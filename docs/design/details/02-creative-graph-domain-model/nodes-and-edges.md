# 节点与边

## 目标与边界

Project Canvas 用 ProjectCanvas、GraphNode、GraphEdge、ProductionFrame、ReferenceGroup 和 CanvasBlueprintInstance 表达创作对象、上下文、任务和结果之间的生产关系。Creative Graph 是 Project Canvas 内的结构化关系层，只保存画布结构和展示摘要，领域对象的完整内容由各自目录负责。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 节点创建请求、领域对象 ID、画布位置、来源入口 | GraphNode、可选领域对象引用和 source event |
| 连线请求、source、target、relation | 通过矩阵校验的 GraphEdge |
| 删除、移动、折叠、状态刷新事件 | 更新后的 ProjectCanvas version 和 audit 摘要 |
| Blueprint 插入请求 | 一组节点、边、Frame、输入槽和恢复语义 |

## 核心对象或规则

ProjectCanvas 是画布保存单元，GraphNode 是画布对象，GraphEdge 是生产关系。核心规则是 Canvas 只保存关系和展示摘要，领域对象完整内容通过 `refId` 读取；任何会影响生产上下文的边都必须经过 Studio Command 和合法关系矩阵校验。PromptRun 不是 GraphNode，也不是 GraphEdge endpoint；运行来源由 Prompt、Package、Result 等领域对象字段和 run records 引用。

ProjectCanvas 同时保存双层组的画布结构：`ReferenceGroup` 是内层输入组，`ProductionFrame` 是外层可运行生产组。两者都是图谱上下文和布局对象，不是 Shot、Prompt、Asset 或 Package 的领域真相。FrameRun 是用户可理解的历史入口，可引用 PromptRun、provider attempt 和输出 Asset，但不作为 GraphNode 或 GraphEdge endpoint。

```ts
type CanvasNodeCategory =
  | "source"
  | "concept"
  | "continuity"
  | "production"
  | "output"
  | "review"
  | "organizer"
  | "handoff"

type NodeSource = "human" | "agent" | "system" | "imported"

interface ProjectCanvas {
  id: string
  projectId: string
  schemaVersion: string
  version: number
  viewport: { x: number; y: number; zoom: number }
  theme: "dark" | "warm_light"
  grid: { visible: boolean; size: number; opacity: number }
  nodes: GraphNode[]
  edges: GraphEdge[]
  frames: ProductionFrame[]
  referenceGroups: ReferenceGroup[]
  blueprintInstances: CanvasBlueprintInstance[]
  selectedNodeIds: string[]
  updatedAt: string
}

interface GraphNode {
  id: string
  kind: GraphNodeKind
  category: CanvasNodeCategory
  title: string
  refId?: string
  x: number
  y: number
  width: number
  height: number
  status?: string
  badges?: string[]
  source: NodeSource
  sourceEventId?: string
  data?: Record<string, unknown>
}

interface GraphEdge {
  id: string
  sourceNodeId: string
  targetNodeId: string
  relation: GraphEdgeRelation
  createdAt: string
}

interface ReferenceGroup {
  id: string
  title: string
  inputNodeIds: string[]
  role: string
  priority: number
  notes?: string
}

interface ProductionFrame {
  id: string
  title: string
  referenceGroupIds: string[]
  outputNodeIds: string[]
  taskIntent: string
  requiredCapabilities: string[]
  historySummary: {
    currentRunId?: string
    favoriteRunIds: string[]
    latestSuccessfulRunId?: string
  }
}

interface CanvasBlueprintInstance {
  id: string
  blueprintId: string
  title: string
  frameIds: string[]
  nodeIds: string[]
  edgeIds: string[]
  insertedBy: NodeSource
  insertedAt: string
}
```

## 画布节点类别

| category | 用途 | 典型 kind |
| --- | --- | --- |
| `source` | 输入资料和剧本来源 | `script`、`note` |
| `concept` | 设定、场景氛围和风格原则 | `scene`、`style` |
| `continuity` | 需要跨镜头锁定的对象 | `character`、`prop` |
| `production` | 可运行或可编译的生产单元 | `shot`、`prompt`、`fusion` |
| `output` | 运行输出或外部回收结果 | `video_result` |
| `review` | 评审结论和修订说明 | `note`、review badge |
| `organizer` | 画布组织结构 | `ReferenceGroup`、`ProductionFrame` |
| `handoff` | 离线交接兜底 | `package` |

