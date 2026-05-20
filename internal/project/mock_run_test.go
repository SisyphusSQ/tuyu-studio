package project

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMockRunStartWritesRunEventsOutputAndAudit(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC), "mock-run-start")

	result := store.StartMockRun(MockRunCommand{
		Root:          root,
		ShotID:        "shot_002",
		SelectionIDs:  []string{"node_shot_002"},
		CorrelationID: "corr-mock-start",
	})
	if !result.OK || result.Run == nil {
		t.Fatalf("StartMockRun() = %#v, want completed run", result)
	}
	run := result.Run
	if run.ProviderMode != MockRunProviderMode || run.TaskMode != MockRunProviderMode {
		t.Fatalf("run mode = provider %q task %q, want mock_local", run.ProviderMode, run.TaskMode)
	}
	if run.Status != MockRunStatusCompleted {
		t.Fatalf("run status = %q, want completed", run.Status)
	}
	if run.Attempt != 1 {
		t.Fatalf("run attempt = %d, want 1", run.Attempt)
	}
	if run.Output == nil {
		t.Fatal("run output = nil, want placeholder output")
	}
	if strings.HasPrefix(run.RunPath, "/") || strings.Contains(run.RunPath, "..") {
		t.Fatalf("run path = %q, want project-relative safe path", run.RunPath)
	}
	if !strings.HasPrefix(run.RunPath, "prompts/runs/"+run.RunID+"/") {
		t.Fatalf("run path = %q, want run directory under prompts/runs", run.RunPath)
	}
	if !strings.HasPrefix(run.Output.RelativePath, "assets/outputs/mock-run/"+run.RunID+"/") {
		t.Fatalf("output path = %q, want mock output outside assets/results", run.Output.RelativePath)
	}
	if strings.Contains(run.Output.RelativePath, "assets/results") {
		t.Fatalf("output path = %q, must not use result import directory", run.Output.RelativePath)
	}
	if len(result.Events) != 4 {
		t.Fatalf("events = %#v, want queued/running/progress/completed", result.Events)
	}
	if result.Events[0].State != MockRunStatusQueued || result.Events[len(result.Events)-1].State != MockRunStatusCompleted {
		t.Fatalf("event sequence = %#v, want queued -> completed", result.Events)
	}

	runData := readTextFile(t, filepath.Join(root, filepath.FromSlash(run.RunPath)))
	if !strings.Contains(runData, `"providerMode": "mock_local"`) {
		t.Fatalf("run record missing providerMode:\n%s", runData)
	}
	eventsData := readTextFile(t, filepath.Join(root, filepath.FromSlash(run.EventsPath)))
	if !strings.Contains(eventsData, `"eventType":"run.complete"`) {
		t.Fatalf("events jsonl missing run.complete:\n%s", eventsData)
	}
	outputData := readTextFile(t, filepath.Join(root, filepath.FromSlash(run.Output.RelativePath)))
	if !strings.Contains(outputData, "Tuyu Studio mock_local placeholder output") {
		t.Fatalf("output file missing placeholder body:\n%s", outputData)
	}
	auditData := readTextFile(t, filepath.Join(root, filepath.FromSlash(AuditEventsRelativePath)))
	if !strings.Contains(auditData, `"eventType":"run.complete"`) {
		t.Fatalf("audit missing run.complete:\n%s", auditData)
	}

	graph := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "corr-mock-graph"})
	if !graph.OK || graph.Canvas == nil || len(graph.Canvas.Frames) == 0 {
		t.Fatalf("GraphView() = %#v, want frame history", graph)
	}
	if graph.Canvas.Frames[0].HistorySummary.CurrentRunID != run.RunPath {
		t.Fatalf("current run = %q, want %q", graph.Canvas.Frames[0].HistorySummary.CurrentRunID, run.RunPath)
	}
	if result.Health == nil || result.Health.HasBlocking() {
		t.Fatalf("mock run health = %#v, want no blocking items", result.Health)
	}
}

func TestMockRunRetryCreatesNewAttemptWithoutOverwritingPreviousOutput(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 15, 5, 0, 0, time.UTC), "mock-run-retry")

	first := store.StartMockRun(MockRunCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-mock-first"})
	if !first.OK || first.Run == nil || first.Run.Output == nil {
		t.Fatalf("first StartMockRun() = %#v, want output", first)
	}
	second := store.RetryMockRun(MockRunCommand{Root: root, RunID: first.Run.RunID, CorrelationID: "corr-mock-retry"})
	if !second.OK || second.Run == nil || second.Run.Output == nil {
		t.Fatalf("RetryMockRun() = %#v, want output", second)
	}
	if second.Run.Attempt != first.Run.Attempt+1 {
		t.Fatalf("retry attempt = %d, want %d", second.Run.Attempt, first.Run.Attempt+1)
	}
	if second.Run.RetryOfRunID != first.Run.RunID {
		t.Fatalf("retryOfRunId = %q, want %q", second.Run.RetryOfRunID, first.Run.RunID)
	}
	if second.Run.Output.RelativePath == first.Run.Output.RelativePath {
		t.Fatalf("retry reused output path %q", second.Run.Output.RelativePath)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(first.Run.Output.RelativePath))); err != nil {
		t.Fatalf("first output was overwritten or removed: %v", err)
	}
	if !strings.Contains(readTextFile(t, filepath.Join(root, filepath.FromSlash(second.Run.EventsPath))), `"eventType":"run.retry"`) {
		t.Fatalf("retry events missing run.retry")
	}
}

