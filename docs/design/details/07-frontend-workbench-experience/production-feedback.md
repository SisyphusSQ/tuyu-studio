# 生产反馈

## 目标与边界

定义工作台在保存、失败、运行队列、上下文过期、缺失资产、包导出、任务失败时的反馈和可行动入口。本文不定义底层错误码全集。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 保存结果 | Top Bar 保存状态、字段标记、Bottom Bar 摘要 |
| 运行队列事件 | 队列计数、进度、取消、详情入口 |
| context_dirty 标记 | 节点徽标、Inspector 风险提示 |
| 缺失资产检查 | 节点角标、Health View 列表、重新定位入口 |
| 交接包导出结果 | PackageNode 状态、打开目录、校验详情 |
| Agent Studio Command | 节点来源标记、候选状态、CommandResult、Agent-visible error |
| 任务失败记录 | 节点状态、PromptRun 错误、可重试性 |

## 核心对象或规则

| 反馈 | 展示规则 |
| --- | --- |
| saved | Top Bar 显示最近保存时间，字段清除 dirty |
| save_failed | 字段保留 dirty，Bottom Bar 显示错误和重试 |
| run_queue | Top Bar 显示队列数量，点击打开队列面板 |
| context_dirty | Shot 或 Prompt 节点显示徽标，任务 tab 解释过期来源 |
| missing_asset | 节点角标和 Health View 同步列出影响对象 |
| package_exported | PackageNode 展示导出版本、状态和打开目录 |
| task_failed | 节点徽标、运行记录、Bottom Bar 三处可达 |
| agent_command | Agent 创建、移动、连线、运行和写回都显示来源、状态和接管入口 |
| candidate_output | Agent 候选节点或连线用 candidate 标记，用户确认后进入正式上下文 |

可行动错误显示：

- 用户提示说明发生了什么和下一步。
- 技术详情包含错误码、对象 ID、路径摘要和可重试性。
- 调试详情链接到项目内错误记录，例如 `{studio_root}/runs/{run_id}/error.json`。
- 不在错误提示中展示机器私有路径、凭据或未脱敏外部地址。

## 状态推进

| 场景 | 状态推进 |
| --- | --- |
| 保存成功 | dirty -> saving -> saved |
| 保存失败 | dirty -> saving -> save_failed -> dirty |
| 任务排队 | task_preview -> queued -> running -> succeeded 或 failed |
| Agent 命令 | command_received -> validated -> applied 或 rejected |
| Agent 候选 | candidate -> accepted 或 rejected |
| 上下文变化 | context_ready -> context_dirty -> context_ready |
| 资产缺失 | asset_ready -> missing_asset -> relocated 或 removed |
| 包导出 | draft -> ready |
| 用户交接 | ready -> handed_off |
| 结果回收 | handed_off -> result_received |
| 包失效 | ready 或 handed_off -> stale；校验失败进入 invalid |

Shot.status 可使用 `package_ready` 表示镜头已达到导出准备态；GenerationPackage.status 使用 `draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid`。

## 错误语义

| 错误 | 用户语义 | 可重试 |
| --- | --- | --- |
| save_failed | 当前更改尚未保存，输入仍保留 | 是 |
| queue_submit_failed | 任务未进入队列 | 是 |
| context_dirty | 上下文已变化，需重新预览 | 修复后可继续 |
| missing_asset | 引用文件缺失或不可读 | 重新定位后可继续 |
| package_validation_failed | 交接包引用或 manifest 不完整 | 修复后重新导出 |
| task_failed | 运行失败，正式对象未被污染 | 视错误码而定 |
| agent_command_failed | Agent 操作失败，画布未被修改 | 修复权限、selection 或参数后可重试 |
| candidate_rejected | 用户拒绝 Agent 候选 | 可从审计记录重新查看 |

## 恢复与重试

- 保存失败可以重试保存，或继续编辑并保留 dirty 状态。
- 任务失败从 PromptRun 详情重试，重试生成新 run，不覆盖旧记录。
- context_dirty 需要重新编译预览，不能直接复用旧输出。
- 缺失资产可重新定位、替换引用、移除绑定或打开影响列表。
- 包导出失败保留旧版本，新导出使用新版本目录。
- 队列卡住时允许取消未开始任务，运行中任务按能力声明决定是否可取消。
- Agent 命令失败可查看 command audit、复制错误摘要、调整 selection 后重试。
- Agent 候选被拒绝后不进入正式上下文，但保留审计证据。

## 验收标准

- 保存成功和失败都能在 Top Bar、字段或 Bottom Bar 得到一致反馈。
- 运行队列可查看排队、运行、完成、失败和取消状态。
- context_dirty 有明确来源和重新预览入口。
- 缺失资产能定位影响对象并提供重新定位或解绑路径。
- 包导出后可查看版本、状态、校验记录和目录入口。
- 任务失败展示用户提示、技术详情、调试记录和可重试性。
- Agent 操作、候选、接管和失败都能在画布、Inspector 和 Bottom Bar 中互相定位。
