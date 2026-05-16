# Image Generation Adapter

## 目标与边界

定义图像生成适配器的文本生成参考图、参考图改图、多图融合、输出路径分配、运行后文件校验、资产创建和失败行为。本文不定义视频交接包。

## 输入 / 输出

输入：

- 文字提示、Shot 上下文、风格、比例和输出规格。
- 角色、场景、道具等参考图。
- 预分配输出路径和 PromptRun。

输出：

- 图像文件。
- `Asset` 记录、digest、mime、尺寸和绑定。
- 失败记录和可重试建议。

## 核心对象或规则

任务类型：

| 类型 | 输入 | 输出 |
| --- | --- | --- |
| text-to-image | 提示、比例、风格、输出路径 | image Asset |
| reference-edit | 单张或少量参考图、修改说明 | image Asset |
| multi-image fusion | 角色图、场景图、道具图、Shot 说明 | fusion prompt 和可选 image Asset |

输出路径分配：

- 后端先分配 `assets/outputs/{run_id}/{asset_id}.png`。
- 运行时只允许写入预分配目录。
- 不从自然语言回复中猜测文件位置。

post-run validation：

- 文件存在。
- mime 为允许图片类型。
- digest 可计算。
- 尺寸满足任务要求或记录偏差。
- Asset 创建成功并绑定 PromptRun。

## 状态推进

```text
prepared -> running -> file_written -> validating -> asset_created
running -> failed
validating -> validation_failed
asset_created -> bound
```

- `file_written` 不代表成功，必须通过 validation。
- `bound` 后才可进入资产库和连续性引用。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `image_output_missing` | 未生成文件 | failed，可重试 |
| `image_mime_invalid` | 文件类型不支持 | validation_failed |
| `image_digest_failed` | 摘要计算失败 | validation_failed |
| `image_reference_missing` | 参考图缺失 | 阻断启动 |
| `image_asset_create_failed` | 文件存在但 Asset 写入失败 | 保留文件并提示恢复 |
| `image_generation_cancelled` | 用户取消 | cancelled，不绑定正式资产 |

## 恢复与重试

- 缺失参考图需补齐或移除后重试。
- validation failed 保留文件到 run 目录，不进入正式资产库。
- Asset 创建失败可在文件仍存在时重试资产化。
- 图像生成重试使用新输出路径，不覆盖旧文件。
- 用户取消后可保留临时文件，但需人工确认是否导入。

## 验收标准

- 文本生成参考图成功后创建 image Asset。
- 参考图改图和多图融合能记录所有输入 AssetRefs。
- 输出路径由后端预分配。
- 文件缺失或 mime 不合法不会创建正式 Asset。
- 成功产物绑定 PromptRun。
- 重试不覆盖旧图。
