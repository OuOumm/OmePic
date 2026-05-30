# State Management

> State rules for the current checked-in SvelteKit frontend.

---

## Current State

- Shared state exists under `frontend/src/lib/stores/`.
- The active frontend uses Svelte 5 runes (`$state`) rather than Zustand.
- Current global stores:
  - `frontend/src/lib/stores/preferences.svelte.ts`
  - `frontend/src/lib/stores/toast.svelte.ts`
  - `frontend/src/lib/stores/upload-queue.svelte.ts`
- Browser persistence uses localStorage for UI preferences, selected upload storage, and admin token.
- Upload history uses IndexedDB through `frontend/src/lib/indexeddb/upload-history.ts`.

---

## State Categories

- **Component-local UI state**: drag-over state, dialogs, forms, selected rows, current preview image, loading flags. Keep this inside the route/component that owns it.
- **Shared client state**: language, theme, selected storage key, admin token, runtime settings, toast queue, and upload task list. Keep this in Svelte runes stores under `frontend/src/lib/stores/`.
- **Browser persistence**:
  - `omepic-client-token` for anonymous upload/delete ownership token.
  - `omepic-ui-preferences` for language/theme.
  - `omepic-upload-preferences` for selected storage key.
  - `omepic-admin-token` for admin token.
  - IndexedDB `omepic/uploads` for upload history records.
- **Server state**: admin stats, image lists, storage configs, system settings, announcements, IP bans, and abuse results. Treat the backend as the source of truth and refetch after mutations.

---

## Shared Store Contracts

### Preferences Store

Signatures:

- `preferences.language: Language`
- `preferences.theme: Theme`
- `preferences.selectedStorageKey: string`
- `preferences.adminToken: string | null`
- `preferences.runtimeSettings: PublicRuntimeSettings | null`
- `setLanguage(language)`
- `setTheme(theme)`
- `setSelectedStorageKey(key)`
- `setRuntimeSettings(settings)`
- `setAdminToken(token)`
- `clearAdminToken()`
- `resolvedTheme()`

Contracts:

- Use setter functions for persisted fields so localStorage stays synchronized.
- Do not write localStorage directly from route components for these shared values.
- Keep `runtimeSettings` in memory; do not persist it unless the backend contract explicitly changes.
- Admin routes should read `preferences.adminToken` and clear it with `clearAdminToken()` when backend validation fails.

### Toast Store

Signatures:

- `toast.success(message)`
- `toast.error(message)`
- `toast.info(message)`
- `toasts.items`

Contracts:

- Toasts auto-remove after the store timeout.
- Toasts are supplemental; destructive flows still need confirm dialogs or visible state.

---

## When To Use Global State

Promote state into a shared runes store only when:

- multiple routes/components must react to the same browser value
- the value must persist across page reloads
- a shared shell and nested route both need the same state

Do not put server responses into a global store just because they were fetched once.

---

## Server State

- Keep server responses in route-local state unless there is a real cross-route sharing need.
- Admin pages should use typed helpers in `frontend/src/lib/api.ts` for request logic.
- IndexedDB upload history is a client convenience, not backend truth.
- A successful backend mutation must remain successful even if local IndexedDB persistence fails afterward; surface local persistence failures separately.

---

## Common Mistakes To Avoid

- Reintroducing Zustand or React hook store assumptions into the SvelteKit frontend.
- Duplicating admin table data in component state, global state, and IndexedDB.
- Storing storage secrets or long-lived sensitive config objects in browser state.
- Creating one giant app-wide store instead of scoped runes stores.
- Treating IndexedDB history as authorization for delete operations instead of sending the current `X-Token`.
- Mirroring admin storage catalogs or IP-ban lists into global state without a cross-route need.

---

## Scenario: Anonymous Upload Token

### 1. Scope / Trigger

- Trigger: public upload/delete flows need a browser-owned `X-Token` that is independent from admin JWT authentication.

### 2. Signatures

- Helper: `frontend/src/lib/client-token.ts`
- Storage key: `omepic-client-token`
- API header: `X-Token`

### 3. Contracts

