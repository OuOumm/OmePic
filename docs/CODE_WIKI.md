# OmePic Code Wiki

## 1. 项目概览

OmePic 是一个单仓库图片托管系统。后端使用 Go + Gin 提供图片上传、访问、删除、管理、安全治理、SQLite 持久化、Redis 缓存与限流、多存储后端等能力；前端使用 SvelteKit 2 + Svelte 5 + Vite + TypeScript 构建静态 SPA，提供公开上传页、上传历史、API 示例、管理后台、图片治理、存储设置、系统设置、公告管理和安全/滥用分析页面。

核心能力：

- 图片上传、AVIF 转码、公开访问与删除
- 基于客户端 `X-Token` 的普通用户上传与删除所有权校验
- 基于管理员密码登录和 JWT 的管理端 API 鉴权
- SQLite 作为事实数据源
- Redis 用于 UID 图片缓存、MD5 去重缓存、API/upload fixed-window 限流
- 本地文件系统、S3 兼容对象存储、WebDAV 三类存储 provider
- 按 storage key 隔离的 MD5 去重
- 运行时配置：站点名称、标语、公开 URL、上传大小、MIME 白名单、维护模式、限流策略
- 公告系统：公开列表与后台 CRUD/archive
- 可信代理真实 IP 解析
- IP 封禁、滥用概览、IP 详情、封禁 IP 图片清理
- 前端 IndexedDB 本地上传历史
- SvelteKit 静态构建后复制到 `backend/web/`，由 Go 后端单端口托管

## 2. 仓库结构

```text
OmePic/
├── backend/
│   ├── cmd/server/main.go                 后端启动入口与依赖装配
│   ├── internal/
│   │   ├── auth/                          JWT 与 Bearer token 处理
│   │   ├── cache/                         Redis 图片 UID/MD5 缓存
│   │   ├── config/                        环境变量、存储配置、可信代理配置
│   │   ├── http/
│   │   │   ├── clientip/                  可信代理真实客户端 IP 解析
│   │   │   ├── handler/                   Gin HTTP handlers
│   │   │   ├── middleware/                管理鉴权、日志、限流中间件
│   │   │   └── router/                    路由注册与静态前端 fallback
│   │   ├── model/                         图片、公告、IP ban、abuse 模型
│   │   ├── ratelimit/                     Redis fixed-window 限流器
│   │   ├── repository/                    SQLite schema、迁移、查询、聚合
│   │   ├── response/                      统一 JSON 响应 envelope
│   │   ├── service/                       图片、后台、公告、运行时设置、安全治理业务
│   │   ├── storage/                       local / S3 / WebDAV 存储抽象
│   │   └── uid/                           加密公开 UID 编码器
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/routes/                        SvelteKit 文件路由
│   ├── src/lib/
│   │   ├── actions/                       可复用 Svelte actions
│   │   ├── components/studio/             Svelte 业务 UI 组件
│   │   ├── indexeddb/                     IndexedDB 上传历史
│   │   ├── stores/                        Svelte 5 runes 状态
│   │   ├── types/                         前端共享类型
│   │   ├── api.ts                         前端 API client
│   │   ├── i18n.ts                        中英文文案
│   │   ├── preferences.ts                 客户端 token 与 legacy preference helpers
│   │   └── utils.ts                       工具函数
│   ├── src/app.css                        全局样式、主题变量、studio classes
│   ├── scripts/copy-static-to-backend.mjs 静态产物复制脚本
│   ├── svelte.config.js
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── package.json
├── docs/
│   └── CODE_WIKI.md                       本文档
├── .env.example                           环境变量示例
├── README.md                              项目说明
└── AGENTS.md                              Trellis 协作说明
```

## 3. 技术栈与依赖

### 3.1 后端

后端 Go module：`omepic/backend`，当前 `go.mod` 使用 Go `1.25.0`。

主要依赖：

| 依赖 | 用途 |
|---|---|
| `github.com/gin-gonic/gin` | HTTP server、router、handler |
| `github.com/gin-contrib/cors` | CORS 中间件 |
| `github.com/golang-jwt/jwt/v5` | 管理端 JWT |
| `modernc.org/sqlite` | SQLite 驱动 |
| `github.com/redis/go-redis/v9` | Redis 缓存与限流 |
| `github.com/minio/minio-go/v7` | S3 兼容对象存储 |
| `github.com/studio-b12/gowebdav` | WebDAV 存储 |
| `github.com/gen2brain/avif` | AVIF 编码 |
| `golang.org/x/image` | 图片解码辅助 |
| `github.com/google/uuid` | UUID 相关能力 |

### 3.2 前端

前端是 SvelteKit 静态 SPA。

主要依赖：

| 依赖 | 用途 |
|---|---|
| `@sveltejs/kit` | SvelteKit 应用框架 |
| `svelte` | Svelte 5 UI 框架 |
| `vite` | 开发服务器与构建工具 |
| `@sveltejs/adapter-static` | 输出静态站点到 `out/` |
| `typescript` | 类型系统 |
| `svelte-check` | Svelte + TypeScript 检查 |
| `tailwindcss` | 样式系统 |
| `lucide-svelte` | 图标 |
| `marked` | Markdown 解析 |
| `dompurify` | HTML sanitize |
| `clsx`、`tailwind-merge` | className 组合 |

## 4. 整体架构

### 4.1 分层视图

```text
Browser
  ↓
SvelteKit static SPA
  ↓ HTTP JSON / multipart
Go Gin router
  ↓
HTTP handlers
  ↓
Services
  ↓
Repository / Redis cache / Redis limiter / Storage manager / Runtime settings / UID codec / Client IP resolver
  ↓
SQLite / Redis / Local FS / S3 / WebDAV
```

### 4.2 后端分层职责

- `cmd/server`：进程入口，集中装配配置、数据库、缓存、限流、存储、service、handler、router。
- `config`：读取环境变量，提供 app config、S3/WebDAV config、默认 storage config、可信代理配置。
- `http/router`：注册公共 API、管理 API、中间件和静态前端 fallback。
- `http/handler`：HTTP 参数解析、文件读取、调用 service、错误映射。
- `http/middleware`：Admin JWT 鉴权、请求日志、Redis 限流。
- `http/clientip`：可信代理下解析真实客户端 IP。
- `service`：业务流程，包括上传、去重、删除、解析、配置、公告、IP ban、abuse。
- `repository`：SQLite schema、迁移、CRUD、搜索、聚合查询。
- `cache`：Redis UID 缓存与 MD5 去重缓存。
- `ratelimit`：Redis Lua fixed-window 限流。
- `storage`：统一 local/S3/WebDAV 存储 provider。
- `uid`：公开 UID 生成、解码、校验。
- `model`：领域模型。
- `response`：统一 JSON 响应 envelope。

