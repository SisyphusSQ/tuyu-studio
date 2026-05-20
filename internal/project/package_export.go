package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	CodePackageShotNotReady     = "package_shot_not_ready"
	CodePackagePathInvalid      = "package_path_invalid"
	CodePackageWriteFailed      = "package_write_failed"
	CodePackageRedactionFailed  = "package_redaction_failed"
	CodePackageReferenceIgnored = "package_reference_ignored"
)

const (
	GenerationPackageStatusDraft          = "draft"
	GenerationPackageStatusReady          = "ready"
	GenerationPackageStatusHandedOff      = "handed_off"
	GenerationPackageStatusResultReceived = "result_received"
	GenerationPackageStatusStale          = "stale"
	GenerationPackageStatusInvalid        = "invalid"
)

type ExportGenerationPackageCommand struct {
	Root              string `json:"root"`
	ShotID            string `json:"shotId"`
	ProviderProfileID string `json:"providerProfileId,omitempty"`
	CreatedBy         string `json:"createdBy,omitempty"`
	CorrelationID     string `json:"correlationId"`
}

type GenerationPackageResult struct {
	OK      bool                  `json:"ok"`
	Package *GenerationPackageDTO `json:"package,omitempty"`
	Health  *HealthReport         `json:"health,omitempty"`
	Error   *OperationError       `json:"error,omitempty"`
	Events  []ProjectEvent        `json:"events"`
}

type GenerationPackageDTO struct {
	PackageID               string                          `json:"packageId"`
	ProjectID               string                          `json:"projectId"`
	SceneID                 string                          `json:"sceneId"`
	ShotID                  string                          `json:"shotId"`
	PackageVersion          int                             `json:"packageVersion"`
	ProviderProfileID       string                          `json:"providerProfileId"`
	GenerationPackageStatus string                          `json:"generationPackageStatus"`
	ContextDigest           string                          `json:"contextDigest"`
	RelativePath            string                          `json:"relativePath"`
	ManifestPath            string                          `json:"manifestPath"`
	PromptPath              string                          `json:"promptPath"`
	ScriptExcerptPath       string                          `json:"scriptExcerptPath"`
	ContinuityPath          string                          `json:"continuityPath"`
	UploadChecklistPath     string                          `json:"uploadChecklistPath"`
	References              []GenerationPackageReferenceDTO `json:"references"`
	CreatedAt               string                          `json:"createdAt"`
}

