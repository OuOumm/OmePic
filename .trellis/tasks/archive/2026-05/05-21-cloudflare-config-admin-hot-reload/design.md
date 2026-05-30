# Design: Cloudflare config admin hot reload

## Scope

将 Cloudflare 单 URL purge 的 Zone ID、API Token、API Base URL 纳入 SQLite runtime settings，并通过 `GET|PUT /admin/system-settings` 与后台设置页编辑。保存后更新内存中的 runtime settings，后续手动 purge 与删除图片自动 purge 使用最新配置。

## Data Model / Persistence

在现有 SQLite `config` key/value 表新增 runtime keys：

- `cloudflare_zone_id`: string，默认空。
- `cloudflare_api_token`: string，默认空，敏感值。
- `cloudflare_api_base_url`: string，默认空；空值在 Cloudflare purger 内解析为 `https://api.cloudflare.com/client/v4`。

`RuntimeSettingsManager.Load(ctx, repo)` 继续通过 insert-missing semantics 写入缺失默认值，不覆盖已有值。

不从 `CLOUDFLARE_*` 环境变量迁移或回退。升级后管理员在后台填写配置，避免后台热更新与环境变量双来源。

## Backend Contracts

扩展：

- `RuntimeSettings`
- `RuntimeSettingsUpdateInput`
- `runtimeConfigFields`
- `RuntimeSettingsToConfigValues`
- `runtimeSettingsFromValues`
- `ValidateRuntimeSettingsInput`

Validation：

- `cloudflare_api_base_url` 非空时必须是合法 `http` / `https` URL，保存时 trim 并去掉末尾 `/`。
- `cloudflare_purge_enabled=true` 时：
  - `public_base_url` 必须合法且非空；
  - `cloudflare_zone_id` 必须非空；
  - `cloudflare_api_token` 必须非空。
- validation 必须在任何 DB 写入前完成，失败返回 `invalid_input` 且无部分保存。

## Secret Masking Contract

API Token 是 admin-only secret：

- SQLite 存储真实 token。
- `GET /admin/system-settings` 的 `runtime.cloudflare_api_token` 返回遮罩值或空值。
- 遮罩逻辑复用 `maskSecret` 语义：空值返回空；短值返回 `****`；长值仅暴露尾部 4 位。
- `PUT /admin/system-settings`：
  - 若传入值等于当前已存 token 的遮罩值，则保留旧 token；
  - 若传入空值，则清空 token；
  - 其他值按新 token 保存。

为实现保留旧 token，`AdminService.UpdateSystemSettings` 需要在 validation 前读取当前完整 runtime settings 或 SQLite config，并合并 token 字段后再验证与保存。`RuntimeSettingsManager.Current()` 可以保留完整 token 供后端使用；返回 admin view 时做脱敏。

## Cloudflare Purger Hot Reload

推荐简化现有 `ImageURLCachePurger` 使用方式：

- 保留 `CloudflareCachePurger` 作为单次配置客户端，但接口需支持 `PurgeURLs(ctx, []string)`，并让单 URL `PurgeURL` 委托给多 URL 方法。
- `CloudflareCachePurger` 请求体使用 Cloudflare 支持的 `files` 数组：`{ "files": ["url1", "url2"] }`。
- 新增或使用 factory/helper，在 `ImageService.PurgeImageURLCache` / `PurgeImageURLCaches` / `CloudflarePurgeConfigured` 调用时从 `RuntimeSettingsManager.Current()` 取当前 Cloudflare 配置，构造 purger 并调用。
- 或新增 `RuntimeCloudflareCachePurger`，内部持有 `RuntimeSettingsManager`，每次 purge 动态读取当前 settings。

关键点：

