package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAssetBindingPersistsProfileAndLineage(t *testing.T) {
	now := time.Date(2026, 5, 20, 13, 0, 0, 0, time.UTC)
	store := testStore(now, "asset-binding")
	root := createAssetBindingProject(t)
	writeCharacterProfile(t, root, "char_mina")
	asset := importBindingTextAsset(t, store, root, "mina-reference.txt", "Mina reference\n")

	result := store.BindAsset(BindAssetCommand{
		Root:          root,
		AssetID:       asset.ID,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		Purpose:       "identity",
		CreatedBy:     "local_user",
		CorrelationID: "bind-character",
	})
	if !result.OK || result.Asset == nil || result.Profile == nil {
		t.Fatalf("BindAsset() = %#v, want bound asset/profile", result)
	}
	if result.Profile.MainReferenceAssetID != "" {
		t.Fatalf("mainReferenceAssetId = %q, want empty until set explicitly", result.Profile.MainReferenceAssetID)
	}
	if result.Profile.BindingCount != 1 || len(result.Profile.ReferenceAssetIDs) != 1 || result.Profile.ReferenceAssetIDs[0] != asset.ID {
		t.Fatalf("profile = %#v, want one reference binding", result.Profile)
	}
	if len(result.Asset.Bindings) != 1 {
		t.Fatalf("asset bindings = %#v, want one binding", result.Asset.Bindings)
	}
	binding := result.Asset.Bindings[0]
	if binding.AssetID != asset.ID || binding.TargetType != BindingTargetCharacter || binding.TargetID != "char_mina" || binding.Purpose != "identity" || binding.Locked {
		t.Fatalf("binding = %#v, want identity binding with locked=false", binding)
	}

	var profileRaw map[string]any
	readJSONFile(t, filepath.Join(root, "characters", "char-mina.json"), &profileRaw)
	if got := stringSliceFromAny(profileRaw["referenceAssetIds"]); len(got) != 1 || got[0] != asset.ID {
		t.Fatalf("profile referenceAssetIds = %#v, want %s", got, asset.ID)
	}
	if _, ok := profileRaw["lockedRules"]; !ok {
		t.Fatalf("profile raw = %#v, want baseline lockedRules", profileRaw)
	}

	reopened := testStore(now.Add(time.Minute), "asset-binding-reopen").ListAssetBindings(ListAssetBindingsCommand{
		Root:          root,
		CorrelationID: "binding-list",
	})
	if !reopened.OK || len(reopened.Profiles) != 1 || len(reopened.Lineage) != 1 {
		t.Fatalf("ListAssetBindings() = %#v, want persisted profile and lineage", reopened)
	}
	if reopened.Lineage[0].TargetSummaries[0].Name != "Mina" {
		t.Fatalf("lineage = %#v, want target summary name", reopened.Lineage)
	}
}

func TestAssetBindingDuplicatePolicy(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 13, 5, 0, 0, time.UTC), "asset-binding-duplicate")
	root := createAssetBindingProject(t)
	writeCharacterProfile(t, root, "char_mina")
	asset := importBindingTextAsset(t, store, root, "duplicate-reference.txt", "same binding\n")
	command := BindAssetCommand{
		Root:          root,
		AssetID:       asset.ID,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		Purpose:       "identity",
		CorrelationID: "binding-duplicate",
	}
	if first := store.BindAsset(command); !first.OK {
		t.Fatalf("first BindAsset() = %#v, want success", first)
	}

	duplicate := store.BindAsset(command)
	if duplicate.OK || duplicate.Error == nil {
		t.Fatalf("duplicate BindAsset() = %#v, want duplicate error", duplicate)
	}
	if duplicate.Error.Code != CodeBindingDuplicate || duplicate.Duplicate == nil {
		t.Fatalf("duplicate = %#v, want %s and duplicate dto", duplicate, CodeBindingDuplicate)
	}

	command.DuplicatePolicy = AssetDuplicatePolicyReuse
	reused := store.BindAsset(command)
	if !reused.OK || reused.Duplicate == nil {
		t.Fatalf("reuse BindAsset() = %#v, want reused duplicate", reused)
	}
	if len(reused.Asset.Bindings) != 1 {
		t.Fatalf("bindings = %#v, want no duplicate write", reused.Asset.Bindings)
	}
}