### 4.3 前端分层职责

- `src/routes`：SvelteKit 文件路由、页面级数据加载和局部状态。
- `src/routes/+layout.svelte`：导入全局 CSS，用 `AppShell` 包裹所有页面。
- `src/routes/admin/dashboard/+layout.svelte`：管理端布局、侧边栏、二级导航、登出。
- `src/lib/components/studio`：项目 UI 组件，如上传区域、图片表格、弹窗、抽屉、存储管理、公告管理。
- `src/lib/api.ts`：所有 public/admin API helper。
- `src/lib/stores`：Svelte 5 `$state` stores，保存 UI 偏好、admin token、runtime settings、toast 队列。
- `src/lib/indexeddb`：本地上传历史持久化。
- `src/lib/types`：前端 API/domain types。
- `src/app.css`：主题 tokens、全局样式、studio 风格类。

### 4.4 后端启动流程

```text
main.go
→ config.Load()
→ repository.New(cfg.DatabasePath)
→ repo.Migrate(ctx)
→ repo.InitializeStorageCatalog(ctx, config.DefaultStorageConfig())
→ storage.NewManager(storageCatalog.StorageConfigs)
→ uid.NewCodec(cfg.UIDPrefix, cfg.UIDEncryptionKey)
→ cache.NewClient(cfg.RedisURL)
→ cache.NewWithClient(redisClient)
→ ratelimit.NewRedisLimiter(redisClient)
→ repo.Ping(ctx) / imageCache.Ping(ctx)
→ RuntimeSettingsManager.Load(ctx, repo)
→ NewImageService(...)
→ NewAdminService(...)
→ NewAnnouncementService(...)
→ clientip.NewResolver(nil, "")
→ imageService.Preheat(ctx)
→ router.New(router.Dependencies{...})
→ engine.Run(cfg.HTTPAddr)
```

### 4.5 核心业务数据流

#### 上传

```text
POST /v1/image
→ uploadLimiter
→ ImageHandler.Upload
→ clientip.Resolver.Resolve(request)
→ ImageService.Upload
→ ensureIPAllowed(ip)
→ 校验 X-Token
→ 校验维护模式、文件大小、扩展名、MIME 白名单
→ 解析 storage_key 或默认存储
→ 计算原始上传字节 MD5
→ scoped MD5 查重：Redis + SQLite
→ 重复：复用已有物理文件，写入新 UID 记录
→ 不重复：转换 AVIF，保存到 provider，写入 SQLite
→ 写 Redis uid 与 md5 cache
→ 返回 UploadOutput
```

#### 图片访问

```text
GET /i/:uid.avif
→ ImageHandler.Serve
→ ImageService.Resolve
→ 校验 public UID 与 .avif 后缀
→ Redis uid cache 命中则返回
→ miss 后查 SQLite 并回填 Redis
→ storage.Manager.ForKey(record.StorageKey)
→ Provider.Open(record.FilePath)
→ 返回 image/avif 文件流
```

#### 普通用户删除

```text
DELETE /i/:uid.avif
→ apiLimiter
→ ImageHandler.Delete
→ clientip.Resolver.Resolve(request)
→ ImageService.Delete(isAdmin=false)
→ ensureIPAllowed(ip)
→ 校验 UID 与 X-Token 所有权
→ 删除 SQLite UID 记录
→ 删除 Redis uid cache
→ 修复或删除 scoped MD5 cache
→ 保留物理文件
```

#### 管理端删除

```text
DELETE /admin/images
→ apiLimiter
→ AdminAuth
→ AdminHandler.DeleteImages
→ AdminService.DeleteImages
→ ImageService.Delete(isAdmin=true)
```

#### IP 封禁与滥用治理

```text
/admin/ip-bans 或 /admin/abuse/*
→ apiLimiter
→ AdminAuth
→ AdminHandler
→ AdminService
→ Repository ip_bans + images aggregate queries
```

## 5. 后端模块详解

### 5.1 `cmd/server`

关键文件：`backend/cmd/server/main.go`

主要职责：

- 创建 JSON `slog` logger。
- 加载 `AppConfig`。
- 创建 `data` 目录。
- 初始化 SQLite repository，并执行 migration。
- 初始化 storage catalog 和 storage manager。
- 初始化 UID codec。
- 初始化 Redis client、图片 cache、rate limiter。
- 加载 runtime settings。
- 创建 `ImageService`、`AdminService`、`AnnouncementService`。
- 创建 `clientip.Resolver`。
- 启动前预热图片缓存。
- 注入 router dependencies 并启动 Gin server。

关键函数：

- `main()`：唯一 Go command 入口。

### 5.2 `internal/config`

关键文件：`backend/internal/config/config.go`

关键类型：

- `AppConfig`：仅 HTTP、SQLite、Redis、UID、JWT 等启动必需环境配置。
- `RuntimeStorageConfig`：运行时 storage instance 配置。
- `RuntimeStorageCatalog`：默认 storage key 与 storage configs。
- `RuntimeStorageUpdate`：storage config patch 输入。

重要配置字段：

- `HTTPAddr`
- `DatabasePath`
- `RedisURL`
- `UIDPrefix`
- `UIDEncryptionKey`
- `JWTSecret`

公开访问基准 URL、存储配置、上传策略、维护模式、限流和管理员密码均保存在 SQLite，不再通过 `AppConfig` 从环境变量读取。

关键函数：

- `Load()`：从环境变量读取配置。
- `DefaultStorageConfig()`：生成初始默认 storage instance。
- `BootstrapStorageKey()`：生成初始 storage key。
- `BootstrapStorageName()`：生成初始 storage 展示名。

### 5.3 `internal/http/clientip`

关键文件：`backend/internal/http/clientip/resolver.go`

职责：

- 解析可信客户端 IP。
- 解析可信客户端 IP。
- 当前启动配置不读取可信代理环境变量，默认以无可信代理方式创建 resolver。
- 因没有可信代理，默认不信任 `X-Forwarded-For` / `X-Real-IP`，直接使用 remote IP。

关键类型与函数：

- `Resolver`
- `NewResolver(trustedProxyCIDRs []string, realIPHeader string)`
- `Resolve(req *http.Request) string`

使用点：

- `ImageHandler.Upload`
- `ImageHandler.Delete`
- `middleware.RateLimit`
- `images.ip_address`
- IP ban enforcement
- abuse analytics

### 5.4 `internal/http/router`

关键文件：

- `backend/internal/http/router/router.go`
- `backend/internal/http/router/frontend.go`

职责：

- 创建 Gin engine。
- 注册 recovery、CORS、请求日志。
- 构造 API/upload 限流 middleware。
- 注册公共 API。
- 注册 `/admin` group，并添加 API 限流和 JWT 鉴权。
- 注册静态前端 fallback。