- 不再在 `cmd/server/main.go` 用启动环境变量固定注入一次 purger。
- `CloudflarePurgeConfigured()` 反映当前 runtime settings 中 Zone ID + API Token 是否完整。
- 前台用户删除、后台单张删除、后台批量删除和后台手动 purge 都调用同一条动态配置路径。
- 后台批量删除多个图片时应先查出待删除记录，构造 `{public_base_url}/i/{uid}.avif` URL 列表，一次调用 Cloudflare `files` 数组 purge；Cloudflare 成功后再逐条删除 SQLite/Redis/MD5 缓存，避免逐张调用 Cloudflare。

## Startup Config Compatibility

`backend/internal/config.AppConfig` 应移除或停止暴露 `CloudflareZoneID` / `CloudflareAPIToken` / `CloudflareAPIBaseURL`。`.env.example`、部署文档、config tests 同步更新，避免继续宣传环境变量配置。

## Public Runtime MIME Contract

当前 `allowed_mime_types` 与 `effective_allowed_mime_types` 在行为上重复。此任务顺带收敛公开 contract：

- `GET /v1/runtime-settings.upload` 仅保留 `allowed_mime_types`。
- 前端首页、上传队列、文件选择 accept 列表统一读取 `upload.allowed_mime_types`。
- 后端 upload validation 继续基于 runtime settings 的 `AllowedMIMETypes`；不再需要对外暴露重复的 `effective_allowed_mime_types`。

## Frontend

扩展 `frontend/src/lib/types/index.ts` 的 `RuntimeSettings`：

- `cloudflare_zone_id: string`
- `cloudflare_api_token: string`
- `cloudflare_api_base_url: string`

后台 settings 页 Cloudflare 区块新增输入控件：

- Zone ID 文本输入；
- API Token 密码/文本输入，显示后端遮罩值；保存遮罩值时后端保留旧 token；清空输入表示清空 token；
- API Base URL 文本输入，提示空值使用默认 Cloudflare API。

更新 en/zh i18n 文案。手动 purge 按钮继续依赖 `readonly.service.cloudflare_purge_configured`，该值由当前 runtime settings 计算。

## Docs / Specs

需要更新：

- `.trellis/spec/backend/runtime-settings.md`
- `.trellis/spec/frontend/type-safety.md`
- `docs/cloudflare-single-url-cache-purge.md`
- `docs/api-reference.md`
- `docs/running-and-deployment.md`
- `.env.example`

删除“Cloudflare 凭据只能来自环境变量 / 不写入 SQLite”的旧约束，改为“admin-only runtime secret，返回时脱敏”。同时删除 public runtime settings 中重复的 `effective_allowed_mime_types` 文档描述。

## Batch Delete Flow

新增/调整服务层批量删除路径：

1. `AdminService.DeleteImages` 接收 UID 列表。
2. 若 Cloudflare purge 未启用，保持逐张删除即可。
3. 若启用，先按 UID 查出所有图片记录并校验存在；使用当前 `public_base_url` 生成所有目标图片 URL。
4. 一次调用 `ImageService.PurgeImageURLCaches(ctx, urls)`，该方法使用当前 runtime Cloudflare 配置提交 `files` 数组。
5. purge 成功后再删除每个图片记录和对应 Redis UID/MD5 缓存。若 purge 失败，不删除任何记录。

前台用户删除和后台单张删除继续走单 URL 路径，但底层可复用 `PurgeImageURLCaches`。

## Compatibility / Rollback

- 现有部署升级后 Cloudflare purge 默认变为未配置；管理员需在后台重新填写。
- 已有 `cloudflare_purge_enabled=true` 但未填写新 runtime 凭据的数据库，在 load/validate 时可能失败；实现时应确保缺失新 keys 会先写入空默认值，但启用状态下需要管理员修正配置。若启动 Load 因已有 enabled=true 且新凭据空导致失败，可在实现中对 load 路径保留可启动策略：读取旧值时不因启用但缺凭据阻断进程，PUT 保存时严格校验。
- 回滚可恢复启动环境变量注入 purger，并移除新 runtime fields。
