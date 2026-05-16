# Shot Card Lifecycle

## 目标与边界

定义 ShotCard 从创建、补齐上下文、准备提示词、review 到修订循环的规则。本文不定义剧本拆场和交接包文件结构。

## 输入 / 输出

输入：

- 已确认 `ScriptScene` 或用户手动创建的镜头。
- 脚本展开表格中的候选行。
- 角色、场景、道具、参考资产和用户补充说明。
- review 记录和修订请求。

输出：

- `ShotCard`。
- 状态变更记录、review records、修订版本。
- 从脚本展开行确认而来的 Shot 来源链路。
- 可供提示编译和交接包导出的镜头上下文。

## 核心对象或规则

必填字段：

- `title`
- `description`
- `durationSeconds`
- `aspectRatio`
- `shotType`
- `cameraMovement`
- `action`
- `emotion`
- 至少一个角色、场景说明或明确的空场景理由

`context_ready` 规则：

- 必填字段完整。
- 需要连续性的角色、场景、道具已经绑定 profile 或明确豁免。
- referenceAssetIds 中的资产存在且可读。
- 无 blocking 连续性冲突。

`prompt_ready` 规则：

- Shot 已处于 `context_ready` 或更高状态。
- 已生成或手动确认 PromptObject。
- PromptObject 与当前 Shot context digest 一致。
- 不存在缺失的锁定参考资产。

Shot.status 允许使用 `package_ready` 表示 ready GenerationPackage 已存在，且 manifest 已通过验证。

脚本展开行确认：

| 来源字段 | 写入 ShotCard | 规则 |
| --- | --- | --- |
| 镜号 | `index` | 必须保持同一场次内唯一，冲突时要求用户确认插入或覆盖 |
| 时长 | `durationSeconds` | 必须大于 0；缺失时不能确认 |
| 画面描述 | `description`、`action` | 保留原文来源，不让 Agent 覆盖人工锁定字段 |
| 角色 / 角色描述 | `characterIds`、上下文摘要 | 能匹配 CharacterProfile 时绑定，不能匹配时保留候选文本 |
| 角色图 | `referenceAssetIds` | 必须是项目内 Asset 或受控引用 |
| source range | `sourceRange` 或审计 metadata | 用于追溯回 ScriptDocument |

从脚本展开表确认的 Shot 默认进入 `draft` 或 `context_ready`，取决于必填字段和连续性检查是否通过。表格确认不能绕过 ShotCard 的必填字段、连续性和资产检查。

review record：

```ts
interface ReviewRecord {
  id: string
  targetType: "shot" | "prompt" | "result"
  targetId: string
  decision: "approved" | "needs_revision" | "rejected"
  notes: string
  createdAt: string
}
```

## 状态推进

```text
draft -> context_ready -> prompt_ready -> package_ready -> submitted -> generated
generated -> approved
generated -> needs_revision -> context_ready | prompt_ready
draft -> invalid
```

- `needs_revision` 必须记录原因。
- `submitted` 表示 ready Package 已被用户手动交接，或未来 provider job attempt 已进入提交后状态。
- 修订 Shot 后，依赖旧上下文的 PromptObject 和 Package 标记为 stale。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `shot_required_field_missing` | 必填字段缺失 | 阻断 context_ready |
| `shot_expansion_source_missing` | 从展开表确认的行缺少 source range 或 ScriptScene | 阻断创建正式 Shot |
| `shot_context_conflict` | 连续性 blocking 冲突 | 阻断状态推进 |
| `shot_reference_missing` | 参考资产缺失 | 阻断提示生成或导出 |
| `shot_prompt_stale` | 提示词与当前上下文不一致 | 要求重编译或确认沿用 |
| `shot_review_reason_missing` | 退回无原因 | 阻断 review 提交 |

## 恢复与重试

- 用户补齐字段后可重新执行 context check。
- 展开行确认失败时保留候选行，用户可补字段、重新匹配角色或手动创建 Shot。
- 修订循环保留旧版本和 review 记录。
- stale PromptObject 可重编译，也可由用户确认后作为历史版本保留。
- 错误进入高状态时，健康检查可回退到最近满足条件的状态。

## 验收标准

- 缺少必填字段的 Shot 不能进入 `context_ready`。
- 脚本展开行确认后能创建 ShotCard，并能追溯到 ScriptScene、行号和 source range。
- 上下文完整且无阻断冲突的 Shot 可进入 `context_ready`。
- 提示词与上下文一致后可进入 `prompt_ready`。
- 导出成功且 Package ready 后可进入 `package_ready`。
- review 退回必须记录原因并进入修订循环。
- Shot 修订后旧 Package 标记为 stale，不被覆盖。
