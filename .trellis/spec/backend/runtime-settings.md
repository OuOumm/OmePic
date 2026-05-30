# Runtime Settings Guidelines

> Cross-layer contracts for runtime-managed site metadata, upload policy, and public/admin settings views.

---

## Current State

- Runtime settings are owned by `backend/internal/service/runtime_settings.go` and loaded from the SQLite `config` key/value table through `RuntimeSettingsManager`.
- Public runtime settings are exposed through `GET /v1/runtime-settings`; admin runtime settings are exposed through `GET /admin/system-settings` and updated through `PUT /admin/system-settings`.
- The Svelte frontend consumes these contracts in `frontend/src/lib/types/index.ts`, `frontend/src/lib/api.ts`, `frontend/src/routes/+page.svelte`, and `frontend/src/routes/admin/dashboard/settings/+page.svelte`.

---

## Scenario: Runtime Upload Policy And Site Metadata

### 1. Scope / Trigger

- Trigger: a change touches site name/tagline, upload size, allowed MIME types, AVIF conversion quality/speed/concurrency/timeout, image pixel limits, Cloudflare purge runtime credentials, public runtime settings, admin system settings, or upload validation.
- This is cross-layer because the same setting is persisted in SQLite, normalized in Go, returned by admin/public APIs, typed in TypeScript, rendered in settings UI, and enforced during upload.

### 2. Signatures

- Backend structs:
  - `RuntimeSettings{SiteName, SiteTagline, PublicBaseURL, CloudflarePurgeEnabled, CloudflareZoneID, CloudflareAPIToken, CloudflareAPIBaseURL, MaxUploadSizeMB, AllowedMIMETypes, AvifQuality, AvifSpeed, MaxImagePixels, AVIFMaxConcurrency, AVIFConversionTimeoutSeconds, RealIPSource, ...}`
  - `RuntimeSettingsUpdateInput{site_name, site_tagline, public_base_url, cloudflare_purge_enabled, cloudflare_zone_id, cloudflare_api_token, cloudflare_api_base_url, max_upload_size_mb, allowed_mime_types, avif_quality, avif_speed, max_image_pixels, avif_max_concurrency, avif_conversion_timeout_seconds, real_ip_source, ...}`
  - `PublicRuntimeSettingsView{site, upload, features, storage}`
  - `AdminSystemSettingsView{runtime, readonly}`
- Backend service methods:
  - `RuntimeSettingsManager.Load(ctx, repo)`
  - `RuntimeSettingsManager.Current() RuntimeSettings`
  - `RuntimeSettingsManager.Reconfigure(settings)`
- HTTP APIs:
  - `GET /v1/runtime-settings`
  - `GET /admin/system-settings`
  - `PUT /admin/system-settings`
  - `POST /v1/image`
- Frontend helpers/types:
  - `getRuntimeSettings(): Promise<PublicRuntimeSettings>`
  - `adminGetSystemSettings(token): Promise<AdminSystemSettings>`
  - `adminUpdateSystemSettings(token, runtime): Promise<AdminSystemSettings>`
  - `RuntimeSettings.allowed_mime_types: string[]`
  - `RuntimeSettings.avif_quality: number`
  - `RuntimeSettings.avif_speed: number`
  - `RuntimeSettings.cloudflare_api_token: string` is admin-only and must be masked in admin GET responses.

### 3. Contracts

