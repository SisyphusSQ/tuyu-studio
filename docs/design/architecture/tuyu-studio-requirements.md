# Tuyu Studio 产品需求文档

**项目名**：图屿 Studio / Tuyu Studio  
**文档类型**：产品需求文档（PRD）  
**版本**：v0.2  
**日期**：2026-05-13  
**定位**：本地 AI 影视创作编排器  

---

## 1. 产品定位

Tuyu Studio 不是单纯的 AI 图片画布，也不是直接替代视频生成平台。它的目标是成为 **AI 视频创作者的本地前期制作台**。

核心定义：

> Tuyu Studio 是一个本地 AI 影视创作编排工具，用无限画布组织剧本、人物、场景、道具、参考图、提示词、分镜和生成包，并通过可配置的指令栈与技能系统，引导 Codex app-server 进行提示词优化、生图、融图文和纳逗Pro投喂包导出。

第一阶段不直接做视频生成。第一阶段只完成：

```text
剧本 / 角色 / 场景 / 参考图
  ↓
分镜与提示词优化
  ↓
融图文 / 参考图生成
  ↓
纳逗Pro生成包导出
  ↓
视频结果回收与管理
```

---

## 2. 背景与判断

AI 视频生成的难点不只是“生成视频”，而是前置生产流程：

- 剧本如何拆成场次和镜头。
- 人物如何保持一致。
- 场景、道具、服装、光线如何保持连续性。
- 普通描述如何转成影视级提示词。
- 多张人物图、场景图、道具图如何与剧本文字融合。
- 生成视频前，如何把素材整理成可投喂的视频生成包。
- 视频生成后，如何回收结果并绑定到对应分镜。

公开资料显示，纳逗Pro已经被定位为专业影视 AI 智能体平台，覆盖编剧、美术、分镜、视效等影视流程，并接入奇智、即梦、可灵、Vidu、海螺、Wan 等模型，支持文本、图像、视频与音频生成。但当前没有看到公开、稳定、可直接集成的纳逗Pro API 文档，因此 Tuyu Studio v1 应该优先做 **手动投喂包导出**，而不是 API 自动提交。

---

## 3. 产品目标

### 3.1 核心目标

1. **本地化管理 AI 影视前期资产**  
   剧本、人物、场景、道具、参考图、提示词、分镜、生成包、结果文件都保存在本地项目目录。

2. **把无限画布升级为 Creative Graph**  
   画布不是普通图片堆叠，而是一个节点化创作图谱：剧本节点、人物节点、场景节点、Prompt 节点、Shot 节点、Fusion 节点、Package 节点、Result 节点之间可以建立关系。

3. **建立 Instruction Stack 指令栈**  
   支持全局原则、项目原则、风格圣经、人物圣经、Provider Profile、Skill、任务模板和画布上下文的分层拼接。

4. **支持可加载 Skill**  
   提供项目级和全局级 skill，用于更稳定地引导 Codex app-server 完成提示词优化、生图、融图文、生成包导出、连续性检查等任务。

5. **输出可投喂纳逗Pro的标准生成包**  
   每个 Shot 或 Scene 可以导出一个文件夹或压缩包，包含 prompt、manifest、剧本片段、参考图、连续性说明和上传清单。

6. **预留视频接口，但 v1 不实现自动视频生成**  
   架构上保留 VideoProvider Interface，当前 provider 为 `nadou_pro/manual_export`。

### 3.2 非目标

v1 不做以下内容：

- 不做云端 SaaS。
- 不做多人协作。
- 不做用户登录系统。
- 不做纳逗Pro自动 API 提交。
- 不直接融合视频或生成最终视频。
- 不做复杂时间轴剪辑器。
- 不替代专业剪辑软件。

---

## 4. 目标用户

### 4.1 个人 AI 视频创作者

需求：

- 管理人物图、场景图、剧本和提示词。
- 把普通想法改成影视级提示词。
- 批量整理可以投喂纳逗Pro的视频生成素材。

### 4.2 AI 短剧 / 漫剧创作者

需求：

- 把短剧剧本拆成分镜。
- 保持主角、服装、场景、道具一致。
- 让每个镜头都有可复制、可追溯的生成 prompt。

### 4.3 广告片 / 产品视频创作者

需求：

