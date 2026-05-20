package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerationPackageExportReadyShot(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 13, 10, 0, 0, time.UTC), "package-export")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:              root,
		ShotID:            "shot_002",
		ProviderProfileID: "provider_manual_handoff",
		CreatedBy:         "test",
		CorrelationID:     "corr-package-export",
	})
	if !result.OK || result.Package == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want ready package", result)
	}
	pkg := *result.Package
	if pkg.GenerationPackageStatus != GenerationPackageStatusReady {
		t.Fatalf("package status = %q, want %s", pkg.GenerationPackageStatus, GenerationPackageStatusReady)
	}
	if pkg.PackageVersion != 1 {
		t.Fatalf("package version = %d, want 1", pkg.PackageVersion)
	}
	if len(pkg.References) < 3 {
		t.Fatalf("package references = %#v, want shot/profile references", pkg.References)
	}
	if !strings.Contains(pkg.RelativePath, "packages/scene_001/shot_002_pkg_") || strings.Contains(pkg.RelativePath, "_v001") {
		t.Fatalf("package relative path = %q, want scene/shot index timestamp path without version suffix", pkg.RelativePath)
	}

	manifestPath := filepath.Join(root, filepath.FromSlash(pkg.ManifestPath))
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read package manifest: %v", err)
	}
	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode package manifest: %v", err)
	}
	if manifest.ScriptExcerptPath != "script_excerpt.md" {
		t.Fatalf("scriptExcerptPath = %q, want script_excerpt.md", manifest.ScriptExcerptPath)
	}
	if manifest.GenerationPackageStatus != GenerationPackageStatusReady {
		t.Fatalf("manifest status = %q, want ready", manifest.GenerationPackageStatus)
	}
	for _, relative := range append([]string{manifest.PromptPath, manifest.ScriptExcerptPath, manifest.ContinuityPath, manifest.UploadChecklistPath}, manifest.References...) {
		if strings.HasPrefix(relative, "/") || strings.Contains(relative, "..") {
			t.Fatalf("manifest reference %q is not package-relative", relative)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(relative))); err != nil {
			t.Fatalf("manifest reference %q missing: %v", relative, err)
		}
	}

	prompt := readTextFile(t, filepath.Join(root, filepath.FromSlash(pkg.PromptPath)))
	excerpt := readTextFile(t, filepath.Join(root, filepath.FromSlash(pkg.ScriptExcerptPath)))
	continuity := readTextFile(t, filepath.Join(root, filepath.FromSlash(pkg.ContinuityPath)))
	checklist := readTextFile(t, filepath.Join(root, filepath.FromSlash(pkg.UploadChecklistPath)))
	for label, content := range map[string]string{
		"prompt":     prompt,
		"excerpt":    excerpt,
		"continuity": continuity,
		"checklist":  checklist,
	} {
		if strings.TrimSpace(content) == "" {
			t.Fatalf("%s is empty", label)
		}
		assertNoSensitiveText(t, label, content)
	}
	if !strings.Contains(excerpt, "灯笼被风吹偏") {
		t.Fatalf("script excerpt does not include selected source range:\n%s", excerpt)
	}
	if !strings.Contains(continuity, "rule_prop_lantern_position") {
		t.Fatalf("continuity snapshot missing shot rule:\n%s", continuity)
	}

	shot := readShotJSON(t, root, "shot_002")
	if !containsString(shot.PackageIDs, pkg.PackageID) {
		t.Fatalf("shot packageIds = %#v, want %s", shot.PackageIDs, pkg.PackageID)
	}
	graph := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "corr-package-graph"})
	if !graph.OK || graph.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want package node visible", graph)
	}
	found := false
	for _, node := range graph.Canvas.Nodes {
		if node.RefID == pkg.ManifestPath && node.Kind == "package" && node.Status == GenerationPackageStatusReady {
			found = true
		}
	}
	if !found {
		t.Fatalf("GraphView nodes missing exported package %s", pkg.ManifestPath)
	}
	if result.Health == nil || result.Health.HasBlocking() {
		t.Fatalf("export health = %#v, want no blocking items", result.Health)
	}
}

