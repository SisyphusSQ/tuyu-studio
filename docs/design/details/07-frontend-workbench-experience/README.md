# 07. Canvas-first 桌面工作台体验设计

## 子文档

| 文档 | 作用 |
| --- | --- |
| `workspace-layout.md` | Top Bar、左侧入口、Project Canvas、Inspector、Bottom Bar、Dark / 暖浅画布 |
| `canvas-interactions.md` | 节点创建、Blueprint、连线、框选、拖拽、搜索、双层 Frame |
| `inspector-and-task-panels.md` | 属性、关联、连续性、任务、运行记录、审计 |
| `production-feedback.md` | 保存失败、任务失败、缺失资产、上下文过期、导出状态 |

本 README 只作为模块总览；详细规则、清单和矩阵以子文档为准。

## 目标

桌面工作台是创作者每天反复使用的生产界面。它要安静、密集、可扫描，打开项目后第一屏就是 Project Canvas，服务剧本整理、节点编排、资产绑定、提示词编译、Agent 共创、运行、结果 review 和必要的 handoff，不做营销式页面，也不把常用操作藏进说明文字。

## 当前前端落地栈

当前产品落地前端采用 Vue 3 + TypeScript + Vite + Ant Design Vue + AntV，并通过 Wails 生成的 TypeScript binding 调用 Go 后端服务。详细技术分层见 `../../architecture/technology-stack.md`。

| 技术 | 工作台职责 |
| --- | --- |
| Vue 3 + TypeScript | 页面、组件、面板、交互状态和类型化调用 |
| Ant Design Vue | 布局、菜单、表单、表格、抽屉、弹窗、通知、上传、进度和状态控件 |
| AntV G6 | Project Canvas / Creative Graph 画布、节点、边、生产组、参考组和图谱交互 |
| AntV G2 | 生产状态、模块覆盖、健康检查摘要等图表 |
| Wails Binding | 前端命令、查询和事件订阅的类型化 bridge |

前端不直接读写本地文件，也不把 Pinia 或组件状态当作领域 truth。所有项目、图谱、资产、运行记录和交接包的持久数据都通过 Go service DTO 返回。

### 前端分层约定

| 层 | 职责 | 边界 |
| --- | --- | --- |
| View / Page | 路由入口、页面布局、把用户动作组织成 command | 不直接 import Wails generated binding |
| Feature Components | GraphCanvas、AssetShelf、ShotInspector、RunQueue、ProviderSettings、PackageExportPanel 等局部交互 | 可以调用 composable 或 API wrapper，不保存领域 truth |
| API Wrapper | `projectApi`、`graphApi`、`assetApi`、`runtimeApi`、`providerApi`、`packageApi` | 唯一可以调用 `wailsjs/go/...` 的前端层 |
| UI Session Store | 当前选择、面板折叠、过滤器、排序、画布视口、运行队列投影、错误抽屉 | 不保存 Project、Graph、Asset、PromptRun、Package、Result 的正式状态 |
| Shared Components | StatusChip、ErrorNotice、ActionList、EmptyState、ConfirmAction、RelativePathText | 只负责展示和基础交互，不知道领域内部实现 |

Pinia 或等价 store 只保存 UI session。项目重开、页面刷新或事件断连后，前端必须从 Go DTO 和 run records 重新构建 UI 投影。

### Ant Design Vue 与 AntV 使用约定

| 类型 | 优先技术 | 使用范围 |
| --- | --- | --- |
| Shell | Ant Design Vue Layout / Menu / Tabs | Top Bar、Left Panel、Bottom Panel、视图切换 |
| 编辑 | Ant Design Vue Form / Input / Select / Switch / Radio / Checkbox | 领域对象、provider profile、项目设置 |
| 列表 | Ant Design Vue Table / Tree / List | 资产、镜头、运行记录、审计、健康项 |
| 确认与反馈 | Ant Design Vue Drawer / Modal / Popconfirm / Alert / Notification / Progress | Inspector、危险确认、长任务、错误提示 |
| 图谱 | AntV G6 | Creative Graph、ProductionFrame、ReferenceGroup、连线、布局和状态徽标 |
| 统计 | AntV G2 | 生产状态、健康趋势、模块覆盖、运行耗时、provider 能力覆盖 |

