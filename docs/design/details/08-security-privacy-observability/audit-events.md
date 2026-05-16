# 审计事件

## 目标与边界

定义审计事件结构、分类、必记事件、动作摘要，以及交接包导出和路径拒绝示例。本文不替代运行记录；PromptRun 仍保存任务级细节。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 用户关键操作 | audit/events.jsonl 事件 |
| 状态变更 | 目标类型、目标 ID、前后状态摘要 |
| 安全拒绝 | security 分类事件 |
| 导出、回收、删除 | 可追踪动作摘要 |
| PromptRun 状态 | run 分类事件和 runId 引用 |

## 核心对象或规则

事件 schema：

```json
{
  "id": "evt_20260514_102000_001",
  "timestamp": "2026-05-14T10:20:00+08:00",
  "actor": "local_user",
  "category": "package",
  "action": "package.exported",
  "projectId": "proj_001",
  "targetType": "GenerationPackage",
  "targetId": "pkg_001",
  "summary": "Exported handoff package for shot_001",
  "metadata": {
    "shotId": "shot_001",
    "packageStatus": "ready",
    "path": "packages/pkg_001"
  }
}
```

分类：

| 分类 | 事件 |
| --- | --- |
| project | create、open、close、migrate、recover、lock_takeover |
| asset | import、bind、unbind、missing、relocate、trash_restore |
| graph | node_create、edge_create、edge_reject、layout_save、delete |
| run | create、start、progress、complete、fail、cancel、retry |
| package | validate、export、reexport、handoff、mark_stale |
| result | import、bind、review、discard、restore |
| security | path_reject、permission_fail、external_confirm、privacy_block |

必记事件：

- 项目创建、打开失败、迁移、恢复、锁接管。
- 资产导入、重新定位、删除、恢复。
- 图节点创建、删除、非法边拒绝。
- PromptRun 创建、完成、失败、取消、重试。
- GenerationPackage 校验、导出、重新导出、交接、失效。
- VideoResult 导入、绑定、review、废弃、恢复。
- 路径拒绝、外部导出确认、隐私阻断。

动作摘要：

- summary 面向用户，可读且不包含敏感数据。
- metadata 使用对象 ID、相对路径、错误码和状态摘要。
- 不记录长正文、凭据、机器私有路径或完整外部地址。

## 状态推进

| 状态 | 推进规则 |
| --- | --- |
| event_pending | 业务操作已确认，事件待写入 |
| event_written | 写入 jsonl 并可被审计 tab 读取 |
| event_deferred | 审计写入暂时失败，进入本地缓冲 |
| event_replayed | 缓冲事件重新写入成功 |
| event_failed | 多次失败后进入 Project Health |

审计写入失败不能让危险操作静默完成；删除、外部导出、锁接管这类操作在审计不可用时需要阻断或要求用户处理。

## 错误语义

| 错误 | 语义 |
| --- | --- |
| audit_write_failed | 审计事件写入失败 |
| audit_schema_invalid | 事件结构不符合 schema |
| audit_redaction_failed | 事件包含敏感字段 |
| audit_read_failed | 审计 tab 无法读取事件 |

## 恢复与重试

- 可延迟事件写入本地缓冲，并在项目恢复时重放。
- 审计 schema 错误必须修复后再写入正式日志。
- 脱敏失败时阻断事件和对应敏感操作，提示具体字段摘要。
- 审计日志损坏时，保留损坏文件，创建新日志并在 Health View 提示。

## 验收标准

- 审计事件包含 id、timestamp、actor、category、action、projectId、targetType、targetId、summary、metadata。
- 包导出事件记录包状态、镜头 ID、相对路径和校验摘要。
- 路径拒绝事件记录拒绝原因、对象 ID 和路径摘要。
- 安全拒绝、删除、回收、导出、运行失败都可在审计 tab 找到。
- 审计内容通过脱敏检查，不包含长正文、凭据或机器私有路径。
