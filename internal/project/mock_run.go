package project

import (
	"crypto/sha256"
	"encoding/hex"
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
	CodeMockRunModeUnsupported = "mock_run_mode_unsupported"
	CodeMockRunTargetRequired  = "mock_run_target_required"
	CodeMockRunTargetMissing   = "mock_run_target_missing"
	CodeMockRunPathRejected    = "mock_run_path_rejected"
	CodeMockRunWriteFailed     = "mock_run_write_failed"
	CodeMockRunReadFailed      = "mock_run_read_failed"
)

const (
	MockRunProviderMode = "mock_local"

	MockRunStatusQueued    = "queued"
	MockRunStatusRunning   = "running"
	MockRunStatusCompleted = "completed"
	MockRunStatusFailed    = "failed"
	MockRunStatusCancelled = "cancelled"
)

type MockRunCommand struct {
	Root          string   `json:"root"`
	RunID         string   `json:"runId,omitempty"`
	ShotID        string   `json:"shotId,omitempty"`
	PackageID     string   `json:"packageId,omitempty"`
	SelectionIDs  []string `json:"selectionIds,omitempty"`
	TaskMode      string   `json:"taskMode,omitempty"`
	RetryOfRunID  string   `json:"retryOfRunId,omitempty"`
	CancelReason  string   `json:"cancelReason,omitempty"`
	CreatedBy     string   `json:"createdBy,omitempty"`
	CorrelationID string   `json:"correlationId"`
}

type MockRunResult struct {
	OK     bool              `json:"ok"`
	Run    *MockRunDTO       `json:"run,omitempty"`
	Health *HealthReport     `json:"health,omitempty"`
	Error  *OperationError   `json:"error,omitempty"`
	Events []MockRunEventDTO `json:"events"`
}

