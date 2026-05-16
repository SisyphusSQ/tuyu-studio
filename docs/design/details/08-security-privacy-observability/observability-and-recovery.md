# 可观测性与恢复

## 目标与边界

定义 PromptRun 记录、包校验记录、健康报告、三层错误和恢复动作。本文不定义外部能力的具体实现。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 任务提交和事件 | PromptRun 记录、事件流、输出或错误 |
| 包导出和校验 | GenerationPackage 校验记录 |
| 项目健康检查 | Health Report |
| 错误对象 | 用户提示、技术详情、调试记录 |
| 恢复操作 | 恢复记录和审计事件 |

## 核心对象或规则

PromptRun 记录：

- 保存 runId、taskType、sourceObjectIds、contextDigest、compiledInstruction 摘要、schemaId、status、events、outputRef、errorRef。
- PromptRun 不是 GraphNode 或 GraphEdge 端点，只通过领域字段和运行记录引用。
- 正式对象写入必须发生在输出校验通过之后。

包校验记录：

- GenerationPackage 导出前后都要记录 manifest 校验、引用文件清单、相对路径检查、摘要校验和状态。
- GenerationPackage.status 使用 `draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid`。
- 包失效时记录原因，例如引用资产变更、Shot 更新、manifest 不一致。

健康报告：

| 检查 | 内容 |
| --- | --- |
| graph | 断边、非法关系、孤立关键节点 |
| asset | 缺失文件、摘要不一致、未使用大文件 |
| package | manifest 缺项、引用缺失、状态失效 |
| run | 卡住任务、失败未处理、输出未归档 |
| security | 路径越界、外部引用、审计不可写 |
| schema | 版本不匹配、迁移未完成 |

三层错误：

1. 用户提示：一句话说明和下一步。
2. 技术详情：错误码、对象 ID、状态、路径摘要、可重试性。
3. 调试记录：项目内 JSON 记录，供排障和报告导出脱敏使用。

## 状态推进

| 对象 | 状态推进 |
| --- | --- |
| PromptRun | created -> queued -> running -> succeeded / failed / canceled |
| GenerationPackage | draft -> ready -> handed_off -> result_received |
| GenerationPackage 异常 | ready 或 handed_off -> stale；校验失败 -> invalid |
| Health Report | pending -> running -> clean / warning / failed |
| Recovery Action | proposed -> confirmed -> applied -> audited |

## 错误语义

| 错误 | 语义 |
| --- | --- |
| run_output_invalid | 任务输出未通过结构校验 |
| package_reference_missing | 包引用文件不存在或不可读 |
| health_check_failed | 健康检查执行失败 |
| recovery_not_applicable | 当前对象状态不支持该恢复动作 |
| debug_record_unavailable | 调试记录缺失或不可读 |

错误要标注是否可重试、是否污染正式对象、是否需要用户选择。

## 恢复与重试

- PromptRun 失败后可从相同上下文重试，但必须创建新 runId。
- 输出结构校验失败时，不写正式对象，只保留错误记录和原始输出摘要。
- 包引用缺失时，可重新定位资产、重新导出或标记包 invalid。
- 健康检查发现 schema 不匹配时，先备份，再迁移，失败可回滚。
- 审计不可写时，危险恢复动作必须阻断。
- 项目锁异常时，Health View 展示风险，用户确认后接管并写审计。

## 验收标准

- 每次任务都有 PromptRun 记录，能追踪输入摘要、事件、输出和错误。
- 包导出和重新导出都有校验记录和状态变更。
- Project Health 能列出 graph、asset、package、run、security、schema 风险。
- 错误展示满足用户提示、技术详情、调试记录三层。
- 恢复动作有确认、执行结果和审计事件。
