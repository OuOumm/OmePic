# 🖼️ OmePic

**自托管图片托管服务 — 自动 AVIF 转换 · MD5 去重 · 多后端存储**

> [US](docs/README_EN.md)

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

---

## ✨ 核心功能

- **自动 AVIF 转换** — 上传图片自动转换为 AVIF 格式，支持后台可配置的质量（0–100）和速度（0–10）参数
- **MD5 去重** — 相同内容的上传复用已有物理文件，按存储实例作用域隔离
- **多后端存储** — 支持本地文件系统、S3 兼容服务和 WebDAV，运行时动态管理，无需重启
- **管理后台** — JWT 保护的管理面板，支持图片管理、存储配置和系统设置
- **IP 封禁与滥用监控** — 封禁恶意 IP，按 IP 和 Token 追踪上传量
- **公告系统** — 发布带时间窗口和优先级的公告
- **运行时配置** — 站点名称、上传限制、MIME 白名单、AVIF 参数、维护模式、速率限制，全部可在后台 UI 中编辑
- **Token 认证** — 无需注册账户，客户端生成的 Token 标识上传者并授权删除操作
- **拖拽 / 粘贴 / URL 上传** — 灵活的上传方式，上传历史通过 IndexedDB 本地持久化
- **单端口部署** — 生产构建将前端编译进 Go 二进制文件，单一端口同时提供 API 和前端

## 📸 演示 / 截图

> 截图待补充

## 🛠️ 技术栈

