# 任务：软删除与回收站（已取消并移除）

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` 原 P0 删除生命周期重构条目。
- 2026-05-25 用户确认：**回收站功能不需要**，因此本任务不再作为产品能力保留。

## 预估人日

4 人日（原估算，已取消）

## 当前结论

- 不再提供回收站列表、恢复图片、软删除字段或软删除查询语义。
- 图片删除保持直接删除数据库记录；存储对象仍遵循既有“不主动物理删除对象”的边界。
- 删除后继续清理 UID 缓存并修复/清理 scoped MD5 映射，避免已删记录被缓存命中。
- 不新增 `deleted_at`、`deleted_by`、`delete_reason`、`purge_after` 字段。
- 不注册 `GET /admin/images/trash` 与 `POST /admin/images/:uid/restore` 管理员 API。

## 开发

### 已移除范围

- 后端模型中的删除元数据字段。
- SQLite schema 与索引中的软删除列/索引。
- 仓储层 `FindDeletedByUID`、`SearchDeletedImages`、`SoftDeleteByUID`、`RestoreByUID`。
- 服务层回收站列表、恢复逻辑与软删除逻辑。
- 管理员回收站/恢复路由与 handler。
- 前端图片管理页回收站 tab、恢复按钮、回收站 API helper、类型与文案。
- 对应后端/前端测试用例。

### 完成标准

- [x] 代码中不存在回收站/软删除 API 与数据字段。
- [x] 图片删除链路为直接删除数据库记录并修复缓存映射。
- [x] 前端不展示回收站入口或恢复操作。
- [x] 本地验证通过。

## 测试

### 验证命令

- `cd backend && gofmt -w ./cmd ./internal`
- `cd backend && go test ./...`
- `cd frontend && npm run lint`
- `cd frontend && npm run typecheck`
- `cd frontend && npm run test`
- `cd frontend && npm run build:backend`

## 复查

### 复查关注点

- 删除路径不得再写入或依赖删除元数据字段。
- 管理端不得暴露回收站或恢复入口。
- 公开/管理查询不再拼接软删除过滤条件。
- SQLite 迁移不再创建软删除列或 deleted_at 复合索引。

## 完成报告

### 实施摘要

- 已按用户要求彻底移除回收站功能。
- 后端删除链路由软删除改回直接删除记录，保留 UID 缓存清理与 MD5 映射修复。
- 移除了回收站/恢复 API、路由、service、repository 方法与测试。
- 前端移除了回收站 tab、恢复 API helper、类型与 i18n 文案。
- SQLite schema 不再包含软删除字段，核心索引不再包含 deleted_at 索引。

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
- `frontend/src/lib/api.ts`
- `frontend/src/lib/api.test.ts`
- `frontend/src/lib/types/index.ts`
- `frontend/src/lib/i18n.ts`
- `frontend/src/routes/admin/dashboard/+layout.svelte`
- `frontend/src/routes/admin/dashboard/images/+page.svelte`
- `docs/tasks/soft-delete-recycle-bin.md`
- `docs/tasks/sqlite-core-indexes.md`
- `docs/status/task-list.md`
- `docs/status/progress-report.md`
- `.trellis/tasks/05-22-remediation-extension-implementation/design.md`
- `.trellis/tasks/05-22-remediation-extension-implementation/implement.md`

### 验证记录

- `cd backend && gofmt -w ./cmd ./internal && go test ./...`：通过，250 项后端测试通过。
- `cd frontend && npm run lint`：通过，无 ESLint 问题。
- `cd frontend && npm run typecheck`：通过，0 errors / 0 warnings。
- `cd frontend && npm run test`：通过，10 个测试文件 / 55 项测试通过。
- `cd frontend && npm run build:backend`：通过，前端构建并复制到 `backend/web/`。

### 复核结论

- 需求已更新为“不需要回收站功能”，当前实现与需求一致。
- 代码侧已确认不存在回收站/恢复 API、软删除字段、软删除仓储方法、前端回收站入口或恢复调用。