- Persisted config keys include:
  - `site_name`: non-empty after normalization; defaults to `OmePic`
  - `site_tagline`: non-empty after normalization; defaults to `上传、分享和管理图片`
  - `public_base_url`: configured public base URL; empty means request-host fallback
  - `max_upload_size_mb`: integer megabytes; default is `20`
  - `allowed_mime_types`: comma-separated image MIME values stored as the actual configured allow-list
  - `avif_quality`: integer AVIF encoder quality, default `60`, valid `0..100` (`100` means lossless)
  - `avif_speed`: integer AVIF encoder speed, default `8`, valid `0..10` (lower is usually slower with better compression/quality trade-offs)
  - `max_image_pixels`: integer decoded pixel limit, default `40000000`, must be positive
  - `avif_max_concurrency`: integer AVIF conversion concurrency limit, default `2`, must be positive; public runtime `upload.avif_max_concurrency` exposes this value for frontend queue guidance
  - `avif_conversion_timeout_seconds`: integer per-image conversion timeout, default `30`, must be positive
  - `cloudflare_purge_enabled`: bool, default `false`
  - `cloudflare_zone_id`: admin-managed Cloudflare Zone ID, default empty
  - `cloudflare_api_token`: admin-only Cloudflare API Token, default empty; SQLite stores the real value, admin responses return only a masked value or empty string
  - `cloudflare_api_base_url`: optional Cloudflare API base URL, default empty; empty means `https://api.cloudflare.com/client/v4`
  - `real_ip_source`: client IP source for upload/delete/rate-limit/IP-ban/abuse flows; default `remote-addr`; valid values are `remote-addr`, `x-forwarded-for`, `x-real-ip`, and `cf-connecting-ip`
  - `allow_storage_selection`, `maintenance_mode`, `maintenance_message`, `rate_limit_window_minutes`, `rate_limit_max_requests`, `upload_rate_limit_window_minutes`, `upload_rate_limit_max_requests`
- `RuntimeSettingsManager.Load(ctx, repo)` must persist missing default runtime keys to SQLite with insert-missing semantics before loading settings, so first run has durable defaults without overwriting existing admin changes.
- `PUBLIC_BASE_URL` may be read at startup only as a bootstrap seed into `RuntimeSettings.PublicBaseURL` when the runtime setting is still empty. After startup, `RuntimeSettings.PublicBaseURL` is the only configured public URL source. `AdminEnvironmentStatus` exposes only `public_base_url_source` plus `runtime_public_base_url_set` for that state.
- In `APP_ENV=production`, `public_base_url` must be configured at startup and admin runtime updates must not clear it; production upload URLs must not fall back to the request Host header.
- `AdminSystemSettingsView.readonly.security` must expose `configured` status for `jwt_secret`, `admin_password`, and `uid_encryption_key` so the admin settings UI can warn about missing required startup secrets without exposing any secret material.
- `allowed_mime_types` must not be treated as a hidden backend fallback. The admin input should display the runtime field directly, and upload validation must use the configured runtime list.
- `image/jpg` is accepted as an admin input alias and normalized to `image/jpeg` before persistence and API response.
- SVG is not allowed in this upload pipeline even though it is an `image/*` MIME type.
- `GET /v1/runtime-settings.site.name` drives visible site branding.
- `GET /v1/runtime-settings.site.tagline` is only browser-title metadata. On the homepage title, render `site.name - site.tagline` when tagline is present; do not use the tagline as the upload dropzone subtitle.
- Cloudflare image URL purge credentials are runtime admin settings, not startup environment variables. `CLOUDFLARE_ZONE_ID`, `CLOUDFLARE_API_TOKEN`, and `CLOUDFLARE_API_BASE_URL` must not be read by `config.Load()`.
- Cloudflare purge requests must use Cloudflare's `files` array. Single URL purge is supported, and admin batch deletion should combine the deleted image URLs into one `{ "files": [...] }` request when Cloudflare purge is enabled; do not implement `purge_everything` or prefix purge.
- `GET /admin/system-settings.runtime.cloudflare_api_token` must return `""` for no token or a masked value using `maskSecret` semantics; never return plaintext.
- `PUT /admin/system-settings` must preserve the current Cloudflare API token when the submitted value equals the current masked token, clear it when submitted as empty, and save any other submitted value as a new token.
- `cloudflare_api_base_url` must be normalized by trimming spaces and trailing `/`; non-empty values must be valid `http` or `https` URLs.
- `cloudflare_purge_enabled=true` requires a non-empty valid `public_base_url`, non-empty `cloudflare_zone_id`, and non-empty `cloudflare_api_token`; validation failures return `invalid_input` and must not partially save runtime settings.
- `RuntimeSettingsManager.Load` may load existing databases with `cloudflare_purge_enabled=true` and missing new Cloudflare credential keys so upgrades remain bootable; strict credential validation applies to admin saves.
- Upload validation must check MIME through `runtimeSettingsAllowsMIME(settings, input.MIMEType)` and must not maintain a separate extension allow-list that can drift from `allowed_mime_types`.
- `real_ip_source` must stay admin-only (not exposed in `GET /v1/runtime-settings`) and must hot-reload through `clientip.Resolver` so IP bans/rate limits observe settings changes without restarting.
- New physical uploads must pass the current runtime `avif_quality`, `avif_speed`, `avif_max_concurrency`, and `avif_conversion_timeout_seconds` into the AVIF conversion path after checking `max_image_pixels`; duplicate uploads that hit original-byte MD5 deduplication must reuse the existing physical object and skip conversion.
- Frontend settings UI may edit MIME types as a comma-separated string, but it must send `allowed_mime_types` as a string array in the admin update request.

