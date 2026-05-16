# 路径守卫与权限

## 目标与边界

定义项目相对路径、用户选择外部文件、符号链接、路径穿越拒绝、外部导出确认和安全审计。本文不定义操作系统权限实现细节。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 项目相对路径 | 解析后的项目内受控路径 |
| 用户选择的外部文件 | selectionId 和只读或复制权限 |
| 导出目标 | 项目内路径或外部目录确认记录 |
| 路径校验失败 | 拒绝错误和 security 审计事件 |
| 文件读写请求 | 允许、拒绝或要求用户确认 |
| Agent / CLI 命令 | 权限检查、命令结果和 audit event |

## 核心对象或规则

项目内路径规则：

- 项目清单、交接包 manifest、对象文件和包清单只保存项目相对路径。
- 示例路径使用 `{studio_root}/projects/{project_id}/assets/asset_001.png`。
- 后端解析后必须确认真实路径仍在项目根内。
- 禁止 `..`、空路径、控制字符、URL 伪装、本地卷根路径和隐藏绝对形式。

用户选择外部文件：

- 用户选择单个文件只授予该文件读取或复制权限，不扩展为父目录权限。
- 外部文件进入项目时优先复制到资产区，再以项目相对路径引用。
- 如果必须保持外部受控引用，需要记录 selectionId、摘要、用途和失效处理。

符号链接：

- 项目内符号链接必须解析真实目标后再判断是否仍在项目根内。
- 指向项目外的符号链接默认拒绝。
- 导出包不得包含可跳出包目录的符号链接。

外部导出确认：

- 默认导出到项目内 packages 目录。
- 用户选择外部目录时，必须展示目标目录摘要、将写入的文件数量和覆盖策略。
- 确认记录写入审计，后续不能复用为无限期写入授权。

Agent / CLI 权限：

| 权限项 | 默认 | 要求 |
| --- | --- | --- |
| 读取 ProjectCanvas 摘要 | 允许已启用 Agent Skill 后读取 | 不返回项目根以外绝对路径 |
| 读取选中上下文 | 需要用户选择或显式命令 scope | 只返回 selection 内对象、digest 和摘要 |
| 创建 / 移动 / 连线节点 | 需要 Studio Command 校验 | 写入 command audit 和 ProjectCanvas version |
| 发起运行 | 需要 ProviderMode、输出契约和网络/外发确认 | internal_provider 与 external_agent 同样审计 |
| 写文件 | 默认拒绝直接写文件 | 只能通过 bind_result、asset import、package export 等受控命令 |
| 访问凭据 | 默认拒绝 | Agent 只能看到 credentialRef 摘要，不读取密钥值 |

## 状态推进

| 状态 | 推进规则 |
| --- | --- |
| path_requested | 收到读写请求，尚未解析 |
| path_resolved | 已规范化并解析真实路径 |
| project_allowed | 路径确认在项目根内 |
| external_selected | 来自用户显式选择，权限绑定 selectionId |
| confirmation_required | 外部导出或覆盖需要确认 |
| path_rejected | 拒绝并写 security 审计事件 |

## 错误语义

| 错误 | 语义 |
| --- | --- |
| path_traversal_rejected | 路径尝试跳出项目根 |
| symlink_escape_rejected | 符号链接目标位于项目外 |
| external_write_unconfirmed | 外部目录写入缺少用户确认 |
| agent_scope_denied | Agent 请求超出授权 Project 或 selection |
| direct_file_write_denied | Agent 或 CLI 试图绕过 Studio Command 直接写文件 |
| selection_expired | 用户选择授权已失效 |
| permission_denied | 当前进程无读写权限 |
| unsafe_package_entry | 包内条目可能写出包目录 |

错误详情只展示路径摘要和对象 ID，不展示机器私有路径。

## 恢复与重试

- path_traversal_rejected 不能自动重试，用户需选择合法项目内目标。
- selection_expired 允许重新选择文件。
- external_write_unconfirmed 允许回到确认面板或改为项目内导出。
- permission_denied 提供选择其他目录、修改权限后重试、取消操作。
- symlink_escape_rejected 允许复制真实文件到项目资产区。

## 验收标准

- 项目文件引用均以项目相对路径保存。
- 路径穿越、项目外符号链接和包内逃逸条目被拒绝。
- 用户选择外部文件不会获得目录级通配权限。
- 外部导出必须有明确确认和审计事件。
- 所有安全拒绝都有错误码、用户提示和可追踪审计摘要。
- Agent / CLI 操作不能越过 Project root、selection scope 和 Studio Command。
