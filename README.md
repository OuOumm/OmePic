# 🖼️ OmePic

**自托管图片托管服务 — 自动 AVIF 转换 · MD5 去重 · 多后端存储**

> [English](docs/language/README_EN.md)

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

---

## 📑 目录

- [📸 截图](#-截图)
- [✨ 功能概览](#-功能概览)
- [🚀 快速开始](#-快速开始)
- [🖥️ 平台支持](#️-平台支持)
- [💾 存储后端](#-存储后端)
- [📡 API 概览](#-api-概览)
- [⚙️ 环境变量](#️-环境变量)
- [⚙️ 运行时配置](#️-运行时配置-1)
- [🛠️ 技术栈](#️-技术栈)
- [🏗️ 架构](#️-架构)
- [📂 项目结构](#-项目结构)
- [🧑‍💻 开发指南](#-开发指南)
- [🤝 贡献](#-贡献)
- [📄 许可证](#-许可证)

---

## 📸 截图

<p align="center">
  <img src="docs/screenshots/home.png" width="90%" alt="上传首页" />
</p>

<p align="center">
  <img src="docs/screenshots/admin-login.png" width="44%" alt="管理后台登录" />
  <img src="docs/screenshots/admin-dashboard.png" width="44%" alt="管理后台" />
</p>

## ✨ 功能概览

| 分类 | 特性 |
|------|------|
| **图片处理** | 自动 AVIF 转换（质量/速度/并发可调）、动画 GIF 保留（最多 300 帧）、MD5 去重 |
| **存储** | 本地文件系统 / S3 / WebDAV，运行时动态管理，敏感凭据 AES-256-GCM 加密 |
| **上传方式** | 拖拽 / 粘贴 / URL 上传，上传历史通过 IndexedDB 本地持久化 |
| **管理后台** | JWT 保护的管理面板、图片管理、存储配置、系统设置、IP 封禁、滥用监控、公告系统 |
| **安全** | 强制密钥配置、JWT 短 TTL + 撤销、CORS 分离、限流 fail-closed、凭据加密存储 |
| **部署** | 单端口部署（API + 前端）、Docker Compose 一键启动、运行时配置热更新 |

## 🚀 快速开始

### 环境要求

- **Go** 1.25+ · **Node.js** 20+ · **Redis** 7+

### 启动

```bash
git clone https://github.com/OuOumm/OmePic.git && cd OmePic

# 配置环境变量
cp .env.production.example .env
# 必填: JWT_SECRET, UID_ENCRYPTION_KEY, SECRET_ENCRYPTION_KEY (均 ≥32 字符)

# 方式一：直接运行
cd backend && go run ./cmd/server

# 方式二：Docker Compose（推荐生产环境）
docker compose up -d
```

服务在 `http://localhost:8080` 启动。

### 前端开发

```bash
cd frontend && npm install && npm run dev
```

开发服务器带热重载，API 请求代理到后端。

### 生产单端口构建

```bash
cd frontend && npm run build:backend   # 编译静态资源到 backend/web/
cd ../backend && go run ./cmd/server   # 单端口同时提供 API 和前端
```

### 首次登录

打开 `http://localhost:8080/admin`，默认密码 **`admin123`**，登录后请立即在 **设置 → 密码** 中修改。

> ⚠️ 未修改默认密码前，存储配置和系统设置等高风险操作会被拒绝。

## 🖥️ 平台支持

| 环境 | AVIF 编码器 | GIF 动画 | 说明 |
|------|-------------|----------|------|
| **Docker / Linux** | [lilliput](https://github.com/discord/lilliput) | ✅ | 生产推荐 |
| **macOS** | [lilliput](https://github.com/discord/lilliput) | ✅ | 需安装 lilliput C 依赖 |
| **Windows** | [gen2brain/avif](https://github.com/gen2brain/avif) | ❌ | 仅静态 AVIF，动画 GIF 会报错 |

## 💾 存储后端

| 后端 | 适用场景 |
|------|----------|
| **本地** `local` | 服务器本地文件系统（默认：`data/images/`） |
| **S3** `s3` | AWS S3、MinIO 或任何 S3 兼容服务 |
| **WebDAV** `webdav` | 任何 WebDAV 兼容服务器 |

每种后端可创建多个实例，上传时可选存储目标。S3/WebDAV 凭据在 SQLite 中加密存储。

## 📡 API 概览

<details>
<summary><strong>公开端点</strong>（CORS 支持）</summary>

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/health` | 健康检查 |
| `GET` | `/v1/runtime-settings` | 站点/上传配置 |
| `GET` | `/v1/announcements` | 已发布公告 |
| `POST` | `/v1/image` | 上传图片（需要 `X-Token`） |
| `GET` | `/i/:uid.avif` | 获取图片 |
| `DELETE` | `/i/:uid.avif` | 删除图片（需要相同 `X-Token`） |

</details>

<details>
<summary><strong>管理端点</strong>（JWT Bearer 认证，严格同源）</summary>

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/admin/login` | 管理员认证 |
| `PUT` | `/admin/password` | 修改密码 |
| `GET` | `/admin/status` | 全局上传统计 |
| `GET` | `/admin/images` | 分页图片列表 |
| `DELETE` | `/admin/images` | 批量删除图片 |
| `GET/PUT` | `/admin/system-settings` | 运行时设置 |
| `GET/POST/PUT/DELETE` | `/admin/config/storage-instances` | 存储实例管理 |
| `POST` | `/admin/config/default` | 设置默认存储 |
| `GET/POST` | `/admin/storage/health` | 存储健康检查 |
| `GET/POST/DELETE` | `/admin/ip-bans` | IP 封禁管理 |
| `GET` | `/admin/abuse/overview` | 滥用统计 |
| `GET/POST/PUT/DELETE` | `/admin/announcements` | 公告管理 |
| `POST` | `/admin/cloudflare/purge-image-cache` | Cloudflare 缓存清理 |

</details>

## ⚙️ 环境变量

<details>
<summary>展开查看完整列表</summary>

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `JWT_SECRET` | **是** | — | JWT 签名密钥，≥32 字符 |
| `UID_ENCRYPTION_KEY` | **是** | — | UID 混淆密钥，≥32 字符 |
| `SECRET_ENCRYPTION_KEY` | **是** | — | AES-256-GCM 凭据加密密钥，恰好 32 字节 |
| `PUBLIC_BASE_URL` | 生产必填 | — | 公开访问 URL |
| `HTTP_ADDR` | 否 | `:8080` | 监听地址 |
| `DATABASE_PATH` | 否 | `data/omepic.db` | SQLite 路径 |
| `REDIS_URL` | 否 | `redis://localhost:6379/0` | Redis 连接地址 |
| `UID_PREFIX` | 否 | `omeo_` | UID 明文前缀 |
| `APP_ENV` | 否 | — | 设为 `production` 启用严格检查 |

> 其他设置（存储、上传限制、AVIF 参数、维护模式、速率限制等）均通过管理后台配置，无需环境变量。

</details>

## ⚙️ 运行时配置

在管理后台 **设置** 中管理，修改后立即生效。

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 站点名称 | `OmePic` | UI 和页面标题 |
| 最大上传大小 | `20 MB` | 单文件限制 |
| 允许的 MIME 类型 | `image/jpeg, png, gif, webp, avif` | 接受的格式 |
| AVIF 质量 / 速度 | `60` / `8` | 编码器参数（0-100 / 0-10） |
| 最大图片像素 | `40,000,000` | 解码后像素上限 |
| AVIF 并发 / 超时 | `2` / `30s` | 转换并发与单张超时 |
| 允许选择存储 | `true` | 让上传者选择存储目标 |
| 维护模式 | `false` | 阻止上传并显示自定义消息 |
| 速率限制 | `120/min` · `20/10min` | 通用 / 上传专用 |
| Cloudflare purge | `false` | 可选清理图片 URL 缓存 |
| 真实 IP 来源 | `remote-addr` | 反向代理后解析真实 IP |

## 🛠️ 技术栈

| 层次 | 技术 |
|------|------|
| 后端 | Go + [Gin](https://github.com/gin-gonic/gin) |
| 数据库 | SQLite ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)，纯 Go) |
| 缓存 | Redis ([go-redis](https://github.com/redis/go-redis)) |
| 图片转换 | [lilliput](https://github.com/discord/lilliput) (Linux/macOS) · [gen2brain/avif](https://github.com/gen2brain/avif) (Windows) |
| 前端 | Svelte 5 + SvelteKit 2 + Tailwind CSS |
| 认证 | [golang-jwt/v5](https://github.com/golang-jwt/jwt) + Redis 撤销 |
| 存储 | [minio-go](https://github.com/minio/minio-go) (S3) · [gowebdav](https://github.com/studio-b12/gowebdav) (WebDAV) |

## 🏗️ 架构

<details>
<summary>展开查看架构图与请求流向</summary>

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
│   中间件（安全头 / CORS / 认证 / 限流）         │
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
  │ + Redis  │          │ WebDAV 提供者    │
  └──────────┘          └──────────────────┘
```

**请求流向**：浏览器 → Gin 路由（安全头 → CORS → Body 限制 → 鉴权/限流） → 业务服务层 → SQLite + Redis + 存储后端

</details>

## 📂 项目结构

<details>
<summary>展开查看目录结构</summary>

```
OmePic/
├── backend/
│   ├── cmd/server/              # 启动入口 + 密钥校验
│   ├── internal/
│   │   ├── auth/                # JWT + Redis 撤销
│   │   ├── cache/               # Redis 客户端
│   │   ├── config/              # 环境变量加载
│   │   ├── http/
│   │   │   ├── handler/         # HTTP 处理器
│   │   │   ├── middleware/      # 安全头、CORS、认证、限流
│   │   │   └── router/          # 路由注册
│   │   ├── repository/          # SQLite 数据访问
│   │   ├── secrets/             # AES-256-GCM 加密
│   │   ├── service/             # 业务逻辑
│   │   ├── storage/             │ 本地 / S3 / WebDAV
│   │   └── uid/                 # UID 编码（Snowflake + XOR + Base62）
│   └── web/                     # 生产前端资源（构建生成）
├── frontend/
│   └── src/
│       ├── lib/
│       │   ├── api.ts           # API 客户端
│       │   ├── components/      # UI 组件
│       │   ├── stores/          # Svelte 5 runes 状态
│       │   └── i18n.ts          # 国际化
│       └── routes/              # SvelteKit 页面
├── .github/workflows/ci.yml    # CI 流水线
├── Dockerfile                   # 多阶段构建
└── docker-compose.yml           # Docker Compose
```

</details>

## 🧑‍💻 开发指南

```bash
# 后端验证
cd backend && go vet ./... && go test ./... && go build ./...

# 前端验证
cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend
```

## 🤝 贡献

欢迎任何形式的贡献！无论是 Bug 报告、功能建议、文档改进还是代码提交，都非常感谢。

### 如何贡献

1. **Fork** 本仓库
2. 从 `main` 创建特性分支：`git checkout -b feat/your-feature`
3. 进行修改，确保通过本地验证
4. 使用 [约定式提交](https://www.conventionalcommits.org/zh-hans/) 格式：`<类型>(<范围>): <描述>`
5. 提交 **Pull Request** 到 `main`

### 提交规范

| 类型 | 用途 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(backend): 添加 WebP 上传支持` |
| `fix` | 修复 | `fix(frontend): 修复移动端上传按钮` |
| `docs` | 文档 | `docs: 补全 API 文档` |
| `refactor` | 重构 | `refactor(service): 抽离去重逻辑` |
| `chore` | 杂务 | `chore: 升级依赖版本` |

### 报告 Bug

提交 Issue 时请包含：运行环境、复现步骤、预期与实际行为、相关日志。

### 代码规范

- **后端**：`go vet` 无警告
- **前端**：ESLint + Prettier 配置
- **新功能**：为关键逻辑编写测试

## 🙏 致谢

感谢 [Linux.do](https://linux.do/) 社区的支持与反馈。

## 📄 许可证

[MIT](LICENSE) © ououmm