- 管理产品图、品牌规则、广告文案和视觉风格。
- 生成产品广告视频的镜头包。
- 用同一套品牌原则批量优化 prompt。

### 4.4 开发者型创作者

需求：

- 本地运行。
- 可自定义 skill、prompt 模板和 provider profile。
- 可接入 Codex app-server 和未来的视频生成 API。

---

## 5. 核心概念

### 5.1 Creative Graph

Creative Graph 是 Tuyu Studio 的主工作区。它不是传统图层画布，而是由节点和边构成的创作图谱。

节点类型：

| 节点 | 说明 |
|---|---|
| ScriptNode | 剧本、场次、对白、旁白 |
| CharacterNode | 人物设定、参考图、三视图、表情、服装 |
| SceneNode | 场景设定、地点、时代、气氛、光线 |
| PropNode | 道具设定、关键物品、产品、武器等 |
| StyleNode | 项目风格、色彩、镜头语言、美术原则 |
| PromptNode | 原始 prompt、优化 prompt、负面 prompt |
| ShotNode | 单个镜头，包含时长、景别、动作、运镜 |
| FusionNode | 图图融合、图文融合、多图文融合任务 |
| PackageNode | 纳逗Pro生成包或其他 provider 生成包 |
| VideoResultNode | 从纳逗Pro等平台回收的视频结果 |

边类型：

| 边 | 说明 |
|---|---|
| uses | Shot 使用某个角色、场景、道具 |
| generated_by | 生成图 / prompt / package 由某个节点生成 |
| refines | Prompt B 是 Prompt A 的优化版本 |
| packaged_as | Shot 被导出为某个生成包 |
| result_of | 视频结果属于某个 Shot 或 Package |

### 5.2 Instruction Stack

Instruction Stack 是一次 Codex 调用前的提示词编译层。

拼接顺序建议：

```text
1. App Operating Contract
2. Global Studio Profile
3. Project Principles
4. Style Bible / Character Bible / Scene Bible
5. Provider Profile
6. Selected Skill Instructions
7. Task Template
8. Dynamic Canvas Context
9. User Request
10. Output Contract
```

其中：

- **Global Studio Profile**：全局创作偏好和默认规则。
- **Project Principles**：项目级原则，例如世界观、时代背景、禁用元素。
- **Provider Profile**：目标平台的 prompt 风格，例如纳逗Pro提示词结构。
- **Skill**：可加载的工作流能力，例如提示词优化、融图文、生成包导出。
- **Dynamic Canvas Context**：当前选中的画布节点、关联图片、剧本片段和上下文。

### 5.3 Skill

Skill 是一套可加载、可复用的 Markdown 指令与辅助资源。

建议内置以下 skill：

| Skill | 用途 |
|---|---|
| `tuyu-script-breakdown` | 剧本拆分为场次、人物、场景、道具和镜头建议 |
| `tuyu-character-bible` | 从人物图和描述生成角色圣经 |
| `tuyu-shot-prompt-optimizer` | 普通描述转影视级提示词 |
| `tuyu-image-text-fusion` | 人物图、场景图、道具图与剧本文字融合成镜头 prompt |
| `tuyu-imagegen-reference` | 调用 Codex `$imagegen` 生成或改图 |
| `tuyu-nadou-package-export` | 导出纳逗Pro生成包 |
| `tuyu-continuity-check` | 检查人物、场景、道具、镜头连续性 |
| `tuyu-prompt-score` | 对 prompt 可生成性和平台适配度打分 |

### 5.4 Provider Profile

Provider Profile 用于定义不同视频生成平台的投喂格式。

v1 默认 provider：

```text
provider = nadou_pro
mode = manual_export
```

未来可扩展：

```text
kling
vidu
wan
hailuo
custom
```

### 5.5 Shot Package / 生成包

“打包”建议在产品里命名为：

```text
生成包
Shot Package
Video Generation Brief
纳逗Pro投喂包
```

生成包是文件夹或 zip，不是一段 prompt。它至少包含：

```text
manifest.json
prompt_nadou.txt
script_excerpt.md
continuity.md
reference images
storyboard image（可选）
upload_checklist.md
```

---

## 6. 核心工作流

### 6.1 剧本到分镜

```text
用户粘贴剧本
  ↓
AnalyzeScript
  ↓
提取人物、场景、道具、情绪、动作、对白
  ↓
生成 ScriptScene 列表
  ↓
生成 ShotCard 候选
  ↓
在 Creative Graph 中生成 ScriptNode 和 ShotNode
```

