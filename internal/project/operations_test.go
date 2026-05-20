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

func TestProjectCreateOpenSaveBaseline(t *testing.T) {
	now := time.Date(2026, 5, 20, 5, 15, 0, 0, time.UTC)
	store := testStore(now, "test-instance")
	root := filepath.Join(t.TempDir(), "alpha-project")

	created := store.CreateProject(CreateProjectCommand{
		Root:          root,
		ProjectID:     "proj_ops_001",
		Name:          "Ops Project",
		CorrelationID: "corr-create",
	})
	if !created.OK {
		t.Fatalf("CreateProject() failed: %#v", created.Error)
	}
	if created.Summary == nil || created.Summary.OpenMode != OpenModeReadWrite {
		t.Fatalf("created summary = %#v, want read_write", created.Summary)
	}
	if created.Summary.GraphVersion != 1 {
		t.Fatalf("created graph version = %d, want 1", created.Summary.GraphVersion)
	}

	for _, dir := range RequiredProjectDirectories() {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil || !info.IsDir() {
			t.Fatalf("required directory %q missing or not a directory: %v", dir, err)
		}
	}

	opened := store.OpenProject(OpenProjectCommand{Root: root, CorrelationID: "corr-open"})
	if !opened.OK {
		t.Fatalf("OpenProject() failed: %#v", opened.Error)
	}
	if opened.Summary == nil || opened.Summary.LockState != LockStateOwned {
		t.Fatalf("opened summary = %#v, want owned lock", opened.Summary)
	}

	saved := store.SaveProject(SaveProjectCommand{Root: root, CorrelationID: "corr-save"})
	if !saved.OK {
		t.Fatalf("SaveProject() failed: %#v", saved.Error)
	}

	manifest, report, err := store.readManifest(root)
	if err != nil || report.HasBlocking() {
		t.Fatalf("saved manifest invalid: err=%v report=%#v", err, report.Items)
	}
	if manifest.Graph.Version != 2 {
		t.Fatalf("saved graph version = %d, want 2", manifest.Graph.Version)
	}
	if !manifest.Integrity.LastCleanShutdown {
		t.Fatal("saved manifest must be clean")
	}
	if manifest.Integrity.LastGraphVersion != manifest.Graph.Version {
		t.Fatalf("lastGraphVersion = %d, want graph version %d", manifest.Integrity.LastGraphVersion, manifest.Graph.Version)
	}
}

func TestProjectOpenReportsDistinctManifestErrors(t *testing.T) {
	now := time.Date(2026, 5, 20, 5, 20, 0, 0, time.UTC)
	store := testStore(now, "test-instance")

	invalidRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(invalidRoot, ManifestFileName), []byte(`{"schemaVersion":`), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	invalid := store.OpenProject(OpenProjectCommand{Root: invalidRoot, CorrelationID: "corr-invalid"})
	if invalid.OK || invalid.Error == nil {
		t.Fatalf("invalid open result = %#v, want error", invalid)
	}
	if invalid.Error.Code != CodeProjectManifestInvalid {
		t.Fatalf("invalid error code = %q, want %s", invalid.Error.Code, CodeProjectManifestInvalid)
	}
	if invalid.Health == nil || !hasHealthCode(*invalid.Health, CodeManifestJSONInvalid) {
		t.Fatalf("invalid health = %#v, want %s", invalid.Health, CodeManifestJSONInvalid)
	}

	unsupportedRoot := t.TempDir()
	manifest := testManifest(t)
	manifest.SchemaVersion = "9.9.9"
	writeManifestBypassValidation(t, unsupportedRoot, manifest)

	unsupported := store.OpenProject(OpenProjectCommand{Root: unsupportedRoot, CorrelationID: "corr-unsupported"})
	if unsupported.OK || unsupported.Error == nil {
		t.Fatalf("unsupported open result = %#v, want error", unsupported)
	}
	if unsupported.Error.Code != CodeSchemaUnsupported {
		t.Fatalf("unsupported error code = %q, want %s", unsupported.Error.Code, CodeSchemaUnsupported)
	}
}

