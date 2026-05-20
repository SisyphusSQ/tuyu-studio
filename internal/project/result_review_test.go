package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResultImportCreatesAssetVideoResultAndReviewPending(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "placeholder-result.txt", "mock result frame\n")
	store := testStore(time.Date(2026, 5, 20, 15, 0, 0, 0, time.UTC), "result-import")

	result := store.ImportResult(ImportResultCommand{
		Root:          root,
		SourcePath:    source,
		ShotID:        "shot_002",
		CreatedBy:     "test",
		CorrelationID: "result-import",
	})
	if !result.OK || result.Result == nil {
		t.Fatalf("ImportResult() = %#v, want result", result)
	}
	imported := result.Result
	if imported.ProjectID != "proj_alpha_fixture" || imported.ShotID != "shot_002" {
		t.Fatalf("imported target = %s/%s, want fixture shot_002", imported.ProjectID, imported.ShotID)
	}
	if imported.TakeNumber != 1 {
		t.Fatalf("takeNumber = %d, want 1", imported.TakeNumber)
	}
	if imported.Status != ResultStatusReviewPending || imported.ReviewStatus != ResultReviewPending {
		t.Fatalf("status/review = %s/%s, want review_pending/pending", imported.Status, imported.ReviewStatus)
	}
	if !strings.HasPrefix(imported.RelativePath, "assets/results/shot_002/") {
		t.Fatalf("relativePath = %q, want shot result directory", imported.RelativePath)
	}
	if imported.AssetID == "" || imported.RecordPath == "" {
		t.Fatalf("result lacks asset/record linkage: %#v", imported)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(imported.RelativePath))); err != nil {
		t.Fatalf("copied result missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(imported.RecordPath))); err != nil {
		t.Fatalf("result record missing: %v", err)
	}

	assetIndex, err := store.loadAssetIndex(root, "proj_alpha_fixture")
	if err != nil {
		t.Fatalf("load asset index: %v", err)
	}
	if !assetIDExists(assetIndex.Assets, imported.AssetID) {
		t.Fatalf("asset index missing result asset %s", imported.AssetID)
	}
	shot := readShotJSON(t, root, "shot_002")
	if !containsString(shot.ResultIDs, imported.ID) {
		t.Fatalf("shot resultIds = %#v, want %s", shot.ResultIDs, imported.ID)
	}
}

func TestTakeNumberingAndDuplicateDigestPolicy(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "same-result.txt", "same result digest\n")
	store := testStore(time.Date(2026, 5, 20, 15, 2, 0, 0, time.UTC), "result-duplicate")

	first := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, ShotID: "shot_002", CorrelationID: "result-first"})
	if !first.OK || first.Result == nil {
		t.Fatalf("first ImportResult() = %#v, want ok", first)
	}
	duplicate := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, ShotID: "shot_002", CorrelationID: "result-duplicate"})
	if duplicate.OK || duplicate.Error == nil {
		t.Fatalf("duplicate ImportResult() = %#v, want duplicate warning", duplicate)
	}
	if duplicate.Error.Code != CodeResultDuplicateDigest {
		t.Fatalf("duplicate code = %q, want %s", duplicate.Error.Code, CodeResultDuplicateDigest)
	}
	if duplicate.Duplicate == nil || duplicate.Duplicate.ExistingResultID != first.Result.ID {
		t.Fatalf("duplicate dto = %#v, want first result %s", duplicate.Duplicate, first.Result.ID)
	}

	second := store.ImportResult(ImportResultCommand{
		Root:            root,
		SourcePath:      source,
		ShotID:          "shot_002",
		DuplicatePolicy: ResultDuplicatePolicyNewTake,
		CorrelationID:   "result-new-take",
	})
	if !second.OK || second.Result == nil {
		t.Fatalf("new take ImportResult() = %#v, want ok", second)
	}
	if second.Result.TakeNumber != first.Result.TakeNumber+1 {
		t.Fatalf("takeNumber = %d, want %d", second.Result.TakeNumber, first.Result.TakeNumber+1)
	}
	if second.Result.ID == first.Result.ID || second.Result.AssetID == first.Result.AssetID {
		t.Fatalf("new take reused id/asset: first=%#v second=%#v", first.Result, second.Result)
	}
}