输出：

- 剧本场次列表。
- 分镜卡片。
- 角色候选。
- 场景候选。
- 道具候选。

### 6.2 人物资产管理

```text
上传人物图
  ↓
创建 CharacterNode
  ↓
填写或生成角色圣经
  ↓
绑定三视图、表情图、服装图、动作参考图
  ↓
后续 Shot 引用该 CharacterNode
```

人物卡字段：

- 名字。
- 年龄、性别、身份。
- 外貌描述。
- 服装与道具。
- 性格与情绪基调。
- 视觉连续性规则。
- 禁止变化项。
- 参考图列表。

### 6.3 提示词优化

```text
原始描述 / 剧本片段 / ShotCard
  ↓
Instruction Stack
  ↓
Skill Router 选择 tuyu-shot-prompt-optimizer
  ↓
Codex app-server
  ↓
输出结构化优化结果
```

输出字段：

```ts
interface OptimizedPrompt {
  rawPrompt: string
  cinematicPrompt: string
  nadouReadyPrompt: string
  negativePrompt: string
  shotType: string
  cameraMovement: string
  lighting: string
  mood: string
  action: string
  characterContinuity: string
  sceneContinuity: string
  warnings: string[]
}
```

### 6.4 融图 / 融图文

融图文的输入可以是：

```text
人物图
+ 场景图
+ 道具图
+ 剧本片段
+ ShotCard
+ 项目原则
+ Provider Profile
```

输出可以是：

- 融合后的镜头 prompt。
- 可用于纳逗Pro的参考说明。
- 可选的 Codex `$imagegen` 参考图。
- FusionNode 与生成结果节点。

### 6.5 生成包导出

```text
选择 ShotNode 或 Scene
  ↓
收集角色、场景、道具、prompt、剧本片段
  ↓
运行 tuyu-nadou-package-export
  ↓
生成文件夹 / zip
  ↓
用户手动复制 prompt、上传参考图到纳逗Pro
```

输出目录示例：

```text
packages/scene_001/shot_001_nadou_package/
├─ manifest.json
├─ prompt_nadou.txt
├─ script_excerpt.md
├─ continuity.md
├─ upload_checklist.md
├─ character_refs/
├─ scene_refs/
├─ prop_refs/
└─ storyboard/
```

### 6.6 视频结果回收

```text
用户从纳逗Pro导出 video.mp4
  ↓
拖入 Tuyu Studio
  ↓
绑定到对应 PackageNode 或 ShotNode
  ↓
生成 VideoResultNode
  ↓
更新状态：generated / approved / needs_revision
```

---

## 7. 功能需求

优先级定义：

| 优先级 | 含义 |
|---|---|
| P0 | MVP 必须实现 |
| P1 | v1 增强功能 |
| P2 | 后续扩展 |

### 7.1 项目管理与本地存储

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-PROJ-001 | 创建项目 | P0 | 创建本地项目目录和 `project.json` |
| R-PROJ-002 | 打开项目 | P0 | 从本地目录加载项目 |
| R-PROJ-003 | 自动保存 | P0 | 自动保存 graph、节点和设置 |
| R-PROJ-004 | 最近项目 | P1 | 显示最近打开项目 |
| R-PROJ-005 | 项目导出 | P1 | 导出项目 zip，用于迁移 |

### 7.2 Creative Graph 画布

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-GRAPH-001 | 无限画布 | P0 | 支持平移、缩放、节点选择 |
| R-GRAPH-002 | 节点创建 | P0 | 支持 Script、Character、Scene、Prompt、Shot、Package 节点 |
| R-GRAPH-003 | 节点连线 | P0 | 支持 uses、generated_by、packaged_as 等关系 |
| R-GRAPH-004 | 节点属性面板 | P0 | 选中节点后在右侧编辑属性 |
| R-GRAPH-005 | 画布状态保存 | P0 | 保存节点坐标、尺寸、关系、视口状态 |
| R-GRAPH-006 | 搜索与定位 | P1 | 按节点名、人物名、Shot 编号搜索 |
| R-GRAPH-007 | 分组 / Frame | P2 | 支持按场次、章节、资产组分组 |

