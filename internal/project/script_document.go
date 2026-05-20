package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	CodeScriptEmpty        = "script_empty"
	CodeScriptAssetMissing = "script_asset_missing"
	CodeScriptSaveFailed   = "save_failed"
	CodeScriptLoadFailed   = "script_load_failed"
	CodeScriptInvalidID    = "script_invalid_id"
)

const ScriptDocumentsRelativeDir = "assets/inputs/scripts"

var scriptIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type SaveScriptDocumentCommand struct {
	Root          string `json:"root"`
	ScriptID      string `json:"scriptId,omitempty"`
	Title         string `json:"title"`
	SourceAssetID string `json:"sourceAssetId,omitempty"`
	RawText       string `json:"rawText"`
	Logline       string `json:"logline,omitempty"`
	Synopsis      string `json:"synopsis,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type LoadScriptDocumentCommand struct {
	Root          string `json:"root"`
	ScriptID      string `json:"scriptId,omitempty"`
	CorrelationID string `json:"correlationId"`
}

type ScriptDocumentResult struct {
	OK       bool               `json:"ok"`
	Document *ScriptDocumentDTO `json:"document,omitempty"`
	Error    *OperationError    `json:"error,omitempty"`
	Events   []ProjectEvent     `json:"events"`
}

type ScriptDocumentDTO struct {
	ID            string           `json:"id"`
	ProjectID     string           `json:"projectId"`
	Title         string           `json:"title"`
	SourceAssetID string           `json:"sourceAssetId,omitempty"`
	RawText       string           `json:"rawText"`
	Logline       string           `json:"logline,omitempty"`
	Synopsis      string           `json:"synopsis,omitempty"`
	Scenes        []ScriptSceneDTO `json:"scenes"`
	CreatedAt     string           `json:"createdAt"`
	UpdatedAt     string           `json:"updatedAt"`
}

type ScriptSceneDTO struct {
	ID            string             `json:"id"`
	Index         int                `json:"index"`
	Title         string             `json:"title"`
	Location      string             `json:"location"`
	TimeOfDay     string             `json:"timeOfDay"`
	Characters    []string           `json:"characters"`
	Props         []string           `json:"props"`
	Action        string             `json:"action"`
	Dialogue      []DialogueLineDTO  `json:"dialogue"`
	EmotionalBeat string             `json:"emotionalBeat"`
	SourceRange   *ScriptSourceRange `json:"sourceRange,omitempty"`
}

type DialogueLineDTO struct {
	CharacterName string `json:"characterName"`
	Text          string `json:"text"`
	Intent        string `json:"intent,omitempty"`
}

type ScriptSourceRange struct {
	StartLine int `json:"startLine"`
	EndLine   int `json:"endLine"`
}

func (s *Store) SaveScriptDocument(command SaveScriptDocumentCommand) ScriptDocumentResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptDocumentFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before saving a script.")
	}

	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		return s.scriptDocumentFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "Project is open in another active session.", "Open the project read-only or confirm takeover before saving the script.")
	}
	if lockInfo.State == LockStateStale {
		return s.scriptDocumentFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "Project has a stale lock that needs takeover before saving the script.", "Open the project with takeover, then retry script save.")
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		return ScriptDocumentResult{
			OK:     false,
			Error:  &operationError,
			Events: []ProjectEvent{s.event("script.document.save", "blocked", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	scriptID, idErr := normalizeScriptDocumentID(command.ScriptID)
	if idErr != nil {
		return s.scriptDocumentFailure(CodeScriptInvalidID, SeverityBlocking, false, correlationID, "Script document id is invalid.", idErr.Error())
	}

	rawText := command.RawText
	sourceAssetID := cleanProjectRelativePath(command.SourceAssetID)
	if sourceAssetID != "" {
		if issues := validateProjectPath("script.sourceAssetId", sourceAssetID); len(issues) > 0 {
			return s.scriptDocumentFailure(CodeScriptAssetMissing, SeverityBlocking, true, correlationID, "Script source asset path is invalid.", issues[0].TechnicalDetail)
		}
	}
	if strings.TrimSpace(rawText) == "" && sourceAssetID != "" {
		var sourceErr error
		rawText, sourceErr = s.readScriptSourceAsset(root, sourceAssetID)
		if sourceErr != nil {
			return s.scriptDocumentFailure(CodeScriptAssetMissing, SeverityBlocking, true, correlationID, "Script source asset could not be imported.", sourceErr.Error())
		}
	}
	if strings.TrimSpace(rawText) == "" {
		return s.scriptDocumentFailure(CodeScriptEmpty, SeverityBlocking, false, correlationID, "Script text is empty.", "Paste script text or choose a non-empty script_source asset before saving.")
	}

	now := s.now().UTC()
	existing, err := s.loadScriptDocumentFile(root, scriptID)
	if err != nil {
		return s.scriptDocumentFailure(CodeScriptLoadFailed, SeverityBlocking, true, correlationID, "Existing ScriptDocument could not be read before save.", err.Error())
	}
	createdAt := now.Format(time.RFC3339)
	scenes := []ScriptSceneDTO{}
	if existing != nil {
		createdAt = existing.CreatedAt
		scenes = existing.Scenes
	}

	document := ScriptDocumentDTO{
		ID:            scriptID,
		ProjectID:     manifest.Project.ID,
		Title:         scriptTitle(command.Title, sourceAssetID),
		SourceAssetID: sourceAssetID,
		RawText:       rawText,
		Logline:       command.Logline,
		Synopsis:      command.Synopsis,
		Scenes:        scenes,
		CreatedAt:     createdAt,
		UpdatedAt:     now.Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return s.scriptDocumentFailure(CodeScriptSaveFailed, SeverityBlocking, true, correlationID, "Script document could not be encoded.", err.Error())
	}
	data = append(data, '\n')

	filename := scriptDocumentPath(root, scriptID)
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return s.scriptDocumentFailure(CodeScriptSaveFailed, SeverityBlocking, true, correlationID, "Script document directory could not be created.", err.Error())
	}
	if err := s.atomicWrite(filename, data, 0o644, s.instanceID); err != nil {
		return s.scriptDocumentFailure(CodeScriptSaveFailed, SeverityBlocking, true, correlationID, "Script document save failed before replacing the previous version.", err.Error())
	}

	_ = s.appendAudit(root, auditEntry{
		EventID:       s.eventID("script.document.saved", correlationID),
		EventType:     "script.document.saved",
		ProjectID:     manifest.Project.ID,
		CorrelationID: correlationID,
		CreatedAt:     now.Format(time.RFC3339),
		Summary:       "ScriptDocument saved.",
		Details: map[string]any{
			"scriptId":      document.ID,
			"sourceAssetId": document.SourceAssetID,
			"rawTextBytes":  len([]byte(document.RawText)),
			"scenes":        len(document.Scenes),
		},
	})

	return ScriptDocumentResult{
		OK:       true,
		Document: &document,
		Events:   []ProjectEvent{s.event("script.document.saved", "completed", "ScriptDocument saved with raw text preserved.", correlationID, nil)},
	}
}

func (s *Store) LoadScriptDocument(command LoadScriptDocumentCommand) ScriptDocumentResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.scriptDocumentFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "Project root is required.", "Choose a project folder before loading a script.")
	}

	scriptID, err := normalizeScriptDocumentID(command.ScriptID)
	if err != nil {
		return s.scriptDocumentFailure(CodeScriptInvalidID, SeverityBlocking, false, correlationID, "Script document id is invalid.", err.Error())
	}

	document, err := s.readScriptDocument(root, scriptID)
	if err != nil {
		return s.scriptDocumentFailure(CodeScriptLoadFailed, SeverityBlocking, true, correlationID, "Script document could not be loaded.", err.Error())
	}

	return ScriptDocumentResult{
		OK:       true,
		Document: &document,
		Events:   []ProjectEvent{s.event("script.document.loaded", "completed", "ScriptDocument loaded.", correlationID, nil)},
	}
}

func (s *Store) readScriptSourceAsset(root string, sourceAssetID string) (string, error) {
	if sourceAssetID == "" {
		return "", errors.New("source asset id is required")
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(sourceAssetID)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *Store) loadScriptDocumentFile(root string, scriptID string) (*ScriptDocumentDTO, error) {
	document, err := s.readScriptDocument(root, scriptID)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return &document, nil
}

func (s *Store) readScriptDocument(root string, scriptID string) (ScriptDocumentDTO, error) {
	data, err := os.ReadFile(scriptDocumentPath(root, scriptID))
	if err != nil {
		return ScriptDocumentDTO{}, err
	}

	var document ScriptDocumentDTO
	if err := json.Unmarshal(data, &document); err != nil {
		return ScriptDocumentDTO{}, err
	}
	if document.Scenes == nil {
		document.Scenes = []ScriptSceneDTO{}
	}
	return document, nil
}

func (s *Store) scriptDocumentFailure(code string, severity string, retryable bool, correlationID string, userMessage string, technicalDetail string) ScriptDocumentResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, scriptRecoveryActions(code, technicalDetail))
	return ScriptDocumentResult{
		OK:     false,
		Error:  &err,
		Events: []ProjectEvent{s.event("script.document", "blocked", userMessage, correlationID, &err)},
	}
}

func scriptRecoveryActions(code string, detail string) []string {
	switch code {
	case CodeScriptEmpty:
		return []string{"Paste non-empty script text or import a non-empty script_source asset."}
	case CodeScriptAssetMissing:
		return []string{"Choose an existing project-relative script_source asset or paste the script text manually."}
	case CodeScriptSaveFailed:
		return []string{"Retry save. Keep editing because the previous script document remains active."}
	case CodeScriptLoadFailed:
		return []string{"Create the ScriptDocument again or restore the script JSON from project backup."}
	default:
		if strings.TrimSpace(detail) == "" {
			return []string{"Retry after fixing the ScriptDocument input."}
		}
		return []string{detail}
	}
}

func normalizeScriptDocumentID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "script_main", nil
	}
	if !scriptIDPattern.MatchString(value) {
		return "", fmt.Errorf("script id %q must use letters, numbers, '_' or '-'", value)
	}
	return value, nil
}

func scriptDocumentPath(root string, scriptID string) string {
	return filepath.Join(root, filepath.FromSlash(path.Join(ScriptDocumentsRelativeDir, scriptID+".script.json")))
}

func scriptTitle(title string, sourceAssetID string) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	if sourceAssetID != "" {
		return strings.TrimSuffix(path.Base(sourceAssetID), path.Ext(sourceAssetID))
	}
	return "Untitled Script"
}

func cleanProjectRelativePath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if value == "" {
		return ""
	}
	return path.Clean(value)
}
