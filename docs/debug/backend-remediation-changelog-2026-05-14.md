# Backend Remediation Changelog

Date: 2026-05-14  
Commit: `1a1b20e`  
Commit message: `feat(upload): streamline upload pipeline and runtime warnings`

## Changed

- Upload requests now prefer `Source + DeclaredSize` instead of in-memory `[]byte` as the primary production path.
- Service now owns upload-source preparation, temporary file spooling, original-byte MD5 calculation, and temp-file cleanup.
- AVIF encoding output now streams directly into storage providers through `SaveStream()`.
- Runtime `public_base_url` now narrows allowed CORS origins when configured.
- Startup now warns when default JWT / UID secrets or first-boot admin password bootstrap are still active.
- Admin settings UI now surfaces runtime security warnings.

## Added

- `storage.Provider.SaveStream(...)` for local / S3 / WebDAV storage backends.
- `encodeAVIFToWriter(...)` for streaming AVIF encoding.
- `NewUploadInputFromBytes(...)` compatibility helper for tests / non-stream callers.
- Debug summary documents for the remediation pass.

## Removed

- Unused repository method `FindByMD5()`.

## Verified

- `cd backend && go test ./...`
- `cd frontend && npm run typecheck`
- `cd frontend && npm test -- --run src/lib/api.test.ts src/lib/ui-errors.test.ts`
- `cd frontend && npm run build:backend`
- `git diff --check`
