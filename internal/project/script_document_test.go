package project

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScriptDocumentSavePreservesRawText(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
	store := testStore(now, "script-preserve")
	root := createScriptDocumentProject(t)
	rawText := "INT. MARKET - NIGHT\n\nMina:  Keep the lantern lit.\n  \n[beat]\n"

	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		ScriptID:      "main_script",
		Title:         "Market Script",
		RawText:       rawText,
		Logline:       "A lantern keeps a promise alive.",
		Synopsis:      "Mina protects the market route.",
		CorrelationID: "corr-script-save",
	})

	if !result.OK {
		t.Fatalf("SaveScriptDocument() failed: %#v", result.Error)
	}
	if result.Document == nil {
		t.Fatal("SaveScriptDocument() document is nil")
	}
	if result.Document.RawText != rawText {
		t.Fatalf("raw text changed\nwant: %q\n got: %q", rawText, result.Document.RawText)
	}
	if result.Document.ProjectID != "proj_alpha_001" {
		t.Fatalf("project id = %q, want proj_alpha_001", result.Document.ProjectID)
	}
	if result.Document.Title != "Market Script" || result.Document.Logline == "" || result.Document.Synopsis == "" {
		t.Fatalf("document metadata not preserved: %#v", result.Document)
	}
	if result.Document.CreatedAt != now.Format(time.RFC3339) || result.Document.UpdatedAt != now.Format(time.RFC3339) {
		t.Fatalf("timestamps = %s/%s, want %s", result.Document.CreatedAt, result.Document.UpdatedAt, now.Format(time.RFC3339))
	}
	if len(result.Document.Scenes) != 0 {
		t.Fatalf("scenes = %#v, want no scene generation in TOO-172", result.Document.Scenes)
	}

	reopened := store.LoadScriptDocument(LoadScriptDocumentCommand{Root: root, ScriptID: "main_script"})
	if !reopened.OK || reopened.Document == nil {
		t.Fatalf("LoadScriptDocument() failed: %#v", reopened.Error)
	}
	if reopened.Document.RawText != rawText {
		t.Fatalf("reopened raw text changed\nwant: %q\n got: %q", rawText, reopened.Document.RawText)
	}
}

func TestScriptDocumentRejectsEmptyText(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 5, 0, 0, time.UTC), "script-empty")
	root := createScriptDocumentProject(t)

	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		Title:         "Empty Script",
		RawText:       " \n\t",
		CorrelationID: "corr-script-empty",
	})

	if result.OK || result.Error == nil {
		t.Fatalf("SaveScriptDocument() = %#v, want empty script error", result)
	}
	if result.Error.Code != CodeScriptEmpty {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeScriptEmpty)
	}
	if !hasRecoveryAction(result.Error.RecoveryActions, "non-empty script text") {
		t.Fatalf("recovery actions = %#v, want non-empty script text guidance", result.Error.RecoveryActions)
	}
}

func TestScriptDocumentImportsSourceAsset(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 10, 0, 0, time.UTC), "script-asset")
	root := createScriptDocumentProject(t)
	rawText := "SCENE 1\nA painter marks the first frame.\n"
	sourceAssetID := "assets/inputs/source-script.md"
	writeScriptAsset(t, root, sourceAssetID, rawText)

	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		ScriptID:      "",
		SourceAssetID: sourceAssetID,
		CorrelationID: "corr-script-import",
	})

	if !result.OK {
		t.Fatalf("SaveScriptDocument() import failed: %#v", result.Error)
	}
	if result.Document == nil {
		t.Fatal("SaveScriptDocument() document is nil")
	}
	if result.Document.ID != "script_main" {
		t.Fatalf("script id = %q, want default script_main", result.Document.ID)
	}
	if result.Document.Title != "source-script" {
		t.Fatalf("title = %q, want source-script", result.Document.Title)
	}
	if result.Document.SourceAssetID != sourceAssetID || result.Document.RawText != rawText {
		t.Fatalf("imported document = %#v, want source asset and exact raw text", result.Document)
	}
}

func TestScriptDocumentReportsMissingSourceAsset(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 15, 0, 0, time.UTC), "script-missing")
	root := createScriptDocumentProject(t)

	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		SourceAssetID: "assets/inputs/missing-script.md",
		CorrelationID: "corr-script-missing",
	})

	if result.OK || result.Error == nil {
		t.Fatalf("SaveScriptDocument() = %#v, want missing asset error", result)
	}
	if result.Error.Code != CodeScriptAssetMissing {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeScriptAssetMissing)
	}
}

