# Design Details Module Split Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure `docs/design/details` into one subdirectory per approved 00-09 module and continue refining each module into production-ready subdocuments.

**Architecture:** Keep the existing 00-09 module order as the stable navigation layer. Move each current flat module document into that module directory as `README.md`, then add focused subdocuments for requirements, schemas, workflows, state, error handling, recovery, and acceptance. Update root indexes so `docs/design/README.md` and `docs/design/details/README.md` remain the only top-level navigation surfaces.

**Tech Stack:** Markdown documentation, repo harness scripts, POSIX shell checks, `rg`, `find`, `make harness-check`, `make harness-review-gate`.

---

## Scope

This plan implements the approved spec in:

`/Users/suqing/Coding/golang/00_self/tuyu-studio/.agents/plans/2026-05-14-design-details-module-split-spec.md`

The work is docs-only. It does not create Linear issues, update Linear documents, or change application code.

## File Structure

### Modify

- `docs/design/README.md`
  Update reading order links from flat `details/*.md` files to module directory `README.md` files.

- `docs/design/details/README.md`
  Replace the flat document map with a module-directory map, child-document map, cross-module dependency map, and validation checklist.

- `.agents/plans/2026-05-14-design-docs-production-refine.md`
  Update progress / verify summary after the docs split so the earlier design-doc plan remains accurate.

### Delete After Migration

- `docs/design/details/00-product-scope-and-glossary.md`
- `docs/design/details/01-local-project-storage.md`
- `docs/design/details/02-creative-graph-domain-model.md`
- `docs/design/details/03-asset-library-and-continuity.md`
- `docs/design/details/04-script-shot-package-workflow.md`
- `docs/design/details/05-instruction-stack-and-skills.md`
- `docs/design/details/06-ai-runtime-and-provider-adapters.md`
- `docs/design/details/07-frontend-workbench-experience.md`
- `docs/design/details/08-security-privacy-observability.md`
- `docs/design/details/09-delivery-acceptance-and-test-plan.md`

### Create

```text
docs/design/details/
├─ 00-product-scope-and-glossary/
│  ├─ README.md
│  ├─ product-boundary.md
│  ├─ user-workflows.md
│  └─ glossary.md
├─ 01-local-project-storage/
│  ├─ README.md
│  ├─ directory-layout.md
│  ├─ schema-and-migration.md
│  ├─ save-lock-recovery.md
│  └─ health-check.md
├─ 02-creative-graph-domain-model/
│  ├─ README.md
│  ├─ nodes-and-edges.md
│  ├─ graph-state-machine.md
│  ├─ context-resolution.md
│  └─ consistency-rules.md
├─ 03-asset-library-and-continuity/
│  ├─ README.md
│  ├─ asset-ingestion-and-indexing.md
│  ├─ asset-binding-and-lineage.md
│  ├─ continuity-rules.md
│  └─ deletion-and-recovery.md
├─ 04-script-shot-package-workflow/
│  ├─ README.md
│  ├─ script-to-scene.md
│  ├─ shot-card-lifecycle.md
│  ├─ package-export.md
│  └─ result-ingestion-and-review.md
├─ 05-instruction-stack-and-skills/
│  ├─ README.md
│  ├─ instruction-compile-order.md
│  ├─ skill-registry-and-routing.md
│  ├─ output-contracts.md
│  └─ conflict-handling.md
├─ 06-ai-runtime-and-provider-adapters/
│  ├─ README.md
│  ├─ runtime-gateway.md
│  ├─ image-generation-adapter.md
│  ├─ provider-handoff-adapter.md
│  └─ runtime-errors-and-retry.md
├─ 07-frontend-workbench-experience/
│  ├─ README.md
│  ├─ workspace-layout.md
│  ├─ canvas-interactions.md
│  ├─ inspector-and-task-panels.md
│  └─ production-feedback.md
├─ 08-security-privacy-observability/
│  ├─ README.md
│  ├─ path-guard-and-permissions.md
│  ├─ privacy-and-redaction.md
│  ├─ audit-events.md
│  └─ observability-and-recovery.md
└─ 09-delivery-acceptance-and-test-plan/
   ├─ README.md
   ├─ delivery-slices.md
   ├─ acceptance-matrix.md
   ├─ test-strategy.md
   └─ release-readiness.md
```

## Content Rules

Every module `README.md` must preserve the useful content from the current flat document and add a short child-document index.

Every child document must include these sections, adjusted for product-only topics when needed:

