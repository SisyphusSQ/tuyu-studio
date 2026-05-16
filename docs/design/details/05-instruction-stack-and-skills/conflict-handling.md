# Conflict Handling

## 目标与边界

定义指令编译和任务运行前的冲突处理，包括锁定原则冲突、上下文缺失、输出不合法、上下文过大、用户确认和重试行为。本文不定义具体连续性规则字段。

## 输入 / 输出

输入：

- 编译请求、锁定规则、选中上下文、资产引用、输出契约和用户请求。

输出：

- `CompileConflict[]`。
- 用户确认请求。
- retry decision 和修复建议。

## 核心对象或规则

冲突类型：

| 类型 | 行为 |
| --- | --- |
| locked principle conflict | 阻断任务，展示规则和替代建议 |
| missing context | 阻断依赖该上下文的任务 |
| invalid output | 标记运行失败，保留原始输出 |
| overlarge context | 生成摘要或要求用户缩小选择 |
| missing asset | 阻断图像和交接类任务 |
| provider profile missing | 阻断 profile 相关输出 |

用户确认：

- 只有 warning、可忽略缺失资产、手动交接确认等场景可由用户确认继续。
- blocking locked conflict 不能通过普通确认绕过，必须修改请求或解锁规则。
- 用户确认必须写入 run record。

retry behavior：

- 可重试错误给出重试入口。
- 不可重试错误给出需要修复的对象。
- 重试创建新 attempt，不覆盖旧输出。

## 状态推进

```text
detected -> user_confirmation_required -> confirmed | cancelled
detected -> blocked
detected -> warning_recorded -> runnable
invalid_output -> failed -> retrying
```

- `blocked`：必须修复输入或规则。
- `warning_recorded`：允许继续，风险写入记录。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `conflict_locked_principle` | 用户请求违反锁定原则 | 阻断 |
| `conflict_missing_context` | 必需上下文缺失 | 阻断或要求补齐 |
| `conflict_invalid_output` | 输出不合法 | 标记 failed |
| `conflict_context_overlarge` | 上下文过大 | 摘要或要求缩小 |
| `conflict_confirmation_required` | 需要用户确认 | 进入 waiting_user |

## 恢复与重试

- 修改用户请求后重新编译。
- 补齐角色、场景、道具、资产或 profile 后重试。
- 解锁原则需独立确认和审计。
- 输出不合法可重跑，也可从原始输出手动提取。
- 上下文过大可减少选中对象或使用摘要策略。

## 验收标准

- 锁定原则冲突不会继续运行任务。
- 缺少必要上下文时展示具体缺失对象。
- warning 继续运行前写入记录。
- 输出不合法时不更新正式对象。
- 用户确认场景进入 `waiting_user` 并可取消。
- 重试保留旧 attempt 和错误记录。
