# Cloudflare config admin hot reload

## Goal

将 Cloudflare 单 URL 图片缓存清理所需配置从仅启动时环境变量读取，改为可通过后台管理运行时查看/编辑并即时生效，让管理员无需重启服务即可启用、停用或更换 Cloudflare Zone/API Token/API Base URL。

## Confirmed Facts

- 当前 Cloudflare 缓存清理客户端在 `backend/internal/service/cloudflare_cache.go` 中实现，调用 `POST /zones/{zone_id}/purge_cache`，并要求 Zone ID 与 API Token 已配置；Cloudflare API 的 `files` 参数允许一次提交多个 URL。
- 当前启动流程在 `backend/cmd/server/main.go` 中通过 `config.Load()` 读取 `CLOUDFLARE_ZONE_ID`、`CLOUDFLARE_API_TOKEN`、`CLOUDFLARE_API_BASE_URL`，然后只调用一次 `imageService.SetImageURLCachePurger(...)`。
- 当前运行时设置只持久化 `cloudflare_purge_enabled`，后台设置页只能编辑启用开关、查看 `cloudflare_purge_configured` 状态、执行手动单 URL 清理。
- 当前文档明确写着 Cloudflare 凭据不进入 SQLite、不返回前端；本任务会改变该约束，需要同步更新 spec/docs。
- Runtime settings 已通过 SQLite `config` 表持久化，并由后台 `GET|PUT /admin/system-settings` 编辑；已有秘密字段遮罩/保持逻辑可参考存储配置的 masked secret 处理。

## Requirements

- 在后台管理的系统/运行时设置中新增 Cloudflare 配置管理：
  - Cloudflare Zone ID；
  - Cloudflare API Token；
  - Cloudflare API Base URL（空值使用默认 `https://api.cloudflare.com/client/v4`）。
- 保存配置后必须热更新后端 Cloudflare cache purger；后续所有删除路径的自动清理、后台手动清理都使用最新配置，无需重启服务。
- API Token 必须按敏感信息处理：后台 GET 只能返回遮罩值或空值；PUT 发送未变更的遮罩值时必须保留已有 Token；空值按“清空 Token”处理。
- `cloudflare_purge_enabled=true` 时，仍必须要求 `public_base_url` 有效；同时必须有可用的 Zone ID 与 API Token，否则保存应失败并返回 `invalid_input`。
- 所有图片删除路径都必须触发 Cloudflare 缓存清理：前台用户删除、后台单张删除、后台批量删除。
- 后台批量删除多个图片时，应尽量把这些图片 URL 合并到一次 Cloudflare `files` 数组请求中，而不是逐张调用 Cloudflare。
- `cloudflare_purge_enabled=false` 时，允许保存不完整或空的 Cloudflare 凭据，方便管理员先草稿配置或关闭功能。
- Cloudflare API Base URL 若非空，必须是合法 `http` 或 `https` URL；保存时去除首尾空白和末尾 `/`。
- 程序首次运行或缺失 key 时必须写入默认 runtime config，且不得覆盖已有配置。
- 现有 `CLOUDFLARE_*` 环境变量不作为后台配置迁移来源；升级后管理员需要在后台重新填写 Cloudflare 配置。
- 后端启动配置应移除或停止使用 `CLOUDFLARE_ZONE_ID` / `CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_API_BASE_URL`，避免环境变量与后台热更新产生双来源。
- Public runtime settings 不再同时暴露重复的 `effective_allowed_mime_types`；公开上传配置只保留单一 `allowed_mime_types` 字段。
- TypeScript 类型、API 测试、后端服务测试和文档/spec 需同步更新。

## Acceptance Criteria

- [ ] `RuntimeSettings` / `RuntimeSettingsUpdateInput` / 前端 `RuntimeSettings` 类型包含 Cloudflare Zone ID、API Token、API Base URL 管理字段。
- [ ] `GET /admin/system-settings` 返回 Cloudflare 配置状态，并且 API Token 不以明文返回。
- [ ] `PUT /admin/system-settings` 能保存 Cloudflare 配置；遮罩 token 保留旧值，空 token 清空旧值。
- [ ] 保存成功后，后续 Cloudflare 手动 purge 与所有删除路径的自动 purge 立即使用新 Zone ID/API Token/API Base URL。
- [ ] 启用 `cloudflare_purge_enabled` 时缺少 public base URL、Zone ID 或 API Token 会返回 `invalid_input`，且不保存部分配置。
- [ ] `cloudflare_api_base_url` 非空且不是 http/https URL 时返回 `invalid_input`。
- [ ] 后台设置页可编辑 Cloudflare Zone ID、API Token、API Base URL，并随现有保存按钮一起提交。
- [ ] 后台批量删除多个图片时 Cloudflare 请求体为 `{ "files": ["url1", "url2", ...] }`，同批次 URL 尽量一次提交。
- [ ] 文档和 `.trellis/spec/` 不再声明 Cloudflare 凭据只能来自环境变量。
- [ ] `.env.example` 与部署文档不再要求 `CLOUDFLARE_*` 作为启动环境变量。
- [ ] `GET /v1/runtime-settings.upload` 不再返回与 `allowed_mime_types` 重复的 `effective_allowed_mime_types`。
- [ ] 前端首页/上传队列改用 `upload.allowed_mime_types`，不再依赖 `effective_allowed_mime_types`。
- [ ] `cd backend && go test ./...` 通过；前端改动后 `cd frontend && npm run typecheck` 与相关测试通过。

## Out of Scope

- 不做 Cloudflare `purge_everything`、前缀 purge 或非图片删除场景的全站缓存清理。
- 不新增独立的后台“批量输入 URL 清理”工具；批量 purge 仅服务于图片批量删除流程。
- 不新增多套 Cloudflare profile 或按存储实例拆分 Cloudflare 配置。
- 不把 Cloudflare API Token 暴露给公开 API 或非管理员前端。

## Decisions

- 现有环境变量 `CLOUDFLARE_ZONE_ID` / `CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_API_BASE_URL` 不迁移、不回退。新增后台配置默认空值，管理员升级后在后台重新填写。