```markdown
# Product Boundary

## 目标与边界

## 输入 / 输出

## 核心对象或规则

## 状态推进

## 错误语义

## 恢复与重试

## 验收标准
```

Product or glossary documents may replace `状态推进` with `用户路径` or `术语规则`, but still need `错误语义` / `异常路径` and `验收标准` where applicable.

## Task 1: Snapshot Current Details Docs

**Files:**
- Read: `docs/design/details/*.md`
- Read: `docs/design/README.md`
- Read: `docs/design/details/README.md`
- Read: `.agents/plans/2026-05-14-design-details-module-split-spec.md`

- [ ] **Step 1: Confirm current files**

Run:

```bash
find docs/design/details -maxdepth 1 -type f -print | sort
```

Expected: `README.md` plus flat `00-*.md` through `09-*.md`.

- [ ] **Step 2: Confirm no existing module directories**

Run:

```bash
find docs/design/details -mindepth 1 -maxdepth 1 -type d -print | sort
```

Expected: no output.

- [ ] **Step 3: Save current word counts for migration review**

Run:

```bash
wc -l docs/design/details/*.md
```

Expected: line counts for all flat docs. Use this only to compare that migrated `README.md` files are not accidentally empty.

## Task 2: Create Module Directories And Move Existing Content

**Files:**
- Create: all module directories under `docs/design/details/00-*` through `09-*`
- Create: each module `README.md`
- Delete: flat `docs/design/details/00-*.md` through `09-*.md`

- [ ] **Step 1: Create directories**

Run:

```bash
mkdir -p \
  docs/design/details/00-product-scope-and-glossary \
  docs/design/details/01-local-project-storage \
  docs/design/details/02-creative-graph-domain-model \
  docs/design/details/03-asset-library-and-continuity \
  docs/design/details/04-script-shot-package-workflow \
  docs/design/details/05-instruction-stack-and-skills \
  docs/design/details/06-ai-runtime-and-provider-adapters \
  docs/design/details/07-frontend-workbench-experience \
  docs/design/details/08-security-privacy-observability \
  docs/design/details/09-delivery-acceptance-and-test-plan
```

Expected: command exits 0.

- [ ] **Step 2: Move flat documents into module README files**

Run:

```bash
mv docs/design/details/00-product-scope-and-glossary.md docs/design/details/00-product-scope-and-glossary/README.md
mv docs/design/details/01-local-project-storage.md docs/design/details/01-local-project-storage/README.md
mv docs/design/details/02-creative-graph-domain-model.md docs/design/details/02-creative-graph-domain-model/README.md
mv docs/design/details/03-asset-library-and-continuity.md docs/design/details/03-asset-library-and-continuity/README.md
mv docs/design/details/04-script-shot-package-workflow.md docs/design/details/04-script-shot-package-workflow/README.md
mv docs/design/details/05-instruction-stack-and-skills.md docs/design/details/05-instruction-stack-and-skills/README.md
mv docs/design/details/06-ai-runtime-and-provider-adapters.md docs/design/details/06-ai-runtime-and-provider-adapters/README.md
mv docs/design/details/07-frontend-workbench-experience.md docs/design/details/07-frontend-workbench-experience/README.md
mv docs/design/details/08-security-privacy-observability.md docs/design/details/08-security-privacy-observability/README.md
mv docs/design/details/09-delivery-acceptance-and-test-plan.md docs/design/details/09-delivery-acceptance-and-test-plan/README.md
```

Expected: command exits 0 and no flat `00-*.md` files remain.

- [ ] **Step 3: Add child-document indexes to module README files**

Use `apply_patch` to add a `## 子文档` table near the top of each module `README.md`. Example for `01-local-project-storage/README.md`:

```markdown
## 子文档

| 文档 | 作用 |
| --- | --- |
| `directory-layout.md` | Studio root、project root、资产、包、运行记录、审计目录 |
| `schema-and-migration.md` | `project.json`、schema version、迁移、备份和回滚 |
| `save-lock-recovery.md` | 原子写、自动保存、项目锁、崩溃恢复 |
| `health-check.md` | 缺文件、断链、digest、外部引用、迁移风险检查 |
```

Repeat with the exact child files from the spec for each module.

## Task 3: Write 00 Product Scope Child Docs

**Files:**
- Create: `docs/design/details/00-product-scope-and-glossary/product-boundary.md`
- Create: `docs/design/details/00-product-scope-and-glossary/user-workflows.md`
- Create: `docs/design/details/00-product-scope-and-glossary/glossary.md`

