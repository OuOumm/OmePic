# Serve Frontend Build From Backend

## Goal

Enable production deployment to run OmePic through a single backend port by building the Next.js frontend into static assets and serving those assets from the Go/Gin backend.

## What I Already Know

* The frontend is a Next.js 16.2.4 App Router app under `frontend/`.
* `frontend/package.json` currently runs `next build` for production builds.
* `frontend/next.config.ts` is empty, so static export is not configured yet.
* The backend uses Gin and registers API/image/admin routes in `backend/internal/http/router/router.go`.
* The backend currently listens on `http://localhost:8080` by default.
* `frontend/src/lib/api.ts` defaults API calls to `http://localhost:8080`, which works for local split-port development but should prefer same-origin in bundled production.
* Current frontend routes are static-looking app pages: `/`, `/history`, `/api`, `/admin/dashboard`, `/admin/dashboard/images`, `/admin/dashboard/settings`.

## Requirements

* Add a production frontend build path that emits static frontend files usable without a Node.js frontend server.
* Configure the backend to serve the built frontend files when present.
* Keep backend API/image routes authoritative:
  * `/health`
  * `/v1/*`
  * `/i/*`
  * `/admin/login`
  * `/admin/status`, `/admin/images`, `/admin/config`, and related admin config routes
* Add frontend SPA/page fallback so direct browser visits to frontend pages load the built app.
* Keep local development split-port workflow usable.
* Make same-port production API requests work without requiring `NEXT_PUBLIC_API_BASE_URL`.
* Document the production build/run flow.

## Acceptance Criteria

* [ ] `npm run build` in `frontend/` produces static frontend output.
* [ ] Running the Go backend can serve the built frontend from the same HTTP port.
* [ ] Direct visits to frontend routes such as `/`, `/history`, and `/admin/dashboard` load frontend assets instead of 404.
* [ ] Existing API routes still return JSON/API behavior and are not shadowed by frontend fallback.
* [ ] Frontend API calls work on same-origin production hosting.
* [ ] Backend and frontend quality checks pass.

## Definition of Done

* Tests added or updated where appropriate.
* Backend tests pass.
* Frontend lint, typecheck, and build pass.
* README documents the single-port production flow.
* Rollback remains straightforward: remove/ignore frontend build directory and run split-port development as before.

## Technical Approach

Chosen MVP approach:

* Configure Next.js static export via `output: "export"` and ensure route/image settings are compatible with static hosting.
* Add or adjust frontend build scripts so `frontend/out/` is the raw build artifact.
* Add a build/copy workflow that places the static frontend artifact under `backend/web/`.
* Serve `backend/web/` from Gin after all API routes are registered.
* Use a no-route fallback to return `index.html` or matching static HTML for frontend page paths while preserving API route behavior.
* Change frontend API base resolution so production can use relative URLs by default while development can still use `NEXT_PUBLIC_API_BASE_URL` or fall back to `http://localhost:8080`.

## Decision (ADR-lite)

**Context**: The backend should be able to run the complete production app on one port without depending on a live Next.js server.

**Decision**: Build the frontend as static files, then copy the artifact into `backend/web/`; the Go backend serves only that backend-owned static directory.

**Consequences**: Release artifacts are clearer because the backend owns the files it serves. The build flow must keep `backend/web/` synchronized with `frontend/out/`, and generated static assets should be handled so they do not accidentally create unrelated source churn.

## Open Questions

* None.

## Out of Scope

* Replacing Next.js with another frontend toolchain.
* Server-side rendering through the Go backend.
* Reverse proxying a live Next.js server from Go.
* Changing image upload/storage behavior.

## Technical Notes

* Backend route registration: `backend/internal/http/router/router.go`.
* Backend entrypoint: `backend/cmd/server/main.go`.
* Frontend API base: `frontend/src/lib/api.ts`.
* Frontend config: `frontend/next.config.ts`.
* Frontend build command: `frontend/package.json`.
* README currently documents split backend/frontend local development only.
