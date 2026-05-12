# API 参考

## 目录

1. [通用约定](#1-通用约定)
2. [公开 API](#2-公开-api)
3. [管理员 API](#3-管理员-api)
4. [错误码](#4-错误码)

---

## 1. 通用约定

### 基础 URL

- 开发环境: `http://localhost:8080`
- 可通过 `PUBLIC_BASE_URL` 环境变量自定义

### 响应格式

所有响应统一格式：

```json
// 成功
{ "success": true, "data": { ... } }

// 失败
{ "success": false, "error": { "code": "error_code", "message": "Human readable message" } }
```

### 认证方式

- **管理员认证**: `Authorization: Bearer <jwt>`（通过 `/admin/login` 获取）
- **用户认证**: `X-Token: <client-token>`（客户端随机生成，标识上传者）

### 速率限制

失败响应包含以下响应头：
- `X-RateLimit-Limit`: 限制上限
- `X-RateLimit-Remaining`: 剩余次数
- `Retry-After`: 建议等待秒数

当超限时返回 HTTP 429。

---

## 2. 公开 API

### 2.1 健康检查

```
GET /health
```

无需认证。检查 SQLite 和 Redis 连通性。

**响应** (200):
```json
{ "success": true, "data": { "status": "ok" } }
```

**响应** (503):
```json
{ "success": false, "error": { "code": "dependency_unavailable", "message": "sqlite unavailable" } }
```

### 2.2 获取运行时设置

```
GET /v1/runtime-settings
```

公开的运行时配置，包括站点信息、上传限制和存储选项。

**响应** (200):
```json
{
  "success": true,
  "data": {
    "site": {
      "name": "OmePic",
      "tagline": "上传、分享和管理图片"
    },
    "access": {
      "public_base_url": "http://localhost:8080"
    },
    "upload": {
      "max_upload_size_mb": 20,
      "allowed_mime_types": ["image/jpeg", "image/png", "image/gif", "image/webp", "image/avif"],
      "effective_allowed_mime_types": ["image/jpeg", "image/png", "image/gif", "image/webp", "image/avif"]
    },
    "features": {
      "allow_storage_selection": true,
      "maintenance_mode": false,
      "maintenance_message": ""
    },
    "storage": {
      "options": [
        { "storage_key": "local-default", "name": "Default Local Storage", "storage_backend": "local", "is_default": true }
      ]
    }
  }
}
```

### 2.3 获取公告

```
GET /v1/announcements
```

公开的已发布且有效时间窗口内的公告。

**响应** (200):
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "title": "Notice",
        "content": "**Markdown** content",
        "priority": "normal",
        "starts_at": null,
        "ends_at": null,
        "created_at": "2025-01-01T00:00:00Z",
        "updated_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 2.4 上传图片

```
POST /v1/image
```

**请求头**: `X-Token: <client-token>`（必填）

**请求体**: `multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | File | 是 | 图片文件 |
| `storage_key` | String | 否 | 指定存储实例（需 `allow_storage_selection` 开启） |

**响应** (200):
```json
{
  "success": true,
  "data": {
    "uid": "abc123def456",
    "url": "http://localhost:8080/i/abc123def456.avif",
    "md_url": "![image](http://localhost:8080/i/abc123def456.avif)",
    "bbcode": "[img]http://localhost:8080/i/abc123def456.avif[/img]",
    "size": 12345,
    "mime_type": "image/avif",
    "created_at": "2025-01-01T00:00:00Z",
    "duplicate": false,
    "storage_key": "local-default",
    "storage_backend": "local"
  }
}
```

**说明**:
- 上传的图片会自动转换为 AVIF 格式
- `duplicate` 为 true 表示 MD5 去重命中，未创建新文件
- 可通过 `X-Token` 头传递上传者标识（用于后续删除验证）

### 2.5 获取图片

```
GET /i/:uid.avif
```

直接返回图片文件内容。

**响应头**:
- `Content-Type: image/avif`
- `Cache-Control: public, max-age=31536000, immutable`
- `Content-Disposition: inline`

**响应** (404): 图片不存在时返回空 404

### 2.6 删除图片

```
DELETE /i/:uid.avif
```

**请求头**: `X-Token: <client-token>`（必填，需与上传时的 token 一致）

**响应** (200):
```json
{ "success": true, "data": {} }
```

**说明**:
- 逻辑删除：仅删除数据库记录和 Redis 缓存，保留物理文件
- 需使用原上传时的 X-Token 才能删除

---

## 3. 管理员 API

所有管理员 API 需在请求头携带 `Authorization: Bearer <jwt>`。

### 3.1 管理员登录

```
POST /admin/login
```

**请求体**:
```json
{ "password": "admin123" }
```

**响应** (200):
```json
{ "success": true, "data": { "token": "eyJhbGciOiJIUzI1NiIs..." } }
```

JWT 有效期为 24 小时。

### 3.2 获取全局统计

```
GET /admin/status
```

**响应** (200):
```json
{
  "success": true,
  "data": {
    "total_images": 1000,
    "total_storage_size": 104857600,
    "today_uploads": 25,
    "unique_tokens": 50
  }
}
```

### 3.3 图片管理列表

```
GET /admin/images?page=1&pageSize=20&search=abc
```

**参数**:
- `page`: 页码（默认 1）
- `pageSize`: 每页数量（默认 20，最大 100）
- `search`: 搜索关键词（支持 uid, token, ip, md5, storage_key）

**响应** (200):
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "uid": "abc123",
        "token": "client-token-xxx",
        "storage_key": "local-default",
        "storage_backend": "local",
        "mime_type": "image/avif",
        "size": 12345,
        "md5_hash": "d41d8cd98f00b204e9800998ecf8427e",
        "ip_address": "192.168.1.1",
        "ip_address_masked": "192.168.1.*",
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 100
  }
}
```

### 3.4 批量删除图片

```
DELETE /admin/images
```

**请求体**:
```json
{ "uids": ["abc123", "def456"] }
```

**响应** (200):
```json
{ "success": true, "data": {} }
```

### 3.5 IP 封禁管理

#### 创建封禁

```
POST /admin/ip-bans
```

**请求体**:
```json
{
  "uid": "abc123",
  "ip_address": "192.168.1.1",
  "duration_hours": 24,
  "reason": "Abusive upload"
}
```

**说明**: `uid` 和 `ip_address` 至少提供一个。提供 `uid` 时自动使用该图片的 IP。

**响应** (200):
```json
{
  "success": true,
  "data": {
    "ban": { "id": 1, "ip_hash": "...", "ip_address": "192.168.1.1", "ip_address_masked": "192.168.1.*", "reason": "Abusive upload", "expires_at": "2025-01-02T00:00:00Z", ... },
    "affected_image_count": 10,
    "affected_total_size": 500000
  }
}
```

#### 查询封禁列表

```
GET /admin/ip-bans
```

#### 删除封禁

```
DELETE /admin/ip-bans/:id
```

#### 删除封禁关联图片

```
DELETE /admin/ip-bans/:id/images
```

**响应** (200):
```json
{ "success": true, "data": { "deleted_count": 10 } }
```

### 3.6 滥用监控

#### 滥用概览

```
GET /admin/abuse/overview?from=2025-01-01T00:00:00Z&to=2025-01-02T00:00:00Z
```

**参数**:
- `from`: 起始时间（RFC3339）
- `to`: 结束时间（RFC3339）

**响应** (200):
```json
{
  "success": true,
  "data": {
    "from": "2025-01-01T00:00:00Z",
    "to": "2025-01-02T00:00:00Z",
    "upload_count": 100,
    "upload_size": 5000000,
    "active_ip_ban_count": 5,
    "top_ips": [
      { "ip_address": "192.168.1.1", "ip_address_masked": "192.168.1.*", "upload_count": 50, "total_size": 250000, "latest_upload_at": "2025-01-01T12:00:00Z", "is_banned": true, "ban_id": 1 }
    ],
    "top_tokens": [
      { "token": "abc123...", "token_preview": "abc123...", "upload_count": 30, "total_size": 150000, "latest_upload_at": "2025-01-01T10:00:00Z" }
    ]
  }
}
```

#### IP 详情

```
GET /admin/abuse/ip?ip=192.168.1.1
```

**响应** (200):
```json
{
  "success": true,
  "data": {
    "ip_address": "192.168.1.1",
    "ip_address_masked": "192.168.1.*",
    "upload_count": 50,
    "total_size": 250000,
    "is_banned": true,
    "ban": { "id": 1, ... }
  }
}
```

### 3.7 存储配置管理

#### 获取存储配置

```
GET /admin/config
```

**响应** (200):
```json
{
  "success": true,
  "data": {
    "default_storage_key": "local-default",
    "storage_configs": [
      {
        "storage_key": "local-default",
        "name": "Default Local Storage",
        "is_default": true,
        "storage_backend": "local",
        "local_storage_path": "data/images",
        "s3_endpoint": "",
        "s3_region": "",
        "s3_bucket": "",
        "s3_access_key": "",
        "s3_secret_key": "",
        "s3_use_ssl": false,
        "s3_force_path_style": false,
        "webdav_url": "",
        "webdav_user": "",
        "webdav_pass": ""
      }
    ]
  }
}
```

**注意**: 秘密字段（s3_access_key, s3_secret_key, webdav_pass）返回时会被脱敏（仅保留后 4 字符）。

#### 更新存储配置（兼容路由）

```
POST /admin/config
```

支持部分更新和修改默认存储。

#### 创建存储实例

```
POST /admin/config/storage-instances
```

**请求体**:
```json
{
  "storage_key": "my-s3",
  "name": "My S3 Storage",
  "storage_backend": "s3",
  "s3_endpoint": "s3.amazonaws.com",
  "s3_bucket": "my-bucket",
  "s3_access_key": "AKIA...",
  "s3_secret_key": "secret...",
  "s3_region": "us-east-1"
}
```

#### 更新存储实例

```
PUT /admin/config/storage-instances/:storageKey
```

支持部分更新（只传需要修改的字段）。

**注意**: 不允许修改有图片引用的存储实例的后端类型。

#### 删除存储实例

```
DELETE /admin/config/storage-instances/:storageKey
```

**限制**: 不能删除默认存储或有图片引用的存储实例。

#### 设置默认存储

```
POST /admin/config/default
```

**请求体**:
```json
{ "storage_key": "my-s3" }
```

### 3.8 系统设置

#### 获取系统设置

```
GET /admin/system-settings
```

**响应** (200):
```json
{
  "success": true,
  "data": {
    "runtime": {
      "site_name": "OmePic",
      "site_tagline": "上传、分享和管理图片",
      "public_base_url": "",
      "max_upload_size_mb": 20,
      "allowed_mime_types": ["image/jpeg", "image/png", "image/gif", "image/webp", "image/avif"],
      "allow_storage_selection": true,
      "maintenance_mode": false,
      "maintenance_message": "",
      "rate_limit_window_minutes": 1,
      "rate_limit_max_requests": 120,
      "upload_rate_limit_window_minutes": 10,
      "upload_rate_limit_max_requests": 20
    },
    "readonly": {
      "environment": {
        "http_addr": ":8080",
        "database_path": "data/omepic.db",
        "redis_configured": true,
        "public_base_url_source": "request_host",
        "env_public_base_url_set": false,
        "runtime_public_base_url_set": false
      },
      "security": {
        "jwt_secret": { "configured": true, "using_default": false },
        "admin_password": { "configured": true, "using_default": false },
        "uid_encryption_key": { "configured": true, "using_default": false }
      },
      "storage": {
        "default_storage_key": "local-default",
        "storage_config_count": 1,
        "allow_storage_selection": true
      },
      "service": {
        "health": "ok",
        "maintenance_mode": false
      }
    }
  }
}
```

#### 更新系统设置

```
PUT /admin/system-settings
```

**请求体**: 同 `runtime` 字段结构。

### 3.9 公告管理

#### 获取所有公告

```
GET /admin/announcements
```

#### 创建公告

```
POST /admin/announcements
```

**请求体**:
```json
{
  "title": "Notice",
  "content": "**Markdown** content",
  "status": "published",
  "priority": "normal",
  "starts_at": null,
  "ends_at": null,
  "sort_order": 0
}
```

#### 更新公告

```
PUT /admin/announcements/:id
```

#### 删除公告

```
DELETE /admin/announcements/:id
```

#### 归档公告

```
POST /admin/announcements/:id/archive
```

---

## 4. 错误码

### HTTP 状态码

| 状态码 | 含义 | 常见场景 |
|--------|------|----------|
| 200 | 成功 | 请求正常处理 |
| 201 | 创建成功 | 创建资源成功 |
| 400 | 请求参数错误 | 无效输入、缺少必填字段 |
| 401 | 未认证 | 缺少或无效的 Token/JWT |
| 403 | 禁止访问 | Token 不匹配、IP 被封禁 |
| 404 | 资源不存在 | 图片/公告/存储配置未找到 |
| 409 | 冲突 | 资源状态冲突 |
| 429 | 请求超限 | 超过速率限制 |
| 500 | 内部错误 | 未预期的服务端错误 |
| 503 | 服务不可用 | 依赖故障（Redis/SQLite/Storage） |

### 错误码列表

| 错误码 | 说明 |
|--------|------|
| `invalid_input` | 请求参数无效 |
| `missing_token` | 缺少 X-Token 头 |
| `invalid_admin_token` | JWT 无效或过期 |
| `forbidden` | Token 不匹配或无权限 |
| `ip_banned` | IP 被封禁 |
| `not_found` | 资源不存在 |
| `conflict` | 操作冲突 |
| `rate_limited` | 请求频率超限 |
| `dependency_unavailable` | 后端依赖不可用 |
| `internal_error` | 服务器内部错误 |
