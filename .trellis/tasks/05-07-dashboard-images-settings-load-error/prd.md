# 修复 dashboard images/settings 页面加载报错

## Goal

修复 `/admin/dashboard/images/` 与 `/admin/dashboard/settings/` 打开后出现 “This page couldn’t load” 的运行期加载失败，让管理员图片管理和存储设置页面能稳定渲染后端返回的数据。

## What I already know

- 用户报告 `/admin/dashboard/images/` 与 `/admin/dashboard/settings/` 页面显示 “This page couldn’t load”。
- 路由文件存在：`frontend/src/app/admin/dashboard/images/page.tsx` 与 `frontend/src/app/admin/dashboard/settings/page.tsx`。
- 页面只是分别渲染 `ImageTable` 和 `SettingsForm`，崩溃点更可能在客户端组件读取 API 数据时。
- `npm --prefix frontend run typecheck` 通过。
- `npm --prefix frontend run build` 通过，且静态导出包含 `/admin/dashboard/images` 与 `/admin/dashboard/settings`。
- 后端 `/admin/images` 返回 `items`，但前端 `AdminImagesResponse` 与 `ImageTable` 读取 `images`。
- 后端图片字段是 `md5_hash`，但前端 `AdminImage` 与 `ImageTable` 读取 `md5`。
- 后端 `/admin/config` 返回 `storage_configs/default_storage_key`，单项字段为 `storage_backend/local_storage_path/webdav_pass`；前端 `SettingsForm` 读取 `storage_instances/backend/local_path/webdav_password`。

## Requirements

- `/admin/dashboard/images/` 必须能使用后端 `/admin/images` 的实际响应结构渲染图片列表、空列表、分页和预览元数据。
- `/admin/dashboard/settings/` 必须能使用后端 `/admin/config` 的实际响应结构渲染存储实例列表和编辑表单。
- 前端 admin API 类型、读字段和写字段必须与后端 JSON tag 保持一致。
- 修复不得改变管理员登录、token 校验、现有后端 API 路径或静态导出配置。
- 对缺失可选字段保持安全展示，不因某个图片/存储字段为空而整页崩溃。

## Acceptance Criteria

- [ ] `/admin/dashboard/images/` 不再出现 “This page couldn’t load”。
- [ ] `/admin/dashboard/settings/` 不再出现 “This page couldn’t load”。
- [ ] 图片管理页使用 `items` 列表与 `md5_hash` 字段，不再读取不存在的 `images`/`md5`。
- [ ] 存储设置页使用 `storage_configs`、`storage_backend`、`local_storage_path`、`webdav_pass` 等后端字段。
- [ ] 新建/更新/删除/设默认存储实例的请求 payload 使用后端期望字段。
- [ ] `npm --prefix frontend run lint` 通过。
- [ ] `npm --prefix frontend run typecheck` 通过。
- [ ] `npm --prefix frontend run build` 通过。

## Definition of Done

- Tests/checks: frontend lint、typecheck、build 通过。
- 如可运行 UI，则至少手动打开 dashboard 首页、images、settings 验证路由渲染；若无法启动/登录，应明确说明限制。
- 不引入无关 UI 重构或后端 API 变更。

## Technical Approach

对齐前端 admin 类型与后端实际 JSON 合约，优先修正前端 API 类型和组件字段读取/提交 payload；保持后端合约不变，避免破坏已有服务端测试与 API。

## Decision (ADR-lite)

**Context**: 页面编译/导出正常，但运行期读取不存在字段会触发 React 错误边界；根因是前后端 admin 数据合约漂移。

**Decision**: 修复前端类型和组件以匹配后端现有 JSON 字段，而不是修改后端字段兼容前端旧命名。

**Consequences**: 前端 admin 类型会更接近后端真实 API；本次只处理 dashboard images/settings 报错，不扩大到完整 API 版本化或 schema 生成。

## Out of Scope

- 不重构 dashboard 整体布局。
- 不修改 admin 认证流程。
- 不新增后端 API 路径。
- 不处理与本次页面崩溃无关的视觉样式问题。

## Technical Notes

- 相关前端文件：`frontend/src/features/admin/ImageTable.tsx`、`frontend/src/features/admin/SettingsForm.tsx`、`frontend/src/lib/api.ts`、`frontend/src/types/index.ts`。
- 后端响应来源：`backend/internal/service/admin_service.go` 与 `backend/internal/http/handler/admin_handler.go`。
- 静态前端深层路由由 `backend/internal/http/router/frontend.go` 支持，当前构建已生成对应路由。