公共路由：

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/health` | SQLite 与 Redis 健康检查 |
| `GET` | `/v1/runtime-settings` | 公开运行时设置 |
| `GET` | `/v1/announcements` | 公开公告列表 |
| `POST` | `/v1/image` | 上传图片 |
| `GET` | `/i/:uid` | 获取图片 |
| `DELETE` | `/i/:uid` | 普通用户删除图片 |
| `POST` | `/admin/login` | 管理员登录 |

管理路由：

| 方法 | 路径 | 说明 |
|---|---|---|
| `PUT` | `/admin/password` | 修改管理员密码 |
| `GET` | `/admin/status` | 管理状态统计 |
| `GET` | `/admin/images` | 图片列表 |
| `DELETE` | `/admin/images` | 批量删除图片 |
| `GET` | `/admin/ip-bans` | IP 封禁列表 |
| `POST` | `/admin/ip-bans` | 创建 IP 封禁 |
| `DELETE` | `/admin/ip-bans/:id` | 删除 IP 封禁 |
| `DELETE` | `/admin/ip-bans/:id/images` | 删除该封禁 IP 相关图片 |
| `GET` | `/admin/abuse/overview` | 滥用概览 |
| `GET` | `/admin/abuse/ip` | IP 详情 |
| `GET` | `/admin/config` | 存储配置 |
| `POST` | `/admin/config` | 兼容式配置更新 |
| `POST` | `/admin/config/storage-instances` | 新建存储实例 |
| `PUT` | `/admin/config/storage-instances/:storageKey` | 更新存储实例 |
| `DELETE` | `/admin/config/storage-instances/:storageKey` | 删除存储实例 |
| `POST` | `/admin/config/default` | 设置默认存储 |
| `GET` | `/admin/system-settings` | 获取系统设置 |
| `PUT` | `/admin/system-settings` | 更新系统设置 |
| `GET` | `/admin/announcements` | 管理端公告列表 |
| `POST` | `/admin/announcements` | 创建公告 |
| `PUT` | `/admin/announcements/:id` | 更新公告 |
| `DELETE` | `/admin/announcements/:id` | 删除公告 |
| `POST` | `/admin/announcements/:id/archive` | 归档公告 |

静态前端 fallback：

- 若 `backend/web/index.html` 不存在，后端以 API-only 模式运行。
- 若存在，则 `NoRoute` 尝试服务静态文件。
- 对浏览器页面路由 fallback 到 `index.html`。
- 对明确 API 前缀保持 JSON 404。

### 5.5 `internal/http/handler`

关键文件：

- `backend/internal/http/handler/image_handler.go`
- `backend/internal/http/handler/admin_handler.go`
- `backend/internal/http/handler/announcement_handler.go`
- `backend/internal/http/handler/health_handler.go`

#### `ImageHandler`

职责：

- `Upload`：读取 multipart `file`，读取 `X-Token`，读取可选 `storage_key`，解析 client IP，调用 `ImageService.Upload`。
- `RuntimeSettings`：返回公开 runtime settings view。
- `Delete`：解析 UID，读取 `X-Token` 和 client IP，调用 `ImageService.Delete`。
- `Serve`：解析图片记录，按 storage key 打开对象，设置响应头并流式返回。

重要错误映射：

- `ErrMissingToken` -> token 缺失。
- `ErrForbidden` -> 权限不足。
- `ErrIPBanned` -> HTTP 403，错误码 `ip_banned`。
- `ErrInvalidInput` -> 请求错误。
- `ErrNotFound` -> not found。

#### `AdminHandler`

职责：

- `Login`
- `Status`
- `Images`
- `DeleteImages`
- `CreateIPBan`
- `IPBans`
- `DeleteIPBan`
- `DeleteIPBanImages`
- `AbuseOverview`
- `AbuseIPDetail`
- `GetConfig`
- `UpdateConfig`
- storage instance CRUD/default
- system settings get/update

#### `AnnouncementHandler`

职责：

- `PublicList`
- `AdminList`
- `Create`
- `Update`
- `Delete`
- `Archive`

#### `HealthHandler`

职责：

- 同时检查 SQLite repository 和 Redis cache。
- 成功返回 `success: true` 与 `status: ok`。

### 5.6 `internal/http/middleware`

关键文件：

- `auth_middleware.go`
- `logging_middleware.go`
- `rate_limit_middleware.go`

职责：

- `AdminAuth(jwtSecret)`：解析 `Authorization: Bearer <jwt>` 并校验 JWT。
- `RequestLogger(logger)`：记录结构化请求日志。
- `RateLimit(limiter, logger, policy)`：根据 resolved client IP + scope 进行 Redis fixed-window 限流。

限流特点：

- key 格式：`ratelimit:<scope>:ip:<sha256(client_ip)>`。
- scope：`api`、`upload`。
- Redis 出错时 fail-open，记录日志但放行请求。
- 限流命中时返回 `Retry-After`、`X-RateLimit-Limit`、`X-RateLimit-Remaining`。

### 5.7 `internal/service`

#### `ImageService`

关键文件：

- `backend/internal/service/image_service.go`
- `backend/internal/service/image_transform.go`
- `backend/internal/service/errors.go`
- `backend/internal/service/ip_utils.go`

关键类型：

- `ImageService`
- `UploadInput`
- `UploadOutput`
- `ImageResolverOutput`

关键函数：

- `Upload(ctx, input)`：上传主流程。
- `Delete(ctx, uid, token, isAdmin, ipAddress)`：普通用户或管理员删除图片记录。
- `Resolve(ctx, uid)`：解析公开 UID 图片记录。
- `Preheat(ctx)`：启动时预热 Redis cache。
- `EffectivePublicBaseURL(requestBase)`：按 SQLite runtime setting 或请求 Host 计算公开 URL base。
- `ensureIPAllowed(ctx, ipAddress)`：检查 active IP ban。
- `convertToAVIFWithSettings(data, settings)`：按 runtime AVIF 质量/速度参数转换图片。
- `ipHash(ip)`：SHA-256 IP hash。
- `maskIPAddress(ip)`：IPv4/IPv6 脱敏展示。

上传规则：

- 必须有 `X-Token`。
- 先检查 IP ban。
- 维护模式阻止上传。
- `MaxUploadSizeMB <= 0` 表示不限制大小。
- MIME 默认允许 `image/jpeg`、`image/png`、`image/gif`、`image/webp`、`image/avif`。
- 文件扩展名支持常见 raster 图片。
- 查重按 `storage_key + md5_hash` scope。
- 新文件保存为 AVIF。
- 删除当前只删除逻辑记录与缓存，不主动删除物理文件。

#### `AdminService`

关键文件：`backend/internal/service/admin_service.go`

关键类型：

- `AdminService`
- `AdminImageItem`
- `AdminImageList`
- `AdminIPBanCreateInput`
- `AdminIPBanCreateResult`
- `AdminIPBanDeleteImagesResult`
- storage create/update input
- system settings update input

关键函数：

- `Login(password)`：校验 SQLite 中的 bcrypt 密码哈希，生成 24 小时 JWT。
- `ChangePassword(ctx, oldPassword, newPassword)`：校验旧密码和新密码强度（至少 8 位，包含大写、小写和符号）后写入新的 bcrypt 哈希。
- `Status(ctx)`：后台统计。
- `Images(ctx, page, pageSize, search)`：分页搜索图片。
- `DeleteImages(ctx, uids)`：管理员批量删除。
- `CreateIPBan(ctx, input)`：按 UID 或 IP 创建封禁。
- `IPBans(ctx)`：列出封禁。
- `DeleteIPBan(ctx, id)`：解除封禁。
- `DeleteImagesByIPBan(ctx, id)`：删除某封禁 IP 关联图片。
- `AbuseOverview(ctx, from, to)`：滥用概览。
- `AbuseIPDetail(ctx, ip)`：IP 详情。
- `GetConfig(ctx)` / storage CRUD / `SetDefaultStorageConfig(ctx, key)`。
- `GetSystemSettings(ctx)` / `UpdateSystemSettings(ctx, input)`。

#### `RuntimeSettingsManager`

关键文件：`backend/internal/service/runtime_settings.go`

关键类型：

- `RuntimeSettings`
- `PublicRuntimeSettingsView`
- `AdminSystemSettingsView`
- `RuntimeSettingsUpdateInput`
- `RuntimeSettingsManager`

配置字段：

- `site_name`
- `site_tagline`
- `public_base_url`
- `max_upload_size_mb`
- `allowed_mime_types`
- `avif_quality`
- `avif_speed`
- `allow_storage_selection`
- `maintenance_mode`
- `maintenance_message`
- `rate_limit_window_minutes`
- `rate_limit_max_requests`
- `upload_rate_limit_window_minutes`
- `upload_rate_limit_max_requests`

默认值：

- 站点名：`OmePic`
- 标语：`上传、分享和管理图片`
- 维护提示：`系统维护中，请稍后再试`
- API 限流：1 分钟 120 次
- 上传限流：10 分钟 20 次
- AVIF 质量：60（范围 0..100，100 表示无损）
- AVIF 速度：8（范围 0..10，数值越低通常越慢但压缩/质量取舍更好）

关键函数：

- `Load(ctx, repo)`：先向 SQLite config 补齐缺失的默认 runtime settings，再加载设置。
- `Current()`：读取当前设置副本。
- `Reconfigure(settings)`：热更新内存设置。
- `EffectivePublicBaseURL(requestBase)`：优先 SQLite runtime setting，其次 request host。
- `PublicBaseURLSource()`：返回 `runtime` 或 `request_host`。

#### `AnnouncementService`

关键文件：`backend/internal/service/announcement_service.go`

关键函数：

- `PublicAnnouncements(ctx)`
- `AdminAnnouncements(ctx)`
- `CreateAnnouncement(ctx, input)`
- `UpdateAnnouncement(ctx, id, input)`
- `DeleteAnnouncement(ctx, id)`
- `ArchiveAnnouncement(ctx, id)`

#### Abuse 与 IP 工具

关键文件：

- `backend/internal/service/abuse.go`
- `backend/internal/service/ip_utils.go`

规则：

- Abuse overview 默认近 24 小时。
- 最大查询范围 90 天。
- `from` 必须早于 `to`。
- IP hash 使用 SHA-256。
- IPv4 masked 格式类似 `a.b.c.*`。
- IPv6 masked 输出前两段加 `:*`。

### 5.8 `internal/repository`

关键文件：`backend/internal/repository/repository.go`

职责：

- 创建 SQLite 连接并配置 PRAGMA。
- 执行 schema migration。
- 图片 CRUD、搜索、聚合统计。
- Storage config 初始化与持久化。
- Runtime config 读写。
- Announcement CRUD 和公开过滤。
- IP ban CRUD。
- Abuse 聚合查询。

连接特性：

- 使用 `modernc.org/sqlite`。
- 设置 WAL、foreign keys、busy timeout、mmap 等 PRAGMA。
- 控制连接数，适配 SQLite。

核心表：

```text
images
- id, uid, token, storage_key, storage_backend, file_path
- mime_type, size, md5_hash, ip_address, created_at

