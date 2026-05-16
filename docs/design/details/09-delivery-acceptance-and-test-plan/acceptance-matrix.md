# 验收矩阵

## 目标与边界

定义主要能力在功能、数据、错误、恢复、审计和安全六个维度的验收行。本文用于验收设计完整性和实现切片，不替代测试用例。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 主要能力清单 | 六维验收矩阵 |
| 领域对象和状态 | 数据与状态验收 |
| 错误语义 | 错误和恢复验收 |
| 安全要求 | 审计和安全验收 |

## 核心对象或规则

| 能力 | 功能 | 数据 | 错误 | 恢复 | 审计 | 安全 |
| --- | --- | --- | --- | --- | --- | --- |
| 项目创建打开 | 新建、打开、恢复视口 | project、graph、settings 落盘 | 不可写或损坏有错误码 | 保留损坏副本并尝试最新备份恢复；无有效备份时只允许诊断、只读或拒绝打开 | project.create/open | 路径在项目根内 |
| Canvas-first 工作台 | 进入项目即显示 Project Canvas、左侧入口、Inspector、Bottom Bar 和任务健康 | ProjectCanvas viewport、theme、grid、selection 落盘 | 画布损坏显示诊断，不进入空白页 | 从备份或只读模式恢复画布 | canvas.open/theme/update | 不暴露本机绝对路径 |
| 暖浅 / 暗色画布 | Dark 与 Warm Light Canvas 都有独立背景、网格、节点、连线和状态色 | theme、grid.opacity、visual tokens | 主题 token 缺失时回退默认主题并提示 | 用户可重新选择主题和网格亮度 | canvas.theme/update | 不影响项目数据 |
| Graph 编辑 | 创建、移动、连线、删除 | 节点边 version 更新 | 非法边拒绝 | 撤销或依赖处理 | graph.* | 不写越界路径 |
| 脚本展开视图 | ScriptScene 可生成画布表格，行级生成 Shot 候选，可全屏和下载 | ScriptExpansionRow、sourceRange、candidate version | 缺镜号、时长、source range 或角色图引用时阻断行确认 | 保留候选，补字段、重新生成或手动拆场 | script.expand.* / command.* | 下载不含绝对路径或凭据 |
| Human 画布操作 | 人类直接编辑节点、Frame、Blueprint、Shot Card 和 Review | Studio Command、source=human | 命令校验失败不改画布 | 用户可撤销或重做安全变更 | command.human.* | 高风险动作需确认 |
| Agent 可见操作 | Agent 通过 Skill/CLI 创建、移动、连线、运行和写回候选 | Studio Command、source=agent、CommandResult | 越权、非法关系、输出失败写 Agent-visible error | 用户可接管、拒绝 candidate 或重试 | command.agent.* | 不能直接写文件或读越界路径 |
| 资产导入绑定 | 导入、预览、绑定 | Asset、摘要、缩略图 | 类型或缺失提示 | 重新定位或解绑 | asset.* | 外部文件受控复制 |
| 剧本分镜 | 拆场、建 Shot、校验 | ScriptScene、ShotCard | 缺字段阻断导出 | 补字段后继续 | graph/script 事件 | 正文作为 data |
| 指令预览 | 展示上下文和输出契约 | ContextBundle 摘要 | 上下文冲突阻断 | 修复后重新预览 | run.preview | 外部请求需确认 |
| 任务运行 | 创建 PromptRun 并产出 | run 记录、输出引用 | 输出校验失败不污染 | 新 run 重试 | run.* | 凭据不入记录 |
| ProviderMode | internal_provider 与 external_agent 都能发起运行并归一为 Run/Event/Audit | ProviderMode、CanvasSelection、events | unsupported mode 阻断并提示 | 切换模式或补齐配置后重试 | run.mode.* | 外发和网络需确认 |
| 成本 / 扣费占位 | UI 展示外部成本或积分占位 | estimatedCostLabel、creditPlaceholder | 不允许真实扣费或余额判断 | 隐藏或标注为占位 | run.cost.placeholder | 不保存支付或会员凭据 |
| 交接包 | 导出 manifest 和文件 | GenerationPackage 状态 | 引用缺失或路径拒绝 | 重新导出新版本 | package.* | 包内无逃逸路径 |
| 结果回收 | 导入结果并绑定 Shot | VideoResult、take、review | 错绑可提示 | 解绑、废弃、恢复 | result.* | 外部来源摘要脱敏 |
| 健康检查 | 列出风险和影响对象 | Health Report | 检查失败可读 | 修复后重跑 | project.health | 标记外部引用风险 |
| 迁移恢复 | 备份、迁移、回滚 | schema version 更新 | 迁移失败不破坏旧数据 | 回滚备份 | project.migrate | 迁移不泄露路径 |

## 状态推进

| 状态 | 推进规则 |
| --- | --- |
| not_covered | 尚无验收项 |
| specified | 六维矩阵已填写 |
| verified | 有测试、手动脚本或文档检查证明 |
| failed | 发现能力不满足矩阵 |
| accepted | 缺陷修复且证据更新 |

## 错误语义

| 错误 | 语义 |
| --- | --- |
| matrix_row_missing | 主要能力缺少验收行 |
| dimension_missing | 某能力缺少六维之一 |
| evidence_missing | 验收行没有验证证据 |
| status_mismatch | 实现状态与矩阵状态不一致 |

## 恢复与重试

- matrix_row_missing 先补能力行，再继续验收。
- dimension_missing 需要补齐功能、数据、错误、恢复、审计、安全维度。
- evidence_missing 通过测试、手动验收或文档检查补证据。
- failed 行保留失败原因，修复后更新为 verified。

## 验收标准

- 每个主要能力都有功能、数据、错误、恢复、审计、安全六维验收。
- 矩阵能直接映射到切片完成定义和测试策略。
- GenerationPackage.status 验收只使用 `draft`、`ready`、`handed_off`、`result_received`、`stale`、`invalid`。
- 任何安全拒绝和危险操作都有审计验收。
