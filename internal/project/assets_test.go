package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAssetImportCreatesProjectRelativeRecord(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)
	store := testStore(now, "asset-import")
	root := createAssetTestProject(t)
	source := filepath.Join(t.TempDir(), "dialogue-source.md")
	if err := os.WriteFile(source, []byte("# Scene\nMina enters the market.\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result := store.ImportAsset(ImportAssetCommand{
		Root:          root,
		SourcePath:    source,
		Role:          "script_source",
		CorrelationID: "asset-import",
	})
	if !result.OK || result.Asset == nil {
		t.Fatalf("ImportAsset() = %#v, want imported asset", result)
	}

	asset := result.Asset
	if asset.ProjectID != "proj_alpha_001" {
		t.Fatalf("ProjectID = %q, want proj_alpha_001", asset.ProjectID)
	}
	if asset.Type != AssetTypeText || asset.Role != "script_source" {
		t.Fatalf("asset type/role = %q/%q, want text/script_source", asset.Type, asset.Role)
	}
	if filepath.IsAbs(asset.RelativePath) || strings.Contains(asset.RelativePath, "..") {
		t.Fatalf("relativePath = %q, want project-relative safe path", asset.RelativePath)
	}
	if !strings.HasPrefix(asset.RelativePath, "assets/inputs/") {
		t.Fatalf("relativePath = %q, want assets/inputs", asset.RelativePath)
	}
	if !strings.Contains(filepath.Base(asset.RelativePath), asset.ID) {
		t.Fatalf("relativePath = %q, want file name to contain asset id %q", asset.RelativePath, asset.ID)
	}
	if asset.Digest == "" || len(asset.Digest) != 64 || asset.DigestSummary == "" {
		t.Fatalf("digest = %q summary=%q, want sha256 digest", asset.Digest, asset.DigestSummary)
	}
	if asset.BindingCount != 0 || len(asset.Bindings) != 0 {
		t.Fatalf("bindings = %#v count=%d, want empty bindings", asset.Bindings, asset.BindingCount)
	}
	if asset.Missing {
		t.Fatalf("asset reported missing after import: %#v", asset)
	}
	if asset.ThumbnailStatus != AssetThumbnailPlaceholder || asset.ThumbnailPath == "" {
		t.Fatalf("thumbnail = %q/%q, want placeholder path", asset.ThumbnailStatus, asset.ThumbnailPath)
	}
	copiedAsset, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(asset.RelativePath)))
	if err != nil {
		t.Fatalf("copied asset missing: %v", err)
	}
	copiedDigest := sha256.Sum256(copiedAsset)
	if asset.Digest != hex.EncodeToString(copiedDigest[:]) {
		t.Fatalf("asset digest = %q, want digest of copied project file", asset.Digest)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(asset.ThumbnailPath))); err != nil {
		t.Fatalf("thumbnail placeholder missing: %v", err)
	}

	index := readAssetIndexForTest(t, root)
	if len(index.Assets) != 1 || index.Assets[0].ID != asset.ID {
		t.Fatalf("index assets = %#v, want imported asset", index.Assets)
	}
	if result.Health == nil || result.Health.HasBlocking() {
		t.Fatalf("health = %#v, want no blocking health", result.Health)
	}
}

func TestAssetImportContentMimeTypes(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 2, 0, 0, time.UTC)
	tests := []struct {
		name     string
		filename string
		data     []byte
		wantType string
		wantMime string
	}{
		{
			name:     "png content with binary extension",
			filename: "reference.bin",
			data: []byte{
				0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
				0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
			},
			wantType: AssetTypeImage,
			wantMime: "image/png",
		},
		{
			name:     "mp4 content with text extension",
			filename: "clip.txt",
			data: []byte{
				0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p',
				'i', 's', 'o', 'm', 0x00, 0x00, 0x02, 0x00,
			},
			wantType: AssetTypeVideo,
			wantMime: "video/mp4",
		},
		{
			name:     "wav content with generic extension",
			filename: "voice.data",
			data: []byte{
				'R', 'I', 'F', 'F', 0x24, 0x00, 0x00, 0x00,
				'W', 'A', 'V', 'E', 'f', 'm', 't', ' ',
			},
			wantType: AssetTypeAudio,
			wantMime: "audio/wav",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := testStore(now, "asset-"+tt.wantType)
			root := createAssetTestProject(t)
			source := filepath.Join(t.TempDir(), tt.filename)
			if err := os.WriteFile(source, tt.data, 0o644); err != nil {
				t.Fatalf("write source: %v", err)
			}

			result := store.ImportAsset(ImportAssetCommand{
				Root:          root,
				SourcePath:    source,
				Role:          "other",
				CorrelationID: "asset-mime-" + tt.wantType,
			})
			if !result.OK || result.Asset == nil {
				t.Fatalf("ImportAsset() = %#v, want success", result)
			}
			if result.Asset.Type != tt.wantType || result.Asset.MimeType != tt.wantMime {
				t.Fatalf("asset type/mime = %q/%q, want %q/%q", result.Asset.Type, result.Asset.MimeType, tt.wantType, tt.wantMime)
			}
			if filepath.Ext(result.Asset.RelativePath) == filepath.Ext(tt.filename) && tt.wantType != AssetTypeText {
				t.Fatalf("relativePath = %q reused untrusted source extension %q", result.Asset.RelativePath, filepath.Ext(tt.filename))
			}
		})
	}
}

