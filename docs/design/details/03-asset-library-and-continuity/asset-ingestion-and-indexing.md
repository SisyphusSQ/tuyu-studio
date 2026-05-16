# Asset Ingestion and Indexing

## 目标与边界

定义资产进入项目后的导入、校验、索引和预览规则。本文覆盖用户导入、生成产物落盘后的资产化、交接包引用文件登记，以及 managed reference 的风险治理；不定义角色、镜头或交接包的业务字段。

## 输入 / 输出

输入：

- 用户选择或拖入的文件。
- 运行产物的预分配输出路径。
- 可选绑定目标，例如角色、场景、道具、镜头、包或结果。
- 用户对重复文件的处理选择。

输出：

- `Asset` 记录、文件副本或 managed reference 记录。
- `digest`、`mimeType`、尺寸、时长、大小和缩略图索引。
- 重复资产提示、风险提示和导入审计事件。

## 核心对象或规则

导入流程：

1. 读取文件基础 metadata，不信任扩展名。
2. 使用内容识别校验 mime，并与允许列表比对。
3. 计算内容摘要，摘要用于重复检测、完整性检查和结果复用提示。
4. 对图片、视频、音频生成可预览 metadata；图片和视频生成缩略图。
5. 非重复文件复制到 `{studio_root}/projects/{project_id}/assets/...`。
6. 创建 `Asset`，再按用户选择创建 `AssetBinding`。

路径规则：

- 项目内资产路径必须是项目相对路径。
- 文件名保留可读前缀，但最终文件名必须包含稳定 `assetId`。
- 缩略图位于 `assets/thumbnails/{asset_id}.jpg`，可被重新生成。
- 交接包导出的文件是包产物，不反向写入源资产目录，除非用户显式导入。

重复处理：

- 同 digest、同 mime 的文件提示复用已有资产。
- 用户可选择只新增绑定、创建新资产副本，或取消导入。
- 创建新资产副本时必须保留重复来源关系，避免误判为独立素材。

managed reference：

- 仅用于大文件或用户明确选择不复制的文件。
- 必须记录外部引用状态、最后校验时间和迁移风险。
- 不能作为默认导入路径，也不能静默参与需要可移植交接的导出。

## 状态推进

```text
selected -> inspecting -> rejected | duplicate_pending | copying | indexed -> ready
duplicate_pending -> ready | cancelled
copying -> failed
indexed -> thumbnail_failed | ready
```

- `rejected`：mime 不支持、文件不可读或路径越界。
- `thumbnail_failed`：资产可用，但预览缺失，允许后续重建。
- `ready`：Asset、索引和必要绑定均已写入。

## 错误语义

| 错误 | 含义 | 用户提示 |
| --- | --- | --- |
| `asset_file_unreadable` | 文件不可读 | 检查文件权限或重新选择文件 |
| `asset_mime_rejected` | 内容类型不支持 | 选择受支持的图片、视频、音频或文本文件 |
| `asset_digest_failed` | 摘要计算失败 | 文件可能损坏，建议重新导入 |
| `asset_duplicate_found` | 发现同内容资产 | 选择复用、创建副本或取消 |
| `asset_thumbnail_failed` | 缩略图生成失败 | 资产已保留，可稍后重建预览 |
| `managed_reference_risky` | 外部引用不可移植 | 建议复制进项目资产目录 |

## 恢复与重试

- `copying` 失败时清理未完成文件，保留导入错误记录。
- `thumbnail_failed` 可单独重建缩略图，不重复复制源文件。
- digest 已存在时，重试必须复用重复检测结果，不创建隐式副本。
- managed reference 校验失败时标记风险，不删除 Asset 记录。
- 导入中断后，临时文件位于 `{studio_root}/projects/{project_id}/.tmp/imports/`，下次启动提示清理或恢复。

## 验收标准

- 支持导入图片、视频、音频、文本并生成 Asset 记录。
- 同一文件重复导入时提示复用或创建新绑定。
- 图片和视频能生成缩略图，缩略图缺失不破坏原资产。
- mime 校验以内容为准，伪装扩展名不得进入生产链路。
- managed reference 在健康检查中列出迁移风险。
- 所有正式资产路径均使用项目相对路径。
