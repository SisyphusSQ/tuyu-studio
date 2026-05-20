package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestShotContextRequiredFieldsBlockContextReady(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 0, 0, 0, time.UTC), "shot-required")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_missing", ShotCardDTO{
		ID:              "shot_missing",
		ProjectID:       "proj_alpha_001",
		Title:           "Incomplete",
		Description:     "",
		DurationSeconds: 0,
		AspectRatio:     "9:16",
		ShotType:        "unspecified",
		CameraMovement:  "locked",
		Action:          "",
		Emotion:         "unspecified",
		Status:          ShotStatusDraft,
	})

	promoted := store.PromoteShotContext(PromoteShotContextCommand{
		Root:   root,
		ShotID: "shot_missing",
	})
	if promoted.OK || promoted.Error == nil || promoted.Error.Code != CodeShotRequiredFieldMissing {
		t.Fatalf("PromoteShotContext() = %#v, want %s", promoted, CodeShotRequiredFieldMissing)
	}
	if !containsString(promoted.Report.MissingFields, "description") ||
		!containsString(promoted.Report.MissingFields, "durationSeconds") ||
		!containsString(promoted.Report.MissingFields, "action") ||
		!containsString(promoted.Report.MissingFields, "emotion") ||
		!containsString(promoted.Report.MissingFields, "shotType") ||
		!containsString(promoted.Report.MissingFields, "context") {
		t.Fatalf("missing fields = %#v, want required fields and context", promoted.Report.MissingFields)
	}
}

func TestShotContextSourceLineageAndReferencesBlockContextReady(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 10, 0, 0, time.UTC), "shot-lineage")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_bad_refs", validShotContextDTO("shot_bad_refs"))
	var shot ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_bad_refs.json"), &shot)
	shot.SourceCandidateID = "candidate_bad_refs"
	shot.ScriptSceneID = ""
	shot.SourceRange = ScriptSourceRange{}
	shot.CharacterIDs = []string{"char_missing"}
	shot.SceneID = "scene_missing"
	shot.PropIDs = []string{"prop_missing"}
	shot.ReferenceAssetIDs = []string{"asset_missing"}
	writeJSON(t, filepath.Join(root, "shots", "shot_bad_refs.json"), shot)

	checked := store.ValidateShotContext(ValidateShotContextCommand{
		Root:   root,
		ShotID: "shot_bad_refs",
	})
	if !checked.OK || checked.Report.CanEnterContextReady {
		t.Fatalf("ValidateShotContext() = %#v, want blocking report", checked)
	}
	if !hasShotContextCode(checked.Report, CodeShotExpansionSourceMissing) {
		t.Fatalf("report = %#v, want lineage blocker", checked.Report)
	}
	if countShotContextCode(checked.Report, CodeShotReferenceMissing) != 4 {
		t.Fatalf("report = %#v, want four missing reference blockers", checked.Report)
	}
}

func TestShotContextReadyPersistsShotAndGraphStatus(t *testing.T) {
	now := time.Date(2026, 5, 20, 11, 20, 0, 0, time.UTC)
	store := testStore(now, "shot-ready")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_ready", validShotContextDTO("shot_ready"))

	promoted := store.PromoteShotContext(PromoteShotContextCommand{
		Root:   root,
		ShotID: "shot_ready",
	})
	if !promoted.OK || promoted.Shot == nil || !promoted.Report.CanEnterContextReady {
		t.Fatalf("PromoteShotContext() = %#v, want context_ready", promoted)
	}
	if promoted.Shot.Status != ShotStatusContextReady {
		t.Fatalf("shot status = %q, want %s", promoted.Shot.Status, ShotStatusContextReady)
	}

	var shot ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_ready.json"), &shot)
	if shot.Status != ShotStatusContextReady || shot.UpdatedAt != now.Format(time.RFC3339) {
		t.Fatalf("persisted shot = %#v, want context_ready with timestamp", shot)
	}
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if got := firstShotNodeStatus(manifest, "shot_ready"); got != ShotStatusContextReady {
		t.Fatalf("graph shot status = %q, want %s", got, ShotStatusContextReady)
	}
	reopened := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "shot-reopen"})
	if !reopened.OK || reopened.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want canvas", reopened)
	}
	node := findGraphNode(reopened.Canvas.Nodes, "node_shot_ready")
	if node == nil || node.Status != ShotStatusContextReady || !containsString(node.Badges, "context_ready") {
		t.Fatalf("GraphView shot node = %#v, want context_ready badge", node)
	}
}

