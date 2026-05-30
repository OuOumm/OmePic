# Frontend Admin Login Session Redirect

## Goal

When an admin user opens `/admin/dashboard`, the frontend should use that dashboard route as the only admin entry point: show the login form when logged out and show the admin dashboard directly when a valid persisted admin session exists. The `/admin/login` frontend route should be removed.

## Requirements

* Preserve the successful login flow: password submission stores the admin token.
* Remove the `/admin/login` frontend page route.
* Point frontend admin navigation at `/admin/dashboard`.
* On `/admin/dashboard`, wait for the persisted admin session store to hydrate before deciding what to render.
* If no persisted token exists, render the login form directly inside the dashboard route.
* If a persisted admin token exists and backend validation succeeds, render the admin dashboard.
* Keep the dashboard shell as the single place that verifies the stored token against the backend.
* Do not change backend auth contracts or token format.

## Acceptance Criteria

* [ ] `/admin/login` frontend route is removed.
* [ ] The header admin entry links to `/admin/dashboard`.
* [ ] After a successful admin login, `/admin/dashboard` swaps into the admin dashboard without requiring another route hop first.
* [ ] Clicking the header/admin entry while logged in shows the admin dashboard directly.
* [ ] A fresh browser/session with no stored admin token still shows the login form.
* [ ] The dashboard route avoids a misleading form flash while persisted session hydration is still pending.
* [ ] Frontend lint/type-check/build checks pass or any environmental blocker is recorded.

## Definition of Done

* Tests added or updated if the repo has a matching lightweight pattern.
* Frontend lint/typecheck/build pass where possible.
* Trellis spec update considered if the task creates a new durable convention.

## Technical Approach

Update `AdminShell` so `/admin/dashboard` handles all admin session branching. It reads `token` and `hasHydrated` from `useAdminSessionStore`; after hydration, it renders the existing login form for logged-out users or validates the token and renders dashboard children for logged-in users. Remove the `/admin/login` frontend route and point header admin navigation to `/admin/dashboard`.

## Decision (ADR-lite)

**Context**: The admin dashboard already validates tokens in `AdminShell`, and a separate frontend login route creates duplicate route/session branching.

**Decision**: Make `/admin/dashboard` the only frontend admin entry route. It branches on persisted session hydration and renders either the login form or admin dashboard shell directly, without introducing new server-side auth logic.

**Consequences**: Admin navigation has one frontend route. Stale/invalid tokens are still rejected by the dashboard validation path and cleared there.

## Out of Scope

* Backend authentication changes.
* Token expiry/refresh design.
* Backend route changes for `POST /admin/login`.

## Technical Notes

* Existing admin session store: `frontend/src/stores/admin-session-store.ts`.
* Existing login form: `frontend/src/features/admin/LoginForm.tsx`.
* Existing dashboard auth validation: `frontend/src/features/admin/AdminShell.tsx`.
* Frontend spec index: `.trellis/spec/frontend/index.md`.
