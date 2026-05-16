# Runtime Gateway

## 目标与边界

定义任务运行时统一接口、ProviderMode、请求形状、工作目录、可写根、网络开关、事件归一化和任务生命周期。本文不绑定任何外部品牌或单一运行方式。

## 输入 / 输出

输入：

- `RuntimeTaskRequest`、ProviderMode 和 Studio Command 上下文。
- 编译后的指令文件、输出契约、资产引用和预分配输出目录。

输出：

- `RuntimeTaskHandle`。
- 归一化 `TaskEvent`。
- `RuntimeTaskStatus`、输出文件或错误文件。

## 核心对象或规则

request shape：

```go
type RuntimeTaskRequest struct {
    ProjectID          string
    RunID              string
    ProviderMode       ProviderMode
    TaskMode           string
    CanvasSelection    RuntimeCanvasSelection
    WorkingDirectory   string
    CompiledPromptPath string
    OutputSchemaPath   string
    AssetRefs          []RuntimeAssetRef
    WritableRoots      []string
    NetworkAccess      bool
    TimeoutSeconds     int
}

type ProviderMode string

const (
    ProviderModeInternalProvider ProviderMode = "internal_provider"
    ProviderModeExternalAgent    ProviderMode = "external_agent"
)

type RuntimeCanvasSelection struct {
    NodeIDs  []string
    FrameIDs []string
    Version  int
}
```

ProviderMode：

| 模式 | 含义 | 使用场景 | 审计要求 |
| --- | --- | --- | --- |
| `internal_provider` | Studio 内置 provider/profile/adapter 发起运行 | 本地运行、内置模型服务、受控 adapter | 记录 profile、capability、request、events、output contract |
| `external_agent` | 外部 Agent 读取 Skill 并通过 CLI/MCP 风格命令驱动画布任务 | Codex/其他 Agent 接管画布、生成候选、调用外部 CLI | 记录 agent identity、commandId、selection、permission、events |

两种模式都必须落到相同的 Run/Event/Audit 结构；差异只体现在执行器来源、权限确认和用户提示。UI 不应为外部 Agent 建立第二套结果模型。

约束：

- `WorkingDirectory` 默认是 `{studio_root}/projects/{project_id}`。
- `WritableRoots` 默认只包含项目目录和本次 run 输出目录。
- `NetworkAccess` 默认关闭，只有用户明确允许的任务才开启。
- 输出路径由产品层预分配，运行时不得自行决定正式落盘位置。
- 所有事件写入 `prompts/runs/{run_id}/events.jsonl`，并关联 ProjectCanvas version 与 selection。

event normalization：

- 外部进度、日志、警告、错误、等待用户和完成事件都转换为内部 `TaskEvent`。
- 事件必须带 `runId`、`attempt`、`timestamp`、`level` 和 `message`。
- 用户可见 message 不包含敏感路径或凭据。

## 状态推进

```text
queued -> running -> waiting_user -> running
running -> completed
running -> failed
running -> cancelled
failed -> retrying -> queued
```

- `completed` 只有在输出存在且校验通过后进入。
- `cancelled` 保留已产生的临时产物，但不自动写入正式对象。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `runtime_not_configured` | 运行时未配置 | 阻断启动 |
| `provider_mode_unsupported` | 当前任务不支持所选 ProviderMode | 阻断启动 |
| `agent_gateway_unavailable` | 外部 Agent 入口未启用或不可达 | failed，可重试 |
| `runtime_start_failed` | 启动失败 | failed，可重试 |
| `permission_denied` | 读写权限不足 | failed，提示修复 |
| `path_rejected` | 路径越界 | failed，不执行 |
| `network_required` | 任务需要网络但未授权 | waiting_user 或 failed |
| `task_timeout` | 超时 | failed，可重试 |

## 恢复与重试

- 重试创建新 attempt，复用同一 run lineage。
- 取消任务后用户可从同一上下文创建新 run。
- 权限、路径或网络授权修复后可重新启动。
- 运行时崩溃时保留 request、compiled instruction 和 events。

## 验收标准

- 未配置运行时时，任务按钮能说明不可运行原因。
- 每次任务都有 request、compiled instruction、events 和 status。
- 每次任务都记录 ProviderMode、CanvasSelection、Run/Event/Audit。
- internal_provider 和 external_agent 的输出都通过同一输出契约校验。
- 路径越界被拒绝。
- 网络默认关闭，授权状态可追溯。
- completed 状态必须经过输出校验。
- cancelled 不更新正式对象。