func TestShotContextDirtyPersistsAfterRevision(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 30, 0, 0, time.UTC), "shot-dirty")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_dirty", validShotContextDTO("shot_dirty"))
	promoted := store.PromoteShotContext(PromoteShotContextCommand{Root: root, ShotID: "shot_dirty"})
	if !promoted.OK {
		t.Fatalf("PromoteShotContext() failed: %#v", promoted.Error)
	}

	dirty := store.MarkShotContextDirty(MarkShotContextDirtyCommand{
		Root:   root,
		ShotID: "shot_dirty",
		Reason: "description revised",
	})
	if !dirty.OK || dirty.Shot == nil || dirty.Shot.Status != ShotStatusContextDirty {
		t.Fatalf("MarkShotContextDirty() = %#v, want context_dirty", dirty)
	}
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if got := firstShotNodeStatus(manifest, "shot_dirty"); got != ShotStatusContextDirty {
		t.Fatalf("graph shot status = %q, want %s", got, ShotStatusContextDirty)
	}
}

func TestShotContextRejectsUnsupportedPromotionState(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 40, 0, 0, time.UTC), "shot-illegal-state")
	root := createShotContextProject(t)
	shot := validShotContextDTO("shot_prompt_ready")
	shot.Status = "prompt_ready"
	writeShotContextFixture(t, root, "shot_prompt_ready", shot)

	promoted := store.PromoteShotContext(PromoteShotContextCommand{
		Root:   root,
		ShotID: "shot_prompt_ready",
	})
	if promoted.OK || promoted.Error == nil || promoted.Error.Code != CodeShotIllegalTransition {
		t.Fatalf("PromoteShotContext() = %#v, want %s", promoted, CodeShotIllegalTransition)
	}

	var persisted ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_prompt_ready.json"), &persisted)
	if persisted.Status != "prompt_ready" {
		t.Fatalf("persisted status = %q, want prompt_ready", persisted.Status)
	}
}

func TestShotContextRejectsFileIdentityMismatch(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 50, 0, 0, time.UTC), "shot-id-mismatch")
	root := createShotContextProject(t)
	shot := validShotContextDTO("shot_embedded")
	writeShotContextFixture(t, root, "shot_requested", shot)

	promoted := store.PromoteShotContext(PromoteShotContextCommand{
		Root:   root,
		ShotID: "shot_requested",
	})
	if promoted.OK || promoted.Error == nil || promoted.Error.Code != CodeShotIdentityMismatch {
		t.Fatalf("PromoteShotContext() = %#v, want %s", promoted, CodeShotIdentityMismatch)
	}
	if _, err := os.Stat(filepath.Join(root, "shots", "shot_embedded.json")); !os.IsNotExist(err) {
		t.Fatalf("embedded shot file exists or stat failed with non-missing error: %v", err)
	}
}

func TestShotContextCanonicalizesWhitespaceFileID(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 55, 0, 0, time.UTC), "shot-id-canonical")
	root := createShotContextProject(t)
	shot := validShotContextDTO("shot_canonical")
	shot.ID = " shot_canonical "
	writeShotContextFixture(t, root, "shot_canonical", shot)

	promoted := store.PromoteShotContext(PromoteShotContextCommand{
		Root:   root,
		ShotID: "shot_canonical",
	})
	if !promoted.OK || promoted.Shot == nil || promoted.Shot.ID != "shot_canonical" {
		t.Fatalf("PromoteShotContext() = %#v, want canonical shot id", promoted)
	}
	var persisted ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_canonical.json"), &persisted)
	if persisted.ID != "shot_canonical" || persisted.Status != ShotStatusContextReady {
		t.Fatalf("persisted shot = %#v, want canonical id and context_ready", persisted)
	}
	if _, err := os.Stat(filepath.Join(root, "shots", " shot_canonical .json")); !os.IsNotExist(err) {
		t.Fatalf("whitespace shot file exists or stat failed with non-missing error: %v", err)
	}
}