config
- key, value

storage_configs
- id, storage_key, name, backend, is_default
- local_storage_path
- s3_endpoint, s3_region, s3_bucket, s3_access_key, s3_secret_key, s3_use_ssl, s3_force_path_style
- webdav_url, webdav_user, webdav_pass
- created_at, updated_at

announcements
- id, title, content, status, priority
- starts_at, ends_at, sort_order
- created_at, updated_at

ip_bans
- id, ip_hash, ip_address, ip_address_masked
- reason, expires_at, created_at, updated_at
```

关键索引：

- UID 索引
- MD5 索引
- storage + MD5 索引
- storage key 索引
- IP address 索引
- created_at + IP 索引
- created_at + token 索引
- announcement status/time/sort 索引
- IP ban hash 与 expires_at 索引

关键函数：

- `New(databasePath)`
- `Migrate(ctx)`
- `InitializeStorageCatalog(ctx, envDefault)`
- `InsertImage(ctx, record)`
- `FindByUID(ctx, uid)`
- `FindByMD5AndStorageKey(ctx, storageKey, md5)`
- `DeleteByUID(ctx, uid)`
- `SearchImages(ctx, page, pageSize, search)`
- `AggregateStatus(ctx)`
- `ImageSummaryByIP(ctx, ip)`
- `ListImagesByIP(ctx, ip)`
- `CreateIPBan(ctx, ban)`
- `ListIPBans(ctx)`
- `FindActiveIPBanByHash(ctx, hash)`
- `DeleteIPBan(ctx, id)`
- `CountActiveIPBans(ctx)`
- `AbuseOverviewTotals(ctx, from, to)`
- `TopAbuseIPs(ctx, from, to, limit)`
- `TopAbuseTokens(ctx, from, to, limit)`
- `IPDetail(ctx, ip)`
- announcement CRUD functions

### 5.9 `internal/cache`

关键文件：`backend/internal/cache/redis_cache.go`

关键接口：

- `ImageCache`
  - `GetImage(ctx, uid)`
  - `SetImage(ctx, record)`
  - `DeleteImage(ctx, uid)`
  - `GetMD5(ctx, md5Hash)`
  - `SetMD5(ctx, md5Hash, uid)`
  - `SetMD5IfAbsent(ctx, md5Hash, uid)`
  - `DeleteMD5(ctx, md5Hash)`
  - `Ping(ctx)`

Key 约定：

- `uid:<uid>`：图片元数据 JSON。
- `md5:<scoped-hash>`：MD5 查重映射。service 层 scoped hash 含 storage key 语义。

### 5.10 `internal/ratelimit`

关键文件：`backend/internal/ratelimit/redis_limiter.go`

关键类型：

- `Limiter`
- `RedisLimiter`
- `Result`

实现：

- Redis Lua script。
- `INCR` + 首次请求 `PEXPIRE`。
- 返回 allowed、remaining、retry after。

### 5.11 `internal/storage`

关键文件：`backend/internal/storage/storage.go`

关键接口：

```go
Provider interface {
  Name() string
  Save(ctx, objectKey, data, contentType) (string, error)
  Open(ctx, objectKey) (OpenResult, error)
  Delete(ctx, objectKey) error
}
```

关键类型：

- `OpenResult`
- `ResolvedProvider`
- `Manager`
- `localProvider`
- `s3Provider`
- `webdavProvider`

关键函数：

- `NewManager(settings)`
- `Current()`
- `CurrentKey()`
- `CurrentBackend()`
- `ForKey(storageKey)`
- `Reconfigure(settings)`
- `ValidateConfig(settings)`
- `BuildObjectKey(uid, extension)` -> `YYYY/MM/<uid>.<ext>`

### 5.12 `internal/auth`

关键文件：`backend/internal/auth/jwt.go`

关键类型与函数：

- `Claims`
- `GenerateJWT(secret, ttl)`
- `ParseJWT(secret, tokenString)`
- `ValidateJWT(secret, tokenString)`
- `ParseBearer(header)`

### 5.13 `internal/uid`

关键文件：`backend/internal/uid/codec.go`

职责：

- Snowflake SID + prefix + secret + XOR + base64/base62 生成短公开 UID。
- 解码并校验 prefix。
- 最大 public UID 长度约束为 30。

关键类型与函数：

- `Codec`
- `Decoded`
- `SnowflakeGenerator`
- `NewCodec(prefix, secret)`
- `Generate()`
- `Decode(uid)`
- `Validate(uid)`

### 5.14 `internal/model` 与 `internal/response`

核心模型：

- `ImageRecord`
- `CachedImage`
- `AdminStatus`
- `Announcement`
- `IPBan`
- `IPImageSummary`
- `AbuseOverview`
- `AbuseIPRankItem`
- `AbuseTokenRankItem`
- `AbuseIPDetail`

统一响应：

```json
{
  "success": true,
  "data": {}
}
```

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "human readable message"
  }
}
```

