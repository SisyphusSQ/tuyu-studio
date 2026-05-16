# 08. 安全、隐私与可观测性

## 子文档

| 文档 | 作用 |
| --- | --- |
| `path-guard-and-permissions.md` | 路径守卫、权限、用户选择文件、项目外写入拒绝 |
| `privacy-and-redaction.md` | 本地隐私、凭据隔离、提交版文档脱敏 |
| `audit-events.md` | 审计事件结构、分类、关键操作记录 |
| `observability-and-recovery.md` | 运行记录、错误三层展示、恢复与回滚 |

本 README 只作为模块总览；详细规则、清单和矩阵以子文档为准。

## 目标

图屿处理的是用户剧本、角色设定、参考图、生成结果和创作策略，这些内容默认属于本地私有资产。安全设计必须防止路径越界、隐式上传、提示注入、资产丢失、能力包滥用和不可追溯的黑盒运行。

## 安全原则

| 原则 | 要求 |
| --- | --- |
| 本地默认 | 项目数据默认只保存在本地项目目录 |
| 最小权限 | 文件读写只开放当前项目和用户显式选择的文件 |
| 显式联网 | 需要外部能力的任务必须先让用户确认 |
| 可审计 | 导入、运行、导出、回收、删除、迁移都有记录 |
| 可恢复 | 危险操作先进入回收、备份或新版本，不直接覆盖 |
| 数据包装 | 剧本、备注、图片说明、外部文本一律作为 data |
| 供应商隔离 | provider 差异通过 profile 和 adapter 管理 |

## 路径守卫

```go
type PathGuard interface {
    ResolveProjectPath(projectID string, relativePath string) (string, error)
    ResolveUserSelectedPath(selectionID string) (string, error)
    EnsureWritable(path string) error
    EnsureReadable(path string) error
}
```

校验规则：

- 项目内文件必须使用相对路径进入项目清单、交接包 manifest 和对象文件。
- 后端内部解析为绝对路径后，必须确认它仍在项目根内。
- 用户显式选择的外部文件只允许读取或复制，不获得目录通配权限。
- 禁止 `..`、符号链接绕过、隐藏绝对路径、URL 伪装成本地路径。
- 导出交接包时拒绝写入项目外，除非用户选择“导出到外部目录”并确认。

## 隐私边界

| 数据 | 默认行为 |
| --- | --- |
| 剧本文本 | 保存在项目目录，不自动上传 |
| 角色参考图 | 复制到项目资产目录，AI 任务只在用户触发时使用 |
| 运行记录 | 本地保存，包含上下文摘要和产物路径 |
| 交接包 | 用户主动导出，包内不含机器私有路径 |
| 结果文件 | 用户拖入后复制或受控引用 |
| 全局配置 | `~/.tuyu-studio` 只保存非敏感偏好、provider profile 和 credentialRef |
| 项目配置 | 只保存 providerProfileId 和非敏感 override |
| 凭据 | API key、token、refresh token 只进入系统密钥存储 |

如未来引入需要凭据的 provider，凭据必须进入系统级安全存储或独立密钥方案，不写入项目目录、manifest、日志或运行记录。导出交接包、支持包或脱敏报告前必须检查疑似 credential 泄漏。

## 提示注入防护

风险来源：

- 剧本内容中包含伪指令。
- 用户备注要求忽略项目规则。
- 图片说明或外部文本要求改变系统行为。
- 能力包被修改后输出不符合契约。

处理方式：

| 风险 | 防护 |
| --- | --- |
| 剧本文本伪指令 | 包装为 `creative_material`，明确禁止作为系统指令 |
| 用户备注冲突 | 编译阶段比较锁定原则，冲突则阻断 |
| 图片说明越权 | 只作为 `visual_reference` |
| 能力包污染 | 校验来源、版本、任务模式和输出 schema |
| 输出不可信 | 结构化校验失败则不写正式对象 |

