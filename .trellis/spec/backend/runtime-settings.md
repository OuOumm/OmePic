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

- Trigger: a change touches site name/tagline, upload size, allowed MIME types, AVIF conversion quality/speed, public runtime settings, admin system settings, or upload validation.
- This is cross-layer because the same setting is persisted in SQLite, normalized in Go, returned by admin/public APIs, typed in TypeScript, rendered in settings UI, and enforced during upload.

### 2. Signatures

- Backend structs:
  - `RuntimeSettings{SiteName, SiteTagline, PublicBaseURL, MaxUploadSizeMB, AllowedMIMETypes, AvifQuality, AvifSpeed, ...}`
  - `RuntimeSettingsUpdateInput{site_name, site_tagline, public_base_url, max_upload_size_mb, allowed_mime_types, avif_quality, avif_speed, ...}`
  - `PublicRuntimeSettingsView{site, upload, features, storage}`
  - `AdminSystemSettingsView{runtime, readonly}`
- Backend service methods:
  - `RuntimeSettingsManager.Load(ctx, repo)`
  - `RuntimeSettingsManager.Current() RuntimeSettings`
  - `RuntimeSettingsManager.Reconfigure(settings)`
  - `RuntimeSettings.EffectiveAllowedMIMETypes() []string`
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
  - `PublicRuntimeSettings.upload.effective_allowed_mime_types: string[]`

### 3. Contracts

- Persisted config keys include:
  - `site_name`: non-empty after normalization; defaults to `OmePic`
  - `site_tagline`: non-empty after normalization; defaults to `Upload, share, and manage images`
  - `public_base_url`: configured public base URL; empty means request-host fallback
  - `max_upload_size_mb`: integer megabytes; default is `20`
  - `allowed_mime_types`: comma-separated image MIME values stored as the actual configured allow-list
  - `avif_quality`: integer AVIF encoder quality, default `60`, valid `0..100` (`100` means lossless)
  - `avif_speed`: integer AVIF encoder speed, default `8`, valid `0..10` (lower is usually slower with better compression/quality trade-offs)
  - `allow_storage_selection`, `maintenance_mode`, `maintenance_message`, `rate_limit_window_minutes`, `rate_limit_max_requests`, `upload_rate_limit_window_minutes`, `upload_rate_limit_max_requests`
- `RuntimeSettingsManager.Load(ctx, repo)` must persist missing default runtime keys to SQLite with insert-missing semantics before loading settings, so first run has durable defaults without overwriting existing admin changes.
- `PUBLIC_BASE_URL` is not an environment variable. `RuntimeSettings.PublicBaseURL` is the only configured public URL source, and `AdminEnvironmentStatus` exposes only `public_base_url_source` plus `runtime_public_base_url_set` for that state.
- `AdminSystemSettingsView.readonly.security` must expose `configured` / `using_default` status for `jwt_secret`, `admin_password`, and `uid_encryption_key` so the admin settings UI can warn about insecure default bootstrap values without exposing any secret material.
- `allowed_mime_types` must not be treated as a hidden backend fallback. The admin input should display the runtime field directly, and upload validation must use the configured runtime list.
- `image/jpg` is accepted as an admin input alias and normalized to `image/jpeg` before persistence and API response.
- SVG is not allowed in this upload pipeline even though it is an `image/*` MIME type.
- `GET /v1/runtime-settings.site.name` drives visible site branding.
- `GET /v1/runtime-settings.site.tagline` is only browser-title metadata. On the homepage title, render `site.name - site.tagline` when tagline is present; do not use the tagline as the upload dropzone subtitle.
- Upload validation must check MIME through `runtimeSettingsAllowsMIME(settings, input.MIMEType)` and must not maintain a separate extension allow-list that can drift from `allowed_mime_types`.
- New physical uploads must pass the current runtime `avif_quality` and `avif_speed` into AVIF conversion; duplicate uploads that hit original-byte MD5 deduplication must reuse the existing physical object and skip conversion.
- Frontend settings UI may edit MIME types as a comma-separated string, but it must send `allowed_mime_types` as a string array in the admin update request.

### 4. Validation & Error Matrix

- Empty `site_name` -> normalize to default site name, not an empty API field.
- Empty `site_tagline` -> normalize to default tagline, not an empty API field.
- Missing runtime settings config keys during bootstrap/load -> insert default values with `ON CONFLICT DO NOTHING`; do not overwrite existing values.
- Missing `allowed_mime_types` config key during bootstrap/load -> write the default configured list through the same missing-key persistence path; do not defer to upload-time fallback.
- Missing `avif_quality` / `avif_speed` config keys during bootstrap/load -> write defaults `60` / `8` through insert-missing semantics; do not overwrite existing admin values.
- `avif_quality < 0` or `avif_quality > 100` -> return `invalid_input` and do not update runtime settings.
- `avif_speed < 0` or `avif_speed > 10` -> return `invalid_input` and do not update runtime settings.
- `allowed_mime_types` contains `image/jpg` -> normalize to `image/jpeg`.
- `allowed_mime_types` contains non-`image/*`, whitespace/semicolon, or `image/svg+xml` -> return `invalid_input` and do not update runtime settings.
- Upload MIME not in `AllowedMIMETypes` -> reject `POST /v1/image` with `invalid_input` / file MIME type not allowed.
- Frontend receives `allowed_mime_types: null` from an older backend -> guard with `Array.isArray` before joining to avoid runtime crashes.

### 5. Good/Base/Bad Cases

- Good: admin sees `image/avif, image/gif, image/jpeg, image/png, image/webp` directly in the allowed MIME input; saving the same list keeps upload validation aligned with the UI.
- Good: admin enters `image/jpg, image/png`; API response returns `image/jpeg, image/png`, and uploads with `image/jpeg` pass.
- Base: public homepage title becomes `OmePic - Custom subtitle`, while the dropzone keeps the localized upload helper text.
- Bad: backend silently allows `.bmp` because an extension map includes it while `allowed_mime_types` does not.
- Bad: admin UI displays `effective_allowed_mime_types` fallback while sending a different or empty `allowed_mime_types` payload.

### 6. Tests Required

- Backend tests:
  - `RuntimeSettingsManager.Load` persists every default runtime setting key on an empty config table
  - missing-key persistence does not overwrite existing `site_name`, `site_tagline`, `public_base_url`, or other admin-configured values
  - default runtime settings include `max_upload_size_mb = 20`, `avif_quality = 60`, `avif_speed = 8`, and a non-empty configured `allowed_mime_types` list
  - AVIF quality/speed validation rejects out-of-range values without partial settings saves
  - upload passes configured AVIF quality/speed into new physical conversions while duplicate uploads skip conversion
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
