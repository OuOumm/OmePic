# Research: shadcn/ui official examples for Next.js App Router refactor

- Query: Research current official shadcn/ui examples relevant to refactoring a Next.js App Router frontend for an image hosting/admin dashboard app.
- Scope: mixed
- Date: 2026-05-02

## Findings

### Files found

- `frontend/package.json` - Next.js 16.2.4 App Router app with React 19, Tailwind CSS 3.4, `clsx`, `tailwind-merge`, Zustand, and existing scripts for `dev`, `build`, `lint`, and `typecheck`.
- `frontend/tailwind.config.ts` - Current Tailwind setup uses `darkMode: "class"`, `content: ["./src/**/*.{ts,tsx}"]`, custom RGB CSS-variable color tokens, and custom animation/shadow utilities.
- `frontend/src/app/globals.css` - Current global theme is implemented with `:root` and `:root[data-theme="dark"]` RGB variables, body background effects, shared component classes, and `prefers-reduced-motion`.
- `frontend/src/app/layout.tsx` - App Router root layout sets `<html className="dark" data-theme="dark" data-theme-mode="dark">`, runs the existing preference init script, mounts shared header, skip link, and toast provider.
- `frontend/src/components/ui/Button.tsx` - Existing local UI components already follow a shadcn-like local-copy pattern using `React.forwardRef`, `cn`, variants, and Tailwind class composition, but variants are custom and not `class-variance-authority` based.
- `frontend/src/features/admin/AdminShell.tsx` - Current admin dashboard layout is a hand-built sidebar/card shell with route-aware nav items, session verification, and inline SVG icons.
- `frontend/src/features/admin/SettingsForm.tsx` - Current admin settings flow is a large client component with manual form state, selects, inputs, status notices, toasts, and storage-config CRUD.
- `frontend/src/features/admin/ImageTable.tsx` - Current image admin view combines grid/list modes, search debounce, selection, batch delete, preview, and a native HTML table.

### Official shadcn/ui references

- Next.js installation: https://ui.shadcn.com/docs/installation/next
- `components.json` configuration: https://ui.shadcn.com/docs/components-json
- Theming and CSS variables: https://ui.shadcn.com/docs/theming
- Button: https://ui.shadcn.com/docs/components/button
- Card: https://ui.shadcn.com/docs/components/card
- Form: https://ui.shadcn.com/docs/components/form
- Input: https://ui.shadcn.com/docs/components/input
- Select: https://ui.shadcn.com/docs/components/select
- Table: https://ui.shadcn.com/docs/components/table
- Data Table: https://ui.shadcn.com/docs/components/data-table
- Sidebar: https://ui.shadcn.com/docs/components/sidebar
- Dropdown Menu: https://ui.shadcn.com/docs/components/dropdown-menu
- Tabs: https://ui.shadcn.com/docs/components/tabs
- Badge: https://ui.shadcn.com/docs/components/badge
- Dialog: https://ui.shadcn.com/docs/components/dialog
- Sonner: https://ui.shadcn.com/docs/components/sonner
- Dashboard block examples: https://ui.shadcn.com/blocks

### Official setup pattern for Next.js

The official Next.js guide initializes shadcn/ui in an existing Next.js project with the CLI from the app directory. For this repo, the command should be run from `frontend/`, not the repo root, because `package.json`, `tailwind.config.ts`, `tsconfig.json`, and `src/app` live under `frontend/`.

The official install flow expects the CLI to detect framework and Tailwind settings, then create/update:

- `components.json`
- Tailwind config and/or global CSS theme variables
- `src/lib/utils.ts`
- `src/components/ui/*`

Repo mapping:

- `frontend/src/lib/utils.ts` already exists and is imported by current UI components.
- `frontend/tsconfig.json:20-25` already defines the `@/*` alias expected by the default shadcn layout.
- `frontend/tailwind.config.ts:6-7` already enables class-based dark mode and scans `./src/**/*.{ts,tsx}`.
- `frontend/src/app/globals.css:1-3` already imports Tailwind layers and is the correct location for shadcn base variables.