### 4. Validation & Error Matrix

- Empty `site_name` -> normalize to default site name, not an empty API field.
- Empty `site_tagline` -> normalize to default tagline, not an empty API field.
- Missing runtime settings config keys during bootstrap/load -> insert default values with `ON CONFLICT DO NOTHING`; do not overwrite existing values.
- Missing `allowed_mime_types` config key during bootstrap/load -> write the default configured list through the same missing-key persistence path; do not defer to upload-time fallback.
- Missing `avif_quality` / `avif_speed` / `max_image_pixels` / `avif_max_concurrency` / `avif_conversion_timeout_seconds` / `real_ip_source` config keys during bootstrap/load -> write defaults `60` / `8` / `40000000` / `2` / `30` / `remote-addr` through insert-missing semantics; do not overwrite existing admin values.
- Admin save with invalid `real_ip_source` -> return `invalid_input` and do not update runtime settings.
- `APP_ENV=production` startup or admin save with empty `public_base_url` -> return/fail with `invalid_input` and do not partially save other runtime settings.
- `cloudflare_purge_enabled=true` with missing `public_base_url`, `cloudflare_zone_id`, or `cloudflare_api_token` on admin save -> return `invalid_input` and do not update runtime settings.
- `cloudflare_api_base_url` non-empty but not `http`/`https` URL -> return `invalid_input` and do not update runtime settings.
- Submitted `cloudflare_api_token` equals the current masked token -> keep stored plaintext token unchanged.
- Submitted `cloudflare_api_token` is empty -> clear stored token.
- `avif_quality < 0` or `avif_quality > 100` -> return `invalid_input` and do not update runtime settings.
- `avif_speed < 0` or `avif_speed > 10` -> return `invalid_input` and do not update runtime settings.
- `max_image_pixels <= 0`, `avif_max_concurrency <= 0`, or `avif_conversion_timeout_seconds <= 0` -> return `invalid_input` and do not update runtime settings.
- `allowed_mime_types` contains `image/jpg` -> normalize to `image/jpeg`.
- `allowed_mime_types` contains non-`image/*`, whitespace/semicolon, or `image/svg+xml` -> return `invalid_input` and do not update runtime settings.
- Upload MIME not in `AllowedMIMETypes` -> reject `POST /v1/image` with `invalid_input` / file MIME type not allowed.
- Frontend receives `allowed_mime_types: null` from an older backend -> guard with `Array.isArray` before joining to avoid runtime crashes.
- Public runtime settings must expose only one upload MIME list field: `upload.allowed_mime_types`; do not reintroduce duplicate `effective_allowed_mime_types` into the public API.
- Public runtime settings upload payload includes `max_upload_size_mb`, `allowed_mime_types`, and `avif_max_concurrency` only; admin-only quality/speed/pixel/timeout fields stay under `/admin/system-settings.runtime`.