func TestAssetDuplicatePolicies(t *testing.T) {
	now := time.Date(2026, 5, 20, 9, 5, 0, 0, time.UTC)
	store := testStore(now, "asset-duplicate")
	root := createAssetTestProject(t)
	source := filepath.Join(t.TempDir(), "reference.txt")
	if err := os.WriteFile(source, []byte("same content\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	first := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, Role: "character_ref", CorrelationID: "asset-first"})
	if !first.OK || first.Asset == nil {
		t.Fatalf("first ImportAsset() = %#v, want success", first)
	}

	duplicate := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, Role: "character_ref", CorrelationID: "asset-duplicate"})
	if duplicate.OK || duplicate.Error == nil {
		t.Fatalf("duplicate ImportAsset() = %#v, want duplicate error", duplicate)
	}
	if duplicate.Error.Code != CodeAssetDuplicateFound {
		t.Fatalf("duplicate error = %q, want %s", duplicate.Error.Code, CodeAssetDuplicateFound)
	}
	if duplicate.Duplicate == nil || duplicate.Duplicate.ExistingAssetID != first.Asset.ID {
		t.Fatalf("duplicate dto = %#v, want existing asset %s", duplicate.Duplicate, first.Asset.ID)
	}

	reused := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, Role: "character_ref", DuplicatePolicy: AssetDuplicatePolicyReuse, CorrelationID: "asset-reuse"})
	if !reused.OK || reused.Asset == nil || reused.Asset.ID != first.Asset.ID {
		t.Fatalf("reuse ImportAsset() = %#v, want existing asset", reused)
	}
	if len(reused.Assets) != 1 {
		t.Fatalf("reuse assets = %d, want 1", len(reused.Assets))
	}

	copied := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, Role: "character_ref", DuplicatePolicy: AssetDuplicatePolicyCopy, CorrelationID: "asset-copy"})
	if !copied.OK || copied.Asset == nil {
		t.Fatalf("copy ImportAsset() = %#v, want copied asset", copied)
	}
	if copied.Asset.ID == first.Asset.ID {
		t.Fatalf("copy asset id = %q, want a distinct id", copied.Asset.ID)
	}
	if len(copied.Assets) != 2 {
		t.Fatalf("copy assets = %d, want 2", len(copied.Assets))
	}
}

func TestAssetImportRejectsUnsafeRelativeSourcePath(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 8, 0, 0, time.UTC), "asset-path")
	root := createAssetTestProject(t)

	for _, sourcePath := range []string{"../outside.txt", "assets/../../outside.txt", "file:///tmp/private.txt", "https://example.invalid/private.txt"} {
		t.Run(sourcePath, func(t *testing.T) {
			result := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: sourcePath, CorrelationID: "asset-path"})
			if result.OK || result.Error == nil {
				t.Fatalf("ImportAsset() = %#v, want path error", result)
			}
			if result.Error.Code != CodeAssetPathInvalid {
				t.Fatalf("error code = %q, want %s", result.Error.Code, CodeAssetPathInvalid)
			}
		})
	}
}

func TestAssetImportRejectsUnsupportedMime(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 10, 0, 0, time.UTC), "asset-mime")
	root := createAssetTestProject(t)
	source := filepath.Join(t.TempDir(), "binary.bin")
	if err := os.WriteFile(source, []byte{0x00, 0x01, 0x02, 0x03}, 0o644); err != nil {
		t.Fatalf("write binary source: %v", err)
	}

	result := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, CorrelationID: "asset-mime"})
	if result.OK || result.Error == nil {
		t.Fatalf("ImportAsset() = %#v, want rejected mime", result)
	}
	if result.Error.Code != CodeAssetMimeRejected {
		t.Fatalf("error code = %q, want %s", result.Error.Code, CodeAssetMimeRejected)
	}
}

