# 项目命名建议：图屿 / Tuyu Canvas

## 1. 推荐名称

### 中文名

```text
图屿
```

### 英文名

```text
Tuyu Canvas
```

### GitHub Repo

```text
tuyu-canvas
```

### App 标题栏

```text
图屿 Canvas
```

### 命令行 / 内部代号

```text
tuyu
```

---

## 2. 为什么推荐“图屿”

**图屿 = 图片 + 岛屿。**

这个名字和项目形态匹配：

```text
无限画布上分布着图片、Prompt、生成结果、参考图、节点关系。
```

每一张图片、每一个 Prompt、每一次生成结果，都像画布上的一个“小岛”。用户在这些图像节点之间组织想法、延展创意、生成新图片。

“图屿”相比“AI Canvas”“Prompt Canvas”“Image Board”更有辨识度，也不会直接绑定某个模型、平台或供应商。后续即使从 Codex 切换到其他本地模型、OpenAI Image API、ComfyUI、Stable Diffusion 或其他图像生成服务，这个名字也仍然成立。

---

## 3. 名称定位

一句话定位：

```text
图屿是一款本地 AI 无限画布，用 Prompt、参考图和节点关系组织你的图像生成流程。
```

更短的产品介绍：

```text
本地 AI 图像生成画布。
```

偏开发者的介绍：

```text
A local AI image canvas for prompt-driven visual exploration.
```

偏设计师的介绍：

```text
在无限画布上组织灵感、参考图和 AI 生成结果。
```

---

## 4. 推荐命名组合

| 场景 | 推荐名称 |
|---|---|
| 中文产品名 | 图屿 |
| 英文产品名 | Tuyu Canvas |
| App 名称 | 图屿 Canvas |
| GitHub Repo | `tuyu-canvas` |
| 项目目录 | `tuyu-canvas` |
| 命令行工具 | `tuyu` |
| 配置目录 | `~/.tuyu` 或 `~/TuyuCanvas` |
| 默认项目目录 | `~/TuyuCanvas/projects` |
| Go package 前缀 | `tuyu` |
| Wails 项目名 | `tuyu-canvas` |

---

## 5. App 内文案建议

### 顶部标题

```text
图屿 Canvas
```

### 欢迎页标题

```text
欢迎使用图屿
```

### 欢迎页副标题

```text
在本地无限画布上组织参考图、Prompt 和 AI 生成结果。
```

### 空白画布提示

```text
拖入图片，或输入 Prompt 开始生成。
```

### 生成按钮

```text
生成图片
```

### 参考图生成按钮

```text
基于选中图片生成
```

### 保存状态

```text
已保存到本地
```

### 生成中状态

```text
正在生成图片...
```

### 生成完成状态

```text
图片已生成并添加到画布
```

---

## 6. 备选名称

| 名称 | 气质 | 说明 |
|---|---|---|
| 图屿 | 简洁、中文品牌感强 | 最推荐 |
| 像屿 | 更抽象、更设计感 | “图像之岛” |
| 图织 | 偏创作 | 图片和 Prompt 被组织、编织在一起 |
| PromptDeck | 偏英文、偏工具 | Prompt 卡片工作台 |
| NodeCanvas | 偏技术 | 强调节点画布 |
| InkNode | 偏开发者工具 | 图像节点、生成节点 |
| PromptBoard | 直白 | Prompt 白板 |
| PixelDock | 偏桌面工具 | 图片素材停靠站 |
| CanvasForge | 偏生成/锻造 | 适合更硬核的生成工具 |
| Local Loom | 偏本地创作 | 本地编织图像流程 |

---

## 7. 为什么不建议直接叫 AI Canvas

`AI Canvas` 太泛，存在几个问题：

1. **难以形成品牌辨识度**  
   任何带 AI 画布能力的产品都可以叫 AI Canvas。

2. **搜索结果容易混杂**  
   用户很难通过关键词准确找到你的项目。

3. **不利于后续扩展**  
   项目不只是一个 Canvas，而是本地创作流程、Prompt 编译、图像节点、参考图关系和生成历史的组合。

4. **缺少中文产品气质**  
   如果目标用户包含中文创作者，中文名更容易形成记忆点。

---

## 8. 品牌调性

“图屿”的气质可以定成：

```text
本地、轻量、创作、节点、图像、私有化。
```

不建议走过于夸张的 AI 平台感，例如：

```text
超级 AI 创作宇宙
下一代多模态智能工作站
全自动视觉创作平台
```

更适合走冷静、工具化、创作者友好的路线：

```text
本地 AI 图像画布
Prompt-driven image workspace
Local visual exploration canvas
```

---

## 9. Logo 方向

可以考虑三个方向：

### 方向 A：岛屿 + 图片

图形元素：

```text
小岛轮廓 + 图片边框 + 节点连接线
```

适合强调“图屿”的字面意义。

### 方向 B：节点 + 画布

图形元素：

```text
多个方形图片节点 + 连接线 + 无限画布网格
```

适合强调“节点式 AI 图像工作流”。

### 方向 C：极简中文字标

图形元素：

```text
图屿两个字的定制字标
```

适合做中文品牌感。

---

## 10. 配色方向

不需要一开始做复杂品牌系统。可以先用简洁中性色：

```text
背景：浅灰 / 深灰
主色：蓝紫 / 靛蓝 / 青绿色
强调：生成状态用亮色
```

如果使用 Ant Design Vue，可以先沿用默认主题，后续再定制 token。

---

## 11. 项目目录命名建议

### 开发目录

```text
tuyu-canvas/
```

### 用户数据目录

macOS / Linux：

```text
~/TuyuCanvas/
```

Windows：

```text
%USERPROFILE%\TuyuCanvas\
```

### 默认项目结构

```text
~/TuyuCanvas/
└─ projects/
   └─ demo/
      ├─ scene.json
      ├─ inputs/
      ├─ outputs/
      ├─ refs/
      ├─ masks/
      └─ logs/
```

---

## 12. README 开头示例

```md
# 图屿 / Tuyu Canvas

图屿是一款本地 AI 无限画布，用 Prompt、参考图和节点关系组织你的图像生成流程。

它基于 Wails、Go、Vue 3、TypeScript、Ant Design Vue 和 AntV X6 构建，运行在本地桌面环境中。Go 后端负责本地文件系统、Codex app-server 通信和图片生成任务，Vue 前端负责无限画布、Prompt 输入和生成结果管理。

## Core Idea

- Infinite canvas for image nodes and prompt nodes
- Local-first project storage
- Codex app-server integration
- `$imagegen` prompt-driven image generation
- Generated images are saved locally and inserted back into the canvas
```

---

## 13. 发布前检查项

正式公开发布前，建议检查：

```text
1. 是否已有同名软件或开源项目
2. GitHub repo 名是否可用
3. 域名是否可用
4. 商标是否存在冲突
5. App Store / Microsoft Store 是否存在同名应用
6. 中英文搜索结果是否容易混淆
```

如果只是本地自用或小范围开源，优先级可以低一些。  
如果计划商业化，商标和域名需要提前查。

---

## 14. 最终建议

采用：

```text
中文名：图屿
英文名：Tuyu Canvas
Repo：tuyu-canvas
App：图屿 Canvas
```

一句话介绍：

```text
图屿是一款本地 AI 无限画布，用 Prompt、参考图和节点关系组织你的图像生成流程。
```
