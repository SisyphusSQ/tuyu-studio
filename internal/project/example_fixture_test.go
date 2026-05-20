package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestExampleAlphaProjectOpen(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 5, 45, 0, 0, time.UTC), "fixture-open")

	result := store.OpenProject(OpenProjectCommand{
		Root:          root,
		CorrelationID: "fixture-open",
	})
	if !result.OK {
		t.Fatalf("OpenProject() failed: %#v", result.Error)
	}
	if result.Summary == nil {
		t.Fatal("OpenProject() summary = nil")
	}
	if result.Summary.ProjectID != "proj_alpha_fixture" {
		t.Fatalf("ProjectID = %q, want proj_alpha_fixture", result.Summary.ProjectID)
	}
	if result.Summary.OpenMode != OpenModeReadWrite {
		t.Fatalf("OpenMode = %q, want %s", result.Summary.OpenMode, OpenModeReadWrite)
	}
}

func TestExampleAlphaProjectHealthCheck(t *testing.T) {
	root := exampleFixtureRoot(t)
	store := testStore(time.Date(2026, 5, 20, 5, 46, 0, 0, time.UTC), "fixture-health")

	report := store.HealthReport(root)
	if report.HasBlocking() {
		t.Fatalf("HealthReport() has blocking items: %#v", report.Items)
	}
	if report.Status != HealthStatusClean {
		t.Fatalf("HealthReport().Status = %q, want %s; items=%#v", report.Status, HealthStatusClean, report.Items)
	}
}

func TestExampleAlphaProjectCanvasReopenSmoke(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 8, 20, 0, 0, time.UTC), "fixture-canvas-open")

	opened := store.OpenProject(OpenProjectCommand{
		Root:          root,
		CorrelationID: "fixture-canvas-open",
	})
	if !opened.OK || opened.Summary == nil {
		t.Fatalf("OpenProject() = %#v, want open fixture", opened)
	}
	if opened.Summary.OpenMode != OpenModeReadWrite {
		t.Fatalf("OpenMode = %q, want %s", opened.Summary.OpenMode, OpenModeReadWrite)
	}
	if opened.Health == nil || opened.Health.HasBlocking() {
		t.Fatalf("open health = %#v, want no blocking items", opened.Health)
	}

	initial := store.GraphView(GraphViewCommand{
		Root:          root,
		CorrelationID: "fixture-canvas-view",
	})
	if !initial.OK || initial.Canvas == nil {
		t.Fatalf("GraphView() = %#v, want initial Canvas", initial)
	}
	requireAlphaCanvasDisplay(t, *initial.Canvas)

	saved := store.SaveGraphLayout(SaveGraphLayoutCommand{
		Root:                 root,
		ExpectedGraphVersion: initial.Canvas.Version,
		Viewport:             CanvasViewportDTO{X: -144, Y: 96, Zoom: 1.35},
		Theme:                "warm_light",
		Grid:                 CanvasGridDTO{Visible: false, Size: 32, Opacity: 0.16},
		Nodes: []GraphNodeLayoutCommand{
			{
				ID:        "node_shot_001",
				Position:  CanvasPositionDTO{X: 812, Y: 264},
				Size:      CanvasSizeDTO{Width: 276, Height: 136},
				Collapsed: false,
			},
		},
		CorrelationID: "fixture-canvas-save",
	})
	if !saved.OK || saved.Canvas == nil {
		t.Fatalf("SaveGraphLayout() = %#v, want saved Canvas", saved)
	}

	reopenedStore := testStore(time.Date(2026, 5, 20, 8, 22, 0, 0, time.UTC), "fixture-canvas-reopen")
	reopened := reopenedStore.OpenProject(OpenProjectCommand{
		Root:          root,
		Takeover:      true,
		CorrelationID: "fixture-canvas-reopen",
	})
	if !reopened.OK || reopened.Summary == nil {
		t.Fatalf("reopen OpenProject() = %#v, want reopened fixture", reopened)
	}
	if reopened.Summary.GraphVersion != initial.Canvas.Version+1 {
		t.Fatalf("reopened graph version = %d, want %d", reopened.Summary.GraphVersion, initial.Canvas.Version+1)
	}

	restored := reopenedStore.GraphView(GraphViewCommand{
		Root:          root,
		CorrelationID: "fixture-canvas-restored",
	})
	if !restored.OK || restored.Canvas == nil {
		t.Fatalf("reopened GraphView() = %#v, want restored Canvas", restored)
	}
	requireAlphaCanvasDisplay(t, *restored.Canvas)
	if restored.Canvas.Version != initial.Canvas.Version+1 {
		t.Fatalf("restored canvas version = %d, want %d", restored.Canvas.Version, initial.Canvas.Version+1)
	}
	if restored.Canvas.Theme != "warm_light" {
		t.Fatalf("restored theme = %q, want warm_light", restored.Canvas.Theme)
	}
	if restored.Canvas.Grid.Visible || restored.Canvas.Grid.Size != 32 || restored.Canvas.Grid.Opacity != 0.16 {
		t.Fatalf("restored grid = %#v, want saved grid", restored.Canvas.Grid)
	}
	if restored.Canvas.Viewport.X != -144 || restored.Canvas.Viewport.Y != 96 || restored.Canvas.Viewport.Zoom != 1.35 {
		t.Fatalf("restored viewport = %#v, want saved viewport", restored.Canvas.Viewport)
	}

	shot := requireNode(t, restored.Canvas.Nodes, "node_shot_001", "shot", "production", "Mina reaches the stall shelter")
	if shot.Position.X != 812 || shot.Position.Y != 264 {
		t.Fatalf("restored shot position = %#v, want saved position", shot.Position)
	}
	if shot.Size.Width != 276 || shot.Size.Height != 136 {
		t.Fatalf("restored shot size = %#v, want saved size", shot.Size)
	}
}