### 7.3 资产管理

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-ASSET-001 | 本地图片导入 | P0 | 支持拖拽图片到画布 |
| R-ASSET-002 | 本地视频导入 | P0 | 支持拖拽视频结果作为 VideoResultNode |
| R-ASSET-003 | 资产库面板 | P0 | 显示 assets、outputs、results |
| R-ASSET-004 | 资产绑定节点 | P0 | 图片可绑定到 Character、Scene、Prop、Shot |
| R-ASSET-005 | 资产标签 | P1 | 角色、场景、道具、风格、结果等标签 |
| R-ASSET-006 | 资产版本历史 | P1 | 记录生成版本和来源 |

### 7.4 剧本与分镜

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-SCRIPT-001 | 剧本节点 | P0 | 支持粘贴、编辑剧本文本 |
| R-SCRIPT-002 | 剧本解析占位 | P0 | `AnalyzeScript()` 可先用 mock，后续接 Codex |
| R-SCRIPT-003 | 场次列表 | P1 | 自动或手动拆场次 |
| R-SCRIPT-004 | ShotCard | P0 | 每个镜头包含时长、景别、动作、情绪、台词、运镜 |
| R-SCRIPT-005 | 一键生成分镜候选 | P1 | 基于剧本输出多个 ShotCard |

### 7.5 Instruction Stack

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-INST-001 | 全局 Profile | P0 | 全局原则、默认语言、默认风格 |
| R-INST-002 | 项目原则 | P0 | 世界观、风格、连续性、禁用元素 |
| R-INST-003 | Provider Profile | P0 | 纳逗Pro prompt 格式规则 |
| R-INST-004 | Prompt Compiler | P0 | 把项目规则、skill、上下文、用户请求编译成 Codex 请求 |
| R-INST-005 | Compiled Prompt Preview | P0 | 调用前可预览最终提示词 |
| R-INST-006 | 冲突检测 | P1 | 用户请求与锁定项目原则冲突时提醒 |
| R-INST-007 | 运行追溯 | P1 | 保存 `compiled_prompt.md`、上下文和输出 |

### 7.6 Skill 系统

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-SKILL-001 | 全局 skill 目录 | P0 | `~/.TuyuStudio/skills` |
| R-SKILL-002 | 项目 skill 目录 | P0 | `{project}/skills` 或 `{project}/.agents/skills` |
| R-SKILL-003 | Skill Registry | P0 | 扫描、加载、启用、禁用 skill |
| R-SKILL-004 | Skill Router | P0 | 根据任务类型自动选择 skill |
| R-SKILL-005 | Skill 编辑器 | P1 | 在 App 内查看和编辑 `SKILL.md` |
| R-SKILL-006 | Skill 测试 | P1 | 用测试输入预览输出 |
| R-SKILL-007 | Skill 导入导出 | P2 | 导出为 zip 或导入其他 skill 包 |

### 7.7 Codex app-server 集成

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-CODEX-001 | 启动 app-server | P0 | Go 后端以子进程启动 `codex app-server` |
| R-CODEX-002 | JSON-RPC over stdio | P0 | 使用默认 stdio JSONL transport |
| R-CODEX-003 | 初始化线程 | P0 | `initialize`、`initialized`、`thread/start` |
| R-CODEX-004 | turn/start | P0 | 发起提示词优化、生图、融图文任务 |
| R-CODEX-005 | localImage 输入 | P0 | 传入本地人物图、场景图、参考图 |
| R-CODEX-006 | skill input item | P0 | 显式传入 skill 路径 |
| R-CODEX-007 | outputSchema | P1 | 对提示词优化、剧本解析等任务要求结构化输出 |
| R-CODEX-008 | 事件流监听 | P0 | 监听 turn 状态、agent message、错误 |
| R-CODEX-009 | 进程重启 | P1 | app-server 崩溃时可重启 |

### 7.8 生图与参考图

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-IMG-001 | `$imagegen` 文生图 | P0 | 通过 Codex prompt 显式调用 `$imagegen` |
| R-IMG-002 | 参考图改图 | P0 | 通过 localImage + prompt 生成新图 |
| R-IMG-003 | 输出路径固定 | P0 | 要求 Codex 将图片保存到指定 outputs 路径 |
| R-IMG-004 | 生成图回填画布 | P0 | 输出图生成 ImageNode 或 Fusion 结果 |
| R-IMG-005 | prompt 保存 | P1 | 保存每次生图 prompt 与参考图 |
| R-IMG-006 | 批量生图 | P2 | 后续可改走 OpenAI API，而不是 Codex 额度 |

