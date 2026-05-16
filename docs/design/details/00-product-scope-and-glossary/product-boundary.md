# 产品边界

## 目标与边界

图屿是 Canvas-first 私有 AI 视频创作 Studio。产品核心不是复刻某个外部创作平台，也不是只做生成包交接，而是把人和 Agent 对同一张 Project Canvas 的创作操作统一到本地项目状态机中：剧本、角色、场景、参考资产、分镜、提示词、生成任务、结果评审、交接包和审计事件都能被组织、运行、回收、追溯和恢复。

产品边界以“画布上的创作操作是否能落到本地项目、结构化对象、资产 lineage、Run/Event/Audit 和 Review 证据链”为判断依据。任何能力如果不能强化这个闭环，默认不进入核心产品范围。

## 输入 / 输出

| 类别 | 输入 | 输出 |
| --- | --- | --- |
| 项目资料 | 剧本、角色说明、场景说明、道具说明、风格规则、用户画布操作 | 本地 Project、领域对象、Project Canvas 节点和 Frame |
| Agent 操作 | Skill、CLI/MCP 风格命令、选中上下文、Blueprint 输入 | Studio Command、Run/Event/Audit、可见画布变更 |
| 参考资产 | 图片、视频、音频、文本、手工说明 | 受控 Asset 记录、digest、引用关系、连续性锁定 |
| 镜头生产 | Shot、上下文、Prompt、ProviderProfile、ProviderMode | 节点运行、结构化输出、PromptObject、Result placeholder |
| 结果回收 | 外部生成文件、用户备注、评审结论 | Take、Review Status、修订任务和审计记录 |
| 交接兜底 | 需要离线或手工外部生成的 Shot / Prompt | GenerationPackage、manifest、上传清单和 handoff 状态 |

## 核心对象或规则

产品边界围绕 Project、Project Canvas、Shared Canvas Core、Studio Command、Asset、Shot、PromptRun、Run/Event/Audit、GenerationPackage、Take 和 Review Status 建立。核心规则是所有生产产物必须能回到本地项目、画布关系、镜头上下文和审计记录；任何外部能力都只通过 ProviderMode、provider/profile/adapter 或 Agent Gateway 抽象进入设计。

## 用户路径

| 阶段 | 用户动作 | 产品责任 |
| --- | --- | --- |
| 建项目 | 创建本地 Project，进入 Project Canvas，选择暖浅或暗色画布 | 生成可健康检查的项目结构和默认画布布局 |
| 布画布 | 插入 Blueprint、导入角色/场景/参考资产、创建节点和连线 | 建立 Project Canvas 关系、Frame 和连续性规则 |
| 人工创作 | 用户直接编辑节点、Inspector、资产绑定和 Shot Card | 通过 Studio Command 写入结构化对象和审计 |
| Agent 共创 | 外部 Agent 读取 Skill 并调用 CLI/MCP 风格命令操作画布 | 所有变更在同一画布上可见、可撤销、可审计 |
| 发起运行 | 选择节点或 Frame，选择 internal_provider 或 external_agent | 写入 Run/Event/Audit，结果回流到 Asset/Take/Review |
| 做交接 | 必要时导出 GenerationPackage 并手动外部生成 | 保证 manifest、资产和清单完整，作为 fallback 而非主路径 |
| 收结果 | 导入 Take，评审并决定通过或修订 | 绑定结果、保留历史、驱动下一轮修改 |

## 包含能力

| 能力 | 产品内含义 | 完成信号 |
| --- | --- | --- |
| 本地项目 | 项目目录包含配置、画布、图谱、领域对象、资产、运行记录、交接包、审计和备份 | 复制项目后仍可健康检查 |
| Project Canvas | 用无限画布、节点、边、Frame 和 Blueprint 表达剧本、角色、场景、道具、镜头、提示词、任务和结果关系 | 任意结果可追溯到源 Shot、上下文和来源命令 |
| Shared Canvas Core | UI 和 Agent 入口共用的命令校验、状态推进、事件写入和审计边界 | 人和 Agent 不能绕过它直接改项目文件 |
| Agent Skill / CLI | 外部 Agent 通过受控 Skill 和 CLI/MCP 风格命令读取上下文、创建节点、连线、运行和写回结果 | Agent 操作必须可见、可审计、可被人接管 |
| 资产连续性 | 角色、场景、道具、风格的锁定规则和参考图能影响后续 Shot | 上游规则变更后关联 Shot 被标记 dirty |
| 指令栈 | 按项目原则、领域上下文、任务模板、输出契约顺序编译任务输入 | 运行前可预览来源、优先级和冲突 |
| ProviderMode | 将 internal_provider 与 external_agent 都归一为 Run/Event/Audit | 具体模型、外部工具或扣费系统不进入核心领域模型 |
| 交接包 | 为手动外部生成准备结构化包、manifest、提示词、参考资产和清单 | 包可离线检查且引用文件完整，是 handoff/fallback |
| 结果评审 | 外部结果回收为 Take，支持评审、对比、退回和保留历史 | 同一 Shot 可挂多个 Take 并记录结论 |