type MockRunDTO struct {
	SchemaVersion string            `json:"schemaVersion,omitempty"`
	RunID         string            `json:"runId"`
	ProjectID     string            `json:"projectId"`
	ShotID        string            `json:"shotId,omitempty"`
	PackageID     string            `json:"packageId,omitempty"`
	SelectionIDs  []string          `json:"selectionIds"`
	TaskMode      string            `json:"taskMode"`
	ProviderMode  string            `json:"providerMode"`
	ContextDigest string            `json:"contextDigest"`
	Status        string            `json:"status"`
	Attempt       int               `json:"attempt"`
	RetryOfRunID  string            `json:"retryOfRunId,omitempty"`
	CancelReason  string            `json:"cancelReason,omitempty"`
	RunPath       string            `json:"runPath"`
	EventsPath    string            `json:"eventsPath"`
	Output        *MockRunOutputDTO `json:"output,omitempty"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
}

type MockRunOutputDTO struct {
	RunID        string `json:"runId"`
	RelativePath string `json:"relativePath"`
	Digest       string `json:"digest"`
	MimeType     string `json:"mimeType"`
	SizeBytes    int64  `json:"sizeBytes"`
	Summary      string `json:"summary"`
}

type MockRunEventDTO struct {
	EventID     string          `json:"eventId"`
	RunID       string          `json:"runId,omitempty"`
	EventType   string          `json:"eventType"`
	State       string          `json:"state"`
	Progress    int             `json:"progress"`
	TargetType  string          `json:"targetType,omitempty"`
	TargetID    string          `json:"targetId,omitempty"`
	Summary     string          `json:"summary"`
	Error       *OperationError `json:"error,omitempty"`
	NextActions []string        `json:"nextActions"`
	CreatedAt   string          `json:"createdAt"`
}

type mockRunAction string

const (
	mockRunActionStart  mockRunAction = "start"
	mockRunActionCancel mockRunAction = "cancel"
	mockRunActionRetry  mockRunAction = "retry"
)

func (s *Store) StartMockRun(command MockRunCommand) MockRunResult {
	return s.mockRun(command, mockRunActionStart)
}

func (s *Store) CancelMockRun(command MockRunCommand) MockRunResult {
	return s.mockRun(command, mockRunActionCancel)
}

func (s *Store) RetryMockRun(command MockRunCommand) MockRunResult {
	return s.mockRun(command, mockRunActionRetry)
}

func (s *Store) mockRun(command MockRunCommand, action mockRunAction) MockRunResult {
	correlationID := s.correlationID(command.CorrelationID)
	root := strings.TrimSpace(command.Root)
	if root == "" {
		return s.mockRunFailure(CodeProjectRootRequired, SeverityBlocking, false, correlationID, "", "Project root is required.", "Choose a project folder before running a mock task.", nil)
	}
	if failure := s.mockRunWriteLockFailure(root, correlationID); failure != nil {
		return *failure
	}

	manifest, report, err := s.readManifest(root)
	if err != nil {
		operationError := s.errorFromReadFailure(report, err, correlationID)
		operationError.TargetType = "mock_run"
		return MockRunResult{
			OK:     false,
			Health: ptr(s.HealthReport(root)),
			Error:  &operationError,
			Events: []MockRunEventDTO{s.mockRunEvent("", "run.fail", MockRunStatusFailed, 0, "mock_run", "", operationError.UserMessage, correlationID, &operationError)},
		}
	}

	taskMode := normalizeMockRunTaskMode(command.TaskMode)
	if taskMode != MockRunProviderMode {
		return s.mockRunFailure(CodeMockRunModeUnsupported, SeverityBlocking, false, correlationID, "", "Mock run only supports mock_local mode.", "taskMode="+strings.TrimSpace(command.TaskMode), ptr(s.HealthReport(root)))
	}

	previous, errResult := s.previousMockRun(root, manifest, command, action, correlationID)
	if errResult != nil {
		return *errResult
	}
	target, errResult := s.resolveMockRunTarget(root, manifest, command, previous, action, correlationID)
	if errResult != nil {
		return *errResult
	}

	now := s.now().UTC()
	attempt := s.nextMockRunAttempt(root, manifest, target, previous)
	runID := mockRunID(target, attempt)
	runDir := path.Join(cleanProjectRelativePath(manifest.Paths.PromptRuns), runID)
	runPath := path.Join(runDir, "run.json")
	eventsPath := path.Join(runDir, "events.jsonl")
	outputPath := path.Join(cleanProjectRelativePath(manifest.Paths.AssetOutputs), "mock-run", runID, "placeholder-output.txt")
	for field, relative := range map[string]string{
		"mock_run.runPath":    runPath,
		"mock_run.eventsPath": eventsPath,
		"mock_run.outputPath": outputPath,
	} {
		if action == mockRunActionCancel && field == "mock_run.outputPath" {
			continue
		}
		if issues := validateProjectPath(field, relative); len(issues) > 0 {
			return s.mockRunFailure(CodeMockRunPathRejected, SeverityBlocking, false, correlationID, target.id(), "Mock run path was rejected.", issues[0].TechnicalDetail, ptr(s.HealthReport(root)))
		}
	}

	contextDigest := mockRunContextDigest(manifest.Project.ID, target, taskMode)
	status := MockRunStatusCompleted
	if action == mockRunActionCancel {
		status = MockRunStatusCancelled
	}
	run := MockRunDTO{
		SchemaVersion: CurrentSchemaVersion,
		RunID:         runID,
		ProjectID:     manifest.Project.ID,
		ShotID:        target.shotID,
		PackageID:     target.packageID,
		SelectionIDs:  target.selectionIDs,
		TaskMode:      taskMode,
		ProviderMode:  MockRunProviderMode,
		ContextDigest: contextDigest,
		Status:        status,
		Attempt:       attempt,
		RetryOfRunID:  retryOfRunID(command, previous, action),
		CancelReason:  normalizeCancelReason(command.CancelReason, action),
		RunPath:       runPath,
		EventsPath:    eventsPath,
		CreatedAt:     now.Format(time.RFC3339),
		UpdatedAt:     now.Format(time.RFC3339),
	}

	if action != mockRunActionCancel {
		output, writeErr := s.writeMockRunOutput(root, outputPath, run, now)
		if writeErr != nil {
			_ = s.appendMockRunAudit(root, manifest.Project.ID, correlationID, now, "run.fail", "Mock run failed before writing placeholder output.", target, map[string]any{"error": writeErr.Error()})
			return s.mockRunFailure(CodeMockRunWriteFailed, SeverityBlocking, true, correlationID, target.id(), "Mock run placeholder output could not be written.", writeErr.Error(), ptr(s.HealthReport(root)))
		}
		run.Output = output
	}

	events := s.mockRunEvents(run, action, target, correlationID, now)
	if err := s.writeMockRunFiles(root, run, events); err != nil {
		_ = s.appendMockRunAudit(root, manifest.Project.ID, correlationID, now, "run.fail", "Mock run failed while writing run records.", target, map[string]any{"error": err.Error()})
		return s.mockRunFailure(CodeMockRunWriteFailed, SeverityBlocking, true, correlationID, target.id(), "Mock run record could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}
	if err := s.updateManifestForMockRun(root, manifest, run, now); err != nil {
		_ = s.appendMockRunAudit(root, manifest.Project.ID, correlationID, now, "run.fail", "Mock run failed while updating project graph history.", target, map[string]any{"error": err.Error()})
		return s.mockRunFailure(CodeMockRunWriteFailed, SeverityBlocking, true, correlationID, target.id(), "Project manifest could not record the mock run.", err.Error(), ptr(s.HealthReport(root)))
	}
	if err := s.appendMockRunAudits(root, manifest.Project.ID, correlationID, now, action, run, target); err != nil {
		return s.mockRunFailure(CodeMockRunWriteFailed, SeverityBlocking, true, correlationID, target.id(), "Mock run audit events could not be written.", err.Error(), ptr(s.HealthReport(root)))
	}

	health := s.HealthReport(root)
	return MockRunResult{
		OK:     true,
		Run:    &run,
		Health: &health,
		Events: events,
	}
}

func (s *Store) mockRunWriteLockFailure(root string, correlationID string) *MockRunResult {
	lockInfo := s.inspectLock(root)
	if lockInfo.State == LockStateActive {
		result := s.mockRunFailure(CodeProjectActiveLock, SeverityBlocking, false, correlationID, "", "Project is open in another active session.", "Open the project read-only or confirm takeover before starting a mock run.", ptr(s.HealthReport(root)))
		return &result
	}
	if lockInfo.State == LockStateStale {
		result := s.mockRunFailure(CodeProjectStaleLock, SeverityWarning, true, correlationID, "", "Project has a stale lock that needs takeover before starting a mock run.", "Open the project with takeover, then retry mock run.", ptr(s.HealthReport(root)))
		return &result
	}
	return nil
}

func (s *Store) previousMockRun(root string, manifest Manifest, command MockRunCommand, action mockRunAction, correlationID string) (*MockRunDTO, *MockRunResult) {
	previousRunID := strings.TrimSpace(command.RetryOfRunID)
	if previousRunID == "" && (action == mockRunActionCancel || action == mockRunActionRetry) {
		previousRunID = strings.TrimSpace(command.RunID)
	}
	if previousRunID == "" {
		return nil, nil
	}
	previous, ok, err := s.readMockRunByID(root, manifest, previousRunID)
	if err != nil {
		result := s.mockRunFailure(CodeMockRunReadFailed, SeverityBlocking, true, correlationID, previousRunID, "Previous mock run could not be read.", err.Error(), ptr(s.HealthReport(root)))
		return nil, &result
	}
	if !ok {
		result := s.mockRunFailure(CodeMockRunTargetMissing, SeverityBlocking, false, correlationID, previousRunID, "Previous mock run was not found.", previousRunID, ptr(s.HealthReport(root)))
		return nil, &result
	}
	return &previous, nil
}

type mockRunTarget struct {
	shotID       string
	packageID    string
	selectionIDs []string
}

func (target mockRunTarget) id() string {
	if target.packageID != "" {
		return target.packageID
	}
	if target.shotID != "" {
		return target.shotID
	}
	return strings.Join(target.selectionIDs, ",")
}

func (target mockRunTarget) kind() string {
	if target.packageID != "" {
		return "package"
	}
	if target.shotID != "" {
		return "shot"
	}
	return "selection"
}

func (s *Store) resolveMockRunTarget(root string, manifest Manifest, command MockRunCommand, previous *MockRunDTO, action mockRunAction, correlationID string) (mockRunTarget, *MockRunResult) {
	target := mockRunTarget{
		shotID:       strings.TrimSpace(command.ShotID),
		packageID:    strings.TrimSpace(command.PackageID),
		selectionIDs: cleanStringList(command.SelectionIDs),
	}
	if previous != nil {
		if target.shotID == "" {
			target.shotID = previous.ShotID
		}
		if target.packageID == "" {
			target.packageID = previous.PackageID
		}
		if len(target.selectionIDs) == 0 {
			target.selectionIDs = cleanStringList(previous.SelectionIDs)
		}
	}

	if target.packageID != "" {
		pkg, ok := findPackageManifestByID(s.readPackageManifests(root, manifest), target.packageID)
		if !ok {
			result := s.mockRunFailure(CodeMockRunTargetMissing, SeverityBlocking, false, correlationID, target.packageID, "Generation package for mock run was not found.", target.packageID, ptr(s.HealthReport(root)))
			return mockRunTarget{}, &result
		}
		if target.shotID == "" {
			target.shotID = strings.TrimSpace(pkg.ShotID)
		}
	}

	if target.shotID == "" && len(target.selectionIDs) == 0 {
		result := s.mockRunFailure(CodeMockRunTargetRequired, SeverityBlocking, false, correlationID, "", "Mock run requires a shot, package, or selection.", "Select a shot or exported package before starting mock run.", ptr(s.HealthReport(root)))
		return mockRunTarget{}, &result
	}
	if target.shotID != "" {
		record, shotErr := s.readShotContextRecord(root, manifest, target.shotID, correlationID)
		if shotErr != nil {
			result := s.mockRunFailureFromShotResult(*shotErr, correlationID)
			return mockRunTarget{}, &result
		}
		target.shotID = record.shot.ID
	}
	if len(target.selectionIDs) == 0 {
		target.selectionIDs = []string{target.id()}
	}
	sort.Strings(target.selectionIDs)

	_ = action
	return target, nil
}

func findPackageManifestByID(packages map[string][]packageManifest, packageID string) (packageManifest, bool) {
	for _, group := range packages {
		for _, pkg := range group {
			if strings.TrimSpace(pkg.PackageID) == packageID {
				return pkg, true
			}
		}
	}
	return packageManifest{}, false
}

func (s *Store) mockRunFailureFromShotResult(result ShotContextResult, correlationID string) MockRunResult {
	if result.Error == nil {
		return s.mockRunFailure(CodeMockRunTargetMissing, SeverityBlocking, false, correlationID, "", "Shot could not be loaded for mock run.", "Shot context result did not include an error.", nil)
	}
	err := *result.Error
	err.TargetType = "mock_run"
	return MockRunResult{
		OK:     false,
		Error:  &err,
		Events: []MockRunEventDTO{s.mockRunEvent("", "run.fail", MockRunStatusFailed, 0, "mock_run", err.TargetID, err.UserMessage, correlationID, &err)},
	}
}

func (s *Store) nextMockRunAttempt(root string, manifest Manifest, target mockRunTarget, previous *MockRunDTO) int {
	attempt := 1
	if previous != nil && previous.Attempt >= attempt {
		attempt = previous.Attempt + 1
	}
	runs, err := s.readMockRuns(root, manifest)
	if err != nil {
		return attempt
	}
	for _, run := range runs {
		if run.ShotID == target.shotID && run.PackageID == target.packageID && sameStringSet(run.SelectionIDs, target.selectionIDs) && run.Attempt >= attempt {
			attempt = run.Attempt + 1
		}
	}
	for s.mockRunDirectoryExists(root, manifest, mockRunID(target, attempt)) {
		attempt++
	}
	return attempt
}

func (s *Store) mockRunDirectoryExists(root string, manifest Manifest, runID string) bool {
	runDir := path.Join(cleanProjectRelativePath(manifest.Paths.PromptRuns), strings.TrimSpace(runID))
	if issues := validateProjectPath("mock_run.runDir", runDir); len(issues) > 0 {
		return false
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(runDir)))
	return err == nil && info.IsDir()
}

func mockRunID(target mockRunTarget, attempt int) string {
	return "run_mock_" + safeFileToken(target.kind()+"_"+target.id()) + fmt.Sprintf("_%03d", attempt)
}

func retryOfRunID(command MockRunCommand, previous *MockRunDTO, action mockRunAction) string {
	if action != mockRunActionRetry {
		return ""
	}
	value := strings.TrimSpace(command.RetryOfRunID)
	if value == "" {
		value = strings.TrimSpace(command.RunID)
	}
	if value == "" && previous != nil {
		value = previous.RunID
	}
	return value
}

func normalizeCancelReason(value string, action mockRunAction) string {
	if action != mockRunActionCancel {
		return ""
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "user_cancelled"
	}
	return value
}

func normalizeMockRunTaskMode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return MockRunProviderMode
	}
	return value
}

func mockRunContextDigest(projectID string, target mockRunTarget, taskMode string) string {
	payload := strings.Join([]string{
		strings.TrimSpace(projectID),
		strings.TrimSpace(target.shotID),
		strings.TrimSpace(target.packageID),
		strings.Join(cleanStringList(target.selectionIDs), ","),
		strings.TrimSpace(taskMode),
	}, "\n")
	sum := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (s *Store) writeMockRunOutput(root string, outputPath string, run MockRunDTO, now time.Time) (*MockRunOutputDTO, error) {
	summary := "Deterministic mock_local placeholder output."
	body := strings.Join([]string{
		"Tuyu Studio mock_local placeholder output",
		"run_id=" + run.RunID,
		"project_id=" + run.ProjectID,
		"shot_id=" + run.ShotID,
		"package_id=" + run.PackageID,
		"task_mode=" + run.TaskMode,
		"context_digest=" + run.ContextDigest,
		"created_at=" + now.Format(time.RFC3339),
		"summary=" + summary,
		"",
	}, "\n")
	absolute := filepath.Join(root, filepath.FromSlash(outputPath))
	if err := s.atomicWrite(absolute, []byte(body), 0o644, s.instanceID); err != nil {
		return nil, err
	}
	digest, size, err := digestAssetFile(absolute)
	if err != nil {
		return nil, err
	}
	return &MockRunOutputDTO{
		RunID:        run.RunID,
		RelativePath: outputPath,
		Digest:       digest,
		MimeType:     "text/plain; charset=utf-8",
		SizeBytes:    size,
		Summary:      summary,
	}, nil
}

func (s *Store) mockRunEvents(run MockRunDTO, action mockRunAction, target mockRunTarget, correlationID string, now time.Time) []MockRunEventDTO {
	if action == mockRunActionCancel {
		return []MockRunEventDTO{
			s.mockRunEventAt(run.RunID, "run.create", MockRunStatusQueued, 0, target.kind(), target.id(), "Mock run cancellation attempt was queued.", correlationID, nil, now),
			s.mockRunEventAt(run.RunID, "run.cancel", MockRunStatusCancelled, 100, target.kind(), target.id(), "Mock run was cancelled locally before provider submission.", correlationID, nil, now),
		}
	}
	events := []MockRunEventDTO{
		s.mockRunEventAt(run.RunID, "run.create", MockRunStatusQueued, 0, target.kind(), target.id(), "Mock run was queued locally.", correlationID, nil, now),
		s.mockRunEventAt(run.RunID, "run.start", MockRunStatusRunning, 15, target.kind(), target.id(), "Mock run started with deterministic local runner.", correlationID, nil, now),
		s.mockRunEventAt(run.RunID, "run.progress", MockRunStatusRunning, 60, target.kind(), target.id(), "Mock run generated placeholder output.", correlationID, nil, now),
	}
	if action == mockRunActionRetry {
		events = append([]MockRunEventDTO{
			s.mockRunEventAt(run.RunID, "run.retry", MockRunStatusQueued, 0, target.kind(), target.id(), "Mock run retry created a new attempt.", correlationID, nil, now),
		}, events...)
	}
	events = append(events, s.mockRunEventAt(run.RunID, "run.complete", MockRunStatusCompleted, 100, target.kind(), target.id(), "Mock run completed with placeholder output.", correlationID, nil, now))
	return events
}

func (s *Store) mockRunEvent(runID string, eventType string, state string, progress int, targetType string, targetID string, summary string, correlationID string, err *OperationError) MockRunEventDTO {
	return s.mockRunEventAt(runID, eventType, state, progress, targetType, targetID, summary, correlationID, err, s.now().UTC())
}

func (s *Store) mockRunEventAt(runID string, eventType string, state string, progress int, targetType string, targetID string, summary string, correlationID string, err *OperationError, now time.Time) MockRunEventDTO {
	nextActions := []string{}
	if err != nil {
		nextActions = err.RecoveryActions
	}
	idBase := strings.Join([]string{runID, eventType, state, s.correlationID(correlationID)}, "-")
	return MockRunEventDTO{
		EventID:     "evt-" + safeFileToken(idBase),
		RunID:       runID,
		EventType:   eventType,
		State:       state,
		Progress:    progress,
		TargetType:  targetType,
		TargetID:    targetID,
		Summary:     summary,
		Error:       err,
		NextActions: nextActions,
		CreatedAt:   now.UTC().Format(time.RFC3339),
	}
}

func (s *Store) writeMockRunFiles(root string, run MockRunDTO, events []MockRunEventDTO) error {
	runAbs := filepath.Join(root, filepath.FromSlash(run.RunPath))
	eventsAbs := filepath.Join(root, filepath.FromSlash(run.EventsPath))
	if _, err := os.Stat(filepath.Dir(runAbs)); err == nil {
		return fmt.Errorf("mock run directory already exists: %s", run.RunID)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(runAbs), 0o755); err != nil {
		return err
	}
	for _, event := range events {
		if err := appendMockRunEventJSONL(eventsAbs, event); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return s.atomicWrite(runAbs, data, 0o644, s.instanceID)
}

func appendMockRunEventJSONL(filename string, event MockRunEventDTO) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}

func (s *Store) updateManifestForMockRun(root string, manifest Manifest, run MockRunDTO, now time.Time) error {
	nodeID := firstNodeIDByKind(manifest.Graph.Nodes, "prompt_run")
	if nodeID == "" {
		nodeID = "node_run_" + safeFileToken(run.RunID)
		manifest.Graph.Nodes = append(manifest.Graph.Nodes, Node{
			ID:     nodeID,
			Kind:   "prompt_run",
			RefID:  run.RunPath,
			Status: run.Status,
		})
	} else {
		for index := range manifest.Graph.Nodes {
			node := &manifest.Graph.Nodes[index]
			if strings.TrimSpace(node.ID) == nodeID {
				node.RefID = run.RunPath
				node.Status = run.Status
			}
		}
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

func (s *Store) appendMockRunAudits(root string, projectID string, correlationID string, now time.Time, action mockRunAction, run MockRunDTO, target mockRunTarget) error {
	if action == mockRunActionRetry {
		if err := s.appendMockRunAudit(root, projectID, correlationID, now, "run.retry", "Mock run retry requested.", target, map[string]any{
			"runId":        run.RunID,
			"retryOfRunId": run.RetryOfRunID,
			"attempt":      run.Attempt,
		}); err != nil {
			return err
		}
	}
	events := []struct {
		eventType string
		summary   string
	}{
		{"run.create", "Mock run record created."},
	}
	if action == mockRunActionCancel {
		events = append(events, struct {
			eventType string
			summary   string
		}{"run.cancel", "Mock run cancelled locally."})
	} else {
		events = append(events,
			struct {
				eventType string
				summary   string
			}{"run.start", "Mock run started locally."},
			struct {
				eventType string
				summary   string
			}{"run.complete", "Mock run completed locally."},
		)
	}
	for _, event := range events {
		details := map[string]any{
			"runId":         run.RunID,
			"providerMode":  run.ProviderMode,
			"taskMode":      run.TaskMode,
			"status":        run.Status,
			"attempt":       run.Attempt,
			"targetType":    target.kind(),
			"targetId":      target.id(),
			"contextDigest": run.ContextDigest,
		}
		if run.Output != nil {
			details["outputPath"] = run.Output.RelativePath
			details["outputDigest"] = run.Output.Digest
		}
		if run.CancelReason != "" {
			details["cancelReason"] = run.CancelReason
		}
		if err := s.appendMockRunAudit(root, projectID, correlationID, now, event.eventType, event.summary, target, details); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) appendMockRunAudit(root string, projectID string, correlationID string, now time.Time, eventType string, summary string, target mockRunTarget, details map[string]any) error {
	if details == nil {
		details = map[string]any{}
	}
	details["targetType"] = target.kind()
	details["targetId"] = target.id()
	return s.appendAudit(root, auditEntry{
		EventID:       s.eventID(eventType, correlationID),
		EventType:     eventType,
		ProjectID:     projectID,
		CorrelationID: correlationID,
		CreatedAt:     now.UTC().Format(time.RFC3339),
		Summary:       summary,
		Details:       details,
	})
}

func (s *Store) readMockRunByID(root string, manifest Manifest, runID string) (MockRunDTO, bool, error) {
	runs, err := s.readMockRuns(root, manifest)
	if err != nil {
		return MockRunDTO{}, false, err
	}
	for _, run := range runs {
		if strings.TrimSpace(run.RunID) == strings.TrimSpace(runID) {
			return run, true, nil
		}
	}
	return MockRunDTO{}, false, nil
}

func (s *Store) readMockRuns(root string, manifest Manifest) ([]MockRunDTO, error) {
	dirRelative := cleanProjectRelativePath(manifest.Paths.PromptRuns)
	if issues := validateProjectPath("paths.promptRuns", dirRelative); len(issues) > 0 {
		return nil, errors.New(issues[0].TechnicalDetail)
	}
	dir := filepath.Join(root, filepath.FromSlash(dirRelative))
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	runs := make([]MockRunDTO, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			run, ok, err := readMockRunFile(filepath.Join(dir, entry.Name(), "run.json"))
			if err != nil {
				return nil, err
			}
			if ok {
				runs = append(runs, run)
			}
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		run, ok, err := readMockRunFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if ok {
			runs = append(runs, run)
		}
	}
	return runs, nil
}

func readMockRunFile(filename string) (MockRunDTO, bool, error) {
	data, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		return MockRunDTO{}, false, nil
	}
	if err != nil {
		return MockRunDTO{}, false, err
	}
	var run MockRunDTO
	if err := json.Unmarshal(data, &run); err == nil && strings.TrimSpace(run.RunID) != "" {
		if run.Attempt == 0 {
			run.Attempt = 1
		}
		return run, true, nil
	}

	var legacy struct {
		RunID      string `json:"runId"`
		Status     string `json:"status"`
		TargetType string `json:"targetType"`
		TargetID   string `json:"targetId"`
		CreatedAt  string `json:"createdAt"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return MockRunDTO{}, false, err
	}
	if strings.TrimSpace(legacy.RunID) == "" {
		return MockRunDTO{}, false, nil
	}
	run = MockRunDTO{
		RunID:        legacy.RunID,
		Status:       legacy.Status,
		ProviderMode: MockRunProviderMode,
		TaskMode:     MockRunProviderMode,
		Attempt:      1,
		CreatedAt:    legacy.CreatedAt,
		UpdatedAt:    legacy.CreatedAt,
	}
	if legacy.TargetType == "shot" {
		run.ShotID = legacy.TargetID
		run.SelectionIDs = []string{legacy.TargetID}
	}
	return run, true, nil
}