func TestReviewStateRequiresReason(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "review-result.txt", "review result\n")
	store := testStore(time.Date(2026, 5, 20, 15, 4, 0, 0, time.UTC), "result-review")
	imported := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, ShotID: "shot_002", CorrelationID: "result-review-import"})
	if !imported.OK || imported.Result == nil {
		t.Fatalf("ImportResult() = %#v, want ok", imported)
	}

	rejected := store.UpdateResultReview(UpdateResultReviewCommand{
		Root:          root,
		ResultID:      imported.Result.ID,
		ReviewStatus:  ResultReviewRejected,
		CorrelationID: "result-review-reject",
	})
	if rejected.OK || rejected.Error == nil {
		t.Fatalf("UpdateResultReview(rejected without reason) = %#v, want error", rejected)
	}
	if rejected.Error.Code != CodeResultReviewNoteMissing {
		t.Fatalf("review code = %q, want %s", rejected.Error.Code, CodeResultReviewNoteMissing)
	}

	revision := store.UpdateResultReview(UpdateResultReviewCommand{
		Root:          root,
		ResultID:      imported.Result.ID,
		ReviewStatus:  ResultReviewNeedsRevision,
		Reason:        "lantern continuity mismatch",
		ReviewNotes:   "redo with the lantern fixed",
		CorrelationID: "result-review-revision",
	})
	if !revision.OK || revision.Result == nil {
		t.Fatalf("UpdateResultReview(needs_revision) = %#v, want ok", revision)
	}
	if revision.Result.Status != ResultStatusNeedsRevision || revision.Result.ReviewReason == "" {
		t.Fatalf("revision result = %#v, want needs_revision with reason", revision.Result)
	}
}

func TestResultTargetRejectsShotPackageMismatch(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "mismatch-result.txt", "mismatch result\n")
	store := testStore(time.Date(2026, 5, 20, 15, 5, 0, 0, time.UTC), "result-mismatch")

	imported := store.ImportResult(ImportResultCommand{
		Root:          root,
		SourcePath:    source,
		ShotID:        "shot_002",
		PackageID:     "pkg_scene001_shot001",
		CorrelationID: "result-mismatch-import",
	})
	if imported.OK || imported.Error == nil {
		t.Fatalf("ImportResult(mismatched target) = %#v, want error", imported)
	}
	if imported.Error.Code != CodeResultTargetMismatch {
		t.Fatalf("mismatch code = %q, want %s", imported.Error.Code, CodeResultTargetMismatch)
	}

	valid := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, ShotID: "shot_002", CorrelationID: "result-valid"})
	if !valid.OK || valid.Result == nil {
		t.Fatalf("ImportResult(valid) = %#v, want ok", valid)
	}
	rebound := store.RebindResult(RebindResultCommand{
		Root:          root,
		ResultID:      valid.Result.ID,
		PackageID:     "pkg_scene001_shot001",
		Reason:        "operator selected the wrong package",
		CorrelationID: "result-mismatch-rebind",
	})
	if rebound.OK || rebound.Error == nil {
		t.Fatalf("RebindResult(mismatched target) = %#v, want error", rebound)
	}
	if rebound.Error.Code != CodeResultTargetMismatch {
		t.Fatalf("rebind mismatch code = %q, want %s", rebound.Error.Code, CodeResultTargetMismatch)
	}
}

func TestUnboundResultCannotBeApprovedBeforeBinding(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "unbound-result.txt", "unbound result\n")
	store := testStore(time.Date(2026, 5, 20, 15, 5, 30, 0, time.UTC), "result-unbound")

	imported := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, CorrelationID: "result-unbound-import"})
	if !imported.OK || imported.Result == nil {
		t.Fatalf("ImportResult(unbound) = %#v, want ok", imported)
	}
	if imported.Result.Status != ResultStatusBindingPending || imported.Result.TakeNumber != 0 {
		t.Fatalf("unbound result = %#v, want binding_pending take 0", imported.Result)
	}

	approved := store.UpdateResultReview(UpdateResultReviewCommand{
		Root:          root,
		ResultID:      imported.Result.ID,
		ReviewStatus:  ResultReviewApproved,
		CorrelationID: "result-unbound-approve",
	})
	if approved.OK || approved.Error == nil {
		t.Fatalf("UpdateResultReview(unbound approve) = %#v, want error", approved)
	}
	if approved.Error.Code != CodeResultTargetMissing {
		t.Fatalf("unbound approve code = %q, want %s", approved.Error.Code, CodeResultTargetMissing)
	}

	rebound := store.RebindResult(RebindResultCommand{
		Root:          root,
		ResultID:      imported.Result.ID,
		ShotID:        "shot_002",
		Reason:        "bind pending result to reviewed shot",
		CorrelationID: "result-unbound-rebind",
	})
	if !rebound.OK || rebound.Result == nil {
		t.Fatalf("RebindResult(unbound to shot) = %#v, want ok", rebound)
	}
	if rebound.Result.Status != ResultStatusReviewPending || rebound.Result.ReviewStatus != ResultReviewPending || rebound.Result.ReviewReason != "" {
		t.Fatalf("rebound review state = %#v, want reset pending review", rebound.Result)
	}
	if rebound.Result.ShotID != "shot_002" || rebound.Result.TakeNumber != 1 || len(rebound.Result.TakeHistory) != 1 {
		t.Fatalf("rebound target/history = %#v, want shot_002 take history", rebound.Result)
	}
}

