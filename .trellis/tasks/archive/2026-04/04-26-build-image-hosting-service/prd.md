# PRD: Build OmePic Image Hosting Service

## Source

- Primary requirements source: `start.md`

## Goal

Build a full-stack image hosting service with:

- Go backend
- Next.js frontend
- SQLite persistence
- Redis permanent cache and startup preheat
- file deduplication by MD5
- user token auth for upload/delete
- JWT-protected admin area

The repository currently starts from bootstrap-only state, so this task must create the first real application code and project structure.

## Required repository structure

```text
/backend
/frontend
/README.md
/.env.example
```

Optional:

```text
/docker-compose.yml
```

## Backend requirements

### Stack

- Go
- Gin
- SQLite
- Redis

### Storage backends

Support all of:

- local filesystem storage
- S3-compatible storage
- WebDAV storage

Support runtime-managed storage instances instead of a single global storage config:

- admins can create multiple named storage instances
- multiple instances of the same backend type are allowed, including multiple S3 and multiple WebDAV entries
- one instance is marked as the default active storage for new uploads
- each image record must persist the specific storage instance identifier used for that upload, not only the backend type
- admins can edit existing storage instances
- admins can delete storage instances that are not the active default and are not referenced by existing images
- read responses must still mask secrets

### Authentication

- Upload and user delete routes must require `X-Token`.
- The frontend must generate a UUID v4 token on first visit and persist it in `localStorage`.
- Admin login must validate `ADMIN_PASSWORD` and return a JWT signed with `JWT_SECRET`.
- All `/admin/*` routes must require `Authorization: Bearer <token>`.

### Redis caching and deduplication

Implement Redis as a permanent cache layer with startup preheat.

On startup:

- load all rows from the SQLite `images` table
- populate `uid:{uid}` entries with enough JSON to resolve and serve the image
- populate `md5:{md5_hash}` with the first seen `uid` for that hash if absent

On upload:

- accept `multipart/form-data` with field name `file`
- require `X-Token`
- enforce 20 MB limit
- allow only `png`, `jpg`, `jpeg`, `gif`, `webp`, `svg`, `bmp`
- compute MD5 of file content
- if `md5:{md5}` exists, do not store the file again
- when duplicate, create a new UID and a new DB row that points at the same physical file and storage instance
- set `uid:{new_uid}` in Redis
- do not overwrite the existing `md5:{md5}` mapping
- return `duplicate: true` for deduplicated uploads

On delete:

- route: `DELETE /i/:uid`
- require matching `X-Token`
- delete physical storage only if no other DB row points to the same stored file
- delete DB row for the UID
- remove `uid:{uid}` from Redis
- remove `md5:{md5}` only if no rows remain for that hash

On serve:

- route: `GET /i/:uid`
- resolve from Redis first
- fallback to SQLite on Redis miss
- repopulate Redis on fallback
- serve correct content type
- set `Cache-Control: public, max-age=31536000, immutable`

### Required backend routes

- `POST /v1/image`
- `DELETE /i/:uid`
- `GET /i/:uid`
- `POST /admin/login`
- `GET /admin/status`
- `GET /admin/images`
- `DELETE /admin/images`
- `GET /admin/config`
- `POST /admin/config`
- `GET /health`

Admin config APIs must expose storage-instance management for the admin UI:

- list all storage instances
- create a storage instance
- update a storage instance
- delete a storage instance with validation
- switch the default active storage instance

### Backend data model

Create SQLite tables and auto-migrate on first run.

`images` table fields:

- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `uid TEXT UNIQUE NOT NULL`
- `token TEXT NOT NULL`
- `original_filename TEXT`
- `storage_key TEXT NOT NULL`
- `storage_backend TEXT DEFAULT 'local'`
- `file_path TEXT`
- `mime_type TEXT`
- `size INTEGER`
- `md5_hash TEXT NOT NULL`
- `ip_address TEXT`
- `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`

`config` table fields:

- `key TEXT PRIMARY KEY`
- `value TEXT`

Add a dedicated runtime storage catalog for instance management.

`storage_configs` table fields:

- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `storage_key TEXT UNIQUE NOT NULL`
- `name TEXT NOT NULL`
- `backend TEXT NOT NULL`
- `is_default INTEGER NOT NULL DEFAULT 0`
- backend-specific fields for local, S3, and WebDAV settings
- `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`
- `updated_at DATETIME DEFAULT CURRENT_TIMESTAMP`

### Backend middleware and runtime

- CORS
- admin auth middleware
- structured logging
- startup migration before HTTP server starts
- Redis preheat after migration and before serving traffic
- `GET /health` returns `200` only when SQLite and Redis are reachable

## Frontend requirements

### Stack

- Next.js 16.2.4 App Router
- TypeScript
- Tailwind CSS
- shadcn/ui
- Zustand
- react-hot-toast

### Pages

#### `/`

- ensure UUID token exists in `localStorage`
- upload drop zone and file input
- show upload progress
- show preview card on success
- provide copy buttons for URL, Markdown, BBCode
- persist upload history to IndexedDB
- show last 10 uploads

#### `/history`

- show all IndexedDB records
- allow delete only when current token matches record token
- send `DELETE /i/{uid}` with `X-Token`
- remove local record on success
- clear local history action

#### `/api`

- show static API documentation for upload and delete endpoints with curl examples

#### `/admin/login`

- submit password to `POST /admin/login`
- persist JWT client-side
- redirect to `/admin/dashboard` on success

#### `/admin/dashboard`

- protect all admin routes
- redirect to `/admin/login` if JWT missing or invalid
- inject `Authorization: Bearer <token>` on admin requests
- include sub-pages or sections for:
  - status
  - image management with pagination, search, multi-select, batch delete
  - settings with config load/save

## Configuration requirements

Provide `.env.example` with at least:

- bootstrap defaults for the first local storage instance and optional first S3/WebDAV instance values
- if legacy single-storage env vars are already present, startup may use them only to seed the initial default storage instance when the runtime catalog is empty
- `ADMIN_PASSWORD`
- `JWT_SECRET`
- `REDIS_URL`

Admin config endpoints must support masking secrets in read responses.

## Acceptance criteria

- The repository contains working `backend/` and `frontend/` applications.
- Backend starts, creates SQLite tables, connects to Redis, and preheats cache.
- Upload, deduplication, serve, and delete flows resolve by stored `storage_key` and work with the selected default instance.
- Admin can add, edit, switch, and delete storage instances from the dashboard.
- Multiple S3 and multiple WebDAV instances can coexist.
- Deleting an in-use or default storage instance is rejected with a clear validation error.
- Admin login and protected admin endpoints work with JWT.
- Frontend upload and history flows work against the backend contract.
- IndexedDB stores upload history records with the required fields.
- The implementation follows the bootstrap specs under `.trellis/spec/`.
- Provide basic setup and usage documentation in `README.md`.

## Delivery notes

- Prefer a coherent first end-to-end implementation over optional extras.
- If some optional infrastructure such as Docker is skipped, state that clearly.
- Keep the code organized so later tasks can extend storage backends, admin metrics, and UI behavior without major rewrites.