func TestProjectSavePreservesPreviousManifestOnCommitFailure(t *testing.T) {
	now := time.Date(2026, 5, 20, 5, 25, 0, 0, time.UTC)
	root := filepath.Join(t.TempDir(), "save-failure")
	store := testStore(now, "test-instance")

	created := store.CreateProject(CreateProjectCommand{
		Root:          root,
		ProjectID:     "proj_save_failure",
		Name:          "Save Failure",
		CorrelationID: "corr-create",
	})
	if !created.OK {
		t.Fatalf("CreateProject() failed: %#v", created.Error)
	}
	before, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		t.Fatalf("read manifest before save: %v", err)
	}

	failingStore := NewStore(StoreOptions{
		Now:        func() time.Time { return now.Add(time.Minute) },
		InstanceID: "test-instance",
		Hostname:   "test-host",
		PID:        os.Getpid(),
		AtomicWrite: func(filename string, data []byte, perm fs.FileMode, instanceID string) error {
			return errors.New("injected rename failure")
		},
	})

	result := failingStore.SaveProject(SaveProjectCommand{Root: root, CorrelationID: "corr-save-fail"})
	if result.OK || result.Error == nil {
		t.Fatalf("SaveProject() = %#v, want injected failure", result)
	}
	if result.Error.Code != CodeProjectSaveCommit {
		t.Fatalf("save error code = %q, want %s", result.Error.Code, CodeProjectSaveCommit)
	}

	after, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		t.Fatalf("read manifest after failed save: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("manifest changed after failed save\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestProjectLockStatesAndTakeover(t *testing.T) {
	now := time.Date(2026, 5, 20, 5, 30, 0, 0, time.UTC)
	root := filepath.Join(t.TempDir(), "lock-project")
	manifest := testManifest(t)
	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}

	writeLock(t, root, LockMetadata{
		AppInstanceID: "other-instance",
		OpenedAt:      now.Format(time.RFC3339),
		HeartbeatAt:   now.Format(time.RFC3339),
		PID:           os.Getpid(),
		Host:          "test-host",
	})

	store := testStore(now, "test-instance")
	active := store.OpenProject(OpenProjectCommand{Root: root, CorrelationID: "corr-active"})
	if !active.OK {
		t.Fatalf("active lock open should return read-only summary, got error %#v", active.Error)
	}
	if active.Summary == nil || active.Summary.OpenMode != OpenModeReadOnly || active.Summary.LockState != LockStateActive {
		t.Fatalf("active lock summary = %#v, want read_only active", active.Summary)
	}
	if active.Health == nil || !hasHealthCode(*active.Health, CodeProjectActiveLock) {
		t.Fatalf("active lock health = %#v, want %s", active.Health, CodeProjectActiveLock)
	}

	writeLock(t, root, LockMetadata{
		AppInstanceID: "aged-live-instance",
		OpenedAt:      now.Add(-10 * time.Minute).Format(time.RFC3339),
		HeartbeatAt:   now.Add(-10 * time.Minute).Format(time.RFC3339),
		PID:           os.Getpid(),
		Host:          "test-host",
	})
	agedLive := store.OpenProject(OpenProjectCommand{Root: root, CorrelationID: "corr-aged-live"})
	if !agedLive.OK {
		t.Fatalf("aged live lock open should return read-only summary, got error %#v", agedLive.Error)
	}
	if agedLive.Summary == nil || agedLive.Summary.LockState != LockStateActive || agedLive.Summary.OpenMode != OpenModeReadOnly {
		t.Fatalf("aged live lock summary = %#v, want active read_only", agedLive.Summary)
	}

	save := store.SaveProject(SaveProjectCommand{Root: root, CorrelationID: "corr-save-active"})
	if save.OK || save.Error == nil || save.Error.Code != CodeProjectActiveLock {
		t.Fatalf("save with active lock = %#v, want %s", save, CodeProjectActiveLock)
	}

	writeLock(t, root, LockMetadata{
		AppInstanceID: "old-instance",
		OpenedAt:      now.Add(-10 * time.Minute).Format(time.RFC3339),
		HeartbeatAt:   now.Add(-10 * time.Minute).Format(time.RFC3339),
		PID:           -1,
		Host:          "test-host",
	})

	takeover := store.OpenProject(OpenProjectCommand{Root: root, Takeover: true, CorrelationID: "corr-takeover"})
	if !takeover.OK {
		t.Fatalf("takeover open failed: %#v", takeover.Error)
	}
	if takeover.Summary == nil || takeover.Summary.LockState != LockStateTakeover || takeover.Summary.OpenMode != OpenModeReadWrite {
		t.Fatalf("takeover summary = %#v, want takeover read_write", takeover.Summary)
	}

	audit, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(AuditEventsRelativePath)))
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if !strings.Contains(string(audit), "project.lock_takeover") {
		t.Fatalf("audit does not contain takeover event:\n%s", string(audit))
	}
}

