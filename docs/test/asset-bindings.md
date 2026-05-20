# Asset Binding And Main Reference

## Scope

This note records the TOO-176 verification surface for binding indexed Assets to Character, Scene, and Prop profiles, setting or clearing a profile main reference, projecting compact reference metadata into Canvas/Inspector, and reopening the example project with committed continuity fixture data.

## Current Run Summary

| Check | Result | Evidence |
| --- | --- | --- |
| Go focused binding tests | Passed | `go test ./internal/project -run 'TestAssetBinding|TestMainReference' -count=1` |
| Example fixture and binding tests | Passed | `go test ./internal/project -run 'TestExampleFixture|TestAssetBinding|TestMainReference' -count=1` |
| Project and shell suites | Passed | `go test ./internal/project ./internal/shell` |
| Frontend unit tests | Passed | `npm --prefix frontend run test:unit` |
| Frontend typecheck | Passed | `npm --prefix frontend run typecheck` |
| Frontend build | Passed | `npm --prefix frontend run build` |
| Full alpha-shell gate | Passed | `make alpha-shell-verify` |
| Browser smoke | Passed | Playwright CLI opened Vite on `127.0.0.1:5192`, injected mocked Wails bindings, verified first-load continuity rows, exercised List bindings, Bind, Main ref, Clear main ref, verified 4 graph reloads, visible binding event summary, and captured `output/playwright/too-176-asset-bindings.png` |
| Diff hygiene | Passed | `git diff --check` |
| Independent review | Passed after rework | Initial review found 1 high and 3 medium findings; rework added empty-asset validation, reserved generic `main_reference` binding, fixture MIME alignment, and initial continuity loading. Focused re-review reported no blocking/high/medium findings. |

## Coverage Notes

- Binding supports Character, Scene, and Prop targets only; Shot dirty propagation, Continuity lock behavior, image understanding, automatic recognition, and similarity matching stay out of TOO-176 scope.
- Generic Bind rejects `main_reference`; main reference changes must use the dedicated set/clear action so profile `mainReferenceAssetId` and Asset index bindings stay consistent.
- Main reference set now requires a non-empty Asset id unless `clear=true`; malformed API calls do not clear an existing profile main reference.
- Duplicate binding defaults to `binding_duplicate`; choosing `reuse` preserves the existing binding and refreshes the profile reference list.
- Main reference set/replace removes the previous `main_reference` binding for that target before adding the new one; clear removes main-reference binding state without deleting ordinary reference bindings.
- Profile writes preserve existing JSON fields and add baseline `name`, type-specific descriptors, `referenceAssetIds`, `mainReferenceAssetId`, and `lockedRules` where needed.
- Canvas Graph View carries only compact display metadata for profiles: profile id, reference asset ids, binding count, main reference asset id/path/thumbnail, and main-reference status. Full profile arrays and raw design fields are not exposed through Canvas DTOs.
- The Workbench Asset rail lists bindings on first load, binds an Asset to the selected target, sets or clears the main reference, refreshes Asset/Profile rows, and reloads the graph after successful mutating actions.
- The committed alpha fixture now includes `assets/index.json` with importer-compatible MIME values and real digest/size values for the three reference assets, plus baseline main-reference fields on the Character, Scene, and Prop profiles.
- Design source files under `/Users/suqing/Coding/design/tuyu-design` are read-only reference for this issue and were not modified.

## Browser Smoke

Use a local Vite target with a mocked Wails bridge. The mock must expose `ProjectAssetsList`, `ProjectAssetBindingsList`, `ProjectAssetBind`, `ProjectMainReferenceSet`, and `ProjectGraphView`.

Expected visible behavior:

- Asset rail loads existing indexed reference assets and the continuity profile list.
- `Bindings` displays Character, Scene, and Prop profiles with binding counts and main reference labels.
- `Bind` sends the selected asset, target type, target id, purpose, duplicate policy, and correlation id through the Wails wrapper.
- `Main ref` updates the selected profile main reference, refreshes the Asset/Profile rows, and reloads Canvas.
- `Clear main reference` returns a warning state for that profile and the Inspector sees `mainReferenceStatus=missing` after graph reload.
- Runs/Audit include binding events, binding health items, and any structured binding errors or recovery actions.