func (s *Store) mockRunFailure(code string, severity string, retryable bool, correlationID string, targetID string, userMessage string, technicalDetail string, health *HealthReport) MockRunResult {
	err := s.operationError(code, severity, retryable, correlationID, userMessage, technicalDetail, []string{technicalDetail})
	err.TargetType = "mock_run"
	err.TargetID = targetID
	if code == CodeProjectRootRequired {
		err.RecoveryActions = []string{technicalDetail}
	}
	if code == CodeMockRunModeUnsupported {
		err.RecoveryActions = []string{"Use mock_local mode for this Alpha shell issue."}
	}
	if code == CodeMockRunTargetRequired {
		err.RecoveryActions = []string{"Select a context_ready shot or exported generation package before starting a mock run."}
	}
	if code == CodeMockRunPathRejected {
		err.RecoveryActions = []string{"Keep mock run records and outputs under project-relative promptRuns and assetOutputs paths."}
	}
	if code == CodeMockRunWriteFailed {
		err.RecoveryActions = []string{"Check project write permissions and retry to create a new mock run attempt."}
	}
	return MockRunResult{
		OK:     false,
		Health: health,
		Error:  &err,
		Events: []MockRunEventDTO{s.mockRunEvent("", "run.fail", MockRunStatusFailed, 0, "mock_run", targetID, userMessage, correlationID, &err)},
	}
}

func sameStringSet(left []string, right []string) bool {
	left = cleanStringList(left)
	right = cleanStringList(right)
	if len(left) != len(right) {
		return false
	}
	sort.Strings(left)
	sort.Strings(right)
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
