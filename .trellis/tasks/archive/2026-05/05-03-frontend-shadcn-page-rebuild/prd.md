# Refactor Frontend Pages With shadcn/ui

## Goal

Rebuild the OmePic frontend visual system and page layouts around the official shadcn/ui product-app style. Discard the current glassy neon design language and replace it with a restrained, modern, component-driven interface that feels close to shadcn/ui examples while preserving existing frontend behavior and backend API contracts.

## Requirements

* Rework the frontend UI style globally, including `globals.css`, Tailwind tokens, shared UI primitives, shared page layout components, and major page compositions.
* Use shadcn/ui "new-york" conventions already configured in `frontend/components.json`: neutral slate base, CSS variables, Radix-backed primitives, lucide icons, compact borders, predictable spacing, and card/header/content/footer composition.
* Replace decorative glass/neon/gradient-heavy presentation with a product UI style: flat background, subtle borders, muted panels, focused accent usage, restrained shadows, and compact information hierarchy.
* Rebuild the public upload page layout so upload remains the primary workflow, with storage selection, progress/status, latest result, client token, and recent uploads organized in shadcn-like cards and responsive grids.
* Rebuild supporting public pages (`history`, `api`) so they share the same page shell, section headers, cards, tables/grids, empty states, and action styling.
* Rebuild admin pages (`login`, dashboard status, images, settings) toward official shadcn dashboard patterns: sidebar/inset-style layout, sticky or stable header areas, section cards, data-table style image management, clear forms, and consistent destructive/secondary states.
* Preserve existing user-facing functionality: upload, selectable storage, recent/local history, copy actions, language/theme controls, admin login/session, admin status, image table actions, and settings management.
* Preserve existing API request/response contracts, routes, state stores, auth behavior, and data flow unless a visual component boundary requires internal-only refactoring.
* Keep accessibility intact or improve it: keyboard focus, labels/descriptions, aria status/error handling, contrast, reduced-motion behavior, and responsive text/layout without overlap.
* Do not add a marketing landing page or explanatory feature text. The first screen remains the usable upload application.

## Acceptance Criteria

* [ ] The app no longer reads as glass/neon/gradient-heavy; it reads as a shadcn/ui-style product app with neutral surfaces, borders, compact rhythm, and restrained accent usage.
* [ ] Shared UI primitives follow shadcn-style composition and state behavior, including button, card, badge, input, select, textarea, table, tabs, dialog/dropdown where touched.
* [ ] Public upload, history, API, admin login, admin dashboard, admin images, and admin settings pages have coherent shared layout and spacing.
* [ ] Upload workflow behavior remains intact, including storage option loading/refresh, progress, duplicate result display, copy buttons, recent uploads, and local-history save error handling.
* [ ] Admin behavior remains intact, including session verification, navigation, status display, image actions/search/loading, and settings forms.
* [ ] Desktop and mobile layouts remain stable with no incoherent overlap, clipped button text, or layout shift from hover/status labels.
* [ ] Light/dark/system theme and zh/en preference controls still work across public and admin pages.
* [ ] `npm run lint`, `npm run typecheck`, and `npm run build` pass from `frontend/`.

## Definition of Done

* Code changes are scoped to frontend UI/layout/style and task bookkeeping.
* Existing uncommitted user/WIP changes are preserved and integrated rather than reverted.
* Lint, typecheck, and build have been run or any blocker is reported with exact command output.
* Specs are reviewed for whether a new durable frontend convention should be recorded.

## Technical Approach

Use a full rebuild of the frontend visual layer, not light polish. Start from the existing shadcn setup (`components.json`, Radix dependencies, Tailwind CSS variables), then make shared primitives and layout wrappers carry the design language so page files do not accumulate one-off styling. Favor official shadcn patterns: card composition, neutral sections, sidebar/dashboard rhythm, table/card density, lucide icons, and clear state variants.

## Decision (ADR-lite)

**Context**: The current UI has accumulated glass panels, gradient text, radial backgrounds, large rounded shapes, and decorative glow from previous visual tasks. The user explicitly asked to discard the old design and reference official shadcn/ui.

**Decision**: Rebuild the frontend around shadcn/ui's restrained product-app style while preserving app behavior and backend contracts. Treat this as a visual/layout refactor, not a backend or API redesign.

**Consequences**: This will touch many frontend files and may replace prior visual conventions. The implementation should centralize new styling in shared primitives and layout components so later pages inherit the shadcn-like system instead of recreating local styles.

## Out of Scope

* Backend API, storage, database, UID, deduplication, or auth contract changes.
* Adding new product features beyond UI affordances required by the redesign.
* Replacing the existing Next.js/Tailwind/Radix/shadcn setup with a different design system.
* Introducing charting, analytics, or dashboard data that does not already exist.

## Technical Notes

* `frontend/components.json` already uses shadcn/ui style `new-york`, base color `slate`, CSS variables, and lucide icons.
* Current WIP already touches shared UI primitives and major page components; do not revert unknown changes.
* Official shadcn references are recorded in `research/shadcn-official-style.md`.
* Memory from prior OmePic frontend work says the homepage route should stay thin and `UploadPageClient` is the control point for homepage hierarchy.
