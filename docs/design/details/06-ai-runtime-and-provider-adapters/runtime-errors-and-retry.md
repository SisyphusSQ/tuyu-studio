# Runtime Errors and Retry

## 目标与边界

定义运行时错误码、用户提示、可重试性、取消、超时、幂等键和 attempt records。本文覆盖运行任务的错误治理，不定义具体 provider 能力。

## 输入 / 输出

输入：

- RuntimeGateway、Adapter 和输出校验阶段产生的错误。
- 用户取消、超时配置和重试命令。

输出：

- 标准错误对象。
- attempt record。
- 用户可读提示和下一步操作。

## 核心对象或规则

错误对象：

```ts
interface RuntimeError {
  code: string
  message: string
  retryable: boolean
  userAction?: string
  attempt: number
  occurredAt: string
}
```

错误码：

- `runtime_not_configured`
- `runtime_start_failed`
- `task_timeout`
- `permission_denied`
- `path_rejected`
- `asset_missing`
- `output_schema_invalid`
- `output_missing`
- `network_required`
- `provider_profile_missing`
- `provider_mode_unsupported`
- `agent_gateway_unavailable`
- `cost_placeholder_only`
- `user_cancelled`
- `unknown`

idempotency keys：

| 场景 | 幂等键 | 行为 |
| --- | --- | --- |
| 结构化文本任务重试 | `runId + contextDigest` | 新 attempt，保留旧输出 |
| 图像任务重试 | `runId + attempt` | 新输出路径 |
| 交接包重新导出 | `shotId + profileId + packageVersion` | 新目录 |
| 结果回收重复文件 | `digest + shotId` | 提示复用或新 take |

attempt records：

- 每次启动、重试、取消和超时都记录 attempt。
- attempt 记录 request 摘要、开始时间、结束时间、错误和输出路径。
- failed attempt 不被删除。

## 状态推进

```text
attempt_created -> running -> completed
running -> failed_retryable -> retrying -> attempt_created
running -> failed_final
running -> cancelling -> cancelled
running -> timed_out
```

- `failed_retryable`：用户可直接重试或先修复配置。
- `failed_final`：必须修改输入、配置或规则。

## 错误语义

| 错误 | 用户提示 | 可重试 |
| --- | --- | --- |
| `runtime_not_configured` | 运行时未配置 | 配置后可重试 |
| `runtime_start_failed` | 运行时启动失败 | 可重试 |
| `task_timeout` | 任务超时 | 可重试 |
| `permission_denied` | 项目目录不可写或资产不可读 | 修复后可重试 |
| `path_rejected` | 输入或输出路径越界 | 修正后可重试 |
| `asset_missing` | 参考资产缺失 | 补齐后可重试 |
| `output_schema_invalid` | 输出不符合任务契约 | 可重试或手动提取 |
| `output_missing` | 任务完成但产物未落盘 | 可重试 |
| `network_required` | 当前任务需要网络授权 | 授权后可重试 |
| `provider_profile_missing` | 缺少目标 profile | 补齐后可重试 |
| `provider_mode_unsupported` | 当前任务不支持所选 ProviderMode | 切换模式或补齐适配器后可重试 |
| `agent_gateway_unavailable` | 外部 Agent 入口未启用或不可达 | 启用 Agent Skill/CLI 后可重试 |
| `cost_placeholder_only` | 成本、扣费或积分仅为占位信息 | 不阻断运行，不执行真实扣费 |
| `user_cancelled` | 用户取消任务 | 由用户决定 |

## 成本 / 扣费 / 积分占位

当前只保留 provider cost、credit、point 和 quota 的 UI 占位与事件字段，用于提示“可能有外部成本”。这些字段不参与余额计算、不执行扣费、不阻断功能验收，也不作为任务调度依据。

| 字段 | 当前语义 | 禁止行为 |
| --- | --- | --- |
| `estimatedCostLabel` | 展示型成本文案，例如“外部成本占位” | 不能当作真实价格 |
| `creditPlaceholder` | 展示型积分占位 | 不能扣减、充值或结算 |
| `quotaWarning` | 用户提示可能受 provider 限制 | 不能伪造会员状态 |

## 恢复与重试

- 重试不覆盖旧 attempt、旧输出或旧错误。
- 取消后不自动重试。
- 超时后可沿用同一上下文创建新 attempt，也可调整超时后重试。
- idempotency key 命中时，系统提示已有 attempt，用户可查看或新建。
- 不可重试错误必须说明需要修复的对象。

## 验收标准

- 所有运行错误转换为标准错误码。
- 成本、扣费和积分仅作为占位可见，不形成真实财务或权限逻辑。
- 用户提示包含原因、影响和下一步。
- retryable 与不可重试错误区分明确。
- 取消和超时都有独立状态与记录。
- 重试创建新 attempt，不覆盖旧产物。
- 幂等键能阻止无提示的重复写入。
