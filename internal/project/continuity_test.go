package project

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestContinuityRuleSavePersistsLockedRuleAndMarksShotsDirty(t *testing.T) {
	now := time.Date(2026, 5, 20, 14, 0, 0, 0, time.UTC)
	store := testStore(now, "continuity-rule-save")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_rule_dirty", validShotContextDTO("shot_rule_dirty"))
	if promoted := store.PromoteShotContext(PromoteShotContextCommand{Root: root, ShotID: "shot_rule_dirty"}); !promoted.OK {
		t.Fatalf("PromoteShotContext() = %#v, want ready", promoted)
	}

	result := store.SaveContinuityRule(SaveContinuityRuleCommand{
		Root:          root,
		ID:            "rule_char_mina_raincoat",
		TargetType:    BindingTargetCharacter,
		TargetID:      "char_mina",
		Rule:          "Mina keeps the yellow raincoat visible in rainy exterior shots.",
		Severity:      ContinuitySeverityBlocking,
		Locked:        true,
		CreatedBy:     "local_user",
		CorrelationID: "continuity-save",
	})
	if !result.OK || result.Rule == nil || !result.Rule.Locked {
		t.Fatalf("SaveContinuityRule() = %#v, want locked rule", result)
	}
	if result.Impact == nil || !containsTestString(result.Impact.AffectedShots, "shot_rule_dirty") {
		t.Fatalf("impact = %#v, want affected shot", result.Impact)
	}

	var profile ProfileDTO
	record, err := readProfileRecordFile(filepath.Join(root, "characters", "char-mina.json"), "characters/char-mina.json", BindingTargetCharacter)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	profile = record.dto
	if !containsTestString(profile.LockedRules, "rule_char_mina_raincoat") || len(profile.ContinuityRules) != 1 {
		t.Fatalf("profile = %#v, want locked continuity rule", profile)
	}
	var shot ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_rule_dirty.json"), &shot)
	if shot.Status != ShotStatusContextDirty || shot.UpdatedAt != now.Format(time.RFC3339) {
		t.Fatalf("shot = %#v, want context_dirty timestamp", shot)
	}
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if got := firstShotNodeStatus(manifest, "shot_rule_dirty"); got != ShotStatusContextDirty {
		t.Fatalf("graph shot status = %q, want %s", got, ShotStatusContextDirty)
	}
}

func TestContinuityRuleSaveCannotSilentlyUnlockLockedRule(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 14, 5, 0, 0, time.UTC), "continuity-rule-save-locked")
	root := createShotContextProject(t)
	if saved := store.SaveContinuityRule(SaveContinuityRuleCommand{
		Root:       root,
		ID:         "rule_char_mina_raincoat",
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		Rule:       "Mina keeps the yellow raincoat visible.",
		Locked:     true,
	}); !saved.OK {
		t.Fatalf("SaveContinuityRule(locked) = %#v, want success", saved)
	}

	unlockedBySave := store.SaveContinuityRule(SaveContinuityRuleCommand{
		Root:       root,
		ID:         "rule_char_mina_raincoat",
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		Rule:       "Mina keeps the yellow raincoat visible.",
		Locked:     false,
	})
	if unlockedBySave.OK || unlockedBySave.Error == nil || unlockedBySave.Error.Code != CodeContinuityUnlockReason {
		t.Fatalf("SaveContinuityRule(locked=false) = %#v, want explicit unlock reason error", unlockedBySave)
	}
	record, err := readProfileRecordFile(filepath.Join(root, "characters", "char-mina.json"), "characters/char-mina.json", BindingTargetCharacter)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if !containsTestString(record.dto.LockedRules, "rule_char_mina_raincoat") || !record.dto.ContinuityRules[0].Locked {
		t.Fatalf("profile = %#v, want rule still locked", record.dto)
	}
}

