# Database Guidelines

> SQLite, Redis, and upload storage pipeline rules for the current backend implementation.

---

## Current State

- SQLite schema, Redis cache access, and startup migration code are implemented under `backend/internal/repository/` and `backend/internal/cache/`.
- SQLite remains the source of truth.
- Redis remains a permanent cache that is preheated from SQLite on startup.
- Treat the schema and flow in [start.md](../../../start.md) plus the checked-in backend code as authoritative.

---

## Source-of-Truth Rules

- SQLite stores durable image metadata and configuration.
- Runtime-managed storage instances live in SQLite `storage_configs`; new uploads must resolve through the current default instance, while existing image rows keep their own persisted `storage_key`.
- Redis mirrors lookup data needed for:
  - `GET /i/:uid`
  - storage-scoped MD5-based deduplication during `POST /v1/image`
- A request is not complete until SQLite and Redis updates are consistent enough for the chosen operation:
  - write to SQLite first
  - then populate or clear Redis keys
  - if Redis mutation fails after SQLite success, surface the error and decide whether to retry or rebuild from SQLite on the next startup

Do not make Redis the only place where image metadata exists.

---

## Query Patterns

- Repositories own database access. Gin handlers must not assemble SQL statements directly.
- Repository methods should accept `context.Context`.
- Keep multi-step write workflows in services, with repository helpers used inside transactions when necessary.
- Deduplication paths must treat `md5_hash` lookups and `uid` inserts as a race-sensitive workflow. Use a transaction or an explicit locking strategy around:
  - find existing MD5
  - decide whether a physical file write is needed
  - insert the new `uid` row
- Delete flows are logical deletes in the request path: remove the SQL row, clear `uid:{uid}`, and then repair or clear scoped `md5:{storage_key}:{hash}`.
  - Do not delete the physical asset in the online delete path, even when the deleted row was the last logical reference.
  - Treat `storage_backend + file_path` as the physical storage identity for future orphan-cleanup flows, because different backends can legitimately reuse the same object key text.
  - If the deleted row owned the cached `md5:{storage_key}:{hash}` pointer and other rows still share that hash in the same storage key, repoint the cache entry to a surviving UID instead of leaving a stale Redis pointer behind.

Current implementations (checked in):

- `backend/internal/repository/image_repository.go` — `InsertImage`, `FindByUID`, `FindByMD5`, `FindByMD5AndStorageKey`, `DeleteByUID`, `CountByMD5`, `CountByMD5AndStorageKey`, `CountByStoredFile`, `SearchImages`, `ListAllImages`, `ImageSummaryByIP`, `ListImagesByIP`, and admin-oriented image listing methods.
- `backend/internal/repository/config_repository.go` — `GetConfig`, `GetAllConfig`, `SetConfig`, `InsertMissingConfigValues`.
- `backend/internal/repository/storage_repository.go` — `GetStorageConfig`, `ListStorageConfigs`, `InsertStorageConfig`, `UpdateStorageConfig`, `DeleteStorageConfig`, `SetDefaultStorageKey`, `GetDefaultStorageKey`.
- `backend/internal/repository/ip_ban_repository.go` — `InsertIPBan`, `FindActiveIPBanByHash`, `ListActiveIPBans`, `DeleteIPBan`, `DeleteImagesByBanID`.
- `backend/internal/repository/announcement_repository.go` — `InsertAnnouncement`, `ListAnnouncements`, `UpdateAnnouncement`, `DeleteAnnouncement`, `ListCurrentAnnouncements`.
- `backend/internal/repository/helpers.go` — shared query helpers (`ensureRowsAffected`, `scanImage`, `scanImages`, `countByQuery`).

---

## Migrations

- Table creation and migration logic lives in `backend/internal/repository/migration.go` (`r.Migrate(ctx)`), which is called on startup to create tables idempotently.
- No external migration tool is used.
- In local development, do not add compatibility rebuilds for obsolete SQLite schemas just to preserve throwaway data. If a dev database predates the current schema, delete/reset it and let bootstrap recreate the tables.
  - For the default local backend workflow (`cd backend && go run ./cmd/server`), that usually means deleting `backend/data/omepic.db` before restarting.
