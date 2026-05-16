# Continuity Rules

## 目标与边界

定义角色、场景、道具和项目风格在多镜头生产中的连续性对象、规则强度和冲突输出。本文覆盖任务前检查与结果 review 检查，不覆盖资产导入和包导出。

## 输入 / 输出

输入：

- `CharacterProfile`、`SceneProfile`、`PropProfile` 和项目风格规则。
- ShotCard 上下文、参考资产绑定和用户当前请求。
- 已锁定或未锁定的 `ContinuityRule`。

输出：

- 连续性检查结果：通过、警告、建议或阻断。
- 冲突列表、替代建议和 review 记录。
- 可写回 Shot 或 Result 的连续性偏离说明。

## 核心对象或规则

`CharacterProfile` 必须表达身份、外观、服装、发妆、性格、动作习惯、对白风格、参考资产和禁改项。

`SceneProfile` 必须表达地点、时间、时代、气氛、光线、空间关系、参考资产和锁定规则。

`PropProfile` 必须表达名称、类别、外观、使用方式、出现范围、参考资产和锁定规则。

`ContinuityRule`：

```ts
interface ContinuityRule {
  id: string
  targetType: "character" | "scene" | "prop" | "style" | "project"
  targetId: string
  rule: string
  severity: "blocking" | "warning" | "suggestion"
  locked: boolean
  createdBy: "user" | "system"
}
```

强度语义：

- `blocking`：违反时任务不可运行或结果不可直接通过。
- `warning`：允许继续，但必须展示风险并写入 run record。
- `suggestion`：不阻断，用于优化提示和 review 提醒。
- `locked=true`：用户当前请求、能力包和自动生成内容都不得覆盖。

冲突输出：

```json
{
  "status": "blocked",
  "conflicts": [
    {
      "ruleId": "rule_char_costume_001",
      "targetType": "character",
      "severity": "blocking",
      "message": "镜头服装描述与锁定规则冲突",
      "suggestion": "保留锁定服装，只调整动作、表情和构图"
    }
  ]
}
```

## 状态推进

```text
draft_rule -> active -> locked
active -> disabled
locked -> unlock_requested -> active | locked
check_pending -> passed | warning | blocked
```

review-time checks：

- 结果通过时，可记录“连续性符合”。
- 结果偏离但可接受时，记录 warning 和原因。
- 结果违反 blocking 规则时，review 不能直接通过，除非用户先修改规则状态。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `continuity_context_missing` | 缺少角色、场景或道具上下文 | 阻断需要该上下文的任务 |
| `continuity_locked_conflict` | 用户请求违反锁定规则 | 阻断并给出替代建议 |
| `continuity_rule_ambiguous` | 规则文本无法判定 | 降为 warning 并要求人工确认 |
| `continuity_reference_missing` | 规则依赖的参考资产缺失 | 阻断图像类任务或提示补齐 |

## 恢复与重试

- 用户可补齐缺失 profile 或参考资产后重试。
- 用户可请求解锁规则，但必须记录解锁原因和审计事件。
- warning 可继续运行，冲突摘要必须写入 PromptRun 或 review record。
- 规则歧义时允许用户改写规则文本，再重新检查。

## 验收标准

- 角色、场景、道具和项目风格均可创建连续性规则。
- 锁定 blocking 规则被违反时任务被阻断。
- warning 和 suggestion 不阻断，但必须可见且可追溯。
- 结果 review 能记录连续性偏离和处理决定。
- 冲突输出包含规则 ID、目标、强度、说明和替代建议。
- 解锁锁定规则需要确认并产生审计事件。