## 6. 前端模块详解

### 6.1 SvelteKit 配置

关键文件：

- `frontend/svelte.config.js`
- `frontend/vite.config.ts`
- `frontend/tsconfig.json`
- `frontend/tailwind.config.ts`

配置要点：

- 使用 `@sveltejs/adapter-static`。
- pages/assets 输出目录为 `out`。
- SPA fallback 为 `index.html`。
- `prerender.entries = ['*']`。
- alias：`@/*` 指向 `frontend/src/lib/*`。
- Vite 启用 SvelteKit plugin。
- Tailwind 扫描 `./src/**/*.{html,js,svelte,ts}`。

### 6.2 SvelteKit 路由

| 路由 | 文件 | 说明 |
|---|---|---|
| `/` | `frontend/src/routes/+page.svelte` | 主上传页：runtime settings、公告、上传、URL 上传、最近上传、删除 |
| `/history` | `frontend/src/routes/history/+page.svelte` | IndexedDB 本地上传历史 |
| `/api` | `frontend/src/routes/api/+page.svelte` | API 示例与客户端 token 展示 |
| `/admin/dashboard` | `frontend/src/routes/admin/dashboard/+page.svelte` | 管理后台状态概览 |
| `/admin/dashboard/images` | `frontend/src/routes/admin/dashboard/images/+page.svelte` | 图片管理：搜索、分页、预览、删除、封禁 IP |
| `/admin/dashboard/security` | `frontend/src/routes/admin/dashboard/security/+page.svelte` | abuse 概览、IP bans、限流设置 tabs |
| `/admin/dashboard/settings` | `frontend/src/routes/admin/dashboard/settings/+page.svelte` | runtime/storage/announcements 设置 tabs |
| root layout | `frontend/src/routes/+layout.svelte` | 引入全局 CSS，用 `AppShell` 包裹页面 |
| admin layout | `frontend/src/routes/admin/dashboard/+layout.svelte` | 管理后台侧边栏、二级导航、登录态、登出 |

### 6.3 Studio 组件

关键目录：`frontend/src/lib/components/studio/`

主要组件：

| 组件 | 职责 |
|---|---|
| `AppShell.svelte` | 全站导航、语言切换、主题切换、移动菜单、Toast 容器 |
| `CanvasDropzone.svelte` | 上传拖拽/选择区域 |
| `ImageDataTable.svelte` | 图片/历史表格展示与操作入口 |
| `ImagePreviewDialog.svelte` | 公开上传和历史图片预览，支持复制、下载、删除、前后导航 |
| `ImageDetailDrawer.svelte` | 管理端图片详情抽屉，支持删除、IP 详情、IP 封禁、前后导航 |
| `BanIPDialog.svelte` | 创建 IP ban 表单 |
| `IPDetailPanel.svelte` | IP 级 abuse 详情 |
| `StorageInspector.svelte` | 存储配置概览 |
| `StorageInstanceManager.svelte` | 存储实例 CRUD 与默认存储设置 |
| `AnnouncementManager.svelte` | 公告 CRUD、归档、预览 |
| `AnnouncementDialog.svelte` | 公开公告弹窗 |
| `ConfirmDialog.svelte` | 通用确认弹窗 |
| `ToastViewport.svelte` | Toast 显示容器 |
| `MetricStrip.svelte` | 指标展示条 |
| `PageTitle.svelte` | 页面标题区 |

组件特点：

- 当前代码同时存在 Svelte 5 rune 写法和部分传统 `export let` 写法。
- 多个弹窗组件使用 `accessibleDialog` action 管理焦点与 Escape 关闭。
- 可复用 UI 优先放入 studio 组件，而不是散落在 route 页面中。

### 6.4 Actions

关键目录：`frontend/src/lib/actions/`

- `accessible-dialog.ts`：焦点捕获、Escape 关闭、销毁时恢复焦点。
- `click-outside.ts`：监听组件外点击。

### 6.5 Stores / 客户端状态

关键目录：`frontend/src/lib/stores/`

#### `preferences.svelte.ts`

使用 Svelte 5 `$state` 保存：

