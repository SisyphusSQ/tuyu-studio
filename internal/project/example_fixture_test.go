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
