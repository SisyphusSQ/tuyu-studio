# Spec: docs/design/details module split and production refinement

## Goal

将 `docs/design/details` 从当前平铺的 00-09 设计文档，重构为“每个模块一个子目录”的长期文档结构，并在迁移时继续细化需求、领域对象、状态、错误语义、恢复方式、验收矩阵和后续 issue 消费边界。

成功标准：

- 当前 00-09 模块编号和阅读顺序保持稳定。
- 每个模块都有独立目录和 `README.md`。
- 每个模块至少拆出 2-4 篇专题细化文档。
- 原平铺文档内容不丢失，不保留双 truth。
- 正式 `docs/design` 不回流外部品牌主叙事和降级交付口径。
- 新结构能直接支撑后续 Linear issue 拆分和实现计划。

## Approved Direction

用户已确认采用方案：

```text
按当前 00-09 每篇一个子目录：
- 保留现有编号和模块名。
- 原文档迁移为各模块目录下 README.md。
- 每个模块继续拆 2-4 篇生产细节文档。
- details/README.md 作为总索引和覆盖矩阵。
```

不采用：

- 不把 00-09 合并成 5-6 个大域目录。
- 不按后续 issue 重新打散当前模块顺序。
- 不保留平铺旧文档和新目录两套入口。

## Scope

| 范围 | 说明 |
| --- | --- |
| 目录迁移 | 将 `docs/design/details/00-*.md` 到 `09-*.md` 迁入同名模块目录的 `README.md` |
| 索引更新 | 更新 `docs/design/details/README.md` 和 `docs/design/README.md` 的链接与阅读顺序 |
| 子文档新增 | 每个模块新增专题细化文档，承接需求、schema、状态、错误和验收 |
| 内容细化 | 按生产可用标准扩展需求，不写演示版或临时口径 |
| 验证 | 检查链接、敏感词、目录结构、harness gate |

## Non-Goals

| 不做 | 原因 |
| --- | --- |
| 不改业务代码 | 本轮是设计文档重构和细化 |
| 不创建 Linear issue | 用户当前目标是文档拆分和细化 |
| 不重写 architecture | architecture 已有脱敏总览，本轮只补必要链接 |
| 不引入外部品牌主叙事 | 继续保持 provider/profile/adapter 抽象 |
| 不做二级深目录 | 先控制复杂度，等具体模块进入实现时再按需细拆 |

## Target Directory Map

