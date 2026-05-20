# ScriptScene And Shot Candidate Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-173` ScriptScene and ShotCard candidate confirmation
- Current conclusion: focused Go checks, frontend API checks, browser smoke, and the alpha-shell gate passed before independent review
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw command logs, full temporary paths, or machine-private runtime state are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| `go test ./internal/project` | Passed | ScriptScene required fields, source range bounds, overlap handling, candidate save/reject/retry, accepted lifecycle guards, duplicate accepted output blocking, formal draft Shot confirm, source lineage, and missing-source blocking were verified |
| `go test ./...` | Passed | Project and shell package checks passed |
| `npm --prefix frontend run typecheck` | Passed | ScriptScene/candidate UI and API wrapper types compiled |
| `npm --prefix frontend run test:unit` | Passed | ScriptScene/candidate command-builder and failed-action row-preservation tests passed with existing Workbench tests |
| Browser smoke | Passed | Mocked Wails bindings drove Script tab scene confirmation, candidate save, Canvas expansion-table projection, and per-row candidate confirmation to `accepted` without a shared Shot id field |
| `make alpha-shell-verify` | Passed | Harness verification, Go tests, frontend typecheck/unit/build, desktop shell build, alpha-shell smoke, and whitespace checks passed |

## Scope

This runbook covers `TOO-173` only: manual ScriptScene confirmation from ScriptDocument line ranges, Shot candidate row storage, candidate reject/retry, and candidate confirmation into a formal draft ShotCard.

It does not validate full Shot required fields, move Shots to `context_ready`, bind Asset Continuity, generate storyboards, call providers, optimize prompts, export packages, or run review flows.

## Commands

```bash
go test ./internal/project
go test ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
make alpha-shell-verify
```

## Expected Behavior

- A valid line range plus scene title, location, and action writes a ScriptScene into `ScriptDocument.scenes`.
- Missing title, location, action, or invalid line ranges block scene confirmation with explicit errors.
- Overlapping scene ranges block by default and can be explicitly allowed.
- Shot candidates are stored separately from formal ShotCards and start as `candidate`.
- Rejected candidates do not create Shot files and can be edited back to `candidate` for retry.
- Accepted candidates cannot be rejected or confirmed again through the service API.
- Confirming a valid candidate writes a formal draft ShotCard with candidate id, ScriptScene id, source range, confirmer, confirmation time, and overwritten field metadata.
- Candidate confirmation blocks duplicate accepted output for the same ScriptScene/index or existing Shot id.
- Candidate confirmation does not set `context_ready`.
- Workbench has baseline scene/candidate controls and a Canvas script expansion table projection.
- Failed candidate actions preserve existing Canvas expansion rows until a successful load or candidate mutation replaces them.

## Storage

- ScriptScene truth: `assets/inputs/scripts/{scriptId}.script.json`
- Candidate truth: `assets/inputs/shot-candidates/{scriptId}.candidates.json`
- Confirmed draft Shot truth: `shots/{shotId}.json`

## Known Limits

- Candidate generation is manual in this slice; deterministic or AI-assisted generation is not included.
- Browser smoke uses mocked Wails bindings for UI projection; Go tests cover the real service behavior.
- Full Shot validation and `context_ready` belong to `TOO-174`.