- `language`
- `theme`
- `selectedStorageKey`
- `adminToken`
- `runtimeSettings`

本地持久化 key：

- `omepic-ui-preferences`
- `omepic-upload-preferences`
- `omepic-admin-token`

关键函数：

- `setLanguage(language)`
- `setTheme(theme)`
- `setSelectedStorageKey(key)`
- `setRuntimeSettings(settings)`
- `setAdminToken(token)`
- `clearAdminToken()`
- `resolvedTheme()`

#### `toast.svelte.ts`

保存 toast 队列：

- `toasts.items`
- `toast.success(message)`
- `toast.error(message)`
- `toast.info(message)`

Toast 默认约 3200ms 自动移除。

### 6.6 API Client

关键文件：`frontend/src/lib/api.ts`

核心类与函数：

- `ApiError`：封装后端错误 message、code、status、retryAfter。
- `apiFetch<T>(path, options)`：统一构造 URL、query params、`cache: no-store`、解析 `ApiResponse<T>`。
- `uploadImageWithProgress(file, token, onProgress, storageKey?)`：使用 XHR 上传 multipart，支持进度与 Retry-After。
- `adminHeaders(token)`：管理端 `Authorization: Bearer <token>`。

公共 API helper：

- `getRuntimeSettings(signal?)`
- `getAnnouncements(signal?)`
- `deleteImageByUid(uid, token)`

管理 API helper：

- `adminLogin(password)`
- `adminChangePassword(token, oldPassword, newPassword)`
- `adminGetStatus(token)`
- `adminGetImages(token, page, pageSize, search?)`
- `adminDeleteImages(token, uids)`
- `adminCreateIPBan(token, input)`
- `adminGetIPBans(token)`
- `adminDeleteIPBan(token, id)`
- `adminDeleteIPBanImages(token, id)`
- `adminGetAbuseOverview(token, params?)`
- `adminGetAbuseIPDetail(token, ip)`
- storage config CRUD/default helpers
- system settings get/update helpers
- announcement CRUD/archive helpers

### 6.7 Utils 与 API Base URL

关键文件：`frontend/src/lib/utils.ts`

关键函数：

- `cn(...)`
- `formatBytes(bytes)`
- `formatMegabytes(mb)`
- `formatDate(input)`
- `getApiBaseUrl()`
- `getAbsoluteUrl(url)`
- `getImageUrl(uid)`

API base 规则：

- 非浏览器环境 fallback 到 `http://localhost:8080`。
- 浏览器端优先 `import.meta.env.VITE_API_BASE_URL`。
- 浏览器端未设置时返回空字符串，使用同源 API。

### 6.8 IndexedDB

关键文件：`frontend/src/lib/indexeddb/upload-history.ts`

配置：

- DB：`omepic`
- Store：`uploads`
- keyPath：`uid`

关键函数：

- `saveUploadToHistory(record)`
- `getRecentUploads(limit)`
- `getAllUploads()`
- `deleteUploadFromHistory(uid)`
- `clearUploadHistory()`
- `getUploadCount()`

### 6.9 Preferences helper

关键文件：`frontend/src/lib/preferences.ts`

职责：

- 获取或生成客户端上传 token。
- token 存储 key：`omepic-client-token`。
- 生成 32 位随机 token。
- 提供 legacy UI/upload/admin preference 读取写入辅助。

### 6.10 I18n

关键文件：`frontend/src/lib/i18n.ts`

职责：

- 内置英文/中文翻译字典。
- `t(language, key, params?)` 解析文案。
- 新增可见 UI 文案时应同步中英文。

### 6.11 类型模型

关键文件：`frontend/src/lib/types/index.ts`

核心类型：

- `ApiResponse<T>`
- `UploadResult`
- `StorageOption`
- `AdminStatus`
- `AdminImage`
- `AdminImagesResponse`
- `AdminIPBan`
- `AdminIPBanCreateResult`
- `AdminIPBanDeleteImagesResult`
- `AdminAbuseOverview`
- `AdminAbuseIPRankItem`
- `AdminAbuseTokenRankItem`
- `AdminAbuseIPDetail`
- `StorageInstance`
- `AdminConfig`
- `RuntimeSettings`
- `PublicRuntimeSettings`
- `AdminSystemSettings`
- `SecretStatus`
- `Announcement`
- `AnnouncementListResponse`
- `AnnouncementInput`
- `UploadHistoryRecord`
- `Language`
- `Theme`
- `ViewMode`

### 6.12 样式系统

关键文件：

- `frontend/src/app.css`
- `frontend/tailwind.config.ts`

特点：

- Tailwind layers：`base`、`components`、`utilities`。
- CSS variables：`--paper`、`--ink`、`--marker-yellow`、`--marker-pink` 等。
- 暗色主题通过 `html.dark` 覆盖变量。
- 自定义类：`studio-panel`、`studio-button`、`studio-input`、`studio-table-row`、`marker-highlight`、`blueprint-grid`。

## 7. 前后端接口契约

### 7.1 统一响应 envelope

成功：

```json
{
  "success": true,
  "data": {}
}
```

失败：

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "error message"
  }
}
```

### 7.2 Public API

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/health` | 健康检查 |
| `GET` | `/v1/runtime-settings` | 公开运行时设置 |
| `GET` | `/v1/announcements` | 公开公告列表 |
| `POST` | `/v1/image` | 上传图片，multipart `file`，header `X-Token`，可选 `storage_key` |
| `GET` | `/i/:uid.avif` | 获取 AVIF 图片 |
| `DELETE` | `/i/:uid.avif` | 普通用户删除图片，header `X-Token` |

### 7.3 Admin API

除 `/admin/login` 外，管理端 API 均需要：

```text
Authorization: Bearer <jwt>
```

主要接口：

- `POST /admin/login`
- `GET /admin/status`
- `GET /admin/images`
- `DELETE /admin/images`
- `GET /admin/ip-bans`
- `POST /admin/ip-bans`
- `DELETE /admin/ip-bans/:id`
- `DELETE /admin/ip-bans/:id/images`
- `GET /admin/abuse/overview`
- `GET /admin/abuse/ip`
- `GET /admin/config`
- `POST /admin/config`
- `POST /admin/config/storage-instances`
- `PUT /admin/config/storage-instances/:storageKey`
- `DELETE /admin/config/storage-instances/:storageKey`
- `POST /admin/config/default`
- `GET /admin/system-settings`
- `PUT /admin/system-settings`
- `GET /admin/announcements`
- `POST /admin/announcements`
- `PUT /admin/announcements/:id`
- `DELETE /admin/announcements/:id`
- `POST /admin/announcements/:id/archive`

### 7.4 上传响应与前端映射

前端 `UploadResult` 保存字段：

