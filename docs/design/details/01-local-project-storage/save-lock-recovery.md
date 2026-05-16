# 保存、锁与恢复

## 目标与边界

保存系统要保证结构化数据不会因为进程退出、磁盘错误或并发打开而被半写破坏。项目锁用于提示并发风险，不用于替代文件系统权限或远程协作控制。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 待保存的 Project manifest、领域对象、Package 或 Review | 原子替换后的正式文件和保存 audit |
| 自动保存事件、用户显式保存、任务状态变化 | 节流后的持久化结果和 dirty 状态 |
| `session.lock`、`write.lock`、`lastCleanShutdown` | 项目打开模式、接管决策和恢复检查报告 |

## 核心对象或规则

保存链路以正式文件、临时文件、锁文件和 audit 事件为核心对象。规则是先校验再写 tmp，rename 成功后才推进生产状态；任何保存失败都必须保留上一版有效文件，并让用户看到可重试或恢复路径。

## 原子写序列

```text
1. serialize: 在内存中生成完整新内容
2. validate: 校验 JSON、schema、路径和必要引用
3. write tmp: 写入 {file}.tmp.{instance_id}
4. fsync tmp: 刷新临时文件
5. rename: 原子替换目标文件
6. fsync dir: 刷新父目录
7. audit: 记录保存事件、版本和 digest
8. cleanup: 清理本实例遗留 tmp
```

| 文件 | 策略 |
| --- | --- |
| `project.tuyu.json` | 原子写，保存前校验 schema；图谱分区 `version` 每次递增 |
| 领域对象 JSON | 单对象原子写 |
| `audit/*.jsonl` | append-only，失败时写 fallback error |
| 大型资产 | 先写 staging，再完成 digest 后移入目标目录 |

## 自动保存

| 数据 | 触发 | 节流 | 失败处理 |
| --- | --- | --- | --- |
| 项目清单的图谱分区 | 节点移动、缩放、连线、删除 | 500-1000ms 合并保存 | 保留 dirty 状态并提示重试 |
| Shot / Character / Scene / Prop | 表单保存、失焦、显式保存 | 即时或短延迟 | 表单保留未保存标记 |
| PromptRun | 状态变化、输出片段、失败事件 | 事件驱动 | 写 fallback error 文件 |
| Package manifest | 导出完成前统一写入 | 不节流 | 失败则 Package 不进入正式状态 |
| Review | 用户提交评审结论 | 即时 | 阻断状态推进，保留输入 |

## 状态推进

### 项目锁状态

| 状态 | 检测 | 行为 |
| --- | --- | --- |
| clean open | 无 `session.lock` 或锁属于当前实例 | 正常读写 |
| stale lock | 锁文件存在，但 pid 不存在或心跳过期 | 提示恢复并允许接管 |
| active lock | 锁文件显示其他实例仍活跃 | 默认只读打开 |
| takeover | 用户确认接管 active 或 stale 项目 | 写 takeover audit，替换锁 |

`session.lock` 示例：

```json
{
  "appInstanceId": "instance_001",
  "openedAt": "2026-05-14T10:00:00+08:00",
  "heartbeatAt": "2026-05-14T10:05:00+08:00",
  "pid": 12345,
  "host": "local-machine"
}
```

### 保存状态

| 状态 | 进入条件 | 下一步 |
| --- | --- | --- |
| dirty | 内存数据相对磁盘有变化 | 自动保存或显式保存 |
| validating | 保存前 schema、路径、引用检查中 | 通过后写 tmp，失败则保留 dirty |
| writing_tmp | 临时文件写入中 | fsync tmp |
| committed | rename 和父目录 fsync 成功 | 写 audit 并清理 tmp |
| recoverable_error | rename、fsync 或 audit 失败 | 保留旧文件，进入恢复检查 |

## 崩溃恢复

当 `project.tuyu.json.integrity.lastCleanShutdown=false` 时，打开项目必须进入恢复检查：

1. 扫描临时文件、staging asset、未完成 Package 和 fallback error。
2. 校验 `project.tuyu.json`、图谱分区和领域对象 JSON。
3. 对比 `lastGraphVersion` 与 `graph.version`。
4. 展示可恢复项：保留上一版、恢复临时写入、清理废弃临时文件。
5. 恢复完成后写 `audit/repairs.jsonl` 并更新 clean 状态。

## 错误语义

| 错误 | 用户提示 | 系统行为 |
| --- | --- | --- |
| write validation failed | 保存内容未通过校验 | 不写 tmp，不改变旧文件 |
| rename failed | 保存替换失败 | 保留旧文件和 tmp，提示重试 |
| fsync failed | 保存可能未落盘 | 标记项目需要恢复检查 |
| active lock | 项目可能在另一实例打开 | 只读或确认接管 |
| dirty shutdown | 上次未正常关闭 | 打开前执行恢复检查 |

## 恢复与重试

- 原子写的 tmp 只在校验通过后才能被恢复为正式文件。
- 未完成 Package 目录保留为 `draft`；对应 Shot 不推进到 `package_ready`。
- staging asset 若 digest 完整可重新入库，否则清理并提示重新导入。
- active lock takeover 必须写入 audit，便于后续排查并发写入。

## 验收标准

- 任意结构化文件写入中断后，上一版有效文件仍可解析并用于打开项目。
- `lastCleanShutdown=false` 时，项目打开前显示恢复检查结果，不静默跳过。
- stale lock 可接管，active lock 默认只读。
- 自动保存失败不会清除用户未保存输入，也不会推进生产状态。