func TestAssetImportThumbnailFailureDoesNotBlockAssetReady(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 15, 0, 0, time.UTC), "asset-thumbnail")
	root := createAssetTestProject(t)
	if err := os.RemoveAll(filepath.Join(root, "assets", "thumbnails")); err != nil {
		t.Fatalf("remove thumbnails dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "thumbnails"), []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write thumbnails blocker: %v", err)
	}
	source := filepath.Join(t.TempDir(), "note.txt")
	if err := os.WriteFile(source, []byte("thumbnail failure still imports\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, CorrelationID: "asset-thumbnail"})
	if !result.OK || result.Asset == nil {
		t.Fatalf("ImportAsset() = %#v, want success despite thumbnail failure", result)
	}
	if result.Asset.ThumbnailStatus != AssetThumbnailFailed {
		t.Fatalf("thumbnail status = %q, want %s", result.Asset.ThumbnailStatus, AssetThumbnailFailed)
	}
	if !hasProjectEventError(result.Events, CodeAssetThumbnailFailed) {
		t.Fatalf("events = %#v, want thumbnail failure event", result.Events)
	}
	if result.Asset.Missing {
		t.Fatalf("asset missing = true, want asset ready")
	}
}

func TestAssetManagedReferenceCreatesHealthRisk(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 20, 0, 0, time.UTC), "asset-managed")
	root := createAssetTestProject(t)
	source := filepath.Join(t.TempDir(), "external-reference.txt")
	if err := os.WriteFile(source, []byte("external reference tracked explicitly\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	result := store.ImportAsset(ImportAssetCommand{
		Root:             root,
		SourcePath:       source,
		Role:             "scene_ref",
		ManagedReference: true,
		CorrelationID:    "asset-managed",
	})
	if !result.OK || result.Asset == nil {
		t.Fatalf("ImportAsset() = %#v, want managed reference asset", result)
	}
	if result.Asset.Source.Kind != AssetSourceManagedReference {
		t.Fatalf("source kind = %q, want %s", result.Asset.Source.Kind, AssetSourceManagedReference)
	}
	if filepath.IsAbs(result.Asset.RelativePath) || !strings.HasSuffix(result.Asset.RelativePath, ".managed-reference.json") {
		t.Fatalf("relativePath = %q, want project-relative managed reference placeholder", result.Asset.RelativePath)
	}
	placeholder, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(result.Asset.RelativePath)))
	if err != nil {
		t.Fatalf("read managed reference placeholder: %v", err)
	}
	if strings.Contains(string(placeholder), source) {
		t.Fatalf("managed reference placeholder leaked source path: %s", placeholder)
	}

	health := store.HealthReport(root)
	if !hasHealthCode(health, CodeManagedReferenceRisky) {
		t.Fatalf("health = %#v, want %s", health.Items, CodeManagedReferenceRisky)
	}
}

func TestAssetListReportsMissingIndexedAsset(t *testing.T) {
	store := testStore(time.Date(2026, 5, 20, 9, 25, 0, 0, time.UTC), "asset-list")
	root := createAssetTestProject(t)
	source := filepath.Join(t.TempDir(), "missing.txt")
	if err := os.WriteFile(source, []byte("will be removed\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	imported := store.ImportAsset(ImportAssetCommand{Root: root, SourcePath: source, CorrelationID: "asset-list-import"})
	if !imported.OK || imported.Asset == nil {
		t.Fatalf("ImportAsset() = %#v, want success", imported)
	}
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(imported.Asset.RelativePath))); err != nil {
		t.Fatalf("remove imported file: %v", err)
	}

	listed := store.ListAssets(ListAssetsCommand{Root: root, CorrelationID: "asset-list"})
	if !listed.OK || len(listed.Assets) != 1 {
		t.Fatalf("ListAssets() = %#v, want one asset", listed)
	}
	if !listed.Assets[0].Missing {
		t.Fatalf("listed asset = %#v, want missing=true", listed.Assets[0])
	}
	if listed.Health == nil || !hasHealthCode(*listed.Health, CodeAssetMissing) {
		t.Fatalf("health = %#v, want %s", listed.Health, CodeAssetMissing)
	}
}

func createAssetTestProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "asset-project")
	if err := CreateBaselineProject(root, testManifest(t)); err != nil {
		t.Fatalf("CreateBaselineProject() error = %v", err)
	}
	return root
}

func readAssetIndexForTest(t *testing.T, root string) AssetIndex {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(AssetIndexRelativePath)))
	if err != nil {
		t.Fatalf("read asset index: %v", err)
	}
	var index AssetIndex
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("decode asset index: %v", err)
	}
	return index
}

func hasProjectEventError(events []ProjectEvent, code string) bool {
	for _, event := range events {
		if event.Error != nil && event.Error.Code == code {
			return true
		}
	}
	return false
}
