# Implementation Plan

## Checklist

1. Backend runtime settings model
   - Add `CloudflareZoneID`, `CloudflareAPIToken`, `CloudflareAPIBaseURL` to `RuntimeSettings` and `RuntimeSettingsUpdateInput`.
   - Add `cloudflare_zone_id`, `cloudflare_api_token`, `cloudflare_api_base_url` to `runtimeConfigFields` with empty defaults.
   - Normalize Cloudflare API base URL by trimming spaces and trailing slash.
   - Update default persistence and round-trip tests.

2. Backend validation and secret masking
   - Split/extend validation so admin `PUT /admin/system-settings` strictly rejects enabling Cloudflare purge without public base URL, zone ID, or API token.
   - Ensure startup/load does not make existing databases unbootable solely because old `cloudflare_purge_enabled=true` lacks newly introduced credentials; strict validation applies to saves.
   - In `AdminService.UpdateSystemSettings`, load current runtime settings before validation and preserve existing API token when the submitted value equals `maskSecret(current.CloudflareAPIToken)`.
   - Return masked API token from `GetSystemSettings`; never return plaintext token to frontend.
   - Add tests for masked token view, masked-token preservation, empty-token clearing, invalid base URL, and no partial save on invalid enabled config.

3. Cloudflare hot reload and multi-file purge path
   - Remove startup-time `NewCloudflareCachePurger(cfg.Cloudflare...)` injection from `cmd/server/main.go`.
   - Add dynamic runtime-backed purge behavior in `ImageService` or a runtime-aware `ImageURLCachePurger` so every purge reads current runtime settings.
   - Extend the purger seam to support multi-file `PurgeURLs(ctx, []string)` while preserving single URL manual purge via delegation.
   - Update `CloudflarePurgeConfigured()` to use current runtime settings.
   - Ensure前台用户删除、后台单张删除、后台批量删除都会触发 Cloudflare purge when enabled.
   - For admin batch delete, build all image URLs and send one Cloudflare `{ "files": [...] }` request before deleting records/caches.
   - Add/update tests proving manual purge, frontend delete purge, and admin batch delete multi-file purge use changed runtime values without service restart.

4. Startup config cleanup
   - Remove Cloudflare fields from `backend/internal/config.AppConfig` and `Load()`.
   - Update config tests so `CLOUDFLARE_*` are not part of startup environment contract.
   - Update `.env.example` to remove `CLOUDFLARE_*` startup variables.

5. Frontend admin settings + public runtime MIME cleanup
   - Extend `frontend/src/lib/types/index.ts` `RuntimeSettings` with `cloudflare_zone_id`, `cloudflare_api_token`, `cloudflare_api_base_url`.
   - Remove duplicated `effective_allowed_mime_types` from public runtime types.
   - Update homepage / upload queue to use `upload.allowed_mime_types` directly.
   - Add settings page inputs for Zone ID, API Token, and API Base URL in the Cloudflare block.
   - Update Cloudflare configured/not-configured copy to refer to runtime/admin settings instead of environment variables.
   - Update `frontend/src/lib/api.test.ts` fixtures.
   - Add en/zh i18n strings.

6. Docs/spec
   - Update `.trellis/spec/backend/runtime-settings.md` with admin-only Cloudflare runtime secret contract.
   - Update `.trellis/spec/frontend/type-safety.md` with new fields.
   - Update `docs/cloudflare-single-url-cache-purge.md`, `docs/api-reference.md`, `docs/running-and-deployment.md`.

7. Validation
   - `cd backend && go test ./...`
   - `cd frontend && npm run typecheck`
   - `cd frontend && npm test -- --run src/lib/api.test.ts`
   - `cd frontend && npm run build:backend`
   - `git diff --check`

## Risk / Rollback

- API Token becomes a SQLite admin-only secret. Mask every admin response and avoid logging full runtime payloads.
- Existing deployments with Cloudflare env vars must re-enter credentials in the admin UI by design.
- Existing DBs with `cloudflare_purge_enabled=true` and no new credentials should not block process startup; strict checks happen on admin save.
- Rollback restores startup env-driven purger and removes the new runtime fields/UI controls.
