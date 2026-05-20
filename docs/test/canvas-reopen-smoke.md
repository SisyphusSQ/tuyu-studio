# Canvas Reopen Smoke

## Current Verification Result

- Recorded at: 2026-05-20
- Recorded directory: repository root
- Task: `TOO-171` example project Canvas reopen smoke
- Current conclusion: passed with ProjectStore-backed persistence proof and mocked-Wails browser display smoke
- Sensitive information handling: do not record credentials, tokens, cookies, database hosts, connection strings, row keys, raw logs, full temporary paths, full local-only screenshots, or machine-private runtime state.

### This Run Result

| Step | Result | Sanitized summary |
| --- | --- | --- |
| Targeted ProjectStore reopen test | Passed | `examples/alpha-project` opened, projected into Canvas, saved layout/theme/grid/viewport, reopened through a fresh Store, and restored the saved Canvas state |
| Canvas display projection | Passed | Graph View projected domain nodes, valid edges, one ReferenceGroup, one ProductionFrame, and required status badge semantics |
| Browser mocked-Wails smoke | Passed | Workbench rendered a non-empty Canvas surface, Inspector selected-node details, relation summary, Bottom Bar save/health/event feedback, Warm theme, and saved graph version after reload |
| Scope guard | Passed | Smoke used local fixture/DTO paths only; no provider, model generation, external Agent, cloud sync, billing, or account path was required |

## Scope

This runbook covers the Canvas reopen acceptance card for the `TOO-155` Canvas Master.

It verifies that the alpha fixture can be opened, projected into Canvas, saved with changed Canvas presentation state, and reopened with the same persisted Canvas state. It does not verify Script-to-Shot, Asset Library, Continuity business logic, Handoff, Mock Run, Review, provider execution, or a full screenshot regression matrix.

## Commands

Run the focused ProjectStore reopen proof:

```bash
go test ./internal/project -run 'TestExampleAlphaProject(CanvasReopen|Open|Health|Paths)' -count=1
```

Run the full local gate:

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

## Expected Reopen Behavior

- `examples/alpha-project` opens read/write with no blocking health items.
- Graph View returns a non-empty Canvas for `proj_alpha_fixture`.
- Canvas projection includes domain nodes, valid relations, one ReferenceGroup, one ProductionFrame, and status badges for `context_ready`, `context_dirty`, `package_ready`, and `pending_review`.
- Saving Canvas layout increments the graph version.
- A fresh Store can reopen the fixture and read back the saved node position, node size, viewport, Warm Light theme, and grid settings.
- Reopen does not require a real provider, real model generation, external Agent MCP, account state, or cloud sync.

## Browser Smoke

The browser smoke starts the Vite frontend and injects a mocked Wails `window.go.main.App` before app load. The mock returns a fixture-shaped Graph View DTO and persists layout saves in memory so the UI can exercise display, save, and reload behavior.

Expected browser observations:

- Top Bar displays the fixture graph version and clean health state.
- Canvas overlay displays nonzero node, edge, frame, and reference counts.
- ReferenceGroup and ProductionFrame labels render in the Canvas.
- Selecting a node opens Inspector details and relation summaries.
- Switching to Warm Light keeps Canvas, nodes, panels, and badges readable.
- Saving Canvas layout updates save feedback and increments the graph version.
- Reload reads the saved mocked Canvas state and preserves Warm Light feedback.
- Bottom Bar displays selection, zoom, grid, save status, health, and latest event.

## Current Smoke Observations

- Targeted ProjectStore reopen test passed and restored:
  - graph version increment from `3` to `4`
  - viewport `x=-144`, `y=96`, `zoom=1.35`
  - theme `warm_light`
  - grid `visible=false`, `size=32`, `opacity=0.16`
  - `node_shot_001` position and size
- Browser smoke readback:
  - graph view calls: 2
  - save commands: 1
  - Inspector tabs: `Props`, `Links`, `Rules`, `Tasks`, `Runs`, `Audit`
  - Canvas overlay: `4 nodes`, `3 edges`, `1 frames`, `1 refs`
  - Bottom Bar: selected `Script Source`, saved state, clean health, latest graph event

## Known Limits

- Browser smoke uses mocked Wails bindings and does not replace Go ProjectStore tests or Wails build.
- The browser smoke records a sanitized textual result rather than committing raw screenshots.
- Headless direct Canvas node hit-testing can be environment-sensitive; the smoke verifies selection via the existing Canvas command path.
- Vite production build still emits the known non-blocking G6 chunk-size warning.

## Follow-up Candidates

- Full cross-platform Canvas smoke.
- Screenshot diff regression matrix for Canvas and Inspector.
- Graph relation edit smoke.
- Blueprint, search, minimap, and keyboard navigation execution cards.
