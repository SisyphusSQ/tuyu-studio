# 图谱状态机

## 目标与边界

状态机用于约束 Shot、PromptRun、Package 和 VideoResult 的生产推进。状态变化必须由明确事件触发，并写入项目数据和 audit，不能只存在于界面内存。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 用户动作、健康检查结果、PromptRun 事件、Package 导出结果 | Shot、Package、VideoResult 的新状态 |
| 上游上下文、asset digest、Prompt 或 Review 变化 | dirty / stale 标记和影响报告 |
| 非法跳转请求 | 拒绝原因、缺失条件和当前状态 |

## 核心对象或规则

状态机覆盖 Shot、PromptRun、Package、VideoResult 和 Take。核心规则是状态只能由可审计事件推进，完成类状态必须有对应产物或用户确认，dirty / stale 状态优先于继续导出或评审。

## 状态推进

### Shot 状态推进

```text
draft
  -> context_ready
  -> prompt_ready
  -> package_ready
  -> submitted
  -> generated
  -> approved

generated -> needs_revision -> context_ready
context_ready -> context_dirty
prompt_ready -> context_dirty
package_ready -> context_dirty
```

| 状态 | 进入条件 | 退出条件 |
| --- | --- | --- |
| `draft` | 创建 Shot | 必填字段和基础引用补齐 |
| `context_ready` | 剧本片段、Scene、必要 Character / Prop / Style 检查通过 | Prompt 确认或上游变更 |
| `prompt_ready` | PromptObject 已确认并绑定 Shot | Package 导出或上游变更 |
| `package_ready` | GenerationPackage manifest 校验通过 | 用户标记 submitted 或上游变更 |
| `submitted` | 用户显式标记已手动交接 | 结果回收 |
| `generated` | VideoResult / Take 绑定成功 | 用户评审 |
| `approved` | 用户评审通过 | 后续手动重开修订 |
| `needs_revision` | 用户评审要求修改 | 修改上下文并重新检查 |
| `context_dirty` | 上游连续性、资产、Prompt 或 Package 依赖变化 | 用户重新确认上下文 |

### PromptRun 状态

| 状态 | 进入条件 | 允许动作 |
| --- | --- | --- |
| `queued` | 用户确认运行，任务进入队列 | 取消 |
| `running` | RuntimeGateway 开始执行 | 取消、记录事件 |
| `waiting_user` | 需要用户确认冲突、裁剪或补充输入 | 补充、取消 |
| `completed` | 输出符合契约并完成写入 | 接受写回、复用 |
| `failed` | 执行、解析或契约校验失败 | 重试、查看错误、复制诊断 |
| `cancelled` | 用户取消或系统安全终止 | 重跑 |

PromptRun 完成不自动覆盖用户已确认 Prompt；写回必须由用户确认或由明确规则允许。

### Package 状态

| 状态 | 含义 | 转换 |
| --- | --- | --- |
| `draft` | 正在准备导出 | 校验通过 -> `ready` |
| `ready` | manifest、资产和文本完整 | 用户标记交接 -> `handed_off` |
| `handed_off` | 用户已手动交接 | 回收结果 -> `result_received` |
| `result_received` | 结果已导入并绑定 | 后续上游变化生成新包或让新版本 stale；本包保留为不可变历史证据 |
| `stale` | 上游 Shot、Prompt、资产或连续性变化 | 重新导出生成新 Package |
| `invalid` | manifest 或引用损坏 | 修复或废弃 |

### VideoResult / Take 状态

| 状态 | 含义 | 转换 |
| --- | --- | --- |
| `unbound` | 文件已导入但未匹配 Shot 或 Package | 绑定 -> `pending_review` |
| `pending_review` | 已绑定，等待用户查看 | 评审 -> `approved` / `needs_revision` / `rejected` |
| `approved` | 结果通过 | 可锁定为当前推荐 Take |
| `needs_revision` | 需要修改 | 生成修订说明并回到 Shot |
| `rejected` | 废弃但保留历史 | 可恢复为 pending_review |

### dirty 状态处理

| 上游变化 | 影响对象 | 处理 |
| --- | --- | --- |
| Character 连续性变更 | 引用该 Character 的 Shot、Prompt、Package | 标记 `context_dirty` 或 `stale` |
| Scene 规则变更 | 所属 Shot 和 Package | 标记 dirty 并列出变化字段 |
| Asset digest 变化 | 引用 asset 的 Shot、Package、Result | 阻断导出，要求检查 |
| Prompt 修改 | 相关 Package | Package 标记 `stale` |
| Package 重新导出 | 旧 Package | 保留旧版本，新增版本，不覆盖审计 |

## 错误语义

| 错误 | 行为 |
| --- | --- |
| 非法状态跳转 | 拒绝并返回当前状态、目标状态、缺失条件 |
| 状态与引用不一致 | 健康检查标记 warning 或 blocking |
| 完成状态缺少 audit | 阻断 closeout 类动作，要求补写或重建记录 |
| dirty 未处理直接导出 | 拒绝导出，要求重新确认上下文 |

## 恢复与重试

- PromptRun 从 `failed` 或 `cancelled` 重试时创建新 run，不覆盖旧 run。
- Package 处于 `invalid` 时可重新导出新版本，旧 Package 保留诊断状态。
- Take 误判为 rejected 时可恢复到 `pending_review`，但必须写 review audit。
- dirty Shot 重新确认上下文后回到 `context_ready`，关联 Package 仍需重新导出。

## 验收标准

- Shot 不能在缺少上下文时进入 `prompt_ready` 或 `package_ready`。
- PromptRun 失败保留错误类型和可重试性，不写成 completed。
- Package 上游变化后旧包标记 stale，重新导出创建新版本。
- Result 评审为 `needs_revision` 后，关联 Shot 回到可修订路径并保留评审原因。