func TestGenerationPackageReexportCreatesNewVersion(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 13, 12, 0, 0, time.UTC), "package-reexport")

	first := store.ExportGenerationPackage(ExportGenerationPackageCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-package-first"})
	if !first.OK || first.Package == nil {
		t.Fatalf("first export = %#v, want ok", first)
	}
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest after first export: %v", err)
	}
	if packages := store.readPackageManifests(root, manifest)["shot_002"]; len(packages) == 0 {
		t.Fatalf("readPackageManifests found no package for shot_002 after first export; first=%#v", first.Package)
	} else if packages[0].PackageVersion != 1 {
		t.Fatalf("readPackageManifests package = %#v, want version 1", packages[0])
	}
	second := store.ExportGenerationPackage(ExportGenerationPackageCommand{Root: root, ShotID: "shot_002", CorrelationID: "corr-package-second"})
	if !second.OK || second.Package == nil {
		t.Fatalf("second export = %#v, want ok", second)
	}
	if first.Package.RelativePath == second.Package.RelativePath {
		t.Fatalf("re-export reused package path %q", first.Package.RelativePath)
	}
	if second.Package.PackageVersion != first.Package.PackageVersion+1 {
		t.Fatalf("second version = %d, want %d", second.Package.PackageVersion, first.Package.PackageVersion+1)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(first.Package.ManifestPath))); err != nil {
		t.Fatalf("first package manifest was overwritten or removed: %v", err)
	}
}

func TestGenerationPackageExportRejectsSensitiveManifestMetadata(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 13, 18, 0, 0, time.UTC), "package-sensitive-manifest")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:              root,
		ShotID:            "shot_002",
		ProviderProfileID: "token=secret-provider",
		CorrelationID:     "corr-package-sensitive-manifest",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want redaction error", result)
	}
	if result.Error.Code != CodePackageRedactionFailed {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodePackageRedactionFailed)
	}
}

func TestGenerationPackageExportRejectsSensitiveTextReference(t *testing.T) {
	root := copyExampleFixture(t)
	if err := os.WriteFile(filepath.Join(root, "assets", "refs", "prop-red-lantern.txt"), []byte("token=reference-secret\n"), 0o644); err != nil {
		t.Fatalf("write sensitive reference: %v", err)
	}
	store := testStore(time.Date(2026, 5, 20, 13, 19, 0, 0, time.UTC), "package-sensitive-reference")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-sensitive-reference",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want redaction error", result)
	}
	if result.Error.Code != CodePackageRedactionFailed {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodePackageRedactionFailed)
	}
}

func TestGenerationPackageExportIncludesCharacterReferenceAsset(t *testing.T) {
	root := copyExampleFixture(t)
	writeFile(t, filepath.Join(root, "assets", "refs", "character-pose.txt"), "Mina direct pose reference")
	appendPackageTestAsset(t, root, AssetDTO{
		ID:              "asset_ref_character_pose",
		ProjectID:       "proj_alpha_fixture",
		Type:            "text",
		Role:            "character_ref",
		RelativePath:    "assets/refs/character-pose.txt",
		OriginalName:    "character-pose.txt",
		MimeType:        "text/plain",
		SizeBytes:       26,
		Digest:          "sha256-character-pose",
		ThumbnailStatus: "none",
		CreatedAt:       "2026-05-20T05:40:00Z",
		UpdatedAt:       "2026-05-20T05:40:00Z",
	})
	shot := readShotJSON(t, root, "shot_002")
	shot.CharacterRefs = append(shot.CharacterRefs, ShotCharacterRefDTO{
		CharacterID:      "char_mina",
		Name:             "Mina",
		ReferenceAssetID: "asset_ref_character_pose",
	})
	writePackageTestShot(t, root, shot)

	store := testStore(time.Date(2026, 5, 20, 13, 22, 0, 0, time.UTC), "package-character-ref")
	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-character-ref",
	})
	if !result.OK || result.Package == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want package with character reference", result)
	}
	ref, ok := packageReferenceByAssetID(result.Package.References, "asset_ref_character_pose")
	if !ok {
		t.Fatalf("package references = %#v, want character reference asset", result.Package.References)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(result.Package.RelativePath), filepath.FromSlash(ref.PackagePath))); err != nil {
		t.Fatalf("character reference package copy missing: %v", err)
	}
}

