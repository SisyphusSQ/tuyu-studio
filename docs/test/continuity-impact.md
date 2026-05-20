# Continuity Impact

## Scope

This runbook covers TOO-177 only: ContinuityRule storage/display, locked rule and locked binding unlock semantics, missing asset/digest health impact, and Shot `context_dirty` propagation when continuity references change.

It does not cover provider execution, model-side image understanding, package export internals, cloud sync, billing/account state, or automated visual continuity repair.

## Verification Matrix

| Check | Status | Evidence |
| --- | --- | --- |
| Focused Go tests | Passed | `go test ./internal/project ./internal/shell` covers locked rules, explicit unlock reason/audit, locked main reference blocking, binding unlock, main-reference dirty propagation, and Health/Canvas affected Shot badges. |
| Frontend typecheck | Passed | `npm --prefix frontend run typecheck` accepts the new Continuity DTOs, Wails bindings, and Workbench rule controls. |
| Frontend unit tests | Passed | `npm --prefix frontend run test:unit` covers continuity command builders and DTO normalization. |
| Wails binding refresh | Passed | `make desktop-shell-build` regenerated `frontend/wailsjs` with `ProjectContinuityList`, `ProjectContinuityRuleSave`, `ProjectContinuityRuleUnlock`, and `ProjectAssetBindingUnlock`. |
| Browser smoke | Passed | Vite on `127.0.0.1:5193` with mocked Wails bridge rendered Continuity controls, missing asset / DIRTY Canvas badges, Save rule and Unlock rule actions, and captured `output/playwright/too-177-continuity.png`. |
| Independent review | Passed after rework | Initial review found locked-rule save bypass, missing main-reference unlock UI path, and underspecified binding unlock selector. All were reworked and focused re-review reported no findings. |

## Expected Behavior

- `ContinuityRule` rows include `targetType`, `targetId`, `rule`, `severity`, `locked`, and `createdBy`.
- Locked rules require an unlock reason and write an audit event before they become editable/unlocked.
- Locked main-reference bindings cannot be replaced or cleared by the main-reference command; they must be unlocked with a reason first.
- Asset missing, digest mismatch, managed-reference risk, and stale binding health items include affected asset/profile/Shot/package identifiers where available.
- Character/Scene/Prop main-reference or continuity-rule changes mark referencing Shots `context_dirty`.
- Canvas nodes receive continuity badges through HealthReport affected objects, including `missing_asset`, `context_dirty`, `stale_binding`, and `managed_reference`.

## Residual Risks

- ImpactReport package coverage is limited to package manifests already present under the local package directory.
- Browser smoke uses a mocked Wails bridge; real service behavior is covered by Go tests and generated Wails binding refresh.
