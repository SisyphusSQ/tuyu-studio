# 健康检查

## 目标与边界

健康检查用于判断项目是否可打开、可编辑、可导出和可迁移。它不应自动删除用户数据；修复动作必须可解释、可预览，并写入 audit。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| `project.tuyu.json`、领域对象、asset metadata、Package manifest | HealthReport |
| 资产文件和 digest | 缺失、篡改或重复报告 |
| 审计与备份目录 | 恢复建议 |

HealthReport 至少包含：

```json
{
  "status": "blocking",
  "checkedAt": "2026-05-14T10:00:00+08:00",
  "items": [
    {
      "severity": "warning",
      "code": "asset_missing",
      "path": "assets/refs/ref_001.png",
      "affectedObjects": ["shot_001", "package_001"]
    }
  ]
}
```

## 核心对象或规则

健康检查以 HealthReport、check item、severity、affectedObjects 和 repair action 为核心对象。核心规则是每个问题都要能定位到文件、对象或关系；blocking 项阻断对应生产动作，warning 项允许继续编辑但必须在导出前显式处理。

## 状态推进

| 阶段 | 进入条件 | 输出 |
| --- | --- | --- |
| queued | 打开项目、导出 Package、迁移前或用户手动触发 | 检查任务 |
| scanning | 正在读取项目清单、图谱分区、资产、Package 和 audit | 中间发现项 |
| reported | 所有检查完成 | HealthReport 和严重级别汇总 |
| repair_pending | 用户选择修复项 | 修复计划和确认提示 |
| repaired | 修复动作成功 | repairs audit 和可重跑检查范围 |
| blocked | 存在未处理 blocking 项 | 阻断打开、保存、导出或迁移中的对应动作 |

## 检查项

| 检查 | 说明 | 典型级别 |
| --- | --- | --- |
| JSON validity | `project.tuyu.json`、领域对象、manifest 必须可解析；图谱分区必须存在 | blocking |
| graph refs | `GraphNode.refId`、edge source/target 必须存在或可标记占位 | warning/blocking |
| asset missing | asset index 或领域对象引用的文件不存在 | warning |
| digest mismatch | 文件存在但 digest 与记录不一致 | blocking |
| package references | manifest 引用的 Shot、Prompt、asset、结果文件缺失 | warning/blocking |
| absolute path leakage | 项目文件、manifest、Package 内含绝对路径或本机路径 | blocking |
| schema support | `project.tuyu.json.schemaVersion` 是否在支持范围 | blocking |
| lock state | stale 或 active lock 是否影响写入 | info/warning |

## 严重级别模型

| 级别 | 含义 | 是否阻断 |
| --- | --- | --- |
| blocking | 会导致项目无法安全打开、保存、迁移或导出 | 阻断对应动作 |
| warning | 项目可继续编辑，但部分对象、Package 或结果不可用 | 不阻断编辑，可能阻断导出 |
| info | 状态提示或可选清理建议 | 不阻断 |

## 修复动作

| 问题 | 修复动作 |
| --- | --- |
| JSON 无效 | 从 backups 恢复、保留损坏副本、输出诊断 |
| graph ref 缺失 | 创建 broken reference placeholder 或移除无效边 |
| asset missing | 重新定位文件、标记缺失、从 Package 重新导入 |
| digest mismatch | 阻断使用，允许重新计算并生成新 asset 版本 |
| Package 引用缺失 | 重新导出 Package 或标记旧 Package 不完整 |
| 绝对路径泄漏 | 拒绝导出，列出字段位置，要求改为相对路径 |
| stale lock | 用户确认后接管并写 audit |

## 错误语义

| 错误 | 用户提示 | 系统行为 |
| --- | --- | --- |
| health check incomplete | 检查未完成，不能给出安全结论 | 阻断导出和迁移 |
| repair failed | 修复失败，项目未被完全修复 | 保留原状态，写 repairs 事件 |
| affected object unknown | 找到问题但无法定位影响对象 | 提升为 warning，要求人工检查 |
| unsafe export | 存在 blocking 项，不能导出交接包 | 拒绝导出 |

## 恢复与重试

- 用户修复资产路径后，可只重跑受影响对象检查。
- 修复 graph ref 后，需要刷新节点状态和相关 Package 状态。
- digest mismatch 修复不能覆盖旧记录，应创建新版本或重新导入。
- 健康检查报告可保存到 `audit/`，但不得包含真实机器本地路径。

## 验收标准

- 删除一个被 Shot 和 Package 引用的资产后，健康检查能报告具体缺失路径、受影响 Shot、受影响 Package 和建议动作。
- JSON 损坏、digest mismatch、绝对路径泄漏会产生 blocking 结果。
- managed reference 缺失不会删除绑定关系，而是给出重新定位入口。
- 健康检查报告可用于判断项目是否可打开、可保存、可导出和可迁移。