## 能力包安全

| 能力包行为 | 当前策略 |
| --- | --- |
| Markdown 指令 | 允许 |
| 引用资料 | 允许，限制在能力包目录 |
| 静态 schema | 允许 |
| 本地脚本执行 | 默认禁止 |
| 网络访问 | 默认禁止 |
| 修改项目文件 | 只能通过图屿任务输出契约，不直接写 |

能力包启用前展示：

- 名称、描述、版本、来源路径。
- 支持的任务模式。
- 将参与的输出 schema。
- 是否来自项目目录或用户全局目录。

## 审计事件

`audit/events.jsonl` 记录结构：

```json
{
  "id": "evt_20260514_102000_001",
  "timestamp": "2026-05-14T10:20:00+08:00",
  "actor": "local_user",
  "action": "package.exported",
  "projectId": "proj_20260514_001",
  "targetType": "package",
  "targetId": "pkg_001",
  "summary": "Exported shot handoff package",
  "metadata": {
    "shotId": "shot_001",
    "profileId": "provider_manual_handoff"
  }
}
```

事件分类：

| 分类 | 示例 |
| --- | --- |
| project | 创建、打开、关闭、迁移、恢复 |
| asset | 导入、绑定、解绑、缺失、恢复 |
| graph | 创建节点、连线、删除、布局保存 |
| run | 创建任务、开始、进度、完成、失败、取消 |
| package | 导出、校验、重新导出、打开目录 |
| result | 回收、绑定、review、废弃 |
| security | 路径拒绝、权限失败、联网请求、冲突阻断 |

## 可观测性

| 对象 | 记录 |
| --- | --- |
| PromptRun | request、compiled instruction、context、schema、events、output/error |
| GenerationPackage | manifest、引用文件清单、导出摘要、校验结果 |
| Project Health | 缺失资产、断链、schema、外部引用、digest mismatch |
| Runtime Task | 状态、耗时、事件、取消、重试、错误码 |
| User Operation | 关键写操作和危险操作摘要 |

错误详情分三层：

1. 用户提示：一句话说明和下一步。
2. 技术详情：错误码、对象 ID、文件路径摘要。
3. 调试记录：完整错误 JSON，保存在项目 audit 或 run 目录。

## 脱敏规则

正式 repo 文档和提交版测试摘要不写入：

- 真实用户剧本全文。
- 真实外部平台账号、凭据、会话密钥。
- 机器私有路径。
- 私有素材完整下载地址。
- 未脱敏的客户、品牌或商业项目名称。
- 原始命令输出中的敏感主机、路径或凭据。
- provider API key、token、refresh token、credentialRef 对应的真实值。

项目内部运行记录可以保存用户本地资产信息，但导出问题报告时必须经过脱敏摘要。

## 恢复与回滚

| 操作 | 恢复方式 |
| --- | --- |
| 误删节点 | undo 或从 audit 恢复节点摘要 |
| 误删资产绑定 | 从 audit 还原 binding |
| 删除资产文件 | 回收区恢复 |
| 导出包错误 | 重新导出新版本，旧版本保留 |
| 运行任务失败 | 根据 PromptRun 重试 |
| schema 迁移失败 | 回滚迁移前备份 |
| 项目锁异常 | 健康检查后用户确认接管 |

## 验收标准

- 路径穿越和项目外写入被拒绝并记录安全事件。
- AI 任务默认只获得当前项目写权限。
- 剧本和用户备注在编译结果中被标记为 data。
- 任务失败后用户能看到错误码、提示和可重试性。
- 导出包不包含机器私有路径。
- 删除资产先进入回收区，可恢复。
- 项目健康检查能列出断链、缺失资产、schema 和外部引用风险。
- 提交版文档不包含真实凭据、本机路径或外部平台私有细节。
- 项目、运行记录、manifest、日志和支持包不包含 provider API key 或 token。