func TestContinuityRuleUnlockRequiresReasonAndWritesAudit(t *testing.T) {
	now := time.Date(2026, 5, 20, 14, 10, 0, 0, time.UTC)
	store := testStore(now, "continuity-rule-unlock")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_unlock_dirty", validShotContextDTO("shot_unlock_dirty"))
	if promoted := store.PromoteShotContext(PromoteShotContextCommand{Root: root, ShotID: "shot_unlock_dirty"}); !promoted.OK {
		t.Fatalf("PromoteShotContext() = %#v, want ready", promoted)
	}
	if saved := store.SaveContinuityRule(SaveContinuityRuleCommand{
		Root:       root,
		ID:         "rule_char_mina_raincoat",
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		Rule:       "Mina keeps the yellow raincoat visible.",
		Locked:     true,
	}); !saved.OK {
		t.Fatalf("SaveContinuityRule() = %#v, want success", saved)
	}
	if promoted := store.PromoteShotContext(PromoteShotContextCommand{Root: root, ShotID: "shot_unlock_dirty"}); !promoted.OK {
		t.Fatalf("PromoteShotContext(re-ready) = %#v, want ready", promoted)
	}

	missingReason := store.UnlockContinuityRule(UnlockContinuityRuleCommand{
		Root:   root,
		RuleID: "rule_char_mina_raincoat",
	})
	if missingReason.OK || missingReason.Error == nil || missingReason.Error.Code != CodeContinuityUnlockReason {
		t.Fatalf("UnlockContinuityRule(empty reason) = %#v, want reason error", missingReason)
	}

	unlocked := store.UnlockContinuityRule(UnlockContinuityRuleCommand{
		Root:          root,
		RuleID:        "rule_char_mina_raincoat",
		Reason:        "director approved a costume exception",
		CorrelationID: "continuity-unlock",
	})
	if !unlocked.OK || unlocked.Rule == nil || unlocked.Rule.Locked {
		t.Fatalf("UnlockContinuityRule() = %#v, want unlocked rule", unlocked)
	}
	var shot ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_unlock_dirty.json"), &shot)
	if shot.Status != ShotStatusContextDirty {
		t.Fatalf("shot status = %q, want %s", shot.Status, ShotStatusContextDirty)
	}
	auditData, err := os.ReadFile(filepath.Join(root, AuditEventsRelativePath))
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if !containsBytes(auditData, []byte("continuity.rule.unlocked")) || !containsBytes(auditData, []byte("director approved")) {
		t.Fatalf("audit = %s, want unlock event and reason", auditData)
	}
}

func TestUnlockAssetBindingRejectsIncompleteSelector(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 14, 15, 0, 0, time.UTC), "binding-unlock-selector")
	root := createAssetBindingProject(t)
	writeCharacterProfile(t, root, "char_mina")
	asset := importBindingTextAsset(t, store, root, "selector-reference.txt", "selector reference\n")
	if bound := store.BindAsset(BindAssetCommand{
		Root:       root,
		AssetID:    asset.ID,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		Purpose:    "reference",
	}); !bound.OK {
		t.Fatalf("BindAsset() = %#v, want success", bound)
	}

	for _, command := range []UnlockAssetBindingCommand{
		{Root: root, Reason: "missing selector"},
		{Root: root, AssetID: asset.ID, Reason: "partial selector"},
		{Root: root, AssetID: asset.ID, TargetType: BindingTargetCharacter, TargetID: "char_mina", Reason: "missing purpose"},
	} {
		result := store.UnlockAssetBinding(command)
		if result.OK || result.Error == nil || result.Error.Code != CodeContinuityBindingSelector {
			t.Fatalf("UnlockAssetBinding(%#v) = %#v, want selector error", command, result)
		}
	}
}