func TestGenerationPackageExportRejectsMissingCharacterReferenceAsset(t *testing.T) {
	root := copyExampleFixture(t)
	shot := readShotJSON(t, root, "shot_002")
	shot.CharacterRefs = append(shot.CharacterRefs, ShotCharacterRefDTO{
		CharacterID:      "char_mina",
		Name:             "Mina",
		ReferenceAssetID: "asset_ref_character_missing",
	})
	writePackageTestShot(t, root, shot)
	store := testStore(time.Date(2026, 5, 20, 13, 23, 0, 0, time.UTC), "package-character-ref-missing")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-character-ref-missing",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want missing character reference", result)
	}
	if result.Error.Code != CodePackageReferenceMissing {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodePackageReferenceMissing)
	}
}

func TestGenerationPackageExportRejectsMissingShotGraphNode(t *testing.T) {
	root := copyExampleFixture(t)
	removePackageTestShotGraphNode(t, root, "shot_002")
	store := testStore(time.Date(2026, 5, 20, 13, 24, 0, 0, time.UTC), "package-shot-node-missing")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-shot-node-missing",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want missing graph node", result)
	}
	if result.Error.Code != CodeShotGraphNodeMissing {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeShotGraphNodeMissing)
	}
	shot := readShotJSON(t, root, "shot_002")
	if len(shot.PackageIDs) != 0 {
		t.Fatalf("shot packageIds = %#v, want no linkage after blocked export", shot.PackageIDs)
	}
}

func TestGenerationPackageMarkedStaleWhenShotDirty(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 13, 20, 0, 0, time.UTC), "package-stale")

	exported := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-before-dirty",
	})
	if !exported.OK || exported.Package == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want ready package", exported)
	}
	dirty := store.MarkShotContextDirty(MarkShotContextDirtyCommand{
		Root:          root,
		ShotID:        "shot_002",
		Reason:        "manual revision",
		CorrelationID: "corr-package-dirty",
	})
	if !dirty.OK {
		t.Fatalf("MarkShotContextDirty() = %#v, want ok", dirty)
	}

	manifest := readPackageManifestForTest(t, root, exported.Package.ManifestPath)
	if manifest.GenerationPackageStatus != GenerationPackageStatusStale {
		t.Fatalf("package manifest status = %q, want stale", manifest.GenerationPackageStatus)
	}
	graph := store.GraphView(GraphViewCommand{Root: root, CorrelationID: "corr-package-stale-graph"})
	if !graph.OK || graph.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want canvas", graph)
	}
	var packageNode *GraphNodeDTO
	for index := range graph.Canvas.Nodes {
		if graph.Canvas.Nodes[index].RefID == exported.Package.ManifestPath {
			packageNode = &graph.Canvas.Nodes[index]
			break
		}
	}
	if packageNode == nil || packageNode.Status != GenerationPackageStatusStale {
		t.Fatalf("package graph node = %#v, want stale", packageNode)
	}
}

func TestGenerationPackageMarkedStaleWhenContinuityDirty(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 13, 21, 0, 0, time.UTC), "package-continuity-stale")

	exported := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-continuity-before",
	})
	if !exported.OK || exported.Package == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want ready package", exported)
	}
	changed := store.SaveContinuityRule(SaveContinuityRuleCommand{
		Root:          root,
		ID:            "rule_prop_lantern_position",
		TargetType:    BindingTargetProp,
		TargetID:      "prop_red_lantern",
		Rule:          "The red lantern stays fixed as the warm practical light cue.",
		Severity:      ContinuitySeverityWarning,
		Locked:        true,
		CreatedBy:     "test",
		CorrelationID: "corr-package-continuity-dirty",
	})
	if !changed.OK {
		t.Fatalf("SaveContinuityRule() = %#v, want ok", changed)
	}

	manifest := readPackageManifestForTest(t, root, exported.Package.ManifestPath)
	if manifest.GenerationPackageStatus != GenerationPackageStatusStale {
		t.Fatalf("package manifest status = %q, want stale", manifest.GenerationPackageStatus)
	}
}