表单、列表、确认和反馈优先使用 Ant Design Vue，不手写一套类似控件。G6 只负责画布交互状态：layout、viewport、selection、hover、drag preview 和 temporary edge；正式节点、边、生产组、状态徽标都来自 `GraphViewDTO`。G2 用于统计图，不替代表格或可操作列表。

### 错误与状态展示约定

| 场景 | 表达 |
| --- | --- |
| blocking error | Alert + recovery actions |
| field validation | Ant Form item error |
| async task | RunQueue item + Progress |
| capability missing | Tag / Badge + disabled action + tooltip |
| destructive action | Popconfirm 或 Modal，必要时要求 audit reason |
| credential missing | ProviderSettings CTA，不在业务页面输入 API key |

长标题、长路径和长错误详情必须可折叠、可复制或摘要显示。绝对路径默认脱敏或摘要显示。主要操作不能只藏在 hover 状态里；任务失败必须给下一步动作。

## 信息架构

```text
┌──────────────────────────────────────────────────────────────┐
│ Top Bar: 项目 / 保存状态 / 运行队列 / 导出 / 设置               │
├──────────────┬──────────────────────────────┬────────────────┤
│ Left Panel   │ Creative Graph Canvas         │ Inspector      │
│ 资产 / 剧本   │ 节点、边、生产组、状态徽标       │ 属性、上下文、任务 │
├──────────────┴──────────────────────────────┴────────────────┤
│ Bottom Bar: 当前选择 / 健康检查 / 任务事件 / 错误摘要           │
└──────────────────────────────────────────────────────────────┘
```

| 区域 | 主要用途 | 设计要求 |
| --- | --- | --- |
| Top Bar | 项目切换、保存状态、运行队列、导出、设置 | 高度稳定，不因项目名过长导致布局跳动 |
| Left Panel | 资产库、剧本、场次、搜索、过滤 | 支持列表和缩略图，提供拖拽到画布 |
| Canvas | Project Canvas 主工作区 | 平移、缩放、选择、连线、框选、Blueprint、生产组、撤销重做 |
| Inspector | 节点属性、上下文、连续性、任务、预览 | 选中对象变化时快速切换，不丢失未保存编辑 |
| Bottom Bar | 状态、错误、健康检查、任务进度 | 只显示可行动信息，点击可展开详情 |

## 主要视图

| 视图 | 入口 | 作用 |
| --- | --- | --- |
| Project Canvas | 默认 | 编排节点、边、Frame、Blueprint、上下文、运行和结果 |
| Script View | 左侧剧本 tab 或 ScriptNode 打开 | 长文本编辑、拆场、创建镜头候选 |
| Asset View | 左侧资产 tab | 导入、过滤、绑定、预览、缺失修复 |
| Run View | 运行队列 / PromptRun 打开 | 查看编译指令、事件流、输出和错误 |
| Package View | PackageNode 打开 | 检查 manifest、提示词、引用图、上传清单 |
| Review View | VideoResultNode 打开 | 播放结果、标记通过/需修改/废弃、写 review |
| Project Health View | Bottom Bar 打开 | 查看断链、缺失资产、schema、外部引用风险 |

## 视觉模式

| 模式 | 画布底色 | 网格 | 节点层级 | 状态要求 |
| --- | --- | --- | --- | --- |
| Dark Canvas | 深色低眩光背景 | 点阵网格可调亮度，默认中低对比 | 面板和节点浮在画布上，边界清晰 | 风险、成功、进行中、阻断状态必须可辨认 |
| Warm Light Canvas | 偏暖、低眩光浅色背景，不使用纯白 | 点阵网格可调亮度，默认更柔和 | 节点阴影更轻、边框更明确 | 不简单反色，需单独定义连线、选中、hover、active |

