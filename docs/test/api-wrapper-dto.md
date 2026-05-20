# API Wrapper And DTO Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-162` Go service, API wrapper, Error/Event DTO
- Current conclusion: passed before independent review
- Automated entries: Go tests, frontend typecheck, frontend unit tests, frontend build, harness verify, Wails desktop build
- Smoke entries: Workbench success probe and structured-error probe through API wrapper UI projection
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw logs, full temporary paths, or full local-only screenshots are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| `npm --prefix frontend install` | Passed | Vitest dependency installed with no vulnerability report |
| `npm --prefix frontend run typecheck` | Passed | Vue TypeScript check completed |
| `npm --prefix frontend run test:unit` | Passed | DTO normalization focused tests passed |
| `npm --prefix frontend run build` | Passed | Vite production build completed with non-blocking chunk-size warning |
| `go test ./...` | Passed | root package had no tests; `internal/shell` passed focused service/DTO tests |
| `make harness-verify` | Passed | harness check passed |
| `make desktop-shell-build` | Passed | Wails v2 darwin/arm64 package completed and generated bindings refreshed |
| `git diff --check` | Passed | no whitespace errors |
| Binding boundary | Passed | `frontend/src/api/workbench.ts` is the only `frontend/src` file importing generated Wails binding |
| UI smoke | Passed | success probe, structured error summary, recovery actions, event projection summary, and bottom event count displayed through the Workbench UI |

## Scope

This runbook covers `TOO-162` only: a minimal Workbench-to-Go-service call path through `frontend/src/api/*`, plus baseline `AppErrorDTO` and `RuntimeEventDTO` rendering.

## Commands

```bash
npm --prefix frontend install
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
npm --prefix frontend run build
go test ./...
make harness-verify
make desktop-shell-build
```

## UI Smoke

Open the Workbench in the built Wails app or Vite web target. For Web smoke, inject a Wails binding mock before app load; real Go service DTO behavior is covered by Go tests and generated binding refresh.

Expected visible state:

- Trigger `Probe Go service`.
- Success summary displays a structured service name, status, capabilities, and event summary.
- Trigger `Show structured error`.
- Error summary displays `AppErrorDTO` code, severity, correlation id, and recovery actions; technical detail remains in DTO state for debugging/writeback.
- Event projection displays at least one `RuntimeEventDTO` entry with event type, state, progress, summary, and createdAt; event id and next actions remain in DTO state for later detailed event panels.

Known limits:

- This runbook does not verify ProjectStore, RuntimeGateway, Canvas command, provider adapter, or real generation.
- Runtime events are baseline DTO projections for this issue, not the full Run/Event/Audit model.
