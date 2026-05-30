# Directory Structure

> Frontend layout guidance for the current checked-in SvelteKit application.

---

## Current Repository Reality

- The repository contains a checked-in `frontend/` tree.
- The active frontend framework is SvelteKit 2 + Svelte 5 + Vite + TypeScript, not Next.js.
- Paths below describe the current SvelteKit file-route layout and studio component boundaries.
- Keep this document aligned with real source paths when frontend files move.

---

## Current Layout

```text
frontend/
|-- src/
|   |-- routes/
|   |   |-- +layout.svelte
|   |   |-- +page.svelte
|   |   |-- +error.svelte
|   |   |-- history/+page.svelte
|   |   |-- api/+page.svelte
|   |   `-- admin/dashboard/
|   |       |-- +layout.svelte
|   |       |-- +page.svelte
|   |       |-- images/+page.svelte
|   |       |-- security/+page.svelte
|   |       `-- settings/+page.svelte
|   |-- lib/
|   |   |-- actions/
|   |   |   |-- accessible-dialog.ts
|   |   |   `-- viewport-portal.ts
|   |   |-- api.ts
|   |   |-- api.test.ts
|   |   |-- client-token.ts
|   |   |-- client-token.test.ts
|   |   |-- clipboard.ts
|   |   |-- clipboard.test.ts
|   |   |-- i18n.ts
|   |   |-- password-policy.ts
|   |   |-- password-policy.test.ts
|   |   |-- performance-utils.test.ts
|   |   |-- preferences.ts
|   |   |-- ui-errors.ts
|   |   |-- ui-errors.test.ts
|   |   |-- upload-queue.ts
|   |   |-- upload-queue.test.ts
|   |   |-- utils.ts
|   |   |-- utils.test.ts
|   |   |-- components/
|   |   |   `-- studio/             # Reusable business UI components
|   |   |-- indexeddb/
|   |   |   |-- upload-history.ts
|   |   |   `-- upload-history.test.ts
|   |   |-- stores/
|   |   |   |-- preferences.svelte.ts
|   |   |   |-- preferences.test.ts
|   |   |   |-- toast.svelte.ts
|   |   |   `-- upload-queue.svelte.ts
|   |   `-- types/
|   |       `-- index.ts
|   `-- app.css
|-- scripts/copy-static-to-backend.mjs
|-- svelte.config.js
|-- vite.config.ts
|-- tailwind.config.ts
`-- package.json
```

This structure keeps page routes under SvelteKit `routes/` while shared business UI, API contracts, IndexedDB, types, and runes stores live under `src/lib/`.

---

## Module Organization