- If raw SQL migrations are added later, keep them under `backend/migrations/` and update this spec with the exact workflow.
- Avoid ad hoc schema creation spread across handlers or services.

---

## Naming Conventions

Use the table and column names already declared in the checked-in repository:

- tables: `images`, `config`, `storage_configs`, `announcements`, `ip_bans`
- image columns: `uid`, `token`, `storage_key`, `storage_backend`, `file_path`, `mime_type`, `size`, `md5_hash`, `ip_address`, `created_at`
- IP-ban columns: `ip_hash`, `ip_address`, `ip_address_masked`, `reason`, `expires_at`, `created_at`, `updated_at`

Rules:

- Use `snake_case` for SQLite column names.
- Keep table names short and explicit.
- `images.storage_key` is the durable runtime instance identity for serve/delete/dedup follow-up flows.
- Once images reference a `storage_key`, keep that instance's backend type stable. Connection/detail edits are allowed, but switching `local`/`s3`/`webdav` under the same key must be rejected because historical `file_path` values resolve through that key.
- Keep cache keys aligned with the API model:
  - `uid:{uid}`
  - `md5:{storage_key}:{md5_hash}`
- Store `images.ip_address` using the trusted client IP resolver output, not raw request headers.
- Store IP bans with both `ip_hash` for lookup and `ip_address_masked` for UI display.
- Active IP bans are rows where `expires_at` is null/empty or later than the current time.

---

## Scenario: Public UID + AVIF Storage Pipeline

### 1. Scope / Trigger

- Trigger: image upload, serve, delete, cache preheat, and frontend URL generation now depend on one shared UID/storage contract.

### 2. Signatures

- `POST /v1/image`
- `GET /v1/storage-options`
- `GET /i/{uid}.avif`
- `DELETE /i/{uid}.avif`
- Admin delete keeps using canonical bare `uid` values inside `/admin/images`.

### 3. Contracts

- `images.uid` stores the opaque public UID token, not the decrypted
  `normalized UID_PREFIX + shortId` plaintext.
- SQLite image rows do not persist `original_filename`; keep upload filenames only in request-local response formatting or client-side IndexedDB history.
- `uid:{uid}` Redis keys use the same encrypted token stored in SQLite.
- Scoped `md5:{storage_key}:{hash}` Redis keys must point to the encrypted UID token that owns the deduplicated object for that storage instance.
- `md5_hash` is computed from the original uploaded bytes and acts as the deduplication key for the stored AVIF object.
- `POST /v1/image` may include optional multipart `storage_key`. Empty or missing values resolve to the current backend default storage at upload time; non-empty values must match an existing runtime storage instance before any file write or image row insert.
- Accepted upload source extensions are raster-only: `.avif`, `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`, and `.bmp`. AVIF source uploads are valid inputs to the same decode/re-encode pipeline; stored/served output still uses `image/avif`.
- Deduplication is scoped by the resolved `storage_key`. Same original bytes plus same storage key reuse the first physical AVIF object for that storage; same original bytes plus a different storage key must store a separate physical object in the selected storage.
- Redis MD5 mappings must use scoped keys shaped as `md5:{storage_key}:{md5_hash}` so one storage instance cannot make another selected instance reuse its physical object.
- `mime_type` for stored/served images is always `image/avif`.
- `images.storage_key` must be persisted on every row and mirrored into `uid:{uid}` cache payloads so later reads resolve the exact runtime-managed storage instance, not just the backend type.
- `GET /v1/storage-options` is public and must expose only safe display fields: `storage_key`, `name`, `storage_backend`, and `is_default`. It must not include local paths, bucket names, endpoints, access keys, passwords, or masked secret values.
- The first persisted object for a unique upload must use that row's generated encrypted UID as the filename/key base; object keys and stored paths must end in `.avif`.
- Public UID encoding must keep the token opaque at the head as well as the tail: do not XOR a payload that begins with the fixed `UID_PREFIX` bytes in a fixed secret position. Start the reversible payload with the varying short-id portion and carry any XOR-offset metadata needed for decode so different generated UIDs do not collapse into the same visible prefix.
- Environment keys:
  - `UID_PREFIX`: required contract value prefix for decrypted plaintext, default `omeo_`
    - Trailing underscores are normalized so the decrypted plaintext keeps one separator before the 8-character short id.
  - `UID_ENCRYPTION_KEY`: XOR secret material for the raw UID plaintext; if omitted locally, the process falls back to `JWT_SECRET`

