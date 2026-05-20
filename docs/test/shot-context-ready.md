# Shot Context Ready Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-174` Shot required-field validation and `context_ready` transition
- Current conclusion: review rework, focused service/UI checks, browser smoke, and full `alpha-shell-verify` passed before independent re-review
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw command logs, full temporary paths, or machine-private runtime state are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| `go test ./internal/project` | Passed | Required fields, source lineage, reference resolution, `context_ready` persistence, graph node status, reopen projection, and `context_dirty` rollback were verified |
| `npm --prefix frontend run typecheck` | Passed | Shot context API, DTOs, and Inspector feedback controls compiled |
| `npm --prefix frontend run test:unit` | Passed | Shot context command builders and status badge mappings passed with existing Workbench tests |
| Independent review rework | Passed | Illegal higher-state demotion is blocked, Shot file id mismatch is rejected and canonicalized, missing graph nodes are blocked, unsafe reference paths cannot resolve, Canvas projects missing Shot references, and `context_dirty` is labeled distinctly |
| Browser smoke with mocked Wails bridge | Passed | Inspector Tasks showed blocking missing-field feedback, eligible/context_ready/context_dirty labels, ready promotion, and dirty rollback |
| `make alpha-shell-verify` | Passed | Harness check, Go tests, frontend typecheck/unit/build, Wails production build, alpha shell smoke, and whitespace check all passed after rework |

## Scope

This runbook covers `TOO-174` only: Shot required-field checks, basic local reference checks, candidate source lineage checks, `context_ready` promotion, and `context_dirty` rollback.

It does not generate PromptObject, export Package, call providers, submit/poll runtime jobs, import results, create Takes, or run Review flows.

## Commands

```bash
go test ./internal/project
go test ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
make alpha-shell-verify
```

## Expected Behavior

- Missing title, description, durationSeconds, aspectRatio, shotType, cameraMovement, action, emotion, or minimum context blocks `context_ready`.
- Candidate-created Shots must keep sourceCandidateID, scriptSceneID, and positive sourceRange lineage.
- characterIds, sceneProfileId or sceneId, propIds, and referenceAssetIds must resolve to local project objects or declared reference assets before `context_ready`.
- Valid Shots persist `context_ready` to both the Shot file and matching Canvas graph node status.
- Revised Shots can be marked `context_dirty` and the graph node status follows.
- Inspector feedback shows missing fields, blocking issues, reference status, and context state for selected Shot nodes.

## Known Limits

- Reference checks are local and structural; deep Asset Library digest/import/thumbnail checks belong to the Asset Continuity slice.
- Package stale propagation is not implemented in this slice.
- Prompt, Package, Runtime, Result, Take, and Review status transitions remain out of scope.
