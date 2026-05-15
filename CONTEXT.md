# OmePic Domain Context

## Core Concepts

### Image
- **ImageRecord**: 图片元数据记录，包含 UID、Token、存储信息、MD5 哈希等
- **CachedImage**: Redis 缓存中的图片元数据
- **UploadInput**: 上传输入，包含文件内容、MIME 类型、存储选择等
- **UploadOutput**: 上传输出，包含公共 URL 和去重标志

### Storage
- **StorageConfig**: 存储实例配置（本地、S3、WebDAV）
- **StorageManager**: 存储管理器，负责 Provider 生命周期和热重载
- **Provider**: 存储提供者接口（Save、Open、Delete）

### Runtime Settings
- **RuntimeSettings**: 运行时配置，包含站点信息、上传策略、AVIF 参数等
- **RuntimeSettingsManager**: 运行时配置管理器，负责加载、验证和持久化

### Admin
- **AdminService**: 管理后台服务，包含登录、图片管理、IP 封禁、滥用分析等
- **AdminAuth**: 管理员认证（JWT）

### Security
- **IPBan**: IP 封禁记录
- **AbuseOverview**: 滥用统计概览

### UID
- **UIDCodec**: UID 编解码器，基于 Snowflake + XOR + Base62

### Cache
- **ImageCache**: 图片缓存接口（Redis）
- **MD5Mapping**: MD5 去重映射

## Key Relationships

1. **Upload Flow**: UploadInput → ImageService.Upload → Storage.Provider.Save → ImageRecord
2. **Deduplication**: MD5 哈希 → Redis/SQLite 查找 → 复用已有物理文件
3. **AVIF Conversion**: 原始图片 → AVIF 编码器（使用 RuntimeSettings.AvifQuality/AvifSpeed）
4. **Storage Resolution**: StorageKey → StorageManager.ForKey → Provider

## Architecture Layers

- **Handler Layer**: HTTP 请求解析和响应构造
- **Service Layer**: 业务逻辑和事务协调
- **Repository Layer**: SQLite 数据访问
- **Cache Layer**: Redis 缓存操作
- **Storage Layer**: 文件存储抽象

## Design Principles

1. **Separation of Concerns**: 每层只关注自己的职责
2. **Interface Segregation**: 使用小接口隔离依赖
3. **Dependency Injection**: 通过构造函数注入依赖
4. **Error Handling**: 统一的错误类型和映射
5. **Configuration Management**: 运行时配置与代码分离