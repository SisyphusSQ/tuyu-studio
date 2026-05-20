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
	CodeShotInvalidID            = "shot_invalid_id"
	CodeShotMissing              = "shot_missing"
	CodeShotLoadFailed           = "shot_load_failed"
	CodeShotSaveFailed           = "shot_save_failed"
	CodeShotIdentityMismatch     = "shot_identity_mismatch"
	CodeShotGraphNodeMissing     = "shot_graph_node_missing"
	CodeShotRequiredFieldMissing = "shot_required_field_missing"
	CodeShotReferenceMissing     = "shot_reference_missing"
	CodeShotContextConflict      = "shot_context_conflict"
	CodeShotIllegalTransition    = "shot_illegal_transition"
)

const (
	ShotStatusContextReady = "context_ready"
	ShotStatusContextDirty = "context_dirty"
)

type ValidateShotContextCommand struct {
	Root          string `json:"root"`
	ShotID        string `json:"shotId"`
	CorrelationID string `json:"correlationId"`
}

type PromoteShotContextCommand struct {
	Root          string `json:"root"`
	ShotID        string `json:"shotId"`
	CorrelationID string `json:"correlationId"`
}

type MarkShotContextDirtyCommand struct {
	Root          string `json:"root"`
	ShotID        string `json:"shotId"`
	Reason        string `json:"reason,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type ShotContextResult struct {
	OK     bool                 `json:"ok"`
	Shot   *ShotCardDTO         `json:"shot,omitempty"`
	Report ShotContextReportDTO `json:"report"`
	Error  *OperationError      `json:"error,omitempty"`
	Events []ProjectEvent       `json:"events"`
}

type ShotContextReportDTO struct {
	ShotID               string                `json:"shotId"`
	Status               string                `json:"status"`
	CanEnterContextReady bool                  `json:"canEnterContextReady"`
	MissingFields        []string              `json:"missingFields"`
	Blocking             []ShotContextIssueDTO `json:"blocking"`
	Warnings             []ShotContextIssueDTO `json:"warnings"`
	References           []ShotReferenceDTO    `json:"references"`
	CheckedAt            string                `json:"checkedAt"`
}

type ShotContextIssueDTO struct {
	Code            string   `json:"code"`
	Severity        string   `json:"severity"`
	Field           string   `json:"field"`
	ReferenceID     string   `json:"referenceId,omitempty"`
	UserMessage     string   `json:"userMessage"`
	RecoveryActions []string `json:"recoveryActions"`
}

type ShotReferenceDTO struct {
	Field       string `json:"field"`
	Kind        string `json:"kind"`
	ReferenceID string `json:"referenceId"`
	Status      string `json:"status"`
	Path        string `json:"path,omitempty"`
}

type shotContextRecord struct {
	raw  map[string]any
	shot ShotCardDTO
}

type referenceAssetEntry struct {
	id       string
	path     string
	declared bool
}

func (s *Store) ValidateShotContext(command ValidateShotContextCommand) ShotContextResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.shotContextFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before checking Shot context.", ShotContextReportDTO{})
	}

	manifest, errResult := s.readShotContextManifest(root, correlationID, "checking Shot context")
	if errResult != nil {
		return *errResult
	}
	record, errResult := s.readShotContextRecord(root, manifest, command.ShotID, correlationID)
	if errResult != nil {
		return *errResult
	}
	report := s.validateShotContext(root, manifest, record.shot)
	state := "completed"
	if !report.CanEnterContextReady {
		state = "blocked"
	}
	return ShotContextResult{
		OK:     true,
		Shot:   &record.shot,
		Report: report,
		Events: []ProjectEvent{s.event("shot.context_checked", state, "Shot context validation completed.", correlationID, nil)},
	}
}

func (s *Store) PromoteShotContext(command PromoteShotContextCommand) ShotContextResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.shotContextFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before promoting Shot context.", ShotContextReportDTO{})
	}
	if failure := s.shotContextWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, errResult := s.readShotContextManifest(root, correlationID, "promoting Shot context")
	if errResult != nil {
		return *errResult
	}
	record, errResult := s.readShotContextRecord(root, manifest, command.ShotID, correlationID)
	if errResult != nil {
		return *errResult
	}
	report := s.validateShotContext(root, manifest, record.shot)
	if !report.CanEnterContextReady {
		issue := firstShotContextIssue(report)
		return s.shotContextFailure(issue.Code, issue.Severity, false, correlationID, "Shot cannot enter context_ready.", issue.UserMessage, report)
	}
	if isShotAlreadyContextReady(record.shot.Status) {
		return ShotContextResult{
			OK:     true,
			Shot:   &record.shot,
			Report: report,
			Events: []ProjectEvent{s.event("shot.context_ready", "completed", "Shot is already context_ready.", correlationID, nil)},
		}
	}
	if !canPromoteShotContextStatus(record.shot.Status) {
		return s.shotContextFailure(
			CodeShotIllegalTransition,
			SeverityBlocking,
			false,
			correlationID,
			"Shot status cannot enter context_ready from its current state.",
			"current status: "+strings.TrimSpace(record.shot.Status),
			report,
		)
	}
	if !manifestHasShotNode(root, manifest, record.shot.ID) {
		return s.shotContextFailure(CodeShotGraphNodeMissing, SeverityBlocking, false, correlationID, "Shot graph node is missing.", "shots/"+record.shot.ID+".json", report)
	}

	now := s.now().UTC()
	shot := record.shot
	shot.Status = ShotStatusContextReady
	shot.UpdatedAt = now.Format(time.RFC3339)
	patchShotRaw(record.raw, shot)
	if err := s.writeShotRaw(root, manifest, shot.ID, record.raw); err != nil {
		return s.shotContextFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, "Shot context_ready state could not be saved.", err.Error(), report)
	}
	if err := s.updateManifestShotStatus(root, manifest, shot.ID, ShotStatusContextReady, now); err != nil {
		return s.shotContextFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, "Shot graph status could not be saved.", err.Error(), report)
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("shot.context_ready", correlationID),
		EventType:     "shot.context_ready",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Shot promoted to context_ready.",
		Details: map[string]any{
			"shotId": shot.ID,
		},
	})
	report = s.validateShotContext(root, manifest, shot)
	report.Status = ShotStatusContextReady
	return ShotContextResult{
		OK:     true,
		Shot:   &shot,
		Report: report,
		Events: []ProjectEvent{s.event("shot.context_ready", "completed", "Shot promoted to context_ready.", correlationID, nil)},
	}
}

func (s *Store) MarkShotContextDirty(command MarkShotContextDirtyCommand) ShotContextResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.shotContextFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before marking Shot context dirty.", ShotContextReportDTO{})
	}
	if failure := s.shotContextWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, errResult := s.readShotContextManifest(root, correlationID, "marking Shot context dirty")
	if errResult != nil {
		return *errResult
	}
	record, errResult := s.readShotContextRecord(root, manifest, command.ShotID, correlationID)
	if errResult != nil {
		return *errResult
	}
	if !manifestHasShotNode(root, manifest, record.shot.ID) {
		return s.shotContextFailure(CodeShotGraphNodeMissing, SeverityBlocking, false, correlationID, "Shot graph node is missing.", "shots/"+record.shot.ID+".json", ShotContextReportDTO{})
	}
	now := s.now().UTC()
	shot := record.shot
	shot.Status = ShotStatusContextDirty
	shot.UpdatedAt = now.Format(time.RFC3339)
	patchShotRaw(record.raw, shot)
	if err := s.writeShotRaw(root, manifest, shot.ID, record.raw); err != nil {
		return s.shotContextFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, "Shot context_dirty state could not be saved.", err.Error(), ShotContextReportDTO{})
	}
	if err := s.updateManifestShotStatus(root, manifest, shot.ID, ShotStatusContextDirty, now); err != nil {
		return s.shotContextFailure(CodeShotSaveFailed, SeverityBlocking, true, correlationID, "Shot graph dirty status could not be saved.", err.Error(), ShotContextReportDTO{})
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("shot.context_dirty", correlationID),
		EventType:     "shot.context_dirty",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Shot marked context_dirty.",
		Details: map[string]any{
			"shotId": shot.ID,
			"reason": strings.TrimSpace(command.Reason),
		},
	})
	report := s.validateShotContext(root, manifest, shot)
	report.Status = ShotStatusContextDirty
	return ShotContextResult{
		OK:     true,
		Shot:   &shot,
		Report: report,
		Events: []ProjectEvent{s.event("shot.context_dirty", "completed", "Shot marked context_dirty after revision.", correlationID, nil)},
	}
}

func (s *Store) validateShotContext(root string, manifest Manifest, shot ShotCardDTO) ShotContextReportDTO {
	report := ShotContextReportDTO{
		ShotID:     shot.ID,
		Status:     strings.TrimSpace(shot.Status),
		CheckedAt:  s.now().UTC().Format(time.RFC3339),
		Blocking:   []ShotContextIssueDTO{},
		Warnings:   []ShotContextIssueDTO{},
		References: []ShotReferenceDTO{},
	}
	for _, field := range missingShotRequiredFields(shot) {
		report.MissingFields = append(report.MissingFields, field)
		report.Blocking = append(report.Blocking, shotContextIssue(
			CodeShotRequiredFieldMissing,
			SeverityBlocking,
			field,
			"",
			"Shot required field is missing or still uses a placeholder.",
			"Fill "+field+" before promoting the Shot to context_ready.",
		))
	}
	if hasCandidateSource(shot) && (strings.TrimSpace(shot.ScriptSceneID) == "" || !validPositiveSourceRange(shot.SourceRange)) {
		report.Blocking = append(report.Blocking, shotContextIssue(
			CodeShotExpansionSourceMissing,
			SeverityBlocking,
			"sourceRange",
			shot.SourceCandidateID,
			"Candidate-created Shot is missing ScriptScene or sourceRange lineage.",
			"Restore sourceCandidateID, scriptSceneID, and sourceRange before promoting the Shot.",
		))
	}
	report = s.validateShotReferences(root, manifest, shot, report)
	report.CanEnterContextReady = len(report.Blocking) == 0
	sort.Strings(report.MissingFields)
	return report
}

func (s *Store) validateShotReferences(root string, manifest Manifest, shot ShotCardDTO, report ShotContextReportDTO) ShotContextReportDTO {
	for _, id := range cleanStringList(shot.CharacterIDs) {
		ref := s.referenceCheck(root, manifest.Paths.Characters, "characterIds", "character", id)
		report.References = append(report.References, ref)
		if ref.Status != "resolved" {
			report.Blocking = append(report.Blocking, shotContextIssue(CodeShotReferenceMissing, SeverityBlocking, "characterIds", id, "Shot character reference cannot be resolved.", "Create or relink the Character profile before promoting context."))
		}
	}
	if sceneID := shotSceneReferenceID(shot); sceneID != "" {
		ref := s.referenceCheck(root, manifest.Paths.Scenes, "sceneProfileId", "scene", sceneID)
		report.References = append(report.References, ref)
		if ref.Status != "resolved" {
			report.Blocking = append(report.Blocking, shotContextIssue(CodeShotReferenceMissing, SeverityBlocking, "sceneProfileId", sceneID, "Shot scene reference cannot be resolved.", "Create or relink the Scene profile before promoting context."))
		}
	}
	for _, id := range cleanStringList(shot.PropIDs) {
		ref := s.referenceCheck(root, manifest.Paths.Props, "propIds", "prop", id)
		report.References = append(report.References, ref)
		if ref.Status != "resolved" {
			report.Blocking = append(report.Blocking, shotContextIssue(CodeShotReferenceMissing, SeverityBlocking, "propIds", id, "Shot prop reference cannot be resolved.", "Create or relink the Prop profile before promoting context."))
		}
	}
	assets := s.referenceAssetIndex(root, manifest)
	for _, id := range cleanStringList(shot.ReferenceAssetIDs) {
		ref := s.referenceAssetCheck(root, assets, id)
		report.References = append(report.References, ref)
		if ref.Status != "resolved" {
			report.Blocking = append(report.Blocking, shotContextIssue(CodeShotReferenceMissing, SeverityBlocking, "referenceAssetIds", id, "Shot reference asset cannot be resolved.", "Restore or relink the project reference asset before promoting context."))
		}
	}
	for _, characterRef := range shot.CharacterRefs {
		id := strings.TrimSpace(characterRef.ReferenceAssetID)
		if id == "" {
			continue
		}
		ref := s.referenceAssetCheck(root, assets, id)
		ref.Field = "characterRefs.referenceAssetId"
		report.References = append(report.References, ref)
		if ref.Status != "resolved" {
			report.Blocking = append(report.Blocking, shotContextIssue(CodeShotReferenceMissing, SeverityBlocking, "characterRefs.referenceAssetId", id, "Shot character reference asset cannot be resolved.", "Restore or relink the character reference asset before promoting context."))
		}
	}
	return report
}

func (s *Store) referenceCheck(root string, directory string, field string, kind string, id string) ShotReferenceDTO {
	relative, ok := s.projectObjectPathByID(root, directory, id)
	if ok {
		return ShotReferenceDTO{Field: field, Kind: kind, ReferenceID: id, Status: "resolved", Path: relative}
	}
	return ShotReferenceDTO{Field: field, Kind: kind, ReferenceID: id, Status: "missing"}
}

func (s *Store) referenceAssetCheck(root string, assets map[string]referenceAssetEntry, id string) ShotReferenceDTO {
	if entry, ok := assets[id]; ok {
		status := "declared_missing"
		if safePath, ok := safeProjectFilePath(root, entry.path); ok && fileExists(safePath) {
			status = "resolved"
		} else if strings.TrimSpace(entry.path) != "" && len(validateProjectPath("referenceAsset.path", entry.path)) > 0 {
			status = "invalid_path"
		}
		return ShotReferenceDTO{Field: "referenceAssetIds", Kind: "asset", ReferenceID: id, Status: status, Path: entry.path}
	}
	if safePath, ok := safeProjectFilePath(root, id); ok && fileExists(safePath) {
		return ShotReferenceDTO{Field: "referenceAssetIds", Kind: "asset", ReferenceID: id, Status: "resolved", Path: id}
	}
	return ShotReferenceDTO{Field: "referenceAssetIds", Kind: "asset", ReferenceID: id, Status: "missing"}
}

func (s *Store) projectObjectPathByID(root string, directory string, id string) (string, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", false
	}
	dirRelative := path.Clean(strings.ReplaceAll(directory, "\\", "/"))
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dirRelative)))
	if err != nil {
		return "", false
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		relative := path.Join(dirRelative, entry.Name())
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			continue
		}
		var value struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(data, &value) == nil && strings.TrimSpace(value.ID) == id {
			return relative, true
		}
	}
	return "", false
}

func (s *Store) referenceAssetIndex(root string, manifest Manifest) map[string]referenceAssetEntry {
	assets := map[string]referenceAssetEntry{}
	for _, directory := range []string{manifest.Paths.Characters, manifest.Paths.Scenes, manifest.Paths.Props} {
		s.collectReferenceAssets(root, directory, assets)
	}
	if index, err := s.loadAssetIndex(root, manifest.Project.ID); err == nil {
		for _, asset := range index.Assets {
			id := strings.TrimSpace(asset.ID)
			if id == "" {
				continue
			}
			assets[id] = referenceAssetEntry{
				id:       id,
				path:     cleanProjectRelativePath(asset.RelativePath),
				declared: true,
			}
		}
	}
	return assets
}

func (s *Store) collectReferenceAssets(root string, directory string, assets map[string]referenceAssetEntry) {
	dirRelative := path.Clean(strings.ReplaceAll(directory, "\\", "/"))
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dirRelative)))
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path.Join(dirRelative, entry.Name()))))
		if err != nil {
			continue
		}
		var value struct {
			ReferenceAssetIDs []string `json:"referenceAssetIds"`
			ReferencePaths    []string `json:"referencePaths"`
		}
		if json.Unmarshal(data, &value) != nil {
			continue
		}
		for index, id := range cleanStringList(value.ReferenceAssetIDs) {
			asset := referenceAssetEntry{id: id, declared: true}
			if index < len(value.ReferencePaths) {
				asset.path = cleanProjectRelativePath(value.ReferencePaths[index])
			}
			assets[id] = asset
		}
	}
}

func (s *Store) readShotContextManifest(root string, correlationID string, action string) (Manifest, *ShotContextResult) {
	manifest, report, err := s.readManifest(root)
	if err == nil {
		return manifest, nil
	}
	operationError := s.errorFromReadFailure(report, err, correlationID)
	return Manifest{}, &ShotContextResult{
		OK:     false,
		Error:  &operationError,
		Events: []ProjectEvent{s.event("shot.context", "blocked", "Project manifest could not be loaded before "+action+".", correlationID, &operationError)},
	}
}

func (s *Store) readShotContextRecord(root string, manifest Manifest, shotIDValue string, correlationID string) (shotContextRecord, *ShotContextResult) {
	shotID, err := normalizeExistingShotID(shotIDValue)
	if err != nil {
		result := s.shotContextFailure(CodeShotInvalidID, SeverityBlocking, false, correlationID, "Shot id is invalid.", err.Error(), ShotContextReportDTO{})
		return shotContextRecord{}, &result
	}
	filename := shotPath(root, manifest, shotID)
	data, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		result := s.shotContextFailure(CodeShotMissing, SeverityBlocking, false, correlationID, "Shot file is missing.", "shots/"+shotID+".json", ShotContextReportDTO{})
		return shotContextRecord{}, &result
	}
	if err != nil {
		result := s.shotContextFailure(CodeShotLoadFailed, SeverityBlocking, true, correlationID, "Shot file could not be read.", err.Error(), ShotContextReportDTO{})
		return shotContextRecord{}, &result
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		result := s.shotContextFailure(CodeShotLoadFailed, SeverityBlocking, false, correlationID, "Shot file is invalid JSON.", err.Error(), ShotContextReportDTO{})
		return shotContextRecord{}, &result
	}
	var shot ShotCardDTO
	if err := json.Unmarshal(data, &shot); err != nil {
		result := s.shotContextFailure(CodeShotLoadFailed, SeverityBlocking, false, correlationID, "Shot file shape is invalid.", err.Error(), ShotContextReportDTO{})
		return shotContextRecord{}, &result
	}
	if strings.TrimSpace(shot.ID) == "" {
		shot.ID = shotID
	} else if strings.TrimSpace(shot.ID) != shotID {
		result := s.shotContextFailure(CodeShotIdentityMismatch, SeverityBlocking, false, correlationID, "Shot file id does not match the requested Shot id.", "requested "+shotID+", found "+strings.TrimSpace(shot.ID), ShotContextReportDTO{})
		return shotContextRecord{}, &result
	}
	shot.ID = shotID
	if strings.TrimSpace(shot.ProjectID) == "" {
		shot.ProjectID = manifest.Project.ID
	}
	shot.CharacterIDs = cleanStringList(shot.CharacterIDs)
	shot.SceneIDRefs = cleanStringList(shot.SceneIDRefs)
	shot.PropIDs = cleanStringList(shot.PropIDs)
	shot.ReferenceAssetIDs = cleanStringList(shot.ReferenceAssetIDs)
	shot.PackageIDs = cleanStringList(shot.PackageIDs)
	shot.ResultIDs = cleanStringList(shot.ResultIDs)
	shot.ContinuityRuleIDs = cleanStringList(shot.ContinuityRuleIDs)
	return shotContextRecord{raw: raw, shot: shot}, nil
}

func (s *Store) writeShotRaw(root string, manifest Manifest, shotID string, raw map[string]any) error {
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	filename := shotPath(root, manifest, shotID)
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return s.atomicWrite(filename, data, 0o644, s.instanceID)
}

func (s *Store) updateManifestShotStatus(root string, manifest Manifest, shotID string, status string, now time.Time) error {
	shotNodeID := manifestShotNodeID(root, manifest, shotID)
	updated := false
	for index := range manifest.Graph.Nodes {
		node := &manifest.Graph.Nodes[index]
		if strings.TrimSpace(node.Kind) == "shot" && strings.TrimSpace(node.ID) == shotNodeID {
			node.Status = status
			updated = true
		}
	}
	packageUpdated, err := s.markPackagesStaleForShots(root, &manifest, []string{shotID}, now)
	if err != nil {
		return err
	}
	if !updated && !packageUpdated {
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

func (s *Store) shotContextWriteLockFailure(root string, correlationID string) *ShotContextResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.shotContextFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before editing Shot context.", ShotContextReportDTO{})
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.shotContextFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before editing Shot context.", "Open the project with takeover, then retry.", ShotContextReportDTO{})
		return &result
	}
	return nil
}

func (s *Store) shotContextFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string, report ShotContextReportDTO) ShotContextResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, shotContextRecoveryActions(code, technicalDetail))
	return ShotContextResult{
		OK:     false,
		Report: report,
		Error:  &err,
		Events: []ProjectEvent{s.event("shot.context", "blocked", userMessage, correlationID, &err)},
	}
}

func shotContextRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeShotRequiredFieldMissing:
		return []string{"Fill the missing Shot fields before promoting context."}
	case CodeShotExpansionSourceMissing:
		return []string{"Restore the candidate source, ScriptScene id, and source range before promoting context."}
	case CodeShotReferenceMissing:
		return []string{"Restore or relink the missing Shot reference before promoting context."}
	case CodeShotContextConflict:
		return []string{"Resolve blocking continuity conflicts before promoting context."}
	case CodeShotIllegalTransition:
		return []string{"Only draft or context_dirty Shots can be promoted to context_ready in this workflow."}
	case CodeShotIdentityMismatch:
		return []string{"Fix the Shot file id so it matches the selected Shot filename before retrying."}
	case CodeShotGraphNodeMissing:
		return []string{"Restore the Shot graph node before changing Shot context state."}
	default:
		if strings.TrimSpace(detail) == "" {
			return []string{"Fix the Shot context input and retry."}
		}
		return []string{detail}
	}
}

func missingShotRequiredFields(shot ShotCardDTO) []string {
	required := map[string]string{
		"title":          shot.Title,
		"description":    shot.Description,
		"aspectRatio":    shot.AspectRatio,
		"shotType":       shot.ShotType,
		"cameraMovement": shot.CameraMovement,
		"action":         shot.Action,
		"emotion":        shot.Emotion,
	}
	var missing []string
	for field, value := range required {
		if isMissingShotValue(value) {
			missing = append(missing, field)
		}
	}
	if shot.DurationSeconds <= 0 {
		missing = append(missing, "durationSeconds")
	}
	if len(cleanStringList(shot.CharacterIDs)) == 0 && shotSceneReferenceID(shot) == "" && strings.TrimSpace(shot.EmptySceneReason) == "" {
		missing = append(missing, "context")
	}
	sort.Strings(missing)
	return missing
}

func isMissingShotValue(value string) bool {
	value = strings.TrimSpace(value)
	switch value {
	case "", "unspecified", "unknown", "todo", "tbd":
		return true
	default:
		return false
	}
}

func hasCandidateSource(shot ShotCardDTO) bool {
	return strings.TrimSpace(shot.SourceCandidateID) != ""
}

func validPositiveSourceRange(sourceRange ScriptSourceRange) bool {
	return sourceRange.StartLine > 0 && sourceRange.EndLine >= sourceRange.StartLine
}

func shotSceneReferenceID(shot ShotCardDTO) string {
	for _, value := range []string{shot.SceneProfileID, shot.SceneID} {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	if len(shot.SceneIDRefs) > 0 {
		return strings.TrimSpace(shot.SceneIDRefs[0])
	}
	return ""
}

func firstShotContextIssue(report ShotContextReportDTO) ShotContextIssueDTO {
	if len(report.Blocking) > 0 {
		return report.Blocking[0]
	}
	return shotContextIssue(CodeShotContextConflict, SeverityBlocking, "", "", "Shot context is not ready.", "Fix blocking validation issues before retrying.")
}

func isShotAlreadyContextReady(status string) bool {
	return strings.TrimSpace(status) == ShotStatusContextReady
}

func canPromoteShotContextStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "", ShotStatusDraft, ShotStatusContextDirty:
		return true
	default:
		return false
	}
}

func manifestHasShotNode(root string, manifest Manifest, shotID string) bool {
	return manifestShotNodeID(root, manifest, shotID) != ""
}

func shotContextIssue(code string, severity string, field string, referenceID string, userMessage string, recoveryAction string) ShotContextIssueDTO {
	return ShotContextIssueDTO{
		Code:            code,
		Severity:        severity,
		Field:           field,
		ReferenceID:     strings.TrimSpace(referenceID),
		UserMessage:     userMessage,
		RecoveryActions: []string{recoveryAction},
	}
}

func normalizeExistingShotID(shotID string) (string, error) {
	shotID = strings.TrimSpace(shotID)
	if shotID == "" {
		return "", errors.New("shot id is required")
	}
	if !scriptIDPattern.MatchString(shotID) {
		return "", fmt.Errorf("invalid shot id %q", shotID)
	}
	return shotID, nil
}

func shotPath(root string, manifest Manifest, shotID string) string {
	if relative := shotRelativePathByID(root, manifest, shotID); relative != "" {
		return filepath.Join(root, filepath.FromSlash(relative))
	}
	return filepath.Join(root, filepath.FromSlash(path.Join(manifest.Paths.Shots, shotID+".json")))
}

func shotRelativePathByID(root string, manifest Manifest, shotID string) string {
	shotID = strings.TrimSpace(shotID)
	if shotID == "" {
		return ""
	}
	dirRelative := path.Clean(strings.ReplaceAll(manifest.Paths.Shots, "\\", "/"))
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dirRelative)))
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		relative := path.Join(dirRelative, entry.Name())
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			continue
		}
		var value struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(data, &value) == nil && strings.TrimSpace(value.ID) == shotID {
			return relative
		}
	}
	return ""
}

func patchShotRaw(raw map[string]any, shot ShotCardDTO) {
	raw["id"] = shot.ID
	raw["projectId"] = shot.ProjectID
	raw["status"] = shot.Status
	raw["updatedAt"] = shot.UpdatedAt
	if strings.TrimSpace(shot.CreatedAt) != "" {
		raw["createdAt"] = shot.CreatedAt
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}