- [ ] **Step 1: Write `product-boundary.md`**

Content must cover:

- Product definition: local AI film pre-production and generation handoff workbench.
- Included capabilities: local project, Creative Graph, asset continuity, instruction stack, handoff package, result review.
- Excluded capabilities: SaaS collaboration, professional editing timeline, silent external submission, generic file manager, black-box AI agent.
- Acceptance criteria: a future implementation issue can decide whether a feature is in or out from this file.

- [ ] **Step 2: Write `user-workflows.md`**

Content must cover these workflows:

- New project to first Shot package.
- Character/scene asset import to continuity-locked Shot.
- Prompt run preview to structured output.
- Handoff package export to manual external generation.
- Result ingestion to review and revision.

Each workflow must include trigger, happy path, interruption point, failure path, and acceptance criteria.

- [ ] **Step 3: Write `glossary.md`**

Content must define:

- Product-level terms: Project, Studio Root, Creative Graph, Node, Edge.
- Production terms: Shot, PromptRun, GenerationPackage, Take, Review Status.
- Adapter terms: RuntimeGateway, ProviderProfile, HandoffAdapter, ImageGenerationAdapter.
- Naming rules: use provider/profile/adapter abstractions in formal docs; avoid binding architecture to a specific vendor.

## Task 4: Write 01 Storage Child Docs

**Files:**
- Create: `docs/design/details/01-local-project-storage/directory-layout.md`
- Create: `docs/design/details/01-local-project-storage/schema-and-migration.md`
- Create: `docs/design/details/01-local-project-storage/save-lock-recovery.md`
- Create: `docs/design/details/01-local-project-storage/health-check.md`

- [ ] **Step 1: Write `directory-layout.md`**

Must include:

- Full tree for `{studio_root}` and `{project_id}`.
- Responsibility table for `project.json`, `graph.json`, `assets/`, `prompts/runs/`, `packages/`, `audit/`, `backups/`, `locks/`.
- Path rules: internal references are project-relative; external references require explicit managed-reference metadata.
- Acceptance: project copy to another directory remains health-checkable.

- [ ] **Step 2: Write `schema-and-migration.md`**

Must include:

- `project.json` schema shape.
- `schemaVersion` vs app version distinction.
- Migration flow: detect, backup, migrate, validate, rollback.
- Error semantics: unsupported schema, failed migration, invalid backup.
- Acceptance: failed migration leaves original project recoverable.

- [ ] **Step 3: Write `save-lock-recovery.md`**

Must include:

- Atomic write sequence.
- Autosave triggers and throttle.
- Project lock states: clean open, stale lock, active lock, takeover.
- Crash recovery path when `lastCleanShutdown=false`.
- Acceptance: interrupted write keeps previous valid version.

- [ ] **Step 4: Write `health-check.md`**

Must include:

- Checks for JSON validity, graph refs, asset missing, digest mismatch, package references, absolute path leakage.
- Severity model: blocking, warning, info.
- Repair actions: restore backup, relink asset, regenerate package, clear stale lock.
- Acceptance: deleting one referenced asset produces a precise affected-object report.

## Task 5: Write 02 Graph Child Docs

**Files:**
- Create: `docs/design/details/02-creative-graph-domain-model/nodes-and-edges.md`
- Create: `docs/design/details/02-creative-graph-domain-model/graph-state-machine.md`
- Create: `docs/design/details/02-creative-graph-domain-model/context-resolution.md`
- Create: `docs/design/details/02-creative-graph-domain-model/consistency-rules.md`

- [ ] **Step 1: Write `nodes-and-edges.md`**

Must include:

- `GraphDocument`, `GraphNode`, `GraphEdge` roles.
- Supported node kinds and relation types.
- Ref boundary: `GraphNode.refId` points to domain object; node data is not domain truth.
- Legal relation matrix and rejection rules.

- [ ] **Step 2: Write `graph-state-machine.md`**

Must include:

- Shot state transitions from draft to approved / needs_revision.
- PromptRun states from queued to completed / failed / cancelled.
- Package and VideoResult status transitions.
- Dirty state handling when upstream context changes.

- [ ] **Step 3: Write `context-resolution.md`**

Must include:

- Selected node expansion rules.
- Traversal depth for Shot, Character, Scene, Prop, Prompt, Package, Result.
- Context digest inputs.
- Overlarge context behavior: summarize, warn, ask user to narrow selection.