- `uid`
- `url`
- `mime_type`
- `size`
- `created_at`
- `is_duplicate`
- `storage_key`
- `storage_backend`
- `markdown`
- `bbcode`

上传历史 `UploadHistoryRecord` 额外保存：

- `client_token`
- `original_filename`
- `saved_at`

### 7.5 IP Ban / Abuse 契约

IP ban 创建可以基于：

- `uid`
- 或 `ip_address`

Active ban 条件：

```text
expires_at IS NULL OR expires_at = '' OR expires_at > now
```

Abuse overview：

- 默认范围：近 24 小时。
- 最大范围：90 天。
- 返回 upload_count、upload_size、active_ip_ban_count、top_ips、top_tokens。

## 8. 依赖关系

### 8.1 后端依赖图

```text
main
├── config.Load
├── repository.New / Migrate / InitializeStorageCatalog
├── storage.NewManager
├── uid.NewCodec
├── cache.NewClient / NewWithClient
├── ratelimit.NewRedisLimiter
├── RuntimeSettingsManager
├── ImageService
├── AdminService
├── AnnouncementService
├── clientip.Resolver
└── router.New
    ├── middleware.RequestLogger
    ├── middleware.RateLimit
    ├── middleware.AdminAuth
    ├── ImageHandler → ImageService → Repository / Cache / Storage / Settings / UID / ClientIP
    ├── AdminHandler → AdminService → Repository / Storage / ImageService / Settings / Auth
    ├── AnnouncementHandler → AnnouncementService → Repository
    └── HealthHandler → Repository / Cache
```

### 8.2 前端依赖图

```text
src/routes
├── +layout.svelte → AppShell → preferences / ToastViewport
├── +page.svelte → CanvasDropzone / AnnouncementDialog / ImagePreviewDialog / API / IndexedDB
├── history/+page.svelte → IndexedDB / deleteImageByUid / ImagePreviewDialog
├── api/+page.svelte → preferences / token helpers
└── admin/dashboard
    ├── +layout.svelte → preferences.adminToken / adminGetStatus / sidebar tabs
    ├── +page.svelte → adminGetStatus / adminGetSystemSettings
    ├── images/+page.svelte → ImageDataTable / ImageDetailDrawer / BanIPDialog / admin image APIs
    ├── security/+page.svelte → MetricStrip / BanIPDialog / ConfirmDialog / admin abuse + IP ban + rate limit APIs
    └── settings/+page.svelte → StorageInspector / StorageInstanceManager / AnnouncementManager / settings APIs
```

### 8.3 数据存储关系

```text
SQLite
- images、config、storage_configs、announcements、ip_bans
- 事实来源

Redis
- uid:<uid> 图片访问缓存
- md5:<scoped-hash> 去重缓存
- ratelimit:<scope>:ip:<hash> 限流窗口

Storage Provider
- local / S3 / WebDAV 保存 AVIF 物理对象

IndexedDB
- 浏览器本地上传历史 uploads store

localStorage
- omepic-client-token
- omepic-ui-preferences
- omepic-upload-preferences
- omepic-admin-token
```

## 9. 配置项

| 变量 | 默认值 | 说明 |
|---|---|---|
| `HTTP_ADDR` | `:8080` | 后端监听地址 |
| `DATABASE_PATH` | `data/omepic.db` | SQLite 文件路径 |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis 地址 |
| `UID_PREFIX` | `omeo_` | UID 编码前缀 |
| `UID_ENCRYPTION_KEY` | `change-me-uid-secret` | UID 编码密钥 |
| `JWT_SECRET` | `change-me-too` | 管理 JWT 签名密钥，生产必须修改 |

存储配置、公开访问基准 URL、上传策略、维护模式、限流和管理员密码均保存在 SQLite。

## 10. 项目运行方式

### 10.1 后端本地运行

前置条件：

- 安装 Go。
- 启动 Redis。
- 按需配置 `.env.example`。

命令：

```powershell
cd backend
go mod tidy
go run ./cmd/server
```

默认地址：

```text
http://localhost:8080
```

### 10.2 前端本地运行

```powershell
cd frontend
npm install
npm run dev
```

`npm run dev` 当前等价于：

```text
vite dev --host 0.0.0.0
```

前端 API base：

- 浏览器端优先 `VITE_API_BASE_URL`。
- 未设置时使用同源空 base。
- 非浏览器环境 fallback 到 `http://localhost:8080`。

### 10.3 单端口生产构建

前端静态构建并复制到后端：

```powershell
cd frontend
npm run build:backend
```

该命令等价于：

```text
vite build && node scripts/copy-static-to-backend.mjs
```

结果：

- 生成 `frontend/out/`。
- 删除旧 `backend/web/`。
- 将 `frontend/out/` 复制到 `backend/web/`。

后端构建运行：

```powershell
cd ../backend
go build ./cmd/server
./server
```

运行行为：

- 后端优先处理 `/health`、`/v1/*`、`/i/*`、`/admin/*` 等 API。
- 如果 `backend/web/index.html` 存在，其他浏览器页面路由 fallback 到静态 SPA。
- 如果 `backend/web/index.html` 不存在，后端以 API-only 模式运行。

## 11. 质量检查

后端：

```powershell
cd backend
go test ./...
go build ./cmd/server
```

前端：

```powershell
cd frontend
npm run lint
npm run typecheck
npm run build:backend
```

项目约定：每次前端 build 验证使用 `npm run build:backend`。`npm run build` 可作为局部诊断，但不能替代最终 build 验证。

## 12. 测试布局

后端已有测试覆盖包括：

- `internal/auth/jwt_test.go`
- `internal/config/config_test.go`
- `internal/http/handler/image_handler_test.go`
- `internal/http/router/frontend_test.go`
- `internal/repository/repository_test.go`
- `internal/service/admin_service_test.go`
- `internal/service/announcement_service_test.go`
- `internal/service/image_service_test.go`
- `internal/storage/storage_test.go`
- `internal/uid/codec_test.go`

前端当前主要通过：

- ESLint
- `svelte-check`
- `npm run build:backend`

## 13. 关键设计决策

### 13.1 SQLite 是事实来源，Redis 是加速层

Redis 中的 UID 和 MD5 cache 都可从 SQLite 恢复。启动时 `ImageService.Preheat()` 会预热 Redis。

### 13.2 图片删除是逻辑删除

删除移除 UID 记录和缓存，但不主动删除物理文件，避免重复上传共享物理对象时误删仍被引用的文件。

### 13.3 MD5 去重按存储实例隔离

相同 MD5 只在同一 storage key 范围内复用，避免跨 local/S3/WebDAV 或跨实例复用不存在对象。

### 13.4 真实 IP 默认不信任转发头

