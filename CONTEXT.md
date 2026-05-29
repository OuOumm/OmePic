# OmePic Domain Context

## Core Concepts

### Image
- **ImageRecord**: 图片元数据记录，包含 UID、Token、存储信息、MD5 哈希、上传 IP 等。
- **CachedImage**: Redis 缓存中的图片元数据，必须包含 `storage_key` 以便按实例解析对象。
- **UploadInput**: 上传输入，包含文件流/字节、MIME 类型、大小、客户端 Token、客户端 IP、可选 `storage_key`。
- **UploadOutput**: 上传输出，包含公共 URL 和 `duplicate` 标记；Markdown/BBCode 由前端从 URL 派生。

### Storage
- **StorageConfig / RuntimeStorageConfig**: 存储实例配置（本地、S3、WebDAV），通过 SQLite `storage_configs` 运行时管理。
- **StorageManager**: 存储管理器，负责 Provider 生命周期、按 `storage_key` 解析和热重载。
- **Provider**: 存储提供者接口（`Save` 兼容 helper、`SaveStream` 流式写入、`Open`、`Delete`）。
- **StorageHealthCheck**: 存储探测历史记录，保存在 `storage_health_checks`，后台 UI 可查看最新状态和趋势。

### Runtime Settings
- **RuntimeSettings**: 运行时配置，包含站点信息、公开 URL、上传策略、MIME 白名单、AVIF 质量/速度/并发/超时、图片像素上限、维护模式、限流、Cloudflare purge 配置。
- **RuntimeSettingsManager**: 运行时配置管理器，负责加载、默认值补齐、验证、持久化和热更新。
- **PublicRuntimeSettingsView**: 公开配置视图，暴露站点、上传限制、`upload.avif_max_concurrency`、功能开关和安全存储选项。
- **CORS Allowed Origins**: 通过后台运行时设置配置允许跨域的 Origin 列表，支持热更新。公开 API 支持跨域，管理 API 限制同源。
- **Real IP Source**: 后台运行时设置，指定真实客户端 IP 来源头（如 `X-Forwarded-For`、`X-Real-IP`、`CF-Connecting-IP` 或直连 `RemoteAddr`），支持热更新。

### Admin
- **AdminService**: 管理后台服务，包含登录、密码修改、图片管理、存储配置、系统设置、IP 封禁、滥用分析、Cloudflare purge、存储健康检查代理等。
- **AdminAuth**: 管理员认证（JWT）。采用单管理员模式，TTL 较短（如 2–4 小时）。
- **Admin Revocation**: Redis 存储 `admin_revoked_before` 时间戳；修改密码时写入当前时间；JWT 验证时比对 `iat`，早于该时间则视为失效。

### Client Token (X-Token)
- **X-Token**: 前端自生成的匿名凭据（UUID/hex），后端不签发。上传时原样存入 `images.token`，删除时比对校验。
- **双重角色**: 既是归属标记（"谁上传的"），也是删除授权凭据（"谁有权删"）。
- **已知取舍**: 用户清除浏览器 localStorage 后永久丢失删除能力。这是核心设计的有意取舍，不通过引入账户体系修复。

### Security
- **IPBan**: IP 封禁记录，持久化 `ip_hash` 和单一展示字段 `ip_address`；不要恢复旧的 masked IP API 字段。
- **AbuseOverview / AbuseIPDetail**: 滥用统计概览与 IP 详情。
- **Client IP Resolver**: 当前启动配置使用 `clientip.NewResolver(nil, "")`，默认不信任转发头。

### UID
- **UIDCodec**: UID 编解码器，基于 Snowflake + XOR 混淆 + Base62；公开 UID 不应暴露稳定明文前缀，XOR 不作为强加密安全边界。

### Cache
- **ImageLookupCache**: Redis UID 图片缓存接口。
- **MD5MappingCache**: 按 `model.MD5MappingKey` 构造的 `storage_key + md5` 去重映射。
- **RateLimit**: Redis fixed-window 限流，按解析后的客户端 IP 哈希分 scope 计数。

### Announcement
- **Announcement**: 公告实体，支持草稿/发布/归档、优先级、时间窗口和 Markdown 内容。
- **Public Announcement Acknowledgement**: 前端只有点击确认按钮才写入最新公告已读时间；关闭/遮罩/Esc 不应标记已读。

### Cloudflare Image Cache Purge
- Cloudflare Zone ID、API Token 和 API Base URL 是后台运行时配置，不是启动环境变量。
- 删除图片前可清理 `{public_base_url}/i/{uid}.avif`；批量删除应尽量使用 Cloudflare `files` 数组合并清理。

### Frontend Modal Portal
- 所有共享 `fixed inset-0` 模态/抽屉根节点使用 `attachViewportPortal()` 挂载到 `document.body`，避免在 history/admin 等嵌套路由中遮罩被容器裁剪。
- `attachViewportPortal()` 不替代 `attachAccessibleDialog()`；焦点陷阱、Esc 和 ARIA 语义仍由 accessible dialog 负责。

## Key Relationships

1. **Upload Flow**: UploadInput → ImageService.Upload → 原始字节 MD5 → storage-scoped 去重 → AVIF 流式转换 → Storage.Provider.SaveStream → ImageRecord。异步模式下:接收 multipart → 存原始文件 → 立即返回 UID + `status: "processing"` → 后台协程完成 AVIF 转换 + 存储写入 → 前端轮询 `GET /v1/image/:uid/status` 直到 `status: "ready"`。

### Upload Status
- **processing**: 已接收原始文件，AVIF 转换与存储写入仍在后台进行。
- **ready**: 异步处理完成，图片可正常访问。
- **failed**: 后台转换或存储失败；不做自动重试。服务启动时扫描遗留 `processing` 记录并标记为 `failed`。
2. **Deduplication**: `storage_key + original_md5` → Redis/SQLite 查找 → 同存储实例复用已有物理文件；不同存储实例分别保存。
3. **AVIF Conversion**: 原始图片 → 像素上限校验 → AVIF 编码器（RuntimeSettings quality/speed/concurrency/timeout）。
4. **Storage Resolution**: `storage_key` → StorageManager.ForKey → Provider。
5. **Admin Storage Health**: StorageHealthService → Provider probe → SQLite `storage_health_checks` → Admin API/frontend chart。
6. **Static Frontend Serving**: `npm run build:backend` → `frontend/out/` → `backend/web/` → Go router filesystem serving。

## Architecture Layers

- **Handler Layer**: HTTP 请求解析和响应构造。
- **Service Layer**: 业务逻辑、事务协调、运行时配置和跨依赖流程。
- **Repository Layer**: SQLite schema、迁移、CRUD、聚合和健康历史。
- **Cache Layer**: Redis UID 缓存、MD5 去重缓存、限流。
- **Storage Layer**: 本地/S3/WebDAV 文件存储抽象。
- **Frontend Layer**: SvelteKit 静态 SPA、共享 studio 组件、runes stores、API helpers、IndexedDB 历史。

## Design Principles

1. **Separation of Concerns**: 每层只关注自己的职责。
2. **Interface Segregation**: 使用窄接口隔离依赖，避免业务路径持有不需要的能力。
3. **Dependency Injection**: 通过构造函数注入依赖。
4. **Error Handling**: 统一的错误类型和 HTTP 映射。
5. **Configuration Management**: 启动密钥/路径来自环境变量，运行时业务配置保存在 SQLite。
6. **Cross-layer Contracts**: API 响应字段、前端类型、文档和 Trellis spec 必须同步更新。