## 画布交互

### 基础操作

| 操作 | 行为 |
| --- | --- |
| 拖入文件 | 创建 Asset，必要时创建对应节点 |
| 双击空白 | 打开节点创建菜单 |
| 插入 Blueprint | 创建一组节点、边、Frame、输入槽和输出契约 |
| 拖动节点 | 更新布局并自动保存 |
| 拖出连接线 | 打开关系选择器，只显示合法关系 |
| 框选 | 批量移动、批量导出、批量运行检查 |
| 节点搜索 | 定位节点并临时高亮相关边 |
| 折叠节点 | 节点保留摘要和状态徽标 |
| 生产组 | 组织可运行的生产上下文，包含参考组、输出节点和历史摘要 |

### 双层 Frame

工作台采用双层组模型：

- 外层 `ProductionFrame` 是可运行生产组，承载任务意图、输出节点、能力要求和历史摘要。
- 内层 `ReferenceGroup` 是参考输入组，只组织视频、图片、文本等输入节点，并记录输入角色、优先级和备注。
- 输出节点在生产组内独立展示，不并入参考组；同一生产组可以根据 provider 能力输出文本、图片或视频。
- 历史轨只默认展示当前版本、收藏版本和最近成功版本；失败、取消、草稿进入完整运行详情。
- 缺少图片理解、视频理解或视频生成能力时，不删除节点、不破坏生产组，只禁用或降级对应任务入口。

### 节点视觉

| 节点 | 必显信息 | 状态徽标 |
| --- | --- | --- |
| Script | 标题、场次数、更新时间 | 未拆场、已拆场、需确认 |
| Character | 名字、主参考图、锁定规则数 | 缺参考、已锁定、冲突 |
| Scene | 场景名、地点、光线 | 缺参考、已绑定 |
| Prop | 道具名、用途 | 未绑定、已使用 |
| Shot | 编号、时长、状态、主角色 | 缺字段、可运行、已导出、需修改 |
| Prompt | 版本、来源、评分 | 手动、AI 生成、schema 风险 |
| Package | 版本、目标 profile、导出状态 | 可检查、缺文件、已交接 |
| VideoResult | take、review 状态、时长 | 待看、通过、需修改、废弃 |

节点尺寸要有最小宽高，长标题折行或省略，不能撑破画布布局。

## Inspector 设计

Inspector 采用分区式结构：

```text
Header: 节点类型 / 标题 / 状态 / 关键操作
Tabs:
  - 属性
  - 关联
  - 连续性
  - 任务
  - 运行记录
  - 审计
```

| Tab | 内容 |
| --- | --- |
| 属性 | 领域对象字段编辑，支持保存、撤销、缺失提示 |
| 关联 | 入边、出边、引用资产、相关镜头、包和结果 |
| 连续性 | 锁定规则、冲突、警告、修复建议 |
| 任务 | 可执行 AI 任务、编译预览、运行按钮、输出契约 |
| 运行记录 | PromptRun 列表、contextDigest、输出、错误 |
| 审计 | 用户操作、状态变更、导出、回收和 review |

未保存编辑规则：

- 切换节点时，如果当前表单有未保存更改，提示保存、丢弃或留在当前节点。
- 自动保存失败时，字段显示失败状态，不丢失用户输入。
- 多 tab 编辑同一对象时，以领域对象版本号检测冲突。

## 任务操作体验

AI 任务入口不应只是一个泛化按钮，而是根据当前节点和上下文显示可执行命令：