Practical implication: initialize shadcn/ui inside `frontend/` and review generated diffs carefully. The CLI may try to update files that already contain project-specific theme and utility choices.

### Component conventions from official docs

Official shadcn/ui components are copied into the app source tree instead of consumed as a black-box package. The app owns the generated component files and can edit them. The common convention is:

- shared primitives in `src/components/ui`
- `cn` helper from `@/lib/utils`
- Tailwind classes as the styling surface
- Radix UI primitives for accessible interactive components such as select, dialog, dropdown menu, and tabs
- Lucide icons in examples and blocks
- variant-capable components such as Button implemented with `class-variance-authority`

Repo mapping:

- Current local primitives already live in `frontend/src/components/ui`, so the refactor can be incremental rather than a wholesale folder move.
- Existing imports use PascalCase filenames such as `@/components/ui/Button`; official generated files use lowercase names such as `@/components/ui/button`. Pick one import style for the refactor and apply it consistently.
- Existing components can either be replaced by generated official primitives or wrapped temporarily to avoid touching every caller in one pass.

### Theming and CSS variables

Official shadcn/ui theming centers component colors on semantic CSS variables such as `--background`, `--foreground`, `--card`, `--card-foreground`, `--popover`, `--primary`, `--primary-foreground`, `--secondary`, `--muted`, `--accent`, `--destructive`, `--border`, `--input`, and `--ring`. The Tailwind config maps those variables to utilities such as `bg-background`, `text-foreground`, `bg-card`, `border-border`, and `ring-ring`.

Repo mapping:

- Current globals use project-specific RGB variables such as `--color-surface`, `--color-ink`, `--color-panel`, `--color-muted`, and `--color-accent`.
- Current Tailwind config maps those to custom utilities like `bg-surface`, `text-ink`, `bg-panel`, `text-muted`, and `bg-accent`.
- A shadcn refactor should add official semantic variables while preserving the existing preference system, rather than deleting project tokens in the first pass.
- The existing `data-theme="dark"` dark-mode switch in `frontend/src/app/layout.tsx:19` can coexist with class-based dark mode, but shadcn defaults commonly target `.dark`. Either mirror variables under both `.dark` and `:root[data-theme="dark"]`, or ensure the preference sync always toggles the `.dark` class.

Suggested mapping:

- `--background` from current `--color-surface`
- `--foreground` from current `--color-ink`
- `--card` from current `--color-panel`
- `--card-foreground` from current `--color-ink`
- `--primary` from current `--color-accent`
- `--primary-foreground` from current `--color-accent-foreground`
- `--muted-foreground` from current `--color-muted`
- `--border` from current `--color-border`
- `--destructive` from current `--color-danger`
- `--destructive-foreground` from current `--color-danger-foreground`

### Dashboard and navigation patterns

Official shadcn dashboard examples and blocks are built from composable primitives rather than one monolithic dashboard component. Relevant patterns for this app:

- `Sidebar` for admin navigation and responsive shell behavior
- `Card` for metrics, status summaries, and grouped admin panels
- `Table` or `DataTable` for image-management rows
- `DropdownMenu` for row actions, account/session actions, view options, and destructive actions
- `Tabs` for grouped settings panels or public/admin view modes where the UI is truly tabular
- `Badge` for storage backend, readiness state, token state, and file type labels
- `Dialog` for delete confirmations, image metadata preview actions, and storage-config destructive flows

Repo mapping:

- `AdminShell.tsx` is the natural first target for `Sidebar`, because it already owns route-aware admin navigation.
- `DashboardOverview.tsx` can move status/metric panels to official `Card`, `Badge`, and possibly `Separator`.
- `ImageTable.tsx` can use `Table` first, then optionally `DataTable` if sorting/filtering/column visibility becomes a requirement.
- Avoid importing dashboard block code verbatim unless it matches the app route/session model; treat official blocks as composition examples.

### Forms and settings patterns

The official Form component is designed around `react-hook-form` and schema validation in the examples. For this repo, adopting shadcn form patterns has dependency and migration implications.

Repo mapping:

- `SettingsForm.tsx` currently uses manual React state and backend API payload builders.
- If the implementation wants official `Form`, expect to add `react-hook-form` and a schema validator such as `zod`, then map storage config create/update payloads through typed form schemas.
- A lower-risk first pass is to use official `Input`, `Label`, `Select`, `Switch`, `Textarea`, `Button`, `Card`, and `Alert` primitives while preserving existing submit handlers and payload builders.
- Official `Select` is Radix-based and better suited for storage backend, active storage config, language/theme controls, and upload storage target selection than the current hand-built select.

### Table and image-admin patterns

Official Table is a styled semantic table primitive. Official Data Table examples build richer tables with TanStack Table patterns for sorting, filtering, visibility, selection, and row actions.

Repo mapping:

- `ImageTable.tsx` already has selection, search debounce, batch delete, pagination-ish page state, and grid/list toggle.
- Replace the native table markup with official `Table`, `TableHeader`, `TableBody`, `TableRow`, `TableHead`, and `TableCell` first.
- Consider Data Table only if the refactor scope includes column sorting, column visibility, row action menus, and a stronger table abstraction.
- Keep the existing gallery/grid image-card mode; shadcn table patterns should not force the app into table-only administration.

### Toasts and feedback

Official shadcn/ui currently documents `sonner` as the toast component. This repo uses `react-hot-toast` in `frontend/src/app/layout.tsx:3` and across admin/upload flows.

Repo mapping:

- Replacing toast infrastructure is optional and cross-cutting.
- If switching to official `sonner`, update root layout provider plus all toast calls in one coordinated pass.
- If not switching, shadcn primitives can still coexist with `react-hot-toast`.

### Mapping onto the existing Tailwind app

Recommended migration order:

1. Initialize shadcn/ui in `frontend/` and generate `components.json`.
2. Add official semantic theme variables while keeping existing project variables.
3. Generate only the primitives needed for the first slice: `button`, `card`, `input`, `label`, `select`, `badge`, `table`, `separator`, `dropdown-menu`, `dialog`, and `sidebar` if the admin shell is in scope.
4. Replace local UI primitives with official versions or create compatibility wrappers for existing imports.
5. Convert admin shell/navigation and one admin page at a time.
6. Run `npm run lint`, `npm run typecheck`, and `npm run build` from `frontend/`.

Best first implementation slice for this task:

- Theme bridge in `globals.css` and `tailwind.config.ts`.
- Official `Button`, `Card`, `Input`, `Badge`, and `Select` primitives.
- Admin shell/sidebar plus settings form controls.
- Image table primitive conversion without introducing full Data Table complexity.

### Related specs

- `.trellis/spec/frontend/index.md` - frontend pre-development checklist.
- `.trellis/spec/frontend/directory-structure.md` - expected route, feature, and shared component layout.
- `.trellis/spec/frontend/component-guidelines.md` - component boundary and client/server split rules.
- `.trellis/spec/frontend/hook-guidelines.md` - custom hook and data-fetching rules.
- `.trellis/spec/frontend/state-management.md` - Zustand and persisted preference rules.
- `.trellis/spec/frontend/type-safety.md` - TypeScript and API contract expectations.
- `.trellis/spec/frontend/quality-guidelines.md` - accessibility, lint, build, and verification expectations.

## Caveats / Not Found

- Official shadcn/ui examples are moving targets; verify CLI prompts and generated files at implementation time.
- The official docs and blocks are examples, not a complete app architecture. Do not copy dashboard blocks blindly over OmePic session verification, admin routing, i18n, or storage contracts.
- The current repo uses Tailwind CSS 3.4.17. If the official CLI defaults have shifted toward a newer Tailwind setup, keep the generated configuration aligned with this repo unless the PRD explicitly asks for a Tailwind upgrade.
- No current app source uses official shadcn/ui primitives yet. Existing `frontend/src/components/ui/*` files are custom local primitives with similar naming, so filename/import casing and compatibility wrappers need explicit planning.
- Official Form examples imply `react-hook-form` and schema validation dependencies. That is useful for large settings forms, but it is a bigger behavioral refactor than swapping presentational primitives.