func TestProjectLockTakeoverRequiresWritableAudit(t *testing.T) {
	now := time.Date(2026, 5, 20, 5, 32, 0, 0, time.UTC)
	root := filepath.Join(t.TempDir(), "audit-blocked")
	manifest := testManifest(t)
	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}

	previous := LockMetadata{
		AppInstanceID: "old-instance",
		OpenedAt:      now.Add(-10 * time.Minute).Format(time.RFC3339),
		HeartbeatAt:   now.Add(-10 * time.Minute).Format(time.RFC3339),
		PID:           -1,
		Host:          "test-host",
	}
	writeLock(t, root, previous)
	if err := os.Mkdir(filepath.Join(root, filepath.FromSlash(AuditEventsRelativePath)), 0o755); err != nil {
		t.Fatalf("create audit path directory: %v", err)
	}

	store := testStore(now, "test-instance")
	result := store.OpenProject(OpenProjectCommand{Root: root, Takeover: true, CorrelationID: "corr-audit-blocked"})
	if result.OK || result.Error == nil {
		t.Fatalf("takeover with blocked audit = %#v, want error", result)
	}
	if result.Error.Code != CodeProjectLockWriteFailed {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeProjectLockWriteFailed)
	}

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(LockFileRelativePath)))
	if err != nil {
		t.Fatalf("read lock after failed takeover: %v", err)
	}
	var after LockMetadata
	if err := json.Unmarshal(data, &after); err != nil {
		t.Fatalf("decode lock after failed takeover: %v", err)
	}
	if after.AppInstanceID != previous.AppInstanceID {
		t.Fatalf("lock was replaced despite audit failure: %#v", after)
	}
}

func TestHealthCheckReportsBaselineIssues(t *testing.T) {
	now := time.Date(2026, 5, 20, 5, 35, 0, 0, time.UTC)
	store := testStore(now, "test-instance")

	t.Run("missing graph partition", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ManifestFileName), []byte(`{"schemaVersion":"1.0.0"}`), 0o644); err != nil {
			t.Fatalf("write manifest: %v", err)
		}
		health := store.HealthReport(root)
		if !hasHealthCode(health, CodeGraphPartitionMissing) {
			t.Fatalf("health = %#v, want %s", health.Items, CodeGraphPartitionMissing)
		}
	})

	t.Run("absolute path leakage", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "absolute-path")
		manifest := testManifest(t)
		if err := CreateBaselineProject(root, manifest); err != nil {
			t.Fatalf("CreateBaselineProject() error = %v", err)
		}
		manifest.Paths.AssetRefs = "/private/project/ref.png"
		writeManifestBypassValidation(t, root, manifest)

		health := store.HealthReport(root)
		if !hasHealthCode(health, CodePathAbsolute) {
			t.Fatalf("health = %#v, want %s", health.Items, CodePathAbsolute)
		}
	})

	t.Run("dirty shutdown", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "dirty")
		manifest := testManifest(t)
		if err := CreateBaselineProject(root, manifest); err != nil {
			t.Fatalf("CreateBaselineProject() error = %v", err)
		}
		manifest.Integrity.LastCleanShutdown = false
		writeManifestBypassValidation(t, root, manifest)

		health := store.HealthReport(root)
		if !hasHealthCode(health, CodeProjectDirtyShutdown) {
			t.Fatalf("health = %#v, want %s", health.Items, CodeProjectDirtyShutdown)
		}
	})

	t.Run("digest mismatch", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "digest")
		manifest := testManifest(t)
		if err := CreateBaselineProject(root, manifest); err != nil {
			t.Fatalf("CreateBaselineProject() error = %v", err)
		}
		refPath := filepath.Join(root, "assets", "refs", "ref.txt")
		if err := os.WriteFile(refPath, []byte("actual"), 0o644); err != nil {
			t.Fatalf("write ref: %v", err)
		}
		index := DigestIndex{Files: []DigestEntry{{
			Path:            "assets/refs/ref.txt",
			SHA256:          strings.Repeat("0", 64),
			AffectedObjects: []string{"shot_001", "package_001"},
		}}}
		writeJSON(t, filepath.Join(root, filepath.FromSlash(DigestIndexRelativePath)), index)

		health := store.HealthReport(root)
		if !hasHealthCode(health, CodeDigestMismatch) {
			t.Fatalf("health = %#v, want %s", health.Items, CodeDigestMismatch)
		}
	})

	t.Run("manifest read failure", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ManifestFileName), 0o755); err != nil {
			t.Fatalf("create manifest directory: %v", err)
		}

		health := store.HealthReport(root)
		if !hasHealthCode(health, CodeHealthCheckIncomplete) {
			t.Fatalf("health = %#v, want %s", health.Items, CodeHealthCheckIncomplete)
		}
	})
}

func testStore(now time.Time, instanceID string) *Store {
	return NewStore(StoreOptions{
		Now:        func() time.Time { return now },
		InstanceID: instanceID,
		Hostname:   "test-host",
		PID:        os.Getpid(),
		StaleAfter: time.Minute,
	})
}

func writeManifestBypassValidation(t *testing.T, root string, manifest Manifest) {
	t.Helper()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("create root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ManifestFileName), data, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func writeLock(t *testing.T, root string, lock LockMetadata) {
	t.Helper()
	writeJSON(t, filepath.Join(root, filepath.FromSlash(LockFileRelativePath)), lock)
}

func writeJSON(t *testing.T, filename string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatalf("create JSON dir: %v", err)
	}
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		t.Fatalf("write JSON: %v", err)
	}
}

func hasHealthCode(report HealthReport, code string) bool {
	for _, item := range report.Items {
		if item.Code == code {
			return true
		}
	}
	return false
}