## 支持的节点类型

| kind | 领域对象 | 生产角色 |
| --- | --- | --- |
| `script` | ScriptDocument | 剧本、场次、对白和旁白来源 |
| `scene` | SceneProfile | 地点、时间、氛围、光线和场景连续性 |
| `character` | CharacterProfile | 角色外观、服装、表情、禁改项 |
| `prop` | PropProfile | 道具、产品或关键物件规则 |
| `style` | StyleProfile | 美术、色彩、镜头语言和全局风格 |
| `shot` | ShotCard | 镜头级生产单位 |
| `prompt` | PromptObject | 提示词、负面约束和优化结果 |
| `fusion` | FusionTask | 多素材融合或图像处理任务 |
| `package` | GenerationPackage | 交接包和 manifest |
| `video_result` | VideoResult | 外部结果和 Take |
| `note` | NoteObject | 用户备注、灵感和待办 |

## 支持的关系类型

| relation | 方向 | 语义 |
| --- | --- | --- |
| `uses` | 使用方 -> 被使用方 | Shot 使用 Character、Scene、Prop、Style 或 Asset |
| `belongs_to` | 子对象 -> 父对象 | Shot 归属 Scene，Prompt 归属 Shot |
| `generated_by` | 产物 -> 图谱支持的源节点 | Prompt generated_by Fusion；运行来源写入领域对象字段 |
| `refines` | 新版本 -> 同类旧版本 | Prompt、Shot、Package 或 VideoResult 的版本谱系 |
| `packaged_as` | Shot -> Package | Shot 被打入交接包 |
| `result_of` | Result -> Shot / Package | 结果来自某个 Shot 或 Package |
| `depends_on` | 任务或包 -> 上下文 | Fusion 或 Package 依赖上下文 |

## ref 边界

1. `ProjectCanvas` 是画布保存单元；旧 `GraphDocument` 概念只表示其关系子集。
2. `GraphNode.id` 是画布 ID，生命周期跟随图谱。
3. `GraphNode.refId` 是领域对象 ID，生命周期由领域目录管理。
4. `GraphNode.source` 必须是 `human`、`agent`、`system` 或 `imported`，并能追溯到 audit event。
5. `GraphNode.data` 只允许缓存展示摘要、badge、折叠状态，不保存完整领域对象。
6. 删除 Node 不默认删除 `refId` 指向的领域对象。
7. 删除领域对象必须先生成影响报告，再处理引用它的 Node 和 Edge。
8. `ReferenceGroup.inputNodeIds` 只引用画布节点，不复制输入节点的领域数据。
9. `ProductionFrame.outputNodeIds` 指向输出节点；输出版本和失败记录进入 FrameRun/PromptRun/Asset 历史，不写入节点 `data`。
10. 历史轨只缓存当前、收藏、最近成功 run 的摘要，完整失败、取消、草稿历史从运行记录读取。

## 合法关系矩阵

| source \ target | script | scene | character | prop | style | shot | prompt | fusion | package | video_result | note |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| script | - | `belongs_to` | - | - | - | `uses` | - | - | - | - | `uses` |
| scene | `uses` | - | - | `uses` | `uses` | - | - | - | - | - | `uses` |
| character | - | - | - | `uses` | `uses` | - | - | - | - | - | `uses` |
| prop | - | - | - | - | `uses` | - | - | - | - | - | `uses` |
| style | - | - | - | - | - | - | - | - | - | - | `uses` |
| shot | `uses` | `belongs_to`,`uses` | `uses` | `uses` | `uses` | `refines` | `uses` | `depends_on` | `packaged_as` | - | `uses` |
| prompt | `uses` | `uses` | `uses` | `uses` | `uses` | `belongs_to` | `refines` | `generated_by` | - | - | `uses` |
| fusion | - | `depends_on` | `depends_on` | `depends_on` | `depends_on` | `depends_on` | - | - | - | - | `uses` |
| package | - | `depends_on` | `depends_on` | `depends_on` | `depends_on` | `depends_on` | `depends_on` | - | `refines` | - | `uses` |
| video_result | - | - | - | - | - | `result_of` | - | - | `result_of` | `refines` | `uses` |
| note | `uses` | `uses` | `uses` | `uses` | `uses` | `uses` | `uses` | `uses` | `uses` | `uses` | - |