### 4. Validation & Error Matrix

- Unsupported extension or non-raster payload -> `invalid_input`
- AVIF conversion failure after a valid raster decode -> `dependency_unavailable`
- Invalid encrypted UID, prefix mismatch, or missing `.avif` route suffix on public image routes -> `not_found`
- Redis miss with SQLite hit -> continue and repopulate cache

### 5. Good / Base / Bad Cases

- Good: upload PNG -> store `.avif` under the first generated encrypted UID filename/key, return encrypted `uid`, public URL ends in `.avif`
- Base: duplicate upload to the same resolved storage key -> create a new encrypted `uid` row that reuses the same first UID-based stored `.avif` object and original-bytes `md5_hash` without re-running AVIF conversion
- Base: duplicate upload to a different selected storage key -> create a new physical `.avif` object in the selected storage and persist that selected `storage_key`
- Bad: convert to AVIF before checking the original upload hash, or store the object under the original filename / MD5 hash instead of the first UID-based AVIF key
- Base: the request/service path may spool the original upload stream to a temporary file to avoid keeping the whole original image in memory, but the deduplication key must still be the MD5 of the original uploaded bytes before AVIF conversion
- Preferred service contract: production uploads should enter the service as `io.Reader` + declared size metadata, while in-memory `[]byte` inputs remain a compatibility/testing path rather than the primary request model

### 6. Tests Required

- Upload test asserts returned URL suffix, stored file suffix, stored MIME type, AVIF-decodable bytes, and that the persisted basename equals the first generated UID plus `.avif`
- Dedup test asserts same original upload bytes lead to shared `file_path` and `md5_hash`, that the duplicate path does not invoke AVIF conversion again, and that the shared `file_path` remains the first UID-based stored object
- Selected-storage dedup test asserts same original upload bytes with a different selected `storage_key` do not reuse another storage instance's `file_path`
- Public storage-options test asserts the response omits secret and path fields while marking the current default storage
- Resolve/delete tests assert bare `/i/{uid}` is rejected while `/i/{uid}.avif` succeeds
- Preheat test asserts Redis repopulates `uid:{uid}` and scoped `md5:{storage_key}:{hash}` using encrypted UIDs
- UID codec test asserts different generated UIDs do not share a stable visible head solely because `UID_PREFIX` is constant, and that decode still reconstructs the canonical `UID_PREFIX + shortId` plaintext

### 7. Wrong vs Correct

#### Wrong

- Store original PNG/JPEG bytes, or name the persisted AVIF object from the original filename / MD5 hash and only append `.avif` in the response URL
- Build the XOR payload as `UID_PREFIX + shortId` at a fixed secret offset, which makes the public token head repeat when the business prefix is constant

#### Correct

- Hash the original upload bytes first, short-circuit duplicates before transformation, then persist the first AVIF object under its generated encrypted UID key and reuse that same stored object for later duplicate rows while keeping the encrypted UID token contract consistent in SQLite, Redis, API responses, and public routes
- Keep the XOR codec reversible while making the payload head depend on changing short-id data rather than the fixed prefix, so the public UID remains opaque across generations

---

## Scenario: Admin Config Compatibility Update Route

### 1. Scope / Trigger

- Trigger: `POST /admin/config` remains as the legacy flat settings update route while storage settings now live in the runtime `storage_configs` catalog.

### 2. Signatures

- `POST /admin/config`
- Request fields:
  - `storage_key`: optional selected storage instance to patch
  - `default_storage_key`: optional storage instance to make default; when no `storage_key` is present and patch fields are present, this also selects the patched instance
  - storage patch fields such as `name`, `storage_backend`, `local_storage_path`, `s3_*`, and `webdav_*`
- Service entrypoint: `AdminService.UpdateConfig(ctx, AdminConfigUpdateInput)`

### 3. Contracts

