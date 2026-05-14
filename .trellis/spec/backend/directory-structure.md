# Directory Structure

> Backend layout guidance for the current checked-in service tree.

---

## Current Repository Reality

- The repository contains a checked-in `backend/` tree with the package layout documented below.
- These paths are observed current structure, not speculative first-pass targets.
- Keep this doc aligned with the real package boundaries when files or responsibilities move.

---

## Current Layout

```text
backend/
|-- cmd/
|   `-- server/
|       `-- main.go
|-- web/                  # Generated frontend static export served in production
|-- internal/
|   |-- auth/
|   |-- cache/
|   |-- config/
|   |-- http/
|   |   |-- clientip/          # Trusted-proxy client IP resolver
|   |   |-- handler/
|   |   |-- middleware/
|   |   `-- router/
|   |-- iputil/                # Shared IP utility functions (hashing, masking)
|   |-- model/
|   |-- ratelimit/
|   |-- repository/
|   |-- response/
|   |-- service/
|   |-- storage/
|   `-- uid/
|-- migrations/           # Only add if SQL files become necessary
`-- go.mod
```

This layout matches the responsibilities implemented from [start.md](../../../start.md):

- Gin HTTP transport
- SQLite persistence
- Redis cache preheat and lookups
- multiple storage backends: local, S3-compatible, WebDAV
- separate admin and user authentication flows
- trusted-proxy client IP resolution
- Redis-backed API/upload rate limiting
- admin IP-ban and abuse investigation flows

---

## Module Organization

- `backend/cmd/server/main.go` owns process startup, dependency wiring, migration kickoff, Redis preheat, and HTTP server boot.
- `backend/internal/http/handler/` contains thin Gin handlers. Handlers should parse requests, call services, and write responses. They should not own deduplication, token checks, or storage branching logic.
- `backend/internal/http/router/` owns route registration and backend-owned static frontend serving. Put single-port frontend fallback logic here, not in handlers or services.
- `backend/internal/http/clientip/` owns trusted-proxy client IP resolution. Handlers and middleware must use the resolver instead of trusting raw `X-Forwarded-For` values.
- `backend/internal/iputil/` owns shared IP utility functions (SHA256 hashing, IP masking) used across services, handlers, and repositories.
- `backend/internal/http/middleware/` owns auth, request logging, and rate-limit middleware.
- `backend/internal/ratelimit/` owns Redis fixed-window limiter scripts and key semantics.
- `backend/internal/service/` contains business workflows such as upload deduplication, temporary upload-source preparation (including temp-file spooling for reader-backed uploads), delete reference counting, config updates, admin status aggregation, IP bans, and abuse statistics.
- `backend/internal/repository/` owns SQLite access. Keep SQL and transaction boundaries here or in closely related repository helpers, including the runtime-managed `storage_configs` catalog plus image/storage-key backfill logic.
- `backend/internal/cache/` owns Redis key conventions and the `ImageCache` interface (`GetImage`, `SetImage`, `DeleteImage`, `GetMD5`, `SetMD5`, `DeleteMD5`). Scoped MD5 keys (`{storage_key}:{md5_hash}`) are constructed by the caller in the service layer.
- `backend/internal/storage/` contains interchangeable storage adapters for local disk, S3-compatible APIs, and WebDAV. Providers expose both `Save([]byte)` compatibility helpers and `SaveStream(io.Reader, size, contentType)` for streaming writes from the service layer, so AVIF encoding results do not need to be materialized as one large output buffer before persistence.
- `backend/internal/auth/` contains `X-Token` validation helpers, JWT generation, and JWT verification middleware support.
- `backend/internal/response/` should centralize JSON success/error envelope helpers so handlers do not hand-roll response bodies.
- `backend/internal/model/` holds database-facing structs and DTOs shared across repository and service code.

---

## Naming Conventions

- Use lowercase Go package directories: `service`, `repository`, `storage`.
- Name files by role, not by generic utility buckets:
  - good: `image_service.go`, `admin_handler.go`, `redis_cache.go`
  - avoid: `utils.go`, `common.go`, `helpers.go`
- Keep transport-oriented files near the route they serve:
  - `backend/internal/http/handler/image_handler.go` for `POST /v1/image`, `DELETE /i/:uid.avif`, and `GET /i/:uid.avif`
  - `backend/internal/http/handler/admin_handler.go` for `/admin/*`
- Put shared middleware in `backend/internal/http/middleware/`, for example `auth_middleware.go` or `logging_middleware.go`.

---

## Current Examples

- `backend/cmd/server/main.go` wires SQLite, Redis, storage adapters, UID codec, and Gin.
- `backend/internal/service/image_service.go` implements upload deduplication, IP-ban enforcement, AVIF conversion, shared-file reference tracking, and `.avif` route normalization.
- `backend/internal/service/admin_service.go` implements admin status, image management, storage configuration, IP-ban operations, abuse statistics, and runtime settings.
- `backend/internal/http/clientip/resolver.go` implements trusted-proxy client IP resolution for uploads, deletes, rate limits, and security analytics.
- `backend/internal/repository/repository.go` isolates reads/writes to the `images`, `config`, `storage_configs`, `announcements`, and `ip_bans` tables.
- `backend/internal/ratelimit/redis_limiter.go` encapsulates Redis fixed-window rate limiting.
- `backend/internal/storage/storage.go` encapsulates local, S3-compatible, and WebDAV file operations without leaking backend-specific details into handlers.

---

## Scenario: Single-Port Frontend Static Serving

### 1. Scope / Trigger

- Trigger: Production deployments can serve the frontend through the Go backend without a live Node.js frontend server.
- Scope: Backend serves generated static files from `backend/web/` after registering API/image/admin routes.

### 2. Signatures

- Startup wiring: `router.Dependencies{FrontendDir: "web"}` from `backend/cmd/server/main.go`.
- Router helper: `registerFrontendRoutes(engine *gin.Engine, frontendDir string, logger *slog.Logger)`.
- Build artifact location: `backend/web/index.html` enables frontend serving; when missing, backend remains API-only.

### 3. Contracts

- `backend/web/` is generated output copied from the frontend static export and must stay ignored by git.
- API routes must be registered before frontend fallback.
- Frontend fallback serves only `GET` and `HEAD` browser requests.
- API paths keep API 404 behavior instead of receiving `index.html`:
  - `/health`
  - `/v1/*`
  - `/i/*`
  - `/admin/config`, `/admin/config/*`, `/admin/system-settings`
  - `/admin/ip-bans`, `/admin/ip-bans/*`, `/admin/abuse/*`
  - `/admin/announcements`, `/admin/announcements/*`.

### 4. Validation & Error Matrix

- Missing `backend/web/index.html` -> no frontend fallback; unmatched routes return Gin 404/API-only behavior.
- Missing static asset with a file extension -> `404`, never `index.html`.
- Unknown browser page path without a file extension -> frontend `index.html`.
- Unknown preserved API route/method -> JSON API-style `404`, never frontend HTML.

### 5. Good/Base/Bad Cases

- Good: `GET /history` serves the exported frontend page or app fallback.
- Base: `GET /_next/static/...` serves the static asset from `backend/web/`.
- Bad: `GET /v1/missing` or `POST /admin/login` returns frontend HTML.

### 6. Tests Required

- Router tests must assert page fallback, static asset serving, missing asset 404, API route preservation, and API method preservation.
- At minimum run `go test ./...` from `backend/` after route fallback changes.

### 7. Wrong vs Correct

#### Wrong

```go
engine.NoRoute(func(c *gin.Context) {
	c.File("web/index.html")
})
```

#### Correct

```go
engine.NoRoute(func(c *gin.Context) {
	if shouldKeepAsAPI404(c.Request.Method, c.Request.URL.Path) {
		c.JSON(http.StatusNotFound, gin.H{...})
		return
	}
	// Serve static file, exported HTML page, or index fallback for browser routes.
})
```

If the layout changes, update this file immediately instead of leaving stale bootstrap text behind.
