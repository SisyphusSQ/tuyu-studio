package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	CodeBindingAssetMissing       = "binding_asset_missing"
	CodeBindingTargetMissing      = "binding_target_missing"
	CodeBindingDuplicate          = "binding_duplicate"
	CodeBindingPurposeReserved    = "binding_purpose_reserved"
	CodeBindingProfileInvalid     = "binding_profile_invalid"
	CodeBindingWriteFailed        = "binding_write_failed"
	CodeMainReferenceAssetMissing = "main_reference_asset_missing"
)

const (
	BindingTargetCharacter = "character"
	BindingTargetScene     = "scene"
	BindingTargetProp      = "prop"
)

const (
	BindingPurposeReference     = "reference"
	BindingPurposeMainReference = "main_reference"
)

type ListAssetBindingsCommand struct {
	Root          string `json:"root"`
	CorrelationID string `json:"correlationId"`
}

type BindAssetCommand struct {
	Root            string `json:"root"`
	AssetID         string `json:"assetId"`
	TargetType      string `json:"targetType"`
	TargetID        string `json:"targetId"`
	Purpose         string `json:"purpose,omitempty"`
	DuplicatePolicy string `json:"duplicatePolicy,omitempty"`
	CreatedBy       string `json:"createdBy,omitempty"`
	CorrelationID   string `json:"correlationId"`
}