func TestExampleAlphaProjectPaths(t *testing.T) {
	root := exampleFixtureRoot(t)

	for _, dir := range RequiredProjectDirectories() {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			t.Fatalf("required fixture directory %q missing: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("required fixture path %q is not a directory", dir)
		}
	}

	data, err := os.ReadFile(filepath.Join(root, ManifestFileName))
	if err != nil {
		t.Fatalf("read fixture manifest: %v", err)
	}
	manifest, report := DecodeManifest(data)
	if report.HasBlocking() {
		t.Fatalf("fixture manifest has blocking report: %#v", report.Items)
	}
	if manifest.Graph.Nodes == nil || manifest.Graph.Edges == nil {
		t.Fatal("fixture graph nodes/edges must be explicit arrays")
	}

	assertFixtureCoverage(t, root)
	assertFixtureGraphReferences(t, root, manifest)
	assertFixtureJSONFilesParse(t, root)
	assertFixtureDigestIndex(t, root)
	assertFixtureHasNoSensitivePatterns(t, root)
}

func requireAlphaCanvasDisplay(t *testing.T, canvas ProjectCanvasDTO) {
	t.Helper()

	if canvas.ID == "" || canvas.ProjectID != "proj_alpha_fixture" {
		t.Fatalf("canvas id/project = %q/%q, want alpha fixture Canvas", canvas.ID, canvas.ProjectID)
	}
	if len(canvas.Nodes) < 8 {
		t.Fatalf("canvas nodes = %d, want projected alpha fixture nodes", len(canvas.Nodes))
	}
	if len(canvas.Edges) < 6 {
		t.Fatalf("canvas edges = %d, want projected alpha fixture relations", len(canvas.Edges))
	}
	if len(canvas.ReferenceGroups) != 1 {
		t.Fatalf("reference groups = %d, want 1", len(canvas.ReferenceGroups))
	}
	if len(canvas.Frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(canvas.Frames))
	}

	script := requireNode(t, canvas.Nodes, "node_script_source", "script", "source", "Script Source")
	if !containsString(script.Badges, "context_ready") {
		t.Fatalf("script badges = %#v, want context_ready", script.Badges)
	}
	shot := requireNode(t, canvas.Nodes, "node_shot_001", "shot", "production", "Mina reaches the stall shelter")
	if !containsString(shot.Badges, "context_dirty") {
		t.Fatalf("shot badges = %#v, want context_dirty", shot.Badges)
	}
	packageNode := requireNode(t, canvas.Nodes, "node_package_001", "package", "handoff", "Package pkg_scene001_shot001")
	if !containsString(packageNode.Badges, "package_ready") {
		t.Fatalf("package badges = %#v, want package_ready", packageNode.Badges)
	}
	review := requireNode(t, canvas.Nodes, "node_review_001", "note", "review", "Review note review_stub_001")
	if !containsString(review.Badges, "pending_review") {
		t.Fatalf("review badges = %#v, want pending_review", review.Badges)
	}

	for _, edge := range canvas.Edges {
		if edge.Validity != GraphEdgeValidityValid {
			t.Fatalf("edge %q validity = %q error=%#v, want valid", edge.ID, edge.Validity, edge.Error)
		}
	}
	if len(canvas.Frames[0].OutputNodeIDs) < 3 {
		t.Fatalf("frame outputs = %#v, want package/result/review outputs", canvas.Frames[0].OutputNodeIDs)
	}
	if len(canvas.ReferenceGroups[0].InputNodeIDs) < 4 {
		t.Fatalf("reference group inputs = %#v, want script/context inputs", canvas.ReferenceGroups[0].InputNodeIDs)
	}
}

