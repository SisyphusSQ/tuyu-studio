# ScriptDocument Import And Edit Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-172` ScriptDocument import and edit
- Current conclusion: passed local focused checks before independent review
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw command logs, full temporary paths, or machine-private runtime state are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| `go test ./internal/project` | Passed | ScriptDocument save, load, source asset import, empty input, missing asset, unsafe id/path, scene preservation, and failed-write preservation were verified |
| `go test ./...` | Passed | Root package had no tests; project and shell package checks passed |
| `npm --prefix frontend run typecheck` | Passed | ScriptDocument UI/API TypeScript boundary compiled |
| `npm --prefix frontend run test:unit` | Passed | API command-builder tests passed with existing Workbench tests |
| `make desktop-shell-build` | Passed | Wails bindings were regenerated and the darwin/arm64 app package built |
| Browser mocked-Wails smoke | Passed | Script rail rendered, default source asset import filled raw text, edited raw text saved, and load restored the edited text with saved/loaded events visible |

## Scope

This runbook covers `TOO-172` only: importing or pasting a script into a persistent `ScriptDocument`, editing its metadata/raw text, reopening the same document, and exposing the flow through the Workbench Script rail.

It does not create scenes, shots, image prompts, run packages, model provider jobs, review artifacts, or handoff exports. Those belong to later Script-to-Shot and Handoff execution cards.

## Commands

```bash
go test ./internal/project
go test ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:unit
make desktop-shell-build
```

## Expected Behavior

- A script can be saved from pasted raw text without changing whitespace or line breaks.
- A script can be imported from a project-relative `script_source` asset.
- Empty input returns `script_empty`.
- Missing or unsafe source paths return `script_asset_missing` with recovery actions.
- Failed writes return `save_failed` and leave the previous document file intact.
- Reopen loads the same `ScriptDocument` JSON and preserves any existing manual scene records.
- The Workbench Script rail can import the default fixture source asset, load the current document, edit metadata/raw text, save, and display DTO events/errors.

## Browser Smoke

The browser smoke starts the Vite frontend and injects a mocked Wails `window.go.main.App` before app load. The mock covers Graph View and ScriptDocument save/load calls so the Workbench can exercise the rail editor in a local browser while Go service behavior remains covered by Go tests and generated bindings.

Expected browser observations:

- Script tab renders `script_main`, title, source asset, logline, synopsis, raw text, Import asset, Load, and Save controls.
- Import asset writes fixture raw text into the editor and emits `script.document.saved`.
- Editing raw text enables Save.
- Save then Load preserves the edited raw text and emits `script.document.loaded`.
- The Bottom Bar latest event reflects the ScriptDocument event stream.

## Storage

- Default script id: `script_main`
- Project-relative storage directory: `assets/inputs/scripts/`
- File shape: `{scriptId}.script.json`

## Known Limits

- Browser UI smoke is covered by generated Wails build and frontend type/unit checks in this card; full screenshot smoke can be added when Script-to-Shot UI grows beyond the rail editor.
- Scene and shot creation remains out of scope for `TOO-172`.
