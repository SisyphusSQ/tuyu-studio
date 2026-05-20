package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScriptSceneConfirmValidatesRangeRequiredFieldsAndOverlap(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC), "scene-confirm")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "line one\nline two\nline three\nline four\n")

	missing := store.ConfirmScriptScene(ConfirmScriptSceneCommand{
		Root:        root,
		Location:    "Market",
		Action:      "Mina enters.",
		SourceRange: ScriptSourceRange{StartLine: 1, EndLine: 2},
	})
	if missing.OK || missing.Error == nil || missing.Error.Code != CodeSceneRequiredFieldMissing {
		t.Fatalf("missing required result = %#v, want %s", missing, CodeSceneRequiredFieldMissing)
	}

	outOfBounds := store.ConfirmScriptScene(ConfirmScriptSceneCommand{
		Root:        root,
		Title:       "Opening",
		Location:    "Market",
		Action:      "Mina enters.",
		SourceRange: ScriptSourceRange{StartLine: 1, EndLine: 8},
	})
	if outOfBounds.OK || outOfBounds.Error == nil || outOfBounds.Error.Code != CodeSourceRangeInvalid {
		t.Fatalf("out-of-bounds result = %#v, want %s", outOfBounds, CodeSourceRangeInvalid)
	}

	confirmed := store.ConfirmScriptScene(ConfirmScriptSceneCommand{
		Root:          root,
		SceneID:       "scene_market",
		Title:         "Opening",
		Location:      "Market",
		TimeOfDay:     "night",
		Characters:    []string{"Mina", "Vendor"},
		Props:         []string{"Lantern"},
		Action:        "Mina enters.",
		EmotionalBeat: "relief",
		SourceRange:   ScriptSourceRange{StartLine: 1, EndLine: 2},
	})
	if !confirmed.OK || confirmed.Document == nil {
		t.Fatalf("ConfirmScriptScene() failed: %#v", confirmed.Error)
	}
	if len(confirmed.Document.Scenes) != 1 {
		t.Fatalf("scenes = %#v, want one confirmed scene", confirmed.Document.Scenes)
	}
	scene := confirmed.Document.Scenes[0]
	if scene.ID != "scene_market" || scene.SourceRange == nil || scene.SourceRange.StartLine != 1 || scene.SourceRange.EndLine != 2 {
		t.Fatalf("confirmed scene = %#v, want source range and id", scene)
	}

	overlap := store.ConfirmScriptScene(ConfirmScriptSceneCommand{
		Root:        root,
		SceneID:     "scene_overlap",
		Title:       "Overlap",
		Location:    "Market",
		Action:      "Mina turns.",
		SourceRange: ScriptSourceRange{StartLine: 2, EndLine: 3},
	})
	if overlap.OK || overlap.Error == nil || overlap.Error.Code != CodeSourceRangeInvalid {
		t.Fatalf("overlap result = %#v, want %s", overlap, CodeSourceRangeInvalid)
	}

	allowed := store.ConfirmScriptScene(ConfirmScriptSceneCommand{
		Root:         root,
		SceneID:      "scene_overlap",
		Title:        "Overlap",
		Location:     "Market",
		Action:       "Mina turns.",
		SourceRange:  ScriptSourceRange{StartLine: 2, EndLine: 3},
		AllowOverlap: true,
	})
	if !allowed.OK || allowed.Document == nil || len(allowed.Document.Scenes) != 2 {
		t.Fatalf("allowed overlap = %#v, want second scene", allowed)
	}
}

