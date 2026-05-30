# Quality Guidelines

> Frontend quality bar for the checked-in implementation.

---

## Current State

- The frontend package, lint config, TypeScript config, and Vitest test runner exist under `frontend/`.
- Existing frontend tests cover API helper contracts, URL/MIME utilities, client token generation, upload queue helpers, and performance-oriented helper logic.
- These rules define the minimum quality expectations for frontend changes.

---

## Forbidden Patterns

- Reintroducing Next.js App Router, React Client Component, or Zustand assumptions into the current SvelteKit frontend.
- Hiding API calls directly inside presentational components.
- Duplicating request logic across upload, history, admin, security, and settings flows instead of sharing `frontend/src/lib/api.ts`.
- Using `any` or broad type assertions to get past response typing.
- Putting tokens or admin secrets into query strings.
- Creating one unrelated global state bucket outside the existing Svelte runes stores.

---

## Required Patterns

- Keep SvelteKit route files focused on page orchestration and use `frontend/src/lib/components/studio/` for reusable UI.
- Centralize auth header injection for admin requests in `frontend/src/lib/api.ts`.
- Keep upload history persistence behind `frontend/src/lib/indexeddb/upload-history.ts` instead of scattered direct IndexedDB calls.
- Use Svelte 5 runes stores under `frontend/src/lib/stores/` for shared client preferences, admin token, runtime settings, and toast state.
- Surface async success and failure clearly in the UI; toasts are helpful but not sufficient by themselves for destructive or blocking flows.

---

## Testing Requirements

Current frontend unit tests run with Vitest. Prioritize coverage for high-risk client logic:

- token bootstrap logic
- upload queue concurrency/progress mapping and copy helpers
- delete authorization behavior in history UI
- admin session persistence and route gating
- security-page IP ban and abuse flows
- Svelte runes store behavior if non-trivial stores are introduced

Minimum local checks:

- `npm run lint`
- `npm run typecheck`
- `npm run test`
- `npm run build:backend`

Command-line conventions:

- Use `git` directly in project commands; do not hardcode a machine-specific Git executable path.
- Use the project shell's default `node`, `npm`, and `npx` commands directly. Do not prepend a hardcoded local Node tool directory or rewrite `PATH` unless a command actually fails because Node is missing.
- Do not manually delete `frontend/.svelte-kit` as routine build hygiene. SvelteKit refreshes this generated directory during `svelte-kit sync` and `vite build`; only remove it when debugging stale generated output or after confirming the directory is causing a concrete failure.

Use `npm run build:backend` for every frontend-affecting task's final verification. Do not substitute `npm run build` as the final build check, because the project deploys the static frontend through the Go backend's `backend/web/` assets.

If unit or component tests are added, record the exact command here instead of assuming one exists today.

---

## Scenario: Static Export For Backend Serving

### 1. Scope / Trigger

- Trigger: Production frontend changes that affect the single-port Go backend deployment.
- Scope: SvelteKit adapter-static exports static files, then the build flow copies them into `backend/web/`.

### 2. Signatures

- Frontend build script: `npm run build:backend`.
- Raw export directory: `frontend/out/`.
- Backend-served copy: `backend/web/`.
- SvelteKit config: `@sveltejs/adapter-static` outputs pages/assets to `out` with `fallback: 'index.html'` and `precompress: true`.
- Vite env key: `VITE_API_BASE_URL`.

### 3. Contracts

- `frontend/out/` and `backend/web/` are generated artifacts and must stay ignored by git.
- Adapter-static precompression should stay enabled so `build:backend` emits `.br` and `.gz` assets alongside static output; backend serving should prefer those files when the request accepts them.
- In production, frontend API helpers should default to same-origin relative URLs unless `VITE_API_BASE_URL` is explicitly set.
- In development/non-browser contexts, the frontend may continue to fall back to `http://localhost:8080` so the split-port workflow remains usable.

### 4. Validation & Error Matrix

- Missing `VITE_API_BASE_URL` in production browser -> API calls use relative paths such as `/v1/image`.
- Missing `VITE_API_BASE_URL` in non-browser context -> API calls may target `http://localhost:8080`.
- Static export build failure -> do not copy stale assets into `backend/web/`.
- Copy failure -> production single-port deployment is invalid until `build:backend` succeeds.

### 5. Good/Base/Bad Cases

- Good: `npm run build:backend` runs `vite build`, emits compressed static assets, and then copies the current `frontend/out/` into `backend/web/`.
- Base: `npm run build` may be useful for isolated SvelteKit diagnostics, but it is not sufficient as the final build verification.
- Bad: hardcoding `http://localhost:8080` into production-exported frontend bundles.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend` for frontend changes.
- Pair with backend router tests when frontend route shape or backend fallback behavior changes.

### 7. Wrong vs Correct

#### Wrong

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";
```

#### Correct

```ts
const envBase = import.meta.env.VITE_API_BASE_URL;
const API_BASE = typeof window === "undefined" ? "http://localhost:8080" : envBase?.replace(/\/+$/, "") || "";
```

---

## Code Review Checklist

- Does the change preserve the route contract defined in [README.md](../../../README.md)?
- Are browser-only APIs guarded behind browser/runtime checks in Svelte components and helpers?
- Are admin requests consistently authenticated with `Authorization: Bearer <token>`?
- Are `X-Token` flows kept separate from admin auth flows?
- Did the author avoid inventing parallel state layers for the same data?
- Are keyboard access, labels, and error states present for upload, delete, login, security, and settings interactions?
- If the change touches global UI preferences, does it update both the shared translation source and the root theme/lang sync path?
- If the change introduces new visible UI copy, is it covered by both `en` and `zh` dictionaries instead of hardcoded in a single component?