- Patch fields update one storage instance, not the whole catalog.
- If `storage_key` is present, patch that instance.
- If `storage_key` is absent and `default_storage_key` is present with patch fields, patch the `default_storage_key` instance and then make it default.
- If both keys are absent but patch fields are present, patch the current default instance.
- When `default_storage_key` is present, validate that it is non-empty and exists before applying any patch. Do not save a partial storage update and then fail the default switch.
- Secret fields may be omitted or sent as the exact masked value from `GET /admin/config`; in both cases the stored secret must be preserved.

### 4. Validation & Error Matrix

- Empty or whitespace `default_storage_key` -> `invalid_input`, no config write
- Unknown `default_storage_key` -> `not_found`, no config write
- Unknown `storage_key` patch target -> `not_found`, no config write
- Backend type change for an in-use storage instance -> `conflict`, no config write
- Invalid backend-specific settings -> `invalid_input`, no config write
- SQLite/storage-manager failure after a valid update attempt -> `dependency_unavailable`

### 5. Good / Base / Bad Cases

- Good: request includes `storage_key`, patch fields, and a valid different `default_storage_key`; the patch is applied, then the default instance switches.
- Base: legacy request sends only flat storage fields; the current default instance is patched.
- Bad: request patches `storage_key = local-default` and also sends `default_storage_key = missing`; the patch must not be saved before returning `not_found`.

### 6. Tests Required

- Patch-only request updates the current default instance and reloads the storage manager.
- Default-only request switches the default instance and reloads the storage manager.
- Patch plus missing `default_storage_key` returns `not_found` and leaves the patched instance unchanged.
- Patch plus empty `default_storage_key` returns `invalid_input` and leaves the patched instance unchanged.
- Backend type changes for in-use instances stay rejected.

### 7. Wrong vs Correct

#### Wrong

- Apply the storage patch first, then validate `default_storage_key`; a bad default key leaves a saved config mutation behind even though the request failed.

#### Correct

- Validate any requested default target before mutating the selected storage instance, then apply the patch and default switch in that order.

---

## Scenario: Original Filename Stays Client-Local

### 1. Scope / Trigger

- Trigger: upload UX still needs the browser-side filename, but SQLite/admin search must not persist or query it.

### 2. Signatures

- SQLite `images` table excludes `original_filename`
- `POST /v1/image` upload request still receives multipart filename metadata
- IndexedDB history record keeps `{ uid, url, md_url, bbcode, token, original_filename, size, mime_type, created_at }`

### 3. Contracts

- Repository insert/select/search code must ignore `original_filename`.
- Upload response formatting may still use the request-local filename for Markdown alt text.
- Admin image list/search responses do not expose or filter by `original_filename`.

### 4. Validation & Error Matrix

- Admin search string matches only UID/token/IP/MD5 -> valid query
- Local dev SQLite file still contains `original_filename` column -> delete/reset the stale dev database instead of adding repository compatibility code
  - Default local path: `backend/data/omepic.db` when the server is launched from `backend/`

### 5. Good / Base / Bad Cases

- Good: upload preserves client-side filename in IndexedDB history while SQLite stores only UID/token/storage metadata
- Base: duplicate upload returns Markdown built from the current request filename without persisting it; a stale dev DB gets recreated from the current schema
- Bad: reintroduce `original_filename` to repository SQL or admin search because a UI table wants a display label

### 6. Tests Required

- Repository schema test asserts newly bootstrapped `images` omits `original_filename`
- Repository search test asserts queries only match UID/token/IP/MD5-backed columns, not client-local filenames
- Upload service test asserts Markdown output still uses the request filename

### 7. Wrong vs Correct

#### Wrong

- Persist browser-provided filenames in SQLite just to populate admin tables or search

#### Correct

- Keep filenames only in request-local/frontend-local contexts and use UID/token/IP/MD5 for durable SQLite/admin workflows

---

---

## Scenario: AVIF Stream Conversion Contract

### 1. Scope / Trigger

- Trigger: `saveConvertedAVIF` uses concurrent goroutines to encode and persist AVIF in a streaming pipeline.
- This is a correctness contract: if either side fails, the upload must fail quickly without hanging.

### 2. Signatures

- `ImageService.saveConvertedAVIF(ctx, provider, objectKey, source, settings) (int64, string, error)`
- `ImageService.encoder` field: `func(io.Reader, io.Writer, AVIFConversionSettings) error` — injectable for testing.
- `AVIFConversionSettings{Quality int; Speed int}` — conversion options derived from runtime settings.

