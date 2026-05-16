# Skill Registry and Routing

## 目标与边界

定义能力包目录结构、来源信任状态、启停规则、路由解析、项目覆盖、Canvas Blueprint、Agent Skill / CLI 入口和 schema 兼容性。本文不定义具体能力包内容。

## 输入 / 输出

输入：

- 内置能力包目录、用户全局目录、项目目录和外部导入包。
- Canvas Blueprint、task mode、用户显式选择和项目覆盖配置。
- 外部 Agent 读取的 Studio Skill 与 CLI/MCP 风格命令。

输出：

- `SkillRegistry` 索引。
- Blueprint registry、路由结果、启停状态、信任状态和兼容性诊断。
- Agent 可见的命令清单、权限摘要和错误语义。

## 核心对象或规则

目录结构：

```text
{skill_name}/
├─ SKILL.md
├─ references/
├─ examples/
├─ schemas/
└─ assets/
```

Agent Skill / CLI 入口：

| 命令 | 作用 | 写入边界 |
| --- | --- | --- |
| `project.open` | 打开本地 Project 并读取画布摘要 | 只读，返回 ProjectCanvas version |
| `canvas.list` | 读取节点、边、Frame、Blueprint 和选中上下文 | 只读，隐藏未授权路径 |
| `canvas.create_node` | 创建节点和可选领域对象 | Studio Command |
| `canvas.move_node` | 移动节点、Frame 或 ReferenceGroup | Studio Command |
| `canvas.connect_nodes` | 创建合法关系边 | Studio Command + 关系矩阵 |
| `canvas.insert_blueprint` | 插入 Blueprint 候选或正式实例 | Studio Command + 用户确认策略 |
| `run.start` | 基于节点或 Frame 发起运行 | Studio Command + ProviderMode |
| `result.bind` | 将输出绑定为 Asset/Take/Review 入口 | Studio Command + 输出契约 |

Agent 不直接写项目文件，不读取项目根以外路径；所有写操作都返回 CommandResult、audit event 和 UI 可见变更摘要。

Canvas Blueprint 声明：

```yaml
id: storyboard_to_prompt
title: 剧本到镜头提示词
inputs:
  - script
  - character
  - scene
nodes:
  - kind: shot
  - kind: prompt
edges:
  - relation: uses
outputs:
  - prompt_object
recovery:
  missing_input: waiting_user
  contract_failed: recoverable_error
```

来源信任状态：

| 来源 | 默认状态 | 行为 |
| --- | --- | --- |
| 内置 | enabled | 可自动参与路由 |
| 用户全局目录 | disabled | 用户启用后参与路由 |
| 项目目录 | project_enabled | 仅当前项目可用 |
| 外部导入包 | quarantined | 检查结构并经确认后启用 |
| Agent Skill | disabled | 用户启用后允许外部 Agent 读取命令说明 |

route resolution：

1. 用户显式选择优先。
2. 未显式选择时按 task mode 使用默认能力包。
3. 项目能力包可覆盖全局默认，但必须展示覆盖来源。
4. Blueprint 输入槽、输出契约和 required output schema 必须兼容 task mode。
5. disabled 或 quarantined 能力包不得自动运行。
6. Agent 入口只能暴露当前 Project、选中上下文和命令 schema，不暴露全局目录扫描能力。

schema compatibility：

- 能力包声明的 task modes 必须包含当前任务。
- 输出 schema ID 必须与任务输出契约匹配。
- 版本不兼容时可预览，但不可运行。

## 状态推进

```text
discovered -> validated -> enabled
discovered -> invalid
validated -> disabled
external_imported -> quarantined -> validated
enabled -> disabled
```

- `invalid`：结构缺失、frontmatter 不合法或 schema 不兼容。
- `quarantined`：只允许查看 metadata，不参与路由。

## 错误语义

| 错误 | 含义 | 行为 |
| --- | --- | --- |
| `skill_manifest_invalid` | SKILL.md frontmatter 不合法 | 标记 invalid |
| `skill_required_file_missing` | 必需文件缺失 | 标记 invalid |
| `skill_disabled` | 能力包未启用 | 不参与自动路由 |
| `skill_schema_incompatible` | 输出 schema 不匹配 | 阻断运行 |
| `skill_override_conflict` | 项目覆盖存在冲突 | 要求用户选择 |
| `agent_skill_disabled` | 外部 Agent Skill 未启用 | 拒绝 Agent 命令 |
| `blueprint_contract_invalid` | Blueprint 输入或输出契约不兼容 | 阻断插入或运行 |

## 恢复与重试

- 修复能力包文件后可重新扫描。
- 用户启用 disabled 能力包后重新路由。
- quarantined 包通过结构检查和用户确认后进入 validated。
- schema 不兼容只能通过升级能力包或切换任务契约恢复。

## 验收标准

- Registry 能列出来源、版本、task modes、启停和信任状态。
- Agent Skill 能列出可读范围、可写命令、ProviderMode 和错误语义。
- Blueprint 能声明节点、边、Frame、输入槽、输出契约和恢复语义。
- disabled 和 quarantined 能力包不会被自动选择。
- 用户显式选择优先于默认路由。
- 项目覆盖来源可见。
- schema 不兼容时任务不可运行。
- 重新扫描能反映文件变化。
