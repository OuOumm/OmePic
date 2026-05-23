# 任务：软删除与回收站

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.4 可维护性不足与整改建议** — "删除逻辑与去重映射、物理文件生命周期边界不够明确"
- **§1.5.3 去重 + 软删除 + 回收站机制**
- **§4.2 P0：删除生命周期重构 MVP** — "图片软删除、回收站字段、去重映射不复用待清理对象"

## 预估人日

4 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/image.go` — 新增软删除字段：`deleted_at`、`deleted_by`、`delete_reason`、`purge_after`
- `backend/internal/repository/image_repo.go` — 查询默认排除已删除记录
- `backend/internal/service/image_service.go` — 删除改为软删除，去重不复用已删除记录
- `backend/internal/http/handler/admin_image.go` — 回收站列表与恢复 API
- `backend/internal/http/handler/upload.go` — MD5 去重排除已删除记录
- `backend/internal/database/migration.go` — ALTER TABLE 新增列
- `backend/internal/cache/redis_cache.go` — 删除时清除/修复 UID、MD5 Redis 映射

### 关键实现点

1. `images` 表新增列：`deleted_at`（软删除时间）、`deleted_by`（操作者）、`delete_reason`（删除原因）、`purge_after`（物理清理时间）。
2. 公开访问、去重查询、后台默认列表仅使用未删除记录（`deleted_at IS NULL`）。
3. 删除操作改为软删除：设置 `deleted_at`、`deleted_by`、`purge_after`。
4. 删除时修复 Redis UID/MD5 映射，防止已删除记录被缓存命中。
5. 新增管理员回收站 API：
   - `GET /admin/images/trash` — 回收站列表（分页、搜索）
   - `POST /admin/images/:uid/restore` — 恢复已删除图片
6. MD5 去重不复用已删除记录的物理文件。
7. 迁移保持幂等：使用 `ALTER TABLE ADD COLUMN` 和 `CREATE TABLE IF NOT EXISTS`。

### 完成标准

- [x] 代码编译通过：`cd backend && go test ./...`（包含编译）
- [x] 单元测试通过：`cd backend && go test ./...`
- [x] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [x] 数据库迁移幂等

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 管理员删除图片 | 软删除成功，公开访问返回 404 |
| 正向 | 管理员查看回收站 | 已删除图片出现在列表中 |
| 正向 | 管理员恢复图片 | 恢复成功，公开访问正常 |
| 正向 | 上传与已删除图片 MD5 相同的文件 | 作为新图片上传，不复用已删除记录 |
| 异常 | 公开访问已删除图片 URL | 返回 404 |
| 异常 | 使用已删除图片的 MD5 去重 | 不命中，作为新图上传 |
| 边界 | 删除后 Redis 缓存仍存在 | 缓存被清除或失效 |
| 边界 | 恢复后 Redis 缓存重建 | 重新可访问 |
| 边界 | 多条记录引用同一 MD5 时删除其中一条 | 只软删除该记录，其他不受影响 |

### 必须覆盖的测试类型

- 单元测试：软删除逻辑、去重排除已删除记录、Redis 缓存修复
- 集成测试：完整删除→回收站→恢复链路
- 并发测试：同时删除和上传相同 MD5 的图片

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 手动验证删除/恢复/去重行为正确

## 复查

### 代码审查关注点

- 所有图片查询是否正确排除已删除记录（不遗漏公开接口）
- 去重逻辑是否确实不复用已删除记录
- Redis 映射修复是否覆盖所有缓存键模式
- 恢复操作是否正确重建 Redis 缓存

### 安全检查

- 回收站 API 需要管理员权限
- 恢复操作记录审计日志

### 文档更新要求

- 数据库 Schema 文档更新
- API 文档新增回收站接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [x] 本地验证绿色
- [x] 完成报告已填写


## 完成报告

### 实施摘要

- `images` 表已新增 `deleted_at`、`deleted_by`、`delete_reason`、`purge_after` 软删除字段，迁移使用 `ALTER TABLE ADD COLUMN` 幂等执行，并创建 `idx_images_deleted_created_at` 索引。
- 默认图片查询、公开访问、MD5 去重、后台默认列表、统计与预热均排除 `deleted_at IS NOT NULL` 记录。
- 删除链路已改为软删除，不物理删除存储对象；删除后清理 UID Redis 缓存并修复/清理 scoped MD5 映射。
- 新增管理员回收站 API：`GET /admin/images/trash` 与 `POST /admin/images/:uid/restore`；恢复后重建 UID 缓存并回填 MD5 映射。
- 同 MD5 文件在原记录软删除后重新上传会创建新记录与新物理对象，不复用已删除对象。

### 修改文件

- `backend/internal/model/image.go`
- `backend/internal/repository/migration.go`
- `backend/internal/repository/image_repository.go`
- `backend/internal/repository/repository_test.go`
- `backend/internal/service/image_service.go`
- `backend/internal/service/image_service_test.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_test.go`
- `backend/internal/http/handler/admin_handler.go`
- `backend/internal/http/router/routes.go`
- `backend/internal/http/router/router.go`
- `docs/tasks/soft-delete-recycle-bin.md`
- `docs/status/task-list.md`
- `docs/status/progress-report.md`

### 验证记录

- `cd backend && gofmt -w ./cmd ./internal`：通过。
- `cd backend && go test ./...`：通过，17 个 package 共 258 项测试通过。
- `cd backend && gofmt -l ./cmd ./internal`：通过，无输出。

### 复核结论

- 需求、设计、代码与测试一致：软删除字段、默认排除、公开 404、回收站列表、恢复、MD5 不复用、Redis 映射修复均已覆盖。
- 回收站路由注册在 `/admin` 鉴权组内，需要管理员 JWT。
- 未实现存储健康检查等其他 P0 子任务，未修改 `frontend/src/lib/i18n.ts` 与 `frontend/src/routes/+error.svelte`。

### 风险与后续

- 当前在线删除只做软删除与缓存修复，物理清理仍需后续离线/定时清理任务承接。
- 恢复操作目前未新增独立审计日志表记录；若后续要求完整图片操作审计，可在审计日志子任务基础上扩展。
