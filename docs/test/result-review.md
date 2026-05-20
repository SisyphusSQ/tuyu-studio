# Result Review Verification

## Scope

TOO-180 covers local placeholder result import, VideoResult/Take records, Shot/Package binding, review state transitions, traceability, wrong-binding recovery, duplicate digest handling, and missing-file reporting.

It does not cover real provider download, submit/poll, automatic scoring, advanced playback, final render, billing, cloud sync, or Alpha shell E2E closeout.

## Verification Matrix

| Check | Status | Evidence |
| --- | --- | --- |
| Result import | Passed | Focused Go tests copy a local placeholder file into `assets/results/{shot_id}/`, create an Asset record, create a VideoResult record, and link the Shot `resultIds`. |
| Take numbering and duplicate digest | Passed | Focused Go tests reject duplicate digest by default, allow `new_take`, and allocate stable per-Shot take numbers. |
| Target consistency | Passed | Focused Go tests reject imports and rebinds where the explicit Shot does not match the selected Package's owning Shot. |
| Review state guard | Passed | Focused Go tests block `needs_revision` / `rejected` without a reason, block review decisions while a result is still `binding_pending`, and reset review state when binding changes. |
| Trace | Passed | Focused Go tests trace Result -> Asset -> Shot -> Package -> mock run and mark the package `result_received` after import. |
| Wrong binding recovery | Passed | Focused Go tests rebind a result from the wrong Shot to the correct Shot, keep the imported file path, and preserve take history. |
| Missing file | Passed | Focused Go tests remove the copied file and verify `missing_file` projection plus health warning without deleting the record. |
| Frontend API/UI | Passed | Unit tests cover command builders for import/list/trace/review/rebind; `vue-tsc` accepts the DTOs and Wails wrapper; Tasks UI exposes Trace and Shot/Package unbind recovery controls. |
| Wails binding refresh | Passed | `make desktop-shell-build` regenerated `ProjectResultImport`, `ProjectResultsList`, `ProjectResultTrace`, `ProjectResultReviewUpdate`, and `ProjectResultRebind`. |
| Browser smoke | Passed | Headless Chrome against Vite with mocked Wails bridge covered Import result -> Trace -> Save review -> unbind Package -> Rebind. |

## Commands

```bash
go test ./internal/project -run 'TestResultImport|TestTakeNumbering|TestReviewState|TestResultTarget|TestUnboundResult|TestResultTrace|TestWrongBindingRecovery|TestMissingResultFile' -count=1
go test ./internal/project ./internal/shell
npm --prefix frontend run test:unit
npm --prefix frontend run typecheck
make desktop-shell-build
make alpha-shell-verify
```

## Recovery Notes

- A duplicate digest returns `result_duplicate_digest`; choose `reuse` or `new_take`.
- An unbound result remains `binding_pending` and is not treated as formally reviewable.
- A Shot/Package mismatch returns `result_target_mismatch`; choose a package exported from the same Shot or clear the conflicting target.
- A missing copied result file keeps the VideoResult and Asset records, reports `missing_file`, and surfaces `asset_missing` health evidence.
- Rebinding requires a reason and does not move or delete the imported result file.
