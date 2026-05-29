# OmePic 项目架构总览

## 一、项目概览

OmePic 是一个功能完整的自托管图片托管服务，支持图片上传、存储、浏览和管理。图片上传后自动转换为 AVIF 格式，支持 MD5 去重，提供多种存储后端选择。

### 核心特性

- 图片上传（支持拖拽、粘贴、URL 抓取）
- 自动转换为 AVIF 格式（质量、速度、并发、超时和像素上限可在后台 runtime/system settings 中调整）
- Redis 支持的 MD5 去重（按存储实例作用域）
- 多种存储后端：本地文件系统、S3 兼容、WebDAV
- 前端 Upload History（IndexedDB 持久化）
- JWT 保护的管理后台
- 运行时动态配置（无需重启）
- 存储健康检查与历史趋势
- Cloudflare 图片 URL 缓存清理
- 公告系统
- IP 封禁与滥用监控
- 令牌认证（非用户系统，基于 X-Token 头）

## 二、整体架构

```text
┌─────────────────────────────────────────────────────┐
│                    Browser                           │
│  SvelteKit SPA (Static Export)                      │
│  ┌───────────┐ ┌──────────┐ ┌──────────────────┐   │
│  │ Upload UI │ │ Admin UI │ │ Settings / Stats  │   │
│  └─────┬─────┘ └────┬─────┘ └────────┬─────────┘   │
│        │            │                │              │
└────────┼────────────┼────────────────┼──────────────┘
         │            │                │
         ▼            ▼                ▼
┌──────────────────────────────────────────────────────┐
│              Gin HTTP Router (backend)                │
│  ┌────────────┐ ┌──────────┐ ┌──────────────────┐   │
│  │ Middleware  │ │ Handlers │ │  Frontend SPA    │   │
│  │ (Auth/Rate/ │ │ (API)    │ │  Static Serving  │   │
│  │  Logging)   │ │          │ │  (production)    │   │
│  └────────────┘ └────┬─────┘ └──────────────────┘   │
│                      │                               │
└──────────────────────┼───────────────────────────────┘
                       │
          ┌────────────┼────────────┬────────────────┬────────────────┐
          ▼            ▼            ▼                ▼                ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐  ┌──────────────┐ ┌──────────────┐
   │ Service  │ │ Service  │ │ Service  │  │  Storage     │ │ Health /     │
   │ Image    │ │ Admin    │ │ Announce │  │  Manager     │ │ StorageHealth│
   └────┬─────┘ └────┬─────┘ └──────────┘  └──────┬───────┘ └──────┬───────┘
        │          │                               │                │
        ▼          ▼                               ▼                ▼
   ┌──────────┐ ┌──────────┐              ┌──────────────────┐ ┌──────────┐
   │Repository│ │  Cache   │              │ Local / S3 /     │ │Health    │
   │(SQLite)  │ │ (Redis)  │              │ WebDAV Provider  │ │History DB │
   └──────────┘ └──────────┘              └──────────────────┘ └──────────┘
```

## 三、后端分层

| 层次 | 包路径 | 职责 |
|------|--------|------|
| **配置层** | `internal/config` | 从环境变量加载配置，定义运行时存储配置结构 |
| **路由层** | `internal/http/router` | Gin 路由注册，CORS/中间件组装 |
| **中间件层** | `internal/http/middleware` | JWT 认证、速率限制、请求日志 |
| **处理器层** | `internal/http/handler` | HTTP 请求解析、Service 调用、响应构造 |
| **服务层** | `internal/service` | 业务逻辑：图片上传/删除、管理操作、公告、滥用检测、存储健康、Cloudflare purge |
| **仓储层** | `internal/repository` | SQLite 数据访问、schema migration、CRUD、聚合与存储健康历史 |
| **缓存层** | `internal/cache` | Redis 缓存：UID 元数据、MD5 去重映射 |
| **存储层** | `internal/storage` | 存储抽象：Local / S3 / WebDAV 实现 |
| **模型层** | `internal/model` | 数据结构和类型定义 |
| **工具层** | `internal/uid`, `internal/auth`, `internal/iputil`, `internal/ratelimit` | UID 编码、JWT 工具、IP 工具、速率限制器 |

## 四、前端分层

| 层次 | 路径 | 职责 |
|------|------|------|
| **页面层** | `src/routes/` | SvelteKit 页面路由 |
| **组件层** | `src/lib/components/studio/` | UI 组件（上传区、数据表格、对话框、存储健康图表等） |
| **API 层** | `src/lib/api.ts` | 后端 API 调用封装 |
| **状态层** | `src/lib/stores/` | 偏好设置（Svelte runes）、Toast 通知 |
| **工具层** | `src/lib/` | 剪贴板、客户端令牌、i18n、上传队列、IndexedDB |
| **类型层** | `src/lib/types/index.ts` | TypeScript 类型定义 |

## 五、数据流示例：图片上传

```text
1. 用户拖拽图片到前端上传区
2. CanvasDropzone 组件触发 handleFiles()
3. 前端验证文件大小和 MIME 类型
4. 通过 XHR POST /v1/image (需 X-Token 头)
5. ImageHandler.Upload (Handler)
   ├─ 读取 multipart/form-data 中的文件
   ├─ 解析 Content-Type
   ├─ 检查文件大小限制
   └─ 调用 ImageService.Upload (Service)
6. ImageService.Upload
   ├─ 检查维护模式、IP 封禁
   ├─ 校验 runtime settings（大小/MIME）
   ├─ 解析目标 storage_key 或默认存储
   ├─ 计算原始上传字节 MD5，并按 storage_key+md5 加 keyed mutex
   ├─ 查 Redis (md5:{storageKey}:{hash}) → 回退 SQLite
   ├─ [去重] → 创建新 UID 行，复用同存储实例文件路径
   ├─ [新文件] → 像素上限校验 → 流式 AVIF 转换 → Storage.Provider.SaveStream()
   │   ├─ Local → io.Copy 到文件
   │   ├─ S3 → minio.PutObject(stream)
   │   └─ WebDAV → 流式写入
   ├─ 写入 SQLite (images 表)
   └─ 写入 Redis (uid:{uid}, md5:{storageKey}:{hash})
7. 前端接收 UploadOutput.url/duplicate → 派生 UID、Markdown、BBCode
8. 保存到 IndexedDB (upload_history) 持久化
```

## 六、关键技术选型

| 组件 | 技术 | 原因 |
|------|------|------|
| HTTP 框架 | Gin | Go 生态最流行的 Web 框架，性能优异 |
| 数据库 | SQLite (modernc.org/sqlite) | 纯 Go 实现，无需外部依赖，适合中小规模 |
| 缓存 | Redis (go-redis) | 高性能 KV 缓存，支持去重预热 |
| 图片转换 | gen2brain/avif | 纯 Go AVIF 编码，无需 CGO |
| ID 生成 | Snowflake + XOR + Base62 | 分布式友好、URL 安全、公开 token 不透明；XOR 属于混淆而非强加密 |
| 前端框架 | Svelte 5 + SvelteKit | 编译型框架，打包体积小，响应式新 runes API |
| 样式 | Tailwind CSS 3 | 实用优先，快速构建 |
| JWT | golang-jwt/v5 | 成熟标准库 |
| S3 | minio-go/v7 | 官方推荐，兼容 AWS S3 和私有存储 |
| WebDAV | gowebdav | 轻量级 WebDAV 客户端 |