func TestShotContextRejectsMissingGraphNode(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 11, 58, 0, 0, time.UTC), "shot-missing-graph")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_orphan", validShotContextDTO("shot_orphan"))
	removeShotContextGraphNode(t, root, "shot_orphan")

	promoted := store.PromoteShotContext(PromoteShotContextCommand{
		Root:   root,
		ShotID: "shot_orphan",
	})
	if promoted.OK || promoted.Error == nil || promoted.Error.Code != CodeShotGraphNodeMissing {
		t.Fatalf("PromoteShotContext() = %#v, want %s", promoted, CodeShotGraphNodeMissing)
	}
	var persisted ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_orphan.json"), &persisted)
	if persisted.Status != ShotStatusDraft {
		t.Fatalf("persisted status = %q, want draft", persisted.Status)
	}
}

func TestGraphViewProjectsMissingShotReferences(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC), "shot-graph-ref")
	root := createShotContextProject(t)
	shot := validShotContextDTO("shot_missing_refs")
	shot.CharacterIDs = []string{"char_missing"}
	shot.SceneID = "scene_missing"
	shot.PropIDs = []string{"prop_missing"}
	shot.ReferenceAssetIDs = []string{"asset_missing"}
	writeShotContextFixture(t, root, "shot_missing_refs", shot)

	view := store.GraphView(GraphViewCommand{
		Root:          root,
		CorrelationID: "graph-missing-shot-refs",
	})
	if !view.OK || view.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want canvas", view)
	}
	node := findGraphNode(view.Canvas.Nodes, "node_shot_missing_refs")
	if node == nil || !containsString(node.Badges, "missing_context") || !containsString(node.Badges, "blocked") {
		t.Fatalf("GraphView shot node = %#v, want missing_context and blocked badges", node)
	}
	missing := node.Data["missingFields"]
	for _, field := range []string{"characterIds", "sceneProfileId", "propIds", "referenceAssetIds"} {
		if !strings.Contains(missing, field) {
			t.Fatalf("missingFields = %q, want %s", missing, field)
		}
	}
}

func TestShotContextRejectsUnsafeReferenceAssetPath(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 12, 10, 0, 0, time.UTC), "shot-unsafe-asset")
	root := createShotContextProject(t)
	if err := os.WriteFile(filepath.Join(root, "..", "outside-ref.txt"), []byte("outside"), 0o644); err != nil {
		t.Fatalf("write outside ref: %v", err)
	}
	writeJSON(t, filepath.Join(root, "characters", "char-mina.json"), map[string]any{
		"id":                "char_mina",
		"displayName":       "Mina",
		"shortDescription":  "Courier",
		"referenceAssetIds": []string{"asset_ref_escape"},
		"referencePaths":    []string{"../outside-ref.txt"},
	})
	shot := validShotContextDTO("shot_unsafe_asset")
	shot.ReferenceAssetIDs = []string{"asset_ref_escape"}
	writeShotContextFixture(t, root, "shot_unsafe_asset", shot)

	checked := store.ValidateShotContext(ValidateShotContextCommand{
		Root:   root,
		ShotID: "shot_unsafe_asset",
	})
	if !checked.OK || checked.Report.CanEnterContextReady {
		t.Fatalf("ValidateShotContext() = %#v, want unsafe asset blocker", checked)
	}
	if !hasShotReferenceStatus(checked.Report, "asset_ref_escape", "invalid_path") {
		t.Fatalf("references = %#v, want invalid_path", checked.Report.References)
	}
}

func createShotContextProject(t *testing.T) string {
	t.Helper()
	root := createScriptDocumentProject(t)
	writeJSON(t, filepath.Join(root, "characters", "char-mina.json"), map[string]any{
		"id":                "char_mina",
		"displayName":       "Mina",
		"shortDescription":  "Courier",
		"referenceAssetIds": []string{"asset_ref_character_mina"},
		"referencePaths":    []string{"assets/refs/character-mina.txt"},
	})
	writeJSON(t, filepath.Join(root, "scenes", "scene-market.json"), map[string]any{
		"id":                "scene_market",
		"displayName":       "Market",
		"shortDescription":  "Laneway market",
		"referenceAssetIds": []string{"asset_ref_scene_market"},
		"referencePaths":    []string{"assets/refs/scene-market.txt"},
	})
	writeJSON(t, filepath.Join(root, "props", "prop-lantern.json"), map[string]any{
		"id":                "prop_lantern",
		"displayName":       "Lantern",
		"shortDescription":  "Red paper lantern",
		"referenceAssetIds": []string{"asset_ref_prop_lantern"},
		"referencePaths":    []string{"assets/refs/prop-lantern.txt"},
	})
	writeFile(t, filepath.Join(root, "assets", "refs", "character-mina.txt"), "Mina reference")
	writeFile(t, filepath.Join(root, "assets", "refs", "scene-market.txt"), "Scene reference")
	writeFile(t, filepath.Join(root, "assets", "refs", "prop-lantern.txt"), "Prop reference")
	return root
}

