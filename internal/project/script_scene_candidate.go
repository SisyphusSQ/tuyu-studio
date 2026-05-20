package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	CodeSourceRangeInvalid         = "source_range_invalid"
	CodeSceneRequiredFieldMissing  = "scene_required_field_missing"
	CodeScriptExpansionRowInvalid  = "script_expansion_row_invalid"
	CodeShotExpansionSourceMissing = "shot_expansion_source_missing"
	CodeShotCandidateMissing       = "shot_candidate_missing"
	CodeShotCandidateConflict      = "shot_candidate_conflict"
	CodeShotCandidateSaveFailed    = "shot_candidate_save_failed"
	CodeShotConfirmFailed          = "shot_confirm_failed"
)

const ShotCandidatesRelativeDir = "assets/inputs/shot-candidates"

const (
	ShotCandidateStatusCandidate = "candidate"
	ShotCandidateStatusAccepted  = "accepted"
	ShotCandidateStatusRejected  = "rejected"
	ShotStatusDraft              = "draft"
)

type ConfirmScriptSceneCommand struct {
	Root          string            `json:"root"`
	ScriptID      string            `json:"scriptId,omitempty"`
	SceneID       string            `json:"sceneId,omitempty"`
	Title         string            `json:"title"`
	Location      string            `json:"location"`
	TimeOfDay     string            `json:"timeOfDay,omitempty"`
	Characters    []string          `json:"characters"`
	Props         []string          `json:"props"`
	Action        string            `json:"action"`
	Dialogue      []DialogueLineDTO `json:"dialogue"`
	EmotionalBeat string            `json:"emotionalBeat,omitempty"`
	SourceRange   ScriptSourceRange `json:"sourceRange"`
	AllowOverlap  bool              `json:"allowOverlap"`
	CorrelationID string            `json:"correlationId"`
}

type SaveShotCandidateCommand struct {
	Root              string                `json:"root"`
	ScriptID          string                `json:"scriptId,omitempty"`
	CandidateID       string                `json:"candidateId,omitempty"`
	ScriptSceneID     string                `json:"scriptSceneId"`
	Index             int                   `json:"index"`
	DurationSeconds   int                   `json:"durationSeconds"`
	VisualDescription string                `json:"visualDescription"`
	CharacterRefs     []ShotCharacterRefDTO `json:"characterRefs"`
	SourceRange       ScriptSourceRange     `json:"sourceRange"`
	CorrelationID     string                `json:"correlationId"`
}

type ListShotCandidatesCommand struct {
	Root          string `json:"root"`
	ScriptID      string `json:"scriptId,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type ConfirmShotCandidateCommand struct {
	Root          string `json:"root"`
	ScriptID      string `json:"scriptId,omitempty"`
	CandidateID   string `json:"candidateId"`
	ShotID        string `json:"shotId,omitempty"`
	ConfirmedBy   string `json:"confirmedBy,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type RejectShotCandidateCommand struct {
	Root            string `json:"root"`
	ScriptID        string `json:"scriptId,omitempty"`
	CandidateID     string `json:"candidateId"`
	RejectionReason string `json:"rejectionReason,omitempty"`
	CorrelationID   string `json:"correlationId"`
}

type ScriptSceneCandidateResult struct {
	OK         bool               `json:"ok"`
	Document   *ScriptDocumentDTO `json:"document,omitempty"`
	Candidate  *ShotCandidateDTO  `json:"candidate,omitempty"`
	Candidates []ShotCandidateDTO `json:"candidates"`
	Shot       *ShotCardDTO       `json:"shot,omitempty"`
	Error      *OperationError    `json:"error,omitempty"`
	Events     []ProjectEvent     `json:"events"`
}