- `frontend/src/routes/` owns route pages and layouts.
- `frontend/src/routes/+layout.svelte` wraps all pages with `AppShell`.
- `frontend/src/routes/+error.svelte` owns the global error boundary page.
- `frontend/src/routes/admin/dashboard/+layout.svelte` owns admin login-state validation, sidebar layout, and child page rendering.
- `frontend/src/lib/components/studio/` owns reusable project UI such as shell, upload dropzone, image table, image preview/detail, storage management, storage health charts, announcement management, confirm dialog, toast viewport, image switch button, page title component, metric strip, and markdown content renderer.
- `frontend/src/lib/api.ts` owns all public and admin request helpers. Do not duplicate endpoint fetch logic inside route pages.
- `frontend/src/lib/api.test.ts` owns API helper contract tests.
- `frontend/src/lib/client-token.ts` owns anonymous client token generation and persistence for upload/delete flows.
- `frontend/src/lib/client-token.test.ts` owns token generation tests.
- `frontend/src/lib/upload-queue.ts` owns pure upload queue helpers (bounded concurrency, progress deduplication).
- `frontend/src/lib/password-policy.ts` owns password strength helper text/validation used by the admin settings UI.
- `frontend/src/lib/stores/` owns Svelte 5 runes stores: `preferences.svelte.ts` (language, theme, admin token, runtime settings), `toast.svelte.ts` (notification queue), and `upload-queue.svelte.ts` (file upload tasks with progress and concurrency).
- `frontend/src/lib/indexeddb/upload-history.ts` owns browser upload-history persistence (IndexedDB).
- `frontend/src/lib/actions/accessible-dialog.ts` owns the accessible dialog focus-trap Svelte action.
- `frontend/src/lib/actions/viewport-portal.ts` owns body-level modal portal attachment so `fixed inset-0` overlays are not clipped or constrained by route/layout containers.
- `frontend/src/lib/types/index.ts` owns all shared API response types, admin types, storage configuration types, runtime settings types, and upload history record types.
- `frontend/src/lib/i18n.ts` owns internationalization utilities and translation dictionaries (English and Chinese).
- `frontend/src/lib/ui-errors.ts` owns UI-specific error mapping utilities (API error code to user-facing messages).
- `frontend/src/lib/clipboard.ts` owns clipboard copy utilities.
- `frontend/src/lib/preferences.ts` re-exports `getClientToken` from `client-token.ts` for backwards compatibility.
- `frontend/src/lib/utils.ts` owns shared utilities: `cn()` (tailwind-merge + clsx), `formatBytes()`, `formatDate()`, `getApiBaseUrl()`, `safeImageUrl()`, MIME-type helpers, theme helpers, and `initialThemeScript()` for flicker-free dark mode.
- `frontend/src/app.css` owns CSS variables, dark/light theme tokens, and shared studio-style classes.

---

## Naming Conventions

- Svelte route files use SvelteKit conventions: `+layout.svelte`, `+page.svelte`.
- Shared Svelte components use `PascalCase.svelte`.
- Svelte runes stores use `*.svelte.ts`, for example `preferences.svelte.ts`.
- Browser API helpers and pure utilities use regular `.ts` files.
- New pure helper modules should sit under `frontend/src/lib/` and include colocated `*.test.ts` coverage when they encode behavior, for example `client-token.ts` and `upload-queue.ts`.
- Route folders remain lowercase: `history`, `api`, `admin`, `dashboard`, `images`, `security`, `settings`.

---

## Current Examples

