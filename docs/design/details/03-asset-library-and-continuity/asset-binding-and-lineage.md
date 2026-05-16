# Asset Binding and Lineage

## 目标与边界

定义资产如何绑定到业务对象，以及如何从任意资产追溯来源、使用范围和删除影响。本文不定义导入细节和连续性规则内容，只定义绑定、来源和查询口径。

## 输入 / 输出

输入：

- 已就绪的 `Asset`。
- 绑定目标：角色、场景、道具、风格、镜头、提示对象、生成包或结果。
- 来源对象：用户导入、运行记录、生成包、结果回收或 managed reference。

输出：

- `AssetBinding`。
- 来源记录和反向索引。
- 影响分析结果，供删除、导出、review 和健康检查使用。

## 核心对象或规则

绑定模型：

```ts
interface AssetBinding {
  id: string
  assetId: string
  targetType: "character" | "scene" | "prop" | "style" | "shot" | "prompt" | "package" | "result"
  targetId: string
  purpose: string
  locked: boolean
  createdBy: "user" | "system"
  createdAt: string
}
```

来源模型：

```ts
interface AssetSource {
  kind: "imported" | "generated" | "exported" | "managed_reference" | "result_import"
  sourceAssetId?: string
  promptRunId?: string
  packageId?: string
  resultId?: string
  importedName?: string
}
```

规则：

- 一个 Asset 可有多个绑定，绑定是使用关系，不代表文件所有权。
- `locked=true` 表示该绑定参与连续性或交接验证，解绑必须要求确认。
- PromptRun 不是 GraphNode/GraphEdge endpoint；资产通过 `promptRunId` 或 run record 字段引用它。
- GenerationPackage 与 Asset 的关系通过 manifest 引用和 `packageId` 记录表达。
- 结果文件必须绑定到 Shot 或 Package 之一，允许两者同时存在。

典型 lineage 查询：

- 从 Asset 查来源：导入文件、PromptRun、Package 或 Result。
- 从 Asset 查使用：角色、场景、道具、Shot、Prompt、Package、Result。
- 从 PromptRun 查产物：输出 Asset、错误文件、导出包。
- 从 Package 查引用：包内引用资产、对应 Shot、对应 PromptRun。
- 从 Shot 查全链路：参考资产、提示词、Package、Result take。

## 状态推进

```text
unbound -> bound -> locked_bound
bound -> unbound
locked_bound -> unlock_requested -> bound | locked_bound
bound -> stale_binding
```

- `stale_binding`：目标对象不存在或资产文件缺失。
- `unlock_requested`：用户正在确认是否解除锁定用途。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `binding_target_missing` | 目标对象不存在 | 阻断绑定写入 |
| `binding_duplicate` | 同用途绑定已存在 | 提示复用已有绑定 |
| `binding_locked` | 锁定绑定不允许静默移除 | 要求用户确认 |
| `lineage_source_missing` | 来源对象缺失 | 标记健康检查风险 |
| `delete_impact_blocked` | 删除影响包含锁定对象 | 阻断删除并展示影响 |

## 恢复与重试

- stale binding 不自动删除，健康检查提供修复入口。
- 目标对象恢复后可重新校验绑定状态。
- 错绑结果可先解绑再绑定到正确 Shot 或 Package，Asset 文件保留。
- PromptRun 或 Package 记录缺失时，保留 Asset，并将来源标记为不可完整追溯。

## 验收标准

- 任意 Asset 能列出来源对象和所有绑定目标。
- 删除影响分析能列出受影响角色、场景、道具、镜头、提示对象、包和结果。
- 锁定绑定不能被静默解除。
- 从生成图片能追溯到对应 PromptRun。
- 从交接包能列出包内所有引用资产。
- 错绑结果可恢复，不删除结果文件。
