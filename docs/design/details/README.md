# 细节设计索引

## 覆盖目标

`docs/design/details` 承接生产级设计细节。当前结构按 00-09 模块拆成子目录：每个模块的 `README.md` 保留模块总览，子文档承接可执行规则、对象字段、状态机、异常语义、恢复路径和验收标准。

每个模块都要回答：

- 输入从哪里来，用户或系统如何触发。
- 输出是什么，落到哪个项目文件、领域对象或状态。
- 状态如何推进，失败如何表达，哪些操作会被阻断。
- 运行记录如何追溯，哪些动作可重试、恢复、撤销或重新生成。
- 交付后如何验收，不靠口头判断。

## 模块地图

| 模块 | 主题 | 关键问题 |
| --- | --- | --- |
| `00-product-scope-and-glossary/README.md` | 产品边界与术语 | 图屿是什么、不是什么、面向谁、闭环到哪里 |
| `01-local-project-storage/README.md` | 本地项目存储 | 项目怎么落盘、怎么迁移、怎么恢复、怎么防损坏 |
| `02-creative-graph-domain-model/README.md` | Project Canvas / Creative Graph | 节点、边、Frame、Blueprint 和 Agent 操作如何进入同一张画布 |
| `03-asset-library-and-continuity/README.md` | 资产与连续性 | 人物、场景、道具、参考图如何复用、锁定和追溯 |
| `04-script-shot-package-workflow/README.md` | 剧本到生成包 | 剧本、场次、镜头、交接包和结果如何形成闭环 |
| `05-instruction-stack-and-skills/README.md` | 指令栈、Blueprint 与 Skill/CLI | 如何编译上下文，如何让 Agent 通过 Skill/CLI 操作画布 |
| `06-ai-runtime-and-provider-adapters/README.md` | 运行时、ProviderMode 与适配器 | internal_provider 与 external_agent 如何统一为 Run/Event/Audit |
| `07-frontend-workbench-experience/README.md` | Canvas-first 工作台体验 | 无限画布、左右面板、Dark/暖浅 Light Canvas、运行反馈和人机共创 |
| `08-security-privacy-observability/README.md` | 安全与观测 | 如何控制本地隐私、权限、审计、脱敏和错误定位 |
| `09-delivery-acceptance-and-test-plan/README.md` | 交付与测试 | 如何验收画布主路径、Agent 可见操作、失败恢复和 handoff fallback |

## 子文档地图

| 模块 | 子文档 | 负责内容 |
| --- | --- | --- |
| 00 产品边界与术语 | `product-boundary.md` | 产品定位、目标用户、能力边界、非目标和功能归属判断 |
| 00 产品边界与术语 | `user-workflows.md` | 新项目、资产导入、提示词运行、交接包导出、结果回收主路径 |
| 00 产品边界与术语 | `glossary.md` | 项目、图谱、镜头、运行、交接、结果、适配器术语 |
| 01 本地项目存储 | `directory-layout.md` | Studio root、project root、资产、包、运行记录、审计目录 |
| 01 本地项目存储 | `schema-and-migration.md` | 单一项目清单、schema version、迁移、备份和回滚 |
| 01 本地项目存储 | `save-lock-recovery.md` | 原子写、自动保存、项目锁、崩溃恢复 |
| 01 本地项目存储 | `health-check.md` | 缺文件、断链、digest、外部引用、迁移风险检查 |
| 02 Project Canvas / Creative Graph | `nodes-and-edges.md` | 节点、边、Frame、Blueprint、领域对象引用和合法关系 |
| 02 Creative Graph | `graph-state-machine.md` | 图谱对象状态、阻断条件、删除和版本推进 |
| 02 Creative Graph | `context-resolution.md` | 镜头上下文解析、资产选择、连续性规则和摘要 |
| 02 Creative Graph | `consistency-rules.md` | refId、边类型、跨项目拒绝、孤儿对象和修复策略 |
| 03 资产与连续性 | `asset-ingestion-and-indexing.md` | 导入、digest、去重、缩略图、来源和索引 |
| 03 资产与连续性 | `asset-binding-and-lineage.md` | 资产绑定、派生关系、生成来源和变更影响 |
| 03 资产与连续性 | `continuity-rules.md` | 人物、场景、道具、风格的连续性锁定和冲突检查 |
| 03 资产与连续性 | `deletion-and-recovery.md` | 删除、隔离、恢复、缺失资产和引用修复 |
| 04 剧本到生成包 | `script-to-scene.md` | 剧本导入、拆场、候选确认和 Scene 生成 |
| 04 剧本到生成包 | `shot-card-lifecycle.md` | ShotCard 字段、必填校验、状态推进和 review |
| 04 剧本到生成包 | `package-export.md` | 交接包目录、manifest、提示词、参考图和清单 |
| 04 剧本到生成包 | `result-ingestion-and-review.md` | 外部结果回收、take、绑定、review 和重做 |
| 05 指令栈与能力包 | `instruction-compile-order.md` | 项目原则、上下文、能力包、模板和输出契约编译顺序 |
| 05 指令栈、Blueprint 与 Skill/CLI | `skill-registry-and-routing.md` | 能力包目录、Blueprint、Agent Skill/CLI 路由、版本和可用性检查 |
| 05 指令栈与能力包 | `output-contracts.md` | 结构化输出 schema、校验、写入策略和失败处理 |
| 05 指令栈与能力包 | `conflict-handling.md` | 指令冲突、素材包装、阻断规则和用户确认 |
| 06 运行时、ProviderMode 与适配器 | `runtime-gateway.md` | 任务创建、ProviderMode、队列、取消、状态、运行记录和可观测 |
| 06 运行时与适配器 | `image-generation-adapter.md` | 图像任务输入、输出资产、错误映射和重试 |
| 06 运行时与适配器 | `provider-handoff-adapter.md` | 视频生成交接 profile、包格式、上传清单和结果绑定 |
| 06 运行时与适配器 | `provider-configuration-and-credentials.md` | provider profile 复用、项目引用、credentialRef 和系统密钥存储 |
| 06 运行时与适配器 | `runtime-errors-and-retry.md` | 错误分类、重试策略、恢复入口和用户提示 |
| 07 Canvas-first 工作台体验 | `workspace-layout.md` | 左侧入口、无限画布、资产栏、Inspector、任务面板和状态栏 |
| 07 工作台体验 | `canvas-interactions.md` | 节点创建、连线、布局、选择、快捷操作和冲突提示 |
| 07 工作台体验 | `inspector-and-task-panels.md` | 领域编辑、校验、任务队列、输出预览和审计入口 |
| 07 工作台体验 | `production-feedback.md` | 保存、运行、导出、回收、错误、恢复和无障碍反馈 |
| 08 安全与观测 | `path-guard-and-permissions.md` | 路径守卫、权限、导入导出授权和拒绝语义 |
| 08 安全与观测 | `privacy-and-redaction.md` | 数据包装、脱敏、报告、日志边界和用户确认 |
| 08 安全与观测 | `audit-events.md` | 审计事件、字段、关联对象、保留和查询 |
| 08 安全与观测 | `observability-and-recovery.md` | 健康检查、运行诊断、恢复报告和支持包 |
| 09 交付与测试 | `delivery-slices.md` | 实现切片、依赖、完成定义和文档同步 |
| 09 交付与测试 | `acceptance-matrix.md` | 功能、数据、错误、恢复、审计、安全验收矩阵 |
| 09 交付与测试 | `test-strategy.md` | 单元、集成、手动验收、文档检查和回归命令 |
| 09 交付与测试 | `release-readiness.md` | 发布门禁、迁移、脱敏、可移植、恢复和适配器替换 |

