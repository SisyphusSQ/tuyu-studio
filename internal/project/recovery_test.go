package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProjectRecoveryBadProjectFixtures(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 6, 5, 0, 0, time.UTC), "recovery-fixture")

	invalid := store.HealthReport(badProjectFixtureRoot(t, "invalid-json"))
	requireHealthCode(t, invalid, CodeManifestJSONInvalid)
	if invalid.Status != HealthStatusBlocking {
		t.Fatalf("invalid fixture status = %q, want %s; items=%#v", invalid.Status, HealthStatusBlocking, invalid.Items)
	}

	missing := store.HealthReport(badProjectFixtureRoot(t, "missing-references"))
	requireHealthCode(t, missing, CodeGraphReferenceMissing)
	requireHealthCode(t, missing, CodePackageReferenceMissing)
	requireHealthCode(t, missing, CodeAssetMissing)
	requireHealthCode(t, missing, CodeProjectDirtyShutdown)
	forbidHealthCode(t, missing, CodeProjectDirectoryMissing)
	if missing.Status != HealthStatusWarning {
		t.Fatalf("missing fixture status = %q, want %s; items=%#v", missing.Status, HealthStatusWarning, missing.Items)
	}
}

func TestHealthCheckMissingAssetAndPackageReferences(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 6, 10, 0, 0, time.UTC), "missing-reference")
	root := filepath.Join(t.TempDir(), "missing-reference")
	manifest := testManifest(t)
	manifest.Graph.Nodes = []Node{
		{ID: "node_missing_character", Kind: "character", RefID: "characters/missing-character.json", Status: "missing"},
		{ID: "node_broken_package", Kind: "package", RefID: "packages/scene-001/broken-package/manifest.json", Status: "draft"},
	}
	manifest.Graph.Edges = []Edge{
		{ID: "edge_missing_endpoint", Source: "node_missing_character", Target: "node_missing_package", Kind: "broken_target"},
	}
	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}

	writeJSON(t, filepath.Join(root, "packages", "scene-001", "broken-package", "manifest.json"), map[string]any{
		"schemaVersion":           CurrentSchemaVersion,
		"projectId":               manifest.Project.ID,
		"sceneId":                 "scene_001",
		"shotId":                  "shot_missing",
		"packageId":               "pkg_missing_refs",
		"promptPath":              "prompt-directory",
		"continuityPath":          "continuity-missing.md",
		"uploadChecklistPath":     "upload-checklist-missing.md",
		"references":              []string{"references/ref-missing.txt"},
		"generationPackageStatus": "draft",
	})
	if err := os.Mkdir(filepath.Join(root, "packages", "scene-001", "broken-package", "prompt-directory"), 0o755); err != nil {
		t.Fatalf("create package directory reference: %v", err)
	}
	writeJSON(t, filepath.Join(root, filepath.FromSlash(DigestIndexRelativePath)), DigestIndex{Files: []DigestEntry{{
		Path:            "assets/refs/ref-missing.txt",
		SHA256:          strings.Repeat("0", 64),
		AffectedObjects: []string{"shot_missing", "pkg_missing_refs"},
	}}})

	health := store.HealthReport(root)
	requireHealthCode(t, health, CodeGraphReferenceMissing)
	requireHealthCode(t, health, CodePackageReferenceMissing)
	requireHealthCode(t, health, CodeAssetMissing)

	item := findHealthCode(t, health, CodePackageReferenceMissing)
	if len(item.AffectedObjects) == 0 {
		t.Fatalf("package reference health item missing affected objects: %#v", item)
	}
	if len(item.RecoveryActions) == 0 {
		t.Fatalf("package reference health item missing recovery actions: %#v", item)
	}
	directoryItem := findHealthPath(t, health, "packages/scene-001/broken-package/prompt-directory")
	if directoryItem.Code != CodePackageReferenceMissing {
		t.Fatalf("directory package reference item = %#v, want %s", directoryItem, CodePackageReferenceMissing)
	}
}

func TestHealthCheckPathLeak(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 6, 15, 0, 0, time.UTC), "path-leak")
	root := filepath.Join(t.TempDir(), "path-leak")
	manifest := testManifest(t)
	manifest.Graph.Nodes = []Node{{
		ID:     "node_unsafe_package",
		Kind:   "package",
		RefID:  "packages/scene-001/unsafe-package/manifest.json",
		Status: "draft",
	}}
	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}
	writeJSON(t, filepath.Join(root, "packages", "scene-001", "unsafe-package", "manifest.json"), map[string]any{
		"schemaVersion":           CurrentSchemaVersion,
		"projectId":               manifest.Project.ID,
		"sceneId":                 "scene_unsafe",
		"shotId":                  "shot_unsafe",
		"packageId":               "pkg_unsafe",
		"promptPath":              "file://external/prompt.txt",
		"continuityPath":          "continuity.md",
		"uploadChecklistPath":     "upload_checklist.md",
		"references":              []string{"~/secret/ref.txt"},
		"diagnosticNote":          "token=redacted-placeholder",
		"generationPackageStatus": "draft",
	})

	health := store.HealthReport(root)
	requireHealthCode(t, health, CodePathURL)
	requireHealthCode(t, health, CodePathAbsolute)
	requireHealthCode(t, health, CodeSensitiveValue)
	if health.Status != HealthStatusBlocking {
		t.Fatalf("path leak status = %q, want %s; items=%#v", health.Status, HealthStatusBlocking, health.Items)
	}
}

