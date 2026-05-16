# Result Ingestion and Review

## 目标与边界

定义外部结果文件回收到项目后的编号、绑定、review、错绑恢复和缺失处理。本文不定义生成包导出或外部提交接口。

## 输入 / 输出

输入：

- 用户拖入或选择的结果文件。
- 目标 Shot、Package 或二者。
- review 决策、备注和重做要求。

输出：

- `VideoResult` 或图片结果记录。
- 结果 Asset、take 编号、绑定关系和 review record。
- Shot 与 Package 的状态更新。

## 核心对象或规则

结果导入：

- 结果文件复制到 `assets/results/{shot_id}/`。
- 按 Shot 维度分配 `takeNumber`，同一 Shot 的 take 从 1 递增。
- 结果必须绑定到 Shot 或 Package；未选择时进入待绑定状态，不进入正式 review。
- 同 digest 文件再次导入时提示复用已有结果或创建新 take。

绑定规则：

- Shot 绑定表达该结果属于哪个镜头。
- Package 绑定表达该结果来自哪次交接包。
- 结果回收不覆盖 PromptObject 或 GenerationPackage。

review states：

```text
pending -> approved
pending -> needs_revision
pending -> rejected
needs_revision -> pending
```

wrong-binding recovery：

- 允许解绑错误 Shot 或 Package。
- 重新绑定后保留原导入审计和 take 历史。
- 若 take 编号冲突，重新计算目标 Shot 下的显示编号，但保留原记录 ID。

## 状态推进

```text
imported -> binding_pending -> review_pending -> approved | needs_revision | rejected
review_pending -> missing_file
needs_revision -> revision_requested
```

- `missing_file`：结果记录保留，文件缺失被健康检查发现。
- `revision_requested`：用户选择基于 review 创建重做任务。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `result_target_missing` | 未选择 Shot 或 Package | 进入待绑定状态 |
| `result_file_missing` | 文件不存在或后续丢失 | 显示缺失，不删除记录 |
| `result_duplicate_digest` | 重复导入同内容文件 | 提示复用或创建新 take |
| `result_wrong_binding` | 用户确认绑定错误 | 支持解绑重绑 |
| `result_review_note_missing` | 退回或拒绝缺少原因 | 阻断 review 提交 |

## 恢复与重试

- 文件复制失败可重试导入，不创建正式结果记录。
- 待绑定结果可在后续选择目标后进入 review。
- 缺失文件可通过重新定位或重新导入恢复。
- 错绑恢复不得删除结果 Asset。
- needs_revision 可生成新的 Shot 修订或新的 Package 导出请求。

## 验收标准

- 拖入结果文件能创建结果 Asset 和 VideoResult。
- 同一 Shot 支持多个 take，并可独立 review。
- 结果可绑定到 Shot 和 Package。
- 错误绑定可解绑并重新绑定，不删除文件。
- 缺失结果文件显示缺失状态并保留历史。
- review 退回或拒绝必须记录原因。
