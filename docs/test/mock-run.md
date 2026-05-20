# Deterministic Mock Run Smoke

## Scope

This runbook covers TOO-179 only: deterministic `mock_local` run execution, run record persistence, event JSONL, audit events, placeholder output, and Workbench Run queue projection.

It does not cover real model generation, internal providers, external Agent MCP, provider submit/poll, Result import, VideoResult, Take, Review status, billing, cloud sync, or account state.

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-179` deterministic mock run
- Current conclusion: passed local verification and independent review rework
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw command logs, full temporary paths, or machine-private runtime state are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| Focused Go mock run tests | Passed | `go test ./internal/project -run 'TestMockRun' -count=1`; covers start, retry, cancel, unsafe manifest path, output write failure, partial run-record recovery, and JSON round trip |
| Project and shell suites | Passed | `go test ./internal/project ./internal/shell` |
| Full Go suite | Passed | `go test ./...` |
| Frontend typecheck | Passed | `npm --prefix frontend run typecheck` |
| Frontend unit tests | Passed | `npm --prefix frontend run test:unit`; 12 files / 56 tests |
| Frontend build | Passed | `npm --prefix frontend run build`; known non-blocking chunk-size warning remains |
| Wails binding refresh and desktop build | Passed | `make desktop-shell-build`; generated `ProjectMockRunStart/Cancel/Retry` bindings refreshed |
| Browser smoke | Passed | Local Vite plus mocked Wails bridge verified Start, Retry, and Cancel UI paths with run id, `mock_local`, placeholder output, lifecycle events, and 100% progress visible |
| Full alpha-shell gate | Passed | `make alpha-shell-verify` |
| Harness review gate | Passed | `make harness-review-gate PLAN=.agents/plans/2026-05-20-too-179-mock-run.md`; `blocking_findings=none` |
| Diff hygiene | Passed | `git diff --check` |
| Independent review | Passed after rework | Review found no blocking findings and one medium partial-write recovery risk; rework added partial run directory attempt advancement and focused coverage |

## Expected Behavior

- Mock run mode is explicitly `mock_local`; it must not appear as `internal_provider` or `external_agent`.
- Starting a mock run creates:
  - `prompts/runs/{run_id}/run.json`
  - `prompts/runs/{run_id}/events.jsonl`
  - `assets/outputs/mock-run/{run_id}/placeholder-output.txt`
  - `audit/project-events.jsonl` entries for run lifecycle events
- Run record includes `runId`, `projectId`, `shotId` or `packageId`, `selectionIds`, `taskMode`, `providerMode`, `contextDigest`, `status`, `attempt`, `createdAt`, and `updatedAt`.
- Event JSONL includes deterministic queued, running, progress, and completed states for the start path.
- Cancel creates a cancelled attempt and does not write a placeholder output.
- Retry creates a new attempt and does not overwrite the old run record or old output.
- Placeholder output is stored under `assets/outputs/mock-run/`, not `assets/results/`, so TOO-180 can own Result import separately.
- Workbench shows mock run state in Queue, Runs tab, and Bottom Bar using the same RuntimeEventDTO-style event projection.

## Commands

```bash
go test ./internal/project -run 'TestMockRun' -count=1
go test ./internal/project ./internal/shell
go test ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
npm --prefix frontend run build
make desktop-shell-build
make alpha-shell-verify
make harness-review-gate PLAN=.agents/plans/2026-05-20-too-179-mock-run.md
git diff --check
```

## Browser Smoke

Use a local Vite target with a mocked Wails bridge or the built Wails shell after generated bindings refresh.

Expected visible behavior:

- The Top Bar has a Mock action that is disabled until a shot/package/selection exists.
- Inspector Tasks shows Start mock, Cancel, and Retry controls.
- Starting a mock run switches the Inspector to Runs and displays a completed `mock_local` run.
- Runs tab shows run status, attempt, run record path, output digest, and event timeline.
- Cancel shows cancelled state without output metadata.
- Retry shows a new attempt and a new output path.
- Bottom Bar progress reaches the latest runtime event progress.

## Known Limits

- The mock runner is synchronous and deterministic for Alpha shell validation; it is not a background provider runner.
- Placeholder output is a text artifact for traceability and later result-import testing; it is not treated as an imported result in this card.
- Browser smoke may use mocked Wails bindings for UI projection; real service behavior is covered by Go tests and generated Wails binding refresh.