| 当前选择 | 可用命令 |
| --- | --- |
| ScriptNode | 拆场、生成镜头候选、摘要、检查人物/道具 |
| ShotNode | 优化提示词、生成参考图、连续性检查、导出交接包 |
| CharacterNode | 生成角色圣经、检查参考图一致性、创建表情/姿态需求 |
| SceneNode | 生成场景说明、检查镜头引用、补充光线/氛围 |
| 多选 Character + Scene + Shot | 融合上下文、生成镜头提示词、生成参考图 |
| ProductionFrame | 优化提示词、生成图片、生成视频、分析分镜候选、写回确认 |
| PackageNode | 检查包完整性、重新导出、打开包目录 |
| VideoResultNode | review、创建修改任务、生成下一版建议 |

每个任务运行前显示：

- 将使用的节点和资产。
- 将使用的能力包、ProviderProfile 和具体 capability。
- 输出位置和输出 schema。
- 是否需要联网或外部权限。
- 可能阻断任务的冲突。

任务入口按能力分级：

| 状态 | 表达 | 行为 |
| --- | --- | --- |
| 可运行 | 按钮可用，显示预计输出 | 直接创建 PromptRun 或 FrameRun |
| 可预览 | 可生成提示词或交接包，但不能自动提交 | 提供复制、下载或手动导入路径 |
| 可降级 | 缺少图片/视频理解能力，但文本上下文可用 | 保留参考节点，要求用户补充描述或改用文本任务 |
| 不可运行 | 缺少必需能力、资产或权限 | 禁用入口并给出修复动作 |

## 资产面板

资产面板支持：

- 按类型、角色、绑定状态、来源、缺失状态过滤。
- 缩略图网格和紧凑列表切换。
- 拖拽资产到节点建立绑定。
- 批量选择并创建角色、场景、道具或参考组。
- 缺失资产重新定位。
- 查看 digest、来源、绑定对象和使用次数。

资产不可只按文件夹浏览，必须以项目语义组织。

## 生产反馈

| 反馈 | 表达 |
| --- | --- |
| 已保存 | Top Bar 显示稳定保存状态和最近保存时间 |
| 保存失败 | Bottom Bar 显示错误，字段保持未保存标记 |
| 任务运行 | 运行队列显示 task、进度、取消、查看详情 |
| 任务失败 | 节点徽标 + Bottom Bar 错误，可打开 error.json |
| 缺失资产 | 节点角标 + Health View 列表 |
| 上下文已过期 | Shot 或 Prompt 显示 `context_dirty` |
| 历史轨更新 | 生产组侧栏只更新当前、收藏、最近成功版本 |
| 导出完成 | PackageNode 状态变更，可打开包目录 |
| Agent 操作 | 节点来源标记、命令事件、候选状态、接管入口同步出现 |

## 可访问性与键盘

| 能力 | 要求 |
| --- | --- |
| 键盘导航 | 画布对象、面板 tab、表单控件可键盘访问 |
| 对比度 | 状态颜色不作为唯一信息，必须有文本或图标 |
| 图标按钮 | 图标按钮有 tooltip 和可访问名称 |
| 长文本 | 剧本、提示词、错误详情使用可复制文本区域 |
| 窄屏 | 保证面板可折叠，核心任务不被遮挡 |
| 误操作 | 删除、覆盖、接管锁、清理回收区必须二次确认 |

## 验收标准

- 默认进入就是 Graph View，不出现营销落地页。
- 用户可在不读说明的情况下完成：导入资产、创建 Shot、优化提示词、导出包、回收结果。
- 任意节点打开 Inspector 后都能看到属性、关联、任务、运行记录和审计入口。
- 任务运行前能预览上下文、能力包、Provider capability、输出契约和风险。
- 生产组能同时展示内层参考组、输出节点、历史轨和任务抽屉。
- 缺少媒体理解或视频生成能力时，画布保留节点和 Frame，只禁用、降级或转入手动交接。
- 保存失败、缺失资产、任务失败都有可行动入口。
- 长标题、长项目名、长错误信息不会撑破按钮、面板或节点。
- 常用操作有键盘路径和清晰撤销/确认机制。