| 层次 | 技术 | 用途 |
|------|------|------|
| 后端 | **Go** + [Gin](https://github.com/gin-gonic/gin) | HTTP API、中间件、路由 |
| 数据库 | **SQLite** (modernc.org/sqlite) | 元数据和配置持久化（纯 Go，无 CGO） |
| 缓存 | **Redis** (go-redis) | UID/MD5 缓存、去重查询 |
| 图片转换 | [gen2brain/avif](https://github.com/gen2brain/avif) | AVIF 编码（纯 Go） |
| 前端 | **Svelte 5** + **SvelteKit 2** + **Tailwind CSS** | SPA，静态适配器导出 |
| ID 生成 | Snowflake + XOR + Base62 | 不透明、URL 安全、不可预测的 UID |
| 认证 | [golang-jwt/v5](https://github.com/golang-jwt/jwt) | 管理员 JWT 会话 |
| S3 | [minio-go/v7](https://github.com/minio/minio-go) | S3 兼容对象存储 |
| WebDAV | [gowebdav](https://github.com/studio-b12/gowebdav) | WebDAV 存储客户端 |

## 🏗️ 架构

```
┌─────────────────────────────────────────────────┐
│                  浏览器                          │
│   SvelteKit SPA（静态导出）                      │
│   上传界面 · 管理后台 · 系统设置                 │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│            Gin HTTP 路由（Go）                   │
│   中间件（认证 / 速率限制 / 日志）               │
│   处理器 · 前端静态资源托管                      │
└───────┬──────────┬──────────┬───────────────────┘
        ▼          ▼          ▼
  ┌──────────┐ ┌────────┐ ┌────────────┐
  │  图片    │ │ 管理   │ │  存储      │
  │  服务    │ │ 服务   │ │  管理器    │
  └────┬─────┘ └────────┘ └─────┬──────┘
       ▼                        ▼
  ┌──────────┐          ┌──────────────────┐
  │  SQLite  │          │ 本地 / S3 /      │
  │ （仓储） │          │ WebDAV 提供者    │
  └────┬─────┘          └──────────────────┘
       ▼
  ┌──────────┐
  │  Redis   │
  │ （缓存） │
  └──────────┘
```

**请求流向**：浏览器 → Gin 路由（中间件鉴权/限流） → 业务服务层 → SQLite 持久化 + Redis 缓存 + 存储后端写入

## 🚀 快速开始

### 环境要求

- **Go** 1.22+
- **Node.js** 18+（含 npm）
- **Redis** 7+

### 克隆项目

```bash
git clone https://github.com/your-username/OmePic.git
cd OmePic
```

### 环境变量配置

复制示例文件并按需编辑：

```bash
cp .env.example .env
```

必需变量（完整列表见[环境变量](#-环境变量)）：

```env
HTTP_ADDR=:8080
DATABASE_PATH=data/omepic.db
REDIS_URL=redis://localhost:6379/0
UID_PREFIX=omeo_
UID_ENCRYPTION_KEY=change-me-uid-secret
JWT_SECRET=change-me-too
```

### 后端启动

```bash
cd backend
go run ./cmd/server
```

服务启动在 `HTTP_ADDR`（默认 `:8080`），SQLite 数据库和本地存储目录自动创建。

### 前端开发启动

```bash
cd frontend
npm install
npm run dev
```

开发服务器在独立端口运行，带热重载。API 请求代理到后端。

### 生产单端口构建

```bash
cd frontend
npm run build:backend
cd ../backend
go run ./cmd/server
```

`build:backend` 将 SvelteKit 应用编译为静态资源并复制到 `backend/web/`。Go 二进制在单一端口同时提供 API 和前端服务。

### 首次登录

1. 打开 `http://localhost:8080/admin`
2. 使用默认密码登录：**`admin123`**
3. 在 **设置 → 密码** 中立即修改密码

> ⚠️ 默认密码首次登录时自动哈希写入 SQLite。请在公开部署前修改。

## 🔧 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `HTTP_ADDR` | 否 | `:8080` | HTTP 服务监听地址 |
| `DATABASE_PATH` | 否 | `data/omepic.db` | SQLite 数据库文件路径 |
| `REDIS_URL` | 否 | `redis://localhost:6379/0` | Redis 连接地址 |
| `UID_PREFIX` | 否 | `omeo_` | UID 加密前的明文前缀（尾部下划线自动规范化） |
| `UID_ENCRYPTION_KEY` | **是** | `change-me-uid-secret` | UID 加密用的 XOR 密钥（为空时回退到 `JWT_SECRET`） |
| `JWT_SECRET` | **是** | `change-me-too` | 签发管理员 JWT 的密钥 |

> 其他所有设置（存储、上传限制、AVIF 参数、维护模式、速率限制）均通过管理后台运行时配置，无需设置环境变量。

## 📡 API 概览

### 公开端点

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/health` | 健康检查（SQLite + Redis） |
| `GET` | `/v1/runtime-settings` | 获取公开的站点/上传配置 |
| `GET` | `/v1/announcements` | 获取已发布的公告 |
| `GET` | `/v1/storage-options` | 获取可用存储实例（仅展示用） |
| `POST` | `/v1/image` | 上传图片（需要 `X-Token`） |
| `GET` | `/i/:uid.avif` | 获取图片（返回 AVIF 字节） |
| `DELETE` | `/i/:uid.avif` | 删除图片（需要与上传时相同的 `X-Token`） |

### 管理端点

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/admin/login` | 管理员认证，返回 JWT |
| `PUT` | `/admin/password` | 修改管理员密码 |
| `GET` | `/admin/status` | 全局上传统计 |
| `GET` | `/admin/images` | 分页图片列表（支持搜索） |
| `DELETE` | `/admin/images` | 按 UID 批量删除图片 |
| `GET` | `/admin/system-settings` | 获取运行时 + 只读设置 |
| `PUT` | `/admin/system-settings` | 更新运行时设置 |
| `GET` | `/admin/config` | 获取存储目录 |
| `POST` | `/admin/config` | 更新存储配置（兼容路由） |
| `POST/PUT/DELETE` | `/admin/config/storage-instances` | 存储实例增删改查 |
| `POST` | `/admin/config/default` | 设置默认存储 |
| `GET/POST/DELETE` | `/admin/ip-bans` | 管理 IP 封禁 |
| `GET` | `/admin/abuse/overview` | 滥用统计概览 |
| `GET` | `/admin/abuse/ip` | 指定 IP 的滥用详情 |
| `GET/POST/PUT/DELETE` | `/admin/announcements` | 管理公告 |

> 完整 API 文档：[docs/api-reference.md](docs/api-reference.md)

## 💾 存储后端

OmePic 支持三种存储后端，通过管理后台运行时配置，无需重启：

| 后端 | 键值 | 适用场景 |
|------|------|----------|
| **本地** | `local` | 文件存储在服务器本地文件系统（默认：`data/images/`） |
| **S3** | `s3` | AWS S3、MinIO 或任何 S3 兼容服务 |
| **WebDAV** | `webdav` | 任何 WebDAV 兼容服务器 |

- 每种后端可创建多个实例（如两个 S3 存储桶）
- 上传时可选让用户选择存储目标
- 每张图片记录其 `storage_key`，不允许对已使用的存储实例切换后端类型

## ⚙️ 运行时配置

所有运行时配置在管理后台（`/admin → 设置`）中管理，修改后立即生效，无需重启。

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 站点名称 | `OmePic` | UI 和页面标题中显示 |
| 站点标语 | `上传、分享和管理图片` | 浏览器标题元数据 |
| 公开 URL | *（自动）* | 覆盖公开访问地址（默认使用请求 Host） |
| 最大上传大小 | `20` MB | 单文件上传限制 |
| 允许的 MIME 类型 | `image/jpeg, png, gif, webp, avif` | 接受的上传格式 |
| AVIF 质量 | `60` | 编码器质量（0=最低，100=无损） |
| AVIF 速度 | `8` | 编码器速度（0=最慢/最佳压缩，10=最快） |
| 允许选择存储 | `true` | 允许上传者选择存储目标 |
| 维护模式 | `false` | 开启后阻止上传并显示自定义消息 |
| 速率限制 | `120 次/分钟` | 通用 API 速率限制 |
| 上传速率限制 | `20 次/10分钟` | 上传接口专用速率限制 |

## 📂 项目结构

```
OmePic/
├── backend/
│   ├── cmd/server/              # 启动入口
│   ├── internal/
│   │   ├── auth/                # JWT 生成与验证
│   │   ├── cache/               # Redis 客户端与预热
│   │   ├── config/              # 环境变量配置加载
│   │   ├── http/
│   │   │   ├── handler/         # HTTP 处理器（图片、管理、健康检查）
│   │   │   ├── middleware/      # 认证、速率限制、日志
│   │   │   └── router/          # Gin 路由注册
│   │   ├── iputil/              # 可信 IP 解析
│   │   ├── model/               # 数据结构
│   │   ├── ratelimit/           # 速率限制器
│   │   ├── repository/          # SQLite 数据访问层
│   │   ├── response/            # JSON 响应辅助函数
│   │   ├── service/             # 业务逻辑层
│   │   ├── storage/             # 本地 / S3 / WebDAV 提供者
│   │   └── uid/                 # UID 编码（Snowflake + XOR + Base62）
│   ├── web/                     # 生产前端资源（构建生成）
│   └── data/                    # 运行时数据（SQLite、图片）
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api.ts           # API 客户端
│   │   │   ├── components/      # UI 组件（studio/）
│   │   │   ├── indexeddb/       # 上传历史持久化
│   │   │   ├── stores/          # Svelte runes 状态管理
│   │   │   ├── types/           # TypeScript 类型定义
│   │   │   └── utils/           # 工具函数（剪贴板、Token、i18n）
│   │   └── routes/              # SvelteKit 页面
│   │       ├── +page.svelte     # 首页 / 上传
│   │       ├── admin/           # 管理后台
│   │       └── history/         # 上传历史
│   └── package.json
└── docs/
    ├── api-reference.md
    ├── architecture-overview.md
    └── README_EN.md
```

## 🧑‍💻 开发指南

### 后端

```bash
cd backend

# 启动服务
go run ./cmd/server

# 运行所有测试
go test ./...

# 格式检查
gofmt -l .

# 运行特定测试
go test ./internal/service/ -run TestUpload
```

### 前端

```bash
cd frontend

# 开发服务器
npm run dev

# 代码检查
npm run lint

# 类型检查
npm run typecheck

# 运行测试
npm run test

# 生产构建（复制到 backend/web/）
npm run build:backend
```

### 完整验证

```bash
# 后端
cd backend && go test ./...

# 前端
cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend
```

## 📄 许可证

[MIT](LICENSE) © ououmm

---
