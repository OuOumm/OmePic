# Type Safety

> TypeScript guidance for the current checked-in SvelteKit frontend.

---

## Current State

- The checked-in frontend uses SvelteKit, Svelte 5, Vite, and TypeScript.
- Shared frontend API/domain types live under `frontend/src/lib/types/`.
- `frontend/src/lib/types/index.ts` is the current central export for public, admin, storage, runtime settings, announcement, IP-ban, and abuse types.
- No runtime validation library is established yet.
- TypeScript remains the main compile-time contract boundary for frontend request and response shapes.

---

## Type Organization

- Keep route-local form and UI state types next to the route/component that owns them.
- Promote shared API contracts and cross-feature domain types into `frontend/src/lib/types/`.
- Keep `frontend/src/lib/api.ts` return types explicit and based on shared types.
- Keep persistence-related shapes explicit:
  - upload history record
  - admin login response
  - upload response
  - admin image list item
  - storage catalog payload
  - runtime settings payload, including admin-only AVIF encoder fields `avif_quality`, `avif_speed`, `max_image_pixels`, `avif_max_concurrency`, and `avif_conversion_timeout_seconds`
  - announcement payload
  - IP-ban and abuse payloads
- When a type mirrors a backend response, name it after the response domain rather than the page consuming it.

---

## Validation

- No runtime validation library is established yet.
- Continue using narrow manual validation at form and fetch boundaries unless a stronger need justifies adding a validation library.
- If a validation library such as Zod is introduced, use it at the boundary where untyped data enters the app:
  - environment parsing
  - server response decoding
  - admin config form submission
  - IP-ban creation form submission
- Do not mix several validation approaches for the same payload shape.

---

## Common Patterns

- Prefer discriminated unions for request states such as `idle | uploading | success | error`.
- Prefer string literal unions for storage backend choices, announcement status/priority, language, theme, and toast tone.
- Use `satisfies` and inference-friendly helpers where that improves safety without hiding the concrete shape.
- Keep date values typed consistently. If the backend returns ISO strings, keep them as strings until intentionally converted for display.
- Keep public runtime/storage option types separate from admin storage catalog types because admin types include sensitive/config-only fields.
- Keep admin `RuntimeSettings` aligned with `GET|PUT /admin/system-settings`; AVIF/image-limit fields are `avif_quality: number` (`0..100`, default `60`), `avif_speed: number` (`0..10`, default `8`), `max_image_pixels: number` (default `40000000`), `avif_max_concurrency: number` (default `2`), and `avif_conversion_timeout_seconds: number` (default `30`). Cloudflare admin runtime fields are `cloudflare_zone_id: string`, `cloudflare_api_token: string`, and `cloudflare_api_base_url: string`; the token is a masked admin-only secret on GET, and sending the unchanged masked value on PUT preserves the stored token.
- Keep public `PublicRuntimeSettings.upload` minimal and single-sourced: expose `max_upload_size_mb`, `allowed_mime_types`, and `avif_max_concurrency` only; homepage, upload queue, and native file-picker accept generation must read `upload.allowed_mime_types` instead of any duplicate effective/fallback field.
- Keep IP-ban, abuse, Cloudflare purge, and storage-health response types in the shared type module so admin images, security pages, and storage settings do not drift.

---

## Scenario: Upload Response and Public URL Contract

### 1. Scope / Trigger

- Trigger: uploads return XOR-obfuscated public UIDs plus public `.avif` URLs, and delete requests mirror that route contract.

### 2. Signatures

- `UploadResult.url: string` -> public URL ending in `.avif` or an absolute URL to that route.
- `UploadResult.duplicate: boolean` -> whether the upload reused an existing physical object.
- `deleteImageByUid(uid: string, token: string)` -> builds `/i/${uid}.avif`.
- Frontend-derived helpers:
  - `uidFromImageUrl(url): string | null`
  - `markdownForImageUrl(url, altText): string`
  - `bbcodeForImageUrl(url): string`

### 3. Contracts

- Backend upload responses stay minimal: return only `url` and `duplicate` inside the standard `{ success, data }` envelope.
- Frontend must derive canonical `uid` from `url` and synthesize `markdown` / `bbcode` locally.
- Upload may send optional multipart `storage_key`; omitting it preserves backend-default storage behavior.
- IndexedDB upload history must persist derived `uid`, `url`, `storage_key`, `storage_backend`, `markdown`, and `bbcode`.
- Successful upload MIME is handled as AVIF in the current pipeline, so the frontend may persist `mime_type: 'image/avif'` for upload-history records.
- Client accept lists must stay aligned with backend raster types and must not add SVG unless backend explicitly supports it.

### 4. Validation & Error Matrix

- Passing bare `uid` into a public route without appending `.avif` -> request contract bug.
- Accepting SVG client-side while backend rejects it -> validation drift.
- Dropping `storage_key` from history -> user cannot inspect where historical uploads landed.

### 5. Good / Base / Bad Cases

- Good: store `uid` for identity/delete and `url`/`markdown`/`bbcode` for sharing.
- Base: preview uses `result.url`; delete reconstructs `/i/${uid}.avif`.
- Bad: persist only public URL and parse UID back out ad hoc.

### 6. Tests Required