当前启动配置不保留可信代理环境变量，服务以无可信代理方式创建 IP resolver，因此默认使用 remote IP，不读取客户端提交的 `X-Forwarded-For` 或 `X-Real-IP`。

### 13.5 限流 fail-open

Redis 限流故障时 middleware 记录日志并放行请求，避免 Redis 单点问题导致主业务完全不可用。

### 13.6 IP 封禁使用 hash 查询并保留脱敏展示

封禁记录保存 IP hash、原始 IP、脱敏 IP。查询 active ban 使用 hash，UI 展示优先使用 masked IP。

### 13.7 前端静态 SPA 由 Go 后端单端口托管

SvelteKit 通过 adapter-static 输出到 `frontend/out/`，`build:backend` 复制到 `backend/web/`，Go 后端同时服务 API 与页面 fallback。

## 14. 新功能开发建议

### 14.1 修改后端 API

1. 在 `service` 定义业务输入输出与核心流程。
2. 在 `repository` 增加 SQL、迁移或聚合查询。
3. 在 `handler` 解析 HTTP 请求并映射错误。
4. 在 `router` 注册路由与必要 middleware。
5. 更新 `frontend/src/lib/types/index.ts` 与 `frontend/src/lib/api.ts`。
6. 增加或更新 Go tests。
7. 运行 `go test ./...` 与 `go build ./cmd/server`。

### 14.2 修改前端页面

1. 路由页面放在 `frontend/src/routes/`。
2. 共享业务 UI 放在 `frontend/src/lib/components/studio/`。
3. API 调用统一放在 `frontend/src/lib/api.ts`。
4. 类型统一放在 `frontend/src/lib/types/index.ts`。
5. 全局状态使用 `frontend/src/lib/stores/` 中的 Svelte runes store。
6. 本地上传历史使用 `frontend/src/lib/indexeddb/upload-history.ts`。
7. 运行 `npm run lint`、`npm run typecheck`、`npm run build:backend`。

### 14.3 修改安全治理能力

1. 涉及真实 IP 时先确认 `clientip.Resolver` 规则。
2. 新增 IP 聚合字段时同步后端 model、repository 查询、service 输出、前端 types。
3. 管理端新增安全操作必须走 AdminAuth。
4. 涉及封禁、解封、删除图片等破坏性操作必须有确认交互。
5. 避免在日志和 UI 中不必要地暴露 token、secret 或完整 IP。

### 14.4 修改存储能力

1. 扩展 `config.RuntimeStorageConfig`。
2. 扩展 SQLite `storage_configs` schema 和 repository 读写。
3. 扩展 `storage.Provider` 或 provider 构造逻辑。
4. 扩展 `AdminService` 配置视图与输入。
5. 扩展前端 storage 类型与设置 UI。
6. 不要在公开接口泄露 secret。

## 15. 快速索引

| 主题 | 文件 |
|---|---|
| 后端入口 | `backend/cmd/server/main.go` |
| 路由注册 | `backend/internal/http/router/router.go` |
| 静态前端 fallback | `backend/internal/http/router/frontend.go` |
| 真实 IP 解析 | `backend/internal/http/clientip/resolver.go` |
| 图片 handler | `backend/internal/http/handler/image_handler.go` |
| 管理 handler | `backend/internal/http/handler/admin_handler.go` |
| 公告 handler | `backend/internal/http/handler/announcement_handler.go` |
| 健康检查 | `backend/internal/http/handler/health_handler.go` |
| 管理鉴权 middleware | `backend/internal/http/middleware/auth_middleware.go` |
| 限流 middleware | `backend/internal/http/middleware/rate_limit_middleware.go` |
| 图片业务 | `backend/internal/service/image_service.go` |
| 图片转换 | `backend/internal/service/image_transform.go` |
| 管理业务 | `backend/internal/service/admin_service.go` |
| Abuse 规则 | `backend/internal/service/abuse.go` |
| IP 工具 | `backend/internal/service/ip_utils.go` |
| 运行时设置 | `backend/internal/service/runtime_settings.go` |
| 公告业务 | `backend/internal/service/announcement_service.go` |
| SQLite repository | `backend/internal/repository/repository.go` |
| Redis cache | `backend/internal/cache/redis_cache.go` |
| Redis 限流 | `backend/internal/ratelimit/redis_limiter.go` |
| 存储抽象 | `backend/internal/storage/storage.go` |
| UID codec | `backend/internal/uid/codec.go` |
| JWT auth | `backend/internal/auth/jwt.go` |
| 前端根 layout | `frontend/src/routes/+layout.svelte` |
| 前端上传页 | `frontend/src/routes/+page.svelte` |
| 前端历史页 | `frontend/src/routes/history/+page.svelte` |
| 前端 API 页 | `frontend/src/routes/api/+page.svelte` |
| 管理端 layout | `frontend/src/routes/admin/dashboard/+layout.svelte` |
| 管理端概览页 | `frontend/src/routes/admin/dashboard/+page.svelte` |
| 管理端图片页 | `frontend/src/routes/admin/dashboard/images/+page.svelte` |
| 管理端安全页 | `frontend/src/routes/admin/dashboard/security/+page.svelte` |
| 管理端设置页 | `frontend/src/routes/admin/dashboard/settings/+page.svelte` |
| App Shell | `frontend/src/lib/components/studio/AppShell.svelte` |
| 上传区域 | `frontend/src/lib/components/studio/CanvasDropzone.svelte` |
| 图片表格 | `frontend/src/lib/components/studio/ImageDataTable.svelte` |
| 图片预览弹窗 | `frontend/src/lib/components/studio/ImagePreviewDialog.svelte` |
| 管理图片详情抽屉 | `frontend/src/lib/components/studio/ImageDetailDrawer.svelte` |
| IP 封禁弹窗 | `frontend/src/lib/components/studio/BanIPDialog.svelte` |
| IP 详情面板 | `frontend/src/lib/components/studio/IPDetailPanel.svelte` |
| 存储实例管理 | `frontend/src/lib/components/studio/StorageInstanceManager.svelte` |
| 公告管理 | `frontend/src/lib/components/studio/AnnouncementManager.svelte` |
| 前端 API client | `frontend/src/lib/api.ts` |
| 前端类型 | `frontend/src/lib/types/index.ts` |
| 前端偏好状态 | `frontend/src/lib/stores/preferences.svelte.ts` |
| Toast 状态 | `frontend/src/lib/stores/toast.svelte.ts` |
| IndexedDB 上传历史 | `frontend/src/lib/indexeddb/upload-history.ts` |
| 前端工具函数 | `frontend/src/lib/utils.ts` |
| i18n 文案 | `frontend/src/lib/i18n.ts` |
| 静态复制脚本 | `frontend/scripts/copy-static-to-backend.mjs` |