- [ ] **Step 4: Write `consistency-rules.md`**

Must include:

- Delete node vs delete domain object.
- Cascading impact report.
- Broken reference placeholders.
- Dirty marking for referenced Shots when continuity changes.

## Task 6: Write 03 Asset Child Docs

**Files:**
- Create: `docs/design/details/03-asset-library-and-continuity/asset-ingestion-and-indexing.md`
- Create: `docs/design/details/03-asset-library-and-continuity/asset-binding-and-lineage.md`
- Create: `docs/design/details/03-asset-library-and-continuity/continuity-rules.md`
- Create: `docs/design/details/03-asset-library-and-continuity/deletion-and-recovery.md`

- [ ] **Step 1: Write `asset-ingestion-and-indexing.md`**

Must include import flow, digest, mime validation, thumbnail generation, duplicate handling, managed reference risk, and acceptance criteria.

- [ ] **Step 2: Write `asset-binding-and-lineage.md`**

Must include binding model, source model, lineage queries, PromptRun and Package relationships, and delete-impact analysis.

- [ ] **Step 3: Write `continuity-rules.md`**

Must include CharacterProfile, SceneProfile, PropProfile, ContinuityRule, severity, locked rules, conflict output, and review-time checks.

- [ ] **Step 4: Write `deletion-and-recovery.md`**

Must include unlink vs delete, recycle bin, restore, permanent purge confirmation, and audit event requirements.

## Task 7: Write 04 Script / Shot / Package Child Docs

**Files:**
- Create: `docs/design/details/04-script-shot-package-workflow/script-to-scene.md`
- Create: `docs/design/details/04-script-shot-package-workflow/shot-card-lifecycle.md`
- Create: `docs/design/details/04-script-shot-package-workflow/package-export.md`
- Create: `docs/design/details/04-script-shot-package-workflow/result-ingestion-and-review.md`

- [ ] **Step 1: Write `script-to-scene.md`**

Must include script import, manual split, AI candidate split, user confirmation, source ranges, and failed-parse fallback.

- [ ] **Step 2: Write `shot-card-lifecycle.md`**

Must include ShotCard fields, required-field validation, context_ready rules, prompt_ready rules, review records, and revision loops.

- [ ] **Step 3: Write `package-export.md`**

Must include package folder layout, manifest schema, prompt file, continuity file, checklist, relative paths, export validation, and re-export versioning.

- [ ] **Step 4: Write `result-ingestion-and-review.md`**

Must include result import, take numbering, Shot/Package binding, review states, wrong-binding recovery, and missing-result behavior.

## Task 8: Write 05 Instruction Child Docs

**Files:**
- Create: `docs/design/details/05-instruction-stack-and-skills/instruction-compile-order.md`
- Create: `docs/design/details/05-instruction-stack-and-skills/skill-registry-and-routing.md`
- Create: `docs/design/details/05-instruction-stack-and-skills/output-contracts.md`
- Create: `docs/design/details/05-instruction-stack-and-skills/conflict-handling.md`

- [ ] **Step 1: Write `instruction-compile-order.md`**

Must include compile order, priority rules, data wrappers, context digest, preview requirements, and acceptance criteria.

- [ ] **Step 2: Write `skill-registry-and-routing.md`**

Must include skill directory structure, source trust states, enable/disable, route resolution, project override, and schema compatibility.

- [ ] **Step 3: Write `output-contracts.md`**

Must include task modes, output schema requirements, parse failure, raw-output preservation, and formal-object pollution prevention.

- [ ] **Step 4: Write `conflict-handling.md`**

Must include locked principle conflict, missing context, invalid output, overlarge context, user confirmation, and retry behavior.

## Task 9: Write 06 Runtime Child Docs

**Files:**
- Create: `docs/design/details/06-ai-runtime-and-provider-adapters/runtime-gateway.md`
- Create: `docs/design/details/06-ai-runtime-and-provider-adapters/image-generation-adapter.md`
- Create: `docs/design/details/06-ai-runtime-and-provider-adapters/provider-handoff-adapter.md`
- Create: `docs/design/details/06-ai-runtime-and-provider-adapters/runtime-errors-and-retry.md`

- [ ] **Step 1: Write `runtime-gateway.md`**

Must include runtime interface, request shape, working directory, writable roots, network flag, event normalization, and task lifecycle.

- [ ] **Step 2: Write `image-generation-adapter.md`**

Must include text-to-image, reference-edit, multi-image fusion, output path allocation, post-run file validation, asset creation, and failure behavior.

