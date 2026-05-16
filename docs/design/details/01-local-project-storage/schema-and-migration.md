# Schema 与迁移

## 目标与边界

Schema 设计要保证项目可以跨版本打开、验证、备份和回滚。`schemaVersion` 描述项目数据结构版本，不等于应用版本；同一个应用版本可能支持多个 schema，也可能只对部分旧 schema 提供只读打开能力。

## 输入 / 输出

| 输入 | 输出 |
| --- | --- |
| `project.tuyu.json`、领域对象和 manifest | 已验证的 Project 运行模型 |
| 旧 schema 项目 | 迁移后项目或明确的拒绝打开错误 |
| 迁移脚本 | backup、audit 事件、迁移报告 |

## 核心对象或规则

迁移以 `project.tuyu.json`、schema support matrix、migration step、pre-migration backup 和 validation report 为核心对象。核心规则是先识别版本和支持范围，再创建可校验备份，最后按版本链迁移并验证；任何一步失败都不能半写覆盖原项目。

## 状态推进

| 阶段 | 进入条件 | 输出 |
| --- | --- | --- |
| detected | 已读取 `schemaVersion` | 支持范围判断和迁移计划 |
| backup_ready | 迁移前备份完成并校验通过 | 可回滚恢复点 |
| migrating | 迁移步骤按版本链执行 | 临时迁移结果和 audit 草稿 |
| validating | 迁移结果完成结构、引用、路径和 digest 校验 | commit 或 rollback 决策 |
| committed | 校验通过并更新 schemaVersion | 可写入的新 schema 项目 |
| rolled_back | 任一步失败 | 原项目保持可恢复，保留失败报告 |

## project.tuyu.json schema 形状

```json
{
  "schemaVersion": "1.0.0",
  "project": {
    "id": "proj_001",
    "name": "project name",
    "type": "series",
    "createdAt": "2026-05-14T10:00:00+08:00",
    "updatedAt": "2026-05-14T10:20:00+08:00"
  },
  "paths": {
    "root": ".",
    "assets": "assets",
    "packages": "packages",
    "audit": "audit",
    "backups": "backups"
  },
  "defaults": {
    "providerProfileId": "provider_manual_handoff",
    "language": "zh",
    "aspectRatio": "9:16"
  },
  "integrity": {
    "lastCleanShutdown": true,
    "lastGraphVersion": 1,
    "lastHealthCheckAt": null
  },
  "graph": {
    "version": 1,
    "nodes": [],
    "edges": [],
    "viewport": {
      "x": 0,
      "y": 0,
      "zoom": 1
    }
  },
  "principles": {
    "summary": "",
    "lockedRules": []
  },
  "styleBible": {
    "summary": "",
    "visualRules": []
  }
}
```

## 字段规则

| 字段 | 规则 |
| --- | --- |
| `schemaVersion` | 语义化数据结构版本；只由迁移流程更新 |
| `project.id` | 项目稳定 ID；目录移动时不改变 |
| `paths.*` | 必须是项目内相对路径 |
| `defaults` | 只保存偏好，不保存凭据和外部敏感凭据材料 |
| `integrity.lastCleanShutdown` | 非 clean 时打开项目必须触发恢复检查 |
| `integrity.lastGraphVersion` | 与 `graph.version` 用于一致性诊断 |
| `graph` | Creative Graph 节点、边、视口、布局和图谱版本 |
| `principles` | 项目原则、禁用元素和锁定规则摘要 |
| `styleBible` | 风格圣经、画面语言、色彩和镜头基调摘要 |

## schemaVersion 与应用版本

| 项 | 含义 | 示例判断 |
| --- | --- | --- |
| schemaVersion | 项目数据结构版本 | `1.0.0` 项目需要迁移到 `1.1.0` |
| app version | 应用发布版本 | 当前应用支持 `1.0.0` 到 `1.2.0` |
| migration range | 可自动迁移的 schema 范围 | `0.9.0 -> 1.0.0 -> 1.1.0` |
| read-only range | 可诊断但不可写入的旧 schema | 允许导出备份，不允许保存 |

## 迁移流程

```text
1. detect: 读取 project.tuyu.json，识别 schemaVersion 和支持矩阵
2. backup: 写入 backups/pre_migration_{timestamp}/
3. migrate: 按版本链执行幂等迁移
4. validate: 运行 JSON、引用、路径、digest 和 package 校验
5. commit: 更新 schemaVersion，写 audit/migrations.jsonl
6. rollback: 任一步失败时恢复备份或保留原项目
```

## 错误语义

| 错误 | 用户提示 | 系统行为 |
| --- | --- | --- |
| unsupported schema | 当前项目数据版本不受支持 | 拒绝写入；允许用户导出诊断包 |
| failed migration | 项目升级失败，已保留升级前备份 | 回滚到迁移前状态，写入失败报告 |
| invalid backup | 迁移前备份不可用 | 停止迁移，不修改原项目 |
| invalid migrated project | 迁移后校验未通过 | 回滚并保留迁移临时目录供诊断 |
| schema downgrade | 项目版本高于当前应用可写范围 | 只读打开或拒绝打开 |

## 恢复与重试

- 迁移脚本必须幂等；重复执行不重复创建领域对象、节点或 Package。
- 迁移前 backup 校验失败时，不允许继续迁移。
- rollback 成功后，`project.tuyu.json` 的 `schemaVersion` 与原始文件一致。
- 迁移失败报告写入 `audit/migrations.jsonl`；如果 audit 不可写，写入备份目录内的 migration report。

## 验收标准

- `schemaVersion` 的升级不会依赖应用版本号比较。
- 迁移前必定存在可校验 backup，backup 无效时不改动原项目。
- 任意 failed migration 都能让原项目保持可恢复，且旧 `project.tuyu.json` 不被半写覆盖。
- unsupported schema 与 failed migration 的用户提示不同，便于用户判断是升级应用还是恢复备份。