func TestLockedMainReferenceBindingRequiresExplicitUnlock(t *testing.T) {
	now := time.Date(2026, 5, 20, 14, 20, 0, 0, time.UTC)
	store := testStore(now, "continuity-binding-unlock")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_binding_dirty", validShotContextDTO("shot_binding_dirty"))
	if promoted := store.PromoteShotContext(PromoteShotContextCommand{Root: root, ShotID: "shot_binding_dirty"}); !promoted.OK {
		t.Fatalf("PromoteShotContext() = %#v, want ready", promoted)
	}
	first := importBindingTextAsset(t, store, root, "main-locked.txt", "locked main reference\n")
	second := importBindingTextAsset(t, store, root, "main-replacement.txt", "replacement main reference\n")
	if set := store.SetMainReference(SetMainReferenceCommand{
		Root:       root,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		AssetID:    first.ID,
	}); !set.OK {
		t.Fatalf("SetMainReference(first) = %#v, want success", set)
	}
	index := readAssetIndexForTest(t, root)
	for assetIndex := range index.Assets {
		for bindingIndex := range index.Assets[assetIndex].Bindings {
			if index.Assets[assetIndex].Bindings[bindingIndex].Purpose == BindingPurposeMainReference {
				index.Assets[assetIndex].Bindings[bindingIndex].Locked = true
			}
		}
	}
	if err := store.writeAssetIndex(root, index); err != nil {
		t.Fatalf("write asset index: %v", err)
	}

	blocked := store.SetMainReference(SetMainReferenceCommand{
		Root:       root,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		AssetID:    second.ID,
	})
	if blocked.OK || blocked.Error == nil || blocked.Error.Code != CodeContinuityLockedBinding {
		t.Fatalf("SetMainReference(locked replacement) = %#v, want locked binding error", blocked)
	}

	unlocked := store.UnlockAssetBinding(UnlockAssetBindingCommand{
		Root:       root,
		AssetID:    first.ID,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		Purpose:    BindingPurposeMainReference,
		Reason:     "approved main reference replacement",
	})
	if !unlocked.OK {
		t.Fatalf("UnlockAssetBinding() = %#v, want success", unlocked)
	}
	replaced := store.SetMainReference(SetMainReferenceCommand{
		Root:       root,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		AssetID:    second.ID,
	})
	if !replaced.OK || replaced.Profile == nil || replaced.Profile.MainReferenceAssetID != second.ID {
		t.Fatalf("SetMainReference(after unlock) = %#v, want replacement", replaced)
	}
	var shot ShotCardDTO
	readJSON(t, filepath.Join(root, "shots", "shot_binding_dirty.json"), &shot)
	if shot.Status != ShotStatusContextDirty {
		t.Fatalf("shot status = %q, want dirty after continuity changes", shot.Status)
	}
}

func TestContinuityListReturnsGlobalImpact(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 14, 25, 0, 0, time.UTC), "continuity-global-impact")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_global_impact", validShotContextDTO("shot_global_impact"))
	asset := importBindingTextAsset(t, store, root, "global-main.txt", "global main reference\n")
	if set := store.SetMainReference(SetMainReferenceCommand{
		Root:       root,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		AssetID:    asset.ID,
	}); !set.OK {
		t.Fatalf("SetMainReference() = %#v, want success", set)
	}

	listed := store.ListContinuity(ListContinuityCommand{Root: root, CorrelationID: "global-impact"})
	if !listed.OK || listed.Impact == nil {
		t.Fatalf("ListContinuity() = %#v, want global impact", listed)
	}
	if !containsTestString(listed.Impact.AffectedAssets, asset.ID) || !containsTestString(listed.Impact.AffectedProfiles, "char_mina") || !containsTestString(listed.Impact.AffectedShots, "shot_global_impact") {
		t.Fatalf("impact = %#v, want asset/profile/shot", listed.Impact)
	}
}

func TestContinuityHealthReportsAssetIssuesToAffectedShots(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 14, 30, 0, 0, time.UTC), "continuity-health")
	root := createShotContextProject(t)
	writeShotContextFixture(t, root, "shot_asset_issue", validShotContextDTO("shot_asset_issue"))
	asset := importBindingTextAsset(t, store, root, "main-missing.txt", "main reference\n")
	if set := store.SetMainReference(SetMainReferenceCommand{
		Root:       root,
		TargetType: BindingTargetCharacter,
		TargetID:   "char_mina",
		AssetID:    asset.ID,
	}); !set.OK {
		t.Fatalf("SetMainReference() = %#v, want success", set)
	}
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(asset.RelativePath))); err != nil {
		t.Fatalf("remove asset: %v", err)
	}

	health := store.HealthReport(root)
	item := firstHealthItemByCode(health, CodeAssetMissing)
	if item == nil {
		t.Fatalf("health = %#v, want asset missing", health.Items)
	}
	if !containsTestString(item.AffectedObjects, "shot_asset_issue") || !containsTestString(item.AffectedObjects, "node_shot_asset_issue") {
		t.Fatalf("affectedObjects = %#v, want shot id and graph node id", item.AffectedObjects)
	}

	graph := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "continuity-health-graph"})
	if !graph.OK || graph.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want canvas", graph)
	}
	node := findGraphNode(graph.Canvas.Nodes, "node_shot_asset_issue")
	if node == nil || !containsTestString(node.Badges, "missing_asset") {
		t.Fatalf("shot node = %#v, want missing_asset badge", node)
	}
}

func containsBytes(value []byte, want []byte) bool {
	return bytes.Contains(value, want)
}

func containsTestString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func firstHealthItemByCode(report HealthReport, code string) *HealthItem {
	for index := range report.Items {
		if report.Items[index].Code == code {
			return &report.Items[index]
		}
	}
	return nil
}