### 3. Contracts

- AVIF encoding writes to a pipe writer; storage `SaveStream` reads from the pipe reader. Both run in separate goroutines coordinated by channels.
- The main goroutine waits for `SaveStream` result first. If `SaveStream` fails, it immediately closes the pipe reader and signals the writer with `CloseWithError`, which unblocks the encoding goroutine if it is still writing.
- After signaling the writer, the main goroutine waits for the encode channel result.
- If the encoding goroutine fails, it must close the pipe writer with `CloseWithError` so the `SaveStream` reader sees EOF/error and terminates.
- Error priority: encoding errors take precedence over save errors because a decode/encode failure is a user-facing `invalid_input` while a save failure is `dependency_unavailable`.
- When both sides fail independently, the encoding error is returned if it is a direct encoding failure; otherwise the save error is wrapped as `dependency_unavailable`.
- No goroutine may leak or hang after `saveConvertedAVIF` returns. Both goroutines must terminate regardless of which side failed first.

### 4. Validation & Error Matrix

- `SaveStream` fails before reading any data -> pipe closed, encode goroutine sees `ClosedWithError`, returns `dependency_unavailable`.
- `SaveStream` fails after reading partial data -> same pipe-close sequence, returns `dependency_unavailable`.
- Encoder fails (bad image data, unsupported format) -> pipe writer closed with `CloseWithError`, `SaveStream` reader sees EOF error, returns `invalid_input`.
- Both succeed -> returns `(counting.size, storedPath, nil)`.

### 5. Good/Base/Bad Cases

- Good: upload with a valid image -> encode and save both succeed, returns size and stored path.
- Good: storage backend returns error -> upload returns quickly with `dependency_unavailable`.
- Good: corrupt image payload -> encode fails with `invalid_input`, pipe signaled, `SaveStream` terminates.
- Bad: `saveConvertedAVIF` returns after timeout instead of immediately on save failure (indicates a missing pipe-close).
- Bad: goroutine leak after upload failure (pipe not properly closed on one side).

### 6. Tests Required

- `TestUploadReturnsQuicklyWhenSaveStreamFailsImmediately` — verifies upload returns error within 2s and no image row is persisted when `SaveStream` fails before reading.
- `TestUploadReturnsQuicklyWhenSaveStreamFailsAfterPartialRead` — verifies upload returns error within 2s and `readCalled` is true when `SaveStream` reads partial data then fails.
- Both tests assert `select` with a 2-second timeout to detect hangs.

### 7. Wrong vs Correct

#### Wrong

```go
// SaveStream result waited after encode completes — encode may hang
// if SaveStream never consumes the pipe.
go func() {
    storedPath, err := provider.SaveStream(ctx, objectKey, pipeReader, -1, mime)
    saveResultCh <- saveResult{storedPath: storedPath, err: err}
}()
// No pipe-close on save failure -> encode goroutine blocks forever
saveResult := <-saveResultCh
if saveResult.err != nil {
    return 0, "", saveResult.err
}
encodeErr := <-encodeErrCh
```

#### Correct

```go
saveResult := <-saveResultCh
if saveResult.err != nil {
    _ = pipeReader.Close()
    _ = pipeWriter.CloseWithError(saveResult.err)
}
encodeErr := <-encodeErrCh
_ = pipeReader.Close()
```

---

## Common Mistakes To Avoid

- Letting handlers talk to SQLite or Redis directly.
- Writing a file to storage before the deduplication check has finished.
- Resolving historical files only by `storage_backend` after the app has moved to multiple named storage instances of the same backend.
- Mutating the backend type for an in-use `storage_key`, which makes historical rows keep the same durable instance key but resolve against a different provider implementation.
- Deleting a physical file in the online request path instead of deferring orphan cleanup.
- Counting cross-backend rows as the same physical file just because they share the same `file_path` string when building future cleanup flows.
- Leaving `md5:{storage_key}:{hash}` mapped to a deleted UID when duplicate rows still exist in that storage key.
- Allowing schema names in code to drift from the names documented in [start.md](../../../start.md).
- Introducing a second source of truth for config values outside SQLite and the environment without documenting precedence.
