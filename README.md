# OmePic

OmePic is a monorepo image hosting service with a Go backend and a Next.js frontend.
This repository now includes the first real application code for:

- user-token authenticated image upload and delete
- Redis-backed MD5 deduplication with startup preheat
- local filesystem storage as the primary verified path
- S3 and WebDAV adapters behind the same storage abstraction
- JWT-protected admin status, image management, and storage config pages
- IndexedDB-backed frontend upload history

## Repository layout

```text
backend/   Go API server, SQLite persistence, Redis cache, storage adapters
frontend/  Next.js App Router UI
.env.example
```

## Backend

### Stack

- Go
- Gin
- SQLite
- Redis
- Opaque XOR/base64/base62 public UID tokens backed by snowflake SIDs
- AVIF storage output

### Main routes

- `GET /v1/storage-options`
- `POST /v1/image`
- `DELETE /i/:uid.avif`
- `GET /i/:uid.avif`
- `POST /admin/login`
- `GET /admin/status`
- `GET /admin/images`
- `DELETE /admin/images`
- `GET /admin/config`
- `POST /admin/config`
- `POST /admin/config/storage-instances`
- `PUT /admin/config/storage-instances/:storageKey`
- `DELETE /admin/config/storage-instances/:storageKey`
- `POST /admin/config/default`
- `GET /health`

### Run locally

1. Copy values from `.env.example` into your shell environment.
2. Start Redis locally.
3. Run the backend:

```powershell
cd backend
go mod tidy
go run ./cmd/server
```

If you have a stale local SQLite file from a schema that still included
`images.original_filename`, delete/reset it before restarting instead of
expecting startup migration code to repair it. In the default local workflow,
that file is `backend/data/omepic.db`.

The server listens on `http://localhost:8080` by default.

## Frontend

### Stack

- Next.js 16.2.4 App Router
- TypeScript
- Tailwind CSS
- Zustand
- react-hot-toast

### Run locally

```powershell
cd frontend
npm install
npm run dev
```

The UI listens on `http://localhost:3000` by default and talks to the backend
through `NEXT_PUBLIC_API_BASE_URL`. When that variable is omitted in
development, the frontend falls back to `http://localhost:8080`.

### Single-port production build

For production deployments that should run without a Node.js frontend server,
export the frontend and copy it into the backend-owned static directory:

```powershell
cd frontend
npm run build:backend
```

This writes the raw Next.js static export to `frontend/out/` and copies it to
`backend/web/`. Then run the backend as usual:

```powershell
cd ../backend
go build ./cmd/server
./server
```

The backend serves API routes first and then serves `backend/web/` for browser
page routes on the same port. Direct visits such as `/`, `/history`, and
`/admin/dashboard` load the exported frontend while `/health`, `/v1/*`,
`/i/*`, and backend admin API routes keep their JSON/API behavior. In the
exported production build, frontend requests use same-origin relative URLs
unless `NEXT_PUBLIC_API_BASE_URL` is explicitly set.

If `backend/web/index.html` is missing, the backend runs in API-only mode, so
the split-port development workflow remains unchanged.

## Storage behavior

- Local storage is the default backend and the main verified flow.
- Public image UIDs are opaque base62 tokens derived from the plaintext
  `normalized UID_PREFIX + shortId` after XOR encryption and base64-byte-stream
  packing. Trailing underscores in `UID_PREFIX` are normalized to a single
  separator before encryption.
- Raster uploads are converted to AVIF before persistence and public URLs end
  in `.avif`.
- The first persisted AVIF object for a unique upload uses that generated UID
  as the saved filename/key base.
- Duplicate uploads reuse the same physical file based on the MD5 of the
  original uploaded bytes only within the resolved storage instance while
  creating a new UID row and token owner record.
- `GET /v1/storage-options` returns a public-safe list of configured storage
  targets: `storage_key`, `name`, `storage_backend`, and `is_default`.
- `POST /v1/image` accepts optional multipart field `storage_key`. When omitted,
  the backend resolves the current default storage at upload time. When present,
  the selected storage instance must exist and is used for the physical write
  and deduplication scope.
- Delete requests are logical deletions: they remove the UID row, clear
  `uid:{uid}`, and repair or clear `md5:{storage_key}:{hash}` as needed.
- Physical files are retained in the request path even after the last logical
  UID is deleted and become deferred-cleanup orphaned assets for future
  maintenance.
- Redis keeps permanent `uid:{uid}` and scoped `md5:{storage_key}:{hash}` keys
  and is preheated from SQLite on startup.

## Admin behavior

- Login uses `POST /admin/login` with the configured `ADMIN_PASSWORD`.
- Admin routes require `Authorization: Bearer <jwt>`.
- `GET /admin/config` returns the runtime storage catalog and masks secrets in
  responses.
- `POST /admin/config` is a compatibility update route for the default or
  selected storage instance and can also switch `default_storage_key`.
- Storage instance CRUD uses `/admin/config/storage-instances/*`; default
  switching uses `POST /admin/config/default`.
- Deleting the default storage instance or an instance referenced by image rows
  is rejected.

## Quality checks

Backend:

```powershell
cd backend
go test ./...
go build ./cmd/server
```

Frontend:

```powershell
cd frontend
npm run lint
npm run typecheck
npm run build
npm run build:backend
```