## 拒绝规则

- source 或 target 不存在时拒绝创建边。
- 任何 UI、CLI 或 Agent 入口不得直接写 ProjectCanvas 文件，必须通过 Studio Command。
- relation 不在合法矩阵中时拒绝创建边。
- `generated_by` 只能连接图谱支持的源节点；PromptRun 来源写入领域对象字段或 run records，不通过 GraphEdge 表达。
- `refines` 只允许同类版本谱系：prompt -> prompt、shot -> shot、package -> package、video_result -> video_result。
- `refines`、`depends_on` 不能形成循环。
- 同一 source、target、relation 的重复边默认拒绝，除非关系 schema 明确允许多标签边。
- Package 与 Result 的关系必须能追溯到 Shot 或 manifest。
- Note 不能提升为高优先级规则，只能作为普通上下文或备注。

## 状态推进

| 操作阶段 | GraphDocument 变化 | 生产影响 |
| --- | --- | --- |
| command accepted | Studio Command 校验来源、版本、权限和参数 | 写入 command audit，并进入画布变更 |
| node draft | 新建 GraphNode，可暂未补齐生产引用 | 只允许编辑，不进入正式上下文 |
| node bound | GraphNode 绑定有效 `refId` | 可参与关系校验和上下文解析 |
| edge validated | GraphEdge 通过矩阵和循环检查 | 可参与依赖、打包或追溯 |
| graph saved | `version` 递增并写入磁盘 | 后续保存以该版本做冲突检测 |
| ref broken | 引用对象或边端点缺失 | 转 placeholder，不参与正式导出 |

## 错误语义

| 错误 | 行为 |
| --- | --- |
| unsupported node kind | 拒绝创建节点，返回支持的 kind 列表 |
| command_source_invalid | 拒绝执行，要求 UI/CLI/Agent 使用受控来源 |
| missing refId | 允许创建无引用 Note；生产节点必须补齐领域对象或进入草稿 |
| invalid relation | 拒绝连线，返回 source、target、relation 和矩阵原因 |
| cycle detected | 拒绝 `refines` 或 `depends_on` 循环 |
| stale graph version | 拒绝覆盖保存，要求重新加载或合并布局变化 |
| frame_missing_output | ProductionFrame 没有输出节点，任务不可运行 |
| reference_group_empty | ReferenceGroup 没有输入节点，只允许编辑，不参与正式运行 |
| provider_capability_missing | 所需 provider 能力缺失，保留 Frame 但禁用或降级任务 |

## 恢复与重试

- 创建节点失败时不写入 GraphDocument；领域对象已创建但节点失败时写入待绑定提示。
- Agent 命令失败时保留 Agent-visible error event，UI 展示该命令未改变画布。
- 连线失败不影响 source 和 target 节点，用户可更换 relation 后重试。
- graph version 冲突时，只允许合并纯布局变化；关系和 ref 变化需要用户确认。
- 断链节点由健康检查转为 placeholder，修复后可重新参与上下文解析。
- 缺图片或视频理解能力时，ReferenceGroup 和 ProductionFrame 保留；自动分析入口降级为人工说明或手动分镜表。
- 缺视频生成能力时，输出节点仍可保留提示词或交接包版本，生成视频按钮禁用并提示配置生成 provider。

## 验收标准

- 所有支持的节点类型都能创建、保存、重新打开并保留布局。
- 合法关系能创建，非法关系、缺失节点和循环依赖会被拒绝并给出明确原因。
- 删除 Node 不会误删领域对象；删除领域对象会列出受影响 Node 和 Edge。
- 从 Result 节点可以沿关系追溯到 Package、Shot、Prompt 和关键上下文。
- ProductionFrame 能保留 ReferenceGroup、输出节点和历史摘要，且历史详情可追溯到运行记录。