## 统一设计口径

| 维度 | 设计要求 |
| --- | --- |
| 完整性 | 每条主工作流都必须有进入、处理、输出、验收和恢复路径 |
| 可靠性 | 保存、导入、任务运行、导出和结果绑定不能依赖单点隐式状态 |
| 可追溯 | 用户输入、上下文来源、能力包、输出契约、产物路径和错误都能回查 |
| 可替换 | 外部运行时和生成平台通过 adapter/profile 表达，不进入核心领域模型 |
| 可测试 | 关键能力都有文件级、领域级、交互级和端到端验收入口 |
| 可维护 | 目录、命名、schema 版本和迁移策略在首个生产闭环中就要稳定 |

## 跨文档不变量

- `Project` 是所有对象的归属根，跨项目引用默认拒绝。
- `Asset` 必须有稳定 `id`、项目内相对路径、来源、digest 和绑定关系。
- `ProjectCanvas` 保存视口、节点、边、Frame、Blueprint 实例和来源事件，是 Studio 的第一工作区。
- `GraphNode.refId` 指向领域对象，节点不是领域数据的唯一真相。
- UI、CLI、MCP 风格入口和外部 Agent 修改项目时都必须经过 Studio Command，不允许直接写项目文件。
- `PromptRun` 是 AI 任务追溯记录，不作为 `GraphNode` 或 `GraphEdge` 的端点；图谱通过领域对象或运行记录引用它。
- `GenerationPackage.status` 使用 `draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid` 作为规范状态。
- `Shot.status=package_ready` 表示对应 `GenerationPackage` 已 ready 且 manifest 校验通过；用户完成交接后才推进到 `submitted`，结果回收后才推进到 `generated`。
- `refines` 只表示同类对象版本链路，例如 prompt 到 prompt、shot 到 shot、package 到 package、video result 到 video result。
- `GenerationPackage` 是交接产物，不反向覆盖源镜头、提示词或连续性规则。
- `VideoResult` 是回收结果和 review 状态，不等同于最终成片。

## 验证命令

```bash
find docs/design/details -maxdepth 3 -type f -print | sort
find docs/design/details -maxdepth 1 -type f -name '[0-9][0-9]-*.md' -print
rg -n "details/[0-9][0-9]-[^/]+\\.md" docs/design
make harness-check
make harness-review-gate PLAN=.agents/plans/2026-05-14-design-docs-production-refine.md
```

预期：

- `find ... -maxdepth 3` 能看到本索引、10 个模块 `README.md` 和 39 个子文档。
- `find ... -maxdepth 1 -name '[0-9][0-9]-*.md'` 无输出。
- 旧 flat 文档链接检查无输出。
- 外部品牌、降级口径和 harness 检查通过；敏感词表以对应计划文档中的验证命令为准，避免正式设计稿保存品牌黑名单本身。
