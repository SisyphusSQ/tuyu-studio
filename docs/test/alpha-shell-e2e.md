# Alpha Shell End-To-End Acceptance

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-181` Alpha shell end-to-end acceptance and Linear writeback
- Current conclusion: E2E validation passed; closeout gates are being recorded before Linear final state update
- Automatic entry points: `make alpha-shell-verify`, `make alpha-shell-smoke-launch`, browser Workbench smoke with mocked Wails bridge
- Sensitive information handling: no credentials, tokens, cookies, database hosts, connection strings, row keys, raw command logs, full temporary paths, complete local screenshots, or machine-private runtime state are written here.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| Full alpha shell gate | Passed | `make alpha-shell-verify` completed harness, Go, frontend, Wails build smoke, and whitespace checks |
| Desktop launch smoke | Passed | `make alpha-shell-smoke-launch` built the Wails app, launched the Alpha shell app process, confirmed startup, and cleaned up the launched process |
| Browser Workbench E2E smoke | Passed | Local Vite plus mocked Wails bridge drove Project Canvas, Local Project, Script-to-Shot, Asset/Continuity, Handoff Package, Mock Run, Result import, Trace, and Review approved |
| Linear relationship audit | Pending closeout writeback | Current project/milestone state was re-read from Linear; final issue state and master gate summary are written to Linear after merge |
| Independent review | Passed | Independent subagent review reported `blocking_findings: none` and no sensitive-data leakage |
| Diff hygiene | Passed | `git diff --check` completed after the closeout document update |
| Harness review gate | Passed | `make harness-review-gate PLAN=.agents/plans/2026-05-20-too-181-alpha-shell-e2e-writeback.md` reported `blocking_findings=none` |

## Scope

This runbook covers the `alpha-shell` cross-slice acceptance path owned by `TOO-181`.

It verifies that the already implemented Alpha shell slices form one runnable local workflow:

1. Desktop shell starts.
2. Workbench opens an Alpha project and displays a Project Canvas.
3. Script material loads and produces accepted Shot candidate state.
4. Asset Library and Continuity show a main reference and locked rule state.
5. Shot context validates as `context_ready`.
6. Handoff Package exports a ready package.
7. Deterministic mock run completes with placeholder output.
8. Result import creates a reviewable result.
9. Trace links Result -> Asset -> Shot -> Package -> Mock Run.
10. Review can be approved from the Workbench.

This runbook does not cover real provider execution, Agent MCP submit/poll, final render, installer signing, auto-update, cloud sync, billing, or cross-platform packaging.

## Acceptance Matrix

| Area | Status | Evidence |
| --- | --- | --- |
| Desktop app startup | Passed | Launch smoke confirms the built Wails bundle starts a `tuyu-studio` process and cleans it up |
| Local project | Passed | Browser smoke opens `Tuyu Studio Alpha Project`, reports clean health, and shows local owned project summary |
| Project Canvas | Passed | Browser smoke shows a loaded canvas and later graph growth from 2 nodes / 1 edge to 4 nodes / 3 edges after package and result steps |
| Script-to-Shot | Passed | Browser smoke loads `Alpha Script` and shows `1 rows / 1 accepted` candidate state |
| Asset Library and Continuity | Passed | Browser smoke shows `asset_mina_ref`, one profile, `Mina`, one lineage row, and one locked continuity rule |
| Shot context | Passed | Browser smoke validates `shot_001` as `context_ready` with no missing or blocking items |
| Handoff Package | Passed | Browser smoke exports `pkg_alpha_shot_002_v1` with ready status and manifest, prompt, checklist, digest, and reference metadata visible |
| Mock Run | Passed | Browser smoke starts `run_mock_shot_002_001`, shows `mock_local`, completed status, attempt count, run paths, and placeholder output |
| Result import | Passed | Browser smoke imports `result_alpha_shot_002_take_001`, shows approved take state, result file, record path, and asset id |
| Result trace | Passed | Browser smoke shows `shot_002`, `pkg_alpha_shot_002_v1`, `run_mock_shot_002_001`, and `asset_result_alpha_002` in the trace summary |
| Review | Passed | Browser smoke saves review and shows Review import as `approved` |
| Linear writeback | Pending closeout writeback | Workpad, plan doc, and Master Done Gate are updated after review and merge gate pass |

## Commands

Run the full reusable local gate:

```bash
make alpha-shell-verify
```

Run the optional launch smoke on macOS:

```bash
make alpha-shell-smoke-launch
```

After independent review, run closeout gates:

```bash
make harness-review-gate PLAN=.agents/plans/2026-05-20-too-181-alpha-shell-e2e-writeback.md
git diff --check
```

## Browser Workbench E2E Smoke

Use a local Vite target with a mocked Wails bridge. This validates Workbench UI projection across the already implemented API wrapper surfaces. Real service behavior remains covered by Go tests, frontend unit tests, generated Wails bindings, and `make alpha-shell-verify`.

Safe replay entry:

```bash
npm --prefix frontend run dev -- --port 5194
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless=new \
  --disable-gpu \
  --remote-debugging-port=9223 \
  --user-data-dir="$(mktemp -d)" \
  about:blank
```

Then drive the Workbench through the Chrome DevTools Protocol by injecting a mocked `window.go.main.App` before assertions. Keep the mock DTOs sanitized and project-relative; do not write raw browser logs, screenshots, temporary directories, or machine-local user paths into committed evidence.

Final smoke checks:

- `assetsContinuity`: passed
- `scriptToShot`: passed
- `canvas`: passed
- `project`: passed
- `shotContext`: passed
- `handoffPackage`: passed
- `mockRun`: passed
- `resultImport`: passed
- `resultTrace`: passed
- `reviewApproved`: passed

The smoke does not persist generated fixture data into the repository.

## Residual Risks

- The browser Workbench E2E uses a mocked Wails bridge for UI projection; Go services, Wails facade bindings, and build smoke cover service truth.
- Desktop launch smoke confirms process startup; content inspection still uses browser-backed Workbench smoke because desktop window automation is not the reliable assertion layer in this repository.
- Vite production build still reports the known non-blocking chunk-size warning.
- Real provider execution, final render, installer signing, auto-update, and cloud sync remain out of Alpha shell scope.

## Recovery And Rerun

| Failure point | Recovery |
| --- | --- |
| `make alpha-shell-verify` fails | Fix the failing harness, Go, frontend, Wails, or whitespace step inside the active issue scope, then rerun the full gate |
| Launch smoke cannot start the app | Rerun `make desktop-shell-build`, then rerun `make alpha-shell-smoke-launch`; record sanitized process-start failure evidence if it still fails |
| Browser smoke cannot reach Workbench controls | Confirm the Vite target is running, inject the Wails bridge mock before driving actions, and rerun the smoke |
| Browser smoke state assertion fails | Inspect visible Workbench state and decide whether the assertion expected hidden form state or a real UI regression |
| Linear writeback fails | Add a blocker comment if provider access is unavailable; otherwise retry Workpad, Plan Doc, and Master gate writeback from the latest merge state |

## Closeout Notes

- Recovery point: branch `suqing/too-181-alpha-shell-e2e-writeback` with the committed sanitized E2E runbook and passing local gates.
- Next action: complete independent subagent review, run closeout gates, merge the branch, mark `TOO-181` Done, then close `TOO-158` if all child Execution Issues remain Done.