func TestMockRunCancelWritesCancelledAttemptWithoutOutput(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 15, 10, 0, 0, time.UTC), "mock-run-cancel")

	result := store.CancelMockRun(MockRunCommand{
		Root:          root,
		ShotID:        "shot_002",
		CancelReason:  "user_cancelled",
		CorrelationID: "corr-mock-cancel",
	})
	if !result.OK || result.Run == nil {
		t.Fatalf("CancelMockRun() = %#v, want cancelled run", result)
	}
	if result.Run.Status != MockRunStatusCancelled {
		t.Fatalf("cancel status = %q, want cancelled", result.Run.Status)
	}
	if result.Run.Output != nil {
		t.Fatalf("cancel output = %#v, want nil", result.Run.Output)
	}
	if len(result.Events) != 2 || result.Events[1].EventType != "run.cancel" {
		t.Fatalf("cancel events = %#v, want create/cancel", result.Events)
	}
	auditData := readTextFile(t, filepath.Join(root, filepath.FromSlash(AuditEventsRelativePath)))
	if !strings.Contains(auditData, `"eventType":"run.cancel"`) {
		t.Fatalf("audit missing run.cancel:\n%s", auditData)
	}
}

func TestMockRunRejectsUnsafeManifestPath(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 15, 15, 0, 0, time.UTC), "mock-run-path")
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest.Paths.AssetOutputs = "../outside"
	writeManifestBypassValidation(t, root, manifest)

	result := store.StartMockRun(MockRunCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-mock-path"})
	if result.OK || result.Error == nil {
		t.Fatalf("StartMockRun() = %#v, want manifest path error", result)
	}
	if result.Error.Code != CodeProjectManifestInvalid {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeProjectManifestInvalid)
	}
}

func TestMockRunReportsOutputWriteFailure(t *testing.T) {
	root := copyExampleFixture(t)
	now := time.Date(2026, 5, 20, 15, 20, 0, 0, time.UTC)
	store := NewStore(StoreOptions{
		Now:        func() time.Time { return now },
		InstanceID: "mock-run-output-fail",
		Hostname:   "test-host",
		PID:        os.Getpid(),
		StaleAfter: time.Minute,
		AtomicWrite: func(filename string, data []byte, perm fs.FileMode, instanceID string) error {
			if strings.Contains(filepath.ToSlash(filename), "/assets/outputs/mock-run/") {
				return errors.New("injected output write failure")
			}
			return atomicWriteFile(filename, data, perm, instanceID)
		},
	})

	result := store.StartMockRun(MockRunCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-mock-output-fail"})
	if result.OK || result.Error == nil {
		t.Fatalf("StartMockRun() = %#v, want output write error", result)
	}
	if result.Error.Code != CodeMockRunWriteFailed {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeMockRunWriteFailed)
	}
}

func TestMockRunRetryAfterPartialRunRecordFailureUsesNextAttempt(t *testing.T) {
	root := copyExampleFixture(t)
	now := time.Date(2026, 5, 20, 15, 25, 0, 0, time.UTC)
	failingStore := NewStore(StoreOptions{
		Now:        func() time.Time { return now },
		InstanceID: "mock-run-record-fail",
		Hostname:   "test-host",
		PID:        os.Getpid(),
		StaleAfter: time.Minute,
		AtomicWrite: func(filename string, data []byte, perm fs.FileMode, instanceID string) error {
			if strings.HasSuffix(filepath.ToSlash(filename), "/run.json") {
				return errors.New("injected run record failure")
			}
			return atomicWriteFile(filename, data, perm, instanceID)
		},
	})

	failed := failingStore.StartMockRun(MockRunCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-mock-record-fail"})
	if failed.OK || failed.Error == nil {
		t.Fatalf("StartMockRun() = %#v, want run record failure", failed)
	}
	if failed.Error.Code != CodeMockRunWriteFailed {
		t.Fatalf("error code = %q, want %s", failed.Error.Code, CodeMockRunWriteFailed)
	}
	partialDir := filepath.Join(root, "prompts", "runs", "run_mock_shot_shot_002_001")
	if info, err := os.Stat(partialDir); err != nil || !info.IsDir() {
		t.Fatalf("partial run dir missing: info=%#v err=%v", info, err)
	}

	recoveryStore := testStore(now.Add(time.Minute), "mock-run-record-retry")
	recovered := recoveryStore.StartMockRun(MockRunCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-mock-record-retry"})
	if !recovered.OK || recovered.Run == nil {
		t.Fatalf("second StartMockRun() = %#v, want recovered attempt", recovered)
	}
	if recovered.Run.Attempt != 2 {
		t.Fatalf("recovered attempt = %d, want 2", recovered.Run.Attempt)
	}
	if recovered.Run.RunID == "run_mock_shot_shot_002_001" {
		t.Fatalf("recovered run reused partial run id %q", recovered.Run.RunID)
	}
}

func TestMockRunRecordJSONRoundTrip(t *testing.T) {
	run := MockRunDTO{
		SchemaVersion: CurrentSchemaVersion,
		RunID:         "run_mock_test_001",
		ProjectID:     "proj",
		ShotID:        "shot_002",
		SelectionIDs:  []string{"shot_002"},
		TaskMode:      MockRunProviderMode,
		ProviderMode:  MockRunProviderMode,
		ContextDigest: "sha256:test",
		Status:        MockRunStatusCompleted,
		Attempt:       1,
		RunPath:       "prompts/runs/run_mock_test_001/run.json",
		EventsPath:    "prompts/runs/run_mock_test_001/events.jsonl",
		CreatedAt:     "2026-05-20T15:20:00Z",
		UpdatedAt:     "2026-05-20T15:20:00Z",
	}
	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	var decoded MockRunDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal run: %v", err)
	}
	if decoded.ProviderMode != MockRunProviderMode || decoded.Status != MockRunStatusCompleted {
		t.Fatalf("decoded run = %#v, want mock completed", decoded)
	}
}
