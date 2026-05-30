# Design: Frontend Svelte 5 audit remediation

## Scope and boundaries

This task updates only the checked-in SvelteKit frontend and removes misleading frontend legacy artifacts. Backend contracts remain unchanged.

Primary frontend boundaries:

- `frontend/src/routes/admin/dashboard/+layout.svelte` owns admin route entry/gating.
- `frontend/src/lib/api.ts` remains the only HTTP helper layer.
- `frontend/src/lib/stores/preferences.svelte.ts` remains the shared preference/runtime/admin-token store.
- Studio components under `frontend/src/lib/components/studio/` own reusable UI behavior.
- `frontend/src/lib/utils.ts` owns reusable URL, MIME, and theme helper logic.

## Technical design

### 1. Admin authentication gate

Add explicit route-level auth state to `admin/dashboard/+layout.svelte`:

- `logged_out`: no token or invalid token; render child only for `/admin/dashboard` login route.
- `checking`: stored token exists but `adminGetStatus()` has not succeeded; render a loading panel only.
- `authenticated`: token validated; render sidebar and protected child route.

The layout should still clear invalid tokens via `clearAdminToken()` and redirect protected routes to `/admin/dashboard`. Children should not render while `checking`.

### 2. Legacy artifact cleanup

Remove obsolete Next.js artifacts:

- `frontend/next-env.d.ts`
- `frontend/eslint.config.mjs`

Remove unused old-style action:

- `frontend/src/lib/actions/click-outside.ts`

Keep `eslint.config.js` as the single active ESLint config. Remove `.next/**` ignore if no longer needed.

### 3. Announcement acknowledgement

In `AnnouncementDialog.svelte`, reset `index` to `0` whenever a new detail-mode open starts or the latest announcement changes. This makes the visible announcement match the home route acknowledgement timestamp (`announcements[0]`).

Dismissal paths remain separate from acknowledgement.

### 4. Upload accept and MIME policy

Add MIME helpers in `utils.ts`:

- normalize/filter image MIME lists
- always reject SVG
- build an `accept` string from runtime effective MIME types

Pass that accept value into `CanvasDropzone` through a new `accept` prop. File selection, drag/drop, paste, and runtime validation continue to rely on `isAllowedImageMimeType()`.

### 5. Image detail abort lifecycle

Move IP detail loading in `ImageDetailDrawer.svelte` to an effect that creates an `AbortController` per selected IP/token. Pass the signal to `adminGetAbuseIPDetail()`, abort in cleanup, and ignore abort errors. Reset loading/detail state when image becomes null.

### 6. Theme initialization and system changes

Keep `getInitialThemeScriptTheme()` as the tested decision helper. Add a helper that emits the inline theme script using the same decision branches, and update tests for script behavior where possible. Add a `matchMedia('(prefers-color-scheme: dark)')` listener in `AppShell.svelte` while `theme === 'system'` so runtime system preference changes update the class.

### 7. i18n and accessible labels

Add missing keys to both dictionaries, including `image.copyUid` and `image.url`. Replace hardcoded image detail labels where existing keys already exist.

### 8. Runtime public URL allow-list

Extend `safeImageUrl()` to accept an optional allow-list of origins. It should:

- allow relative same-origin URLs
- allow absolute URLs from current origin
- allow absolute URLs from runtime `public_base_url` origin when passed
- reject unsafe schemes and unrelated origins

Pass runtime `preferences.runtimeSettings?.access.public_base_url` from history/recent preview call sites and table thumbnail/link call sites.

## Compatibility and migration notes

- No backend endpoint shape changes.
- No localStorage key changes.
- Admin token persistence stays in `preferences.svelte.ts`.
- Existing history records remain valid; only preview URL validation changes.
- Removing Next files should not affect SvelteKit typecheck because `tsconfig.json` extends `.svelte-kit/tsconfig.json`.

## Rollback considerations

- Admin gate changes are isolated to `admin/dashboard/+layout.svelte`.
- URL/MIME helper changes are covered by existing utility tests plus new cases.
- If public URL allow-list causes unexpected preview behavior, revert `safeImageUrl()` call-site changes and keep only same-origin behavior.