- Verify delete helper targets `/i/${uid}.avif`.
- Verify accepted file types exclude SVG in both `isAllowedImageMimeType()` and native file-picker accept generation.
- Verify history records keep `uid`, `url`, `storage_key`, `storage_backend`, `markdown`, and `bbcode`.
- Verify `safeImageUrl()` accepts same-origin image URLs and absolute image URLs whose origin matches `PublicRuntimeSettings.access.public_base_url`, while rejecting unrelated origins and unsafe schemes.

### 7. Wrong vs Correct

#### Wrong

```ts
await deleteImageByUid(record.url, token);
```

#### Correct

```ts
await deleteImageByUid(record.uid, token);
```

#### Wrong

```ts
const imageUrl = safeImageUrl(record.url);
```

#### Correct

```ts
const allowedOrigins = imageUrlAllowedOrigins(preferences.runtimeSettings?.access.public_base_url);
const imageUrl = safeImageUrl(record.url, undefined, allowedOrigins);
```

---

## Scenario: Admin Security Types

### 1. Scope / Trigger

- Trigger: admin security pages consume IP-ban and abuse-analysis endpoints.

### 2. Signatures

- `AdminImage.ip_address: string`
- `AdminIPBan`
- `AdminIPBanCreateResult`
- `AdminIPBanDeleteImagesResult`
- `AdminAbuseOverview`
- `AdminAbuseIPRankItem`
- `AdminAbuseTokenRankItem`
- `AdminAbuseIPDetail`
- API helpers:
  - `adminCreateIPBan(token, input)`
  - `adminGetIPBans(token)`
  - `adminDeleteIPBan(token, id)`
  - `adminDeleteIPBanImages(token, id)`
  - `adminGetAbuseOverview(token, params?)`
  - `adminGetAbuseIPDetail(token, ip)`

### 3. Contracts

- Admin image rows, IP-ban payloads, and abuse-analysis payloads use `ip_address` as the single IP field.
- Ban creation accepts either `uid` or `ip_address` depending on the flow.
- Abuse date filters are ISO strings passed as query params.
- Security pages should use shared types instead of defining local duplicates.

### 4. Validation & Error Matrix

- Missing both UID and IP in ban creation -> backend `invalid_input`; frontend should prevent submission.
- Empty IP detail query -> backend `invalid_input`; frontend should not call detail helper.
- Invalid abuse range -> backend `invalid_input`; frontend should show error and keep previous good data when possible.

### 5. Good / Base / Bad Cases

- Good: `ImageDataTable` passes a UID/IP into `BanIPDialog`, which calls typed `adminCreateIPBan`.
- Base: security page lists bans and abuse overview with shared response types.
- Bad: security route defines its own partial IP-ban interface and silently drops `expires_at`.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend`.
- Verify image page can create a ban from a selected image.
- Verify security page can list, delete, and clean up images for a ban.

### 7. Wrong vs Correct

#### Wrong

```ts
type Ban = { id: number; ip: string };
```

#### Correct

```ts
import type { AdminIPBan } from '@/types';
```

---

## Scenario: Admin Password Change Contract

### 1. Scope / Trigger

- Trigger: admin settings exposes password rotation backed by the backend `PUT /admin/password` endpoint.

### 2. Signatures

- API helper: `adminChangePassword(token: string, oldPassword: string, newPassword: string): Promise<void>`
- HTTP request: `PUT /admin/password`
- Body: `{ old_password: string; new_password: string }`
- Auth: `Authorization: Bearer <jwt>` via shared admin JSON headers.

### 3. Contracts

- Password fields stay as route-local Svelte state in `frontend/src/routes/admin/dashboard/settings/+page.svelte`.
- Do not add password or hash fields to shared admin system settings types.
- Do not store submitted passwords in preferences, localStorage, IndexedDB, URL params, or logs.
- Clear both password inputs after a successful update.
- Backend responses do not return the bcrypt hash; UI must not display or persist any hash.

### 4. Validation & Error Matrix

- Missing admin token -> do not call the helper.
- Empty old or new password -> settings form should keep submit disabled; backend still returns `invalid_input` for empty new password.
- Weak new password (less than 8 characters, missing uppercase, missing lowercase, or missing symbol) -> backend `invalid_input`; show existing API error toast.
- Wrong old password -> backend `forbidden` with a clear password error message; show the backend message in the existing API error toast.
- Expired/invalid JWT -> backend admin auth error; route/session handling remains unchanged.

### 5. Good / Base / Bad Cases

- Good: settings page calls `adminChangePassword(token, oldPassword, newPassword)` and clears both local strings after success.
- Base: failed password update keeps typed input so the admin can correct it.
- Bad: adding `admin_password_hash` to TypeScript types or rendering hash/status details beyond configured/default status.

### 6. Tests Required

- API helper test asserts method `PUT`, path `/admin/password`, admin bearer header, JSON content-type, and body field names.
- Run `npm run typecheck` after settings-page changes.

### 7. Wrong vs Correct

#### Wrong

```ts
localStorage.setItem('admin:new_password', newPassword);
```

#### Correct

```ts
await adminChangePassword(token, oldPassword, newPassword);
oldPassword = '';
newPassword = '';
```

---

## Forbidden Patterns

- `any` in application code.
- Broad `as` assertions to silence type errors.
- Duplicate interface definitions for the same backend payload in multiple routes/components.
- Untyped `fetch(...).json()` results flowing deep into Svelte components without a typed boundary.
- Reusing admin storage or security types for public APIs that intentionally hide sensitive fields.
