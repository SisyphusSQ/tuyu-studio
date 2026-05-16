# Deletion and Recovery

## 目标与边界

定义资产解绑、删除、回收、恢复和永久清理的生产规则。本文只处理资产生命周期，不处理业务对象删除策略。

## 输入 / 输出

输入：

- 用户发起的解绑、删除、恢复或永久清理命令。
- Asset、AssetBinding、lineage 查询和删除影响分析。
- 用户确认信息。

输出：

- 更新后的绑定、回收区记录或清理记录。
- 审计事件。
- 健康检查可读的缺失、恢复或清理状态。

## 核心对象或规则

操作语义：

| 操作 | 文件 | 绑定 | 业务对象 | 是否需要确认 |
| --- | --- | --- | --- | --- |
| 解绑 | 保留 | 删除指定 binding | 保留 | 锁定绑定需要确认 |
| 删除未绑定资产 | 移入回收区 | 无 | 保留 | 需要确认 |
| 删除已绑定资产 | 移入回收区 | 默认保留为 suspended | 保留 | 必须展示影响并确认 |
| 恢复资产 | 移回原位置或新位置 | 恢复可用绑定 | 保留 | 路径冲突需要确认 |
| 永久清理 | 删除回收区文件 | 删除或归档回收记录 | 保留审计摘要 | 必须二次确认 |

回收区：

- 路径：`{studio_root}/projects/{project_id}/.trash/assets/{asset_id}/`。
- 回收记录必须包含原相对路径、digest、绑定摘要、来源摘要、操作者、时间和原因。
- 缩略图可重新生成，不要求进入回收区；但清理事件要记录。

## 状态推进

```text
ready -> unlinking -> ready
ready -> delete_pending -> trashed -> restoring -> ready
trashed -> purge_pending -> purged
trashed -> restore_failed
```

- `delete_pending`：已完成影响分析，等待用户确认。
- `trashed`：文件不在正式资产目录，但记录仍可查询。
- `purged`：文件已永久清理，只保留审计摘要和不可恢复状态。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `delete_impact_requires_confirmation` | 资产仍被使用 | 展示影响对象并等待确认 |
| `delete_locked_binding_blocked` | 存在锁定绑定 | 要求先解锁或确认解除影响 |
| `trash_move_failed` | 移入回收区失败 | 保持原状态，提示重试 |
| `restore_path_conflict` | 原路径已被占用 | 允许选择新路径或取消 |
| `purge_confirmation_missing` | 缺少二次确认 | 阻断永久清理 |

## 恢复与重试

- 删除失败不得留下半删除状态；文件和 Asset 状态必须一致。
- restore 失败保留 `trashed` 状态和失败原因。
- 原路径被占用时，可恢复到新路径并更新 Asset relativePath。
- permanent purge 不可恢复，只能通过外部备份重建为新 Asset。
- 错误绑定造成的删除影响可先解绑，再重新执行删除。

## 验收标准

- 解绑资产不会删除文件。
- 删除已绑定资产前能列出所有影响对象。
- 回收区记录能支持恢复文件和绑定状态。
- 永久清理需要二次确认，并保留审计摘要。
- purge 后 lineage 查询显示不可恢复状态而不是丢失记录。
- 缩略图清理可重建，不影响 Asset 主文件。
