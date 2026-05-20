# Alpha Shell Startup And Build Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-163` startup smoke, build entry, and baseline verification guidance
- Current conclusion: passed with documented desktop-window automation fallback
- Sensitive information handling: do not record credentials, tokens, cookies, database hosts, connection strings, row keys, raw logs, full temporary paths, full local-only screenshots, or machine-private runtime state.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| `make harness-verify` | Passed | Base harness contract check completed |
| `go test ./...` | Passed | Go package and service tests completed |
| `npm --prefix frontend run typecheck` | Passed | Vue TypeScript check completed |
| `npm --prefix frontend run test:unit` | Passed | DTO and wrapper unit tests passed |
| `npm --prefix frontend run build` | Passed | Vite production build completed with the known non-blocking chunk-size warning |
| `make desktop-shell-build` | Passed | Wails build completed and generated bindings were refreshed |
| `make alpha-shell-smoke` | Passed | Build artifact, generated binding, and UI boundary checks passed |
| `make worktree-whitespace-check` | Passed | No whitespace errors in tracked, staged, or untracked text files |
| Desktop launch smoke | Passed | App bundle launched and the `tuyu-studio` process started; process was stopped after verification |
| Computer Use desktop window readback | Limited | Process existed, but the desktop automation window tree returned `cgWindowNotFound` |
| Workbench Web smoke fallback | Passed | Workbench zones were present, success probe and structured-error path rendered through the UI, and no horizontal overflow was detected at 1280x820 or 1024x700 |

## Scope

This runbook covers the Alpha shell baseline across `TOO-160` through `TOO-163`.

It verifies that the current desktop shell can be built, started, and used to reach the Workbench first screen with a minimal Go service success/error path. It does not verify Local Project, Example Project, Canvas graph editing, Script-to-Shot, Asset Library, Review, provider execution, installers, signing, auto-update, or a cross-platform packaging matrix.

## Commands

Run the full reusable local gate:

```bash
make alpha-shell-verify
```

The aggregate target runs:

```bash
make harness-verify
go test ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
npm --prefix frontend run build
make alpha-shell-smoke
make worktree-whitespace-check
```

`make alpha-shell-smoke` builds the Wails app, refreshes generated bindings, then runs `scripts/harness/alpha_shell_smoke.sh --check-build`.

`make alpha-shell-verify` finishes with `make worktree-whitespace-check`, which checks unstaged tracked changes, staged changes, and untracked text files not ignored by Git.

Use the optional launch smoke on macOS when a GUI session is available:

```bash
make alpha-shell-smoke-launch
```

By default the launch smoke captures existing `tuyu-studio` PIDs, starts a new app instance, confirms a new PID appears, and stops only that new PID. Set `TUYU_SMOKE_KEEP_APP=1` when you want to keep the new window open for manual inspection.

## Expected Build Result

- `make harness-verify` passes.
- Go tests pass.
- Frontend typecheck, unit tests, and production build pass.
- Wails build produces the local desktop bundle under ignored build output.
- Generated bindings expose `AppInfo`, `ShellHealth`, and `WorkbenchProbe`.
- Generated models include `AppError`, `RuntimeEvent`, and `WorkbenchProbeResult`.
- `frontend/src/api/*` is the only frontend source layer allowed to import `wailsjs/go/...` or `wailsjs/runtime/...`.
- `frontend/src` keeps Workbench lists on Ant Design Vue list primitives rather than raw `<ul>` / `<li>` markup.

## Desktop Smoke

After a successful build, open the local app bundle or run:

```bash
make alpha-shell-smoke-launch
```

Expected visible state:

- The app starts with title `Tuyu Studio`.
- The first screen is the Workbench, not a landing page.
- Top Bar, Left Panel, Canvas placeholder, Inspector, and Bottom Bar are visible.
- The Workbench exposes the `Probe Go service` and `Show structured error` actions.
- `Probe Go service` displays a structured service/status/capability summary and at least one completed event projection.
- `Show structured error` displays an `AppErrorDTO` code, severity, correlation id, recovery actions, and at least one blocked event projection.

## Web Smoke Fallback

When desktop window automation is unavailable, run a Vite web target and inject a Wails binding mock before app load. This verifies Workbench rendering and API wrapper UI projection. Real Go service DTO behavior is still proven by Go tests and the generated Wails binding refresh.

Expected Web smoke result:

- Workbench zones are visible at desktop-sized and narrower viewports.
- There is no page-level horizontal overflow.
- Success probe, structured error summary, recovery actions, event projection summary, and bottom event count render through the same Workbench UI surface.

## Common Failures And Recovery

| Failure | Recovery |
| --- | --- |
| Missing frontend dependencies | Run `npm --prefix frontend install`, then rerun `make alpha-shell-verify`. |
| Wails CLI download or build failure | Confirm Go module/network access for `go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`, then rerun `make desktop-shell-build`. |
| Generated binding drift | Run `make desktop-shell-build` and commit generated `frontend/wailsjs` changes only when they correspond to Go facade changes. |
| Wails binding imported outside `frontend/src/api` | Move the import into an API wrapper and keep pages/components on wrapper or composable calls. |
| Desktop window cannot be inspected by automation | Use the Web smoke fallback, record that limitation, and keep Go service truth covered by Go tests plus Wails build. |
| Vite chunk-size warning | Treat as non-blocking for this shell; revisit when route or feature splitting exists. |

## Known Limits

- The launch smoke proves the app process starts; manual or browser-backed smoke is still used to inspect Workbench content.
- The Web smoke fallback uses a mocked Wails binding and does not replace Go service tests or Wails build.
- High-fidelity Wails UI automation, installer validation, signing, auto-update, and cross-platform packaging belong to later work.
