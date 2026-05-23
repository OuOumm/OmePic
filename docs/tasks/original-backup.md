# 任务：原图备份

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F2** — "原图备份与恢复"
- **§4.3 P1：原图备份** — "原图存储策略、权限控制、重新转换入口、保留期清理"

## 预估人日

5 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/image.go` — 新增原图备份相关字段
- `backend/internal/service/image_service.go` — 上传时保存原图
- `backend/internal/service/backup_service.go`（新增）— 原图备份管理
- `backend/internal/http/handler/admin_image.go` — 原图下载/重新转换 API
- `backend/internal/database/migration.go` — 迁移脚本
- `backend/internal/repository/setting_repo.go` — 备份配置读写
- `frontend/src/routes/admin/dashboard/images/+page.svelte` — 图片管理增加原图操作

### 关键实现点

1. 配置项：`original_backup_enabled`、`original_retention_days`、`original_max_size`、`original_access_level`（admin_only/public）。
2. 上传时将原图存储到独立路径（如 `backups/original/{uid}.{ext}`）。
3. 原图默认仅管理员可下载，避免泄露 EXIF/GPS。
4. 与异步转换配合：原图作为任务源，转换失败可重试。
5. 后台支持：下载原图、基于原图重新生成派生格式。
6. 保留期清理任务：超过 `original_retention_days` 的原图自动清理。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传图片并启用原图备份 | 原图保存到备份路径 |
| 正向 | 管理员下载原图 | 返回原始文件 |
| 正向 | 基于原图重新转换 | 生成新的 AVIF |
| 异常 | 非管理员下载原图 | 返回 403 |
| 异常 | 原图超过 max_size | 不保存原图，转换正常进行 |
| 边界 | 原图保留期过期 | 清理任务删除过期原图 |
| 边界 | 原图备份关闭 | 上传不保存原图 |

### 必须覆盖的测试类型

- 单元测试：备份保存/读取/清理逻辑
- 集成测试：完整上传→备份→下载链路
- 权限测试：原图访问权限控制

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 原图路径与派生图路径隔离
- 清理任务的幂等性
- 原图不暴露在公开 API 中

### 安全检查

- 原图下载必须验证管理员权限
- EXIF 数据不在公开接口中暴露

### 文档更新要求

- 运行时配置文档说明原图备份选项
- API 文档说明原图下载接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
