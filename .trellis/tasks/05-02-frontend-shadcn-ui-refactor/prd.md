# Refactor Frontend With shadcn/ui

## Goal

Refactor the checked-in `frontend/` application to use shadcn/ui-style primitives and composition patterns, referencing official shadcn/ui setup, theming, component docs, and dashboard examples. The refactor should improve consistency, maintainability, accessibility, and product-grade visual quality while preserving the existing OmePic user-facing workflows and backend API contracts.

## What I Already Know

* The user requested a full frontend refactor with shadcn/ui and official examples as the reference.
* The frontend is a Next.js 16.2.4 App Router app using TypeScript, Tailwind CSS 3.4, Zustand, and `react-hot-toast`.
* Current UI primitives already live under `frontend/src/components/ui`, but they are custom local components rather than official shadcn/ui-generated primitives.
* Current app areas include public upload, upload history, API documentation, admin login, admin dashboard overview, image management, and settings/storage management.
* Existing global preferences for language and theme must keep working across public and admin pages.
* Existing backend contracts, storage selection behavior, upload history behavior, admin auth/session verification, and i18n copy should be preserved unless explicitly changed.

## Requirements

* Initialize or align the frontend with shadcn/ui conventions inside `frontend/`, including `components.json`, local copied primitives, `@/lib/utils`, Tailwind content paths, and semantic theme tokens.
* Use official shadcn/ui docs/examples as design and implementation references, especially Next.js setup, theming, Card, Button, Input, Select, Badge, Table, Dialog, Dropdown Menu, Sidebar, Tabs, and dashboard blocks.
* Preserve the current Next.js App Router routes and user workflows:
  * `/` upload page with storage target selection, drag/drop upload, progress, latest result, copy buttons, recent uploads, and client token display.
  * `/history` upload history with image cards/lightbox behavior.
  * `/api` API documentation page.
  * `/admin/login` admin login flow.
  * `/admin/dashboard` status overview.
  * `/admin/dashboard/images` image management.
  * `/admin/dashboard/settings` storage/settings management.
* Replace or bridge current custom UI primitives with shadcn/ui-style primitives without breaking existing behavior.
* Prefer official accessible primitives for interactive controls such as Select, Dialog, Dropdown Menu, Tabs, and Sidebar instead of hand-rolled equivalents.
* Retain existing language/theme preference state and ensure shadcn semantic variables work in both light and dark modes.
* Improve visual consistency using shadcn-style spacing, surface, border, focus, and state systems while keeping the app visually modern and polished.
* Use lucide-react icons where suitable instead of inline SVG icons, unless a local icon is clearly better.
* Keep the refactor behavior-preserving: no backend API changes, no route changes, and no removal of existing public/admin capabilities.

## Acceptance Criteria

* [ ] `frontend/` contains a valid shadcn/ui setup and reusable shadcn-style primitives.
* [ ] Public upload, history, API docs, admin login, admin dashboard, image management, and settings pages render with the new component system.
* [ ] Existing upload, storage selection, copy actions, history, admin login, admin image actions, and admin settings actions remain wired to the existing API functions.
* [ ] Light/dark/theme preference behavior remains functional and maps cleanly to shadcn semantic CSS variables.
* [ ] Interactive controls have clear focus-visible states, accessible labels, disabled/loading/error states, and responsive mobile layouts.
* [ ] `npm run lint`, `npm run typecheck`, and `npm run build` pass from `frontend/`.

## Definition of Done

* Tests or targeted verification are added/updated where the refactor changes meaningful behavior or state flow.
* Lint, type-check, and build pass.
* shadcn/ui official reference research is persisted under `research/`.
* Trellis frontend specs remain accurate; update specs if the new component/theming conventions become durable project rules.

## Research References

* [`research/shadcn-official-examples.md`](research/shadcn-official-examples.md) - Official shadcn/ui Next.js setup, theming, dashboard/forms/table/navigation patterns, and OmePic Tailwind mapping notes.

## Technical Approach

Recommended approach: a full frontend visual/component refactor with incremental implementation order.

1. Add shadcn/ui setup and semantic theme bridge while preserving existing custom tokens needed by current screens.
2. Generate or implement shadcn-style primitives for the components needed by the app.
3. Replace current `components/ui` imports consistently, choosing one file naming convention and avoiding mixed PascalCase/lowercase component imports.
4. Refactor global layout and shared header to shadcn-style navigation/actions while preserving language/theme controls.
5. Refactor public upload/history/API pages.
6. Refactor admin shell/sidebar, dashboard cards, image table/actions, login form, and settings forms.
7. Run frontend quality checks and fix issues.

## Decision (ADR-lite)

**Context**: The app already has custom Tailwind UI primitives and project-specific theme variables. Official shadcn/ui uses copied local primitives, semantic CSS variables, Radix UI primitives, and dashboard composition examples.

**Decision**: Use shadcn/ui as the frontend component and composition standard, but preserve existing OmePic behavior and migrate through a theme bridge instead of deleting project-specific tokens in one risky step.

**Consequences**: This may add dependencies such as Radix primitives, class-variance-authority, and lucide-react. Official Form examples may imply `react-hook-form` and a schema validator, but the first pass should only adopt them where they reduce risk and complexity.

## Out of Scope

* Backend API changes.
* Database or storage behavior changes.
* Replacing current upload/history/admin business workflows.
* Tailwind major-version upgrade unless shadcn initialization requires it and the change is explicitly accepted.
* Copying official dashboard blocks verbatim over OmePic routing, session, i18n, or storage contracts.

## Technical Notes

* Research found that shadcn initialization should run from `frontend/`, not repo root.
* Current `frontend/src/lib/utils.ts`, `frontend/tsconfig.json` alias, Tailwind setup, and `src/components/ui` location already match many shadcn conventions.
* Current theme sync uses `data-theme` plus class-based dark mode. The shadcn theme bridge should support both the existing preference system and `.dark` variables.
* The implementation should avoid Zustand selector regressions; previous frontend work established that inline object selectors can cause render loops under Zustand 5.

## Confirmation

The user asked to refactor the entire frontend with shadcn/ui and official examples. Treat this task as a comprehensive frontend migration, not a narrow staged polish pass.
