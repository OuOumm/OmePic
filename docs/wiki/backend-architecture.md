# 后端架构详解

## 目录

1. [入口点：main.go](#1-入口点maingo)
2. [配置层：config](#2-配置层config)
3. [仓储层：repository](#3-仓储层repository)
4. [缓存层：cache](#4-缓存层cache)
5. [存储层：storage](#5-存储层storage)
6. [服务层：service](#6-服务层service)
7. [HTTP 处理器层：handler](#7-http-处理器层handler)
8. [中间件层：middleware](#8-中间件层middleware)
9. [路由层：router](#9-路由层router)
10. [工具模块](#10-工具模块)
11. [模型层：model](#11-模型层model)

---

## 1. 入口点：main.go

**路径**: [backend/cmd/server/main.go](file:///d:/Works/MyProject/OmePic/backend/cmd/server/main.go)

### 启动流程

```text
1. 初始化 JSON 格式的 slog Logger
2. config.Load() 加载环境变量配置
3. 创建 data/ 目录
4. repository.New() → 打开/创建 SQLite 数据库
5. repo.Migrate() → 执行数据库迁移（建表）
6. repo.InitializeStorageCatalog() → 初始化存储配置目录
7. storage.NewManager() → 创建存储管理器（含 Provider 初始化）
8. uid.NewCodec() → 创建 UID 编解码器（Snowflake + XOR）
9. cache.NewClient() → 创建 Redis 连接
10. cache.NewWithClient() → 创建图片缓存实例
11. ratelimit.NewRedisLimiter() → 创建速率限制器
12. repo.Ping() + imageCache.Ping() → 双依赖健康检查
13. service.NewRuntimeSettingsManager() → 运行时配置管理器
14. settings.Load() → 补齐缺失的默认运行时配置并从数据库加载
15. 创建 Service：
    - ImageService（图片上传/删除/解析/预热）
    - AdminService（管理后台逻辑、Cloudflare purge、存储健康 API 代理）
    - AnnouncementService（公告管理）
    - HealthService（SQLite/Redis 健康检查）
    - StorageHealthService（存储健康检查与历史）
16. storageHealthService.StartHeartbeat(...) → 后台存储健康心跳
17. clientip.NewResolver(nil, "") → IP 解析器（当前启动配置默认不信任转发头）
18. imageService.Preheat() → Redis 预热
19. router.New() → 组装 Gin Engine
20. engine.Run() → 启动 HTTP 服务器
```

### 关键依赖关系

```
main.go
├─ config.Load()
├─ repository.New() ──────┬─ Migrate()
│                         └─ InitializeStorageCatalog()
├─ storage.NewManager()
├─ uid.NewCodec()
├─ cache.NewClient() ─────┬─ NewWithClient()
│                         └─ NewRedisLimiter()
├─ service:
│  ├─ NewRuntimeSettingsManager() ── Load()
│  ├─ NewImageService(repo, cache, storage, settings, uid.Generate, uid.Validate, logger)
│  ├─ NewAdminService(repo, storage, settings, imageService, jwtSecret, adminEnv)
│  ├─ NewAnnouncementService(repo)
│  ├─ NewHealthService(repo, cache)
│  └─ NewStorageHealthService(repo, storage, logger)
├─ clientip.NewResolver(nil, "")
└─ router.New(handler, middleware, ...)
```

---

## 2. 配置层：config

**路径**: [backend/internal/config/config.go](file:///d:/Works/MyProject/OmePic/backend/internal/config/config.go)

### AppConfig

`AppConfig` 只承载启动必需且不能从 SQLite 读取的环境变量：

| 字段 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| HTTPAddr | `HTTP_ADDR` | `:8080` | 监听地址 |
| DatabasePath | `DATABASE_PATH` | `data/omepic.db` | SQLite 文件路径 |
| RedisURL | `REDIS_URL` | `redis://localhost:6379/0` | Redis 连接 URL |
| UIDPrefix | `UID_PREFIX` | `omeo_` | UID 前缀 |
| UIDEncryptionKey | `UID_ENCRYPTION_KEY` | `change-me-uid-secret` | XOR 混淆密钥（非强加密） |
| JWTSecret | `JWT_SECRET` | `change-me-too` | JWT 签名密钥 |

公开访问基准 URL、存储配置、上传策略、维护模式、限流和管理员密码均由 SQLite 管理，不再读取环境变量。

### 存储后端类型常量

```go
StorageBackendLocal  = "local"
StorageBackendS3     = "s3"
StorageBackendWebDAV = "webdav"
```

### RuntimeStorageConfig

运行时存储配置，支持多存储实例，包含所有后端类型的参数字段（通过指针和零值判断未使用字段）。

### 辅助类型

- **RuntimeStorageCatalog**: 运行时存储目录（含默认存储键和配置列表）
- **RuntimeStorageUpdate**: 存储配置更新输入（全部为指针类型，支持部分更新）

---

## 3. 仓储层：repository

**路径**: [backend/internal/repository/repository.go](file:///d:/Works/MyProject/OmePic/backend/internal/repository/repository.go)

### Repository 结构体

```go
type Repository struct {
    db *sql.DB
}
```

使用 `modernc.org/sqlite` 纯 Go SQLite 驱动。

### SQLite 配置

```go
PRAGMA journal_mode = WAL;     // WAL 模式提升并发
PRAGMA synchronous = NORMAL;   // 平衡性能与安全
PRAGMA busy_timeout = 5000;    // 忙等待 5 秒
PRAGMA foreign_keys = ON;      // 外键约束
PRAGMA temp_store = MEMORY;    // 临时表存内存
PRAGMA mmap_size = 268435456;  // 256MB 内存映射
```

### 数据表（6张）

| 表名 | 用途 | 关键字段 |
|------|------|----------|
| `images` | 图片记录 | uid (UNIQUE), token, storage_key, md5_hash, file_path, ip_address |
| `config` | KV 配置 | key (PRIMARY KEY), value |
| `storage_configs` | 存储实例配置 | storage_key (UNIQUE), backend, is_default, 各后端参数字段 |
| `announcements` | 公告 | title, content, status, priority, starts_at, ends_at |
| `ip_bans` | IP 封禁 | ip_hash, ip_address, reason, expires_at |
| `storage_health_checks` | 存储健康历史 | storage_key, status, latency_ms, error_message, consecutive_failures |

### 核心方法分类

**图片操作**:
- `InsertImage`, `FindByUID`, `FindByMD5AndStorageKey`
- `DeleteByUID`, `CountByMD5`, `CountByMD5AndStorageKey`, `CountByStoredFile`
- `ListAllImages`, `SearchImages` (分页+搜索), `ImageSummaryByIP`, `ListImagesByIP`
- `AggregateStatus` (总览统计)

**存储配置**:
- `ListStorageConfigs`, `GetStorageConfigByKey`
- `CreateStorageConfig`, `UpdateStorageConfig`, `DeleteStorageConfig`
- `SetDefaultStorageConfig`, `CountImagesByStorageKey`
- `InitializeStorageCatalog` (引导初始化)

**KV 配置**:
- `GetAllConfig`, `GetConfigValue`, `SetConfigValue`, `UpsertConfigValues`, `InsertMissingConfigValues`

**IP 封禁**:
- `CreateIPBan`, `ListIPBans`, `GetIPBan`, `DeleteIPBan`
- `FindActiveIPBanByHash`, `FindActiveIPBanByIP`
- `ActiveIPBansByHash`, `CountActiveIPBans`

**公告**:
- `ListPublicAnnouncements`, `ListAnnouncements`, `GetAnnouncement`
- `CreateAnnouncement`, `UpdateAnnouncement`, `DeleteAnnouncement`, `ArchiveAnnouncement`

**存储健康**:
- `UpsertStorageHealthCheck`, `GetStorageHealthCheck`, `GetStorageHealthCheckByID`
- `ListStorageHealthChecks`, `ListStorageHealthHistory`

**滥用统计**:
- `AbuseOverviewTotals`, `TopAbuseIPs`, `TopAbuseTokens`, `IPDetail`

### 关键设计

- 使用 `sql.ErrNoRows` 判断未找到（`IsNotFound()` 辅助函数）
- `ensureImageColumn()`：向后兼容的列添加
- `backfillImageStorageKeys()`：为旧数据填充 storage_key
- `normalizeDefaultStorageConfig()`：确保有且仅有一个默认存储

---

## 4. 缓存层：cache

**路径**: [backend/internal/cache/redis_cache.go](file:///d:/Works/MyProject/OmePic/backend/internal/cache/redis_cache.go)

### 缓存接口

`ImageCache` 是兼容聚合接口，业务代码优先依赖更窄的 seam，避免上传/访问路径拿到不需要的能力：

```go
type ImageLookupCache interface {
    GetImage(ctx, uid) → *CachedImage
    SetImage(ctx, record) → error
    DeleteImage(ctx, uid) → error
}

type ImagePreheatCache interface {
    SetImages(ctx, records) → error
}

type MD5MappingCache interface {
    GetMD5(ctx, key model.MD5MappingKey) → string
    SetMD5(ctx, key model.MD5MappingKey, uid) → error
    SetMD5IfAbsent(ctx, key model.MD5MappingKey, uid) → error
    DeleteMD5(ctx, key model.MD5MappingKey) → error
}

type MD5MappingPreheatCache interface {
    SetMD5Mappings(ctx, []model.MD5Mapping) → error
}

type HealthCache interface { Ping(ctx) → error }
```

MD5 cache key 必须通过 `model.MD5MappingKey.CacheScope()` 生成，调用方不要手写 storage/hash 字符串。

### Redis 键设计

```text
uid:{uid}              → JSON(CachedImage)   # 图片元数据缓存
md5:{storageKey}:{hash} → uid                 # MD5 → UID 映射（按存储实例隔离去重）
```

### Redis 连接配置

```go
DialTimeout:  3s
ReadTimeout:  2s
WriteTimeout: 2s
PoolSize:     16
PoolTimeout:  3s
MinIdleConns: 2
ContextTimeoutEnabled: true
```

### 预热机制

启动时调用 `imageService.Preheat()`：
1. 从 SQLite 加载所有图片记录
2. Pipeline 批量写入 Redis `uid:{uid}` 键
3. Pipeline 批量写入 Redis `md5:{storageKey}:{hash}` 映射

---

## 5. 存储层：storage

**路径**: [backend/internal/storage/storage.go](file:///d:/Works/MyProject/OmePic/backend/internal/storage/storage.go)

### Provider 接口

```go
type Provider interface {
    Name() string
    Save(ctx, objectKey, data, contentType) → (string, error)
    SaveStream(ctx, objectKey, reader, size, contentType) → (string, error)
    Open(ctx, objectKey) → (OpenResult, error)
    Delete(ctx, objectKey) → error
}
```

### 三种实现

#### localProvider
- `Save`: compatibility helper，内部转为 reader 写入
- `SaveStream`: `os.Create` + `io.Copy` + `os.MkdirAll`
- `Open`: `os.Open` + `file.Stat`
- `Delete`: `os.Remove`

#### s3Provider
- 依赖: `minio-go/v7`
- `Save`: compatibility helper，内部转为 `bytes.Reader`
- `SaveStream`: `client.PutObject`
- `Open`: `client.GetObject` + `object.Stat`
- `Delete`: `client.RemoveObject`
- 支持 `BucketLookupPath` (路径风格) 和 `BucketLookupAuto`

#### webdavProvider
- 依赖: `gowebdav`
- `Save`: compatibility helper，内部转为 reader 写入
- `SaveStream`: `client.MkdirAll` + 流式上传/写入
- `Open`: `client.ReadStream` + `client.Stat`
- `Delete`: `client.Remove`

### Manager

```go
type Manager struct {
    mu         sync.RWMutex
    configs    map[string]config.RuntimeStorageConfig
    defaultKey string
    providers  map[string]Provider
}
```

| 方法 | 说明 |
|------|------|
| `NewManager(settings)` | 创建管理器，调用 Reconfigure |
| `Current()` | 获取当前默认存储的 ResolvedProvider |
| `ForKey(key)` | 按 storage_key 获取 Provider（惰性初始化，缓存结果） |
| `Reconfigure(settings)` | 热重载所有存储配置（清空 Provider 缓存） |
| `CurrentKey()` | 返回当前默认存储键 |
| `CurrentBackend()` | 返回当前默认存储后端类型 |

### 对象键生成

```go
BuildObjectKey(uid, extension) → "2025/12/uid.avif"
```

按年月分区存储，便于文件系统管理。

### 辅助函数

- `ValidateConfig()`: 验证存储配置是否有效
- `normalizeConfigs()`: 规范化配置列表，确保唯一键和默认值

---

## 6. 服务层：service

### 6.1 ImageService

**路径**: [backend/internal/service/image_service.go](file:///d:/Works/MyProject/OmePic/backend/internal/service/image_service.go)

#### 核心依赖

```go
type ImageService struct {
    repo                *repository.Repository
    imageCache          cache.ImageLookupCache
    preheat             cache.ImagePreheatCache
    md5Cache            cache.MD5MappingCache
    md5Preheat          cache.MD5MappingPreheatCache
    storage             uploadStorageResolver
    settings            *RuntimeSettingsManager
    logger              *slog.Logger
    generateUID         UIDGenerator
    validateUID         UIDValidator
    encoder             func(io.Reader, io.Writer, AVIFConversionSettings) error
    avifLimiterMu       sync.Mutex
    avifLimiter         chan struct{}
    avifLimiterSize     int
    hashLocks           *keyedMutex
    imageURLCachePurger ImageURLCachePurger
}
```

#### 关键方法

| 方法 | 说明 |
|------|------|
| `Upload(ctx, input)` | 核心上传逻辑：校验 → MD5 去重 → AVIF 转换 → 存储 → 持久化 |
| `Delete(ctx, uid, token, isAdmin, ip)` | 逻辑删除：验证 Token → 删 SQLite → 删 Redis → 修复 MD5 映射 |
| `Open(ctx, uid)` | 图片解析并打开底层存储对象，供 HTTP 层流式输出 |
| `Resolve(ctx, uid)` | 图片解析：优先查 Redis → 回退 SQLite |
| `Preheat(ctx)` | 启动预热：加载全量数据到 Redis |
| `PublicRuntimeSettings(ctx)` | 获取公开运行时配置 |
| `EffectivePublicBaseURL(requestBase)` | 获取有效的公网 BaseURL |

#### 上传流程详解

1. 验证 Token 非空
2. 检查 IP 是否被封禁
3. 检查维护模式
4. 检查文件大小（运行时配置 vs 硬上限 20MB）
5. 检查 MIME 类型白名单
6. 确定目标存储（按 storage_key 参数或默认）
7. 将上传源按需 spool 到临时文件，并计算原始上传字节 MD5
8. **按 `storage_key + md5` 加 keyed mutex**，只串行化同一去重范围，避免全局上传锁
9. 查重：先查 Redis scoped MD5 → 回退 SQLite
10. 去重命中：创建新 UID 行，复用同存储实例下的文件路径，跳过 AVIF 转换
11. 新文件：按 `max_image_pixels` 校验像素 → 受 `avif_max_concurrency` 与 `avif_conversion_timeout_seconds` 限制的流式 AVIF 转换 → Provider.SaveStream() → SQLite Insert → Redis Set
12. 返回 UploadOutput（只含 URL 与 duplicate；Markdown/BBCode 由前端派生）

#### 删除流程

1. 校验 IP（非管理员时）
2. 规范化 UID（带 `.avif` 后缀）
3. SQLite 查询记录
4. 验证 Token 所有权
5. SQLite 删除行
6. Redis 删除 `uid:{uid}`
7. 检查 MD5 引用计数：
   - 若为零 → 删除 `md5:{storageKey}:{hash}`
   - 否则 → 修复 MD5 映射指向剩余的第一个记录

#### UID 规范化

```
serveUID:  "abc123.avif" → 去掉 ".avif" → "abc123"
deleteUID: "abc123.avif" → 同上
storedUID: "abc123"      → 直接验证
```

### 6.2 AdminService

**路径**: [backend/internal/service/admin_service.go](file:///d:/Works/MyProject/OmePic/backend/internal/service/admin_service.go)

| 方法 | 说明 |
|------|------|
| `Login(password)` | 校验 SQLite bcrypt 密码哈希并返回 JWT（24h 有效期） |
| `ChangePassword(ctx, oldPassword, newPassword)` | 校验旧密码和新密码强度（至少 8 位，包含大写、小写和符号）后写入新的 bcrypt 哈希 |
| `Status(ctx)` | 获取全局统计 |
| `Images(ctx, page, pageSize, search)` | 分页搜索图片列表 |
| `DeleteImages(ctx, uids)` | 批量删除图片 |
| `CreateIPBan(ctx, input)` | 创建 IP 封禁（支持按 UID 或 IP 地址） |
| `IPBans(ctx)` | 列出所有封禁记录 |
| `DeleteIPBan(ctx, id)` | 删除封禁 |
| `DeleteImagesByIPBan(ctx, id)` | 删除某封禁关联的所有图片 |
| `AbuseOverview(ctx, input)` | 滥用概览（TOP IP、TOP Token） |
| `AbuseIPDetail(ctx, ip)` | IP 详情 |
| `GetConfig(ctx)` | 获取存储配置 |
| `UpdateConfig(ctx, input)` | 更新存储配置（支持部分更新） |
| `CreateStorageConfig(ctx, input)` | 创建存储实例 |
| `UpdateStorageConfig(ctx, key, input)` | 更新存储实例 |
| `DeleteStorageConfig(ctx, key)` | 删除存储实例（需无引用图片） |
| `SetDefaultStorageConfig(ctx, key)` | 设置默认存储 |
| `GetSystemSettings(ctx)` | 获取系统设置（含只读环境状态） |
| `UpdateSystemSettings(ctx, input)` | 更新运行时配置 |
| `PurgeCloudflareImageCache(ctx, url)` | 手动清理单个 Cloudflare 图片 URL 缓存 |
| `StorageHealth(ctx)` / `StorageHealthHistory(ctx, key, since)` | 查询存储健康状态与历史（当前按调用创建 `StorageHealthService`，`since` 为可选 RFC3339 时间） |
| `CheckStorageHealth(ctx, key)` / `CheckAllStorageHealth(ctx)` | 立即执行存储健康检查 |

#### 存储配置管理规则

- 删除默认存储或仍有图片引用的存储会被拒绝
- 后端类型变更时不幂等，需先确保无图片引用
- 秘密字段（S3 SecretKey、WebDAV Pass）在 API 中脱敏显示（保留后4字符）
- 创建时自动生成 storage_key（slug + timestamp hex）

### 6.3 AnnouncementService

**路径**: [backend/internal/service/announcement_service.go](file:///d:/Works/MyProject/OmePic/backend/internal/service/announcement_service.go)

| 方法 | 说明 |
|------|------|
| `PublicAnnouncements(ctx)` | 获取已发布且时间窗口内的公告（最多 10 条） |
| `AdminAnnouncements(ctx)` | 获取所有公告（含草稿和已归档） |
| `CreateAnnouncement(ctx, input)` | 创建公告 |
| `UpdateAnnouncement(ctx, id, input)` | 更新公告 |
| `DeleteAnnouncement(ctx, id)` | 删除公告 |
| `ArchiveAnnouncement(ctx, id)` | 归档公告 |

#### 公告状态

- `draft`: 草稿
- `published`: 已发布（公开可见）
- `archived`: 已归档

#### 优先级

- `normal`: 普通（默认）
- `important`: 重要
- `urgent`: 紧急（排序最高）

### 6.4 RuntimeSettingsManager

**路径**: [backend/internal/service/runtime_settings.go](file:///d:/Works/MyProject/OmePic/backend/internal/service/runtime_settings.go)

#### RuntimeSettings 字段

| 字段 | 默认值 | 说明 |
|------|--------|------|
| SiteName | `OmePic` | 站点名称 |
| SiteTagline | `上传、分享和管理图片` | 站点标语 |
| PublicBaseURL | `""` | 公网基础 URL |
| MaxUploadSizeMB | `20` | 上传大小上限（MB） |
| AllowedMIMETypes | `image/jpeg,image/png,...` | 允许的 MIME 类型 |
| AvifQuality | `60` | AVIF 编码质量，范围 `0..100`，`100` 表示无损 |
| AvifSpeed | `8` | AVIF 编码速度，范围 `0..10`，数值越低通常越慢但压缩/质量取舍更好 |
| MaxImagePixels | `40000000` | 解码后像素上限，防止超大图片耗尽资源 |
| AVIFMaxConcurrency | `2` | AVIF 转换最大并发数，公开运行时设置也返回该值 |
| AVIFConversionTimeoutSeconds | `30` | 单张图片 AVIF 转换超时时间 |
| AllowStorageSelect | `true` | 允许用户选择存储 |
| MaintenanceMode | `false` | 维护模式 |
| MaintenanceMessage | `系统维护中，请稍后再试` | 维护提示信息 |
| RateLimitWindowMinutes | `1` | 速率限制窗口（分钟） |
| RateLimitMaxRequests | `120` | API 速率上限 |
| UploadRateLimitWindowMinutes | `10` | 上传速率窗口（分钟） |
| UploadRateLimitMaxRequests | `20` | 上传速率上限 |

#### 值来源与持久化

- `RuntimeSettingsManager.Load()` 会先把缺失的默认 runtime settings 写入 SQLite `config` 表。
- 缺失补齐只使用 `INSERT ... ON CONFLICT DO NOTHING`，不会覆盖已有管理员配置。
- `public_base_url` 只来自 SQLite；未配置时请求处理使用 request host 作为运行时回退。
- 新上传且未命中去重的图片会使用当前 `avif_quality` / `avif_speed` / `avif_max_concurrency` / `avif_conversion_timeout_seconds` 转换，并先校验 `max_image_pixels`；重复上传复用已有物理对象，不重新转换。

### 6.5 错误定义

**路径**: [backend/internal/service/errors.go](file:///d:/Works/MyProject/OmePic/backend/internal/service/errors.go)

```go
var (
    ErrInvalidInput          = errors.New("invalid input")
    ErrMissingToken          = errors.New("missing token")
    ErrInvalidAdminToken     = errors.New("invalid admin token")
    ErrForbidden             = errors.New("forbidden")
    ErrIPBanned              = errors.New("ip banned")
    ErrNotFound              = errors.New("not found")
    ErrConflict              = errors.New("conflict")
    ErrDependencyUnavailable = errors.New("dependency unavailable")
)
```

---

## 7. HTTP 处理器层：handler

### 7.1 ImageHandler

**路径**: [backend/internal/http/handler/image_handler.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/handler/image_handler.go)

| 方法 | 路由 | 说明 |
|------|------|------|
| `Upload` | `POST /v1/image` | 处理 multipart 文件上传 |
| `RuntimeSettings` | `GET /v1/runtime-settings` | 返回公开运行时配置 |
| `Delete` | `DELETE /i/:uid.avif` | 图片删除（需 X-Token） |
| `Serve` | `GET /i/:uid.avif` | 图片服务（返回 AVIF 文件内容） |

### 7.2 AdminHandler

**路径**: [backend/internal/http/handler/admin_handler.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/handler/admin_handler.go)

| 方法 | 路由 | 说明 |
|------|------|------|
| `Login` | `POST /admin/login` | 管理员登录 |
| `Status` | `GET /admin/status` | 获取统计 |
| `Images` | `GET /admin/images` | 图片分页列表 |
| `DeleteImages` | `DELETE /admin/images` | 批量删除 |
| `CreateIPBan` | `POST /admin/ip-bans` | 创建 IP 封禁 |
| `IPBans` | `GET /admin/ip-bans` | 封禁列表 |
| `AbuseOverview` | `GET /admin/abuse/overview` | 滥用概览 |
| `AbuseIPDetail` | `GET /admin/abuse/ip` | IP 详情 |
| `DeleteIPBan` | `DELETE /admin/ip-bans/:id` | 删除封禁 |
| `DeleteIPBanImages` | `DELETE /admin/ip-bans/:id/images` | 删除关联图片 |
| `GetConfig` | `GET /admin/config` | 获取存储配置 |
| `UpdateConfig` | `POST /admin/config` | 更新存储配置 |
| `CreateStorageConfig` | `POST /admin/config/storage-instances` | 创建存储实例 |
| `UpdateStorageConfig` | `PUT /admin/config/storage-instances/:storageKey` | 更新实例 |
| `DeleteStorageConfig` | `DELETE /admin/config/storage-instances/:storageKey` | 删除实例 |
| `SetDefaultStorageConfig` | `POST /admin/config/default` | 设置默认存储 |
| `GetSystemSettings` | `GET /admin/system-settings` | 系统设置查看 |
| `UpdateSystemSettings` | `PUT /admin/system-settings` | 系统设置更新 |
| `PurgeCloudflareImageCache` | `POST /admin/cloudflare/purge-image-cache` | 手动清理 Cloudflare 图片 URL 缓存 |
| `StorageHealth` | `GET /admin/storage/health` | 最新存储健康状态 |
| `StorageHealthHistory` | `GET /admin/storage/:key/health-history` | 存储健康历史 |
| `CheckStorageHealth` | `POST /admin/storage/:key/health-check` | 立即检查单个存储 |
| `CheckAllStorageHealth` | `POST /admin/storage/health-check-all` | 立即检查全部存储 |

### 7.3 AnnouncementHandler

**路径**: [backend/internal/http/handler/announcement_handler.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/handler/announcement_handler.go)

| 方法 | 路由 | 说明 |
|------|------|------|
| `PublicList` | `GET /v1/announcements` | 公开公告列表 |
| `AdminList` | `GET /admin/announcements` | 管理后台公告列表 |
| `Create` | `POST /admin/announcements` | 创建公告 |
| `Update` | `PUT /admin/announcements/:id` | 更新公告 |
| `Delete` | `DELETE /admin/announcements/:id` | 删除公告 |
| `Archive` | `POST /admin/announcements/:id/archive` | 归档公告 |

### 7.4 HealthHandler

**路径**: [backend/internal/http/handler/health_handler.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/handler/health_handler.go)

- `GET /health` — 检查 SQLite 和 Redis 连通性

### 7.5 错误处理

**路径**: [backend/internal/http/handler/errors.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/handler/errors.go)

- `writeServiceError()`: 统一的 Service 错误 → HTTP 错误映射
- `sanitizeErrorMessage()`: 去除内部错误前缀（`"dependency unavailable: "` 部分）

### 7.6 响应格式

**路径**: [backend/internal/response/response.go](file:///d:/Works/MyProject/OmePic/backend/internal/response/response.go)

```json
// 成功
{ "success": true, "data": { ... } }

// 失败
{ "success": false, "error": { "code": "not_found", "message": "image not found" } }
```

---

## 8. 中间件层：middleware

**路径**: [backend/internal/http/middleware/](file:///d:/Works/MyProject/OmePic/backend/internal/http/middleware/)

### AdminAuth

```go
func AdminAuth(jwtSecret string) gin.HandlerFunc
```

- 从 `Authorization: Bearer <token>` 中提取 JWT
- 验证 JWT 签名和有效期

### RateLimit

```go
func RateLimit(limiter ratelimit.Limiter, logger, policy) gin.HandlerFunc
```

- 支持两种范围：`api`（通用）和 `upload`（上传专用）
- 使用 IP 哈希作为限流键：`ratelimit:{scope}:ip:{ipHash}`
- 响应头：`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`
- 限流时（429）：upload scope 额外消耗 body 防止重复

### RequestLogger

请求日志中间件（注册在 `gin.Recovery()` 之后）

### 速率限制器

**路径**: [backend/internal/ratelimit/redis_limiter.go](file:///d:/Works/MyProject/OmePic/backend/internal/ratelimit/redis_limiter.go)

基于 Redis Lua 脚本的固定窗口限流器：

```lua
local current = redis.call("INCR", KEYS[1])
if current == 1 then redis.call("PEXPIRE", KEYS[1], ARGV[1]) end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
```

---

## 9. 路由层：router

### Router

**路径**: [backend/internal/http/router/router.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/router/router.go)

```
Gin Engine 组装顺序：
1. gin.Recovery() — 异常恢复
2. cors.New() — 未配置 runtime `public_base_url` 时允许所有来源；配置后仅允许该 public origin
3. RequestLogger — 日志
4. RateLimit (api scope) — API 通用限流
5. RateLimit (upload scope) — 上传限流
6. 公开路由注册
7. /admin 分组（加 AdminAuth 中间件）
8. 前端静态文件服务
```

### Routes

**路径**: [backend/internal/http/router/routes.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/router/routes.go)

所有路由定义集中在此文件，使用 `routeSpec` 结构体管理路径和方法。

### Frontend

**路径**: [backend/internal/http/router/frontend.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/router/frontend.go)

- 生产模式下通过 `NoRoute` 兜底提供 SPA 静态文件服务
- 安全头设置：CSP、X-Content-Type-Options、Referrer-Policy、Permissions-Policy
- API 404 保持 JSON 响应（不返回 HTML）

---

## 10. 工具模块

### UID Codec

**路径**: [backend/internal/uid/codec.go](file:///d:/Works/MyProject/OmePic/backend/internal/uid/codec.go)

UID 生成特性：
- **Snowflake ID**: 41bit 时间戳 + 5bit 数据中心 + 5bit 工作节点 + 12bit 序列号
- **XOR 混淆**: 用密钥对 SID + 前缀做循环 XOR 混淆（偏移量由 SID % 62 决定；不作为强加密安全边界）
- **Base64 编码**: XOR 结果用标准 Base64 编码
- **Base62 压缩**: Base64 字符串转 Base62（URL 安全）
- 最终格式: `{1个Base62字符(偏移)} + {Base62字符串}`

```
生成: SID → sidToBytes → payload = SID + prefix → XOR → Base64 → Base62 → token
解码: token → base62Decode → Base64 Decode → XOR → 分离 SID + prefix
```

### IP 工具

**路径**: [backend/internal/iputil/iputil.go](file:///d:/Works/MyProject/OmePic/backend/internal/iputil/iputil.go)

```go
func Hash(ip) → SHA256(ip)  // 哈希值
func Mask(ip) → "192.168.*.*"  // 脱敏
```

### JWT 工具

**路径**: [backend/internal/auth/jwt.go](file:///d:/Works/MyProject/OmePic/backend/internal/auth/jwt.go)

- `GenerateJWT(secret, ttl)` → 标准 HS256 JWT
- `ParseJWT(secret, token)` → 解析并验证
- `ParseBearer(header)` → 从 Authorization 头提取 Bearer Token

### IP 解析器

**路径**: [backend/internal/http/clientip/resolver.go](file:///d:/Works/MyProject/OmePic/backend/internal/http/clientip/resolver.go)

- 支持根据代理信任配置解析真实客户端 IP
- 支持 `X-Forwarded-For` 和 `X-Real-IP`
- 当前 `main.go` 使用 `clientip.NewResolver(nil, "")`，默认不信任转发头；部署在反向代理后时需先实现/配置可信代理 CIDR 后再信任这些头

---

## 11. 模型层：model

**路径**: [backend/internal/model/](file:///d:/Works/MyProject/OmePic/backend/internal/model/)

| 文件 | 结构体 | 用途 |
|------|--------|------|
| [image.go](file:///d:/Works/MyProject/OmePic/backend/internal/model/image.go) | `ImageRecord`, `CachedImage`, `AdminStatus` | 图片记录、缓存、总览 |
| [ip_ban.go](file:///d:/Works/MyProject/OmePic/backend/internal/model/ip_ban.go) | `IPBan`, `IPImageSummary` | IP 封禁、IP 图片汇总 |
| [abuse.go](file:///d:/Works/MyProject/OmePic/backend/internal/model/abuse.go) | `AbuseOverview`, `AbuseIPRankItem`, `AbuseTokenRankItem`, `AbuseIPDetail` | 滥用统计数据 |
| [announcement.go](file:///d:/Works/MyProject/OmePic/backend/internal/model/announcement.go) | `Announcement` | 公告（含状态/优先级常量） |
| [storage_health.go](file:///d:/Works/MyProject/OmePic/backend/internal/model/storage_health.go) | `StorageHealthCheck` | 存储健康状态与历史 |
