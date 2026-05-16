# Package Export

## 目标与边界

定义生成交接包的目录结构、manifest schema、提示文件、连续性文件、清单、相对路径和重新导出规则。本文不负责 ProviderProfile 的能力定义。

## 输入 / 输出

输入：

- `ShotCard`、PromptObject、ProviderProfile、连续性规则和参考资产。
- 用户选择的单 Shot 导出或 Scene 批量导出。

输出：

- `GenerationPackage`。
- 包目录、manifest、prompt、script excerpt、continuity、upload checklist 和 references。
- 导出验证结果。

## 核心对象或规则

目录结构：

```text
packages/
└─ scene_{scene_index}/
   └─ shot_{shot_index}_pkg_{timestamp}/
      ├─ manifest.json
      ├─ prompt.txt
      ├─ script_excerpt.md
      ├─ continuity.md
      ├─ upload_checklist.md
      └─ references/
```

manifest schema 必须包含：

- `schemaVersion`
- `projectId`
- `sceneId`
- `shotId`
- `packageId`
- `packageVersion`
- `providerProfileId`
- `generationPackageStatus`
- `promptPath`
- `continuityPath`
- `uploadChecklistPath`
- `references[]`
- `createdAt`

GenerationPackage.status 使用：`draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid`。

路径规则：

- manifest 只允许包内相对路径。
- references 文件必须复制进包目录，不能引用项目外路径。
- prompt 文件必须可直接复制使用，并包含镜头目标、连续性摘要和必要负向约束。
- continuity 文件是导出时规则快照，不随项目后续规则自动改变。
- checklist 必须列出用户手动交接步骤、需要上传的参考文件和确认项。

重新导出：

- 不覆盖旧目录。
- 新 packageVersion 指向当前 Shot context digest。
- 旧包若上下文已变化，标记为 `stale`。

## 状态推进

```text
draft -> validating -> ready -> handed_off -> result_received
validating -> invalid
ready -> stale
```

- `ready`：manifest 可解析，引用文件存在，prompt 和 checklist 已生成。
- `handed_off`：用户确认已交接。
- `result_received`：至少一个结果绑定到该 Package。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `package_shot_not_ready` | Shot 未达到导出条件 | 阻断导出 |
| `package_reference_missing` | 引用资产缺失 | 阻断导出或要求用户明确忽略 |
| `package_manifest_invalid` | manifest 不可解析或字段缺失 | 标记 invalid |
| `package_path_invalid` | manifest 含越界路径 | 阻断导出 |
| `package_write_failed` | 文件写入失败 | 保留临时目录并提示重试 |

## 恢复与重试

- 导出失败保留临时目录用于诊断，未通过验证不得创建 ready Package。
- 修复缺失资产后可重新导出，新版本不覆盖旧版本。
- manifest 验证失败时，用户可删除失败包或重新生成。
- Package 标记 stale 后，可继续用于历史追溯，但不作为默认交接包。

## 验收标准

- 单 Shot 能导出完整包目录。
- manifest JSON 可解析，所有引用路径存在且为相对路径。
- prompt、continuity 和 checklist 文件齐全。
- 缺失关键参考资产时导出被阻断或记录用户明确忽略。
- 重新导出生成新目录和新版本。
- 旧包在 Shot 上下文变化后标记为 stale。
