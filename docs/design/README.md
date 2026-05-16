# 图屿设计文档入口

## 文档定位

`docs/design` 是 `tuyu-studio` 的仓库内产品与技术设计真相。本目录基于项目控制面中的三篇源材料整理，但正式设计稿不直接复刻源材料中的外部供应商、平台或实现品牌名称，而是沉淀为可生产落地的产品边界、领域模型、运行机制和验收口径。

本目录遵循三个原则：

1. **产品名归产品，供应商归适配器**：设计稿只固定图屿自身的能力、边界和数据契约，外部生成平台、模型服务或桌面技术栈均通过抽象适配层描述。
2. **画布闭环优先**：任何能力都要写清输入、输出、状态、失败语义、恢复方式、追溯记录和验收标准，并能回到 Project Canvas 上的可见操作。
3. **本地优先且可迁移**：项目数据、资产、运行记录、生成包和结果绑定都以本地项目目录为主，并保留后续替换运行时或 provider 的空间。

## 阅读顺序

| 顺序 | 文档 | 作用 |
| --- | --- | --- |
| 1 | `architecture/redacted-production-architecture.md` | 脱敏后的系统总览、边界和主链路 |
| 2 | `architecture/technology-stack.md` | 当前落地技术栈、分层职责、契约和验证要求 |
| 3 | `details/README.md` | 细节文档索引和覆盖矩阵 |
| 4 | `details/00-product-scope-and-glossary/README.md` | 产品边界、用户、术语和完整交付标准 |
| 5 | `details/01-local-project-storage/README.md` | 本地项目、文件结构、保存、迁移和恢复 |
| 6 | `details/02-creative-graph-domain-model/README.md` | Project Canvas / Creative Graph 节点、边、Frame、Blueprint、状态和一致性 |
| 7 | `details/03-asset-library-and-continuity/README.md` | 资产库、人物/场景/道具连续性和来源追溯 |
| 8 | `details/04-script-shot-package-workflow/README.md` | 剧本、场次、分镜、生成包和结果回收 |
| 9 | `details/05-instruction-stack-and-skills/README.md` | 指令栈、Blueprint、Skill/CLI 和冲突处理 |
| 10 | `details/06-ai-runtime-and-provider-adapters/README.md` | AI 任务运行时、ProviderMode、图像生成适配器和外部 Agent 入口 |
| 11 | `details/07-frontend-workbench-experience/README.md` | Canvas-first 桌面工作台、画布交互、面板和生产操作体验 |
| 12 | `details/08-security-privacy-observability/README.md` | 安全、隐私、脱敏、审计和可观测性 |
| 13 | `details/09-delivery-acceptance-and-test-plan/README.md` | 交付切片、验收矩阵和测试计划 |

## 设计稿边界

| 类别 | 当前固定 | 当前不固定 |
| --- | --- | --- |
| 产品定位 | Canvas-first 私有 AI 视频创作 Studio，本地项目状态机承载人和 Agent 的共同创作 | 某个外部生成平台、会员、积分或云端协作体系 |
| 主工作区 | Project Canvas：无限画布、节点、边、Frame、Blueprint、运行状态、结果评审 | 传统多页面文档工作台、单一图片白板或专业剪辑时间线 |
| 数据存储 | 本地项目目录、结构化 JSON、Markdown 文档、资产文件、运行记录、画布布局和审计事件 | 远端数据库、隐式云同步、不可迁移黑盒项目 |
| AI 编排 | Studio Command、ProviderMode、Run/Event/Audit、Skill/CLI/Agent Gateway | 单一模型、单一外部服务或绕过 Studio Core 的后台代理 |
| 视频生成 | 节点运行、结果回流、Review；Package Export 作为 handoff/fallback | 默认静默提交外部视频 API 或真实积分扣费系统 |
| 安全边界 | 本地白名单、路径守卫、用户显式授权、Agent 操作审计、凭据隔离 | 隐式上传、跨项目任意读写、无痕运行 |

## 当前落地技术栈登记

当前实现方向采用 Wails + Go 后端 + Vue 3 + TypeScript + Ant Design Vue + AntV。该选型只固定落地方式，不改变上面的产品领域边界：

- Wails 负责桌面壳、本地应用 bridge、Go 方法暴露给前端和桌面打包。
- Go 后端负责项目存储、领域服务、路径守卫、任务运行、交接包、审计和恢复。
- Vue 3 + TypeScript 负责桌面工作台前端，承接页面、面板、表单、画布状态和任务反馈。
- Ant Design Vue 负责工作台控件体系，例如布局、表单、表格、抽屉、弹窗、通知、上传和进度。
- AntV 负责图谱和图表，其中 Creative Graph 画布优先评估 G6，生产状态和覆盖统计使用 G2。

详细职责、目录建议、测试要求和前端原型关系见 `architecture/technology-stack.md`。

## 文档质量标准

正式设计稿中不使用降级口径。每个细节设计默认按生产可用目标书写：

- 真实用户工作流要能闭环。
- 数据对象要有稳定字段、状态和一致性规则。
- 错误要有明确语义、用户提示和恢复路径。
- 运行过程要能审计、复现和追踪到输入来源。
- 外部依赖要通过适配器抽象，不进入产品主叙事。
- 验收标准要能被开发、测试和后续 issue 直接消费。
