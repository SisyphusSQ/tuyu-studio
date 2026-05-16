# Instruction Compile Order

## 目标与边界

定义一次任务输入的编译顺序、优先级、data 包装、上下文摘要、预览要求和验收口径。本文不定义能力包目录扫描和运行时提交。

## 输入 / 输出

输入：

- Studio Profile、项目原则、角色/场景/道具/风格设定。
- ProviderProfile、选中的能力包、任务模板、画布上下文和用户请求。
- 输出契约和选中资产。

输出：

- `CompiledInstruction`。
- selected context、asset refs、skill refs、context digest。
- warnings、conflicts 和预览内容。

## 核心对象或规则

编译顺序：

```text
1. App Operating Contract
2. Studio Profile
3. Project Principles
4. Style / Character / Scene / Prop Bible
5. Provider Profile
6. Selected Skill Instructions
7. Task Template
8. Dynamic Canvas Context
9. User Request
10. Output Contract
```

优先级：

```text
locked project principles
  > task output contract
  > user current request
  > shot card
  > character / scene / prop bible
  > project principles
  > studio profile
  > free-form notes
```

data 包装：

- 剧本、用户备注、旧提示词、网页摘录、图片说明都必须作为 data。
- data 不得覆盖上层指令。
- 每段 data 标注来源、角色和对象 ID。

上下文摘要：

- context digest 基于参与编译的对象 ID、版本、关键字段和资产摘要生成。
- digest 变化会使依赖旧上下文的 PromptObject 或 Package 变 stale。
- 预览必须列出被纳入与被排除的对象。

## 状态推进

```text
requested -> collecting_context -> conflict_checking -> preview_ready -> compiled
conflict_checking -> blocked
collecting_context -> context_incomplete
```

- `preview_ready`：用户可查看完整编译输入，但尚未运行。
- `compiled`：运行输入已落盘并可被 RuntimeGateway 使用。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `compile_context_missing` | 必要上下文缺失 | 阻断运行，允许预览失败原因 |
| `compile_locked_conflict` | 用户请求违反锁定原则 | 阻断运行 |
| `compile_output_contract_missing` | 缺少输出契约 | 阻断运行 |
| `compile_context_overlarge` | 上下文超出可控范围 | 生成摘要或要求缩小选择 |
| `compile_asset_ref_missing` | 资产引用缺失 | 阻断依赖资产的任务 |

## 恢复与重试

- 补齐上下文后重新编译并生成新 digest。
- 用户缩小选择范围后可重新预览。
- 锁定冲突只能通过修改请求或解锁规则恢复。
- 编译失败不创建正式 PromptRun completed 记录，可保留 failed run record。

## 验收标准

- 编译结果按固定顺序组织。
- 用户可预览完整指令、上下文来源、能力包来源和输出契约。
- 所有创作素材均以 data 包装。
- locked project principles 优先于用户当前请求。
- context digest 能准确反映参与对象变化。
- 缺少输出契约时任务不可运行。
