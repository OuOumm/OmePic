# 🖼️ OmePic

**A self-hosted image hosting service with automatic AVIF conversion, MD5 deduplication, and multi-backend storage.**

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

---

## ✨ Features

- **Automatic AVIF conversion** — uploads are converted to AVIF with configurable quality (0–100) and speed (0–10)
- **MD5 deduplication** — identical uploads reuse the existing physical file, scoped per storage instance
- **Multi-backend storage** — local filesystem, S3-compatible, and WebDAV, managed at runtime without restarts
- **Admin dashboard** — JWT-protected panel for image management, storage configuration, and system settings
- **IP banning & abuse monitoring** — block abusive IPs, track upload volume by IP and token
- **Announcements** — publish time-windowed announcements with priority levels
- **Runtime configuration** — site name, upload limits, MIME allowlist, AVIF parameters, maintenance mode, and rate limits — all editable from the admin UI
- **Token-based auth** — no user accounts; client-generated tokens identify uploaders and authorize deletes
- **Drag & drop / paste / URL upload** — flexible upload UX with upload history persisted in IndexedDB
- **Single-port deployment** — production build compiles frontend into the Go binary

## 📸 Demo / Screenshots

> Screenshots coming soon

## 🛠️ Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Backend | **Go** + [Gin](https://github.com/gin-gonic/gin) | HTTP API, middleware, routing |
| Database | **SQLite** (modernc.org/sqlite) | Persistent metadata & config (pure Go, no CGO) |
| Cache | **Redis** (go-redis) | UID/MD5 cache, deduplication lookups |
| Image | [gen2brain/avif](https://github.com/gen2brain/avif) | AVIF encoding (pure Go) |
| Frontend | **Svelte 5** + **SvelteKit 2** + **Tailwind CSS** | SPA with static adapter export |
| ID | Snowflake + XOR + Base62 | Opaque, URL-safe, unpredictable UIDs |
| Auth | [golang-jwt/v5](https://github.com/golang-jwt/jwt) | Admin JWT sessions |
| S3 | [minio-go/v7](https://github.com/minio/minio-go) | S3-compatible object storage |
| WebDAV | [gowebdav](https://github.com/studio-b12/gowebdav) | WebDAV storage client |

## 🏗️ Architecture

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
│   Middleware (Auth / Rate Limit / Logging)        │
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
  │  (repo)  │          │ WebDAV Provider  │
  └────┬─────┘          └──────────────────┘
       ▼
  ┌──────────┐
  │  Redis   │
  │  (cache) │
  └──────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go** 1.22+
- **Node.js** 18+ (with npm)
- **Redis** 7+

### Clone

```bash
git clone https://github.com/your-username/OmePic.git
cd OmePic
```

### Environment Variables

Copy the example and edit as needed:

```bash
cp .env.example .env
```

Required variables (see [Environment Variables](#-environment-variables) for full list):

```env
HTTP_ADDR=:8080
DATABASE_PATH=data/omepic.db
REDIS_URL=redis://localhost:6379/0
UID_PREFIX=omeo_
UID_ENCRYPTION_KEY=change-me-uid-secret
JWT_SECRET=change-me-too
```

### Backend

```bash
cd backend
go run ./cmd/server
```

The server starts on `HTTP_ADDR` (default `:8080`). SQLite database and local storage are created automatically.

### Frontend (Development)

```bash
cd frontend
npm install
npm run dev
```

The dev server runs on a separate port with hot reload. API calls proxy to the backend.

### Production (Single-Port Build)

```bash
cd frontend
npm run build:backend
cd ../backend
go run ./cmd/server
```

`build:backend` compiles the SvelteKit app into static assets and copies them into `backend/web/`. The Go binary serves both API and frontend on a single port.

### First Login

1. Open `http://localhost:8080/admin`
2. Log in with the default password: **`admin123`**
3. Change the password immediately in **Settings → Password**

> ⚠️ The default password is auto-hashed into SQLite on first login. Change it before exposing the service publicly.

## 🔧 Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_ADDR` | No | `:8080` | Listen address for the HTTP server |
| `DATABASE_PATH` | No | `data/omepic.db` | Path to the SQLite database file |
| `REDIS_URL` | No | `redis://localhost:6379/0` | Redis connection URL |
| `UID_PREFIX` | No | `omeo_` | Plaintext prefix for encrypted UIDs (trailing underscores normalized) |
| `UID_ENCRYPTION_KEY` | **Yes** | `change-me-uid-secret` | XOR secret for UID encryption (falls back to `JWT_SECRET` if empty) |
| `JWT_SECRET` | **Yes** | `change-me-too` | Secret key for signing admin JWT tokens |

> All other settings (storage, upload limits, AVIF parameters, maintenance mode, rate limits) are managed at runtime through the admin dashboard — no environment variables needed.

## 📡 API Overview

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check (SQLite + Redis) |
| `GET` | `/v1/runtime-settings` | Public site/upload configuration |
| `GET` | `/v1/announcements` | Active published announcements |
| `GET` | `/v1/storage-options` | Available storage instances (display only) |
| `POST` | `/v1/image` | Upload image (requires `X-Token`) |
| `GET` | `/i/:uid.avif` | Serve image (returns AVIF bytes) |
| `DELETE` | `/i/:uid.avif` | Delete image (requires same `X-Token` as upload) |

### Admin Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/admin/login` | Authenticate, returns JWT |
| `PUT` | `/admin/password` | Change admin password |
| `GET` | `/admin/status` | Global upload statistics |
| `GET` | `/admin/images` | Paginated image list with search |
| `DELETE` | `/admin/images` | Batch delete images by UID |
| `GET` | `/admin/system-settings` | Get runtime + readonly settings |
| `PUT` | `/admin/system-settings` | Update runtime settings |
| `GET` | `/admin/config` | Get storage catalog |
| `POST` | `/admin/config` | Update storage config (compat) |
| `POST/PUT/DELETE` | `/admin/config/storage-instances` | CRUD storage instances |
| `POST` | `/admin/config/default` | Set default storage |
| `GET/POST/DELETE` | `/admin/ip-bans` | Manage IP bans |
| `GET` | `/admin/abuse/overview` | Abuse statistics |
| `GET` | `/admin/abuse/ip` | IP-specific abuse detail |
| `GET/POST/PUT/DELETE` | `/admin/announcements` | Manage announcements |

> Full API reference: [docs/api-reference.md](docs/api-reference.md)

## 💾 Storage Backends

OmePic supports three storage backends, configurable at runtime through the admin dashboard:

| Backend | Key | Use Case |
|---------|-----|----------|
| **Local** | `local` | Files stored on the server filesystem (default: `data/images/`) |
| **S3** | `s3` | AWS S3, MinIO, or any S3-compatible service |
| **WebDAV** | `webdav` | Any WebDAV-compatible server |

- Multiple instances of each backend can coexist (e.g., two S3 buckets)
- Uploads can optionally let the user choose a storage target
- Each image stores its `storage_key` — switching a backend type for an in-use instance is blocked

## ⚙️ Runtime Settings

All runtime settings are managed from the admin dashboard (`/admin → Settings`) and take effect immediately — no restart required.

| Setting | Default | Description |
|---------|---------|-------------|
| Site Name | `OmePic` | Displayed in UI and page title |
| Site Tagline | `上传、分享和管理图片` | Browser title metadata |
| Public Base URL | *(auto)* | Override public URL (defaults to request Host) |
| Max Upload Size | `20` MB | Per-file upload limit |
| Allowed MIME Types | `image/jpeg, png, gif, webp, avif` | Accepted upload formats |
| AVIF Quality | `60` | Encoder quality (0=worst, 100=lossless) |
| AVIF Speed | `8` | Encoder speed (0=slowest/best compression, 10=fastest) |
| Allow Storage Selection | `true` | Let uploaders pick storage target |
| Maintenance Mode | `false` | Block uploads with a custom message |
| Rate Limit | `120 req/min` | General API rate limit |
| Upload Rate Limit | `20 req/10min` | Upload-specific rate limit |

## 📂 Project Structure

```
OmePic/
├── backend/
│   ├── cmd/server/              # Entry point
│   ├── internal/
│   │   ├── auth/                # JWT generation & validation
│   │   ├── cache/               # Redis client & preheat
│   │   ├── config/              # Env config loading
│   │   ├── http/
│   │   │   ├── handler/         # HTTP handlers (image, admin, health)
│   │   │   ├── middleware/      # Auth, rate limit, logging
│   │   │   └── router/          # Gin route registration
│   │   ├── iputil/              # Trusted IP resolution
│   │   ├── model/               # Data structures
│   │   ├── ratelimit/           # Rate limiter
│   │   ├── repository/          # SQLite data access
│   │   ├── response/            # JSON response helpers
│   │   ├── service/             # Business logic
│   │   ├── storage/             # Local / S3 / WebDAV providers
│   │   └── uid/                 # UID encoding (Snowflake + XOR + Base62)
│   ├── web/                     # Production frontend assets (generated)
│   └── data/                    # Runtime data (SQLite, images)
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api.ts           # API client
│   │   │   ├── components/      # UI components (studio/)
│   │   │   ├── indexeddb/       # Upload history persistence
│   │   │   ├── stores/          # Svelte runes stores
│   │   │   ├── types/           # TypeScript type definitions
│   │   │   └── utils/           # Helpers (clipboard, token, i18n)
│   │   └── routes/              # SvelteKit pages
│   │       ├── +page.svelte     # Homepage / upload
│   │       ├── admin/           # Admin dashboard
│   │       └── history/         # Upload history
│   └── package.json
└── docs/
    ├── api-reference.md
    └── architecture-overview.md
```

## 🧑‍💻 Development

### Backend

```bash
cd backend

# Run server
go run ./cmd/server

# Run all tests
go test ./...

# Format check
gofmt -l .

# Run specific test
go test ./internal/service/ -run TestUpload
```

### Frontend

```bash
cd frontend

# Dev server
npm run dev

# Lint
npm run lint

# Type check
npm run typecheck

# Run tests
npm run test

# Production build (copies to backend/web/)
npm run build:backend
```

### Full Verification

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend
```

## 📄 License

[MIT](LICENSE) © ououmm
