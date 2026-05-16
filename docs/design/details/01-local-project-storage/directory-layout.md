# 目录布局

## 目标与边界

本地项目目录是生产数据的执行真相。目录布局必须支持长期保存、复制迁移、健康检查、离线交接和崩溃恢复。所有项目内引用默认使用相对路径，避免把本机路径写入可提交或可迁移的项目文件。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| 用户选择的 `{studio_root}` | 可管理的全局配置、能力包和项目列表目录 |
| 新建或导入的 `{project_id}` | 完整项目骨架、单一项目清单和生产子目录 |
| 导入资产、PromptRun、Package、Review | 项目内相对路径、metadata、audit 和健康检查对象 |

## 核心对象或规则

目录布局围绕三个稳定根对象组织：`{studio_root}` 负责全局配置和项目索引，`{project_id}` 负责单项目生产数据，Package 版本目录负责可离线交接的镜头资料。核心规则是项目内生产引用必须相对 `{project_id}`，跨项目或外部文件只能作为显式 managed reference 进入健康检查。

## 完整目录树

```text
{studio_root}/
├─ config/
│  ├─ app_settings.json
│  ├─ studio_profile.md
│  └─ provider_profiles/
│     └─ {profile_id}.json
├─ skills/
│  └─ {skill_name}/
│     ├─ skill.md
│     └─ assets/
└─ projects/
   └─ {project_id}/
      ├─ project.tuyu.json
      ├─ characters/
      │  └─ {character_id}.json
      ├─ scenes/
      │  └─ {scene_id}.json
      ├─ props/
      │  └─ {prop_id}.json
      ├─ shots/
      │  └─ {shot_id}.json
      ├─ prompts/
      │  ├─ prompt_index.json
      │  └─ runs/
      │     └─ {run_id}.json
      ├─ assets/
      │  ├─ inputs/
      │  ├─ refs/
      │  ├─ outputs/
      │  ├─ results/
      │  └─ thumbnails/
      ├─ packages/
      │  └─ {package_id}/
      │     ├─ manifest.json
      │     ├─ prompt.md
      │     ├─ continuity.md
      │     └─ assets/
      ├─ audit/
      │  ├─ events.jsonl
      │  ├─ migrations.jsonl
      │  └─ repairs.jsonl
      ├─ backups/
      │  └─ {backup_id}/
      └─ locks/
         ├─ session.lock
         └─ write.lock
```

## 责任表

| 路径 | 责任 | 写入规则 | 迁移要求 |
| --- | --- | --- | --- |
| `project.tuyu.json` | 项目元信息、schema、默认配置、Creative Graph 分区、原则、风格设定、索引和完整性标记 | 单一清单原子写；只存相对路径和非敏感偏好；图谱分区版本号递增 | schema 迁移主入口，可从备份恢复上一版 |
| `assets/` | 导入、参考、生成、结果和缩略图文件 | copy 为默认；受控引用需显式记录风险 | 健康检查必须验证存在性和 digest |
| `prompts/runs/` | PromptRun 输入摘要、状态、输出、错误 | 单 run 文件追加状态或原子替换 | 迁移保留原始输入摘要 |
| `packages/` | GenerationPackage 版本目录和 manifest | 新版本新目录，不覆盖旧包 | manifest 使用相对路径 |
| `audit/` | 用户操作、运行摘要、迁移、修复记录 | append-only JSONL | 不参与回滚覆盖 |
| `backups/` | 自动备份、迁移前快照、恢复点 | 轮转保留；不可写入新生产对象 | 回滚来源必须校验完整性 |
| `locks/` | 会话锁、写入锁和接管信息 | 短生命周期；打开和保存时维护 | 复制项目时可忽略陈旧锁 |

用户体验口径：`project.tuyu.json` 是项目根唯一用户可感知的核心项目清单。领域对象目录、运行记录和审计日志是内部组织方式，工作台不应把它们包装成需要用户逐个理解和维护的“项目元数据文件”。

## 路径规则

1. 项目清单、交接包 manifest 和领域对象只保存相对 `{project_id}` 根的路径。
2. `{studio_root}` 可移动；项目不得依赖固定父目录。
3. 外部 managed reference 必须记录为显式引用对象，并在健康检查中标为迁移风险。
4. 交接包中不得出现绝对路径、用户主目录、临时目录或完整下载地址。
5. 文件名使用稳定 ID，不依赖用户展示名；展示名只存 metadata。
6. 同一 digest 的导入资产允许复用，复用行为必须写入 audit。

## 状态推进

| 阶段 | 目录变化 | 可用能力 |
| --- | --- | --- |
| studio initialized | 创建 `{studio_root}/config`、`skills`、`projects` | 可配置全局偏好和创建项目 |
| project created | 创建 `{project_id}` 骨架和 `project.tuyu.json` | 可打开、编辑和健康检查 |
| assets imported | 写入 `assets/` 和对象 metadata | 可建立连续性引用 |
| package exported | 写入 `packages/{package_id}` 和 manifest | 可离线交接和回收结果 |
| project copied | 项目目录迁移到新位置 | 相对路径项目仍可检查和打开 |

## 错误语义

| 错误 | 级别 | 行为 |
| --- | --- | --- |
| 缺少 `project.tuyu.json` | blocking | 拒绝按正式项目打开 |
| 缺少图谱分区 | blocking | 尝试从 backups 恢复，否则只允许诊断 |
| 目录不可写 | blocking | 切换只读或要求另存 |
| 绝对路径泄漏 | blocking | 拒绝导出 Package |
| managed reference 缺失 | warning | 标记受影响对象，允许用户重新定位 |

## 恢复与重试

- 项目复制后发现 `locks/` 中存在旧锁时，按 stale lock 处理，不阻断健康检查。
- Package 导出失败时保留旧版本目录，新版本临时目录标记为 recoverable draft。
- 资产复制中断时，未完成文件不进入 asset index；用户可重新导入。

## 验收标准

- 新建项目生成完整目录骨架，`project.tuyu.json` 可解析且核心分区完整。
- 复制 `{project_id}` 到另一个 `{studio_root}/projects/` 下后，健康检查仍可运行。
- Project、Graph、Package、Asset 的引用均可在不依赖原父目录的情况下解析。
- 健康检查能识别项目内绝对路径泄漏并阻断正式导出。