### 7.9 生成包导出

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-PKG-001 | Shot Package 导出 | P0 | 从一个 ShotNode 导出包 |
| R-PKG-002 | Scene Package 导出 | P1 | 从一场戏导出多个 Shot 包 |
| R-PKG-003 | manifest.json | P0 | 结构化记录 shot、角色、参考图、prompt |
| R-PKG-004 | prompt_nadou.txt | P0 | 可直接复制到纳逗Pro |
| R-PKG-005 | continuity.md | P0 | 人物、场景、道具连续性说明 |
| R-PKG-006 | upload_checklist.md | P0 | 手动上传清单 |
| R-PKG-007 | `.tuyupkg` | P2 | zip 形式的生成包文件 |

### 7.10 视频接口预留

| 编号 | 功能 | 优先级 | 说明 |
|---|---|---:|---|
| R-VIDEO-001 | VideoProvider Interface | P0 | 只定义接口，不实现视频生成 |
| R-VIDEO-002 | NadouPro manual_export | P0 | 当前只导出手动投喂包 |
| R-VIDEO-003 | API provider placeholder | P1 | 预留 `submitPackage()`、`pollResult()` |
| R-VIDEO-004 | 视频结果回收 | P0 | 用户拖回本地视频并绑定 Shot |
| R-VIDEO-005 | 状态管理 | P0 | draft、exported、submitted、generated、approved、needs_revision |

---

## 8. 非功能需求

### 8.1 本地优先

- 所有项目文件默认保存在本地。
- 不需要登录。
- 不需要云存储。
- 不主动上传项目文件到第三方平台，除非用户手动投喂或调用 Codex。

### 8.2 可追溯

每一次提示词优化、生图、融图文、导出生成包都应该保存运行记录：

```text
prompts/runs/run_yyyyMMdd_HHmmss/
├─ compiled_prompt.md
├─ selected_context.json
├─ skill_refs.json
├─ output.json
└─ instruction_stack_hash.txt
```

### 8.3 可扩展

- 支持新增 provider profile。
- 支持新增 skill。
- 支持将 manual_export 替换为 api_submit。
- 支持把 Codex app-server 替换或补充为 OpenAI API。

### 8.4 安全边界

- Codex app-server 使用 stdio，不默认开放 WebSocket。
- Codex 的 `cwd` 限制在当前项目目录。
- `writableRoots` 只给当前项目目录。
- 读取图片、导出包、打开文件夹都走 Go 后端白名单。
- 剧本、用户备注、图片说明都作为 data 包装，不作为系统指令执行。

---

## 9. MVP 范围

MVP 应该完成以下闭环：

```text
创建项目
  ↓
拖入人物图 / 场景图
  ↓
创建剧本节点
  ↓
创建 ShotNode
  ↓
通过 Instruction Stack 优化 prompt
  ↓
可选：用 Codex $imagegen 生成参考图
  ↓
导出纳逗Pro生成包
  ↓
用户手动在纳逗Pro生成视频
  ↓
拖回视频结果并绑定 Shot
```

### MVP 必须交付

1. Wails + Vue 3 + TypeScript + Ant Design Vue + AntV X6 桌面应用。
2. 本地项目目录。
3. Creative Graph 画布。
4. ScriptNode、CharacterNode、SceneNode、PromptNode、ShotNode、PackageNode、VideoResultNode。
5. 全局 Profile、项目原则、Provider Profile。
6. Prompt Compiler。
7. 至少 3 个内置 skill：
   - `tuyu-shot-prompt-optimizer`
   - `tuyu-image-text-fusion`
   - `tuyu-nadou-package-export`
8. Codex app-server stdio JSON-RPC 集成。
9. `$imagegen` 生图或参考图生成。
10. 纳逗Pro手动投喂包导出。
11. 视频结果回收。

---

## 10. 版本路线

### Phase 0：原型验证

- Wails 项目初始化。
- Go 后端启动 Codex app-server。
- Vue 前端调用 Go 方法。
- X6 画布显示节点。
- `$imagegen` 输出图片并回填画布。

### Phase 1：MVP

