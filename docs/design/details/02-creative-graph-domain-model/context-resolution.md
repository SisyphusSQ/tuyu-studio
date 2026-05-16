# 上下文解析

## 目标与边界

上下文解析把选中节点、相邻关系和领域对象编译成任务可用的 ContextBundle。解析结果必须可预览、可裁剪、可生成 digest，并能解释每段上下文来源。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| selected node ids、selected frame ids、agent command selection | ContextBundle |
| GraphDocument edges | 上下文来源列表和依赖路径 |
| 领域对象与 asset metadata | 任务输入、digest、缺失报告 |
| ProviderProfile 和任务模板 | 输出契约和格式限制 |

## 核心对象或规则

上下文解析以 selected node、selected frame、traversal profile、ContextBundle、context digest 和 preview 为核心对象。核心规则是按节点类型限定遍历深度，按任务目标裁剪字段，所有参与 digest 的输入都必须可追溯到来源对象、引用路径和来源入口。Agent 只能读取 Studio Command 授权的选中上下文，不能扫描项目外路径。

## 状态推进

| 阶段 | 进入条件 | 输出 |
| --- | --- | --- |
| selected | 用户或 Agent 选中节点 / Frame 并发起任务 | 起点节点、Frame 和任务类型 |
| expanded | 根据 traversal profile 加载相邻对象 | 上下文候选集合 |
| compiled | 合并项目规则、领域对象、资产和历史 | ContextBundle 草稿 |
| previewed | 生成来源、优先级和 digest 预览 | 用户确认或裁剪请求 |
| accepted | 用户确认上下文 | PromptRun 输入或 Package 导出输入 |
| waiting_user | 上下文过大、缺失或冲突 | 裁剪、补充或冲突处理入口 |

## 选中节点展开规则

1. 读取选中节点和直接相邻边。
2. 按节点 kind 加载 `refId` 对应领域对象。
3. 根据任务类型选择 traversal profile，不做无限图遍历。
4. 合并项目原则、Style、Scene、Character、Prop、Shot、Prompt、Package、Result 历史。
5. 给每个上下文片段标注 source、priority、digest input flag。
6. 输出预览，等待用户确认后进入 PromptRun 或 Package 导出。

## 画布选中上下文

| 选择范围 | 上下文边界 | Agent 可见性 |
| --- | --- | --- |
| 单个节点 | 节点摘要、refId 领域对象、直接相邻边 | 可读，需显示来源和 digest |
| 多选节点 | 共同父 Frame、共有上下文、差异列表 | 可读，不自动合并冲突字段 |
| ReferenceGroup | 输入节点、引用角色、用户 notes | notes 作为 data 包装，不提升为系统指令 |
| ProductionFrame | reference groups、output nodes、taskIntent、historySummary | 可读，可请求 start_run |
| Blueprint 候选 | 待插入节点、边、输入槽和输出契约 | 插入前为 candidate，用户可拒绝 |

## 遍历深度

| 起点 | 默认展开 | 最大深度 | 说明 |
| --- | --- | --- | --- |
| Shot | 所属 Scene、使用的 Character / Prop / Style、绑定 Prompt、最近 Package / Result | 2 | Shot 是镜头任务主入口 |
| Character | 角色 profile、参考资产、引用该角色的 Shot 摘要 | 2 | 不展开所有 Result 原文 |
| Scene | 场景 profile、关联 Shot、Style、关键资产 | 2 | 只取相关镜头摘要 |
| Prop | 道具 profile、出现 Shot、参考资产 | 2 | 防止把无关场景带入 |
| Prompt | 所属 Shot、PromptRun 历史、输出契约 | 2 | 旧版本通过 `refines` 追加摘要 |
| Package | manifest、Shot、Prompt、引用资产、交接状态 | 2 | 用于重导出或结果匹配 |
| Result | Take、Review Status、Package、Shot、修订说明 | 3 | 需要追溯到生成上下文 |

## context digest 输入

| 输入 | 是否参与 digest | 说明 |
| --- | --- | --- |
| Shot 核心字段 | 是 | 画面、动作、运镜、时长、台词 |
| Character / Scene / Prop 连续性字段 | 是 | 影响一致性 |
| Style 规则 | 是 | 影响画面语言 |
| asset digest | 是 | 文件内容变化必须触发 dirty |
| Prompt 文本 | 是 | 影响 Package 和 Result 追溯 |
| ProviderProfile 能力声明 | 是 | 影响输出格式 |
| Node 坐标和视口 | 否 | 只影响 UI |
| Note | 默认否 | 用户可手动提升为普通上下文，但不能作为高优先级规则 |

## 过大上下文处理

| 情况 | 行为 |
| --- | --- |
| 文本超过任务限制 | 生成裁剪建议，按优先级保留必需字段 |
| 资产过多 | 要求用户选择关键参考或生成资产清单摘要 |
| 历史 PromptRun 过多 | 只纳入最近有效版本和用户标记版本 |
| Result 历史过多 | 纳入当前推荐 Take、被拒绝原因摘要和最近修订说明 |
| 必需上下文被裁掉 | 阻断运行，要求用户确认替代输入 |

## 错误语义

| 错误 | 行为 |
| --- | --- |
| missing required ref | 阻断运行，列出缺失对象 |
| selection_scope_denied | Agent 读取未授权选择范围，拒绝并写 audit |
| broken asset | 阻断 Package 导出，可继续编辑 |
| context too large | 进入 `waiting_user`，提供裁剪选项 |
| digest unavailable | 阻断正式运行或导出 |
| conflicting rules | 展示冲突来源，要求用户选择优先级 |

## 恢复与重试

- 用户裁剪上下文后，保存裁剪决策并写入 PromptRun 输入摘要。
- 缺失资产重新定位后，只重算受影响 ContextBundle。
- 冲突规则解决后，重新生成 digest，旧 digest 不覆盖。

## 验收标准

- 选中 Shot 可解析出所属 Scene、引用 Character / Prop / Style、Prompt、Package 和 Result 摘要。
- ContextBundle 预览能展示来源、优先级、是否参与 digest。
- 上下文过大时不会静默截断必需字段，而是进入用户确认路径。
- 任一 digest 输入变化后，关联 Shot 或 Package 可被标记 dirty / stale。
