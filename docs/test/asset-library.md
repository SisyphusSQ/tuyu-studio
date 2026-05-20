# Asset Library Import and Indexing

## Scope

This note records the TOO-175 verification surface for Asset import, digest indexing, duplicate handling, thumbnail placeholders, managed-reference health risk, and the Workbench Asset Library list.

## Current Run Summary

| Check | Result | Evidence |
| --- | --- | --- |
| Go focused Asset tests | Passed | `go test ./internal/project -run 'TestAsset' -count=1` |
| Go full suite | Passed | `go test ./...` |
| Frontend unit tests | Passed | `npm --prefix frontend run test:unit` |
| Frontend typecheck | Passed | `npm --prefix frontend run typecheck` |
| Frontend build | Passed | `npm --prefix frontend run build` |
| Full alpha-shell gate | Passed | `make alpha-shell-verify` |
| Browser smoke | Passed | Playwright CLI opened Vite on `127.0.0.1:5191`, injected a mocked Wails bridge, loaded the Asset Library, clicked List, exercised failed import row preservation, imported a managed-reference asset, and verified asset Runs/Audit visibility. |
| Diff hygiene | Passed | `git diff --check` |
| Independent review | Passed | Initial subagent review found no blocking/high findings and one medium digest/copy race; rework added destination digest/size verification before index write. Focused re-review reported no blocking/high/medium findings. |

## Coverage Notes

- Import accepts project-relative or absolute source input but persists only project-relative Asset `relativePath` and thumbnail paths.
- Content mime is detected from file bytes and only allows image, video, audio, or UTF-8 text.
- Focused fixtures cover content-based PNG, MP4 and WAV detection even when the source extension is misleading.
- Health digest checks stream asset files instead of loading full media files into memory.
- Imported file digest is rechecked from the copied project file before `assets/index.json` is written, so a source mutation between pre-copy digest and copy fails with `asset_digest_failed` instead of persisting an inconsistent index.
- Unsafe project-relative source paths and URL-style source paths are rejected before read/copy.
- Duplicate digest and mime defaults to `asset_duplicate_found`; explicit `reuse` returns the existing Asset; explicit `copy` creates a distinct Asset id and file path.
- Thumbnail placeholder failure records `thumbnail_failed` without blocking the Asset record.
- Managed reference creates an explicit project-relative placeholder and health warning `managed_reference_risky`; external source paths are not persisted in the placeholder.
- Failed imports preserve the current visible Asset Library rows in the Workbench UI.
- Asset events and asset health items are included in the Inspector Runs/Audit surfaces.
- Binding, main-reference, Continuity lock, missing-asset Shot dirty propagation, media understanding, and transcoding remain out of TOO-175 scope.
