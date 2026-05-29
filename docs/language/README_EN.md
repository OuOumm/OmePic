# 🖼️ OmePic

**A self-hosted image hosting service with automatic AVIF conversion, MD5 deduplication, and multi-backend storage.**

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-2-FF3E00?logo=svelte&logoColor=white)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

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

- **Automatic AVIF conversion** — uploads are converted to AVIF with configurable quality, speed, concurrency, timeout, and pixel limit
- **Animated GIF preservation** — animated GIFs are automatically converted to animated AVIF, preserving all frames (up to 300 frames)
- **MD5 deduplication** — identical uploads reuse the existing physical file, scoped per storage instance
- **Multi-backend storage** — local filesystem, S3-compatible, and WebDAV, managed at runtime without restarts
- **Admin dashboard** — JWT-protected panel for image management, storage configuration, and system settings
- **Credential encryption** — S3/WebDAV secrets are AES-256-GCM encrypted in SQLite for defense-in-depth
- **JWT session revocation** — changing the admin password invalidates all previously issued JWTs, reducing the leak window
- **IP banning & abuse monitoring** — block abusive IPs, track upload volume by IP and token
- **Announcements** — publish time-windowed announcements with priority levels
- **Runtime configuration** — site name, upload limits, MIME allowlist, AVIF parameters, Cloudflare purge, maintenance mode, and rate limits — all editable from the admin UI
- **Token-based auth** — no user accounts; client-generated tokens (Web Crypto API, no `Math.random` fallback) identify uploaders and authorize deletes
- **Drag & drop / paste / URL upload** — flexible upload UX with upload history persisted in IndexedDB
- **Single-port deployment** — production build copies static frontend assets into `backend/web/`

## 🔒 Security Features

| Feature | Description |
|---------|-------------|
| **Enforced secret configuration** | `JWT_SECRET`, `UID_ENCRYPTION_KEY`, `SECRET_ENCRYPTION_KEY` must be ≥32 chars; server refuses to start if any is missing |
| **CORS separation** | Public APIs support CORS (allowed origins hot-updated from runtime settings); admin APIs enforce same-origin (no CORS headers) |
| **CSP compatible with SvelteKit static output** | Frontend HTML allows inline scripts/styles for SvelteKit bootstrap scripts and dynamic styling, while keeping constraints like `object-src 'none'`, `frame-ancestors 'none'`, and `base-uri 'self'` |
| **Short JWT TTL** | Admin JWT validity is 4 hours (previously 24), reducing leak window |
| **JWT revocation** | Changing password invalidates all old JWTs via Redis `admin_revoked_before` timestamp comparison |
| **Body size limit** | Upload routes set `MaxBytesReader` before multipart parsing; oversized requests are rejected at the HTTP layer |
| **Security headers** | Global `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`; frontend also gets `X-Frame-Options: DENY`; API gets `Cache-Control: no-store` |
| **Rate-limit fail-closed** | Upload and login endpoints reject requests when Redis is unavailable; normal GET requests fail-open |
| **Credential encryption** | S3 keys, WebDAV passwords and other secrets are AES-256-GCM envelope-encrypted before SQLite storage |
| **X-Token security** | Tokens generated with `crypto.randomUUID` / `crypto.getRandomValues`; no `Math.random` fallback — unsupported browsers throw an error |