func TestHealthCheckDigestMismatch(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 6, 20, 0, 0, time.UTC), "digest-mismatch")
	root := filepath.Join(t.TempDir(), "digest-mismatch")
	manifest := testManifest(t)
	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "refs", "ref.txt"), []byte("actual"), 0o644); err != nil {
		t.Fatalf("write ref: %v", err)
	}
	writeJSON(t, filepath.Join(root, filepath.FromSlash(DigestIndexRelativePath)), DigestIndex{Files: []DigestEntry{{
		Path:            "assets/refs/ref.txt",
		SHA256:          strings.Repeat("0", 64),
		AffectedObjects: []string{"shot_001", "package_001"},
	}}})

	health := store.HealthReport(root)
	item := findHealthCode(t, health, CodeDigestMismatch)
	if item.Severity != SeverityBlocking {
		t.Fatalf("digest mismatch severity = %q, want %s", item.Severity, SeverityBlocking)
	}
	if len(item.AffectedObjects) != 2 {
		t.Fatalf("digest mismatch affected objects = %#v", item.AffectedObjects)
	}
}

func TestProjectLockRecoveryAudit(t *testing.T) {
	now := time.Date(2026, 5, 20, 6, 25, 0, 0, time.UTC)
	root := filepath.Join(t.TempDir(), "lock-recovery")
	manifest := testManifest(t)
	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}
	writeLock(t, root, LockMetadata{
		AppInstanceID: "old-instance",
		OpenedAt:      now.Add(-10 * time.Minute).Format(time.RFC3339),
		HeartbeatAt:   now.Add(-10 * time.Minute).Format(time.RFC3339),
		PID:           -1,
		Host:          "test-host",
	})

	store := testStore(now, "new-instance")
	result := store.OpenProject(OpenProjectCommand{
		Root:          root,
		Takeover:      true,
		CorrelationID: "corr-recovery-takeover",
	})
	if !result.OK {
		t.Fatalf("takeover open failed: %#v", result.Error)
	}
	if result.Summary == nil || result.Summary.LockState != LockStateTakeover {
		t.Fatalf("takeover summary = %#v, want takeover", result.Summary)
	}

	projectAudit := readTextFile(t, filepath.Join(root, filepath.FromSlash(AuditEventsRelativePath)))
	recoveryAudit := readTextFile(t, filepath.Join(root, filepath.FromSlash(RecoveryAuditRelativePath)))
	for path, content := range map[string]string{
		AuditEventsRelativePath:   projectAudit,
		RecoveryAuditRelativePath: recoveryAudit,
	} {
		if !strings.Contains(content, "project.lock_takeover") {
			t.Fatalf("%s missing takeover event:\n%s", path, content)
		}
		assertNoSensitiveText(t, path, content)
	}

	var entry auditEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(recoveryAudit)), &entry); err != nil {
		t.Fatalf("decode recovery audit: %v", err)
	}
	if entry.Details["action"] != "replace_session_lock" {
		t.Fatalf("recovery audit details = %#v", entry.Details)
	}
}

func badProjectFixtureRoot(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join("..", "..", "examples", "bad-projects", name)
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("bad project fixture %q missing: %v", name, err)
	}
	return root
}

func requireHealthCode(t *testing.T, report HealthReport, code string) {
	t.Helper()
	if !hasHealthCode(report, code) {
		t.Fatalf("health codes missing %s: %#v", code, report.Items)
	}
}

func forbidHealthCode(t *testing.T, report HealthReport, code string) {
	t.Helper()
	if hasHealthCode(report, code) {
		t.Fatalf("health codes unexpectedly contain %s: %#v", code, report.Items)
	}
}

func findHealthCode(t *testing.T, report HealthReport, code string) HealthItem {
	t.Helper()
	for _, item := range report.Items {
		if item.Code == code {
			return item
		}
	}
	t.Fatalf("health codes missing %s: %#v", code, report.Items)
	return HealthItem{}
}

func findHealthPath(t *testing.T, report HealthReport, path string) HealthItem {
	t.Helper()
	for _, item := range report.Items {
		if item.Path == path {
			return item
		}
	}
	t.Fatalf("health paths missing %s: %#v", path, report.Items)
	return HealthItem{}
}

func readTextFile(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	return string(data)
}

func assertNoSensitiveText(t *testing.T, label string, text string) {
	t.Helper()
	for _, forbidden := range []string{"/Users/", "/var/folders", "api_key=", "token=", "secret=", "credential="} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) {
			t.Fatalf("%s contains forbidden sensitive text %q: %s", label, forbidden, text)
		}
	}
}