type SetMainReferenceCommand struct {
	Root          string `json:"root"`
	TargetType    string `json:"targetType"`
	TargetID      string `json:"targetId"`
	AssetID       string `json:"assetId,omitempty"`
	Clear         bool   `json:"clear,omitempty"`
	CreatedBy     string `json:"createdBy,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type ContinuityLibraryResult struct {
	OK        bool              `json:"ok"`
	Asset     *AssetDTO         `json:"asset,omitempty"`
	Assets    []AssetDTO        `json:"assets"`
	Profile   *ProfileDTO       `json:"profile,omitempty"`
	Profiles  []ProfileDTO      `json:"profiles"`
	Lineage   []AssetLineageDTO `json:"lineage"`
	Duplicate *AssetBindingDTO  `json:"duplicate,omitempty"`
	Health    *HealthReport     `json:"health,omitempty"`
	Error     *OperationError   `json:"error,omitempty"`
	Events    []ProjectEvent    `json:"events"`
}

type ProfileDTO struct {
	ID                         string              `json:"id"`
	Type                       string              `json:"type"`
	Name                       string              `json:"name"`
	Role                       string              `json:"role,omitempty"`
	Identity                   string              `json:"identity,omitempty"`
	VisualDescription          string              `json:"visualDescription,omitempty"`
	Costume                    string              `json:"costume,omitempty"`
	Location                   string              `json:"location,omitempty"`
	TimeOfDay                  string              `json:"timeOfDay,omitempty"`
	Mood                       string              `json:"mood,omitempty"`
	Lighting                   string              `json:"lighting,omitempty"`
	Category                   string              `json:"category,omitempty"`
	Appearance                 string              `json:"appearance,omitempty"`
	Usage                      string              `json:"usage,omitempty"`
	RelativePath               string              `json:"relativePath"`
	ReferenceAssetIDs          []string            `json:"referenceAssetIds"`
	MainReferenceAssetID       string              `json:"mainReferenceAssetId,omitempty"`
	MainReferencePath          string              `json:"mainReferencePath,omitempty"`
	MainReferenceThumbnailPath string              `json:"mainReferenceThumbnailPath,omitempty"`
	LockedRules                []string            `json:"lockedRules"`
	ContinuityRules            []ContinuityRuleDTO `json:"continuityRules"`
	Bindings                   []AssetBindingDTO   `json:"bindings"`
	BindingCount               int                 `json:"bindingCount"`
	MissingMainReference       bool                `json:"missingMainReference"`
}

type AssetLineageDTO struct {
	AssetID         string                    `json:"assetId"`
	SourceKind      string                    `json:"sourceKind"`
	SourceName      string                    `json:"sourceName,omitempty"`
	RelativePath    string                    `json:"relativePath"`
	Bindings        []AssetBindingDTO         `json:"bindings"`
	TargetSummaries []BindingTargetSummaryDTO `json:"targetSummaries"`
}

type BindingTargetSummaryDTO struct {
	TargetType           string `json:"targetType"`
	TargetID             string `json:"targetId"`
	Name                 string `json:"name"`
	RelativePath         string `json:"relativePath"`
	MainReferenceAssetID string `json:"mainReferenceAssetId,omitempty"`
}

type profileRecord struct {
	dto ProfileDTO
	raw map[string]any
}

func (s *Store) ListAssetBindings(command ListAssetBindingsCommand) ContinuityLibraryResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.bindingFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before listing asset bindings.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "asset_binding"
		return ContinuityLibraryResult{
			OK:       false,
			Assets:   []AssetDTO{},
			Profiles: []ProfileDTO{},
			Lineage:  []AssetLineageDTO{},
			Error:    &operationError,
			Events:   []ProjectEvent{s.event("asset_binding.list", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.bindingFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}
	profiles, err := s.loadProfileRecords(root, manifest)
	if err != nil {
		return s.bindingFailure(CodeBindingProfileInvalid, SeverityBlocking, true, correlationID, "Profile library could not be read.", err.Error())
	}

	return s.bindingSuccess(root, index, profiles, nil, nil, nil, []ProjectEvent{
		s.event("asset_binding.listed", "completed", "Asset bindings and profile lineage listed.", correlationID, nil),
	})
}

func (s *Store) BindAsset(command BindAssetCommand) ContinuityLibraryResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.bindingFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before binding an asset.")
	}
	if failure := s.bindingWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "asset_binding"
		return ContinuityLibraryResult{
			OK:       false,
			Assets:   []AssetDTO{},
			Profiles: []ProfileDTO{},
			Lineage:  []AssetLineageDTO{},
			Error:    &operationError,
			Events:   []ProjectEvent{s.event("asset_binding.bind", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.bindingFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}

	assetIndex := findAssetIndexByID(index.Assets, command.AssetID)
	if assetIndex < 0 {
		return s.bindingFailure(CodeBindingAssetMissing, SeverityBlocking, true, correlationID, "Asset cannot be bound because it does not exist.", strings.TrimSpace(command.AssetID))
	}
	targetType, ok := normalizeBindingTargetType(command.TargetType)
	if !ok {
		return s.bindingFailure(CodeBindingTargetMissing, SeverityBlocking, false, correlationID, "Asset binding target type is not supported.", strings.TrimSpace(command.TargetType))
	}
	profile, err := s.loadProfileRecord(root, manifest, targetType, command.TargetID)
	if err != nil {
		return s.bindingFailure(CodeBindingTargetMissing, SeverityBlocking, true, correlationID, "Asset binding target does not exist.", err.Error())
	}

	now := s.now().UTC()
	asset := index.Assets[assetIndex]
	purpose := normalizeBindingPurpose(command.Purpose)
	if purpose == BindingPurposeMainReference {
		return s.bindingFailure(CodeBindingPurposeReserved, SeverityBlocking, false, correlationID, "Main reference must be set through the main-reference action.", "Use ProjectMainReferenceSet instead of generic asset binding for main_reference.")
	}
	binding := AssetBindingDTO{
		ID:         bindingID(asset.ID, targetType, profile.dto.ID, purpose),
		AssetID:    asset.ID,
		TargetType: targetType,
		TargetID:   profile.dto.ID,
		Purpose:    purpose,
		Locked:     false,
		CreatedBy:  normalizeCreatedBy(command.CreatedBy),
		CreatedAt:  now.Format(time.RFC3339),
	}

	if duplicate := findBinding(asset.Bindings, targetType, profile.dto.ID, purpose); duplicate != nil {
		normalizedDuplicate := normalizeAssetBinding(asset.ID, *duplicate)
		if normalizeDuplicatePolicy(command.DuplicatePolicy) == AssetDuplicatePolicyReuse {
			profile.dto.ReferenceAssetIDs = addUniqueString(profile.dto.ReferenceAssetIDs, asset.ID)
			profile.raw = profileRawWithBaseline(profile.raw, profile.dto)
			profiles, err := s.loadProfileRecords(root, manifest)
			if err == nil {
				profiles[profile.dto.key()] = profile
			}
			return s.bindingSuccess(root, index, profiles, &asset, &profile.dto, &normalizedDuplicate, []ProjectEvent{
				s.event("asset_binding.reused", "completed", "Existing asset binding reused.", correlationID, nil),
			})
		}

		err := s.bindingOperationError(CodeBindingDuplicate, SeverityWarning, false, correlationID, "Asset is already bound to this target and purpose.", normalizedDuplicate.ID, []string{"Choose reuse to keep the existing binding, or select a different purpose."})
		return ContinuityLibraryResult{
			OK:        false,
			Assets:    s.hydrateAssetsForList(root, index.Assets),
			Profiles:  profileDTOsWithBindings(root, index, mustProfileRecords(s.loadProfileRecords(root, manifest))),
			Lineage:   assetLineageDTOs(index.Assets, mustProfileRecords(s.loadProfileRecords(root, manifest))),
			Duplicate: &normalizedDuplicate,
			Health:    ptr(s.HealthReport(root)),
			Error:     &err,
			Events:    []ProjectEvent{s.event("asset_binding.duplicate", "blocked", err.UserMessage, correlationID, &err)},
		}
	}

	index.Assets[assetIndex].Bindings = append(index.Assets[assetIndex].Bindings, binding)
	index.Assets[assetIndex].UpdatedAt = now.Format(time.RFC3339)
	index.UpdatedAt = now.Format(time.RFC3339)
	profile.dto.ReferenceAssetIDs = addUniqueString(profile.dto.ReferenceAssetIDs, asset.ID)
	profile.raw = profileRawWithBaseline(profile.raw, profile.dto)

	if err := s.writeProfileAndIndex(root, profile, index); err != nil {
		return s.bindingFailure(CodeBindingWriteFailed, SeverityBlocking, true, correlationID, "Asset binding could not be written.", err.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("asset_binding.created", correlationID),
		EventType:     "asset_binding.created",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Asset bound to continuity profile.",
		Details: map[string]any{
			"assetId":    asset.ID,
			"bindingId":  binding.ID,
			"targetType": targetType,
			"targetId":   profile.dto.ID,
			"purpose":    purpose,
		},
	})

	profiles, _ := s.loadProfileRecords(root, manifest)
	updatedAsset := index.Assets[assetIndex]
	return s.bindingSuccess(root, index, profiles, &updatedAsset, &profile.dto, &binding, []ProjectEvent{
		s.event("asset_binding.created", "completed", "Asset bound to continuity profile.", correlationID, nil),
	})
}

func (s *Store) SetMainReference(command SetMainReferenceCommand) ContinuityLibraryResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.bindingFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before setting a main reference.")
	}
	if failure := s.bindingWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "asset_binding"
		return ContinuityLibraryResult{
			OK:       false,
			Assets:   []AssetDTO{},
			Profiles: []ProfileDTO{},
			Lineage:  []AssetLineageDTO{},
			Error:    &operationError,
			Events:   []ProjectEvent{s.event("asset_binding.main_reference", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}
	index, err := s.loadAssetIndex(root, manifest.Project.ID)
	if err != nil {
		return s.bindingFailure(CodeAssetIndexInvalid, SeverityBlocking, true, correlationID, "Asset index could not be read.", err.Error())
	}
	targetType, ok := normalizeBindingTargetType(command.TargetType)
	if !ok {
		return s.bindingFailure(CodeBindingTargetMissing, SeverityBlocking, false, correlationID, "Main reference target type is not supported.", strings.TrimSpace(command.TargetType))
	}
	profile, err := s.loadProfileRecord(root, manifest, targetType, command.TargetID)
	if err != nil {
		return s.bindingFailure(CodeBindingTargetMissing, SeverityBlocking, true, correlationID, "Main reference target does not exist.", err.Error())
	}

	now := s.now().UTC()
	var selected *AssetDTO
	assetID := strings.TrimSpace(command.AssetID)
	assetIndex := -1
	if !command.Clear {
		if assetID == "" {
			return s.bindingFailure(CodeMainReferenceAssetMissing, SeverityBlocking, false, correlationID, "Main reference asset id is required unless clear is true.", "Choose an Asset id or set clear=true.")
		}
		assetIndex = findAssetIndexByID(index.Assets, assetID)
		if assetIndex < 0 {
			return s.bindingFailure(CodeMainReferenceAssetMissing, SeverityBlocking, true, correlationID, "Main reference asset does not exist.", assetID)
		}
	}
	if locked := lockedMainReferenceBinding(index.Assets, targetType, profile.dto.ID); locked != nil {
		return s.bindingFailure(CodeContinuityLockedBinding, SeverityBlocking, false, correlationID, "Locked main reference binding cannot be replaced or cleared silently.", locked.ID)
	}
	removeMainReferenceBindings(index.Assets, targetType, profile.dto.ID)
	if command.Clear {
		profile.dto.MainReferenceAssetID = ""
	} else {
		profile.dto.MainReferenceAssetID = index.Assets[assetIndex].ID
		profile.dto.ReferenceAssetIDs = addUniqueString(profile.dto.ReferenceAssetIDs, index.Assets[assetIndex].ID)
		binding := AssetBindingDTO{
			ID:         bindingID(index.Assets[assetIndex].ID, targetType, profile.dto.ID, BindingPurposeMainReference),
			AssetID:    index.Assets[assetIndex].ID,
			TargetType: targetType,
			TargetID:   profile.dto.ID,
			Purpose:    BindingPurposeMainReference,
			Locked:     false,
			CreatedBy:  normalizeCreatedBy(command.CreatedBy),
			CreatedAt:  now.Format(time.RFC3339),
		}
		index.Assets[assetIndex].Bindings = append(index.Assets[assetIndex].Bindings, binding)
		index.Assets[assetIndex].UpdatedAt = now.Format(time.RFC3339)
		selected = &index.Assets[assetIndex]
	}
	index.UpdatedAt = now.Format(time.RFC3339)
	profile.raw = profileRawWithBaseline(profile.raw, profile.dto)

	if err := s.writeProfileAndIndex(root, profile, index); err != nil {
		return s.bindingFailure(CodeBindingWriteFailed, SeverityBlocking, true, correlationID, "Main reference could not be written.", err.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("asset_binding.main_reference", correlationID),
		EventType:     "asset_binding.main_reference",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Profile main reference updated.",
		Details: map[string]any{
			"targetType":           targetType,
			"targetId":             profile.dto.ID,
			"mainReferenceAssetId": profile.dto.MainReferenceAssetID,
			"cleared":              command.Clear,
		},
	})
	changedShots, dirtyErr := s.markShotsDirtyForContinuityChange(root, manifest, targetType, profile.dto.ID, "", "main reference updated", correlationID, now)
	if dirtyErr != nil {
		return s.bindingFailure(CodeContinuityDirtyPropagate, SeverityBlocking, true, correlationID, "Main reference was updated, but affected Shot dirty marking failed.", dirtyErr.Error())
	}

	profiles, _ := s.loadProfileRecords(root, manifest)
	state := "completed"
	summary := "Profile main reference updated."
	if profile.dto.MainReferenceAssetID == "" {
		summary = "Profile main reference cleared."
	}
	if len(changedShots) > 0 {
		summary += " Affected Shots were marked context_dirty."
	}
	return s.bindingSuccess(root, index, profiles, selected, &profile.dto, nil, []ProjectEvent{
		s.event("asset_binding.main_reference", state, summary, correlationID, nil),
	})
}

func (s *Store) bindingWriteLockFailure(root string, correlationID string) *ContinuityLibraryResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.bindingFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before changing bindings.")
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.bindingFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before changing bindings.", "Open the project with takeover, then retry binding.")
		return &result
	}
	return nil
}

func (s *Store) bindingSuccess(root string, index AssetIndex, profiles map[string]profileRecord, asset *AssetDTO, profile *ProfileDTO, duplicate *AssetBindingDTO, events []ProjectEvent) ContinuityLibraryResult {
	assets := s.hydrateAssetsForList(root, index.Assets)
	profileDTOs := profileDTOsWithBindings(root, index, profiles)
	lineage := assetLineageDTOs(assets, profiles)
	if asset != nil {
		hydrated := s.hydrateAssetForList(root, *asset)
		asset = &hydrated
	}
	if profile != nil {
		key := profileKey(profile.Type, profile.ID)
		if record, ok := profiles[key]; ok {
			dto := profileDTOWithBindings(record.dto, index, profiles, assetMapByID(assets))
			profile = &dto
		}
	}
	return ContinuityLibraryResult{
		OK:        true,
		Asset:     asset,
		Assets:    assets,
		Profile:   profile,
		Profiles:  profileDTOs,
		Lineage:   lineage,
		Duplicate: duplicate,
		Health:    ptr(s.HealthReport(root)),
		Events:    events,
	}
}

func (s *Store) bindingFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) ContinuityLibraryResult {
	err := s.bindingOperationError(code, severity, retryable, correlationID, userMessage, technicalDetail, bindingRecoveryActions(code, technicalDetail))
	return ContinuityLibraryResult{
		OK:       false,
		Assets:   []AssetDTO{},
		Profiles: []ProfileDTO{},
		Lineage:  []AssetLineageDTO{},
		Error:    &err,
		Events:   []ProjectEvent{s.event("asset_binding.operation", "blocked", userMessage, correlationID, &err)},
	}
}

func (s *Store) bindingOperationError(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string, recoveryActions []string) OperationError {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, recoveryActions)
	err.TargetType = "asset_binding"
	return err
}

func bindingRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeBindingAssetMissing, CodeMainReferenceAssetMissing:
		return []string{"List the Asset Library and choose an existing Asset id before retrying."}
	case CodeBindingTargetMissing:
		return []string{"Choose an existing Character, Scene, or Prop profile target before retrying."}
	case CodeBindingDuplicate:
		return []string{"Choose reuse to keep the existing binding, or select a different purpose."}
	case CodeBindingPurposeReserved:
		return []string{"Use the dedicated main reference action to set or clear main_reference."}
	case CodeContinuityLockedBinding:
		return []string{"Unlock the locked binding with a recorded reason before replacing or clearing it."}
	case CodeBindingProfileInvalid:
		return []string{"Fix the profile JSON shape or restore the profile file from backup."}
	case CodeBindingWriteFailed:
		return []string{"Retry the binding operation. If it fails again, run project health check before continuing."}
	case CodeProjectRootRequired:
		return []string{detail}
	default:
		return []string{"Review the binding target and retry."}
	}
}

func (s *Store) loadProfileRecords(root string, manifest Manifest) (map[string]profileRecord, error) {
	records := map[string]profileRecord{}
	for _, targetType := range []string{BindingTargetCharacter, BindingTargetScene, BindingTargetProp} {
		dir := profileDirectory(manifest, targetType)
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			relative := path.Join(dir, entry.Name())
			record, err := readProfileRecordFile(filepath.Join(root, filepath.FromSlash(relative)), relative, targetType)
			if err != nil {
				return nil, err
			}
			records[record.dto.key()] = record
		}
	}
	return records, nil
}

func (s *Store) loadProfileRecord(root string, manifest Manifest, targetType string, targetID string) (profileRecord, error) {
	records, err := s.loadProfileRecords(root, manifest)
	if err != nil {
		return profileRecord{}, err
	}
	key := profileKey(targetType, strings.TrimSpace(targetID))
	if record, ok := records[key]; ok {
		return record, nil
	}
	return profileRecord{}, fmt.Errorf("%s:%s", targetType, strings.TrimSpace(targetID))
}

func readProfileRecordFile(filename string, relative string, targetType string) (profileRecord, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return profileRecord{}, err
	}
	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return profileRecord{}, err
	}
	id := firstString(raw, "id")
	if id == "" {
		id = strings.TrimSuffix(filepath.Base(relative), filepath.Ext(relative))
	}
	dto := ProfileDTO{
		ID:                   id,
		Type:                 targetType,
		Name:                 profileDisplayName(raw, id),
		Role:                 firstString(raw, "role"),
		Identity:             firstString(raw, "identity", "shortDescription"),
		VisualDescription:    firstString(raw, "visualDescription"),
		Costume:              firstString(raw, "costume"),
		Location:             firstString(raw, "location"),
		TimeOfDay:            firstString(raw, "timeOfDay"),
		Mood:                 firstString(raw, "mood"),
		Lighting:             firstString(raw, "lighting"),
		Category:             firstString(raw, "category"),
		Appearance:           firstString(raw, "appearance"),
		Usage:                firstString(raw, "usage"),
		RelativePath:         relative,
		ReferenceAssetIDs:    cleanStringList(firstStringSlice(raw, "referenceAssetIds")),
		MainReferenceAssetID: strings.TrimSpace(firstString(raw, "mainReferenceAssetId")),
		LockedRules:          cleanStringList(firstStringSlice(raw, "lockedRules")),
	}
	dto.ContinuityRules = parseContinuityRules(raw, targetType, dto.ID, dto.LockedRules)
	if dto.VisualDescription == "" {
		dto.VisualDescription = strings.Join(firstStringSlice(raw, "visualRules"), "; ")
	}
	switch targetType {
	case BindingTargetCharacter:
		if dto.Role == "" {
			dto.Role = "other"
		}
	case BindingTargetScene:
		if dto.Location == "" {
			dto.Location = dto.Name
		}
	case BindingTargetProp:
		if dto.Category == "" {
			dto.Category = "other"
		}
		if dto.Appearance == "" {
			dto.Appearance = strings.Join(firstStringSlice(raw, "visualRules"), "; ")
		}
	}
	if dto.LockedRules == nil {
		dto.LockedRules = []string{}
	}
	if dto.ContinuityRules == nil {
		dto.ContinuityRules = []ContinuityRuleDTO{}
	}
	if dto.ReferenceAssetIDs == nil {
		dto.ReferenceAssetIDs = []string{}
	}
	return profileRecord{dto: dto, raw: raw}, nil
}

func (s *Store) writeProfileAndIndex(root string, profile profileRecord, index AssetIndex) error {
	original, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(profile.dto.RelativePath)))
	if readErr != nil {
		return readErr
	}
	data, err := json.MarshalIndent(profile.raw, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	profilePath := filepath.Join(root, filepath.FromSlash(profile.dto.RelativePath))
	if err := s.atomicWrite(profilePath, data, 0o644, s.instanceID); err != nil {
		return err
	}
	if err := s.writeAssetIndex(root, index); err != nil {
		_ = s.atomicWrite(profilePath, original, 0o644, s.instanceID)
		return err
	}
	return nil
}

func profileRawWithBaseline(raw map[string]any, dto ProfileDTO) map[string]any {
	if raw == nil {
		raw = map[string]any{}
	}
	raw["id"] = dto.ID
	if _, ok := raw["name"]; !ok {
		raw["name"] = dto.Name
	}
	raw["referenceAssetIds"] = dto.ReferenceAssetIDs
	raw["mainReferenceAssetId"] = dto.MainReferenceAssetID
	raw["lockedRules"] = dto.LockedRules
	raw["continuityRules"] = dto.ContinuityRules
	raw["continuityRuleIds"] = continuityRuleIDs(dto.ContinuityRules)
	switch dto.Type {
	case BindingTargetCharacter:
		setStringIfMissing(raw, "role", dto.Role)
		setStringIfMissing(raw, "identity", dto.Identity)
		setStringIfMissing(raw, "visualDescription", dto.VisualDescription)
		setStringIfMissing(raw, "costume", dto.Costume)
	case BindingTargetScene:
		setStringIfMissing(raw, "location", dto.Location)
		setStringIfMissing(raw, "timeOfDay", dto.TimeOfDay)
		setStringIfMissing(raw, "mood", dto.Mood)
		setStringIfMissing(raw, "lighting", dto.Lighting)
	case BindingTargetProp:
		setStringIfMissing(raw, "category", dto.Category)
		setStringIfMissing(raw, "appearance", dto.Appearance)
		setStringIfMissing(raw, "usage", dto.Usage)
	}
	return raw
}

func profileDTOsWithBindings(root string, index AssetIndex, profiles map[string]profileRecord) []ProfileDTO {
	assets := assetMapByID(index.Assets)
	dtos := make([]ProfileDTO, 0, len(profiles))
	for _, record := range profiles {
		dtos = append(dtos, profileDTOWithBindings(record.dto, index, profiles, assets))
	}
	sort.Slice(dtos, func(i int, j int) bool {
		if dtos[i].Type == dtos[j].Type {
			return dtos[i].ID < dtos[j].ID
		}
		return dtos[i].Type < dtos[j].Type
	})
	for i := range dtos {
		if dtos[i].MainReferenceAssetID == "" {
			continue
		}
		if asset, ok := assets[dtos[i].MainReferenceAssetID]; ok {
			dtos[i].MainReferencePath = asset.RelativePath
			dtos[i].MainReferenceThumbnailPath = asset.ThumbnailPath
		} else {
			dtos[i].MissingMainReference = true
		}
	}
	_ = root
	return dtos
}

func profileDTOWithBindings(profile ProfileDTO, index AssetIndex, profiles map[string]profileRecord, assets map[string]AssetDTO) ProfileDTO {
	profile.Bindings = []AssetBindingDTO{}
	for _, asset := range index.Assets {
		for _, binding := range asset.Bindings {
			binding = normalizeAssetBinding(asset.ID, binding)
			if binding.TargetType == profile.Type && binding.TargetID == profile.ID {
				profile.Bindings = append(profile.Bindings, binding)
			}
		}
	}
	sortBindings(profile.Bindings)
	profile.BindingCount = len(profile.Bindings)
	if profile.MainReferenceAssetID != "" {
		if asset, ok := assets[profile.MainReferenceAssetID]; ok {
			profile.MainReferencePath = asset.RelativePath
			profile.MainReferenceThumbnailPath = asset.ThumbnailPath
		} else {
			profile.MissingMainReference = true
		}
	}
	_ = profiles
	return profile
}

func assetLineageDTOs(assets []AssetDTO, profiles map[string]profileRecord) []AssetLineageDTO {
	lineage := make([]AssetLineageDTO, 0, len(assets))
	for _, asset := range assets {
		bindings := make([]AssetBindingDTO, 0, len(asset.Bindings))
		summaries := make([]BindingTargetSummaryDTO, 0, len(asset.Bindings))
		for _, binding := range asset.Bindings {
			binding = normalizeAssetBinding(asset.ID, binding)
			bindings = append(bindings, binding)
			if profile, ok := profiles[profileKey(binding.TargetType, binding.TargetID)]; ok {
				summaries = append(summaries, BindingTargetSummaryDTO{
					TargetType:           binding.TargetType,
					TargetID:             binding.TargetID,
					Name:                 profile.dto.Name,
					RelativePath:         profile.dto.RelativePath,
					MainReferenceAssetID: profile.dto.MainReferenceAssetID,
				})
			}
		}
		sortBindings(bindings)
		sort.Slice(summaries, func(i int, j int) bool {
			if summaries[i].TargetType == summaries[j].TargetType {
				return summaries[i].TargetID < summaries[j].TargetID
			}
			return summaries[i].TargetType < summaries[j].TargetType
		})
		lineage = append(lineage, AssetLineageDTO{
			AssetID:         asset.ID,
			SourceKind:      asset.Source.Kind,
			SourceName:      asset.Source.OriginalName,
			RelativePath:    asset.RelativePath,
			Bindings:        bindings,
			TargetSummaries: summaries,
		})
	}
	sort.Slice(lineage, func(i int, j int) bool { return lineage[i].AssetID < lineage[j].AssetID })
	return lineage
}

func profileDirectory(manifest Manifest, targetType string) string {
	switch targetType {
	case BindingTargetCharacter:
		return manifest.Paths.Characters
	case BindingTargetScene:
		return manifest.Paths.Scenes
	case BindingTargetProp:
		return manifest.Paths.Props
	default:
		return ""
	}
}

func normalizeBindingTargetType(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case BindingTargetCharacter, "characters":
		return BindingTargetCharacter, true
	case BindingTargetScene, "scenes":
		return BindingTargetScene, true
	case BindingTargetProp, "props":
		return BindingTargetProp, true
	default:
		return "", false
	}
}

func normalizeBindingPurpose(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = assetTokenPattern.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return BindingPurposeReference
	}
	return value
}

func normalizeCreatedBy(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "user"
	}
	return value
}

func normalizeAssetBinding(assetID string, binding AssetBindingDTO) AssetBindingDTO {
	binding.AssetID = strings.TrimSpace(binding.AssetID)
	if binding.AssetID == "" {
		binding.AssetID = strings.TrimSpace(assetID)
	}
	binding.TargetType, _ = normalizeBindingTargetType(binding.TargetType)
	binding.TargetID = strings.TrimSpace(binding.TargetID)
	binding.Purpose = normalizeBindingPurpose(firstNonEmpty(binding.Purpose, binding.Role))
	if binding.ID == "" && binding.AssetID != "" && binding.TargetType != "" && binding.TargetID != "" {
		binding.ID = bindingID(binding.AssetID, binding.TargetType, binding.TargetID, binding.Purpose)
	}
	if binding.CreatedBy == "" {
		binding.CreatedBy = "user"
	}
	return binding
}

func findBinding(bindings []AssetBindingDTO, targetType string, targetID string, purpose string) *AssetBindingDTO {
	for index := range bindings {
		binding := normalizeAssetBinding(bindings[index].AssetID, bindings[index])
		if binding.TargetType == targetType && binding.TargetID == targetID && binding.Purpose == purpose {
			return &bindings[index]
		}
	}
	return nil
}

func removeMainReferenceBindings(assets []AssetDTO, targetType string, targetID string) {
	for assetIndex := range assets {
		bindings := assets[assetIndex].Bindings[:0]
		for _, binding := range assets[assetIndex].Bindings {
			binding = normalizeAssetBinding(assets[assetIndex].ID, binding)
			if binding.TargetType == targetType && binding.TargetID == targetID && binding.Purpose == BindingPurposeMainReference {
				continue
			}
			bindings = append(bindings, binding)
		}
		assets[assetIndex].Bindings = bindings
	}
}

func lockedMainReferenceBinding(assets []AssetDTO, targetType string, targetID string) *AssetBindingDTO {
	for assetIndex := range assets {
		for bindingIndex := range assets[assetIndex].Bindings {
			binding := normalizeAssetBinding(assets[assetIndex].ID, assets[assetIndex].Bindings[bindingIndex])
			if binding.TargetType == targetType && binding.TargetID == targetID && binding.Purpose == BindingPurposeMainReference && binding.Locked {
				return &binding
			}
		}
	}
	return nil
}

func bindingID(assetID string, targetType string, targetID string, purpose string) string {
	return "binding_" + safeFileToken(assetID+"_"+targetType+"_"+targetID+"_"+purpose)
}

func (profile ProfileDTO) key() string {
	return profileKey(profile.Type, profile.ID)
}

func profileKey(targetType string, targetID string) string {
	return strings.ToLower(strings.TrimSpace(targetType)) + ":" + strings.TrimSpace(targetID)
}

func findAssetIndexByID(assets []AssetDTO, assetID string) int {
	assetID = strings.TrimSpace(assetID)
	for i := range assets {
		if assets[i].ID == assetID {
			return i
		}
	}
	return -1
}

func assetMapByID(assets []AssetDTO) map[string]AssetDTO {
	byID := make(map[string]AssetDTO, len(assets))
	for _, asset := range assets {
		byID[asset.ID] = asset
	}
	return byID
}

func sortBindings(bindings []AssetBindingDTO) {
	sort.SliceStable(bindings, func(i int, j int) bool {
		if bindings[i].TargetType == bindings[j].TargetType {
			if bindings[i].TargetID == bindings[j].TargetID {
				return bindings[i].Purpose < bindings[j].Purpose
			}
			return bindings[i].TargetID < bindings[j].TargetID
		}
		return bindings[i].TargetType < bindings[j].TargetType
	})
}

func profileDisplayName(raw map[string]any, fallback string) string {
	value := firstString(raw, "name", "displayName", "title")
	if value == "" {
		return fallback
	}
	return value
}

func firstString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return strings.TrimSpace(typed)
				}
			}
		}
	}
	return ""
}

func firstStringSlice(raw map[string]any, keys ...string) []string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case []string:
			if len(typed) > 0 {
				return typed
			}
		case []any:
			values := make([]string, 0, len(typed))
			for _, item := range typed {
				if text, ok := item.(string); ok {
					values = append(values, text)
				}
			}
			if len(values) > 0 {
				return values
			}
		}
	}
	return []string{}
}

func setStringIfMissing(raw map[string]any, key string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if _, ok := raw[key]; !ok {
		raw[key] = value
	}
}

func addUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return cleanStringList(values)
	}
	values = cleanStringList(values)
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func mustProfileRecords(records map[string]profileRecord, err error) map[string]profileRecord {
	if err != nil {
		return map[string]profileRecord{}
	}
	return records
}