func TestResultTraceIncludesShotAssetRunAndPackage(t *testing.T) {
	root := copyExampleFixture(t)
	store := testStore(time.Date(2026, 5, 20, 15, 6, 0, 0, time.UTC), "result-trace")
	pkg := store.ExportGenerationPackage(ExportGenerationPackageCommand{Root: root, ShotID: "shot_002", CorrelationID: "trace-package"})
	if !pkg.OK || pkg.Package == nil {
		t.Fatalf("ExportGenerationPackage() = %#v, want package", pkg)
	}
	run := store.StartMockRun(MockRunCommand{
		Root:          root,
		PackageID:     pkg.Package.PackageID,
		TaskMode:      MockRunProviderMode,
		CorrelationID: "trace-run",
	})
	if !run.OK || run.Run == nil || run.Run.Output == nil {
		t.Fatalf("StartMockRun() = %#v, want output", run)
	}
	imported := store.ImportResult(ImportResultCommand{
		Root:          root,
		RunID:         run.Run.RunID,
		CorrelationID: "trace-import",
	})
	if !imported.OK || imported.Result == nil {
		t.Fatalf("ImportResult(from run) = %#v, want ok", imported)
	}

	trace := store.TraceResult(TraceResultCommand{Root: root, ResultID: imported.Result.ID, CorrelationID: "trace-result"})
	if !trace.OK || trace.Trace == nil {
		t.Fatalf("TraceResult() = %#v, want trace", trace)
	}
	if trace.Trace.Asset == nil || trace.Trace.Shot == nil || trace.Trace.Package == nil || trace.Trace.Run == nil {
		t.Fatalf("trace missing links: %#v", trace.Trace)
	}
	if trace.Trace.Package.GenerationPackageStatus != GenerationPackageStatusResultReceived {
		t.Fatalf("package status = %q, want result_received", trace.Trace.Package.GenerationPackageStatus)
	}
	if trace.Trace.Run.RunID != run.Run.RunID || trace.Trace.Shot.ID != "shot_002" {
		t.Fatalf("trace run/shot = %#v/%#v, want imported run and shot_002", trace.Trace.Run, trace.Trace.Shot)
	}
}

func TestWrongBindingRecoveryPreservesFileAndTakeHistory(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "wrong-binding.txt", "wrong binding result\n")
	store := testStore(time.Date(2026, 5, 20, 15, 8, 0, 0, time.UTC), "result-rebind")
	imported := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, ShotID: "shot_001", CorrelationID: "result-wrong-import"})
	if !imported.OK || imported.Result == nil {
		t.Fatalf("ImportResult() = %#v, want ok", imported)
	}
	originalPath := imported.Result.RelativePath

	rebound := store.RebindResult(RebindResultCommand{
		Root:          root,
		ResultID:      imported.Result.ID,
		ShotID:        "shot_002",
		Reason:        "operator selected the wrong shot",
		CorrelationID: "result-rebind",
	})
	if !rebound.OK || rebound.Result == nil {
		t.Fatalf("RebindResult() = %#v, want ok", rebound)
	}
	if rebound.Result.ID != imported.Result.ID || rebound.Result.RelativePath != originalPath {
		t.Fatalf("rebind changed identity/path: before=%#v after=%#v", imported.Result, rebound.Result)
	}
	if rebound.Result.ShotID != "shot_002" || rebound.Result.TakeNumber != 1 || len(rebound.Result.TakeHistory) < 2 {
		t.Fatalf("rebound result = %#v, want shot_002 take history", rebound.Result)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(originalPath))); err != nil {
		t.Fatalf("original imported file was moved or deleted: %v", err)
	}
	oldShot := readShotJSON(t, root, "shot_001")
	newShot := readShotJSON(t, root, "shot_002")
	if containsString(oldShot.ResultIDs, imported.Result.ID) {
		t.Fatalf("old shot still has result id: %#v", oldShot.ResultIDs)
	}
	if !containsString(newShot.ResultIDs, imported.Result.ID) {
		t.Fatalf("new shot missing result id: %#v", newShot.ResultIDs)
	}
}

func TestMissingResultFile(t *testing.T) {
	root := copyExampleFixture(t)
	source := writeResultSource(t, "missing-result.txt", "missing result\n")
	store := testStore(time.Date(2026, 5, 20, 15, 10, 0, 0, time.UTC), "result-missing")
	imported := store.ImportResult(ImportResultCommand{Root: root, SourcePath: source, ShotID: "shot_002", CorrelationID: "result-missing-import"})
	if !imported.OK || imported.Result == nil {
		t.Fatalf("ImportResult() = %#v, want ok", imported)
	}
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(imported.Result.RelativePath))); err != nil {
		t.Fatalf("remove imported file: %v", err)
	}

	listed := store.ListResults(ListResultsCommand{Root: root, ResultID: imported.Result.ID, CorrelationID: "result-missing-list"})
	if !listed.OK || listed.Result == nil {
		t.Fatalf("ListResults() = %#v, want result", listed)
	}
	if !listed.Result.Missing || listed.Result.Status != ResultStatusMissingFile {
		t.Fatalf("listed result = %#v, want missing_file", listed.Result)
	}
	if listed.Health == nil || !hasHealthCode(*listed.Health, CodeAssetMissing) {
		t.Fatalf("health = %#v, want asset_missing", listed.Health)
	}
}

func writeResultSource(t *testing.T, name string, content string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	return filename
}