type ShotCharacterRefDTO struct {
	CharacterID      string `json:"characterId,omitempty"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	ReferenceAssetID string `json:"referenceAssetId,omitempty"`
}

type ShotCandidateDTO struct {
	ID                string                `json:"id"`
	ProjectID         string                `json:"projectId"`
	ScriptID          string                `json:"scriptId"`
	ScriptSceneID     string                `json:"scriptSceneId"`
	Index             int                   `json:"index"`
	DurationSeconds   int                   `json:"durationSeconds"`
	VisualDescription string                `json:"visualDescription"`
	CharacterRefs     []ShotCharacterRefDTO `json:"characterRefs"`
	SourceRange       ScriptSourceRange     `json:"sourceRange"`
	Status            string                `json:"status"`
	ShotID            string                `json:"shotId,omitempty"`
	RejectionReason   string                `json:"rejectionReason,omitempty"`
	CreatedAt         string                `json:"createdAt"`
	UpdatedAt         string                `json:"updatedAt"`
}

type ShotCardDTO struct {
	ID                string                `json:"id"`
	ProjectID         string                `json:"projectId"`
	SceneID           string                `json:"sceneId"`
	SourceCandidateID string                `json:"sourceCandidateId"`
	ScriptSceneID     string                `json:"scriptSceneId"`
	Index             int                   `json:"index"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	DurationSeconds   int                   `json:"durationSeconds"`
	AspectRatio       string                `json:"aspectRatio"`
	ShotType          string                `json:"shotType"`
	CameraMovement    string                `json:"cameraMovement"`
	Action            string                `json:"action"`
	Emotion           string                `json:"emotion"`
	CharacterIDs      []string              `json:"characterIds"`
	CharacterRefs     []ShotCharacterRefDTO `json:"characterRefs"`
	SourceRange       ScriptSourceRange     `json:"sourceRange"`
	Status            string                `json:"status"`
	ConfirmedBy       string                `json:"confirmedBy"`
	ConfirmedAt       string                `json:"confirmedAt"`
	OverwrittenFields []string              `json:"overwrittenFields"`
	CreatedAt         string                `json:"createdAt"`
	UpdatedAt         string                `json:"updatedAt"`
}

type shotCandidateFile struct {
	ProjectID  string             `json:"projectId"`
	ScriptID   string             `json:"scriptId"`
	Candidates []ShotCandidateDTO `json:"candidates"`
	UpdatedAt  string             `json:"updatedAt"`
}

func (s *Store) ConfirmScriptScene(command ConfirmScriptSceneCommand) ScriptSceneCandidateResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptSceneCandidateFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before confirming a scene.")
	}
	if failure := s.scriptSceneWriteLockFailure(root, correlationID, "scene"); failure != nil {
		return *failure
	}

	manifest, errResult := s.readSceneCandidateManifest(root, correlationID, "confirming a scene")
	if errResult != nil {
		return *errResult
	}

	scriptID, idErr := normalizeScriptDocumentID(command.ScriptID)
	if idErr != nil {
		return s.scriptSceneCandidateFailure(CodeScriptInvalidID, SeverityBlocking, false, correlationID, "Script document id is invalid.", idErr.Error())
	}
	document, err := s.readScriptDocument(root, scriptID)
	if err != nil {
		return s.scriptSceneCandidateFailure(CodeScriptLoadFailed, SeverityBlocking, true, correlationID, "ScriptDocument could not be loaded.", err.Error())
	}
	if missing := missingSceneFields(command.Title, command.Location, command.Action); len(missing) > 0 {
		return s.scriptSceneCandidateFailure(CodeSceneRequiredFieldMissing, SeverityBlocking, false, correlationID, "Scene required fields are missing.", strings.Join(missing, ", "))
	}
	if err := validateSceneSourceRange(document, command.SourceRange, command.AllowOverlap, command.SceneID); err != nil {
		return s.scriptSceneCandidateFailure(CodeSourceRangeInvalid, SeverityBlocking, false, correlationID, "Scene source range is invalid.", err.Error())
	}

	now := s.now().UTC()
	sceneID := normalizeSceneID(command.SceneID, nextSceneIndex(document.Scenes))
	scene := ScriptSceneDTO{
		ID:            sceneID,
		Index:         sceneIndex(document.Scenes, sceneID),
		Title:         strings.TrimSpace(command.Title),
		Location:      strings.TrimSpace(command.Location),
		TimeOfDay:     strings.TrimSpace(command.TimeOfDay),
		Characters:    cleanStringList(command.Characters),
		Props:         cleanStringList(command.Props),
		Action:        strings.TrimSpace(command.Action),
		Dialogue:      cleanDialogue(command.Dialogue),
		EmotionalBeat: strings.TrimSpace(command.EmotionalBeat),
		SourceRange:   &ScriptSourceRange{StartLine: command.SourceRange.StartLine, EndLine: command.SourceRange.EndLine},
	}
	document.Scenes = upsertScriptScene(document.Scenes, scene)
	document.UpdatedAt = now.Format(time.RFC3339)

	if err := s.writeScriptDocument(root, document, correlationID, "ScriptScene confirmed."); err != nil {
		return s.scriptSceneCandidateFailure(CodeScriptSaveFailed, SeverityBlocking, true, correlationID, "ScriptScene could not be saved.", err.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("script.scene.confirmed", correlationID),
		EventType:     "script.scene.confirmed",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "ScriptScene confirmed.",
		Details: map[string]any{
			"scriptId":  scriptID,
			"sceneId":   scene.ID,
			"startLine": scene.SourceRange.StartLine,
			"endLine":   scene.SourceRange.EndLine,
		},
	})

	return ScriptSceneCandidateResult{
		OK:       true,
		Document: &document,
		Events:   []ProjectEvent{s.event("script.scene.confirmed", "completed", "ScriptScene confirmed from ScriptDocument source range.", correlationID, nil)},
	}
}

