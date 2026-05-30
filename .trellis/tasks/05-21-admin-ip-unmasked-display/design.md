# Design: 后台 IP 取消脱敏展示

## Scope and boundaries

本任务同时修改 backend、frontend 与 `.trellis/spec`，目标是让后台管理端彻底停止使用 `ip_address_masked`，统一使用真实 `ip_address`。

边界如下：

- **backend model/service/repository/API**：删除 `ip_address_masked` 字段与生成逻辑。
- **SQLite schema**：`ip_bans` 新结构不再包含 `ip_address_masked`。
- **frontend admin UI/types**：后台图片列表、详情、滥用排行、IP 封禁列表、IP 详情、确认弹窗统一展示 `ip_address`。
- **spec**：同步更新 backend security/database 与 frontend type-safety 等文档。

不包含：

- 公共上传端或访客侧 IP 展示策略调整。
- 旧版 SQLite `ip_bans` 表的兼容迁移；旧库允许手动删库/重建。

## Technical design

### 1. Backend contract shrink

移除以下响应/模型中的 `ip_address_masked`：

- `backend/internal/model/abuse.go`
  - `AbuseIPRankItem`
  - `AbuseIPDetail`
- `backend/internal/model/ip_ban.go`
  - `IPBan`
- `backend/internal/service/admin_service.go`
  - `AdminImageItem`
- 相关 JSON 响应会自然收口到只返回 `ip_address`

`AdminImageItem` 当前是 service 层为后台图片列表拼装的 view model。这里直接删掉 `IPAddressMasked`，避免后端继续承担后台展示脱敏职责。

### 2. Backend logic cleanup

删除后台 IP 脱敏生成逻辑：

- `backend/internal/service/admin_service.go`
  - 图片列表构造时不再调用 `maskIPAddress(item.IPAddress)`
- `backend/internal/service/security_analysis.go`
  - 创建封禁时不再写入 `IPAddressMasked`
  - `Overview()` 的 IP 排行不再填充 masked 字段
  - `IPDetail()` 不再填充 masked 字段
  - 默认原因中的 IP 直接使用真实 `ip_address`
- `backend/internal/service/ip_utils.go`
  - 若仅剩 `ipHash()` 被使用，则删除 `maskIPAddress()` 包装

说明：`backend/internal/iputil/` 可以保留，因为其中哈希能力仍用于 ban lookup；本任务只删除后台使用的 masking 业务契约，不强行删除底层工具包里所有 mask 实现。

### 3. Repository + SQLite schema change

`ip_bans` 仓储改为只读写真实 IP：

- `backend/internal/repository/ip_ban_repository.go`
  - `INSERT` 删除 `ip_address_masked`
  - `SELECT` 删除 `ip_address_masked`
  - `scanIPBan()` 删除对应扫描目标

`backend/internal/repository/migration.go` 中的 `CREATE TABLE IF NOT EXISTS ip_bans` 改为：

- `id`
- `ip_hash`
- `ip_address`
- `reason`
- `expires_at`
- `created_at`
- `updated_at`

不新增旧表重建逻辑。

这与仓库现有 migration 原则一致：对已废弃列不做兼容重建，旧开发库由人工重置。

### 4. Frontend admin contract update

更新 `frontend/src/lib/types/index.ts`，删除以下字段：

- `AdminImage.ip_address_masked`
- `AdminIPBan.ip_address_masked`
- `AdminAbuseIPRankItem.ip_address_masked`
- `AdminAbuseIPDetail.ip_address_masked`

随后同步修改依赖这些类型的组件与页面：

- `frontend/src/routes/admin/dashboard/images/+page.svelte`
- `frontend/src/routes/admin/dashboard/security/+page.svelte`
- `frontend/src/lib/components/studio/ImageDetailDrawer.svelte`
- `frontend/src/lib/components/studio/IPDetailPanel.svelte`

交互规则：

- `BanIPDialog` 的 `target.label` 统一传真实 IP。
- 表格 `title`、确认弹窗 `description`、按钮 `aria-label` 统一使用 `ip_address`。
- 创建封禁成功后的本地状态回填只保留后端仍返回的字段。

### 5. Spec sync

至少更新以下文档中的契约：

- `.trellis/spec/backend/database-guidelines.md`
- `.trellis/spec/backend/security.md`
- `.trellis/spec/frontend/type-safety.md`

必要时补充 frontend/backend 其他文档中的描述，确保不再把 `ip_address_masked` 视为当前有效契约。

## Compatibility and migration notes

- **不兼容旧版 `ip_bans` 表结构**。
- 若本地或部署环境 SQLite 仍含 `ip_address_masked` 列，本次不提供自动迁移；需删库或手动重建该表。
- 该策略与仓库现有 `repository_test.go` 中“已废弃列不做兼容重建”的约定一致。

## Rollback considerations

- 若前端页面异常，可优先回滚 `frontend/src/lib/types/index.ts` 与后台安全页面改动。
- 若 backend API 契约导致前端构建失败，可临时先恢复后端字段，再分步收口。
- 若本地旧库启动失败，按任务约定重建 SQLite，而不是继续追加兼容迁移逻辑。