```text
docs/design/details/
├─ README.md
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

## Module Content Standards

每个模块目录使用同一文档语义：

| 文件 | 职责 |
| --- | --- |
| `README.md` | 模块目标、范围、核心对象、主流程、跨文档索引和验收总览 |
| 专题文档 | 单一主题的生产细节，例如 schema、状态、错误、恢复、验收 |

每篇专题文档至少覆盖：

- `目标与边界`
- `输入 / 输出`
- `核心对象或规则`
- `状态推进`
- `错误语义`
- `恢复与重试`
- `验收标准`

如果某篇文档偏产品而非技术，例如 `user-workflows.md`，可以用用户路径、触发点、成功标准、异常路径替代 schema 章节，但仍要有验收标准。

## Module Split Details

### 00-product-scope-and-glossary

| 子文档 | 内容 |
| --- | --- |
| `product-boundary.md` | 产品定位、目标用户、做什么、不做什么、生产闭环边界 |
| `user-workflows.md` | 创作者从项目创建到结果 review 的主路径和异常路径 |
| `glossary.md` | 统一术语、对象名、状态词、文档内中英文用法 |

### 01-local-project-storage

| 子文档 | 内容 |
| --- | --- |
| `directory-layout.md` | studio root、project root、资产、包、运行记录、审计目录 |
| `schema-and-migration.md` | `project.json`、schema version、迁移、备份和回滚 |
| `save-lock-recovery.md` | 原子写、自动保存、项目锁、崩溃恢复 |
| `health-check.md` | 缺文件、断链、digest、外部引用、迁移风险检查 |

### 02-creative-graph-domain-model

| 子文档 | 内容 |
| --- | --- |
| `nodes-and-edges.md` | 节点、边、关系约束、GraphNode 与领域对象边界 |
| `graph-state-machine.md` | Shot、PromptRun、Package、Result 的状态推进 |
| `context-resolution.md` | 选中节点如何展开上下文、裁剪、摘要和冲突 |
| `consistency-rules.md` | 删除、修改、断链、dirty 标记和级联规则 |

### 03-asset-library-and-continuity

| 子文档 | 内容 |
| --- | --- |
| `asset-ingestion-and-indexing.md` | 导入、digest、缩略图、重复检测、managed reference |
| `asset-binding-and-lineage.md` | 资产绑定、来源追溯、PromptRun 和 Package 关联 |
| `continuity-rules.md` | 角色、场景、道具、风格连续性和冲突输出 |
| `deletion-and-recovery.md` | 解绑、回收区、恢复、永久清理和影响分析 |

### 04-script-shot-package-workflow

| 子文档 | 内容 |
| --- | --- |
| `script-to-scene.md` | 剧本导入、拆场、候选确认、Scene 生成 |
| `shot-card-lifecycle.md` | ShotCard 字段、必填校验、状态和 review |
| `package-export.md` | 交接包目录、manifest、提示词、参考图、清单 |
| `result-ingestion-and-review.md` | 外部结果回收、take、绑定、review、重做 |

### 05-instruction-stack-and-skills

| 子文档 | 内容 |
| --- | --- |
| `instruction-compile-order.md` | 编译顺序、优先级、data 包装、上下文摘要 |
| `skill-registry-and-routing.md` | 能力包结构、来源、信任、启用、路由 |
| `output-contracts.md` | task mode、output schema、解析失败和污染防护 |
| `conflict-handling.md` | 锁定规则冲突、上下文缺失、输出不合法、用户确认 |

### 06-ai-runtime-and-provider-adapters

| 子文档 | 内容 |
| --- | --- |
| `runtime-gateway.md` | 运行时接口、任务生命周期、事件流、权限 |
| `image-generation-adapter.md` | 文生图、参考图改图、多图融合、输出落盘 |
| `provider-handoff-adapter.md` | provider profile、手动交接、未来 API 扩展边界 |
| `runtime-errors-and-retry.md` | 错误码、用户提示、重试、取消、超时、幂等 |

### 07-frontend-workbench-experience

| 子文档 | 内容 |
| --- | --- |
| `workspace-layout.md` | Top Bar、资产面板、画布、Inspector、Bottom Bar |
| `canvas-interactions.md` | 节点创建、连线、框选、拖拽、搜索、Frame |
| `inspector-and-task-panels.md` | 属性、关联、连续性、任务、运行记录、审计 |
| `production-feedback.md` | 保存失败、任务失败、缺失资产、上下文过期、导出状态 |

### 08-security-privacy-observability

| 子文档 | 内容 |
| --- | --- |
| `path-guard-and-permissions.md` | 路径守卫、权限、用户选择文件、项目外写入拒绝 |
| `privacy-and-redaction.md` | 本地隐私、凭据隔离、提交版文档脱敏 |
| `audit-events.md` | 审计事件结构、分类、关键操作记录 |
| `observability-and-recovery.md` | 运行记录、错误三层展示、恢复与回滚 |

### 09-delivery-acceptance-and-test-plan

| 子文档 | 内容 |
| --- | --- |
| `delivery-slices.md` | 后续实现切片、依赖和完成定义 |
| `acceptance-matrix.md` | 功能、数据、错误、恢复验收矩阵 |
| `test-strategy.md` | 单元、集成、手动验收、文档检查 |
| `release-readiness.md` | 发布前检查、脱敏、迁移、替换能力和回归 |

## Migration Strategy

1. 创建 00-09 模块目录。
2. 将当前平铺 `00-*.md` 到 `09-*.md` 内容迁移到对应目录 `README.md`。
3. 新增每个模块的专题文档。
4. 更新 `docs/design/details/README.md` 总索引。
5. 更新 `docs/design/README.md` 阅读顺序，链接改为目录入口。
6. 删除原平铺 `00-*.md` 到 `09-*.md`，避免双 truth。
7. 执行结构、敏感词、链接和 harness 验证。

迁移时不要求把每个原 README 改短；先保证内容不丢，再让专题文档承接新增细节。后续可按模块逐步把 README 中过长的细节移动到专题文档。

## Validation Plan

```bash
find docs/design/details -maxdepth 3 -type f -print | sort
rg -n "MVP|纳逗|可灵|Vidu|海螺|Wan|ComfyUI|Stable Diffusion|OpenAI|Codex|Wails|Ant Design|AntV|X6" docs/design
rg -n "/Users/|token|cookie|secret|password|真实凭据|本机绝对路径" docs/design
make harness-check
make harness-review-gate PLAN=.agents/plans/2026-05-14-design-docs-production-refine.md
```

链接检查策略：

- `docs/design/README.md` 中所有 `details/...` 链接都应指向存在文件或目录。
- `docs/design/details/README.md` 中所有模块入口和专题文档链接都应存在。
- 删除旧平铺文件后，不应再有链接指向 `details/00-*.md` 这类旧路径。

## Risks and Mitigations

| 风险 | 缓解 |
| --- | --- |
| 拆太碎导致阅读成本高 | 每个模块只建一层目录，`README.md` 保留总览 |
| 原内容迁移丢失 | 先移动原文为模块 `README.md`，再补子文档 |
| 链接迁移遗漏 | 最后用 `rg "details/[0-9].*\\.md"` 检查旧链接 |
| 外部品牌名回流 | 保持 provider/profile/adapter 抽象并跑敏感词扫描 |
| 后续 issue 难消费 | 每个专题文档都写验收标准和错误语义 |

## User Review Gate

本 spec 获用户确认后，下一步才进入实际文档拆分和细化。实际拆分应继续遵守当前仓库 `AGENTS.md`：复杂任务先维护 plan，文档写回后跑 harness 验证。