func (s *Store) SaveShotCandidate(command SaveShotCandidateCommand) ScriptSceneCandidateResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptSceneCandidateFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before saving a shot candidate.")
	}
	if failure := s.scriptSceneWriteLockFailure(root, correlationID, "shot candidate"); failure != nil {
		return *failure
	}

	manifest, errResult := s.readSceneCandidateManifest(root, correlationID, "saving a shot candidate")
	if errResult != nil {
		return *errResult
	}
	scriptID, document, errResult := s.readScriptDocumentForSceneCandidate(root, command.ScriptID, correlationID)
	if errResult != nil {
		return *errResult
	}
	scene, ok := findScriptScene(document.Scenes, command.ScriptSceneID)
	if !ok {
		return s.scriptSceneCandidateFailure(CodeShotExpansionSourceMissing, SeverityBlocking, false, correlationID, "Shot candidate source scene is missing.", strings.TrimSpace(command.ScriptSceneID))
	}
	if err := validateShotCandidateCommand(command, document, scene); err != nil {
		return s.scriptSceneCandidateFailure(CodeScriptExpansionRowInvalid, SeverityBlocking, false, correlationID, "Shot candidate row is invalid.", err.Error())
	}

	now := s.now().UTC()
	file, err := s.loadShotCandidateFile(root, manifest.Project.ID, scriptID)
	if err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidates could not be loaded before save.", err.Error())
	}

	candidateID := normalizeCandidateID(command.CandidateID, scene.ID, command.Index)
	existing, hasExisting := findShotCandidate(file.Candidates, candidateID)
	if hasExisting && existing.Status == ShotCandidateStatusAccepted {
		return s.scriptSceneCandidateFailure(CodeScriptExpansionRowInvalid, SeverityBlocking, false, correlationID, "Accepted shot candidate cannot be overwritten.", candidateID)
	}
	createdAt := now.Format(time.RFC3339)
	if hasExisting {
		createdAt = existing.CreatedAt
	}
	candidate := ShotCandidateDTO{
		ID:                candidateID,
		ProjectID:         manifest.Project.ID,
		ScriptID:          scriptID,
		ScriptSceneID:     scene.ID,
		Index:             command.Index,
		DurationSeconds:   command.DurationSeconds,
		VisualDescription: strings.TrimSpace(command.VisualDescription),
		CharacterRefs:     cleanCharacterRefs(command.CharacterRefs),
		SourceRange:       command.SourceRange,
		Status:            ShotCandidateStatusCandidate,
		CreatedAt:         createdAt,
		UpdatedAt:         now.Format(time.RFC3339),
	}
	file.Candidates = upsertShotCandidate(file.Candidates, candidate)
	file.UpdatedAt = now.Format(time.RFC3339)
	if err := s.writeShotCandidateFile(root, scriptID, file); err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidate save failed.", err.Error())
	}

	return ScriptSceneCandidateResult{
		OK:         true,
		Candidate:  &candidate,
		Candidates: file.Candidates,
		Events:     []ProjectEvent{s.event("shot.candidate.saved", "completed", "Shot candidate saved without creating a formal ShotCard.", correlationID, nil)},
	}
}

func (s *Store) ListShotCandidates(command ListShotCandidatesCommand) ScriptSceneCandidateResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptSceneCandidateFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before loading shot candidates.")
	}
	manifest, errResult := s.readSceneCandidateManifest(root, correlationID, "loading shot candidates")
	if errResult != nil {
		return *errResult
	}
	scriptID, document, errResult := s.readScriptDocumentForSceneCandidate(root, command.ScriptID, correlationID)
	if errResult != nil {
		return *errResult
	}
	file, err := s.loadShotCandidateFile(root, manifest.Project.ID, scriptID)
	if err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidates could not be loaded.", err.Error())
	}
	return ScriptSceneCandidateResult{
		OK:         true,
		Document:   &document,
		Candidates: file.Candidates,
		Events:     []ProjectEvent{s.event("shot.candidates.loaded", "completed", "Shot candidates loaded.", correlationID, nil)},
	}
}

