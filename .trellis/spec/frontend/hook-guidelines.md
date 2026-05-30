# Hook / Reusable Client Logic Guidelines

> Guidance for reusable browser-side logic in the current SvelteKit frontend.

---

## Current State

- The active frontend is SvelteKit + Svelte 5, so React hooks do not apply.
- Reusable browser-side logic should be implemented as plain TypeScript helpers, Svelte actions, Svelte components, or Svelte runes stores depending on the behavior.
- Current shared locations:
  - `frontend/src/lib/api.ts`
  - `frontend/src/lib/stores/*.svelte.ts`
  - `frontend/src/lib/indexeddb/*.ts`
  - `frontend/src/lib/actions/*.ts`
  - `frontend/src/lib/components/studio/*.svelte`

---

## Reuse Patterns

- Use a Svelte component when the reusable unit owns markup and UI behavior.
- Use a Svelte action when the reusable unit attaches DOM behavior to an element.
- Use a runes store when multiple components/routes need shared reactive state.
- Use a plain TypeScript helper when the reusable unit is stateless or wraps an API/persistence boundary.
- Keep feature-specific orchestration in the route component when it is not reused elsewhere.

---

## Data Fetching

- Centralize HTTP concerns in `frontend/src/lib/api.ts`.
- Route components may own load/refresh timing and local loading/error state.
- Pass `AbortSignal` through request helpers when a route can cancel stale requests.
- Debounce high-cardinality admin filters such as image search.
- Do not introduce React Query, SWR, or a parallel cache layer without updating this spec and explaining the boundary.

---

## Browser APIs

- Guard `window`, `document`, `localStorage`, `navigator.clipboard`, and IndexedDB access for non-browser contexts.
- Keep IndexedDB operations behind `frontend/src/lib/indexeddb/upload-history.ts`.
- Keep localStorage persistence for shared preferences/admin token behind `frontend/src/lib/stores/preferences.svelte.ts`.

---

## Naming Conventions

- Svelte actions use action-style names and live under `frontend/src/lib/actions/`.
- Svelte runes stores use `*.svelte.ts`.
- Plain helpers should be named after the capability they provide, such as `getApiBaseUrl`, `getImageUrl`, or `saveUploadToHistory`.
- Avoid names that imply React hooks, such as `useSomething`, unless React is reintroduced intentionally and the frontend spec is updated.

---

## Common Mistakes To Avoid

- Reintroducing React hooks into Svelte route/components.
- Reading localStorage directly in many routes instead of using the preferences store.
- Hiding network side effects in generic utilities that make route behavior hard to trace.
- Creating multiple IndexedDB wrappers for upload history.
- Returning untyped JSON from helpers and letting route components infer backend payloads ad hoc.