func TestShotCandidateConfirmCreatesDraftShotWithLineage(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 10, 0, 0, time.UTC)
	store := testStore(now, "candidate-confirm")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "one\ntwo\nthree\n")
	confirmTestScene(t, store, root, "scene_market", 1, 3)

	saved := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		CandidateID:       "candidate_market_001",
		ScriptSceneID:     "scene_market",
		Index:             1,
		DurationSeconds:   6,
		VisualDescription: "Mina reaches the stall.",
		CharacterRefs: []ShotCharacterRefDTO{{
			CharacterID: "char_mina",
			Name:        "Mina",
			Description: "yellow raincoat",
		}},
		SourceRange: ScriptSourceRange{StartLine: 2, EndLine: 2},
	})
	if !saved.OK || saved.Candidate == nil {
		t.Fatalf("SaveShotCandidate() failed: %#v", saved.Error)
	}
	if _, err := os.Stat(filepath.Join(root, "shots", "shot_market_001.json")); !os.IsNotExist(err) {
		t.Fatalf("shot should not exist before candidate confirm: %v", err)
	}

	confirmed := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_market_001",
		ShotID:      "shot_market_001",
		ConfirmedBy: "tester",
	})
	if !confirmed.OK || confirmed.Shot == nil || confirmed.Candidate == nil {
		t.Fatalf("ConfirmShotCandidate() failed: %#v", confirmed.Error)
	}
	if confirmed.Candidate.Status != ShotCandidateStatusAccepted || confirmed.Candidate.ShotID != "shot_market_001" {
		t.Fatalf("candidate after confirm = %#v, want accepted shot id", confirmed.Candidate)
	}
	if confirmed.Shot.Status != ShotStatusDraft || confirmed.Shot.SourceCandidateID != "candidate_market_001" || confirmed.Shot.ScriptSceneID != "scene_market" {
		t.Fatalf("shot lineage = %#v, want draft shot from candidate and scene", confirmed.Shot)
	}
	if confirmed.Shot.ConfirmedBy != "tester" || confirmed.Shot.ConfirmedAt != now.Format(time.RFC3339) {
		t.Fatalf("confirmed metadata = %#v, want tester/%s", confirmed.Shot, now.Format(time.RFC3339))
	}

	var shot ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_market_001.json"), &shot)
	if shot.ID != "shot_market_001" || shot.Status != ShotStatusDraft || shot.Description != "Mina reaches the stall." {
		t.Fatalf("shot file = %#v, want persisted draft shot", shot)
	}
}

func TestShotCandidateRejectRetryDoesNotCreateShotUntilConfirm(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 10, 20, 0, 0, time.UTC), "candidate-retry")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "one\ntwo\nthree\n")
	confirmTestScene(t, store, root, "scene_market", 1, 3)

	save := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		CandidateID:       "candidate_retry_001",
		ScriptSceneID:     "scene_market",
		Index:             1,
		DurationSeconds:   4,
		VisualDescription: "First version.",
		SourceRange:       ScriptSourceRange{StartLine: 1, EndLine: 1},
	})
	if !save.OK {
		t.Fatalf("SaveShotCandidate() failed: %#v", save.Error)
	}
	rejected := store.RejectShotCandidate(RejectShotCandidateCommand{
		Root:            root,
		CandidateID:     "candidate_retry_001",
		RejectionReason: "needs stronger visual",
	})
	if !rejected.OK || rejected.Candidate == nil || rejected.Candidate.Status != ShotCandidateStatusRejected {
		t.Fatalf("RejectShotCandidate() = %#v, want rejected candidate", rejected)
	}
	blocked := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_retry_001",
		ShotID:      "shot_retry_001",
	})
	if blocked.OK || blocked.Error == nil || blocked.Error.Code != CodeScriptExpansionRowInvalid {
		t.Fatalf("confirm rejected = %#v, want %s", blocked, CodeScriptExpansionRowInvalid)
	}
	if _, err := os.Stat(filepath.Join(root, "shots", "shot_retry_001.json")); !os.IsNotExist(err) {
		t.Fatalf("shot should not exist after rejected confirm: %v", err)
	}

	retry := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		CandidateID:       "candidate_retry_001",
		ScriptSceneID:     "scene_market",
		Index:             1,
		DurationSeconds:   5,
		VisualDescription: "Retry version.",
		SourceRange:       ScriptSourceRange{StartLine: 2, EndLine: 2},
	})
	if !retry.OK || retry.Candidate == nil || retry.Candidate.Status != ShotCandidateStatusCandidate {
		t.Fatalf("retry save = %#v, want candidate status", retry)
	}
	confirmed := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_retry_001",
		ShotID:      "shot_retry_001",
	})
	if !confirmed.OK || confirmed.Shot == nil {
		t.Fatalf("confirm retry = %#v, want shot", confirmed)
	}
}