func (s *Store) RejectShotCandidate(command RejectShotCandidateCommand) ScriptSceneCandidateResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptSceneCandidateFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before rejecting a shot candidate.")
	}
	if failure := s.scriptSceneWriteLockFailure(root, correlationID, "shot candidate"); failure != nil {
		return *failure
	}
	manifest, errResult := s.readSceneCandidateManifest(root, correlationID, "rejecting a shot candidate")
	if errResult != nil {
		return *errResult
	}
	scriptID, _, errResult := s.readScriptDocumentForSceneCandidate(root, command.ScriptID, correlationID)
	if errResult != nil {
		return *errResult
	}
	file, err := s.loadShotCandidateFile(root, manifest.Project.ID, scriptID)
	if err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidates could not be loaded.", err.Error())
	}
	candidate, ok := findShotCandidate(file.Candidates, command.CandidateID)
	if !ok {
		return s.scriptSceneCandidateFailure(CodeShotCandidateMissing, SeverityBlocking, false, correlationID, "Shot candidate is missing.", strings.TrimSpace(command.CandidateID))
	}
	if candidate.Status == ShotCandidateStatusAccepted {
		return s.scriptSceneCandidateFailure(CodeShotCandidateConflict, SeverityBlocking, false, correlationID, "Accepted shot candidate cannot be rejected.", acceptedCandidateDetail(candidate))
	}
	candidate.Status = ShotCandidateStatusRejected
	candidate.RejectionReason = strings.TrimSpace(command.RejectionReason)
	candidate.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	file.Candidates = upsertShotCandidate(file.Candidates, candidate)
	file.UpdatedAt = candidate.UpdatedAt
	if err := s.writeShotCandidateFile(root, scriptID, file); err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidate rejection could not be saved.", err.Error())
	}
	return ScriptSceneCandidateResult{
		OK:         true,
		Candidate:  &candidate,
		Candidates: file.Candidates,
		Events:     []ProjectEvent{s.event("shot.candidate.rejected", "completed", "Shot candidate rejected without creating a formal ShotCard.", correlationID, nil)},
	}
}

func (s *Store) ConfirmShotCandidate(command ConfirmShotCandidateCommand) ScriptSceneCandidateResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptSceneCandidateFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before confirming a shot candidate.")
	}
	if failure := s.scriptSceneWriteLockFailure(root, correlationID, "shot candidate"); failure != nil {
		return *failure
	}
	manifest, errResult := s.readSceneCandidateManifest(root, correlationID, "confirming a shot candidate")
	if errResult != nil {
		return *errResult
	}
	scriptID, document, errResult := s.readScriptDocumentForSceneCandidate(root, command.ScriptID, correlationID)
	if errResult != nil {
		return *errResult
	}
	file, err := s.loadShotCandidateFile(root, manifest.Project.ID, scriptID)
	if err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidates could not be loaded.", err.Error())
	}
	candidate, ok := findShotCandidate(file.Candidates, command.CandidateID)
	if !ok {
		return s.scriptSceneCandidateFailure(CodeShotCandidateMissing, SeverityBlocking, false, correlationID, "Shot candidate is missing.", strings.TrimSpace(command.CandidateID))
	}
	if candidate.Status == ShotCandidateStatusAccepted {
		return s.scriptSceneCandidateFailure(CodeShotCandidateConflict, SeverityBlocking, false, correlationID, "Shot candidate is already accepted.", acceptedCandidateDetail(candidate))
	}
	scene, ok := findScriptScene(document.Scenes, candidate.ScriptSceneID)
	if !ok {
		return s.scriptSceneCandidateFailure(CodeShotExpansionSourceMissing, SeverityBlocking, false, correlationID, "Shot candidate source scene is missing.", candidate.ScriptSceneID)
	}
	if err := validateCandidateForConfirm(candidate, scene); err != nil {
		return s.scriptSceneCandidateFailure(CodeScriptExpansionRowInvalid, SeverityBlocking, false, correlationID, "Shot candidate cannot be confirmed.", err.Error())
	}

	now := s.now().UTC()
	shotID := normalizeShotID(command.ShotID, candidate.ID)
	if err := validateShotCandidateConfirmConflicts(root, file.Candidates, candidate, shotID); err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateConflict, SeverityBlocking, false, correlationID, "Shot candidate confirmation conflicts with existing accepted output.", err.Error())
	}
	shot, err := s.shotFromCandidate(root, manifest, candidate, shotID, command.ConfirmedBy, now)
	if err != nil {
		return s.scriptSceneCandidateFailure(CodeShotConfirmFailed, SeverityBlocking, true, correlationID, "ShotCard could not be prepared.", err.Error())
	}
	if err := s.writeShotCard(root, shot); err != nil {
		return s.scriptSceneCandidateFailure(CodeShotConfirmFailed, SeverityBlocking, true, correlationID, "ShotCard confirm failed before replacing the previous version.", err.Error())
	}
	candidate.Status = ShotCandidateStatusAccepted
	candidate.ShotID = shot.ID
	candidate.UpdatedAt = now.Format(time.RFC3339)
	file.Candidates = upsertShotCandidate(file.Candidates, candidate)
	file.UpdatedAt = candidate.UpdatedAt
	if err := s.writeShotCandidateFile(root, scriptID, file); err != nil {
		return s.scriptSceneCandidateFailure(CodeShotCandidateSaveFailed, SeverityBlocking, true, correlationID, "Shot candidate accepted state could not be saved.", err.Error())
	}
	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("shot.candidate.confirmed", correlationID),
		EventType:     "shot.candidate.confirmed",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "Shot candidate confirmed into draft ShotCard.",
		Details: map[string]any{
			"candidateId": candidate.ID,
			"shotId":      shot.ID,
			"sceneId":     scene.ID,
		},
	})

	return ScriptSceneCandidateResult{
		OK:         true,
		Candidate:  &candidate,
		Candidates: file.Candidates,
		Shot:       &shot,
		Events:     []ProjectEvent{s.event("shot.candidate.confirmed", "completed", "Shot candidate confirmed into a draft ShotCard.", correlationID, nil)},
	}
}

