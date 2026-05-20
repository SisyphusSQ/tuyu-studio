# GenerationPackage Export Smoke

## Scope

This runbook covers TOO-178: exporting a local manual handoff `GenerationPackage` for a `context_ready` Shot, including package manifest, prompt, script excerpt, continuity snapshot, upload checklist, copied references, package-relative paths, re-export versioning, and Workbench feedback.

It does not cover provider submit or poll, external Agent MCP, deterministic mock generation, result import, Take, Review, billing, cloud sync, final render, or real model execution.

## Current Run Summary

| Check | Result | Evidence |
| --- | --- | --- |
| Focused package tests | Passed | `go test ./internal/project -run 'TestGenerationPackage' -count=1` |
| Graph/package focused tests | Passed | `go test ./internal/project -run 'TestGraphView|TestGenerationPackage' -count=1`; package coverage includes character-level references, missing Shot graph node blockers, stale package propagation, re-export, and redaction |
| Project and shell tests | Passed | `go test ./internal/project ./internal/shell` |
| Full Go suite | Passed | `go test ./...` |
| Frontend typecheck | Passed | `npm --prefix frontend run typecheck` |
| Frontend unit tests | Passed | `npm --prefix frontend run test:unit` |
| Frontend build | Passed | `npm --prefix frontend run build` completed with the known non-blocking chunk-size warning |
| Wails binding refresh | Passed | `make desktop-shell-build` regenerated `ProjectGenerationPackageExport` and `GenerationPackage` bindings |
| Browser smoke | Passed | Local Vite target with an injected Wails bridge mock rendered Tasks, exported `shot_002`, displayed ready package status, manifest path, and package-local references |
| Full alpha-shell gate | Passed | `make alpha-shell-verify` passed harness, Go, frontend typecheck/unit/build, Wails smoke, and worktree whitespace checks |
| Diff hygiene | Passed | `git diff --check` |

## Expected Behavior

- Export requires an existing Shot with status `context_ready`.
- Export gathers Shot data, script excerpt, continuity rules, asset index references, and profile main/reference assets.
- Export includes direct `characterRefs[].referenceAssetId` assets, even when they are not duplicated in `referenceAssetIds`.
- Export blocks before writing if the Shot's real graph node is missing.
- Export writes a new package directory under `packages/scene_{scene_index}/shot_{shot_index}_pkg_{timestamp}/`.
- Package contents include `manifest.json`, `prompt.txt`, `script_excerpt.md`, `continuity.md`, `upload_checklist.md`, and copied files under `references/`.
- Manifest paths are package-relative and all referenced package files exist.
- Missing reference assets block export with `package_reference_missing` and recovery actions.
- Re-export creates a new directory and increments `packageVersion`; it does not overwrite prior ready packages.
- Export links the new package id to the Shot and adds a package node / handoff edge to the project graph.
- Linked ready packages become `stale` when the Shot is marked dirty directly or through continuity dirty propagation.
- Workbench shows package status, package id/version, manifest path, digest summary, checklist path, and package-local references.

## Commands

```bash
go test ./internal/project -run 'TestGenerationPackage' -count=1
go test ./internal/project -run 'TestGraphView|TestGenerationPackage' -count=1
go test ./internal/project ./internal/shell
go test ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
npm --prefix frontend run build
make desktop-shell-build
```

## Browser Smoke

Use a local Vite target with a pre-load Wails bridge mock. The mock must expose:

- `ProjectGenerationPackageExport`
- `ProjectGraphView`
- `ProjectAssetsList`
- `ProjectAssetBindingsList`
- `ProjectContinuityList`

Expected visible behavior:

- Inspector Tasks tab includes a `Generation package` panel.
- `Export package` sends `shot_002` through the API wrapper.
- Package result shows `ready`, package id/version, manifest path, prompt/checklist paths, and three package-local references.
- Runs/Audit surfaces include the `generation_package.exported` event stream and any package export health items.

## Safety And Redaction

- Committed docs do not include raw logs, real credentials, tokens, cookies, database hosts, connection strings, row keys, full temporary paths, or machine-private browser state.
- Package content scan rejects private paths, external URL markers, and token/credential-like text in generated text files, manifest JSON, and text-like references copied into the package.
- The design prototype repo and `.pen` file are read-only references for this issue.

## Known Limits

- Browser smoke uses a mocked Wails bridge; real service behavior is covered by Go tests and generated Wails bindings.
- Browser screenshot capture timed out in the in-app browser during this run; DOM/state assertions passed.
- Full downstream provider submit, poll, result import and Take/Review remain for later issues.
