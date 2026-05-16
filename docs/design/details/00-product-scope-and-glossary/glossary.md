# 术语表

## 目标与边界

本文统一设计文档、配置和用户界面中的核心术语。正式文档使用稳定抽象，不绑定具体外部厂商、产品或运行时名称。

## 术语规则

| 规则 | 说明 |
| --- | --- |
| 英文对象名 | 持久化 schema、状态机和接口说明使用英文名 |
| 中文解释 | 用户向说明和评审文档可配中文解释 |
| 抽象命名 | 外部能力统一写作 provider、profile、adapter、gateway |
| 禁止绑定 | 正式设计文档不直接写具体外部厂商或产品名 |
| 单复数 | 对象类型用单数，如 `Shot`；集合目录可用复数，如 `shots/` |

## 核心术语

| 术语 | 定义 | 边界 |
| --- | --- | --- |
| Project | 一个创作项目的持久化根，包含画布、图谱、领域对象、资产、提示词、交接包、结果和审计 | 不等同于单个视频文件或单次任务 |
| Studio Root | 本地工作台根目录，保存全局配置、项目列表、能力包和 provider profile | 不应保存具体项目的完整生产资产 |
| Project Canvas | 项目的第一工作台，保存无限画布视口、节点、边、Frame、Blueprint、运行状态和渲染摘要 | 不是临时白板，必须回到本地 Project 和领域对象 |
| Shared Canvas Core | UI 和 Agent 共同调用的项目状态机，负责命令校验、画布更新、资产绑定、运行事件和审计 | 不等同于前端 store 或 CLI 脚本 |
| Studio Command | 所有 UI/CLI/MCP 入口修改项目的唯一命令边界，例如 create_node、connect_nodes、start_run | 不允许绕过它直接写项目文件 |
| Canvas Blueprint | 可插入画布的一组节点、边、Frame、输入槽、输出契约和恢复语义 | 不只是 UI 模板 |
| ProviderMode | 运行来源模式，当前为 internal_provider 或 external_agent | 不等同于具体外部 provider 名称 |
| External Agent | 通过 Skill/CLI/MCP 风格接口操作 Project Canvas 的外部 Agent | 不能直接读取任意路径或绕过审计 |
| Warm Light Canvas | 偏暖、低眩光的浅色画布模式，不是纯白反色主题 | 必须单独设计状态色、连线和节点层级 |
| Creative Graph | Project Canvas 内的结构化关系图，用节点和边表达创作对象、上下文、任务和结果关系 | 不是普通白板，也不是剪辑时间线 |
| Node | Project Canvas / Creative Graph 上的可视化对象，通常通过 `refId` 指向领域对象 | 节点删除不默认删除领域对象 |
| Edge | 节点之间的关系，表达使用、归属、生成、优化、打包、结果来源或依赖 | 边必须符合关系矩阵 |
| Shot | 镜头级生产单位，包含画面、动作、运镜、台词、时长、上下文和状态 | 是交接包和结果绑定的最小生产单元 |
| PromptRun | 一次提示词生成、优化或结构化处理的运行记录 | 必须保存输入摘要、状态、输出、错误和审计信息 |
| GenerationPackage | 面向手动外部生成的交接包，包含 manifest、提示词、连续性说明、引用资产和清单 | 不代表外部任务已经提交 |
| Take | 同一 Shot 的一次回收结果版本 | 多个 Take 可并存，不互相覆盖 |
| Review Status | 用户对 Take 或 Result 的评审状态，如待看、通过、需修改、废弃 | 状态变更必须保留评审记录 |
| RuntimeGateway | 应用内部访问可执行能力的统一入口，负责排队、状态、ProviderMode、取消和错误归一 | 不暴露具体外部实现细节 |
| ProviderProfile | 某类外部生成目标的能力声明、格式偏好、限制和交接规则 | 只描述能力与约束，不保存凭据 |
| HandoffAdapter | 将 Shot、Prompt 和资产编译为交接包的适配器 | 默认服务手动交接，不静默提交 |
| ImageGenerationAdapter | 将图像相关任务输入输出规范化的适配器 | 通过抽象能力描述，不绑定具体产品 |

## 状态词

| 状态词 | 使用对象 | 含义 |
| --- | --- | --- |
| `draft` | Shot、Prompt | 草稿，允许编辑，不可正式交接 |
| `context_ready` | Shot | 必要上下文通过检查 |
| `prompt_ready` | Shot、Prompt | 提示词已确认，可进入 Package |
| `package_ready` | Shot | 对应 GenerationPackage 已达到 `ready`，Shot 可进入手动交接 |
| `submitted` | Shot | 用户已手动标记外部交接 |
| `generated` | Shot、Result | 已回收结果文件 |
| `approved` | Take | 评审通过 |
| `needs_revision` | Take、Shot | 需要修改并重新进入生产链路 |
| `context_dirty` | Shot | 上游上下文变化，镜头生产资料需要重新确认 |

GenerationPackage 使用独立状态集合：`draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid`。`package_ready` 只描述 Shot 已有可交接包，不作为 GenerationPackage 状态。

## 命名示例

| 推荐 | 不推荐 | 原因 |
| --- | --- | --- |
| `ProviderProfile` | 具体外部平台名 profile | 保持可替换 |
| `HandoffAdapter` | 某平台提交器 | 默认不静默提交 |
| `RuntimeGateway` | 某运行时客户端 | 隔离执行细节 |
| `ImageGenerationAdapter` | 某图像产品 SDK | 只描述能力和契约 |

## 错误语义

| 错误 | 处理 |
| --- | --- |
| 术语混用 | 在评审中阻断，要求改为表内标准术语 |
| 外部名称泄漏 | 替换为 provider/profile/adapter 抽象 |
| 状态含义不清 | 回到状态机文档补充进入条件和退出条件 |
| Node 与领域对象混同 | 明确 UI 节点和持久化对象的 ID、生命周期差异 |

## 验收标准

- 新增设计文档能引用本术语表解释核心对象，且不创造同义状态词。
- 持久化对象、接口和状态机使用英文正式名。
- 外部能力设计只使用 RuntimeGateway、ProviderProfile、HandoffAdapter、ImageGenerationAdapter 等抽象。
- 文档中出现新术语时，能判断是否应加入本表或改用既有术语。