func assertFixtureCoverage(t *testing.T, root string) {
	t.Helper()

	requiredFiles := []string{
		"README.md",
		"project.tuyu.json",
		"assets/inputs/script-source.md",
		"characters/char-mina.json",
		"scenes/scene-laneway-market.json",
		"props/prop-red-lantern.json",
		"shots/shot-001.json",
		"shots/shot-002.json",
		"assets/refs/character-mina.txt",
		"assets/refs/scene-laneway-market.txt",
		"assets/refs/prop-red-lantern.txt",
		"assets/results/mock-result-summary.md",
		"prompts/runs/mock-run-001.json",
		"packages/scene-001/shot-001-package/manifest.json",
		"audit/review-stub-001.json",
		"audit/digests.json",
	}

	for _, file := range requiredFiles {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			t.Fatalf("fixture coverage file %q missing: %v", file, err)
		}
		if info.IsDir() {
			t.Fatalf("fixture coverage path %q is a directory", file)
		}
	}
}

func assertFixtureGraphReferences(t *testing.T, root string, manifest Manifest) {
	t.Helper()

	if len(manifest.Graph.Nodes) == 0 {
		t.Fatal("fixture graph must contain nodes")
	}
	if len(manifest.Graph.Edges) == 0 {
		t.Fatal("fixture graph must contain edges")
	}

	nodeIDs := make(map[string]struct{}, len(manifest.Graph.Nodes))
	for _, node := range manifest.Graph.Nodes {
		if node.ID == "" {
			t.Fatal("fixture graph node has empty id")
		}
		if _, exists := nodeIDs[node.ID]; exists {
			t.Fatalf("fixture graph node id %q is duplicated", node.ID)
		}
		nodeIDs[node.ID] = struct{}{}

		if node.RefID == "" {
			t.Fatalf("fixture graph node %q has empty refId", node.ID)
		}
		if issues := validateProjectPath("graph.refId", node.RefID); len(issues) > 0 {
			t.Fatalf("fixture graph node %q has invalid refId %q: %#v", node.ID, node.RefID, issues)
		}
		refPath := filepath.Join(root, filepath.FromSlash(node.RefID))
		info, err := os.Stat(refPath)
		if err != nil {
			t.Fatalf("fixture graph node %q refId %q missing: %v", node.ID, node.RefID, err)
		}
		if info.IsDir() {
			t.Fatalf("fixture graph node %q refId %q is a directory", node.ID, node.RefID)
		}
		if node.Kind == "package" {
			assertFixturePackageReferences(t, refPath)
		}
	}

	edgeIDs := make(map[string]struct{}, len(manifest.Graph.Edges))
	for _, edge := range manifest.Graph.Edges {
		if edge.ID == "" {
			t.Fatal("fixture graph edge has empty id")
		}
		if _, exists := edgeIDs[edge.ID]; exists {
			t.Fatalf("fixture graph edge id %q is duplicated", edge.ID)
		}
		edgeIDs[edge.ID] = struct{}{}
		if _, exists := nodeIDs[edge.Source]; !exists {
			t.Fatalf("fixture graph edge %q source %q does not exist", edge.ID, edge.Source)
		}
		if _, exists := nodeIDs[edge.Target]; !exists {
			t.Fatalf("fixture graph edge %q target %q does not exist", edge.ID, edge.Target)
		}
	}
}