func TestAssetBindingRejectsMissingTarget(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 13, 10, 0, 0, time.UTC), "asset-binding-missing")
	root := createAssetBindingProject(t)
	asset := importBindingTextAsset(t, store, root, "orphan-reference.txt", "orphan binding\n")

	result := store.BindAsset(BindAssetCommand{
		Root:          root,
		AssetID:       asset.ID,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_missing",
		CorrelationID: "binding-missing",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("BindAsset() = %#v, want missing target error", result)
	}
	if result.Error.Code != CodeBindingTargetMissing {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeBindingTargetMissing)
	}
}

func TestAssetBindingRejectsReservedMainReferencePurpose(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 13, 12, 0, 0, time.UTC), "asset-binding-reserved")
	root := createAssetBindingProject(t)
	writeCharacterProfile(t, root, "char_mina")
	asset := importBindingTextAsset(t, store, root, "reserved-reference.txt", "reserved binding\n")

	result := store.BindAsset(BindAssetCommand{
		Root:          root,
		AssetID:       asset.ID,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		Purpose:       BindingPurposeMainReference,
		CorrelationID: "binding-main-reserved",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("BindAsset(main_reference) = %#v, want reserved-purpose error", result)
	}
	if result.Error.Code != CodeBindingPurposeReserved {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeBindingPurposeReserved)
	}

	reopened := store.ListAssetBindings(ListAssetBindingsCommand{Root: root, CorrelationID: "binding-main-reserved-list"})
	if !reopened.OK || len(reopened.Profiles) != 1 {
		t.Fatalf("ListAssetBindings() = %#v, want profile", reopened)
	}
	if reopened.Profiles[0].MainReferenceAssetID != "" || reopened.Profiles[0].BindingCount != 0 {
		t.Fatalf("profile = %#v, want no main reference or binding from generic bind", reopened.Profiles[0])
	}
}

func TestMainReferenceSetReplaceClearAndGraphProjection(t *testing.T) {
	now := time.Date(2026, 5, 20, 13, 15, 0, 0, time.UTC)
	store := testStore(now, "main-reference")
	root := createAssetBindingProject(t)
	writeCharacterProfile(t, root, "char_mina")
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest.Graph.Nodes = []Node{{ID: "node_character_mina", Kind: "character", RefID: "characters/char-mina.json", Status: "ready"}}
	writeManifestBypassValidation(t, root, manifest)

	first := importBindingTextAsset(t, store, root, "main-one.txt", "main reference one\n")
	second := importBindingTextAsset(t, store, root, "main-two.txt", "main reference two\n")

	setFirst := store.SetMainReference(SetMainReferenceCommand{
		Root:          root,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		AssetID:       first.ID,
		CorrelationID: "main-first",
	})
	if !setFirst.OK || setFirst.Profile == nil || setFirst.Profile.MainReferenceAssetID != first.ID {
		t.Fatalf("SetMainReference(first) = %#v, want first main reference", setFirst)
	}

	setSecond := store.SetMainReference(SetMainReferenceCommand{
		Root:          root,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		AssetID:       second.ID,
		CorrelationID: "main-second",
	})
	if !setSecond.OK || setSecond.Profile == nil || setSecond.Profile.MainReferenceAssetID != second.ID {
		t.Fatalf("SetMainReference(second) = %#v, want replacement", setSecond)
	}
	if countPurpose(setSecond.Assets, first.ID, BindingPurposeMainReference) != 0 {
		t.Fatalf("old asset retained main_reference binding: %#v", setSecond.Assets)
	}
	if countPurpose(setSecond.Assets, second.ID, BindingPurposeMainReference) != 1 {
		t.Fatalf("new asset missing main_reference binding: %#v", setSecond.Assets)
	}

	graph := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "graph-main-reference"})
	if !graph.OK || graph.Canvas == nil || len(graph.Canvas.Nodes) != 1 {
		t.Fatalf("GraphView() = %#v, want one character node", graph)
	}
	node := graph.Canvas.Nodes[0]
	if node.Data["mainReferenceAssetId"] != second.ID || node.Data["mainReferenceStatus"] != "ready" {
		t.Fatalf("node data = %#v, want second main reference visible", node.Data)
	}
	if !containsBindingTestString(node.Badges, "main_reference") {
		t.Fatalf("node badges = %#v, want main_reference", node.Badges)
	}

	cleared := store.SetMainReference(SetMainReferenceCommand{
		Root:          root,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		Clear:         true,
		CorrelationID: "main-clear",
	})
	if !cleared.OK || cleared.Profile == nil || cleared.Profile.MainReferenceAssetID != "" {
		t.Fatalf("SetMainReference(clear) = %#v, want empty main reference", cleared)
	}
	if len(cleared.Profile.ReferenceAssetIDs) != 2 {
		t.Fatalf("referenceAssetIds = %#v, want references preserved after clear", cleared.Profile.ReferenceAssetIDs)
	}
	if countPurpose(cleared.Assets, second.ID, BindingPurposeMainReference) != 0 {
		t.Fatalf("main_reference binding remained after clear: %#v", cleared.Assets)
	}
}