func TestGenerationPackageExportRejectsNonReadyShot(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 13, 14, 0, 0, time.UTC), "package-not-ready")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_001",
		CorrelationID: "corr-package-not-ready",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want blocking error", result)
	}
	if result.Error.Code != CodePackageShotNotReady {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodePackageShotNotReady)
	}
}

func TestGenerationPackageExportRejectsMissingReference(t *testing.T) {
	root := copyExampleFixture(t)
	if err := os.Remove(filepath.Join(root, "assets", "refs", "prop-red-lantern.txt")); err != nil {
		t.Fatalf("remove prop ref: %v", err)
	}
	store := testStore(time.Date(2026, 5, 20, 13, 16, 0, 0, time.UTC), "package-missing-ref")

	result := store.ExportGenerationPackage(ExportGenerationPackageCommand{
		Root:          root,
		ShotID:        "shot_002",
		CorrelationID: "corr-package-missing-ref",
	})
	if result.OK || result.Error == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want missing reference", result)
	}
	if result.Error.Code != CodePackageReferenceMissing {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodePackageReferenceMissing)
	}
	if len(result.Error.RecoveryActions) == 0 {
		t.Fatalf("missing reference error lacks recovery actions: %#v", result.Error)
	}
}

func readShotJSON(t *testing.T, root string, shotID string) ShotCardDTO {
	t.Helper()
	store := testStore(time.Date(2026, 5, 20, 13, 18, 0, 0, time.UTC), "package-read-shot")
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	filename := shotPath(root, manifest, shotID)
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read shot %s: %v", shotID, err)
	}
	var shot ShotCardDTO
	if err := json.Unmarshal(data, &shot); err != nil {
		t.Fatalf("decode shot %s: %v", shotID, err)
	}
	return shot
}

func readPackageManifestForTest(t *testing.T, root string, relative string) packageManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatalf("read package manifest %s: %v", relative, err)
	}
	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode package manifest %s: %v", relative, err)
	}
	return manifest
}

func appendPackageTestAsset(t *testing.T, root string, asset AssetDTO) {
	t.Helper()
	filename := filepath.Join(root, "assets", "index.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read asset index: %v", err)
	}
	var index AssetIndex
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("decode asset index: %v", err)
	}
	index.Assets = append(index.Assets, asset)
	writeJSON(t, filename, index)
}

func writePackageTestShot(t *testing.T, root string, shot ShotCardDTO) {
	t.Helper()
	store := testStore(time.Date(2026, 5, 20, 13, 24, 0, 0, time.UTC), "package-write-shot")
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	writeJSON(t, shotPath(root, manifest, shot.ID), shot)
}

func packageReferenceByAssetID(references []GenerationPackageReferenceDTO, assetID string) (GenerationPackageReferenceDTO, bool) {
	for _, reference := range references {
		if reference.AssetID == assetID {
			return reference, true
		}
	}
	return GenerationPackageReferenceDTO{}, false
}

func removePackageTestShotGraphNode(t *testing.T, root string, shotID string) {
	t.Helper()
	store := testStore(time.Date(2026, 5, 20, 13, 24, 0, 0, time.UTC), "package-remove-shot-node")
	manifest, _, err := store.readManifest(root)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	shotNodeID := manifestShotNodeID(root, manifest, shotID)
	if shotNodeID == "" {
		t.Fatalf("shot graph node for %s was already missing", shotID)
	}
	nodes := manifest.Graph.Nodes[:0]
	for _, node := range manifest.Graph.Nodes {
		if node.ID != shotNodeID {
			nodes = append(nodes, node)
		}
	}
	manifest.Graph.Nodes = nodes
	manifest.Graph.Version++
	manifest.Integrity.LastGraphVersion = manifest.Graph.Version
	writeManifestBypassValidation(t, root, manifest)
}