- The frontend generates and persists the anonymous client token locally; the backend must not issue client upload tokens.
- Token generation must require Web Crypto. Prefer `crypto.randomUUID()` and fall back only to `crypto.getRandomValues()` bytes.
- Do not use `Math.random()` or timestamp-derived fallback tokens. If Web Crypto is unavailable, throw a clear unsupported-browser error.
- Keep anonymous `X-Token` separate from `preferences.adminToken` / Bearer admin JWT flows.

### 4. Tests Required

- `client-token.test.ts` must cover `randomUUID`, `getRandomValues`, persisted reuse, SSR/no-window behavior, and Web Crypto unavailable errors.

---

## Scenario: Admin Session Entry

### 1. Scope / Trigger

- Trigger: the frontend needs a single admin dashboard entry route that works for logged-out and already logged-in users.

### 2. Signatures

- Store: `frontend/src/lib/stores/preferences.svelte.ts`
- Admin token field: `preferences.adminToken`
- Entry layout: `frontend/src/routes/admin/dashboard/+layout.svelte`
- Login API: `adminLogin(password)`
- Validation API: `adminGetStatus(token)`

### 3. Contracts

- `/admin/dashboard` remains the admin entry route.
- The frontend should not create a separate `/admin/login` route unless the product flow changes.
- Admin layout validates a stored token through the backend before rendering protected children.
- Failed validation must call `clearAdminToken()`.
- Successful login stores the returned token through `setAdminToken(token)`.

### 4. Validation & Error Matrix

- No stored token -> show login form.
- Stored token exists -> enter a validation/loading state and call `/admin/status` before rendering protected content.
- Token validation succeeds -> render the protected child route and sidebar.
- Token validation fails -> clear token, redirect protected child routes to `/admin/dashboard`, and show the login form without flashing protected content.
- Login API fails -> keep token unset and show an error.

### 5. Good / Base / Bad Cases

- Good: reload on `/admin/dashboard/images` with a valid token validates and renders the admin route.
- Base: fresh browser with no token shows password form in the dashboard shell.
- Bad: storing admin token in a page-local variable or duplicating token storage keys.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend`.
- Verify no token -> login form.
- Verify valid token -> protected dashboard route.
- Verify invalid token -> token cleared and login form shown.

### 7. Wrong vs Correct

#### Wrong

```ts
localStorage.setItem('token', token);
```

#### Correct

```ts
setAdminToken(token);
```

#### Wrong

```svelte
{#if preferences.adminToken}
  {@render children()}
{/if}
```

#### Correct

```svelte
{#if authState === 'authenticated'}
  {@render children()}
{:else if authState === 'checking'}
  <p>{t(preferences.language, 'common.loading')}</p>
{/if}
```

---

## Scenario: Global UI Preferences

### 1. Scope / Trigger

- Trigger: language and theme must be shared across public and admin pages.

### 2. Signatures

- Store: `frontend/src/lib/stores/preferences.svelte.ts`
- Fields:
  - `language: 'en' | 'zh'`
  - `theme: 'light' | 'dark' | 'system'`
- Setters:
  - `setLanguage(language)`
  - `setTheme(theme)`

### 3. Contracts

- Persist language/theme under `omepic-ui-preferences`.
- First-time visitors and invalid stored theme values must default to `system`, so the UI follows the current OS color-scheme preference.
- `setLanguage` must also update `document.documentElement.lang` through the shared store helper.
- `resolvedTheme()` returns the actual light/dark mode for `system`.
- New UI text must be added to both language dictionaries.

### 4. Validation & Error Matrix

- Missing/corrupt localStorage payload -> fall back safely.
- Unsupported language/theme -> normalize to supported defaults (`system` for theme).
- `theme === 'system'` without `window` -> default to light.

### 5. Good / Base / Bad Cases

- Good: language/theme are changed once and all routes reflect them after reload.
- Base: no stored preferences -> browser-language fallback and light theme.
- Bad: route pages read/write independent language/theme keys directly.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend`.
- Manually verify language and theme persistence across public and admin routes.

### 7. Wrong vs Correct

#### Wrong

```ts
preferences.language = 'zh';
```

#### Correct

```ts
setLanguage('zh');
```
