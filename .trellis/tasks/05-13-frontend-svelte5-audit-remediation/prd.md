# Frontend Svelte 5 audit remediation

## Goal

Implement the remediation items documented in `docs/debug/frontend-svelte5-audit-remediation-2026-05-13.md` for the checked-in SvelteKit/Svelte 5 frontend, preserving current product behavior while improving authentication gating, Svelte 5 compliance, upload validation alignment, i18n/accessibility, and public URL handling.

## Background / Confirmed Facts

- The active frontend is `frontend/`, using SvelteKit 2, Svelte 5, Vite, TypeScript, Tailwind CSS, and static adapter output.
- The audit document identifies P1/P2/P3 issues in:
  - `frontend/src/routes/admin/dashboard/+layout.svelte`
  - `frontend/src/lib/components/studio/AnnouncementDialog.svelte`
  - `frontend/src/lib/components/studio/CanvasDropzone.svelte`
  - `frontend/src/lib/components/studio/ImageDetailDrawer.svelte`
  - `frontend/src/lib/components/studio/AppShell.svelte`
  - `frontend/src/lib/components/studio/ImagePreviewDialog.svelte`
  - `frontend/src/lib/components/studio/ImageDataTable.svelte`
  - `frontend/src/lib/i18n.ts`
  - `frontend/src/lib/utils.ts`
  - legacy files `frontend/next-env.d.ts`, `frontend/eslint.config.mjs`, and unused `frontend/src/lib/actions/click-outside.ts`.
- Backend runtime settings support `public_base_url`; upload output builds absolute URLs from the effective public base URL, so frontend image preview should allow a runtime-configured public origin, not only same-origin URLs.
- Project specs require centralized API helpers, Svelte 5 runes conventions, no Next.js/React/Zustand assumptions, no SVG upload acceptance drift, and admin token validation before protected dashboard content renders.
- Current working tree already contains unrelated uncommitted frontend/docs changes; remediation must avoid reverting or overwriting unrelated user changes.

## Requirements

1. Admin route protection
   - Validate a stored admin token through `adminGetStatus()` before rendering protected dashboard children.
   - Show a loading/checking state while validation is pending.
   - Clear invalid tokens and route logged-out users to `/admin/dashboard` without exposing protected child UI.
   - Keep `/admin/dashboard` as the single login/dashboard entry route.

2. Remove legacy Next.js / unused action artifacts
   - Remove obsolete Next.js typing/config files that conflict with the SvelteKit frontend reality.
   - Remove or modernize unused old-style Svelte action code; prefer removal if unused.

3. Announcement acknowledgement correctness
   - When auto-opening latest announcement detail, display the latest announcement and acknowledge the same item.
   - Preserve the existing contract that dismissing by overlay/close/Esc does not update `omepic:announcement:lastSeen`.

4. Upload MIME/accept alignment
   - Derive the native file picker `accept` list from runtime effective MIME types where available.
   - Filter out SVG from native accept and validation paths.
   - Keep drag/drop, paste, and manual file picker behavior aligned with the same MIME policy.

5. Image detail drawer request lifecycle
   - Cancel stale IP-detail requests on image change or drawer close.
   - Avoid stale error toasts and stale loading states after closing or navigating.

6. Theme initialization consistency
   - Reduce duplicated theme-resolution logic between the tested helper and AppShell inline script path.
   - Preserve first-paint dark-mode behavior and preference persistence.
   - Support system-theme changes after initial load where practical.

7. i18n/accessibility cleanup
   - Add missing translation keys used by image detail actions.
   - Replace hardcoded visible labels in image details with existing/new i18n keys where appropriate.
   - Keep copy/action button accessible names complete.

8. Public image URL allow-list
   - Permit safe image preview/download for same-origin URLs and the runtime `public_base_url` origin.
   - Continue rejecting dangerous schemes such as `javascript:`, `data:`, `vbscript:`, and `file:`.
   - Keep behavior covered by tests.

9. Documentation and verification
   - Update the debug remediation document with completion status or add a concise follow-up note if useful.
   - Run frontend validation commands after implementation.

## Acceptance Criteria

- [x] Invalid stored admin token on `/admin/dashboard/images`, `/admin/dashboard/settings`, or `/admin/dashboard/security` shows a validation/loading state and then returns to login without rendering protected child content.
- [x] Valid stored admin token validates successfully before protected child content renders.
- [x] `frontend/next-env.d.ts`, obsolete `frontend/eslint.config.mjs`, and unused old-style `click-outside` action are removed or otherwise no longer present as misleading legacy artifacts.
- [x] Latest announcement detail auto-open resets to the latest announcement; acknowledging stores the timestamp for the displayed latest announcement only through the acknowledgement action.
- [x] File input `accept` excludes SVG and reflects runtime effective MIME types when runtime settings are loaded.
- [x] Image detail IP-detail fetches are abortable and do not update state or toast errors after the image/drawer becomes stale.
- [x] Theme first-paint and runtime theme toggling still work; helper/tests reflect the production path.
- [x] `image.copyUid` and related image detail labels are translated in both English and Chinese.
- [x] Historical image preview/download supports runtime-configured public origin while still blocking unsafe schemes and unrelated cross-origin URLs.
- [x] `npm run lint`, `npm run typecheck`, `npm run test`, and `npm run build:backend` pass from `frontend/` using the Windows `cmd.exe //c` invocation noted in the audit.

## Out of Scope

- Rebuilding the frontend UI or changing visual design beyond necessary labels/loading states.
- Backend API contract changes.
- Adding a new admin login route.
- Adding a new runtime validation library.
- Refactoring unrelated pre-existing uncommitted changes.

## Open Questions

None currently blocking. Repository evidence shows runtime `public_base_url` is an intended public URL contract, so the frontend should allow that configured origin for image preview/download.