func TestShotCandidateConfirmBlocksDuplicateAcceptedOutput(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 10, 25, 0, 0, time.UTC), "candidate-conflict")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "one\ntwo\nthree\n")
	confirmTestScene(t, store, root, "scene_market", 1, 3)
	saveTestShotCandidate(t, store, root, "candidate_first_001", "scene_market", 1, 1, 1)
	saveTestShotCandidate(t, store, root, "candidate_duplicate_index", "scene_market", 1, 2, 2)
	saveTestShotCandidate(t, store, root, "candidate_duplicate_shot", "scene_market", 2, 3, 3)

	first := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_first_001",
		ShotID:      "shot_market_001",
	})
	if !first.OK || first.Shot == nil {
		t.Fatalf("first ConfirmShotCandidate() failed: %#v", first.Error)
	}

	duplicateIndex := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_duplicate_index",
		ShotID:      "shot_market_002",
	})
	if duplicateIndex.OK || duplicateIndex.Error == nil || duplicateIndex.Error.Code != CodeShotCandidateConflict {
		t.Fatalf("duplicate index confirm = %#v, want %s", duplicateIndex, CodeShotCandidateConflict)
	}
	if _, err := os.Stat(filepath.Join(root, "shots", "shot_market_002.json")); !os.IsNotExist(err) {
		t.Fatalf("duplicate index should not create shot: %v", err)
	}

	duplicateShot := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_duplicate_shot",
		ShotID:      "shot_market_001",
	})
	if duplicateShot.OK || duplicateShot.Error == nil || duplicateShot.Error.Code != CodeShotCandidateConflict {
		t.Fatalf("duplicate shot confirm = %#v, want %s", duplicateShot, CodeShotCandidateConflict)
	}
}

func TestAcceptedShotCandidateCannotBeRejectedOrReconfirmed(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 10, 26, 0, 0, time.UTC), "candidate-accepted-guard")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "one\ntwo\n")
	confirmTestScene(t, store, root, "scene_market", 1, 2)
	saveTestShotCandidate(t, store, root, "candidate_accepted_001", "scene_market", 1, 1, 1)

	confirmed := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_accepted_001",
		ShotID:      "shot_accepted_001",
	})
	if !confirmed.OK {
		t.Fatalf("ConfirmShotCandidate() failed: %#v", confirmed.Error)
	}
	reconfirm := store.ConfirmShotCandidate(ConfirmShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_accepted_001",
		ShotID:      "shot_accepted_001",
	})
	if reconfirm.OK || reconfirm.Error == nil || reconfirm.Error.Code != CodeShotCandidateConflict {
		t.Fatalf("reconfirm accepted = %#v, want %s", reconfirm, CodeShotCandidateConflict)
	}
	reject := store.RejectShotCandidate(RejectShotCandidateCommand{
		Root:        root,
		CandidateID: "candidate_accepted_001",
	})
	if reject.OK || reject.Error == nil || reject.Error.Code != CodeShotCandidateConflict {
		t.Fatalf("reject accepted = %#v, want %s", reject, CodeShotCandidateConflict)
	}
	list := store.ListShotCandidates(ListShotCandidatesCommand{Root: root})
	if !list.OK || len(list.Candidates) != 1 || list.Candidates[0].Status != ShotCandidateStatusAccepted {
		t.Fatalf("list after accepted guards = %#v, want accepted candidate preserved", list)
	}
}

