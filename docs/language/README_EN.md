# 🖼️ OmePic

**A self-hosted image hosting service — automatic AVIF conversion · MD5 deduplication · multi-backend storage.**

> [中文](../../README.md)

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

---

## 📑 Table of Contents

- [📸 Screenshots](#-screenshots)
- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [🖥️ Platform Support](#️-platform-support)
- [💾 Storage Backends](#-storage-backends)
- [📡 API Overview](#-api-overview)
- [⚙️ Environment Variables](#️-environment-variables)
- [⚙️ Runtime Settings](#️-runtime-settings-1)
- [🛠️ Tech Stack](#️-tech-stack)
- [🏗️ Architecture](#️-architecture)
- [📂 Project Structure](#-project-structure)
- [🧑‍💻 Development](#-development)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

---

## 📸 Screenshots

<p align="center">
  <img src="../../docs/screenshots/home.png" width="90%" alt="Upload Homepage" />
</p>

<p align="center">
  <img src="../../docs/screenshots/admin-login.png" width="44%" alt="Admin Login" />
  <img src="../../docs/screenshots/admin-dashboard.png" width="44%" alt="Admin Dashboard" />
</p>

## ✨ Features

| Category | Highlights |
|----------|------------|
| **Image Processing** | Automatic AVIF conversion (configurable quality/speed/concurrency), animated GIF preservation (up to 300 frames), MD5 deduplication |
| **Storage** | Local filesystem / S3 / WebDAV, managed at runtime, sensitive credentials AES-256-GCM encrypted |
| **Upload** | Drag & drop / paste / URL upload, upload history persisted in IndexedDB |
| **Admin** | JWT-protected dashboard, image management, storage config, system settings, IP banning, abuse monitoring, announcements |
| **Security** | Enforced secret configuration, short JWT TTL + revocation, CORS separation, rate-limit fail-closed, credential encryption |
| **Deployment** | Single-port deployment (API + frontend), Docker Compose, runtime config hot-reload |

## 🚀 Quick Start

### Prerequisites

- **Go** 1.25+ · **Node.js** 20+ · **Redis** 7+

### Launch

```bash
git clone https://github.com/OuOumm/OmePic.git && cd OmePic

# Configure environment
cp .env.production.example .env
# Required: JWT_SECRET, UID_ENCRYPTION_KEY, SECRET_ENCRYPTION_KEY (all ≥32 chars)

# Option 1: Direct run
cd backend && go run ./cmd/server

# Option 2: Docker Compose (recommended for production)
docker compose up -d
```

The server starts at `http://localhost:8080`.

### Frontend Development

```bash
cd frontend && npm install && npm run dev
```

Dev server runs with hot reload. API calls proxy to the backend.

### Production Single-Port Build

```bash
cd frontend && npm run build:backend   # Compile static assets to backend/web/
cd ../backend && go run ./cmd/server   # Single port serves both API and frontend
```

### First Login

Open `http://localhost:8080/admin`, default password **`admin123`**. Change it immediately in **Settings → Password**.

> ⚠️ High-risk operations (storage config, system settings) are rejected until the default password is changed.

## 🖥️ Platform Support

| Environment | AVIF Encoder | Animated GIF | Notes |
|-------------|-------------|--------------|-------|
| **Docker / Linux** | [lilliput](https://github.com/discord/lilliput) | ✅ | Recommended for production |
| **macOS** | [lilliput](https://github.com/discord/lilliput) | ✅ | Requires lilliput C dependencies |
| **Windows** | [gen2brain/avif](https://github.com/gen2brain/avif) | ❌ | Static AVIF only; animated GIF upload returns an error |

## 💾 Storage Backends

| Backend | Use Case |
|---------|----------|
| **Local** `local` | Server filesystem (default: `data/images/`) |
| **S3** `s3` | AWS S3, MinIO, or any S3-compatible service |
| **WebDAV** `webdav` | Any WebDAV-compatible server |

Multiple instances of each backend can coexist. Uploads can optionally let the user choose a storage target. S3/WebDAV credentials are AES-256-GCM encrypted in SQLite.

## 📡 API Overview

<details>
<summary><strong>Public Endpoints</strong> (CORS supported)</summary>

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/v1/runtime-settings` | Site/upload configuration |
| `GET` | `/v1/announcements` | Published announcements |
| `POST` | `/v1/image` | Upload image (requires `X-Token`) |
| `GET` | `/i/:uid.avif` | Serve image |
| `DELETE` | `/i/:uid.avif` | Delete image (requires same `X-Token`) |

</details>

<details>
<summary><strong>Admin Endpoints</strong> (JWT Bearer auth, strict same-origin)</summary>

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/admin/login` | Admin authentication |
| `PUT` | `/admin/password` | Change password |
| `GET` | `/admin/status` | Global upload statistics |
| `GET` | `/admin/images` | Paginated image list |
| `DELETE` | `/admin/images` | Batch delete images |
| `GET/PUT` | `/admin/system-settings` | Runtime settings |
| `GET/POST/PUT/DELETE` | `/admin/config/storage-instances` | Storage instance management |
| `POST` | `/admin/config/default` | Set default storage |
| `GET/POST` | `/admin/storage/health` | Storage health checks |
| `GET/POST/DELETE` | `/admin/ip-bans` | IP ban management |
| `GET` | `/admin/abuse/overview` | Abuse statistics |
| `GET/POST/PUT/DELETE` | `/admin/announcements` | Announcement management |
| `POST` | `/admin/cloudflare/purge-image-cache` | Cloudflare cache purge |

</details>

## ⚙️ Environment Variables

<details>
<summary>Click to expand full list</summary>

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | **Yes** | — | JWT signing key, ≥32 chars |
| `UID_ENCRYPTION_KEY` | **Yes** | — | UID obfuscation key, ≥32 chars |
| `SECRET_ENCRYPTION_KEY` | **Yes** | — | AES-256-GCM credential encryption key, exactly 32 bytes |
| `PUBLIC_BASE_URL` | Production | — | Public-facing URL |
| `HTTP_ADDR` | No | `:8080` | Listen address |
| `DATABASE_PATH` | No | `data/omepic.db` | SQLite path |
| `REDIS_URL` | No | `redis://localhost:6379/0` | Redis connection URL |
| `UID_PREFIX` | No | `omeo_` | UID plaintext prefix |
| `APP_ENV` | No | — | Set to `production` for strict checks |

> All other settings (storage, upload limits, AVIF parameters, maintenance mode, rate limits) are managed through the admin dashboard — no environment variables needed.

</details>

## ⚙️ Runtime Settings

Managed from the admin dashboard **Settings** page. Changes take effect immediately.

| Setting | Default | Description |
|---------|---------|-------------|
| Site Name | `OmePic` | UI and page title |
| Max Upload Size | `20 MB` | Per-file limit |
| Allowed MIME Types | `image/jpeg, png, gif, webp, avif` | Accepted formats |
| AVIF Quality / Speed | `60` / `8` | Encoder params (0-100 / 0-10) |
| Max Image Pixels | `40,000,000` | Decoded pixel limit |
| AVIF Concurrency / Timeout | `2` / `30s` | Conversion concurrency and per-image timeout |
| Allow Storage Selection | `true` | Let uploaders pick storage target |
| Maintenance Mode | `false` | Block uploads with custom message |
| Rate Limits | `120/min` · `20/10min` | General / upload-specific |
| Cloudflare Purge | `false` | Optional single-image URL cache purge |
| Real IP Source | `remote-addr` | Resolve real client IP behind proxies |

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go + [Gin](https://github.com/gin-gonic/gin) |
| Database | SQLite ([modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite), pure Go) |
| Cache | Redis ([go-redis](https://github.com/redis/go-redis)) |
| Image | [lilliput](https://github.com/discord/lilliput) (Linux/macOS) · [gen2brain/avif](https://github.com/gen2brain/avif) (Windows) |
| Frontend | Svelte 5 + SvelteKit 2 + Tailwind CSS |
| Auth | [golang-jwt/v5](https://github.com/golang-jwt/jwt) + Redis revocation |
| Storage | [minio-go](https://github.com/minio/minio-go) (S3) · [gowebdav](https://github.com/studio-b12/gowebdav) (WebDAV) |

## 🏗️ Architecture

<details>
<summary>Click to expand architecture diagram and request flow</summary>

```
┌─────────────────────────────────────────────────┐
│                  Browser                         │
│   SvelteKit SPA (Static Export)                  │
│   Upload UI · Admin UI · Settings                │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│            Gin HTTP Router (Go)                  │
│   Middleware (Security Headers / CORS /          │
│   Auth / Rate Limiting)                          │
│   Handlers · Frontend Static Serving             │
└───────┬──────────┬──────────┬───────────────────┘
        ▼          ▼          ▼
  ┌──────────┐ ┌────────┐ ┌────────────┐
  │  Image   │ │ Admin  │ │  Storage   │
  │ Service  │ │ Service│ │  Manager   │
  └────┬─────┘ └────────┘ └─────┬──────┘
       ▼                        ▼
  ┌──────────┐          ┌──────────────────┐
  │  SQLite  │          │ Local / S3 /     │
  │ + Redis  │          │ WebDAV Provider  │
  └──────────┘          └──────────────────┘
```

**Request flow**: Browser → Gin router (security headers → CORS → body limit → auth/rate limit) → service layer → SQLite + Redis + storage backend

</details>

## 📂 Project Structure

<details>
<summary>Click to expand directory structure</summary>

```
OmePic/
├── backend/
│   ├── cmd/server/              # Entry point + secret validation
│   ├── internal/
│   │   ├── auth/                # JWT + Redis revocation
│   │   ├── cache/               # Redis client
│   │   ├── config/              # Env config loading
│   │   ├── http/
│   │   │   ├── handler/         # HTTP handlers
│   │   │   ├── middleware/      # Security headers, CORS, auth, rate limit
│   │   │   └── router/          # Route registration
│   │   ├── repository/          # SQLite data access
│   │   ├── secrets/             # AES-256-GCM encryption
│   │   ├── service/             # Business logic
│   │   ├── storage/             # Local / S3 / WebDAV providers
│   │   └── uid/                 # UID encoding (Snowflake + XOR + Base62)
│   └── web/                     # Production frontend assets (generated)
├── frontend/
│   └── src/
│       ├── lib/
│       │   ├── api.ts           # API client
│       │   ├── components/      # UI components
│       │   ├── stores/          # Svelte 5 runes stores
│       │   └── i18n.ts          # Internationalization
│       └── routes/              # SvelteKit pages
├── .github/workflows/ci.yml    # CI pipeline
├── Dockerfile                   # Multi-stage build
└── docker-compose.yml           # Docker Compose
```

</details>

## 🧑‍💻 Development

```bash
# Backend verification
cd backend && go vet ./... && go test ./... && go build ./...

# Frontend verification
cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend
```

## 🤝 Contributing

Contributions of any kind are welcome — bug reports, feature suggestions, documentation improvements, or code submissions.

### How to Contribute

1. **Fork** this repository
2. Create a feature branch from `main`: `git checkout -b feat/your-feature`
3. Make your changes and ensure they pass local verification
4. Use [Conventional Commits](https://www.conventionalcommits.org/) format: `<type>(<scope>): <description>`
5. Submit a **Pull Request** to `main`

### Commit Convention

| Type | Purpose | Example |
|------|---------|---------|
| `feat` | New feature | `feat(backend): add WebP upload support` |
| `fix` | Bug fix | `fix(frontend): fix mobile upload button` |
| `docs` | Documentation | `docs: complete API documentation` |
| `refactor` | Refactoring | `refactor(service): extract dedup logic` |
| `chore` | Maintenance | `chore: upgrade dependencies` |

### Reporting Bugs

Please include: environment, reproduction steps, expected vs actual behavior, and relevant logs.

### Code Standards

- **Backend**: `go vet` with no warnings
- **Frontend**: ESLint + Prettier configuration
- **New features**: Write tests for critical logic

## 🙏 Acknowledgements

Thanks to the [Linux.do](https://linux.do/) community for their support and feedback.

## 📄 License

[MIT](../../LICENSE) © ououmm
