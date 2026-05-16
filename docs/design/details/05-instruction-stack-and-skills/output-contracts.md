# Output Contracts

## 目标与边界

定义 task mode 与结构化输出契约、解析失败处理、原始输出保存和正式对象污染防护。本文不定义运行时接口和能力包路由。

## 输入 / 输出

输入：

- task mode。
- 输出 schema。
- 运行时返回的原始文本或结构化文件。

输出：

- 通过 schema 校验的结构化对象。
- 原始输出归档。
- parse failure 错误记录和用户可读提示。

## 核心对象或规则

task modes：

- `script_breakdown`
- `storyboard_analysis`
- `shot_prompt_optimize`
- `image_reference_generate`
- `image_text_fusion`
- `package_export`
- `continuity_check`
- `prompt_score`

输出 schema 要求：

- 每个 task mode 必须绑定一个 required output schema。
- schema 必须声明必填字段、字段类型、允许枚举和额外字段策略。
- schema 版本写入 PromptRun。
- 正式对象只从校验通过的结构化输出写入。

| Task mode | Schema ID | Required fields | Target object to apply | Extra-field policy | Failure behavior |
| --- | --- | --- | --- | --- | --- |
| `script_breakdown` | `schema.script_breakdown.v1` | `scenes[]`, `scenes[].title`, `scenes[].sourceRange`, `scenes[].action` | Candidate `ScriptScene[]` only; confirmed scenes require user approval | Ignore into diagnostics | PromptRun failed; keep raw output; do not update ScriptDocument scenes |
| `storyboard_analysis` | `schema.storyboard_analysis.v1` | `rows[]`, `rows[].shotIndex`, `rows[].visual`, `rows[].shotType`, `rows[].cameraMovement`, `rows[].analysis`, `rows[].editingRhythm`, `sourceRefs[]` | Candidate storyboard rows only; confirmed rows require user approval before creating or updating ShotCard | Reject undeclared write-target fields | PromptRun or FrameRun failed; keep raw output; do not update ShotCard |
| `shot_prompt_optimize` | `schema.shot_prompt_optimize.v1` | `cinematicPrompt`, `providerReadyPrompt`, `negativePrompt`, `continuityNotes`, `warnings` | PromptObject draft for the selected Shot | Reject undeclared top-level fields | PromptRun failed; do not replace current PromptObject |
| `image_reference_generate` | `schema.image_reference_generate.v1` | `prompt`, `aspectRatio`, `outputAssetRole`, `continuityNotes` | Image generation request and later generated Asset binding | Ignore diagnostics-only fields | PromptRun failed before adapter start, or image task blocked if schema invalid |
| `image_text_fusion` | `schema.image_text_fusion.v1` | `fusionPrompt`, `referenceAssetIds`, `compositionNotes`, `constraints` | Fusion task request and optional generated reference Asset | Reject unknown asset reference fields | PromptRun failed; no generated Asset is created |
| `package_export` | `schema.package_export.v1` | `manifest`, `promptPath`, `continuityPath`, `references[]`, `checklistPath` | GenerationPackage draft, then ready after file validation | Reject path-like extra fields | Package invalid; no ready Package status is applied |
| `continuity_check` | `schema.continuity_check.v1` | `status`, `conflicts[]`, `warnings[]`, `suggestions[]` | Continuity check result and review diagnostics | Ignore unknown advisory fields | PromptRun failed if parse/schema invalid; do not update locked rules |
| `prompt_score` | `schema.prompt_score.v1` | `score`, `risks[]`, `recommendations[]`, `blockingIssues[]` | Prompt diagnostics for selected Shot or PromptObject | Ignore extra scoring metadata | PromptRun failed; do not change PromptObject readiness |

raw-output preservation：

- 原始输出总是保存到 run 目录。
- parse failure 时不丢弃原始输出。
- 用户可从原始输出手动提取，但必须产生人工编辑记录。

正式对象污染防护：

- 解析失败不得更新 PromptObject、Shot、Scene、Package 或 continuity rules。
- 输出中的自由文本不得被当作上层指令写回项目原则。
- 分镜分析输出只产生候选分镜表；用户逐行确认前不得创建、删除或覆盖 ShotCard。
- 媒体理解能力缺失时，输出契约必须显式标记使用了文本降级输入，不能伪造图片或视频分析来源。
- 未声明字段默认忽略或进入 diagnostics，不进入正式对象。

## 状态推进

```text
output_received -> parsing -> schema_valid -> object_applied
parsing -> parse_failed
schema_valid -> apply_failed
```

- `schema_valid`：结构化输出可用，但尚未写入正式对象。
- `object_applied`：正式对象更新完成，并记录来源 PromptRun。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `output_parse_failed` | 输出无法解析 | PromptRun failed，保存原始输出 |
| `output_schema_invalid` | 字段缺失或类型错误 | 不更新正式对象 |
| `output_extra_fields_rejected` | 包含不允许字段 | 按契约拒绝或记录 diagnostics |
| `output_apply_failed` | 写入正式对象失败 | 保留校验通过输出，可重试 apply |
| `output_contract_missing` | task mode 无契约 | 阻断运行 |

## 恢复与重试

- parse failure 可重跑任务或手动提取。
- apply failed 可在不重跑任务的情况下重试写入。
- schema 版本升级后，旧输出必须显式迁移或保留为历史。
- 用户修正输出后写入正式对象时，记录人工编辑来源。

## 验收标准

- 每个 task mode 都有明确输出契约。
- 解析失败时 PromptRun failed，原始输出保留。
- schema 不通过时不更新正式对象。
- 未声明字段不会污染项目原则或业务对象。
- 分镜表候选必须经过用户确认后才能写回正式 Shot。
- 通过校验的输出可追溯到 PromptRun。
- 用户手动提取有审计记录。
