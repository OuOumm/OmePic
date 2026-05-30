# Frontend Selectable Storage for Uploads

## Goal

Let the public upload frontend choose from configured storage instances instead of always using the backend default, while keeping admin-managed multi-storage configuration, hot updates, and backend default-storage selection intact.

## What I Already Know

- User wants the frontend to choose an already configured storage instance.
- User does not want the app limited to one storage target.
- User wants multi-storage configuration, hot updates, and backend-controlled default storage.
- Existing backend already has a runtime `storage_configs` catalog keyed by durable `storage_key`.
- Existing admin UI already supports create/update/delete/default switching for storage instances.
- Existing admin service reloads the storage manager via `storage.Reconfigure(...)` after storage config changes.
- Existing upload service always resolves new unique uploads through `storage.Current()`, so public uploads currently use only the default storage.
- Existing frontend upload request sends only multipart `file` plus `X-Token`; it does not send `storage_key`.

## Assumptions

- Public upload users may see a safe storage-option list containing only non-secret display data: `storage_key`, `name`, `storage_backend`, and `is_default`.
- Admin-only config APIs continue to expose full masked config and remain protected by JWT.
- If the frontend sends no `storage_key`, the backend uses the current default storage.
- "Hot update" means storage config changes take effect without restarting the backend, and frontend upload choices can refresh from the latest safe storage-option endpoint.

## Requirements

- Add a public safe storage-options API for the upload frontend.
- Add optional `storage_key` support to `POST /v1/image`.
- Validate any submitted `storage_key` before storing.
- Use the selected storage for new unique physical writes.
- Scope MD5 deduplication by selected storage instance: duplicate uploads reuse an existing physical object only when the matching `md5_hash` belongs to the same resolved `storage_key`.
- If the same original file is uploaded to a different selected storage instance, store a new physical object in that selected storage and create a new DB row for that storage.
- Preserve existing backend default behavior when no `storage_key` is provided.
- Keep admin default-storage switching as the source of backend default truth.
- Keep storage-manager hot reload after admin storage create/update/delete/default-switch flows.
- Add frontend upload UI control to select the target storage instance.
- Show the default storage clearly in the selector.
- Refresh storage options on page load and after user-triggered refresh.
- Persist uploaded record metadata with the resulting `storage_key`/backend if the backend returns it.
- Keep storage credentials, local paths, S3 bucket secrets, WebDAV credentials, and other sensitive config out of public responses.

## Decisions

### Deduplication Scope

Selected approach: **per-selected-storage deduplication**.

- Same file + same resolved `storage_key` -> duplicate row reuses the existing physical object for that storage.
- Same file + different resolved `storage_key` -> upload stores a new physical object in the selected storage.
- Missing `storage_key` resolves to the current backend default first; deduplication then uses that resolved default storage key.

Rationale: the user's selected storage must be honored. Global deduplication would save space but could make an upload selected for storage B reuse a file physically stored in storage A.

## Acceptance Criteria

- [ ] Public upload page can list configured storage options without admin JWT and without secret fields.
- [ ] Public upload page can select a non-default storage instance before upload.
- [ ] `POST /v1/image` accepts optional `storage_key` multipart field.
- [ ] Missing `storage_key` uses backend default storage.
- [ ] Duplicate upload with same `storage_key` reuses that storage's existing physical file.
- [ ] Duplicate upload with a different `storage_key` writes a separate physical file to the newly selected storage.
- [ ] Unknown, deleted, or invalid `storage_key` returns a clear `invalid_input`/`not_found` style error and does not write a file or DB row.
- [ ] Admin default switch takes effect for later uploads without backend restart.
- [ ] Admin storage create/update/delete/default-switch remains hot-reloaded in the storage manager.
- [ ] Upload response/history can show which storage instance handled the upload.
- [ ] Backend and frontend tests cover selected-storage upload and default fallback.
- [ ] README/API docs mention the optional upload `storage_key` field and public storage-options endpoint.

## Definition of Done

- Tests added/updated for backend selected-storage upload behavior and frontend API/types.
- `go test ./...` passes in `backend/`.
- `go build ./cmd/server` passes in `backend/`.
- `npm run lint`, `npm run typecheck`, and `npm run build` pass in `frontend/`.
- Specs/docs updated for any new cross-layer API contract.

## Out of Scope

- Adding new storage backend types beyond local/S3/WebDAV.
- Making storage credentials public.
- Reworking historical image migration.
- Live push/SSE/WebSocket config updates unless explicitly requested; manual/refetch-on-load refresh is the MVP hot-update UX.

## Technical Notes

- Relevant backend files:
  - `backend/internal/service/image_service.go`: upload currently calls `s.storage.Current()`.
  - `backend/internal/http/handler/image_handler.go`: upload currently reads only `file` and `X-Token`.
  - `backend/internal/storage/storage.go`: manager already supports `Current()`, `ForKey(storageKey)`, and `Reconfigure(settings)`.
  - `backend/internal/service/admin_service.go`: storage mutations reload the manager.
- Relevant frontend files:
  - `frontend/src/lib/api.ts`: upload XHR currently sends only `file`.
  - `frontend/src/features/upload/UploadPageClient.tsx`: upload UI has no storage selector.
  - `frontend/src/types/upload.ts`: upload response/history currently omit storage fields.
  - `frontend/src/features/admin/SettingsForm.tsx`: admin-side storage catalog already exists.
- Relevant specs:
  - `.trellis/spec/backend/database-guidelines.md`
  - `.trellis/spec/frontend/type-safety.md`
  - `.trellis/spec/frontend/state-management.md`