func TestMainReferenceRejectsEmptyAssetWithoutClear(t *testing.T) {
	now := time.Date(2026, 5, 20, 13, 20, 0, 0, time.UTC)
	store := testStore(now, "main-reference-empty")
	root := createAssetBindingProject(t)
	writeCharacterProfile(t, root, "char_mina")
	asset := importBindingTextAsset(t, store, root, "main-existing.txt", "main reference\n")

	set := store.SetMainReference(SetMainReferenceCommand{
		Root:          root,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		AssetID:       asset.ID,
		CorrelationID: "main-existing",
	})
	if !set.OK || set.Profile == nil || set.Profile.MainReferenceAssetID != asset.ID {
		t.Fatalf("SetMainReference(existing) = %#v, want main reference", set)
	}

	empty := store.SetMainReference(SetMainReferenceCommand{
		Root:          root,
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		CorrelationID: "main-empty-without-clear",
	})
	if empty.OK || empty.Error == nil {
		t.Fatalf("SetMainReference(empty) = %#v, want validation error", empty)
	}
	if empty.Error.Code != CodeMainReferenceAssetMissing {
		t.Fatalf("error code = %q, want %s", empty.Error.Code, CodeMainReferenceAssetMissing)
	}

	reopened := store.ListAssetBindings(ListAssetBindingsCommand{Root: root, CorrelationID: "main-empty-reopen"})
	if !reopened.OK || len(reopened.Profiles) != 1 {
		t.Fatalf("ListAssetBindings() = %#v, want profile", reopened)
	}
	if reopened.Profiles[0].MainReferenceAssetID != asset.ID {
		t.Fatalf("mainReferenceAssetId = %q, want preserved %q", reopened.Profiles[0].MainReferenceAssetID, asset.ID)
	}
	if countPurpose(reopened.Assets, asset.ID, BindingPurposeMainReference) != 1 {
		t.Fatalf("assets = %#v, want existing main_reference binding preserved", reopened.Assets)
	}
}

func createAssetBindingProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "binding-project")
	if err := CreateBaselineProject(root, testManifest(t)); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}
	return root
}

func writeCharacterProfile(t *testing.T, root string, id string) {
	t.Helper()
	body := map[string]any{
		"id":               id,
		"displayName":      "Mina",
		"shortDescription": "Courier with a yellow raincoat.",
		"visualRules":      []string{"short black hair", "yellow raincoat"},
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(root, "characters", "char-mina.json"), data, 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
}

func importBindingTextAsset(t *testing.T, store *Store, root string, name string, body string) AssetDTO {
	t.Helper()
	source := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(source, []byte(body), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	result := store.ImportAsset(ImportAssetCommand{
		Root:          root,
		SourcePath:    source,
		Role:          "character_ref",
		CorrelationID: "import-" + name,
	})
	if !result.OK || result.Asset == nil {
		t.Fatalf("ImportAsset() = %#v, want asset", result)
	}
	return *result.Asset
}

func readJSONFile(t *testing.T, filename string, target any) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode %s: %v", filename, err)
	}
}

func stringSliceFromAny(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok {
			values = append(values, text)
		}
	}
	return values
}

func countPurpose(assets []AssetDTO, assetID string, purpose string) int {
	total := 0
	for _, asset := range assets {
		if asset.ID != assetID {
			continue
		}
		for _, binding := range asset.Bindings {
			if binding.Purpose == purpose {
				total++
			}
		}
	}
	return total
}

func containsBindingTestString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