- `frontend/src/routes/+page.svelte` owns the upload home route orchestration (CanvasDropzone, file list, upload queue, announcement dialog, and runtime settings integration).
- `frontend/src/routes/+error.svelte` owns the global error boundary with status-aware messages and retry action.
- `frontend/src/routes/history/+page.svelte` owns upload-history display and user deletion flow.
- `frontend/src/routes/api/+page.svelte` owns API documentation page showing curl examples and sharing formats.
- `frontend/src/routes/admin/dashboard/+page.svelte` owns admin status overview (image count, storage size, today's uploads).
- `frontend/src/routes/admin/dashboard/images/+page.svelte` owns admin image listing, search, deletion, and IP ban creation flow.
- `frontend/src/routes/admin/dashboard/security/+page.svelte` owns abuse overview and IP-ban operations.
- `frontend/src/routes/admin/dashboard/settings/+page.svelte` owns admin runtime settings editor (site metadata, upload policy, AVIF quality/speed/concurrency/timeout, max image pixels, rate limits, password change).
- `frontend/src/lib/components/studio/AppShell.svelte` owns the global shell with navigation bar, theme toggle, language toggle, and mobile menu.
- `frontend/src/lib/components/studio/CanvasDropzone.svelte` owns the file dropzone with click-to-browse and drag-and-drop support.
- `frontend/src/lib/components/studio/ImageDataTable.svelte` owns reusable admin image table/grid display with selection and batch actions.
- `frontend/src/lib/components/studio/ImageDetailDrawer.svelte` owns admin image detail display with navigation, delete, IP detail, and IP ban actions.
- `frontend/src/lib/components/studio/ImagePreviewDialog.svelte` owns public upload/history image preview with sharing links.
- `frontend/src/lib/components/studio/ImageSwitchButton.svelte` owns image search copy-to-clipboard button with check mark feedback.
- `frontend/src/lib/components/studio/BanIPDialog.svelte` owns creating IP bans from UID or IP address.
- `frontend/src/lib/components/studio/IPDetailPanel.svelte` owns IP-level abuse detail display.
- `frontend/src/lib/components/studio/StorageInspector.svelte` owns storage configuration inspection and editing.
- `frontend/src/lib/components/studio/StorageInstanceManager.svelte` owns adding/editing/deleting storage instances.
- `frontend/src/lib/components/studio/AnnouncementManager.svelte` owns admin announcement CRUD.
- `frontend/src/lib/components/studio/AnnouncementDialog.svelte` owns public announcement display with acknowledge/dismiss semantics.
- `frontend/src/lib/components/studio/ConfirmDialog.svelte` owns reusable confirmation dialog for destructive actions.
- `frontend/src/lib/components/studio/ToastViewport.svelte` owns toast notification rendering.
- `frontend/src/lib/components/studio/MarkdownContent.svelte` owns markdown-safe content rendering for announcements and API docs.
- `frontend/src/lib/components/studio/PageTitle.svelte` owns page title and breadcrumb display for admin dashboard.
- `frontend/src/lib/components/studio/MetricStrip.svelte` owns metric display strip for admin status.
- `frontend/src/lib/components/studio/LineChart.svelte` owns compact storage health trend rendering.
- `frontend/src/lib/api.ts` centralizes public upload/runtime/announcement APIs and all admin APIs.
- `frontend/src/lib/indexeddb/upload-history.ts` owns upload-history persistence.
- `frontend/src/lib/stores/preferences.svelte.ts` owns language, theme, selected storage, admin token, and runtime settings.
- `frontend/src/lib/stores/toast.svelte.ts` owns toast notification queue.
- `frontend/src/lib/stores/upload-queue.svelte.ts` owns reactive file upload task list with progress tracking and concurrency limiting.

---

## Scenario: SvelteKit Static SPA For Backend Serving

### 1. Scope / Trigger

- Trigger: Current frontend is SvelteKit static output copied into the Go backend for single-port serving.

### 2. Signatures

- Route files: `frontend/src/routes/**/+page.svelte` and `+layout.svelte`.
- Static adapter config: `frontend/svelte.config.js` with `@sveltejs/adapter-static`.
- Build script: `npm run build:backend`.
- Raw output: `frontend/out/`.
- Backend-served output: `backend/web/`.

### 3. Contracts

- `frontend/out/` and `backend/web/` are generated artifacts.
- SvelteKit static adapter must output to `out/` with SPA fallback `index.html`.
- Production single-port deploy must be verified with `npm run build:backend`, not only `npm run build`.
- API base URL is read from `VITE_API_BASE_URL`; if absent in the browser, API helpers must use same-origin relative URLs.
- Non-browser fallback for API helpers may use `http://localhost:8080` for build/dev compatibility.

### 4. Validation & Error Matrix

- `vite build` failure -> do not copy stale assets into `backend/web/`.
- Copy failure -> production single-port deployment is invalid.
- Missing `VITE_API_BASE_URL` in production browser -> use same-origin paths such as `/v1/image`.
- Missing `VITE_API_BASE_URL` in non-browser context -> fallback to `http://localhost:8080`.

### 5. Good/Base/Bad Cases

- Good: `npm run build:backend` builds SvelteKit output and copies `frontend/out/` into `backend/web/`.
- Base: `npm run build` is acceptable for isolated SvelteKit diagnostics only.
- Bad: reintroducing Next.js `src/app` conventions, Zustand stores, or `NEXT_PUBLIC_API_BASE_URL` into the active SvelteKit frontend.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend` for frontend changes.
- Pair with backend router tests when route shape or static fallback behavior changes.

### 7. Wrong vs Correct

#### Wrong

```ts
const base = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
```

#### Correct

```ts
const envBase = import.meta.env.VITE_API_BASE_URL;
const base = typeof window === 'undefined' ? 'http://localhost:8080' : envBase?.replace(/\/+$/, '') || '';
```