func (s *Store) scriptSceneWriteLockFailure(root string, correlationID string, target string) *ScriptSceneCandidateResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.scriptSceneCandidateFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before saving the "+target+".")
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.scriptSceneCandidateFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before saving the "+target+".", "Open the project with takeover, then retry.")
		return &result
	}
	return nil
}

func (s *Store) readSceneCandidateManifest(root string, correlationID string, action string) (Manifest, *ScriptSceneCandidateResult) {
	manifest, report, err := s.readManifest(root)
	if err == nil {
		return manifest, nil
	}
	operationError := s.errorFromReadFailure(report, err, correlationID)
	return Manifest{}, &ScriptSceneCandidateResult{
		OK:     false,
		Error:  &operationError,
		Events: []ProjectEvent{s.event("script.scene_candidate", "blocked", "Project manifest could not be loaded before "+action+".", correlationID, &operationError)},
	}
}

func (s *Store) readScriptDocumentForSceneCandidate(root string, scriptIDValue string, correlationID string) (string, ScriptDocumentDTO, *ScriptSceneCandidateResult) {
	scriptID, idErr := normalizeScriptDocumentID(scriptIDValue)
	if idErr != nil {
		result := s.scriptSceneCandidateFailure(CodeScriptInvalidID, SeverityBlocking, false, correlationID, "Script document id is invalid.", idErr.Error())
		return "", ScriptDocumentDTO{}, &result
	}
	document, err := s.readScriptDocument(root, scriptID)
	if err != nil {
		result := s.scriptSceneCandidateFailure(CodeScriptLoadFailed, SeverityBlocking, true, correlationID, "ScriptDocument could not be loaded.", err.Error())
		return "", ScriptDocumentDTO{}, &result
	}
	return scriptID, document, nil
}

func (s *Store) writeScriptDocument(root string, document ScriptDocumentDTO, correlationID string, summary string) error {
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	filename := scriptDocumentPath(root, document.ID)
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	if err := s.atomicWrite(filename, data, 0o644, s.instanceID); err != nil {
		return err
	}
	return nil
}

func (s *Store) loadShotCandidateFile(root string, projectID string, scriptID string) (shotCandidateFile, error) {
	filename := shotCandidatePath(root, scriptID)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return shotCandidateFile{
				ProjectID:  projectID,
				ScriptID:   scriptID,
				Candidates: []ShotCandidateDTO{},
			}, nil
		}
		return shotCandidateFile{}, err
	}
	var file shotCandidateFile
	if err := json.Unmarshal(data, &file); err != nil {
		return shotCandidateFile{}, err
	}
	if file.ProjectID == "" {
		file.ProjectID = projectID
	}
	if file.ScriptID == "" {
		file.ScriptID = scriptID
	}
	if file.Candidates == nil {
		file.Candidates = []ShotCandidateDTO{}
	}
	sortShotCandidates(file.Candidates)
	return file, nil
}

func (s *Store) writeShotCandidateFile(root string, scriptID string, file shotCandidateFile) error {
	sortShotCandidates(file.Candidates)
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	filename := shotCandidatePath(root, scriptID)
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return s.atomicWrite(filename, data, 0o644, s.instanceID)
}

