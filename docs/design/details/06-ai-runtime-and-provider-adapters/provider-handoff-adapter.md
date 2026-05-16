# Provider Handoff Adapter

## 目标与边界

定义 ProviderProfile、理解能力、生成能力、手动交接、未来 API 边界，以及 BuildPrompt、ExportPackage、SubmitPackage、PollResult 的语义。本文不绑定具体外部平台。

## 输入 / 输出

输入：

- ProductionFrame、ShotCard、PromptObject、参考资产、连续性规则和 ProviderProfile。
- 用户选择的交接目标和包导出参数。

输出：

- provider-ready prompt。
- `GenerationPackage`。
- 手动交接状态，或未来 provider job / adapter attempt 状态。

## 核心对象或规则

ProviderProfile：

```ts
interface ProviderProfile {
  id: string
  name: string
  mode: "manual_handoff" | "api_submit"
  promptSections: string[]
  assetRules: {
    maxReferenceImages?: number
    acceptedImageTypes: string[]
    acceptedVideoTypes: string[]
  }
  packageRules: {
    includeManifest: boolean
    includeScriptExcerpt: boolean
    includeContinuity: boolean
    includeUploadChecklist: boolean
  }
  capabilityFlags: {
    textUnderstanding: boolean
    imageUnderstanding: boolean
    videoUnderstanding: boolean
    imageGeneration: boolean
    videoGeneration: boolean
    apiSubmit: boolean
    pollResult: boolean
  }
}
```

方法语义：

- `BuildPrompt`：生成目标 profile 下可复制的提示文本，不创建 Package。
- `ExportPackage`：创建本地交接包，Package 初始状态为 `ready` 或 `invalid`。
- `SubmitPackage`：仅在 profile 声明 `apiSubmit` 时启用；否则返回未配置并提示用户执行手动交接。
- `PollResult`：仅在 profile 声明 `pollResult` 时启用；否则返回未配置并提示用户手动回收结果。
- `videoGeneration`：仅表示 provider 可接收生成视频任务，不代表它具备视频理解能力。

能力语义：

| capability | 语义 | 缺失时行为 |
| --- | --- | --- |
| `textUnderstanding` | 可处理剧本、提示词、分镜表和文本备注 | 文本任务不可运行 |
| `imageUnderstanding` | 可读取参考图语义 | 保留图片节点，要求用户补充文字描述 |
| `videoUnderstanding` | 可读取参考视频的动作、节奏和镜头信息 | 保留视频节点，允许手动分镜表或只用文本 |
| `imageGeneration` | 可生成图片输出 | 图片生成入口不可运行 |
| `videoGeneration` | 可生成视频输出 | 视频生成入口转为交接包或手动路径 |
| `apiSubmit` | 可自动提交 provider 任务 | 只允许导出和复制 |
| `pollResult` | 可自动轮询并回收结果 | 结果由用户手动导入 |

未来 API 边界：

- 新增 adapter 只能扩展 SubmitPackage 和 PollResult。
- 不得改变 Shot、PromptRun、Asset、GenerationPackage 的核心数据结构。
- API 返回的远端 job id 只作为 Package 或 Result 的外部引用字段。
- `submitted`、`polling` 这类状态属于 provider job 或 adapter attempt，不写入 GenerationPackage.status。
- 多模态理解运行时、文本理解运行时、视频生成运行时都按 capability 声明进入同一模型，不在画布中硬编码 provider 名称。

## 状态推进

```text
profile_selected -> prompt_built -> package_exported -> handed_off -> result_received
package_exported -> invalid
provider_attempt_created -> submitted -> polling -> provider_result_available
```

GenerationPackage.status 使用：`draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid`。

其中 `package_exported` 对应 GenerationPackage.status=`ready`，`handed_off` 和 `result_received` 分别对应同名 Package 状态；provider attempt 的 `submitted` / `polling` 只描述外部提交进度，不是 Package 状态。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `provider_profile_missing` | 未选择 profile | 阻断 prompt 和包导出 |
| `provider_capability_missing` | profile 不支持该任务 | 阻断并提示切换 |
| `provider_asset_rule_violated` | 参考图数量或类型不符合规则 | 阻断导出 |
| `provider_submit_not_configured` | 未配置自动提交 | 提示手动交接 |
| `provider_poll_not_configured` | 未配置自动回收 | 提示手动导入结果 |

## 恢复与重试

- 切换 ProviderProfile 后重新 BuildPrompt 或 ExportPackage。
- 参考资产不符合规则时，用户可减少引用或转换资产后重试。
- 缺少图片或视频理解能力时，用户可补充文本描述、导入手动分镜表，或切换到具备对应理解能力的 provider。
- 缺少视频生成能力时，用户可导出交接包或等待配置生成 provider，不需要重建 ProductionFrame。
- 手动交接失败由用户重新执行，不改变 Package 文件。
- 未来 API 提交失败时创建新 attempt，不覆盖本地包。

## 验收标准

- ProviderProfile 只通过能力标记和包规则影响导出。
- BuildPrompt 不创建 Package。
- ExportPackage 产出本地包并使用相对路径。
- 默认 SubmitPackage 和 PollResult 给出手动交接提示。
- 生成视频入口只在 `videoGeneration` capability 存在时启用。
- 图片和视频理解缺失时，参考节点仍保留在 Frame 中，并有可执行的降级路径。
- 切换 profile 不破坏 Shot、PromptRun、Asset 和 Package 核心对象。
- 不支持能力时阻断并说明原因。