func writeShotContextFixture(t *testing.T, root string, shotID string, shot ShotCardDTO) {
	t.Helper()
	writeJSON(t, filepath.Join(root, "shots", shotID+".json"), shot)
	manifestData, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest, report := DecodeManifest(manifestData)
	if report.HasBlocking() {
		t.Fatalf("manifest invalid: %#v", report.Items)
	}
	manifest.Graph.Nodes = append(manifest.Graph.Nodes, Node{
		ID:     "node_" + shotID,
		Kind:   "shot",
		RefID:  "shots/" + shotID + ".json",
		Status: shot.Status,
	})
	manifest.Graph.Version++
	manifest.Integrity.LastGraphVersion = manifest.Graph.Version
	writeManifestBypassValidation(t, root, manifest)
}

func removeShotContextGraphNode(t *testing.T, root string, shotID string) {
	t.Helper()
	manifestData, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest, report := DecodeManifest(manifestData)
	if report.HasBlocking() {
		t.Fatalf("manifest invalid: %#v", report.Items)
	}
	nodes := manifest.Graph.Nodes[:0]
	for _, node := range manifest.Graph.Nodes {
		if node.RefID != "shots/"+shotID+".json" {
			nodes = append(nodes, node)
		}
	}
	manifest.Graph.Nodes = nodes
	manifest.Graph.Version++
	manifest.Integrity.LastGraphVersion = manifest.Graph.Version
	writeManifestBypassValidation(t, root, manifest)
}

func validShotContextDTO(shotID string) ShotCardDTO {
	return ShotCardDTO{
		ID:                shotID,
		ProjectID:         "proj_alpha_001",
		SceneID:           "scene_market",
		SourceCandidateID: "candidate_" + shotID,
		ScriptSceneID:     "script_scene_market",
		Index:             1,
		Title:             "Mina reaches the stall",
		Description:       "Mina steps under the food stall awning.",
		DurationSeconds:   6,
		AspectRatio:       "9:16",
		ShotType:          "medium",
		CameraMovement:    "slow push",
		Action:            "She catches her breath.",
		Emotion:           "relieved",
		CharacterIDs:      []string{"char_mina"},
		PropIDs:           []string{"prop_lantern"},
		ReferenceAssetIDs: []string{"asset_ref_character_mina", "asset_ref_scene_market", "asset_ref_prop_lantern"},
		SourceRange:       ScriptSourceRange{StartLine: 1, EndLine: 1},
		Status:            ShotStatusDraft,
		CreatedAt:         "2026-05-20T00:00:00Z",
		UpdatedAt:         "2026-05-20T00:00:00Z",
	}
}

func writeFile(t *testing.T, filename string, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	if err := os.WriteFile(filename, []byte(value), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func hasShotContextCode(report ShotContextReportDTO, code string) bool {
	return countShotContextCode(report, code) > 0
}

func countShotContextCode(report ShotContextReportDTO, code string) int {
	count := 0
	for _, issue := range report.Blocking {
		if issue.Code == code {
			count++
		}
	}
	return count
}

func hasShotReferenceStatus(report ShotContextReportDTO, referenceID string, status string) bool {
	for _, ref := range report.References {
		if ref.ReferenceID == referenceID && ref.Status == status {
			return true
		}
	}
	return false
}

func firstShotNodeStatus(manifest Manifest, shotID string) string {
	for _, node := range manifest.Graph.Nodes {
		if node.RefID == "shots/"+shotID+".json" {
			return node.Status
		}
	}
	return ""
}

func findGraphNode(nodes []GraphNodeDTO, id string) *GraphNodeDTO {
	for index := range nodes {
		if nodes[index].ID == id {
			return &nodes[index]
		}
	}
	return nil
}
