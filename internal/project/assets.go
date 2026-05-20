package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const AssetIndexRelativePath = "assets/index.json"

const (
	CodeAssetFileUnreadable     = "asset_file_unreadable"
	CodeAssetMimeRejected       = "asset_mime_rejected"
	CodeAssetDigestFailed       = "asset_digest_failed"
	CodeAssetDuplicateFound     = "asset_duplicate_found"
	CodeAssetThumbnailFailed    = "asset_thumbnail_failed"
	CodeManagedReferenceRisky   = "managed_reference_risky"
	CodeAssetIndexInvalid       = "asset_index_invalid"
	CodeAssetIndexWriteFailed   = "asset_index_write_failed"
	CodeAssetPathInvalid        = "asset_path_invalid"
	CodeAssetSourcePathRequired = "asset_source_path_required"
)

const (
	AssetTypeImage = "image"
	AssetTypeVideo = "video"
	AssetTypeAudio = "audio"
	AssetTypeText  = "text"
)

const (
	AssetDuplicatePolicyCancel = "cancel"
	AssetDuplicatePolicyReuse  = "reuse"
	AssetDuplicatePolicyCopy   = "copy"
)

const (
	AssetSourceImported         = "imported_file"
	AssetSourceManagedReference = "managed_reference"
)

const (
	AssetThumbnailPlaceholder = "placeholder"
	AssetThumbnailFailed      = "thumbnail_failed"
	AssetThumbnailNone        = "none"
)

var assetTokenPattern = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

type ImportAssetCommand struct {
	Root             string `json:"root"`
	SourcePath       string `json:"sourcePath"`
	Role             string `json:"role,omitempty"`
	DuplicatePolicy  string `json:"duplicatePolicy,omitempty"`
	ManagedReference bool   `json:"managedReference,omitempty"`
	CorrelationID    string `json:"correlationId"`
}

type ListAssetsCommand struct {
	Root          string `json:"root"`
	CorrelationID string `json:"correlationId"`
}

type AssetLibraryResult struct {
	OK        bool               `json:"ok"`
	Asset     *AssetDTO          `json:"asset,omitempty"`
	Assets    []AssetDTO         `json:"assets"`
	Duplicate *AssetDuplicateDTO `json:"duplicate,omitempty"`
	Health    *HealthReport      `json:"health,omitempty"`
	Error     *OperationError    `json:"error,omitempty"`
	Events    []ProjectEvent     `json:"events"`
}

type AssetDuplicateDTO struct {
	ExistingAssetID string `json:"existingAssetId"`
	Digest          string `json:"digest"`
	MimeType        string `json:"mimeType"`
	Policy          string `json:"policy"`
}

