# 交付切片

## 目标与边界

定义实现切片、依赖、完成定义、文档同步和 issue 编写要求。本文用于把设计落到可执行任务，不替代各模块的详细设计。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 设计文档和模块 README | 切片列表和依赖顺序 |
| 领域对象、状态、错误语义 | issue 验收口径 |
| 安全、隐私、可观测要求 | 每个切片的横向完成条件 |
| 验证入口 | 自动、集成、手动或文档检查要求 |

## 核心对象或规则

切片顺序：

| 顺序 | 切片 | 依赖 | 完成定义 |
| --- | --- | --- | --- |
| 0 | Wails 与工作台骨架 | 无 | Wails app、Go service 分层、Vue App Shell、TypeScript binding、基础事件流可运行 |
| 1 | 本地项目与存储 | 0 | 创建、打开、保存、锁、健康检查可落盘恢复 |
| 2 | Project Canvas 基础 | 1 | 无限画布、节点、边、Frame、Blueprint、Inspector、自动保存闭环 |
| 3 | 资产库与绑定 | 1-2 | 导入、摘要、缩略图、绑定、缺失修复 |
| 4 | 剧本与分镜 | 1-3 | Script、Scene、Shot、手动拆分、字段校验 |
| 5 | 指令栈、Blueprint 与 Agent Skill/CLI | 2-4 | 上下文解析、能力包、Blueprint、Agent 命令、输出契约、风险预览 |
| 6 | ProviderMode 任务运行 | 5 | internal_provider / external_agent、PromptRun、结构化输出、失败不污染正式对象 |
| 7 | 生成交接包 fallback | 3-6 | manifest、引用清单、校验、handoff 状态 |
| 8 | 结果回收与 review | 6-7 | VideoResult、take、状态、评审、修改任务 |
| 9 | 安全与可观测 | 横向 | 路径守卫、审计、健康、恢复、脱敏报告覆盖 |
| 10 | 迁移与稳定性 | 1-9 | schema 迁移、备份、跨目录打开、回滚 |

切片 0 的技术范围：

- 初始化 Wails 桌面应用和 Vite + Vue 3 + TypeScript 前端。
- 建立 Go service 分层骨架，但不把领域逻辑写进 Wails app struct。
- 接入 Ant Design Vue 的基础工作台布局和 AntV 的最小画布 smoke。
- 工作台第一屏必须是 Project Canvas，可在 Dark 与暖浅 Light Canvas 之间切换，点阵网格亮度可调。
- 建立 Wails 生成的 TypeScript binding、错误 DTO 和基础事件订阅。
- 提供一个不写真实项目数据的启动 smoke，证明桌面壳、前端、Go 方法调用和事件推送打通。

完成定义：

- 用户路径能通过真实界面或命令入口完成。
- 数据落盘后重启仍可恢复。
- 失败场景有错误码、用户提示和恢复入口。
- 关键状态变更有审计或运行记录。
- 文件写入经过路径守卫，提交版材料使用脱敏摘要。
- 相关设计文档、测试说明、issue 状态同步。

issue 编写 guidance：

- 正文必须包含 Goal、Scope、Out of Scope、Implementation Scope、Acceptance Criteria、Verification Commands、Data / Schema Impact、Security and Recovery Notes。
- 从文档创建 issue 时先通读 README 和子文档，不只摘标题。
- 不确定归属的内容先标为问题或拆分建议，不自行省略。

## 状态推进

| 状态 | 推进规则 |
| --- | --- |
| planned | 切片边界和依赖已明确 |
| ready | 前置切片验收完成，文档口径一致 |
| in_progress | 实现、测试、文档同步并行推进 |
| blocked | 依赖、设计冲突或验证入口缺失 |
| review | 功能、错误、恢复、审计验收完成待确认 |
| accepted | 验证证据和文档同步完成 |

## 错误语义

| 错误 | 语义 |
| --- | --- |
| dependency_unmet | 前置切片未达到完成定义 |
| acceptance_gap | 验收项缺少测试或手动证明 |
| docs_out_of_sync | 实现与设计文档或 issue 不一致 |
| security_gate_missing | 文件、安全或脱敏要求缺失 |
| recovery_path_missing | 失败场景没有恢复路径 |

## 恢复与重试

- dependency_unmet 时回到前置切片补齐验收，不跳过。
- acceptance_gap 需要补测试、手动脚本或文档检查。
- docs_out_of_sync 需要同步设计、issue 和验证摘要。
- security_gate_missing 阻断切片验收，先补路径守卫或脱敏。
- blocked 状态要记录阻塞原因、依赖对象和下一步处理人。

## 验收标准

- 每个切片都有明确依赖、产出和完成定义。
- 切片验收覆盖功能、数据、错误、恢复、审计和安全。
- issue 正文能完整表达实现边界和验收，不丢关键错误语义。
- 文档同步是完成定义的一部分，而不是收尾可选项。