func TestShotCandidateInvalidRowsPreserveExistingCandidates(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC), "candidate-invalid")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "one\ntwo\nthree\n")
	confirmTestScene(t, store, root, "scene_market", 1, 3)
	valid := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		CandidateID:       "candidate_valid_001",
		ScriptSceneID:     "scene_market",
		Index:             1,
		DurationSeconds:   4,
		VisualDescription: "Valid row.",
		SourceRange:       ScriptSourceRange{StartLine: 1, EndLine: 1},
	})
	if !valid.OK {
		t.Fatalf("valid SaveShotCandidate() failed: %#v", valid.Error)
	}

	invalid := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		CandidateID:       "candidate_invalid_001",
		ScriptSceneID:     "scene_market",
		Index:             2,
		DurationSeconds:   0,
		VisualDescription: "Invalid row.",
		SourceRange:       ScriptSourceRange{StartLine: 2, EndLine: 2},
	})
	if invalid.OK || invalid.Error == nil || invalid.Error.Code != CodeScriptExpansionRowInvalid {
		t.Fatalf("invalid SaveShotCandidate() = %#v, want %s", invalid, CodeScriptExpansionRowInvalid)
	}

	list := store.ListShotCandidates(ListShotCandidatesCommand{Root: root})
	if !list.OK || len(list.Candidates) != 1 || list.Candidates[0].ID != "candidate_valid_001" {
		t.Fatalf("ListShotCandidates() = %#v, want only previous valid candidate", list)
	}
}

func TestShotCandidateConfirmRequiresExistingScene(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 10, 40, 0, 0, time.UTC), "candidate-source")
	root := createScriptDocumentProject(t)
	saveTestScriptDocument(t, store, root, "one\ntwo\n")

	result := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		ScriptSceneID:     "scene_missing",
		Index:             1,
		DurationSeconds:   3,
		VisualDescription: "Missing scene.",
		SourceRange:       ScriptSourceRange{StartLine: 1, EndLine: 1},
	})
	if result.OK || result.Error == nil || result.Error.Code != CodeShotExpansionSourceMissing {
		t.Fatalf("missing scene candidate = %#v, want %s", result, CodeShotExpansionSourceMissing)
	}
}

func saveTestScriptDocument(t *testing.T, store *Store, root string, rawText string) {
	t.Helper()
	result := store.SaveScriptDocument(SaveScriptDocumentCommand{
		Root:    root,
		RawText: rawText,
		Title:   "Test Script",
	})
	if !result.OK {
		t.Fatalf("SaveScriptDocument() failed: %#v", result.Error)
	}
}

func confirmTestScene(t *testing.T, store *Store, root string, sceneID string, startLine int, endLine int) {
	t.Helper()
	result := store.ConfirmScriptScene(ConfirmScriptSceneCommand{
		Root:        root,
		SceneID:     sceneID,
		Title:       "Market",
		Location:    "Laneway",
		Action:      "Mina enters.",
		SourceRange: ScriptSourceRange{StartLine: startLine, EndLine: endLine},
	})
	if !result.OK {
		t.Fatalf("ConfirmScriptScene() failed: %#v", result.Error)
	}
}

func saveTestShotCandidate(t *testing.T, store *Store, root string, candidateID string, sceneID string, index int, startLine int, endLine int) {
	t.Helper()
	result := store.SaveShotCandidate(SaveShotCandidateCommand{
		Root:              root,
		CandidateID:       candidateID,
		ScriptSceneID:     sceneID,
		Index:             index,
		DurationSeconds:   5,
		VisualDescription: "Candidate visual.",
		SourceRange:       ScriptSourceRange{StartLine: startLine, EndLine: endLine},
	})
	if !result.OK {
		t.Fatalf("SaveShotCandidate(%s) failed: %#v", candidateID, result.Error)
	}
}

func readJSON(t *testing.T, filename string, value any) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read JSON %s: %v", filename, err)
	}
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatalf("unmarshal JSON %s: %v", filename, err)
	}
}