func TestScriptDocumentSaveAndReopenPreservesScenes(t *testing.T) {
	firstStore := testStore(time.Date(2026, 5, 20, 9, 20, 0, 0, time.UTC), "script-reopen-first")
	root := createScriptDocumentProject(t)
	first := firstStore.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		ScriptID:      "script_main",
		Title:         "First Draft",
		RawText:       "First draft\n",
		CorrelationID: "corr-first",
	})
	if !first.OK || first.Document == nil {
		t.Fatalf("initial SaveScriptDocument() failed: %#v", first.Error)
	}

	first.Document.Scenes = []ScriptSceneDTO{{
		ID:     "scene_001",
		Index:  1,
		Title:  "Manual scene",
		Action: "Scene work stays outside TOO-172.",
	}}
	writeJSON(t, scriptDocumentPath(root, "script_main"), first.Document)

	secondStore := testStore(time.Date(2026, 5, 20, 9, 25, 0, 0, time.UTC), "script-reopen-second")
	second := secondStore.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		ScriptID:      "script_main",
		Title:         "Second Draft",
		RawText:       "Second draft\n",
		Logline:       "Updated line.",
		CorrelationID: "corr-second",
	})
	if !second.OK || second.Document == nil {
		t.Fatalf("second SaveScriptDocument() failed: %#v", second.Error)
	}
	if second.Document.CreatedAt != first.Document.CreatedAt {
		t.Fatalf("createdAt = %q, want previous %q", second.Document.CreatedAt, first.Document.CreatedAt)
	}
	if second.Document.UpdatedAt != time.Date(2026, 5, 20, 9, 25, 0, 0, time.UTC).Format(time.RFC3339) {
		t.Fatalf("updatedAt = %q, want second store time", second.Document.UpdatedAt)
	}
	if len(second.Document.Scenes) != 1 || second.Document.Scenes[0].ID != "scene_001" {
		t.Fatalf("scenes = %#v, want previous manual scene preserved", second.Document.Scenes)
	}

	reopened := secondStore.LoadScriptDocument(LoadScriptDocumentCommand{Root: root})
	if !reopened.OK || reopened.Document == nil || reopened.Document.RawText != "Second draft\n" {
		t.Fatalf("LoadScriptDocument() = %#v, want latest raw text", reopened)
	}
}

func TestScriptDocumentSaveFailureKeepsPreviousFile(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 30, 0, 0, time.UTC)
	root := createScriptDocumentProject(t)
	store := testStore(now, "script-save-before")
	before := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		ScriptID:      "script_main",
		Title:         "Before",
		RawText:       "Before text\n",
		CorrelationID: "corr-before",
	})
	if !before.OK {
		t.Fatalf("initial SaveScriptDocument() failed: %#v", before.Error)
	}
	beforeData, err := os.ReadFile(scriptDocumentPath(root, "script_main"))
	if err != nil {
		t.Fatalf("read previous script document: %v", err)
	}

	failingStore := NewStore(StoreOptions{
		Now:        func() time.Time { return now.Add(time.Minute) },
		InstanceID: "script-save-failure",
		Hostname:   "test-host",
		PID:        os.Getpid(),
		AtomicWrite: func(filename string, data []byte, perm fs.FileMode, instanceID string) error {
			return errors.New("injected script document write failure")
		},
	})
	failed := failingStore.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		ScriptID:      "script_main",
		Title:         "After",
		RawText:       "After text\n",
		CorrelationID: "corr-fail",
	})
	if failed.OK || failed.Error == nil {
		t.Fatalf("SaveScriptDocument() = %#v, want injected failure", failed)
	}
	if failed.Error.Code != CodeScriptSaveFailed {
		t.Fatalf("error code = %q, want %s", failed.Error.Code, CodeScriptSaveFailed)
	}

	afterData, err := os.ReadFile(scriptDocumentPath(root, "script_main"))
	if err != nil {
		t.Fatalf("read script document after failure: %v", err)
	}
	if string(afterData) != string(beforeData) {
		t.Fatalf("script document changed after failed save\nbefore:\n%s\nafter:\n%s", beforeData, afterData)
	}
}

func TestScriptDocumentSaveRefusesActiveLock(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 32, 0, 0, time.UTC)
	root := createScriptDocumentProject(t)
	writeLock(t, root, LockMetadata{
		AppInstanceID: "other-instance",
		OpenedAt:      now.Format(time.RFC3339),
		HeartbeatAt:   now.Format(time.RFC3339),
		PID:           os.Getpid(),
		Host:          "test-host",
	})

	store := testStore(now, "script-lock")
	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		RawText:       "Locked write\n",
		CorrelationID: "corr-lock",
	})

	if result.OK || result.Error == nil {
		t.Fatalf("SaveScriptDocument() = %#v, want active lock error", result)
	}
	if result.Error.Code != CodeProjectActiveLock {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeProjectActiveLock)
	}
}

