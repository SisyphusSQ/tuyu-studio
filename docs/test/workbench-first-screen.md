# Workbench First Screen Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-161` Vue Workbench first-screen layout
- Current conclusion: passed after rework
- Automated entries: frontend typecheck/build, Go tests, harness verify, Wails desktop build, whitespace check, Wails binding isolation check
- Smoke entries: Wails desktop process started; Web smoke confirmed five workbench zones at 1280x820 and 1024x700
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw logs, full temporary paths, or full local-only screenshots are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| `npm --prefix frontend install` | Passed | dependencies installed with no vulnerability report |
| `npm --prefix frontend run typecheck` | Passed | Vue TypeScript check completed |
| `npm --prefix frontend run build` | Passed | Vite production build completed |
| `go test ./...` | Passed | root package had no tests; `internal/shell` passed |
| `make harness-verify` | Passed | harness check passed |
| `make desktop-shell-build` | Passed | Wails v2 darwin/arm64 package completed |
| `git diff --check` | Passed | no whitespace errors |
| Wails binding isolation | Passed | generated Wails binding imports are isolated to `frontend/src/api/*`; page and business components do not import `wailsjs/go` or `wailsjs/runtime` |
| Web smoke 1280x820 | Passed | Top Bar, Left Panel, Canvas placeholder, Inspector, Bottom Bar, and theme switch visible; no page horizontal overflow |
| Web smoke 1024x700 | Passed | five workbench zones visible; no page horizontal overflow |

## Scope

This runbook covers `TOO-161` only: the Vue 3 + TypeScript Workbench first screen on top of the Wails desktop shell.

It verifies that the app opens into a production workbench with five stable zones:

- Top Bar
- Left Panel
- Project Canvas placeholder
- Inspector
- Bottom Bar

## Commands

```bash
npm --prefix frontend install
npm --prefix frontend run typecheck
npm --prefix frontend run build
go test ./...
make harness-verify
make desktop-shell-build
```

## Desktop Smoke

Open `build/bin/tuyu-studio.app` after `make desktop-shell-build`.

Expected visible state:

- Top Bar shows project identity, saved status, queue status, theme toggle, save/export/settings commands.
- Left Panel shows Ant Design Vue Menu/Tabs controls and asset/script/blueprint/search/history entry points.
- Project Canvas area is the dominant surface and shows a placeholder production frame, canvas toolbar, stable node placeholders, and an explicit note that real graph rendering is out of scope for `TOO-161`.
- Inspector shows selection summary, property/relation/task tabs, and the known `TOO-162` API boundary note.
- Bottom Bar shows selection, zoom/grid, health, event, and layout reset status.

Known limits:

- No AntV G6 rendering, drag, edge creation, or saved viewport in this issue.
- No local project create/open/save behavior in this issue.
- No Wails generated binding import is used in page or business components; generated binding imports belong under `frontend/src/api/*`.