func (s *Store) shotFromCandidate(root string, manifest Manifest, candidate ShotCandidateDTO, shotID string, confirmedBy string, now time.Time) (ShotCardDTO, error) {
	existing, err := readExistingShot(root, shotID)
	if err != nil {
		return ShotCardDTO{}, err
	}
	createdAt := now.Format(time.RFC3339)
	if value, ok := existing["createdAt"].(string); ok && strings.TrimSpace(value) != "" {
		createdAt = value
	}
	shot := ShotCardDTO{
		ID:                shotID,
		ProjectID:         manifest.Project.ID,
		SceneID:           candidate.ScriptSceneID,
		SourceCandidateID: candidate.ID,
		ScriptSceneID:     candidate.ScriptSceneID,
		Index:             candidate.Index,
		Title:             fmt.Sprintf("Shot %03d", candidate.Index),
		Description:       candidate.VisualDescription,
		DurationSeconds:   candidate.DurationSeconds,
		AspectRatio:       manifest.Defaults.AspectRatio,
		ShotType:          "unspecified",
		CameraMovement:    "unspecified",
		Action:            candidate.VisualDescription,
		Emotion:           "unspecified",
		CharacterIDs:      characterIDs(candidate.CharacterRefs),
		CharacterRefs:     candidate.CharacterRefs,
		SourceRange:       candidate.SourceRange,
		Status:            ShotStatusDraft,
		ConfirmedBy:       strings.TrimSpace(confirmedBy),
		ConfirmedAt:       now.Format(time.RFC3339),
		CreatedAt:         createdAt,
		UpdatedAt:         now.Format(time.RFC3339),
	}
	if shot.ConfirmedBy == "" {
		shot.ConfirmedBy = "local_user"
	}
	shot.OverwrittenFields = overwrittenShotFields(existing, shot)
	return shot, nil
}

func (s *Store) writeShotCard(root string, shot ShotCardDTO) error {
	data, err := json.MarshalIndent(shot, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	filename := filepath.Join(root, "shots", shot.ID+".json")
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	return s.atomicWrite(filename, data, 0o644, s.instanceID)
}

func (s *Store) scriptSceneCandidateFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) ScriptSceneCandidateResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, sceneCandidateRecoveryActions(code, technicalDetail))
	return ScriptSceneCandidateResult{
		OK:         false,
		Error:      &err,
		Candidates: []ShotCandidateDTO{},
		Events:     []ProjectEvent{s.event("script.scene_candidate", "blocked", userMessage, correlationID, &err)},
	}
}

func sceneCandidateRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeSourceRangeInvalid:
		return []string{"Choose a valid ScriptDocument line range before confirming the scene or candidate."}
	case CodeSceneRequiredFieldMissing:
		return []string{"Fill title, location, and action before confirming the ScriptScene."}
	case CodeScriptExpansionRowInvalid:
		return []string{"Fill candidate index, duration, visual description, and source range before confirming."}
	case CodeShotExpansionSourceMissing:
		return []string{"Confirm the ScriptScene first, then retry the candidate action."}
	case CodeShotCandidateMissing:
		return []string{"Reload shot candidates and choose an existing candidate row."}
	case CodeShotCandidateConflict:
		return []string{"Reload shot candidates, choose an unaccepted row, or use a unique shot/index before confirming."}
	default:
		if strings.TrimSpace(detail) == "" {
			return []string{"Fix the Script-to-Shot input and retry."}
		}
		return []string{detail}
	}
}

func validateSceneSourceRange(document ScriptDocumentDTO, sourceRange ScriptSourceRange, allowOverlap bool, sceneID string) error {
	if err := validateSourceRange(document.RawText, sourceRange); err != nil {
		return err
	}
	if allowOverlap {
		return nil
	}
	for _, scene := range document.Scenes {
		if scene.ID == strings.TrimSpace(sceneID) || scene.SourceRange == nil {
			continue
		}
		if rangesOverlap(*scene.SourceRange, sourceRange) {
			return fmt.Errorf("source range overlaps scene %s", scene.ID)
		}
	}
	return nil
}

func validateSourceRange(rawText string, sourceRange ScriptSourceRange) error {
	if sourceRange.StartLine <= 0 || sourceRange.EndLine <= 0 {
		return errors.New("source range startLine and endLine are required")
	}
	if sourceRange.EndLine < sourceRange.StartLine {
		return fmt.Errorf("source range is reversed: startLine=%d endLine=%d", sourceRange.StartLine, sourceRange.EndLine)
	}
	lineCount := scriptLineCount(rawText)
	if sourceRange.EndLine > lineCount {
		return fmt.Errorf("source range endLine=%d exceeds script line count=%d", sourceRange.EndLine, lineCount)
	}
	return nil
}

func validateShotCandidateCommand(command SaveShotCandidateCommand, document ScriptDocumentDTO, scene ScriptSceneDTO) error {
	if command.Index <= 0 {
		return errors.New("candidate index must be greater than zero")
	}
	if command.DurationSeconds <= 0 {
		return errors.New("candidate durationSeconds must be greater than zero")
	}
	if strings.TrimSpace(command.VisualDescription) == "" {
		return errors.New("candidate visualDescription is required")
	}
	if err := validateSourceRange(document.RawText, command.SourceRange); err != nil {
		return err
	}
	if scene.SourceRange != nil && !rangeContains(*scene.SourceRange, command.SourceRange) {
		return fmt.Errorf("candidate source range must stay within scene %s", scene.ID)
	}
	return nil
}

