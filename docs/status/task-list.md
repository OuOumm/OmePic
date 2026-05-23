# 整改与扩展子任务清单

> 源设计文档：`docs/debug/remediation-extension-design/rectification-and-extension-design.md`
>
> 生成日期：2026-05-22

## P0：必备整改项（约 24 人日）

| # | 任务名称 | 文件路径 | 预估人日 | 状态 | 负责人 |
|---|----------|----------|----------|------|--------|
| 1 | 上传资源保护 | `docs/tasks/upload-resource-guard.md` | 2 | 未开始 | TBD |
| 2 | 文件真实性校验 | `docs/tasks/file-authenticity-validation.md` | 2 | 未开始 | TBD |
| 3 | URL 上传 SSRF 防护 | `docs/tasks/url-upload-ssrf-protection.md` | 3 | 未开始 | TBD |
| 4 | Token 治理基础 | `docs/tasks/token-governance.md` | 3 | 未开始 | TBD |
| 5 | 默认密码安全改造 | `docs/tasks/default-password-hardening.md` | 1.5 | 未开始 | TBD |
| 6 | 软删除与回收站 | `docs/tasks/soft-delete-recycle-bin.md` | 4 | 未开始 | TBD |
| 7 | SQLite 核心索引 | `docs/tasks/sqlite-core-indexes.md` | 1.5 | 已完成 | Implement Agent |
| 8 | 配置审计日志 MVP | `docs/tasks/config-audit-log.md` | 3 | 未开始 | TBD |
| 9 | 存储健康检查 MVP | `docs/tasks/storage-health-check.md` | 4 | 未开始 | TBD |

## P1：高价值扩展与架构升级（约 50 人日）

| # | 任务名称 | 文件路径 | 预估人日 | 状态 | 负责人 |
|---|----------|----------|----------|------|--------|
| 10 | AVIF 异步转换队列 | `docs/tasks/async-avif-queue.md` | 8 | 未开始 | TBD |
| 11 | 原图备份 | `docs/tasks/original-backup.md` | 5 | 未开始 | TBD |
| 12 | 多格式输出 WebP | `docs/tasks/webp-multi-format.md` | 7 | 未开始 | TBD |
| 13 | 图片审核 | `docs/tasks/image-content-review.md` | 8 | 未开始 | TBD |
| 14 | 批量打包下载 | `docs/tasks/batch-zip-download.md` | 5 | 未开始 | TBD |
| 15 | 存储回退策略 | `docs/tasks/storage-fallback.md` | 6 | 未开始 | TBD |
| 16 | Prometheus 指标 | `docs/tasks/prometheus-metrics.md` | 4 | 未开始 | TBD |
| 17 | 请求追踪与结构化日志 | `docs/tasks/request-trace-structured-log.md` | 3 | 未开始 | TBD |
| 18 | 滥用统计聚合 | `docs/tasks/abuse-stats-aggregation.md` | 4 | 未开始 | TBD |

## P2：锦上添花与长期能力（约 57 人日）

| # | 任务名称 | 文件路径 | 预估人日 | 状态 | 负责人 |
|---|----------|----------|----------|------|--------|
| 19 | JPEG-XL 支持 | `docs/tasks/jpegxl-support.md` | 5 | 未开始 | TBD |
| 20 | 短链服务 | `docs/tasks/short-link-service.md` | 5 | 未开始 | TBD |
| 21 | 相册/集合管理 | `docs/tasks/album-management.md` | 7 | 未开始 | TBD |
| 22 | 图片访问统计 | `docs/tasks/image-access-stats.md` | 6 | 未开始 | TBD |
| 23 | 图片处理工具 | `docs/tasks/image-processing-tools.md` | 8 | 未开始 | TBD |
| 24 | 存储迁移与复制 | `docs/tasks/storage-migration.md` | 10 | 未开始 | TBD |
| 25 | API Key / 用户空间 | `docs/tasks/api-key-user-space.md` | 12 | 未开始 | TBD |
| 26 | CDN 预热任务 | `docs/tasks/cdn-prefetch-tasks.md` | 4 | 未开始 | TBD |

## 汇总

| 优先级 | 任务数 | 总预估人日 |
|--------|--------|------------|
| P0 | 9 | 24 |
| P1 | 9 | 50 |
| P2 | 8 | 57 |
| **合计** | **26** | **131** |

## 依赖关系

```
P0 任务群（安全与稳定性基线）
├── upload-resource-guard ← 无依赖
├── file-authenticity-validation ← 无依赖
├── url-upload-ssrf-protection ← file-authenticity-validation（共享 MIME 校验）
├── token-governance ← 无依赖
├── default-password-hardening ← 无依赖
├── soft-delete-recycle-bin ← sqlite-core-indexes（索引包含软删除列）
├── sqlite-core-indexes ← 无依赖
├── config-audit-log ← 无依赖
└── storage-health-check ← 无依赖

P1 任务群（高价值扩展）
├── async-avif-queue ← upload-resource-guard
├── original-backup ← async-avif-queue（原图作为转换源）
├── webp-multi-format ← async-avif-queue
├── image-content-review ← 无强依赖
├── batch-zip-download ← 无强依赖
├── storage-fallback ← storage-health-check
├── prometheus-metrics ← 无强依赖
├── request-trace-structured-log ← 无强依赖
└── abuse-stats-aggregation ← sqlite-core-indexes

P2 任务群（长期能力）
├── jpegxl-support ← webp-multi-format
├── short-link-service ← 无强依赖
├── album-management ← 无强依赖
├── image-access-stats ← request-trace-structured-log
├── image-processing-tools ← 无强依赖
├── storage-migration ← storage-fallback
├── api-key-user-space ← token-governance
└── cdn-prefetch-tasks ← 无强依赖
```
