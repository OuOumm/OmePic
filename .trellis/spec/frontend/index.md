# Frontend Development Guidelines

> Current frontend conventions for the checked-in SvelteKit application.

---

## Current State

- The repository contains a checked-in `frontend/` application.
- The frontend uses SvelteKit 2, Svelte 5, Vite, TypeScript, Tailwind CSS, and static adapter output.
- The active route tree is `frontend/src/routes/`.
- The active shared component tree is `frontend/src/lib/components/studio/`.
- Keep these docs synchronized with concrete source paths such as:
  - `frontend/src/routes/+page.svelte`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte`
  - `frontend/src/routes/admin/dashboard/security/+page.svelte`
  - `frontend/src/lib/api.ts`
  - `frontend/src/lib/types/index.ts`

---

## Pre-Development Checklist

Before touching frontend code, read:

1. [Directory Structure](./directory-structure.md)
2. [Component Guidelines](./component-guidelines.md)
3. [Hook Guidelines](./hook-guidelines.md)
4. [State Management](./state-management.md)
5. [Type Safety](./type-safety.md)
6. [Quality Guidelines](./quality-guidelines.md)

Also confirm the documented contracts still match the repo:

- Re-read [README.md](../../../README.md) if routes, storage behavior, auth flow, or UI requirements change.
- Update the relevant scenario sections when changes affect upload, history, admin settings, security pages, IP-ban flows, public runtime settings, announcements, or shared UI preferences.
- Do not apply old Next.js App Router, React component, or Zustand conventions to the current SvelteKit frontend.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | SvelteKit route, studio component, and shared-file layout | Active |
| [Component Guidelines](./component-guidelines.md) | Component boundaries and UI behavior | Active |
| [Hook Guidelines](./hook-guidelines.md) | Legacy/custom helper guidance; verify against Svelte code before using | Active |
| [State Management](./state-management.md) | Client state and persistence rules | Active |
| [Quality Guidelines](./quality-guidelines.md) | Review, testing, accessibility, and build expectations | Active |
| [Type Safety](./type-safety.md) | TypeScript and runtime contract rules | Active |

---

## Maintenance Rule

These docs intentionally describe the current repository reality and active cross-layer contracts. When implementation details, routes, response shapes, or framework assumptions change, update these docs in the same task.