## 🛠️ Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Backend | **Go** + [Gin](https://github.com/gin-gonic/gin) | HTTP API, middleware, routing |
| Database | **SQLite** (modernc.org/sqlite) | Persistent metadata & config (pure Go, no CGO) |
| Cache | **Redis** (go-redis) | UID/MD5 cache, deduplication, JWT revocation, rate limiting |
| Image | [discord/lilliput](https://github.com/discord/lilliput) + [gen2brain/avif](https://github.com/gen2brain/avif) | lilliput for animated AVIF encoding on Linux/macOS (CGO); gen2brain/avif as Windows fallback (pure Go, static only) |
| Frontend | **Svelte 5** + **SvelteKit 2** + **Tailwind CSS** | SPA with static adapter export |
| ID | Snowflake + XOR + Base62 | Opaque, URL-safe public UIDs (XOR obfuscation key, not a cryptographic boundary) |
| Auth | [golang-jwt/v5](https://github.com/golang-jwt/jwt) | Admin JWT sessions + Redis revocation |
| Encryption | AES-256-GCM (crypto/aes) | Envelope encryption for storage credentials in DB |
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
│   Middleware (Security Headers / CORS Split /    │
│   Body Limit / Auth / Rate Limit /               │
│   JWT Revocation / Logging)                      │
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
  │  (repo + │          │ WebDAV Provider  │
  │  cred    │          │ (credentials     │
  │  encrypt)│          │  encrypted at    │
  └────┬─────┘          │  rest)           │
       ▼                └──────────────────┘
  ┌──────────┐
  │  Redis   │
  │  (cache +│
  │  JWT     │
  │  revoke +│
  │  rate    │
  │  limit)  │
  └──────────┘
```

**Request flow**: Browser → Gin router (security headers → CORS split → body limit → auth/rate limit → JWT revocation check) → service layer → SQLite persistence (credential encryption) + Redis cache + storage backend write

## 🚀 Quick Start

### Prerequisites

- **Go** 1.25+
- **Node.js** 20+ (with npm)
- **Redis** 7+

### Clone

```bash
git clone https://github.com/your-username/OmePic.git
cd OmePic
```

### Environment Variables

Copy the example and fill in all required secrets:

```bash
cp .env.production.example .env
```

**Required secrets** (server refuses to start if any is missing or <32 chars):

```env
JWT_SECRET=            # JWT signing key, ≥32 characters
UID_ENCRYPTION_KEY=    # UID XOR obfuscation key (not a cryptographic boundary), ≥32 characters
SECRET_ENCRYPTION_KEY= # AES-256-GCM credential encryption key, exactly 32 bytes
```

See [Environment Variables](#-environment-variables) for the full list.

### Option 1: Direct Run

```bash
cd backend
go run ./cmd/server
```

### Option 2: Docker Compose

```bash
# Edit .env with all required secrets, then:
docker compose up -d
```

The server starts at `http://localhost:8080` after Redis health checks pass.

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

> ⚠️ The default password is auto-hashed into SQLite on first login. Change it before exposing the service publicly; high-risk updates such as storage and system settings are rejected until the default password is changed.

## 🔧 Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_ADDR` | No | `:8080` | Listen address for the HTTP server |
| `DATABASE_PATH` | No | `data/omepic.db` | Path to the SQLite database file |
| `REDIS_URL` | No | `redis://localhost:6379/0` | Redis connection URL |
| `UID_PREFIX` | No | `omeo_` | Plaintext prefix for obfuscated UIDs (trailing underscores normalized) |
| `UID_ENCRYPTION_KEY` | **Yes** | — | XOR obfuscation key for UID encoding (≥32 chars; name kept as "encryption" for deployment compat; this is NOT a cryptographic boundary) |
| `JWT_SECRET` | **Yes** | — | Secret for signing admin JWT tokens (≥32 chars; TTL 4 hours) |
| `SECRET_ENCRYPTION_KEY` | **Yes** | — | AES-256-GCM key for encrypting storage credentials in SQLite (exactly 32 bytes) |
| `PUBLIC_BASE_URL` | Production | — | Public-facing URL (required in production; server fails to start if unset) |
| `APP_ENV` | No | — | Set to `production` for strict checks; empty or `development` for relaxed mode |

> All other settings (storage, upload limits, AVIF parameters, Cloudflare purge, maintenance mode, rate limits) are managed at runtime through the admin dashboard — no environment variables needed.

## 📡 API Overview

### Public Endpoints (CORS supported; allowed origins hot-updated from runtime settings)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check (SQLite + Redis) |
| `GET` | `/v1/runtime-settings` | Public site/upload configuration |
| `GET` | `/v1/announcements` | Active published announcements |
| `POST` | `/v1/image` | Upload image (requires `X-Token`; body limited to `maxUploadSize + 1 MiB`) |
| `GET` | `/i/:uid.avif` | Serve image (returns AVIF bytes; long cache policy) |
| `DELETE` | `/i/:uid.avif` | Delete image (requires same `X-Token` as upload) |

> Storage options are returned by `GET /v1/runtime-settings` as `storage.options`.

### Admin Endpoints (strict same-origin; no CORS headers; requires JWT Bearer auth)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/admin/login` | Authenticate, returns JWT (TTL 4h) |
| `PUT` | `/admin/password` | Change password (also revokes all existing JWTs) |
| `GET` | `/admin/status` | Global upload statistics |
| `GET` | `/admin/images` | Paginated image list with search |
| `DELETE` | `/admin/images` | Batch delete images by UID |
| `GET` | `/admin/system-settings` | Get runtime + readonly settings (includes secret configuration status) |
| `PUT` | `/admin/system-settings` | Update runtime settings |
| `GET` | `/admin/config` | Get storage catalog |
| `POST` | `/admin/config/storage-instances` | Create storage instance |
| `PUT` | `/admin/config/storage-instances/:storageKey` | Update storage instance |
| `DELETE` | `/admin/config/storage-instances/:storageKey` | Delete storage instance |
| `POST` | `/admin/config/default` | Set default storage |
| `GET` | `/admin/storage/health` | Latest storage health states |
| `GET` | `/admin/storage/:key/health-history` | Storage health history |
| `POST` | `/admin/storage/:key/health-check` | Run one storage health check |
| `POST` | `/admin/storage/health-check-all` | Run health checks for all storage instances |
| `GET/POST/DELETE` | `/admin/ip-bans` | Manage IP bans |
| `GET` | `/admin/abuse/overview` | Abuse statistics |
| `GET` | `/admin/abuse/ip` | IP-specific abuse detail |
| `GET/POST/PUT/DELETE` | `/admin/announcements` | Manage announcements |
| `POST` | `/admin/cloudflare/purge-image-cache` | Purge one Cloudflare image URL cache entry |

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
- **Sensitive credentials** (S3 access/secret keys, WebDAV passwords) are AES-256-GCM encrypted in SQLite

## ⚙️ Runtime Settings

All runtime settings are managed from the admin dashboard (`/admin → Settings`) and take effect immediately — no restart required.

| Setting | Default | Description |
|---------|---------|-------------|
| Site Name | `OmePic` | Displayed in UI and page title |
| Site Tagline | `Upload, share, and manage images` | Browser title metadata |
| Public Base URL | *(required)* | Must be configured in production; overrides the public URL |
| Real IP Source | `remote-addr` | How to resolve real client IP behind proxies (`x-forwarded-for`, `x-real-ip`, `cf-connecting-ip`, or `remote-addr`) |
| Cloudflare purge | `false` | Optional single-image URL cache purge |
| Max Upload Size | `20` MB | Per-file upload limit |
| Allowed MIME Types | `image/jpeg, png, gif, webp, avif` | Accepted upload formats |
| AVIF Quality | `60` | Encoder quality (0=worst, 100=lossless) |
| AVIF Speed | `8` | Encoder speed (0=slowest/best compression, 10=fastest) |
| Max Image Pixels | `40,000,000` | Decoded pixel limit to prevent oversized images |
| AVIF Max Concurrency | `2` | Backend AVIF conversion concurrency limit |
| AVIF Conversion Timeout | `30` seconds | Per-image conversion timeout |
| Allow Storage Selection | `true` | Let uploaders pick storage target |
| Maintenance Mode | `false` | Block uploads with a custom message |
| Rate Limit | `120 req/min` | General API rate limit |
| Upload Rate Limit | `20 req/10min` | Upload-specific rate limit |

## 📂 Project Structure

```
OmePic/
├── backend/
│   ├── cmd/server/              # Entry point + enforced secret validation
│   ├── internal/
│   │   ├── auth/                # JWT generation/validation + Redis revocation check
│   │   ├── cache/               # Redis client & preheat
│   │   ├── config/              # Env config loading
│   │   ├── http/
│   │   │   ├── clientip/        # Client IP resolution (runtime hot-update)
│   │   │   ├── handler/         # HTTP handlers
│   │   │   ├── middleware/      # Security headers, CORS split, body limit, auth, rate limit
│   │   │   └── router/          # Gin route registration
│   │   ├── iputil/              # IP hashing and masking
│   │   ├── model/               # Data structures
│   │   ├── ratelimit/           # Rate limiter (Redis fail-closed/open)
│   │   ├── repository/          # SQLite data access
│   │   ├── response/            # JSON response helpers
│   │   ├── secrets/             # AES-256-GCM envelope encryption/decryption
│   │   ├── service/             # Business logic (credential encrypt-on-write/decrypt-on-read)
│   │   ├── storage/             # Local / S3 / WebDAV providers
│   │   └── uid/                 # UID encoding (Snowflake + XOR + Base62 obfuscation)
│   ├── web/                     # Production frontend assets (generated)
│   └── data/                    # Runtime data (SQLite, images)
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api.ts           # API client
│   │   │   ├── client-token.ts  # X-Token generation (Web Crypto API, no Math.random)
│   │   │   ├── components/      # UI components
│   │   │   ├── indexeddb/       # Upload history persistence
│   │   │   ├── stores/          # Svelte runes stores
│   │   │   ├── types/           # TypeScript type definitions
│   │   │   └── i18n.ts          # Internationalization
│   │   └── routes/              # SvelteKit pages
│   └── package.json
├── .github/workflows/ci.yml     # CI pipeline
├── Dockerfile                   # Multi-stage build
├── docker-compose.yml           # Docker Compose config
└── .env.production.example      # Production env example
```

## 🧑‍💻 Development

### Backend

```bash
cd backend

# Run server (requires .env with mandatory secrets)
go run ./cmd/server

# Run all tests
go test ./...

# Vet check
go vet ./...
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

### Full Verification (matches CI pipeline)

```bash
# Backend
cd backend && go vet ./... && go test ./... && go build ./...

# Frontend
cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend
```

## 🐳 Docker Deployment

```bash
# 1. Copy and edit environment variables
cp .env.production.example .env
# Must fill: JWT_SECRET, UID_ENCRYPTION_KEY, SECRET_ENCRYPTION_KEY, PUBLIC_BASE_URL

# 2. Start
docker compose up -d

# 3. View logs
docker compose logs -f omepic
```

The Dockerfile uses multi-stage builds: frontend Node.js build → Go backend compile → Debian runtime image (lilliput CGO dependencies require glibc).

## 🖥️ Platform Support

| Environment | AVIF Encoder | Animated GIF | Notes |
|-------------|-------------|--------------|-------|
| **Docker/Linux** | [discord/lilliput](https://github.com/discord/lilliput) | ✅ Supported | Recommended for production; animated GIFs auto-converted to animated AVIF |
| **macOS** | [discord/lilliput](https://github.com/discord/lilliput) | ✅ Supported | Requires lilliput C dependencies installed |
| **Windows** | [gen2brain/avif](https://github.com/gen2brain/avif) | ❌ Not supported | Static AVIF only; uploading animated GIF returns an error suggesting Docker |

> lilliput currently only supports Linux and macOS, not Windows. On Windows, the system falls back to gen2brain/avif (pure Go, no CGO) which only supports static images.

## 🙏 Acknowledgements

Thanks to the [Linux.do](https://linux.do/) community for their support and feedback.

## 📄 License

[MIT](../../LICENSE) © ououmm