# 一致性规则

## 目标与边界

一致性规则确保图谱、领域对象、资产和生产状态之间不会静默漂移。所有删除、修改、断链和 dirty 标记都必须可解释，并能给出受影响对象列表。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 删除 Node、删除领域对象、删除 Asset、删除 Result 或 Studio Command 请求 | ImpactReport、CommandResult 和确认决策 |
| Character、Scene、Prop、Style、Asset、Prompt 变更 | dirty / stale 标记和受影响对象列表 |
| 健康检查发现的断链 | placeholder、修复建议和 audit 记录 |

## 核心对象或规则

一致性以 Studio Command、ImpactReport、broken reference placeholder、dirty 标记和 stale Package 为核心对象。规则是所有 UI、CLI、MCP 风格入口和外部 Agent 都先通过命令校验；高风险操作先报告影响再执行，先保留可恢复证据再清理引用，任何连续性输入变化都必须传播到引用它的 Shot 和 Package。

## Studio Command 规则

| 规则 | 要求 |
| --- | --- |
| 单一写入口 | `create_node`、`move_node`、`connect_nodes`、`insert_blueprint`、`start_run`、`bind_result` 等写操作只能通过 Studio Command |
| 来源标记 | 每条命令必须记录 `source=human/agent/system/imported`、操作者、入口、时间和 requestId |
| 版本校验 | 命令必须携带期望的 ProjectCanvas version，版本过旧时拒绝覆盖关系变化 |
| 影响预览 | 删除、覆盖、外发、运行和结果绑定前必须生成影响摘要 |
| 可见失败 | Agent 命令失败时写入 Agent-visible error event，UI 显示未修改画布 |
| 审计落盘 | 命令、结果、错误和恢复动作写入 audit，不能只存在前端 store |

## 删除 Node 与删除领域对象

| 操作 | 默认行为 | 需要确认的情况 |
| --- | --- | --- |
| 删除 Node | 只移除画布节点和相关边 | Node 有 `refId` 时提示领域对象仍保留 |
| 删除领域对象 | 先生成影响报告 | 存在引用 Node、Package、PromptRun、Result 时必须确认 |
| 删除 Asset | 默认不直接删除引用记录 | 存在 Shot 或 Package 引用时必须确认 |
| 删除 Package | 标记废弃或删除 Package 节点 | 已有 Result 关联时必须确认 |
| 删除 Result / Take | 默认标记 rejected 或 archived | 物理删除文件需要二次确认 |

## 级联影响报告

删除或修改高影响对象前必须生成 ImpactReport：

```json
{
  "operation": "delete_domain_object",
  "target": { "kind": "character", "id": "char_001" },
  "affected": {
    "nodes": ["node_001"],
    "edges": ["edge_001"],
    "shots": ["shot_001", "shot_002"],
    "packages": ["package_001"],
    "results": ["result_001"]
  },
  "blockingReasons": ["package_ready_depends_on_target"]
}
```

| 影响项 | 报告内容 |
| --- | --- |
| nodes | 引用该对象的 GraphNode |
| edges | source 或 target 受影响的 GraphEdge |
| shots | 使用该对象或资产的 Shot |
| prompts | digest 输入包含该对象的 Prompt |
| packages | manifest 引用该对象或资产的 Package |
| results | 通过 Package 或 Shot 间接受影响的 Take |

## 断链占位

| 场景 | 占位规则 |
| --- | --- |
| `refId` 缺失 | Node 保留，显示 broken reference placeholder |
| Edge target 缺失 | Edge 标记 invalid，不参与上下文解析 |
| Asset 缺失 | 保留 asset metadata，显示 missing asset |
| Package 引用缺失 | Package 标记 invalid 或 stale |
| Result 文件缺失 | Take 保留评审记录，文件状态为 missing |

占位对象不能进入正式 Package 导出，但可以帮助用户恢复、重新定位或清理。

## 连续性变更 dirty 标记

| 变更 | dirty 目标 | 规则 |
| --- | --- | --- |
| Character 外观、服装、禁改项 | 引用该 Character 的 Shot、Prompt、Package | Shot 标记 `context_dirty`，Package 标记 `stale` |
| Scene 光线、地点、氛围 | 所属 Shot 和 Package | 重新确认前不可导出新包 |
| Prop 外观或用途 | 使用该 Prop 的 Shot、Package | 影响列表写入 audit |
| Style 规则 | 依赖该 Style 的 Shot、Prompt、Package | 批量标记，并允许用户逐项确认 |
| Asset digest 变化 | 引用 asset 的对象 | blocking，要求重新定位或重新导入 |

## 状态推进

| 阶段 | 条件 | 结果 |
| --- | --- | --- |
| impact_pending | 用户请求删除或高影响修改 | 生成 ImpactReport，等待确认 |
| confirmed | 用户选择保留、占位、废弃或物理删除 | 执行对应变更并写 audit |
| placeholder_created | 引用目标缺失但需要保留上下文 | 对象保留断链状态，阻断正式导出 |
| dirty_marked | 连续性或 digest 输入变化 | Shot 标记 dirty，Package 标记 stale |
| repaired | 用户重新定位、恢复备份或重新绑定 | 清除对应断链或 dirty 标记 |

## 错误语义

| 错误 | 行为 |
| --- | --- |
| impact report failed | 阻断删除或高影响修改 |
| command_validation_failed | 拒绝命令，不写入 ProjectCanvas |
| command_source_untrusted | Agent 或 CLI 来源未授权，写安全审计 |
| cascade blocked | 有 Package 或 Result 依赖时，要求用户选择保留占位或废弃 |
| dirty mark failed | 阻断上游保存，避免状态不一致 |
| placeholder unresolved | 健康检查 warning；导出时可能升级为 blocking |
| physical delete unsafe | 拒绝物理删除，建议 archive 或 rejected |

## 恢复与重试

- broken reference 可通过重新选择领域对象、恢复备份或清理 Node 修复。
- missing asset 可重新定位；digest 不一致时必须作为新版本处理。
- dirty Shot 重新确认上下文后可回到 `context_ready`，但旧 Package 不自动恢复为 ready。
- 删除操作取消后，不应留下半删除边或状态变化。

## 验收标准

- 删除 Node 不会误删 Character、Scene、Shot、Prompt 或 Result 等领域对象。
- 删除领域对象前必须列出受影响 Node、Edge、Shot、Package 和 Result。
- 断链对象以 placeholder 保留，便于恢复，不静默消失。
- 修改角色、场景、道具或风格连续性后，引用它们的 Shot 被标记 dirty，相关 Package 标记 stale。