### 5. Good/Base/Bad Cases

- Good: admin sees `image/avif, image/gif, image/jpeg, image/png, image/webp` directly in the allowed MIME input; saving the same list keeps upload validation aligned with the UI.
- Good: homepage, upload queue, and file-picker accept list all read the same `upload.allowed_mime_types` payload from `GET /v1/runtime-settings`.
- Good: admin enters `image/jpg, image/png`; API response returns `image/jpeg, image/png`, and uploads with `image/jpeg` pass.
- Base: public homepage title becomes `OmePic - Custom subtitle`, while the dropzone keeps the localized upload helper text.
- Bad: backend silently allows `.bmp` because an extension map includes it while `allowed_mime_types` does not.
- Bad: admin UI displays a second fallback MIME field while sending a different or empty `allowed_mime_types` payload.

### 6. Tests Required

- Backend tests:
  - `RuntimeSettingsManager.Load` persists every default runtime setting key on an empty config table
  - missing-key persistence does not overwrite existing `site_name`, `site_tagline`, `public_base_url`, or other admin-configured values
  - default runtime settings include `max_upload_size_mb = 20`, `avif_quality = 60`, `avif_speed = 8`, `max_image_pixels = 40000000`, `avif_max_concurrency = 2`, `avif_conversion_timeout_seconds = 30`, `real_ip_source = remote-addr`, empty Cloudflare credential fields, and a non-empty configured `allowed_mime_types` list
  - admin system settings rejects clearing `public_base_url` in production without partially saving other runtime fields
  - invalid `real_ip_source` is rejected without partial settings saves
  - admin system settings masks `cloudflare_api_token`, preserves it when the masked value is submitted back, clears it when empty, rejects invalid `cloudflare_api_base_url`, and rejects enabled purge without public base URL / Zone ID / API Token without partial saves
  - AVIF quality/speed/concurrency/timeout and pixel-limit validation rejects out-of-range values without partial settings saves
  - upload passes configured AVIF quality/speed/concurrency/timeout into new physical conversions while duplicate uploads skip conversion
  - upload rejects images whose decoded dimensions exceed `max_image_pixels` before AVIF persistence
  - MIME normalization converts `image/jpg` to `image/jpeg`, sorts/deduplicates values, and rejects SVG
  - upload rejects MIME types absent from runtime settings even when the filename extension looks image-like
  - upload accepts a configured MIME regardless of filename extension allow-list assumptions
- Frontend checks:
  - `npm run lint`
  - `npm run typecheck`
  - `npm run build:backend`
- Frontend assertions:
  - settings page joins `runtime.allowed_mime_types` with commas and guards `null` with `Array.isArray`
  - homepage title uses site name plus tagline, and dropzone subtitle does not consume site tagline

### 7. Wrong vs Correct

#### Wrong

```go
var allowedExtensions = map[string]struct{}{
	".jpg": {},
	".png": {},
}

if _, ok := allowedExtensions[strings.ToLower(filepath.Ext(input.OriginalFilename))]; !ok {
	return UploadOutput{}, ErrInvalidInput
}
```

#### Correct

```go
if !runtimeSettingsAllowsMIME(runtimeSettings, input.MIMEType) {
	return UploadOutput{}, fmt.Errorf("%w: file MIME type is not allowed", ErrInvalidInput)
}
```

#### Wrong

```svelte
<CanvasDropzone subtitle={preferences.runtimeSettings?.site.tagline} />
```

#### Correct

```svelte
<svelte:head><title>{siteTitle}</title></svelte:head>
<CanvasDropzone language={preferences.language} />
```