func assertFixturePackageReferences(t *testing.T, manifestPath string) {
	t.Helper()

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read package manifest %q: %v", manifestPath, err)
	}

	var manifest struct {
		PromptPath          string   `json:"promptPath"`
		ContinuityPath      string   `json:"continuityPath"`
		UploadChecklistPath string   `json:"uploadChecklistPath"`
		References          []string `json:"references"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode package manifest %q: %v", manifestPath, err)
	}

	packageRoot := filepath.Dir(manifestPath)
	packagePaths := append([]string{
		manifest.PromptPath,
		manifest.ContinuityPath,
		manifest.UploadChecklistPath,
	}, manifest.References...)

	for _, path := range packagePaths {
		if path == "" {
			t.Fatalf("package manifest %q contains an empty package-local path", manifestPath)
		}
		if issues := validateProjectPath("package.path", path); len(issues) > 0 {
			t.Fatalf("package manifest %q has invalid package-local path %q: %#v", manifestPath, path, issues)
		}
		target := filepath.Join(packageRoot, filepath.FromSlash(path))
		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("package manifest %q references missing file %q: %v", manifestPath, path, err)
		}
		if info.IsDir() {
			t.Fatalf("package manifest %q references directory %q", manifestPath, path)
		}
	}
}

func assertFixtureJSONFilesParse(t *testing.T, root string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("fixture JSON parse failed: %v", err)
	}
}

func assertFixtureDigestIndex(t *testing.T, root string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(DigestIndexRelativePath)))
	if err != nil {
		t.Fatalf("read digest index: %v", err)
	}

	var index DigestIndex
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("decode digest index: %v", err)
	}
	if len(index.Files) < 8 {
		t.Fatalf("digest index files = %d, want at least 8", len(index.Files))
	}

	for _, entry := range index.Files {
		if issues := validateProjectPath("digest.path", entry.Path); len(issues) > 0 {
			t.Fatalf("digest path %q is invalid: %#v", entry.Path, issues)
		}
		fileData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.Path)))
		if err != nil {
			t.Fatalf("read digest target %q: %v", entry.Path, err)
		}
		sum := sha256.Sum256(fileData)
		got := hex.EncodeToString(sum[:])
		if !strings.EqualFold(got, entry.SHA256) {
			t.Fatalf("digest mismatch for %q: got %s want %s", entry.Path, got, entry.SHA256)
		}
	}
}

func assertFixtureHasNoSensitivePatterns(t *testing.T, root string) {
	t.Helper()

	forbidden := regexp.MustCompile(`(?i)(/Users/|/var/folders|/tmp/|https?://|sk-[a-z0-9]|ghp_[a-z0-9]|xox[baprs]-|api[_-]?key=|token=|secret=|credential=)`)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if forbidden.Match(data) {
			relative, _ := filepath.Rel(root, path)
			t.Fatalf("fixture file %q contains a forbidden sensitive/path pattern", filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("fixture sensitive scan failed: %v", err)
	}
}

func copyExampleFixture(t *testing.T) string {
	t.Helper()

	source := exampleFixtureRoot(t)
	target := filepath.Join(t.TempDir(), "alpha-project")
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}

		targetPath := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy example fixture: %v", err)
	}

	return target
}

func exampleFixtureRoot(t *testing.T) string {
	t.Helper()

	root := filepath.Join("..", "..", "examples", "alpha-project")
	if _, err := os.Stat(filepath.Join(root, ManifestFileName)); err != nil {
		t.Fatalf("example fixture manifest missing: %v", err)
	}
	return root
}
