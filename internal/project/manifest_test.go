package project

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProjectDirectoryLayoutCreatesRequiredSkeleton(t *testing.T) {
	manifest := testManifest(t)
	root := t.TempDir()

	if err := CreateBaselineProject(root, manifest); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}

	for _, dir := range RequiredProjectDirectories() {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			t.Fatalf("required directory %q missing: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("required path %q is not a directory", dir)
		}
	}

	if _, err := os.Stat(filepath.Join(root, ManifestFileName)); err != nil {
		t.Fatalf("manifest file missing: %v", err)
	}
}

func TestProjectManifestBaselineShapeMatchesGolden(t *testing.T) {
	manifest := testManifest(t)
	encoded, err := EncodeManifest(manifest)
	if err != nil {
		t.Fatalf("EncodeManifest() error = %v", err)
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "minimal-project.tuyu.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if string(encoded) != string(golden) {
		t.Fatalf("encoded manifest mismatch\nwant:\n%s\ngot:\n%s", string(golden), string(encoded))
	}

	decoded, report := DecodeManifest(encoded)
	if report.HasBlocking() {
		t.Fatalf("DecodeManifest() report has blocking issues: %#v", report.Items)
	}
	if decoded.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("SchemaVersion = %q, want %q", decoded.SchemaVersion, CurrentSchemaVersion)
	}
	if decoded.Graph.Nodes == nil || decoded.Graph.Edges == nil {
		t.Fatal("graph nodes and edges must be empty arrays, not nil")
	}
	if decoded.Graph.Viewport.Zoom != 1 {
		t.Fatalf("Viewport.Zoom = %v, want 1", decoded.Graph.Viewport.Zoom)
	}
}

func TestProjectPathValidationRejectsUnsafePaths(t *testing.T) {
	tests := []struct {
		name string
		path string
		code string
	}{
		{name: "absolute unix", path: "/tmp/tuyu-project/assets", code: CodePathAbsolute},
		{name: "absolute windows", path: `C:\Users\demo\assets`, code: CodePathAbsolute},
		{name: "drive relative windows", path: `C:assets`, code: CodePathAbsolute},
		{name: "rooted windows", path: `\Users\demo\private.png`, code: CodePathAbsolute},
		{name: "home", path: "~/assets", code: CodePathAbsolute},
		{name: "url", path: "https://example.invalid/asset.png", code: CodePathURL},
		{name: "file url", path: "file:///Users/demo/private.png", code: CodePathURL},
		{name: "data url", path: "data:text/plain,asset", code: CodePathURL},
		{name: "traversal", path: "../outside", code: CodePathTraversal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := testManifest(t)
			manifest.Paths.AssetRefs = tt.path

			report := ValidateManifest(manifest)
			if !report.HasBlocking() {
				t.Fatalf("ValidateManifest() has no blocking issues for path %q", tt.path)
			}
			if !hasCode(report, tt.code) {
				t.Fatalf("ValidateManifest() codes = %v, want %s", report.Codes(), tt.code)
			}
		})
	}
}

func TestProjectManifestValidationReportsDistinctErrors(t *testing.T) {
	_, invalidJSON := DecodeManifest([]byte(`{"schemaVersion":`))
	if !hasCode(invalidJSON, CodeManifestJSONInvalid) {
		t.Fatalf("invalid JSON codes = %v, want %s", invalidJSON.Codes(), CodeManifestJSONInvalid)
	}

	missingTopLevel := []byte(`{"schemaVersion":"1.0.0"}`)
	_, missingReport := DecodeManifest(missingTopLevel)
	wantMissingCodes := []string{CodeGraphPartitionMissing, CodeManifestRequiredMissing}
	for _, code := range wantMissingCodes {
		if !hasCode(missingReport, code) {
			t.Fatalf("missing report codes = %v, want %s", missingReport.Codes(), code)
		}
	}

	manifest := testManifest(t)
	manifest.SchemaVersion = "9.9.9"
	unsupported := ValidateManifest(manifest)
	if !hasCode(unsupported, CodeSchemaUnsupported) {
		t.Fatalf("unsupported schema codes = %v, want %s", unsupported.Codes(), CodeSchemaUnsupported)
	}

	manifest = testManifest(t)
	manifest.Graph.Nodes = nil
	missingGraph := ValidateManifest(manifest)
	if !hasCode(missingGraph, CodeGraphPartitionMissing) {
		t.Fatalf("missing graph codes = %v, want %s", missingGraph.Codes(), CodeGraphPartitionMissing)
	}
}

func TestProjectManifestRejectsSensitiveValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Manifest)
	}{
		{
			name: "token value",
			mutate: func(manifest *Manifest) {
				manifest.Project.Name = "token=secret-value"
			},
		},
		{
			name: "download URL",
			mutate: func(manifest *Manifest) {
				manifest.Principles.Summary = "https://example.invalid/private-download"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := testManifest(t)
			tt.mutate(&manifest)

			report := ValidateManifest(manifest)
			if !hasCode(report, CodeSensitiveValue) {
				t.Fatalf("ValidateManifest() codes = %v, want %s", report.Codes(), CodeSensitiveValue)
			}
		})
	}
}

func TestRequiredProjectDirectoriesAreStable(t *testing.T) {
	want := []string{
		"characters",
		"scenes",
		"props",
		"shots",
		"prompts",
		"prompts/runs",
		"assets",
		"assets/inputs",
		"assets/refs",
		"assets/outputs",
		"assets/results",
		"assets/thumbnails",
		"packages",
		"audit",
		"backups",
		"locks",
	}

	got := RequiredProjectDirectories()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RequiredProjectDirectories() = %#v, want %#v", got, want)
	}

	got[0] = "mutated"
	if strings.EqualFold(RequiredProjectDirectories()[0], "mutated") {
		t.Fatal("RequiredProjectDirectories must return a defensive copy")
	}
}

func testManifest(t *testing.T) Manifest {
	t.Helper()

	when := time.Date(2026, 5, 20, 4, 0, 0, 0, time.UTC)
	manifest, err := NewBaselineManifest(ManifestInput{
		ProjectID: "proj_alpha_001",
		Name:      "Alpha Fixture",
		Type:      "series",
		CreatedAt: when,
		UpdatedAt: when,
	})
	if err != nil {
		t.Fatalf("NewBaselineManifest() error = %v", err)
	}

	return manifest
}

func hasCode(report ValidationReport, code string) bool {
	for _, got := range report.Codes() {
		if got == code {
			return true
		}
	}
	return false
}