func validateCandidateForConfirm(candidate ShotCandidateDTO, scene ScriptSceneDTO) error {
	if candidate.Status == ShotCandidateStatusRejected {
		return fmt.Errorf("candidate %s is rejected; edit and save it before retry", candidate.ID)
	}
	if candidate.Index <= 0 || candidate.DurationSeconds <= 0 || strings.TrimSpace(candidate.VisualDescription) == "" {
		return fmt.Errorf("candidate %s has incomplete row fields", candidate.ID)
	}
	if scene.SourceRange == nil || !rangeContains(*scene.SourceRange, candidate.SourceRange) {
		return fmt.Errorf("candidate %s source range is not linked to scene %s", candidate.ID, scene.ID)
	}
	return nil
}

func validateShotCandidateConfirmConflicts(root string, candidates []ShotCandidateDTO, candidate ShotCandidateDTO, shotID string) error {
	for _, existing := range candidates {
		if existing.ID == candidate.ID || existing.Status != ShotCandidateStatusAccepted {
			continue
		}
		if existing.ScriptSceneID == candidate.ScriptSceneID && existing.Index == candidate.Index {
			return fmt.Errorf("scene %s index %d is already accepted by candidate %s", candidate.ScriptSceneID, candidate.Index, existing.ID)
		}
		if strings.TrimSpace(existing.ShotID) == shotID {
			return fmt.Errorf("shot %s is already accepted by candidate %s", shotID, existing.ID)
		}
	}
	existingShot, err := readExistingShot(root, shotID)
	if err != nil {
		return err
	}
	if len(existingShot) == 0 {
		return nil
	}
	sourceCandidateID, _ := existingShot["sourceCandidateId"].(string)
	sourceCandidateID = strings.TrimSpace(sourceCandidateID)
	if sourceCandidateID == candidate.ID {
		return nil
	}
	if sourceCandidateID == "" {
		return fmt.Errorf("shot %s already exists without source candidate lineage", shotID)
	}
	return fmt.Errorf("shot %s already belongs to candidate %s", shotID, sourceCandidateID)
}

func acceptedCandidateDetail(candidate ShotCandidateDTO) string {
	if strings.TrimSpace(candidate.ShotID) == "" {
		return fmt.Sprintf("candidate %s is already accepted", candidate.ID)
	}
	return fmt.Sprintf("candidate %s is already accepted as shot %s", candidate.ID, candidate.ShotID)
}

func missingSceneFields(title string, location string, action string) []string {
	var missing []string
	if strings.TrimSpace(title) == "" {
		missing = append(missing, "title")
	}
	if strings.TrimSpace(location) == "" {
		missing = append(missing, "location")
	}
	if strings.TrimSpace(action) == "" {
		missing = append(missing, "action")
	}
	return missing
}

func normalizeSceneID(sceneID string, index int) string {
	sceneID = strings.TrimSpace(sceneID)
	if sceneID != "" && scriptIDPattern.MatchString(sceneID) {
		return sceneID
	}
	return fmt.Sprintf("scene_%03d", index)
}

func normalizeCandidateID(candidateID string, sceneID string, index int) string {
	candidateID = strings.TrimSpace(candidateID)
	if candidateID != "" && scriptIDPattern.MatchString(candidateID) {
		return candidateID
	}
	return fmt.Sprintf("candidate_%s_%03d", sceneID, index)
}

func normalizeShotID(shotID string, candidateID string) string {
	shotID = strings.TrimSpace(shotID)
	if shotID != "" && scriptIDPattern.MatchString(shotID) {
		return shotID
	}
	candidateID = strings.TrimPrefix(strings.TrimSpace(candidateID), "candidate_")
	if candidateID == "" {
		candidateID = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return "shot_" + candidateID
}

func nextSceneIndex(scenes []ScriptSceneDTO) int {
	maxIndex := 0
	for _, scene := range scenes {
		if scene.Index > maxIndex {
			maxIndex = scene.Index
		}
	}
	return maxIndex + 1
}

func sceneIndex(scenes []ScriptSceneDTO, sceneID string) int {
	for _, scene := range scenes {
		if scene.ID == sceneID && scene.Index > 0 {
			return scene.Index
		}
	}
	return nextSceneIndex(scenes)
}

func upsertScriptScene(scenes []ScriptSceneDTO, scene ScriptSceneDTO) []ScriptSceneDTO {
	next := make([]ScriptSceneDTO, 0, len(scenes)+1)
	replaced := false
	for _, existing := range scenes {
		if existing.ID == scene.ID {
			next = append(next, scene)
			replaced = true
			continue
		}
		next = append(next, existing)
	}
	if !replaced {
		next = append(next, scene)
	}
	sort.SliceStable(next, func(i, j int) bool {
		return next[i].Index < next[j].Index
	})
	return next
}

func findScriptScene(scenes []ScriptSceneDTO, sceneID string) (ScriptSceneDTO, bool) {
	sceneID = strings.TrimSpace(sceneID)
	for _, scene := range scenes {
		if scene.ID == sceneID {
			return scene, true
		}
	}
	return ScriptSceneDTO{}, false
}

func upsertShotCandidate(candidates []ShotCandidateDTO, candidate ShotCandidateDTO) []ShotCandidateDTO {
	next := make([]ShotCandidateDTO, 0, len(candidates)+1)
	replaced := false
	for _, existing := range candidates {
		if existing.ID == candidate.ID {
			next = append(next, candidate)
			replaced = true
			continue
		}
		next = append(next, existing)
	}
	if !replaced {
		next = append(next, candidate)
	}
	sortShotCandidates(next)
	return next
}

func findShotCandidate(candidates []ShotCandidateDTO, candidateID string) (ShotCandidateDTO, bool) {
	candidateID = strings.TrimSpace(candidateID)
	for _, candidate := range candidates {
		if candidate.ID == candidateID {
			return candidate, true
		}
	}
	return ShotCandidateDTO{}, false
}

func sortShotCandidates(candidates []ShotCandidateDTO) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].ScriptSceneID == candidates[j].ScriptSceneID {
			return candidates[i].Index < candidates[j].Index
		}
		return candidates[i].ScriptSceneID < candidates[j].ScriptSceneID
	})
}

