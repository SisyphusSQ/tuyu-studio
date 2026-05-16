# Script to Scene

## 目标与边界

定义从剧本文本到 `ScriptScene` 的导入、拆场、候选确认和失败回退规则。本文不负责 ShotCard 字段、交接包导出或结果回收。

## 输入 / 输出

输入：

- 粘贴文本、文本文件或已导入 `script_source` Asset。
- 用户手动拆场操作。
- 结构化拆场候选输出。

输出：

- `ScriptDocument`。
- 经用户确认的 `ScriptScene[]`。
- 脚本展开表格视图、Shot 候选行和行级来源引用。
- 候选解析记录、source range 和失败原因。

## 核心对象或规则

脚本导入：

- 原文必须完整保留在 `ScriptDocument.rawText` 或源 Asset 中。
- 原文不被自动改写；拆场只产生结构化视图。
- 每个 `ScriptScene` 应记录 `sourceRange.startLine` 和 `sourceRange.endLine`。

拆场方式：

- 手动拆场：用户选择文本范围，填写场景标题、地点、时间、人物、道具和动作。
- 候选拆场：系统基于原文生成候选，必须由用户确认后写入正式 scenes。
- 混合拆场：用户可接受候选中的部分场次，并手动补齐其余片段。

## 脚本展开表格视图

脚本展开表格是 Project Canvas 上的可展开 Table View。它用于把剧本文本、角色引用、角色图、时长和画面描述组织为可扫描的分镜候选，不是独立文档页，也不是导出后的静态表格。用户应能在画布中缩放、拖拽、全屏展开，并从表格行继续生成 Shot、Prompt 或分镜。

| 区域 | 内容 | 要求 |
| --- | --- | --- |
| 标题栏 | 场次名，例如“夏日初晴晚风” | 保留来源 ScriptScene 和 source range |
| 表格列 | 镜号、时长、画面描述、角色、角色描述、角色图、第二角色等 | 支持横向滚动和列宽保持，不因长文挤破布局 |
| 行数据 | 一行代表一个 Shot 候选或已确认 Shot 的表格投影 | 行级可确认、重生成、生成分镜、进入 Inspector |
| 顶部动作 | 重新生成、生成分镜、下载 | 动作写入 Run/Event/Audit，不静默覆盖旧候选 |
| 视图切换 | 脚本视图、Shot 视图、Prompt 视图 | 切换只改变投影，不改变领域对象 truth |
| 全屏展开 | 临时放大表格画板 | 只改变 UI viewport，不改变 ProjectCanvas 数据 |

建议字段：

```ts
interface ScriptExpansionRow {
  id: string
  scriptSceneId: string
  shotCandidateId?: string
  shotId?: string
  index: number
  durationSeconds: number
  visualDescription: string
  characterRefs: Array<{
    characterId?: string
    name: string
    description: string
    referenceAssetId?: string
  }>
  sourceRange: {
    startLine: number
    endLine: number
  }
  status: "candidate" | "accepted" | "stale" | "blocked"
}
```

写入规则：

- 表格行默认是 candidate，不直接创建正式 ShotCard。
- 用户确认行后才创建或更新 ShotCard，并记录确认人、确认时间、来源行和被覆盖字段。
- Agent 可通过 Studio Command 写入候选行，但不能覆盖已接受行或锁定连续性字段。
- 重新生成只创建新候选版本，旧候选保留在 run record 中。
- 下载表格是导出视图，不作为项目 truth。

失败回退：

- 解析失败时保留 ScriptDocument。
- UI 提供手动拆场入口和错误摘要。
- 不允许因解析失败丢弃原文或生成空白正式场次。

## 状态推进

```text
imported -> split_pending -> candidate_ready -> confirmed
split_pending -> parse_failed -> manual_split
candidate_ready -> partially_confirmed -> confirmed
```

- `candidate_ready`：候选只读，未进入正式 scenes。
- `confirmed`：用户确认后的 Scene 可继续生成 Shot 候选。
- `expansion_ready`：脚本展开表格已生成，可逐行确认或进入全屏查看。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `script_empty` | 原文为空 | 阻断导入 |
| `script_asset_missing` | 源 Asset 缺失 | 允许粘贴原文重建 |
| `scene_parse_failed` | 候选拆场失败 | 保留原文，进入手动拆场 |
| `source_range_invalid` | 场次范围越界或重叠异常 | 阻断确认并提示修正 |
| `scene_required_field_missing` | 标题、地点或动作缺失 | 不允许确认该场次 |
| `script_expansion_row_invalid` | 展开行缺少镜号、时长、画面描述或 source range | 阻断行确认，保留候选 |

## 恢复与重试

- 候选拆场可基于同一原文重跑，旧候选保留为 run record。
- 脚本展开可基于同一 ScriptScene 重跑，旧展开版本保留为候选历史。
- 手动拆场不依赖候选成功。
- 用户可重新分配 source range，系统记录修订时间。
- 源 Asset 缺失时，已确认 scenes 保留，但健康检查标记原文追溯不完整。

## 验收标准

- 粘贴剧本能创建 ScriptDocument 并保留完整原文。
- 手动拆场能创建带 source range 的 ScriptScene。
- 候选拆场必须经用户确认才进入正式 scenes。
- 脚本展开表格能在 Project Canvas 内展开、全屏查看、逐行确认和发起生成分镜。
- 解析失败后仍可继续手动拆场。
- source range 越界或必填字段缺失时不能确认场次。
- 已确认 Scene 可作为 Shot 候选生成的输入。