- [ ] **Step 3: Write `provider-handoff-adapter.md`**

Must include ProviderProfile, capability flags, manual handoff, API future boundary, BuildPrompt, ExportPackage, SubmitPackage and PollResult semantics.

- [ ] **Step 4: Write `runtime-errors-and-retry.md`**

Must include error codes, user-facing messages, retryability, cancellation, timeout, idempotency keys, and attempt records.

## Task 10: Write 07 Frontend Child Docs

**Files:**
- Create: `docs/design/details/07-frontend-workbench-experience/workspace-layout.md`
- Create: `docs/design/details/07-frontend-workbench-experience/canvas-interactions.md`
- Create: `docs/design/details/07-frontend-workbench-experience/inspector-and-task-panels.md`
- Create: `docs/design/details/07-frontend-workbench-experience/production-feedback.md`

- [ ] **Step 1: Write `workspace-layout.md`**

Must include top bar, left panel, canvas, inspector, bottom bar, density, long text behavior, and no landing-page requirement.

- [ ] **Step 2: Write `canvas-interactions.md`**

Must include drag import, node creation, connection relation picker, frame grouping, search, keyboard support, and illegal action feedback.

- [ ] **Step 3: Write `inspector-and-task-panels.md`**

Must include tabs, unsaved edits, version conflict, task preview, run history, audit, and node-specific task commands.

- [ ] **Step 4: Write `production-feedback.md`**

Must include saved/failed states, run queue, context_dirty, missing asset, package exported, task failed, and actionable error display.

## Task 11: Write 08 Security Child Docs

**Files:**
- Create: `docs/design/details/08-security-privacy-observability/path-guard-and-permissions.md`
- Create: `docs/design/details/08-security-privacy-observability/privacy-and-redaction.md`
- Create: `docs/design/details/08-security-privacy-observability/audit-events.md`
- Create: `docs/design/details/08-security-privacy-observability/observability-and-recovery.md`

- [ ] **Step 1: Write `path-guard-and-permissions.md`**

Must include project-relative paths, user-selected external files, symlink handling, path traversal rejection, external export confirmation, and security audit.

- [ ] **Step 2: Write `privacy-and-redaction.md`**

Must include local privacy defaults, credential exclusion, no silent upload, docs/test redaction, and report export redaction.

- [ ] **Step 3: Write `audit-events.md`**

Must include event schema, categories, required events, action summaries, and examples for package export and path rejection.

- [ ] **Step 4: Write `observability-and-recovery.md`**

Must include PromptRun records, package validation records, health reports, three-layer errors, and recovery actions.

## Task 12: Write 09 Delivery Child Docs

**Files:**
- Create: `docs/design/details/09-delivery-acceptance-and-test-plan/delivery-slices.md`
- Create: `docs/design/details/09-delivery-acceptance-and-test-plan/acceptance-matrix.md`
- Create: `docs/design/details/09-delivery-acceptance-and-test-plan/test-strategy.md`
- Create: `docs/design/details/09-delivery-acceptance-and-test-plan/release-readiness.md`

- [ ] **Step 1: Write `delivery-slices.md`**

Must include implementation slices, dependencies, completion definition, docs sync, and issue-writing guidance.

- [ ] **Step 2: Write `acceptance-matrix.md`**

Must include functional, data, error, recovery, audit, and security acceptance rows for each major capability.

- [ ] **Step 3: Write `test-strategy.md`**

Must include unit, integration, manual acceptance, document validation, and regression commands.

- [ ] **Step 4: Write `release-readiness.md`**

Must include release gates, migration readiness, redaction, package portability, recovery proof, and provider replaceability.

## Task 13: Update Indexes And Remove Old Links

**Files:**
- Modify: `docs/design/README.md`
- Modify: `docs/design/details/README.md`
- Modify: each module `README.md` if child document indexes are missing

- [ ] **Step 1: Update `docs/design/README.md`**

Change reading order links so each detail entry points to module `README.md`, for example:

```markdown
| 4 | `details/01-local-project-storage/README.md` | 本地项目、文件结构、保存、迁移和恢复 |
```

- [ ] **Step 2: Rewrite `docs/design/details/README.md`**

It must include:

- Coverage goal.
- Module directory map.
- Per-module child document table.
- Cross-document invariants.
- Validation commands.

- [ ] **Step 3: Check old flat links are gone**

Run:

```bash
rg -n "details/[0-9][0-9]-[^/]+\\.md" docs/design
```