type AssetDTO struct {
	ID              string            `json:"id"`
	ProjectID       string            `json:"projectId"`
	Type            string            `json:"type"`
	Role            string            `json:"role"`
	RelativePath    string            `json:"relativePath"`
	OriginalName    string            `json:"originalName"`
	MimeType        string            `json:"mimeType"`
	SizeBytes       int64             `json:"sizeBytes"`
	Digest          string            `json:"digest"`
	Source          AssetSourceDTO    `json:"source"`
	Bindings        []AssetBindingDTO `json:"bindings"`
	ThumbnailPath   string            `json:"thumbnailPath,omitempty"`
	ThumbnailStatus string            `json:"thumbnailStatus"`
	DigestSummary   string            `json:"digestSummary"`
	BindingCount    int               `json:"bindingCount"`
	Missing         bool              `json:"missing"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

type AssetSourceDTO struct {
	Kind         string `json:"kind"`
	OriginalName string `json:"originalName,omitempty"`
	ImportedAt   string `json:"importedAt,omitempty"`
}

type AssetBindingDTO struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Role       string `json:"role,omitempty"`
}

type AssetIndex struct {
	ProjectID string     `json:"projectId"`
	Assets    []AssetDTO `json:"assets"`
	UpdatedAt string     `json:"updatedAt"`
}

type assetSourceFacts struct {
	filename string
	size     int64
	mimeType string
	kind     string
	digest   string
	ext      string
}

func (s *Store) ImportAsset(command ImportAssetCommand) AssetLibraryResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.assetFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before importing an asset.")
	}

	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		return s.assetFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before importing an asset.")
	}
	if lockInfo.State == LockStateStale {
		return s.assetFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before importing an asset.", "Open the project with takeover, then retry asset import.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "asset"
		return AssetLibraryResult{
			OK:     false,
			Assets: []AssetDTO{},
			Error:  &operationError,
			Events: []ProjectEvent{s.event("asset.import", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	sourcePath, err := resolveAssetSourcePath(root, command.SourcePath)
	if err != nil {
		code := CodeAssetFileUnreadable
		if errors.Is(err, errAssetSourcePathRequired) {
			code = CodeAssetSourcePathRequired
		} else if errors.Is(err, errAssetPathInvalid) {
			code = CodeAssetPathInvalid
		}
		return s.assetFailure(code, SeverityBlocking, true, correlationID, "Asset source file could not be read.", err.Error())
	}

	facts, err := readAssetSourceFacts(sourcePath)
	if err != nil {
		return s.assetFailure(assetReadErrorCode(err), SeverityBlocking, true, correlationID, "Asset source file could not be imported.", err.Error())
	}

	now := s.now().UTC()
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.assetFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}

	duplicate := findDuplicateAsset(index.Assets, facts.digest, facts.mimeType)
	policy := normalizeDuplicatePolicy(command.DuplicatePolicy)
	if duplicate != nil && policy != AssetDuplicatePolicyCopy {
		duplicateDTO := AssetDuplicateDTO{
			ExistingAssetID: duplicate.ID,
			Digest:          facts.digest,
			MimeType:        facts.mimeType,
			Policy:          policy,
		}
		if policy == AssetDuplicatePolicyReuse {
			asset := s.hydrateAssetForList(root, *duplicate)
			assets := s.hydrateAssetsForList(root, index.Assets)
			return AssetLibraryResult{
				OK:        true,
				Asset:     &asset,
				Assets:    assets,
				Duplicate: &duplicateDTO,
				Health:    ptr(s.HealthReport(root)),
				Events:    []ProjectEvent{s.event("asset.import.reused", "completed", "Existing asset reused for duplicate digest.", correlationID, nil)},
			}
		}

		err := s.assetOperationError(CodeAssetDuplicateFound, SeverityWarning, false, correlationID, "Asset with the same digest and mime type already exists.", duplicate.ID, []string{"Choose reuse to keep the existing asset, copy to create another indexed asset, or cancel the import."})
		return AssetLibraryResult{
			OK:        false,
			Assets:    s.hydrateAssetsForList(root, index.Assets),
			Duplicate: &duplicateDTO,
			Health:    ptr(s.HealthReport(root)),
			Error:     &err,
			Events:    []ProjectEvent{s.event("asset.import.duplicate", "blocked", err.UserMessage, correlationID, &err)},
		}
	}

	role := normalizeAssetRole(command.Role)
	assetID := uniqueAssetID(index.Assets, facts.kind, facts.digest, policy)
	relativePath := assetRelativePath(manifest.Paths, facts, role, assetID)
	thumbnailPath := assetThumbnailPath(manifest.Paths, assetID)
	sourceKind := AssetSourceImported
	if command.ManagedReference {
		sourceKind = AssetSourceManagedReference
		relativePath = managedReferencePath(manifest.Paths, assetID)
	}

	if issues := validateProjectPath("asset.relativePath", relativePath); len(issues) > 0 {
		return s.assetFailure(CodeAssetPathInvalid, SeverityBlocking, false, correlationID, "Asset destination path is invalid.", issues[0].TechnicalDetail)
	}
	if issues := validateProjectPath("asset.thumbnailPath", thumbnailPath); len(issues) > 0 {
		return s.assetFailure(CodeAssetPathInvalid, SeverityBlocking, false, correlationID, "Asset thumbnail path is invalid.", issues[0].TechnicalDetail)
	}

	destination := filepath.Join(root, filepath.FromSlash(relativePath))
	copiedPaths := []string{}
	cleanupOnFailure := true
	defer func() {
		if !cleanupOnFailure {
			return
		}
		for _, filename := range copiedPaths {
			_ = os.Remove(filename)
		}
	}()

	if command.ManagedReference {
		data, marshalErr := json.MarshalIndent(map[string]any{
			"assetId":      assetID,
			"originalName": facts.filename,
			"mimeType":     facts.mimeType,
			"digest":       facts.digest,
			"source":       AssetSourceManagedReference,
			"createdAt":    now.Format(time.RFC3339),
		}, "", "  ")
		if marshalErr != nil {
			return s.assetFailure(CodeAssetIndexWriteFailed, SeverityBlocking, true, correlationID, "Managed reference placeholder could not be encoded.", marshalErr.Error())
		}
		data = append(data, '\n')
		if err := s.atomicWrite(destination, data, 0o644, s.instanceID); err != nil {
			return s.assetFailure(CodeAssetIndexWriteFailed, SeverityBlocking, true, correlationID, "Managed reference placeholder could not be written.", err.Error())
		}
		copiedPaths = append(copiedPaths, destination)
	} else if err := copyFileAtomic(sourcePath, destination, s.instanceID); err != nil {
		return s.assetFailure(CodeAssetFileUnreadable, SeverityBlocking, true, correlationID, "Asset source file could not be copied into the project.", err.Error())
	} else {
		copiedPaths = append(copiedPaths, destination)
		destinationDigest, destinationSize, err := digestAssetFile(destination)
		if err != nil {
			return s.assetFailure(CodeAssetDigestFailed, SeverityBlocking, true, correlationID, "Imported asset file could not be verified.", err.Error())
		}
		if destinationDigest != facts.digest || destinationSize != facts.size {
			return s.assetFailure(CodeAssetDigestFailed, SeverityBlocking, true, correlationID, "Asset source changed during import.", "The copied project asset did not match the source digest and size captured before copy.")
		}
	}

	thumbnailStatus := AssetThumbnailPlaceholder
	thumbnailWriteErr := writeThumbnailPlaceholder(root, thumbnailPath, assetID, facts, s.instanceID, s.atomicWrite)
	if thumbnailWriteErr != nil {
		thumbnailPath = ""
		thumbnailStatus = AssetThumbnailFailed
	} else {
		copiedPaths = append(copiedPaths, filepath.Join(root, filepath.FromSlash(thumbnailPath)))
	}

	asset := AssetDTO{
		ID:              assetID,
		ProjectID:       manifest.Project.ID,
		Type:            facts.kind,
		Role:            role,
		RelativePath:    relativePath,
		OriginalName:    facts.filename,
		MimeType:        facts.mimeType,
		SizeBytes:       facts.size,
		Digest:          facts.digest,
		Source:          AssetSourceDTO{Kind: sourceKind, OriginalName: facts.filename, ImportedAt: now.Format(time.RFC3339)},
		Bindings:        []AssetBindingDTO{},
		ThumbnailPath:   thumbnailPath,
		ThumbnailStatus: thumbnailStatus,
		CreatedAt:       now.Format(time.RFC3339),
		UpdatedAt:       now.Format(time.RFC3339),
	}
	index.Assets = append(index.Assets, asset)
	index.UpdatedAt = now.Format(time.RFC3339)
	sortAssets(index.Assets)

	if err := s.writeAssetIndex(root, index); err != nil {
		return s.assetFailure(CodeAssetIndexWriteFailed, SeverityBlocking, true, correlationID, "Asset index could not be written.", err.Error())
	}
	cleanupOnFailure = false

	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("asset.imported", correlationID),
		EventType:     "asset.imported",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Asset imported and indexed.",
		Details: map[string]any{
			"assetId":          asset.ID,
			"type":             asset.Type,
			"role":             asset.Role,
			"relativePath":     asset.RelativePath,
			"digest":           asset.Digest,
			"thumbnailStatus":  asset.ThumbnailStatus,
			"managedReference": command.ManagedReference,
		},
	})

	events := []ProjectEvent{s.event("asset.imported", "completed", "Asset imported and indexed.", correlationID, nil)}
	if thumbnailWriteErr != nil {
		thumbnailError := s.assetOperationError(CodeAssetThumbnailFailed, SeverityWarning, true, correlationID, "Asset thumbnail placeholder could not be written.", thumbnailWriteErr.Error(), []string{"The asset is ready; rebuild the thumbnail placeholder later."})
		events = append(events, s.event("asset.thumbnail", "failed", thumbnailError.UserMessage, correlationID, &thumbnailError))
	}
	if command.ManagedReference {
		managedError := s.assetOperationError(CodeManagedReferenceRisky, SeverityWarning, true, correlationID, "Managed reference requires health review before production use.", asset.ID, []string{"Copy the file into the project when possible, or keep the managed reference explicitly tracked."})
		events = append(events, s.event("asset.managed_reference", "completed", managedError.UserMessage, correlationID, &managedError))
	}

	asset = s.hydrateAssetForList(root, asset)
	return AssetLibraryResult{
		OK:     true,
		Asset:  &asset,
		Assets: s.hydrateAssetsForList(root, index.Assets),
		Health: ptr(s.HealthReport(root)),
		Events: events,
	}
}

func (s *Store) ListAssets(command ListAssetsCommand) AssetLibraryResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.assetFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before listing assets.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "asset"
		return AssetLibraryResult{
			OK:     false,
			Assets: []AssetDTO{},
			Error:  &operationError,
			Events: []ProjectEvent{s.event("asset.list", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.assetFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}

	return AssetLibraryResult{
		OK:     true,
		Assets: s.hydrateAssetsForList(root, index.Assets),
		Health: ptr(s.HealthReport(root)),
		Events: []ProjectEvent{s.event("asset.listed", "completed", "Asset library listed.", correlationID, nil)},
	}
}

func (s *Store) loadAssetIndex(root string, projectID string) (AssetIndex, error) {
	index := AssetIndex{ProjectID: projectID, Assets: []AssetDTO{}}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(AssetIndexRelativePath)))
	if errors.Is(err, os.ErrNotExist) {
		return index, nil
	}
	if err != nil {
		return AssetIndex{}, err
	}
	if err := json.Unmarshal(data, &index); err != nil {
		return AssetIndex{}, err
	}
	if strings.TrimSpace(index.ProjectID) == "" {
		index.ProjectID = projectID
	}
	if index.Assets == nil {
		index.Assets = []AssetDTO{}
	}
	for i := range index.Assets {
		if index.Assets[i].Bindings == nil {
			index.Assets[i].Bindings = []AssetBindingDTO{}
		}
		if index.Assets[i].ThumbnailStatus == "" {
			index.Assets[i].ThumbnailStatus = AssetThumbnailNone
		}
	}
	sortAssets(index.Assets)
	return index, nil
}

func (s *Store) writeAssetIndex(root string, index AssetIndex) error {
	if index.Assets == nil {
		index.Assets = []AssetDTO{}
	}
	for i := range index.Assets {
		index.Assets[i].DigestSummary = ""
		index.Assets[i].BindingCount = 0
		index.Assets[i].Missing = false
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return s.atomicWrite(filepath.Join(root, filepath.FromSlash(AssetIndexRelativePath)), data, 0o644, s.instanceID)
}

func (s *Store) hydrateAssetsForList(root string, assets []AssetDTO) []AssetDTO {
	hydrated := make([]AssetDTO, 0, len(assets))
	for _, asset := range assets {
		hydrated = append(hydrated, s.hydrateAssetForList(root, asset))
	}
	sortAssets(hydrated)
	return hydrated
}

func (s *Store) hydrateAssetForList(root string, asset AssetDTO) AssetDTO {
	asset.DigestSummary = digestSummary(asset.Digest)
	asset.BindingCount = len(asset.Bindings)
	asset.Missing = assetFileMissing(root, asset.RelativePath)
	if asset.ThumbnailStatus == "" {
		asset.ThumbnailStatus = AssetThumbnailNone
	}
	if asset.Bindings == nil {
		asset.Bindings = []AssetBindingDTO{}
	}
	return asset
}

func (s *Store) checkAssetIndex(root string, manifest Manifest) []HealthItem {
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []HealthItem{{
			Severity:        SeverityBlocking,
			Code:            CodeAssetIndexInvalid,
			Path:            AssetIndexRelativePath,
			AffectedObjects: []string{},
			UserMessage:     "Asset index is invalid or unreadable.",
			TechnicalDetail: err.Error(),
			RecoveryActions: []string{"Restore assets/index.json from backup or rebuild the Asset Library index."},
		}}
	}

	var items []HealthItem
	for _, asset := range index.Assets {
		affected := []string{asset.ID}
		if asset.Source.Kind == AssetSourceManagedReference {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeManagedReferenceRisky,
				Path:            asset.RelativePath,
				AffectedObjects: affected,
				UserMessage:     "Asset is tracked as an explicit managed reference.",
				RecoveryActions: []string{"Copy the file into the project before production handoff, or keep the external reference risk explicitly accepted."},
			})
		}
		if invalid := healthItemsForProjectPath("asset.relativePath", asset.RelativePath, affected); len(invalid) > 0 {
			items = append(items, invalid...)
			continue
		}
		filename := filepath.Join(root, filepath.FromSlash(asset.RelativePath))
		info, err := os.Stat(filename)
		if errors.Is(err, os.ErrNotExist) {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeAssetMissing,
				Path:            asset.RelativePath,
				AffectedObjects: affected,
				UserMessage:     "Asset file indexed in the library is missing.",
				RecoveryActions: []string{"Restore the asset file or remove the stale Asset record after review."},
			})
			continue
		}
		if err != nil {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeHealthCheckIncomplete,
				Path:            asset.RelativePath,
				AffectedObjects: affected,
				UserMessage:     "Asset file could not be read during health check.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Retry health check after verifying project permissions."},
			})
			continue
		}
		if info.IsDir() {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeAssetMissing,
				Path:            asset.RelativePath,
				AffectedObjects: affected,
				UserMessage:     "Asset file indexed in the library points to a directory.",
				RecoveryActions: []string{"Relink the Asset record to a file or reimport the asset."},
			})
			continue
		}
		if asset.Source.Kind == AssetSourceManagedReference {
			continue
		}

		got, _, err := digestAssetFile(filename)
		if err != nil {
			items = append(items, HealthItem{
				Severity:        SeverityWarning,
				Code:            CodeHealthCheckIncomplete,
				Path:            asset.RelativePath,
				AffectedObjects: affected,
				UserMessage:     "Asset file could not be opened during health check.",
				TechnicalDetail: err.Error(),
				RecoveryActions: []string{"Retry health check after verifying project permissions."},
			})
			continue
		}
		if !strings.EqualFold(got, asset.Digest) {
			items = append(items, HealthItem{
				Severity:        SeverityBlocking,
				Code:            CodeDigestMismatch,
				Path:            asset.RelativePath,
				AffectedObjects: affected,
				UserMessage:     "Asset digest does not match the indexed value.",
				RecoveryActions: []string{"Reimport the asset or review it as a new asset version."},
			})
		}
	}
	return dedupeHealthItems(items)
}

func readAssetSourceFacts(filename string) (assetSourceFacts, error) {
	file, err := os.Open(filename)
	if err != nil {
		return assetSourceFacts{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return assetSourceFacts{}, err
	}
	if info.IsDir() {
		return assetSourceFacts{}, fmt.Errorf("%s is a directory", filename)
	}
	if info.Size() <= 0 {
		return assetSourceFacts{}, errAssetMimeRejected
	}

	sniff := make([]byte, 512)
	read, readErr := io.ReadFull(file, sniff)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
		return assetSourceFacts{}, readErr
	}
	sniff = sniff[:read]
	mimeType, kind, ext := detectAssetMime(sniff, filepath.Ext(info.Name()))
	if mimeType == "" || kind == "" {
		return assetSourceFacts{}, errAssetMimeRejected
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return assetSourceFacts{}, err
	}
	digest, err := hashFile(file, sha256.New())
	if err != nil {
		return assetSourceFacts{}, fmt.Errorf("%w: %v", errAssetDigestFailed, err)
	}

	return assetSourceFacts{
		filename: info.Name(),
		size:     info.Size(),
		mimeType: mimeType,
		kind:     kind,
		digest:   digest,
		ext:      ext,
	}, nil
}

func detectAssetMime(sniff []byte, originalExt string) (string, string, string) {
	if len(sniff) == 0 {
		return "", "", ""
	}
	if len(sniff) >= 12 && string(sniff[0:4]) == "RIFF" && string(sniff[8:12]) == "WEBP" {
		return "image/webp", AssetTypeImage, ".webp"
	}
	if len(sniff) >= 12 && string(sniff[0:4]) == "RIFF" && string(sniff[8:12]) == "WAVE" {
		return "audio/wav", AssetTypeAudio, ".wav"
	}
	if len(sniff) >= 12 && string(sniff[4:8]) == "ftyp" {
		brand := strings.ToLower(string(sniff[8:min(len(sniff), 12)]))
		if strings.Contains(brand, "qt") {
			return "video/quicktime", AssetTypeVideo, ".mov"
		}
		return "video/mp4", AssetTypeVideo, ".mp4"
	}
	if len(sniff) >= 3 && string(sniff[:3]) == "ID3" {
		return "audio/mpeg", AssetTypeAudio, ".mp3"
	}
	if len(sniff) >= 2 && sniff[0] == 0xff && (sniff[1]&0xe0) == 0xe0 {
		return "audio/mpeg", AssetTypeAudio, ".mp3"
	}

	detected := http.DetectContentType(sniff)
	base, _, _ := mime.ParseMediaType(detected)
	switch base {
	case "image/png":
		return base, AssetTypeImage, ".png"
	case "image/jpeg":
		return base, AssetTypeImage, ".jpg"
	case "image/gif":
		return base, AssetTypeImage, ".gif"
	case "text/plain":
		if utf8.Valid(sniff) && !containsNUL(sniff) {
			return "text/plain", AssetTypeText, textExtension(originalExt)
		}
	}
	if utf8.Valid(sniff) && !containsNUL(sniff) {
		return "text/plain", AssetTypeText, textExtension(originalExt)
	}
	return "", "", ""
}

func hashFile(reader io.Reader, hasher hash.Hash) (string, error) {
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func digestAssetFile(filename string) (string, int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", 0, err
	}
	if info.IsDir() {
		return "", 0, fmt.Errorf("%s is a directory", filename)
	}
	digest, err := hashFile(file, sha256.New())
	if err != nil {
		return "", 0, err
	}
	return digest, info.Size(), nil
}

func copyFileAtomic(source string, destination string, instanceID string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	tmpName := filepath.Join(filepath.Dir(destination), "."+filepath.Base(destination)+".tmp."+safeFileToken(instanceID))
	output, err := os.OpenFile(tmpName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, destination); err != nil {
		return err
	}
	cleanup = false
	return fsyncDirectory(filepath.Dir(destination))
}

func writeThumbnailPlaceholder(root string, relativePath string, assetID string, facts assetSourceFacts, instanceID string, atomicWrite AtomicWriteFunc) error {
	body := fmt.Sprintf("thumbnail placeholder\nasset_id=%s\ntype=%s\nmime=%s\ndigest=%s\n", assetID, facts.kind, facts.mimeType, facts.digest)
	return atomicWrite(filepath.Join(root, filepath.FromSlash(relativePath)), []byte(body), 0o644, instanceID)
}

func resolveAssetSourcePath(root string, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errAssetSourcePathRequired
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	relative := cleanProjectRelativePath(value)
	if issues := validateProjectPath("asset.sourcePath", relative); len(issues) > 0 {
		return "", fmt.Errorf("%w: %s", errAssetPathInvalid, issues[0].TechnicalDetail)
	}
	return filepath.Join(root, filepath.FromSlash(relative)), nil
}

func assetRelativePath(paths Paths, facts assetSourceFacts, role string, assetID string) string {
	baseDir := assetBaseDir(paths, role)
	readable := safeAssetToken(strings.TrimSuffix(facts.filename, filepath.Ext(facts.filename)))
	if readable == "" {
		readable = facts.kind
	}
	return path.Join(baseDir, readable+"-"+assetID+facts.ext)
}

func assetThumbnailPath(paths Paths, assetID string) string {
	baseDir := cleanProjectRelativePath(paths.AssetThumbnails)
	if baseDir == "" {
		baseDir = DefaultPaths().AssetThumbnails
	}
	return path.Join(baseDir, assetID+".txt")
}

func managedReferencePath(paths Paths, assetID string) string {
	baseDir := cleanProjectRelativePath(paths.AssetRefs)
	if baseDir == "" {
		baseDir = DefaultPaths().AssetRefs
	}
	return path.Join(baseDir, assetID+".managed-reference.json")
}

func assetBaseDir(paths Paths, role string) string {
	defaults := DefaultPaths()
	switch role {
	case "character_ref", "scene_ref", "prop_ref", "style_ref", "reference", "ref":
		return cleanOrDefaultPath(paths.AssetRefs, defaults.AssetRefs)
	case "video_result", "image_result", "result":
		return cleanOrDefaultPath(paths.AssetResults, defaults.AssetResults)
	case "package_file", "generated_output", "output":
		return cleanOrDefaultPath(paths.AssetOutputs, defaults.AssetOutputs)
	default:
		return cleanOrDefaultPath(paths.AssetInputs, defaults.AssetInputs)
	}
}

func cleanOrDefaultPath(value string, fallback string) string {
	value = cleanProjectRelativePath(value)
	if value == "" {
		return fallback
	}
	return value
}

func findDuplicateAsset(assets []AssetDTO, digest string, mimeType string) *AssetDTO {
	for i := range assets {
		if strings.EqualFold(assets[i].Digest, digest) && strings.EqualFold(assets[i].MimeType, mimeType) {
			return &assets[i]
		}
	}
	return nil
}

func uniqueAssetID(assets []AssetDTO, kind string, digest string, duplicatePolicy string) string {
	base := "asset_" + safeAssetToken(kind) + "_" + digestSummary(digest)
	if duplicatePolicy != AssetDuplicatePolicyCopy && !assetIDExists(assets, base) {
		return base
	}
	if !assetIDExists(assets, base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_copy_%03d", base, i)
		if !assetIDExists(assets, candidate) {
			return candidate
		}
	}
}

func assetIDExists(assets []AssetDTO, id string) bool {
	for _, asset := range assets {
		if asset.ID == id {
			return true
		}
	}
	return false
}

func normalizeDuplicatePolicy(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case AssetDuplicatePolicyReuse:
		return AssetDuplicatePolicyReuse
	case AssetDuplicatePolicyCopy, "create_copy", "new_copy":
		return AssetDuplicatePolicyCopy
	default:
		return AssetDuplicatePolicyCancel
	}
}

func normalizeAssetRole(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = assetTokenPattern.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return "other"
	}
	return value
}

func sortAssets(assets []AssetDTO) {
	sort.SliceStable(assets, func(i int, j int) bool {
		if assets[i].CreatedAt == assets[j].CreatedAt {
			return assets[i].ID < assets[j].ID
		}
		return assets[i].CreatedAt < assets[j].CreatedAt
	})
}

func digestSummary(digest string) string {
	digest = strings.TrimSpace(digest)
	if len(digest) <= 12 {
		return digest
	}
	return digest[:12]
}

func safeAssetToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = assetTokenPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_")
	if len(value) > 48 {
		value = value[:48]
		value = strings.Trim(value, "-_")
	}
	return value
}

func textExtension(originalExt string) string {
	ext := strings.ToLower(strings.TrimSpace(originalExt))
	switch ext {
	case ".txt", ".md", ".json", ".csv":
		return ext
	default:
		return ".txt"
	}
}

func containsNUL(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func assetFileMissing(root string, relativePath string) bool {
	if strings.TrimSpace(relativePath) == "" {
		return true
	}
	if len(validateProjectPath("asset.relativePath", relativePath)) > 0 {
		return true
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relativePath)))
	return err != nil || info.IsDir()
}

func assetReadErrorCode(err error) string {
	if errors.Is(err, errAssetMimeRejected) {
		return CodeAssetMimeRejected
	}
	if errors.Is(err, errAssetDigestFailed) {
		return CodeAssetDigestFailed
	}
	return CodeAssetFileUnreadable
}

func (s *Store) assetFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) AssetLibraryResult {
	err := s.assetOperationError(code, severity, retryable, correlationID, userMessage, technicalDetail, assetRecoveryActions(code, technicalDetail))
	return AssetLibraryResult{
		OK:     false,
		Assets: []AssetDTO{},
		Error:  &err,
		Events: []ProjectEvent{s.event("asset.operation", "blocked", userMessage, correlationID, &err)},
	}
}

func (s *Store) assetOperationError(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string, recoveryActions []string) OperationError {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, recoveryActions)
	err.TargetType = "asset"
	return err
}

func assetRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeAssetSourcePathRequired:
		return []string{"Choose a readable source file before importing an asset."}
	case CodeAssetFileUnreadable:
		return []string{"Check source file permissions and retry import with a readable file."}
	case CodeAssetMimeRejected:
		return []string{"Choose a supported image, video, audio, or UTF-8 text file."}
	case CodeAssetDigestFailed:
		return []string{"Retry with a readable source file; if it still fails, treat the file as damaged."}
	case CodeAssetDuplicateFound:
		return []string{"Choose reuse, copy, or cancel for the duplicate asset."}
	case CodeAssetIndexInvalid:
		return []string{"Restore assets/index.json from backup or rebuild the Asset Library index."}
	case CodeAssetIndexWriteFailed:
		return []string{"Retry import. The asset index was not updated."}
	case CodeAssetPathInvalid:
		return []string{"Use a project-relative source path inside the project, or choose a readable absolute source file for copying into the project."}
	default:
		if strings.TrimSpace(detail) != "" {
			return []string{detail}
		}
		return []string{"Retry after fixing the Asset Library input."}
	}
}

var (
	errAssetSourcePathRequired = errors.New("asset source path is required")
	errAssetPathInvalid        = errors.New("asset path invalid")
	errAssetMimeRejected       = errors.New("asset mime rejected")
	errAssetDigestFailed       = errors.New("asset digest failed")
)