- Creative Graph 基础节点。
- 项目本地存储。
- Instruction Stack 基础版。
- Prompt Optimizer。
- Image-Text Fusion。
- NadouPro Shot Package。
- 视频结果回收。

### Phase 2：创作增强

- 剧本解析。
- 角色圣经。
- 风格圣经。
- 连续性检查。
- Prompt 评分器。
- 运行记录可视化。

### Phase 3：生产管理

- Scene Package 批量导出。
- `.tuyupkg` 文件。
- 版本历史。
- Shot 状态看板。
- 生成结果对比。

### Phase 4：视频接口扩展

- 如果纳逗Pro或其他平台开放 API，则接入 VideoProvider。
- 实现 API submit、poll、download、result binding。
- 多 provider 编译：`prompt_nadou.txt`、`prompt_kling.txt`、`prompt_vidu.txt`。

---

## 11. 验收标准

### 11.1 MVP 功能验收

- 可以创建和打开本地项目。
- 可以拖入图片并创建 CharacterNode / SceneNode。
- 可以粘贴剧本并创建 ScriptNode。
- 可以创建 ShotNode 并关联人物、场景、剧本。
- 可以输入原始 prompt 并生成影视化 prompt。
- 可以预览最终 compiled prompt。
- 可以调用 Codex app-server 进行提示词优化。
- 可以调用 `$imagegen` 生成参考图，并将图片回填到画布。
- 可以导出一个纳逗Pro Shot Package。
- 可以拖入生成后的视频，并绑定到对应 ShotNode。

### 11.2 质量验收

- 所有文件均保存在项目目录内。
- 生成包结构稳定、可复制、可复查。
- 每次 Codex 调用都有可追溯记录。
- Skill 可以启用、禁用、替换。
- Prompt Compiler 输出可预览。
- 用户备注和剧本文本不会被当作系统指令执行。

---

## 12. 待确认问题

1. 是否需要第一版就支持 `.tuyupkg` zip，还是先导出文件夹。
2. 是否需要在 App 内置 prompt 评分器。
3. 是否需要内置纳逗Pro prompt 示例库。
4. 是否需要支持 OpenAI Image API 作为 Codex `$imagegen` 的替代路径。
5. 是否需要对视频结果自动抽帧，用作下一镜头参考图。
6. 是否需要做中文 / 英文双语 prompt 输出。
7. 项目 skill 是否放在 `{project}/skills`，还是兼容 Codex 官方 `.agents/skills`。
8. 全局 skill 是否同步到 Codex 官方 `$HOME/.agents/skills`，还是通过 app-server 显式传 skill item。

---

## 13. 参考资料

- OpenAI Codex App Server：说明 app-server 用于 rich client 集成，支持 JSON-RPC over stdio，WebSocket 为 experimental。https://developers.openai.com/codex/app-server
- OpenAI Codex Skills：说明 skill 是包含 `SKILL.md` 的目录，支持显式和隐式调用。https://developers.openai.com/codex/skills
- OpenAI Codex CLI / App image generation：说明 Codex 可通过 `$imagegen` 生成或编辑图片，内置生图使用 `gpt-image-2`。https://developers.openai.com/codex/cli/features
- OpenAI Image Generation Guide：说明 GPT Image 模型支持文本生图与编辑，`gpt-image-2` 为当前 GPT Image 模型之一。https://developers.openai.com/api/docs/guides/image-generation
- Wails Creating a Project：说明 Wails 支持 `vue-ts` 项目模板，frontend 目录可以是任意前端项目。https://wails.io/docs/gettingstarted/firstproject/
- Ant Design Vue：Vue 版 Ant Design UI 组件。https://www.antdv.com/docs/vue/introduce/
- AntV X6：HTML/SVG 图编辑引擎，支持自定义节点和图编辑扩展。https://x6.antv.antgroup.com/en/tutorial/about
- 新华网关于纳逗Pro上线的报道：说明纳逗Pro进入预商用阶段，覆盖影视创作智能体和多模型生成能力。https://www.news.cn/fortune/20260331/95b10ce11e504148a8a4756536bf1ba7/c.html
- 新京报关于纳逗Pro的采访报道：强调专业影视 AI 工具的重点在资产管理、角色一致性和创作者决策空间。https://m.bjnews.com.cn/detail/1777474232168810.html