Expected: no matches. Module directory links such as `details/01-local-project-storage/README.md` are valid and should not be treated as old flat links.

## Task 14: Validate Documentation Structure

**Files:**
- Read: `docs/design/**`
- Modify if needed: any broken links or policy violations found by checks

- [ ] **Step 1: Check final file tree**

Run:

```bash
find docs/design/details -maxdepth 3 -type f -print | sort
```

Expected: `details/README.md`, 10 module `README.md` files, and all child docs listed in the spec.

- [ ] **Step 2: Check old flat files are removed**

Run:

```bash
find docs/design/details -maxdepth 1 -type f -name '[0-9][0-9]-*.md' -print
```

Expected: no output.

- [ ] **Step 3: Check vendor and downgrade wording**

Run:

```bash
rg -n "MVP|纳逗|可灵|Vidu|海螺|Wan|ComfyUI|Stable Diffusion|OpenAI|Codex|Wails|Ant Design|AntV|X6" docs/design
```

Expected: no output.

- [ ] **Step 4: Check sensitive local traces**

Run:

```bash
rg -n "/Users/|token|cookie|secret|password|真实凭据|本机绝对路径" docs/design
```

Expected: only policy statements about not including these items, with no actual credentials or local paths.

- [ ] **Step 5: Run harness**

Run:

```bash
make harness-check
make harness-review-gate PLAN=.agents/plans/2026-05-14-design-docs-production-refine.md
```

Expected: both pass.

## Task 15: Update Plan Summary And Final Review

**Files:**
- Modify: `.agents/plans/2026-05-14-design-docs-production-refine.md`
- Read: `.agents/plans/2026-05-14-design-details-module-split-spec.md`
- Read: this implementation plan

- [ ] **Step 1: Update the production refine plan**

Add a note to `.agents/plans/2026-05-14-design-docs-production-refine.md` that details were subsequently split into module directories under the approved spec. Update verify summary with final validation commands.

- [ ] **Step 2: Run self-review**

Check:

```bash
git diff --check
git status --short
```

Expected: no whitespace errors; changes are limited to `docs/design/**` and the relevant `.agents/plans/**` files.

- [ ] **Step 3: Prepare final summary**

The final response should include:

- New module-directory structure.
- Mention that flat `details/00-*.md` through `09-*.md` were migrated to module `README.md`.
- Validation commands and pass/fail status.
- Any remaining uncommitted state.

## Self-Review

### Spec Coverage

- Directory split: covered by Tasks 2 and 13.
- Per-module child docs: covered by Tasks 3-12.
- Index updates: covered by Task 13.
- No double truth: covered by Task 2 and Task 14.
- Sensitive wording checks: covered by Task 14.
- Harness validation: covered by Task 14.

### Placeholder Scan

This plan contains no placeholder sections or vague implementation tasks. Each child document has a specific file path and required content.

### Scope Check

The spec covers one documentation subsystem: `docs/design/details` restructuring and refinement. It is large but cohesive because all tasks produce one navigable design-doc tree and share the same validation pass.

## Execution Summary

Completed on 2026-05-14.

- Flat `docs/design/details/00-*.md` through `09-*.md` were migrated into matching module directories as `README.md`.
- Added 39 focused child documents under the 10 approved module directories.
- Updated `docs/design/README.md` and `docs/design/details/README.md` to point at module `README.md` files and child documents.
- Updated `.agents/plans/2026-05-14-design-docs-production-refine.md` so the earlier production-refine plan matches the final module split.

Validation result:

- `find docs/design/details -maxdepth 3 -type f -print | sort`: confirmed 50 files under details.
- `find docs/design/details -maxdepth 1 -type f -name '[0-9][0-9]-*.md' -print`: no output.
- `rg -n "details/[0-9][0-9]-[^/]+\\.md" docs/design`: no output.
- `rg -n "MVP|纳逗|可灵|Vidu|海螺|Wan|ComfyUI|Stable Diffusion|OpenAI|Codex|Wails|Ant Design|AntV|X6" docs/design`: no output.
- `rg -n "/Users/|token|cookie|secret|password|真实凭据|本机绝对路径" docs/design`: only policy statements about excluding sensitive material, no concrete secret or local path.
- `git diff --check`: passed.
- `make harness-check`: passed.
- `make harness-review-gate PLAN=.agents/plans/2026-05-14-design-docs-production-refine.md`: passed with `blocking_findings=none`.
