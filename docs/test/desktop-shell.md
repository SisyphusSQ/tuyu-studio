# Desktop Shell Smoke

## Scope

This runbook covers the TOO-160 Wails Desktop App shell only.

It verifies that the repository has a Wails v2 application entry, a thin Go App facade, a frontend asset host, and repeatable commands for later Workbench work. It does not verify project storage, Canvas, Script-to-Shot, Asset Library, provider runtime, or Review flows.

## Commands

```bash
make harness-verify
go test ./...
npm --prefix frontend install
npm --prefix frontend run build
make desktop-shell-build
git diff --check
```

`make desktop-shell-build` uses the pinned Wails CLI module through:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build -clean
```

## Expected Result

- Wails build completes and produces a local desktop bundle under ignored build output.
- The shell window title is `Tuyu Studio`.
- The Wails App facade exposes `AppInfo` and `ShellHealth`.
- The facade does not read or write local project files.
- Frontend output is generated from `frontend/` and is not committed except for `frontend/dist/.gitkeep`.

## Current Result

Verified for TOO-160 on macOS arm64:

| Command | Result |
| --- | --- |
| `make harness-verify` | pass |
| `go test ./...` | pass |
| `npm --prefix frontend install` | pass |
| `npm --prefix frontend run build` | pass |
| `make desktop-shell-build` | pass |
| `open build/bin/tuyu-studio.app` plus process check | pass |
| Computer Use window/readback check | pass |
| `git diff --check` | pass |

The built app was launched from ignored local build output, the `tuyu-studio` process was observed, and the window readback showed `Tuyu Studio`, `Desktop shell ready`, `Shell boundary`, and `Canvas-first surface reserved`. The process was stopped after the smoke. No raw logs, local bundle paths beyond the repo-relative build output, tokens, or machine-private runtime state are committed.