func TestScriptDocumentSaveRefusesStaleLockWithoutChangingFile(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 32, 30, 0, time.UTC)
	root := createScriptDocumentProject(t)
	store := testStore(now, "script-stale-before")
	before := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		RawText:       "Before stale lock\n",
		CorrelationID: "corr-stale-before",
	})
	if !before.OK {
		t.Fatalf("initial SaveScriptDocument() failed: %#v", before.Error)
	}
	beforeData, err := os.ReadFile(scriptDocumentPath(root, "script_main"))
	if err != nil {
		t.Fatalf("read previous script document: %v", err)
	}

	writeLock(t, root, LockMetadata{
		AppInstanceID: "old-instance",
		OpenedAt:      now.Add(-10 * time.Minute).Format(time.RFC3339),
		HeartbeatAt:   now.Add(-10 * time.Minute).Format(time.RFC3339),
		PID:           -1,
		Host:          "test-host",
	})

	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		RawText:       "After stale lock\n",
		CorrelationID: "corr-stale",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("SaveScriptDocument() = %#v, want stale lock error", result)
	}
	if result.Error.Code != CodeProjectStaleLock {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeProjectStaleLock)
	}
	afterData, err := os.ReadFile(scriptDocumentPath(root, "script_main"))
	if err != nil {
		t.Fatalf("read script document after stale lock refusal: %v", err)
	}
	if string(afterData) != string(beforeData) {
		t.Fatalf("script document changed after stale lock refusal\nbefore:\n%s\nafter:\n%s", beforeData, afterData)
	}
}

func TestScriptDocumentSaveRefusesCorruptExistingDocument(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 33, 0, 0, time.UTC), "script-corrupt")
	root := createScriptDocumentProject(t)
	filename := scriptDocumentPath(root, "script_main")
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatalf("create script document dir: %v", err)
	}
	corrupt := []byte(`{"id":"script_main","scenes":[`)
	if err := os.WriteFile(filename, corrupt, 0o644); err != nil {
		t.Fatalf("write corrupt script document: %v", err)
	}

	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		RawText:       "New script text\n",
		CorrelationID: "corr-corrupt",
	})

	if result.OK || result.Error == nil {
		t.Fatalf("SaveScriptDocument() = %#v, want corrupt existing document error", result)
	}
	if result.Error.Code != CodeScriptLoadFailed {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeScriptLoadFailed)
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read script document after refused save: %v", err)
	}
	if string(after) != string(corrupt) {
		t.Fatalf("corrupt document changed after refused save\nbefore:\n%s\nafter:\n%s", corrupt, after)
	}
}

func TestScriptDocumentRejectsUnsafeIDAndAssetPath(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 35, 0, 0, time.UTC), "script-paths")
	root := createScriptDocumentProject(t)

	unsafeID := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:     root,
		ScriptID: "../escape",
		RawText:  "Text\n",
	})
	if unsafeID.OK || unsafeID.Error == nil || unsafeID.Error.Code != CodeScriptInvalidID {
		t.Fatalf("unsafe id result = %#v, want %s", unsafeID, CodeScriptInvalidID)
	}

	unsafeAsset := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		SourceAssetID: "../outside.md",
	})
	if unsafeAsset.OK || unsafeAsset.Error == nil || unsafeAsset.Error.Code != CodeScriptAssetMissing {
		t.Fatalf("unsafe asset result = %#v, want %s", unsafeAsset, CodeScriptAssetMissing)
	}

	unsafeAssetWithRawText := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:          root,
		RawText:       "Raw text should not bypass sourceAssetId validation.\n",
		SourceAssetID: "../outside.md",
	})
	if unsafeAssetWithRawText.OK || unsafeAssetWithRawText.Error == nil || unsafeAssetWithRawText.Error.Code != CodeScriptAssetMissing {
		t.Fatalf("unsafe asset with raw text result = %#v, want %s", unsafeAssetWithRawText, CodeScriptAssetMissing)
	}
}

func createScriptDocumentProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "script-project")
	if err := CreateBaselineProject(root, testManifest(t)); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}
	return root
}

func writeScriptAsset(t *testing.T, root string, sourceAssetID string, rawText string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(sourceAssetID))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatalf("create source asset dir: %v", err)
	}
	if err := os.WriteFile(filename, []byte(rawText), 0o644); err != nil {
		t.Fatalf("write source asset: %v", err)
	}
}

func hasRecoveryAction(actions []string, needle string) bool {
	for _, action := range actions {
		if strings.Contains(action, needle) {
			return true
		}
	}
	return false
}