## 排除能力

| 排除项 | 不做内容 | 边界说明 |
| --- | --- | --- |
| SaaS 协作 | 账号体系、团队权限、实时多人编辑、云同步、云端项目托管 | 后续可导出文件交换，但不成为实时协作平台 |
| 专业剪辑时间线 | 多轨剪辑、音频混音、调色、字幕工程、成片渲染 | 只管理镜头级前期资料和外部结果 |
| 静默外部提交 | 后台替用户上传完整项目、自动提交外部任务、隐藏执行过程 | 外部生成默认由用户手动完成或显式确认 |
| 通用文件管理器 | 管理任意本地文件夹、全盘搜索、替代系统文件浏览器 | 只管理项目内资产和用户显式导入的引用 |
| 黑盒 AI 代理 | 无输入说明、无输出契约、无运行记录、无可见错误语义的自动操作 | 所有任务必须可预览、可审计、可重试或可取消 |
| 真实扣费系统 | 会员、积分、余额、账单、自动扣费和成本结算 | 当前只保留 provider cost / credit placeholder，优先级最低 |

## 决策规则

1. 能直接提升 Project Canvas 创作完整性、连续性、可运行性、可交接性或可追溯性的能力，进入范围。
2. 只提供泛文件整理、泛聊天、泛素材浏览的能力，不进入范围。
3. 需要外部服务或外部 Agent 执行的能力，必须通过 ProviderMode、ProviderProfile、HandoffAdapter、ImageGenerationAdapter 或 Agent Gateway 抽象表达，不在文档中绑定具体厂商。
4. 会改变用户资产、提交外部任务或覆盖历史结果的动作，必须有显式用户确认和审计记录。
5. 无法被健康检查、迁移或审计覆盖的持久化数据，不应成为生产路径依赖。

## 错误语义

| 错误 | 用户可见语义 | 系统行为 |
| --- | --- | --- |
| 越界能力 | 当前操作不属于项目生产闭环 | 拒绝执行并指向可支持的替代路径 |
| 隐式外发 | 操作需要离开本地项目或提交外部任务 | 阻断并要求用户显式确认 |
| 不可追溯产物 | 产物缺少 Shot、PromptRun、Run/Event 或 Package 来源 | 标记为未绑定结果，不进入正式评审 |
| Agent 越权 | Agent 命令绕过 Studio Command、访问越界路径或试图直接写文件 | 拒绝执行，写入 Agent-visible error event |
| 抽象泄漏 | 文档或配置出现具体外部目标名 | 阻断设计评审，改为 provider/profile/adapter 表达 |

## 恢复与重试

- 范围误判导致创建了不适用对象时，允许转为 Note、从画布移除或保留为未接受候选，不自动删除源资产。
- Agent 输出候选节点或连线但用户未接受时，保留为 candidate，不覆盖已锁定对象。
- 外部交接前中断时，保留 GenerationPackage 和 audit 事件；对应 Shot 停留在 `package_ready`，GenerationPackage 停留在 `ready`。
- 结果无法绑定时，文件保留在待处理区，用户可重新选择 Shot 或 Package。

## 验收标准

- 给定一个功能描述，产品评审可以依据“是否服务 Project Canvas 上的人机共创闭环”判断 in scope 或 out of scope。
- 所有外部执行相关设计只出现 provider/profile/adapter 抽象，不出现具体外部厂商或产品名。
- 人工画布操作、Agent 操作、交接、回收、评审路径都能保留本地证据，不依赖静默外部提交。
- 被判定 out of scope 的需求有明确替代处理：拒绝、导出、转 Note 或进入后续独立产品讨论。