type GenerationPackageReferenceDTO struct {
	AssetID     string `json:"assetId"`
	SourcePath  string `json:"sourcePath"`
	PackagePath string `json:"packagePath"`
	Digest      string `json:"digest,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	SizeBytes   int64  `json:"sizeBytes,omitempty"`
}

type packageReferenceSource struct {
	assetID      string
	sourcePath   string
	packagePath  string
	digest       string
	mimeType     string
	sizeBytes    int64
	sourceAbs    string
	packageAbs   string
	originalName string
}

func (s *Store) ExportGenerationPackage(command ExportGenerationPackageCommand) GenerationPackageResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.packageFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before exporting a package.")
	}
	if failure := s.packageWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "generation_package"
		return GenerationPackageResult{
			OK:     false,
			Health: ptr(s.HealthReport(root)),
			Error:  &operationError,
			Events: []ProjectEvent{s.event("generation_package.export", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	record, shotErr := s.readShotContextRecord(root, manifest, command.ShotID, correlationID)
	if shotErr != nil {
		return s.packageFailureFromShotResult(*shotErr, correlationID)
	}
	shot := record.shot
	if strings.TrimSpace(shot.Status) != ShotStatusContextReady {
		return s.packageFailure(CodePackageShotNotReady, SeverityBlocking, false, correlationID, "Shot must be context_ready before package export.", "shot status: "+strings.TrimSpace(shot.Status))
	}
	shotReport := s.validateShotContext(root, manifest, shot)
	if len(shotReport.Blocking) > 0 {
		issue := firstShotContextIssue(shotReport)
		code := CodePackageShotNotReady
		if issue.Code == CodeShotReferenceMissing {
			code = CodePackageReferenceMissing
		}
		return s.packageFailure(code, SeverityBlocking, false, correlationID, "Shot context is not exportable.", issue.UserMessage)
	}
	if !manifestHasShotNode(root, manifest, shot.ID) {
		return s.packageFailure(CodeShotGraphNodeMissing, SeverityBlocking, false, correlationID, "Shot graph node is missing.", "shots/"+shot.ID+".json")
	}

	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.packageFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read before package export.", err.Error())
	}
	profiles, err := s.loadProfileRecords(root, manifest)
	if err != nil {
		return s.packageFailure(CodeBindingProfileInvalid, SeverityBlocking, true, correlationID, "Continuity profile library could not be read before package export.", err.Error())
	}

	now := s.now().UTC()
	version := nextGenerationPackageVersion(s.readPackageManifests(root, manifest)[shot.ID])
	providerProfileID := normalizeProviderProfileID(command.ProviderProfileID, manifest.Defaults.ProviderProfileID)
	packageID := generationPackageID(shot, version)
	packageRel := generationPackageDirectory(manifest, shot, version, now)
	manifestRel := path.Join(packageRel, "manifest.json")
	if issues := validateProjectPath("package.directory", packageRel); len(issues) > 0 {
		return s.packageFailure(CodePackagePathInvalid, SeverityBlocking, false, correlationID, "Package directory path is invalid.", issues[0].TechnicalDetail)
	}
	packageAbs := filepath.Join(root, filepath.FromSlash(packageRel))
	if _, err := os.Stat(packageAbs); err == nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package directory already exists.", packageRel)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package directory could not be inspected.", err.Error())
	}

	references, errResult := s.collectPackageReferences(root, shot, index, profiles, packageAbs, correlationID)
	if errResult != nil {
		return *errResult
	}

	dto := GenerationPackageDTO{
		PackageID:               packageID,
		ProjectID:               manifest.Project.ID,
		SceneID:                 shotSceneReferenceID(shot),
		ShotID:                  shot.ID,
		PackageVersion:          version,
		ProviderProfileID:       providerProfileID,
		GenerationPackageStatus: GenerationPackageStatusReady,
		ContextDigest:           packageContextDigest(shot, references, profiles),
		RelativePath:            packageRel,
		ManifestPath:            manifestRel,
		PromptPath:              path.Join(packageRel, "prompt.txt"),
		ScriptExcerptPath:       path.Join(packageRel, "script_excerpt.md"),
		ContinuityPath:          path.Join(packageRel, "continuity.md"),
		UploadChecklistPath:     path.Join(packageRel, "upload_checklist.md"),
		References:              packageReferenceDTOs(references),
		CreatedAt:               now.Format(time.RFC3339),
	}

	manifestData := packageManifest{
		SchemaVersion:           CurrentSchemaVersion,
		ProjectID:               dto.ProjectID,
		SceneID:                 dto.SceneID,
		ShotID:                  dto.ShotID,
		PackageID:               dto.PackageID,
		PackageVersion:          dto.PackageVersion,
		ProviderProfileID:       dto.ProviderProfileID,
		GenerationPackageStatus: GenerationPackageStatusDraft,
		ContextDigest:           dto.ContextDigest,
		PromptPath:              "prompt.txt",
		ScriptExcerptPath:       "script_excerpt.md",
		ContinuityPath:          "continuity.md",
		UploadChecklistPath:     "upload_checklist.md",
		References:              packageReferencePaths(references),
		CreatedAt:               dto.CreatedAt,
		ManifestPath:            manifestRel,
	}

	files := map[string]string{
		"prompt.txt":          packagePromptText(manifest, shot, profiles, references),
		"script_excerpt.md":   s.packageScriptExcerpt(root, manifest, shot),
		"continuity.md":       packageContinuityText(manifest, shot, profiles),
		"upload_checklist.md": packageUploadChecklistText(shot, references),
	}
	manifestJSON, err := encodePackageManifest(manifestData)
	if err != nil {
		return s.packageFailure(CodePackageManifestInvalid, SeverityBlocking, true, correlationID, "Package manifest could not be encoded.", err.Error())
	}
	scanFiles := copyStringMap(files)
	scanFiles["manifest.json"] = string(manifestJSON)
	if scanErr := scanPackageTextFiles(scanFiles); scanErr != nil {
		return s.packageFailure(CodePackageRedactionFailed, SeverityBlocking, false, correlationID, "Package content failed redaction scan.", scanErr.Error())
	}
	if scanErr := scanPackageReferenceTextFiles(references); scanErr != nil {
		return s.packageFailure(CodePackageRedactionFailed, SeverityBlocking, false, correlationID, "Package reference failed redaction scan.", scanErr.Error())
	}

	if err := os.MkdirAll(filepath.Join(packageAbs, "references"), 0o755); err != nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package directory could not be created.", err.Error())
	}
	for _, reference := range references {
		if err := copyPackageReference(reference.sourceAbs, reference.packageAbs); err != nil {
			return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package reference could not be copied.", err.Error())
		}
	}
	for name, content := range files {
		if err := s.atomicWrite(filepath.Join(packageAbs, name), []byte(content), 0o644, s.instanceID); err != nil {
			return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package file could not be written.", err.Error())
		}
	}
	if err := s.atomicWrite(filepath.Join(packageAbs, "manifest.json"), manifestJSON, 0o644, s.instanceID); err != nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package manifest could not be written.", err.Error())
	}
	if items := s.checkPackageManifest(root, manifestRel); hasBlockingHealthItem(items) {
		return s.packageFailure(CodePackageManifestInvalid, SeverityBlocking, false, correlationID, "Package manifest failed validation after export.", healthCodesDetail(items))
	}

	shot.PackageIDs = addUniqueString(shot.PackageIDs, packageID)
	shot.UpdatedAt = now.Format(time.RFC3339)
	patchShotRaw(record.raw, shot)
	record.raw["packageIds"] = shot.PackageIDs
	if err := s.writeShotRaw(root, manifest, shot.ID, record.raw); err != nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Shot package linkage could not be saved.", err.Error())
	}
	if err := s.updateManifestForPackage(root, manifest, shot, packageID, manifestRel, GenerationPackageStatusDraft, now); err != nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Project manifest package node could not be saved.", err.Error())
	}
	manifestData.GenerationPackageStatus = GenerationPackageStatusReady
	manifestJSON, err = encodePackageManifest(manifestData)
	if err != nil {
		return s.packageFailure(CodePackageManifestInvalid, SeverityBlocking, true, correlationID, "Package manifest could not be encoded.", err.Error())
	}
	if scanErr := scanPackageTextFiles(map[string]string{"manifest.json": string(manifestJSON)}); scanErr != nil {
		return s.packageFailure(CodePackageRedactionFailed, SeverityBlocking, false, correlationID, "Package manifest failed redaction scan.", scanErr.Error())
	}
	if err := s.atomicWrite(filepath.Join(packageAbs, "manifest.json"), manifestJSON, 0o644, s.instanceID); err != nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Package manifest could not be marked ready.", err.Error())
	}
	if items := s.checkPackageManifest(root, manifestRel); hasBlockingHealthItem(items) {
		return s.packageFailure(CodePackageManifestInvalid, SeverityBlocking, false, correlationID, "Package manifest failed validation after ready transition.", healthCodesDetail(items))
	}
	if err := s.updateManifestPackageStatus(root, manifestRel, packageID, GenerationPackageStatusReady, now); err != nil {
		return s.packageFailure(CodePackageWriteFailed, SeverityBlocking, true, correlationID, "Project manifest package status could not be marked ready.", err.Error())
	}

	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("generation_package.exported", correlationID),
		EventType:     "generation_package.exported",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "GenerationPackage exported for manual handoff.",
		Details: map[string]any{
			"packageId":      packageID,
			"shotId":         shot.ID,
			"packageVersion": version,
			"path":           packageRel,
			"references":     len(references),
			"createdBy":      normalizeCreatedBy(command.CreatedBy),
		},
	})

	health := s.HealthReport(root)
	return GenerationPackageResult{
		OK:      true,
		Package: &dto,
		Health:  &health,
		Events:  []ProjectEvent{s.event("generation_package.exported", "completed", "GenerationPackage exported for manual handoff.", correlationID, nil)},
	}
}

func (s *Store) packageWriteLockFailure(root string, correlationID string) *GenerationPackageResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.packageFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before exporting a package.")
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.packageFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before exporting a package.", "Open the project with takeover, then retry package export.")
		return &result
	}
	return nil
}

func (s *Store) packageFailureFromShotResult(result ShotContextResult, correlationID string) GenerationPackageResult {
	if result.Error == nil {
		return s.packageFailure(CodeShotMissing, SeverityBlocking, false, correlationID, "Shot could not be loaded for package export.", "Shot context result did not include an error.")
	}
	err := *result.Error
	err.TargetType = "generation_package"
	return GenerationPackageResult{
		OK:     false,
		Health: nil,
		Error:  &err,
		Events: []ProjectEvent{s.event("generation_package.export", "blocked", err.UserMessage, correlationID, &err)},
	}
}

func (s *Store) packageFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) GenerationPackageResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, packageRecoveryActions(code, technicalDetail))
	err.TargetType = "generation_package"
	return GenerationPackageResult{
		OK:     false,
		Error:  &err,
		Events: []ProjectEvent{s.event("generation_package.export", "blocked", userMessage, correlationID, &err)},
	}
}

func packageRecoveryActions(code string, detail string) []string {
	switch code {
	case CodePackageShotNotReady:
		return []string{"Promote the Shot to context_ready before exporting the handoff package."}
	case CodePackageReferenceMissing:
		return []string{"Restore or relink missing reference assets, then retry package export."}
	case CodePackagePathInvalid:
		return []string{"Keep package paths inside the project and remove absolute, URL, or traversal segments."}
	case CodePackageRedactionFailed:
		return []string{"Remove private paths, credentials, tokens, or external private URLs from package content before retrying."}
	case CodePackageWriteFailed:
		return []string{"Retry export. Existing ready packages are not overwritten by re-export."}
	case CodeProjectActiveLock:
		return []string{"Open the project read-only or confirm lock takeover before exporting."}
	case CodeProjectStaleLock:
		return []string{"Open the project with takeover, then retry package export."}
	case CodeShotGraphNodeMissing:
		return []string{"Restore the Shot graph node before exporting the handoff package."}
	default:
		if strings.TrimSpace(detail) != "" {
			return []string{detail}
		}
		return []string{"Fix the package export input and retry."}
	}
}

func (s *Store) collectPackageReferences(root string, shot ShotCardDTO, index AssetIndex, profiles map[string]profileRecord, packageAbs string, correlationID string) ([]packageReferenceSource, *GenerationPackageResult) {
	assetIDs := shotPackageAssetIDs(shot)
	for _, characterID := range cleanStringList(shot.CharacterIDs) {
		if profile, ok := profiles[profileKey(BindingTargetCharacter, characterID)]; ok {
			assetIDs = append(assetIDs, profilePackageAssetIDs(profile.dto)...)
		}
	}
	if sceneID := shotSceneReferenceID(shot); sceneID != "" {
		if profile, ok := profiles[profileKey(BindingTargetScene, sceneID)]; ok {
			assetIDs = append(assetIDs, profilePackageAssetIDs(profile.dto)...)
		}
	}
	for _, propID := range cleanStringList(shot.PropIDs) {
		if profile, ok := profiles[profileKey(BindingTargetProp, propID)]; ok {
			assetIDs = append(assetIDs, profilePackageAssetIDs(profile.dto)...)
		}
	}
	assetIDs = dedupeSortedStrings(assetIDs)

	assetMap := assetMapByID(index.Assets)
	references := make([]packageReferenceSource, 0, len(assetIDs))
	for _, assetID := range assetIDs {
		asset, ok := assetMap[assetID]
		if !ok {
			result := s.packageFailure(CodePackageReferenceMissing, SeverityBlocking, false, correlationID, "Package reference asset does not exist.", assetID)
			return nil, &result
		}
		sourcePath := strings.TrimSpace(asset.RelativePath)
		if issues := validateProjectPath("package.reference", sourcePath); len(issues) > 0 {
			result := s.packageFailure(CodePackagePathInvalid, SeverityBlocking, false, correlationID, "Package reference asset path is invalid.", issues[0].TechnicalDetail)
			return nil, &result
		}
		sourceAbs, ok := safeProjectFilePath(root, sourcePath)
		if !ok || !fileExists(sourceAbs) {
			result := s.packageFailure(CodePackageReferenceMissing, SeverityBlocking, true, correlationID, "Package reference asset is missing.", sourcePath)
			return nil, &result
		}
		packagePath := path.Join("references", packageReferenceFilename(asset))
		references = append(references, packageReferenceSource{
			assetID:      asset.ID,
			sourcePath:   sourcePath,
			packagePath:  packagePath,
			digest:       asset.Digest,
			mimeType:     asset.MimeType,
			sizeBytes:    asset.SizeBytes,
			sourceAbs:    sourceAbs,
			packageAbs:   filepath.Join(packageAbs, filepath.FromSlash(packagePath)),
			originalName: asset.OriginalName,
		})
	}
	return references, nil
}

func shotPackageAssetIDs(shot ShotCardDTO) []string {
	assetIDs := cleanStringList(shot.ReferenceAssetIDs)
	for _, ref := range shot.CharacterRefs {
		if id := strings.TrimSpace(ref.ReferenceAssetID); id != "" {
			assetIDs = append(assetIDs, id)
		}
	}
	return cleanStringList(assetIDs)
}

func profilePackageAssetIDs(profile ProfileDTO) []string {
	ids := cleanStringList(profile.ReferenceAssetIDs)
	if strings.TrimSpace(profile.MainReferenceAssetID) != "" {
		ids = append(ids, strings.TrimSpace(profile.MainReferenceAssetID))
	}
	return cleanStringList(ids)
}

func packageReferenceFilename(asset AssetDTO) string {
	ext := strings.ToLower(path.Ext(strings.TrimSpace(asset.RelativePath)))
	if ext == "" {
		ext = strings.ToLower(path.Ext(strings.TrimSpace(asset.OriginalName)))
	}
	if ext == "" {
		ext = ".dat"
	}
	return safeFileToken(asset.ID) + ext
}

func packageReferenceDTOs(references []packageReferenceSource) []GenerationPackageReferenceDTO {
	dtos := make([]GenerationPackageReferenceDTO, 0, len(references))
	for _, ref := range references {
		dtos = append(dtos, GenerationPackageReferenceDTO{
			AssetID:     ref.assetID,
			SourcePath:  ref.sourcePath,
			PackagePath: ref.packagePath,
			Digest:      ref.digest,
			MimeType:    ref.mimeType,
			SizeBytes:   ref.sizeBytes,
		})
	}
	return dtos
}

func packageReferencePaths(references []packageReferenceSource) []string {
	paths := make([]string, 0, len(references))
	for _, ref := range references {
		paths = append(paths, ref.packagePath)
	}
	sort.Strings(paths)
	return paths
}

func normalizeProviderProfileID(input string, fallback string) string {
	for _, value := range []string{input, fallback, "provider_manual_handoff"} {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "provider_manual_handoff"
}

func generationPackageID(shot ShotCardDTO, version int) string {
	return "pkg_" + safeFileToken(shot.ID) + fmt.Sprintf("_v%03d", version)
}

func generationPackageDirectory(manifest Manifest, shot ShotCardDTO, version int, now time.Time) string {
	sceneSegment := fmt.Sprintf("scene_%03d", positiveOrDefault(indexSuffix(firstNonEmpty(shot.ScriptSceneID, shot.SceneID, shot.SceneProfileID)), 1))
	shotSegment := fmt.Sprintf("shot_%03d", positiveOrDefault(shot.Index, positiveOrDefault(indexSuffix(shot.ID), 1)))
	timestampTime := now.UTC().Add(time.Duration(version-1) * time.Nanosecond)
	timestamp := timestampTime.Format("20060102T150405") + fmt.Sprintf("%09d", timestampTime.Nanosecond()) + "Z"
	return path.Join(path.Clean(manifest.Paths.Packages), sceneSegment, shotSegment+"_pkg_"+timestamp)
}

func nextGenerationPackageVersion(packages []packageManifest) int {
	version := 1
	for _, pkg := range packages {
		if pkg.PackageVersion >= version {
			version = pkg.PackageVersion + 1
		}
	}
	return version
}

func packageContextDigest(shot ShotCardDTO, references []packageReferenceSource, profiles map[string]profileRecord) string {
	payload := struct {
		Shot       ShotCardDTO                     `json:"shot"`
		References []GenerationPackageReferenceDTO `json:"references"`
		Rules      []ContinuityRuleDTO             `json:"rules"`
	}{
		Shot:       shot,
		References: packageReferenceDTOs(references),
		Rules:      continuityRulesFromProfiles(profiles),
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func packagePromptText(manifest Manifest, shot ShotCardDTO, profiles map[string]profileRecord, references []packageReferenceSource) string {
	lines := []string{
		"# Prompt",
		"",
		"Shot: " + shot.Title,
		"Goal: " + shot.Description,
		"Action: " + shot.Action,
		"Emotion: " + shot.Emotion,
		fmt.Sprintf("Duration: %d seconds", shot.DurationSeconds),
		"Aspect ratio: " + firstNonEmpty(shot.AspectRatio, manifest.Defaults.AspectRatio),
		"Camera: " + strings.TrimSpace(strings.Join([]string{shot.ShotType, shot.CameraMovement}, " / ")),
		"",
		"## Continuity Summary",
	}
	for _, rule := range packageRulesForShot(shot, profiles) {
		lines = append(lines, "- ["+rule.Severity+"] "+rule.Rule)
	}
	if len(references) > 0 {
		lines = append(lines, "", "## References")
		for _, ref := range references {
			lines = append(lines, "- "+ref.assetID+" -> "+ref.packagePath)
		}
	}
	lines = append(lines, "", "## Negative Constraints", "- Do not change locked identity, scene, prop, or style continuity without explicit review.")
	return strings.Join(lines, "\n") + "\n"
}

func (s *Store) packageScriptExcerpt(root string, manifest Manifest, shot ShotCardDTO) string {
	excerpt := selectedScriptExcerpt(root, manifest, shot.SourceRange)
	if strings.TrimSpace(excerpt) == "" {
		excerpt = "- " + firstNonEmpty(shot.Description, shot.Action, shot.Title)
	}
	return strings.Join([]string{
		"# Script Excerpt",
		"",
		"Shot: " + shot.ID,
		fmt.Sprintf("Source range: %d-%d", shot.SourceRange.StartLine, shot.SourceRange.EndLine),
		"",
		excerpt,
		"",
	}, "\n")
}

func selectedScriptExcerpt(root string, manifest Manifest, sourceRange ScriptSourceRange) string {
	if !validPositiveSourceRange(sourceRange) {
		return ""
	}
	for _, node := range manifest.Graph.Nodes {
		if strings.TrimSpace(node.Kind) != "script_document" || strings.TrimSpace(node.RefID) == "" {
			continue
		}
		filename, ok := safeProjectFilePath(root, node.RefID)
		if !ok {
			continue
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		if sourceRange.StartLine > len(lines) {
			continue
		}
		end := sourceRange.EndLine
		if end > len(lines) {
			end = len(lines)
		}
		return strings.Join(lines[sourceRange.StartLine-1:end], "\n")
	}
	return ""
}

func packageContinuityText(manifest Manifest, shot ShotCardDTO, profiles map[string]profileRecord) string {
	lines := []string{
		"# Continuity Snapshot",
		"",
		"Shot: " + shot.ID,
		"Project: " + manifest.Project.ID,
		"",
	}
	rules := packageRulesForShot(shot, profiles)
	if len(rules) == 0 {
		lines = append(lines, "- No locked continuity rules were selected for this Shot.")
	} else {
		for _, rule := range rules {
			lockLabel := "unlocked"
			if rule.Locked {
				lockLabel = "locked"
			}
			lines = append(lines, "- "+rule.TargetType+"/"+rule.TargetID+" ["+rule.ID+", "+rule.Severity+", "+lockLabel+"]: "+rule.Rule)
		}
	}
	if strings.TrimSpace(manifest.StyleBible.Summary) != "" {
		lines = append(lines, "", "## Style", manifest.StyleBible.Summary)
	}
	return strings.Join(lines, "\n") + "\n"
}

func packageRulesForShot(shot ShotCardDTO, profiles map[string]profileRecord) []ContinuityRuleDTO {
	ids := map[string]struct{}{}
	for _, id := range cleanStringList(shot.ContinuityRuleIDs) {
		ids[id] = struct{}{}
	}
	var rules []ContinuityRuleDTO
	for _, profile := range profiles {
		if !shotReferencesProfile(shot, profile.dto.Type, profile.dto.ID) {
			continue
		}
		for _, rule := range profile.dto.ContinuityRules {
			if len(ids) > 0 {
				if _, ok := ids[rule.ID]; !ok {
					continue
				}
			}
			rules = append(rules, rule)
		}
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	return rules
}

func shotReferencesProfile(shot ShotCardDTO, targetType string, targetID string) bool {
	switch targetType {
	case BindingTargetCharacter:
		return stringListContains(shot.CharacterIDs, targetID)
	case BindingTargetScene:
		return shotSceneReferenceID(shot) == targetID || stringListContains(shot.SceneIDRefs, targetID)
	case BindingTargetProp:
		return stringListContains(shot.PropIDs, targetID)
	default:
		return false
	}
}

func packageUploadChecklistText(shot ShotCardDTO, references []packageReferenceSource) string {
	lines := []string{
		"# Upload Checklist",
		"",
		"- Use `prompt.txt` as the manual instruction source.",
		"- Review `script_excerpt.md` and `continuity.md` before handoff.",
		"- Confirm this is a manual handoff package; do not submit automatically from Tuyu Studio.",
	}
	if len(references) > 0 {
		lines = append(lines, "- Attach these package-local references:")
		for _, ref := range references {
			lines = append(lines, "  - "+ref.packagePath+" ("+ref.assetID+")")
		}
	} else {
		lines = append(lines, "- No package-local references were included.")
	}
	lines = append(lines, "- Record external platform actions outside this package until provider submit is implemented.", "- After receiving a result, import it through the Result workflow for Shot `"+shot.ID+"`.")
	return strings.Join(lines, "\n") + "\n"
}

func scanPackageTextFiles(files map[string]string) error {
	for filename, content := range files {
		if packageTextLooksSensitive(content) {
			return fmt.Errorf("%s contains private path, external URL, credential, or token-like text", filename)
		}
	}
	return nil
}

func packageTextLooksSensitive(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if looksSensitive(strings.TrimSpace(line)) {
			return true
		}
		lower := strings.ToLower(line)
		for _, marker := range []string{"/users/", "/var/folders", "api_key=", "apikey=", "token=", "secret=", "credential=", "file://", "http://", "https://"} {
			if strings.Contains(lower, marker) {
				return true
			}
		}
	}
	return false
}

func scanPackageReferenceTextFiles(references []packageReferenceSource) error {
	for _, reference := range references {
		if !packageReferenceLooksTextual(reference) {
			continue
		}
		data, err := os.ReadFile(reference.sourceAbs)
		if err != nil {
			return fmt.Errorf("%s could not be read for redaction scan: %w", reference.assetID, err)
		}
		if packageTextLooksSensitive(string(data)) {
			return fmt.Errorf("%s contains private path, external URL, credential, or token-like text", reference.assetID)
		}
	}
	return nil
}

func packageReferenceLooksTextual(reference packageReferenceSource) bool {
	mimeType := strings.ToLower(strings.TrimSpace(reference.mimeType))
	if strings.HasPrefix(mimeType, "text/") || strings.Contains(mimeType, "json") || strings.Contains(mimeType, "xml") {
		return true
	}
	switch strings.ToLower(path.Ext(reference.sourcePath)) {
	case ".txt", ".md", ".markdown", ".json", ".yaml", ".yml", ".csv", ".srt", ".vtt", ".xml":
		return true
	default:
		return false
	}
}

func encodePackageManifest(manifest packageManifest) ([]byte, error) {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func copyStringMap(values map[string]string) map[string]string {
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func copyPackageReference(source string, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func (s *Store) updateManifestForPackage(root string, manifest Manifest, shot ShotCardDTO, packageID string, manifestRel string, status string, now time.Time) error {
	packageNodeID := "node_package_" + safeFileToken(packageID)
	shotNodeID := manifestShotNodeID(root, manifest, shot.ID)
	if shotNodeID == "" {
		return fmt.Errorf("%s: %s", CodeShotGraphNodeMissing, shot.ID)
	}
	for index := range manifest.Graph.Nodes {
		node := &manifest.Graph.Nodes[index]
		if strings.TrimSpace(node.ID) == packageNodeID {
			node.RefID = manifestRel
			node.Status = status
		}
	}
	if !manifestHasNodeID(manifest, packageNodeID) {
		manifest.Graph.Nodes = append(manifest.Graph.Nodes, Node{
			ID:     packageNodeID,
			Kind:   "package",
			RefID:  manifestRel,
			Status: status,
		})
	}
	edgeID := "edge_" + safeFileToken(shotNodeID+"_"+packageNodeID+"_handoff")
	if !manifestHasEdgeID(manifest, edgeID) {
		manifest.Graph.Edges = append(manifest.Graph.Edges, Edge{
			ID:     edgeID,
			Source: shotNodeID,
			Target: packageNodeID,
			Kind:   "handoff_candidate",
		})
	}
	manifest.Project.UpdatedAt = now.UTC().Format(time.RFC3339)
	manifest.Graph.Version++
	manifest.Integrity.LastGraphVersion = manifest.Graph.Version
	manifest.Integrity.LastCleanShutdown = true
	data, err := EncodeManifest(manifest)
	if err != nil {
		return err
	}
	return s.atomicWrite(filepath.Join(root, ManifestFileName), data, 0o644, s.instanceID)
}

func (s *Store) updateManifestPackageStatus(root string, manifestRel string, packageID string, status string, now time.Time) error {
	manifest, _, err := s.readManifest(root)
	if err != nil {
		return err
	}
	updated := false
	for index := range manifest.Graph.Nodes {
		node := &manifest.Graph.Nodes[index]
		if strings.TrimSpace(node.Kind) != "package" {
			continue
		}
		if path.Clean(strings.ReplaceAll(node.RefID, "\\", "/")) == manifestRel || strings.TrimSpace(node.ID) == "node_package_"+safeFileToken(packageID) {
			node.Status = status
			updated = true
		}
	}
	if !updated {
		return nil
	}
	manifest.Project.UpdatedAt = now.UTC().Format(time.RFC3339)
	manifest.Graph.Version++
	manifest.Integrity.LastGraphVersion = manifest.Graph.Version
	manifest.Integrity.LastCleanShutdown = true
	data, err := EncodeManifest(manifest)
	if err != nil {
		return err
	}
	return s.atomicWrite(filepath.Join(root, ManifestFileName), data, 0o644, s.instanceID)
}

func (s *Store) markPackagesStaleForShots(root string, manifest *Manifest, shotIDs []string, now time.Time) (bool, error) {
	packages := s.readPackageManifests(root, *manifest)
	graphUpdated := false
	for _, shotID := range cleanStringList(shotIDs) {
		for _, packageData := range packages[shotID] {
			if !shouldMarkPackageStale(packageData.GenerationPackageStatus) || strings.TrimSpace(packageData.ManifestPath) == "" {
				continue
			}
			packageData.GenerationPackageStatus = GenerationPackageStatusStale
			if err := s.writePackageManifest(root, packageData); err != nil {
				return graphUpdated, err
			}
			for index := range manifest.Graph.Nodes {
				node := &manifest.Graph.Nodes[index]
				if strings.TrimSpace(node.Kind) != "package" {
					continue
				}
				if path.Clean(strings.ReplaceAll(node.RefID, "\\", "/")) == packageData.ManifestPath || strings.TrimSpace(node.ID) == "node_package_"+safeFileToken(packageData.PackageID) {
					if node.Status != GenerationPackageStatusStale {
						node.Status = GenerationPackageStatusStale
						graphUpdated = true
					}
				}
			}
		}
	}
	return graphUpdated, nil
}

func shouldMarkPackageStale(status string) bool {
	return strings.TrimSpace(status) != GenerationPackageStatusStale
}

func (s *Store) writePackageManifest(root string, manifest packageManifest) error {
	relative := strings.TrimSpace(manifest.ManifestPath)
	if relative == "" {
		return fmt.Errorf("package manifest path is empty")
	}
	if issues := validateProjectPath("package.manifest", relative); len(issues) > 0 {
		return errors.New(issues[0].TechnicalDetail)
	}
	data, err := encodePackageManifest(manifest)
	if err != nil {
		return err
	}
	filename := filepath.Join(root, filepath.FromSlash(relative))
	return s.atomicWrite(filename, data, 0o644, s.instanceID)
}

func positiveOrDefault(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func indexSuffix(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	end := len(value)
	start := end
	for start > 0 && value[start-1] >= '0' && value[start-1] <= '9' {
		start--
	}
	if start == end {
		return 0
	}
	var parsed int
	for _, ch := range value[start:end] {
		parsed = parsed*10 + int(ch-'0')
	}
	return parsed
}

func manifestHasNodeID(manifest Manifest, id string) bool {
	for _, node := range manifest.Graph.Nodes {
		if strings.TrimSpace(node.ID) == id {
			return true
		}
	}
	return false
}

func manifestShotNodeID(root string, manifest Manifest, shotID string) string {
	for _, node := range manifest.Graph.Nodes {
		if strings.TrimSpace(node.Kind) != "shot" || strings.TrimSpace(node.RefID) == "" {
			continue
		}
		filename, ok := safeProjectFilePath(root, node.RefID)
		if !ok {
			continue
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		var value struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(data, &value) == nil && strings.TrimSpace(value.ID) == strings.TrimSpace(shotID) {
			return strings.TrimSpace(node.ID)
		}
	}
	return ""
}

func manifestHasEdgeID(manifest Manifest, id string) bool {
	for _, edge := range manifest.Graph.Edges {
		if strings.TrimSpace(edge.ID) == id {
			return true
		}
	}
	return false
}

func hasBlockingHealthItem(items []HealthItem) bool {
	for _, item := range items {
		if item.Severity == SeverityBlocking {
			return true
		}
	}
	return false
}

func healthCodesDetail(items []HealthItem) string {
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	sort.Strings(codes)
	return strings.Join(codes, ", ")
}