func cleanStringList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

func cleanDialogue(values []DialogueLineDTO) []DialogueLineDTO {
	cleaned := make([]DialogueLineDTO, 0, len(values))
	for _, value := range values {
		character := strings.TrimSpace(value.CharacterName)
		text := strings.TrimSpace(value.Text)
		if character == "" && text == "" {
			continue
		}
		cleaned = append(cleaned, DialogueLineDTO{
			CharacterName: character,
			Text:          text,
			Intent:        strings.TrimSpace(value.Intent),
		})
	}
	return cleaned
}

func cleanCharacterRefs(values []ShotCharacterRefDTO) []ShotCharacterRefDTO {
	cleaned := make([]ShotCharacterRefDTO, 0, len(values))
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		characterID := strings.TrimSpace(value.CharacterID)
		if name == "" && characterID == "" {
			continue
		}
		cleaned = append(cleaned, ShotCharacterRefDTO{
			CharacterID:      characterID,
			Name:             name,
			Description:      strings.TrimSpace(value.Description),
			ReferenceAssetID: strings.TrimSpace(value.ReferenceAssetID),
		})
	}
	return cleaned
}

func characterIDs(refs []ShotCharacterRefDTO) []string {
	var ids []string
	for _, ref := range refs {
		if strings.TrimSpace(ref.CharacterID) != "" {
			ids = append(ids, strings.TrimSpace(ref.CharacterID))
		}
	}
	return ids
}

func scriptLineCount(rawText string) int {
	if rawText == "" {
		return 0
	}
	return strings.Count(rawText, "\n") + 1
}

func rangesOverlap(left ScriptSourceRange, right ScriptSourceRange) bool {
	return left.StartLine <= right.EndLine && right.StartLine <= left.EndLine
}

func rangeContains(parent ScriptSourceRange, child ScriptSourceRange) bool {
	return child.StartLine >= parent.StartLine && child.EndLine <= parent.EndLine
}

func shotCandidatePath(root string, scriptID string) string {
	return filepath.Join(root, filepath.FromSlash(path.Join(ShotCandidatesRelativeDir, scriptID+".candidates.json")))
}

func readExistingShot(root string, shotID string) (map[string]any, error) {
	data, err := os.ReadFile(filepath.Join(root, "shots", shotID+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	var existing map[string]any
	if err := json.Unmarshal(data, &existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func overwrittenShotFields(existing map[string]any, shot ShotCardDTO) []string {
	if len(existing) == 0 {
		return []string{}
	}
	fields := []string{}
	if !jsonEqual(existing["index"], shot.Index) {
		fields = append(fields, "index")
	}
	if !jsonEqual(existing["durationSeconds"], shot.DurationSeconds) {
		fields = append(fields, "durationSeconds")
	}
	if !jsonEqual(existing["description"], shot.Description) {
		fields = append(fields, "description")
	}
	if !jsonEqual(existing["action"], shot.Action) {
		fields = append(fields, "action")
	}
	if !jsonEqual(existing["characterIds"], shot.CharacterIDs) {
		fields = append(fields, "characterIds")
	}
	if !jsonEqual(existing["sourceRange"], shot.SourceRange) {
		fields = append(fields, "sourceRange")
	}
	sort.Strings(fields)
	return fields
}

func jsonEqual(left any, right any) bool {
	leftData, _ := json.Marshal(left)
	rightData, _ := json.Marshal(right)
	return string(leftData) == string(rightData)
}
