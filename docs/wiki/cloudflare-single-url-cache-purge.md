# Cloudflare 图片 URL 缓存清理

## 目标

OmePic 支持通过 Cloudflare API 清理图片 URL 的边缘缓存。Cloudflare Zone ID、API Token 与 API Base URL 由管理员在后台运行时设置中维护，保存后立即生效，无需重启后端。

核心约束：

- 管理后台使用 JWT 保护后台接口。
- 前台上传继续由浏览器生成客户端 token，并通过 `X-Token` 上传和删除。
- Redis 继续负责 UID 元数据与按 `storage_key + original_md5` 的秒传去重缓存。
- Cloudflare 清理只处理图片 URL；手动接口每次单 URL，后台批量删除会尽量合并到一次 `files` 数组请求；不做全站、前缀或 `purge_everything`。

## 配置

公开运行时上传设置 `GET /v1/runtime-settings.upload` 仅保留一个 MIME 列表字段：`allowed_mime_types`。前端首页、上传队列和文件选择 accept 列表统一读取该字段，不再返回重复的 `effective_allowed_mime_types`。

### Runtime settings

Cloudflare 配置保存在 SQLite `config` 表，并通过 `GET|PUT /admin/system-settings` 管理：

| Key | 类型 | 默认值 | 说明 |
|-----|------|--------|------|
| `cloudflare_purge_enabled` | bool | `false` | 是否启用删除图片前的 Cloudflare 图片 URL 缓存清理 |
| `cloudflare_zone_id` | string | 空 | Cloudflare Zone ID |
| `cloudflare_api_token` | string | 空 | 需要具有 Zone Cache Purge 权限的 API Token；admin GET 只返回遮罩值 |
| `cloudflare_api_base_url` | string | 空 | 空值使用 `https://api.cloudflare.com/client/v4`；非空必须是合法 http/https URL |
| `public_base_url` | string | 空 | 启用 Cloudflare 清理时必须配置，用于拼出 `/i/{uid}.avif` 的公开 URL |

`cloudflare_api_base_url` 保存时会 trim 首尾空白并去掉末尾 `/`。

### Token 遮罩与保留

- SQLite 存储真实 `cloudflare_api_token`。
- `GET /admin/system-settings` 返回空值或遮罩值（短值为 `****`，长值仅露出末尾 4 位）。
- `PUT /admin/system-settings` 中：
  - 发送未改变的遮罩值会保留旧 Token；
  - 发送空值会清空 Token；
  - 发送其他值会保存为新 Token。

### 启动环境变量

`CLOUDFLARE_ZONE_ID`、`CLOUDFLARE_API_TOKEN`、`CLOUDFLARE_API_BASE_URL` 不再作为后端启动配置读取，也不会作为后台配置迁移来源。升级后需管理员在后台设置页重新填写 Cloudflare 配置。

## 后端设计

### Cloudflare API 客户端

后端 `CloudflareCachePurger` 调用 Cloudflare purge cache `files` API：

```http
POST /zones/{zone_id}/purge_cache
Authorization: Bearer <cloudflare_api_token>
Content-Type: application/json
```

```json
{ "files": ["https://img.example.com/i/uid-1.avif", "https://img.example.com/i/uid-2.avif"] }
```

发送请求前会规范化 URL：

- 去除首尾空白；
- 仅允许 `http` / `https`；
- 必须存在 host；
- 移除 fragment，例如 `#preview`。

### 热更新路径

`ImageService.PurgeImageURLCache` 每次调用都会从当前 `RuntimeSettingsManager.Current()` 读取 Cloudflare Zone ID、API Token 与 API Base URL 并构造 purger。因此：

- 后台保存设置后，手动 purge 立即使用新配置；
- 删除图片触发的自动 purge 立即使用新配置；
- 后端启动流程不再注入环境变量驱动的固定 purger。

### 删除图片流程

当 `cloudflare_purge_enabled=false` 时，删除流程保持原行为。

当 `cloudflare_purge_enabled=true` 时：

1. 读取当前运行时配置。
2. 校验 `public_base_url` 已配置。
3. 使用 `public_base_url` 拼接目标图片 URL：`{public_base_url}/i/{uid}.avif`。
4. 使用当前 Cloudflare runtime 配置清理目标 URL。
5. 只有 Cloudflare 清理成功后才删除 SQLite 记录和 Redis UID/MD5 缓存。
6. 如果 Cloudflare 清理失败，返回错误并保留图片记录。

后台批量删除多张图片时，会先查出待删除记录并构造全部图片 URL，尽量一次提交 `{ "files": [...] }`；清理成功后再逐条删除记录与缓存，避免逐张调用 Cloudflare。

## 管理后台 API

### 系统设置

`GET /admin/system-settings` 的 `runtime` 包含：

```json
{
  "cloudflare_purge_enabled": false,
  "cloudflare_zone_id": "",
  "cloudflare_api_token": "",
  "cloudflare_api_base_url": ""
}
```

配置过 Token 后，`cloudflare_api_token` 只返回遮罩值。

`PUT /admin/system-settings` 保存同一 runtime 结构。启用 `cloudflare_purge_enabled=true` 时必须同时具备合法 `public_base_url`、非空 `cloudflare_zone_id` 和非空 `cloudflare_api_token`，否则返回 `invalid_input` 且不保存部分配置。

### 手动清理单 URL

```http
POST /admin/cloudflare/purge-image-cache
Authorization: Bearer <jwt>
Content-Type: application/json
```

请求体：

```json
{ "url": "https://img.example.com/i/uid.avif" }
```

响应：

```json
{
  "success": true,
  "data": {
    "url": "https://img.example.com/i/uid.avif"
  }
}
```

说明：

- 必须先启用 `cloudflare_purge_enabled`。
- 必须配置 `public_base_url`。
- 必须配置 Cloudflare Zone ID 与 API Token。
- 该手动接口每次只提交一个 URL 给 Cloudflare；后台批量删除会使用多 URL `files` 数组。

## 前端设计

管理后台设置页 Cloudflare 区块包含：

- `cloudflare_purge_enabled` 开关；
- Zone ID 输入框；
- API Token 密码输入框（显示遮罩值，清空表示清除 Token）；
- API Base URL 输入框（留空使用默认 Cloudflare API）；
- 只读配置状态：`readonly.service.cloudflare_purge_configured`；
- 手动清理 URL 输入框和按钮。

## 错误处理

| 场景 | 错误类别 | 行为 |
|------|----------|------|
| 手动 purge URL 为空或不是 http/https | `invalid_input` | 请求被拒绝，不访问 Cloudflare |
| 启用清理但 `public_base_url` 为空 | `invalid_input` | 保存设置或清理请求失败 |
| 启用清理但 Zone ID / API Token 为空 | `invalid_input` | 保存设置失败，不保存部分配置 |
| `cloudflare_api_base_url` 非空但不是 http/https URL | `invalid_input` | 保存设置失败，不保存部分配置 |
| Cloudflare 返回非 2xx 或 `success=false` | `dependency_unavailable` | 清理失败，删除流程停止 |
| 网络/响应解析失败 | `dependency_unavailable` | 清理失败，删除流程停止 |

## 验证命令

```powershell
cd backend
go test ./...

cd ../frontend
npm run typecheck
npm test -- --run src/lib/api.test.ts
npm run build:backend
```
