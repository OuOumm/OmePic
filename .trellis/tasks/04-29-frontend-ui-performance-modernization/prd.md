# Frontend UI and Performance Modernization

## Goal

Refactor the OmePic frontend into a more modern, product-grade web app while preserving the current upload, history, API documentation, admin auth, admin image management, storage configuration, language/theme preference, and selected-storage upload contracts. Improve visual hierarchy, component craftsmanship, interaction states, and practical frontend performance without introducing a new design system dependency.

## UI Aesthetics Route

- Route: **Refactor + Component Polish + State/Motion Refinement**.
- Visual thesis: match `D:\Works\MyProject\OmePic\img.html`: dark-default minimal glassmorphism with slate/cool-gray structure, violet-to-cyan gradient accents, translucent surfaces, soft blur, fine borders, and calm motion.
- Change level: **full visual rebuild, stable architecture**. Keep the existing Next.js/Tailwind app, routes, feature boundaries, API contracts, Zustand stores, and storage-selection behavior, but restyle every frontend page and shared primitive to the `img.html` direction.

## What I Already Know

- User explicitly invoked `$ui-aesthetics`.
- User revised the brief to follow `D:\Works\MyProject\OmePic\img.html` as the visual reference for all frontend pages.
- User follow-up: **card sections must fully imitate `img.html`**, not just broadly match the palette. Image/recent/history/admin-grid cards should copy the reference card anatomy and interaction states closely.
- The visual prompt requires minimalism + soft glassmorphism, dark mode as the default, light-mode switching support, violet/cyan gradients, slate neutrals, heroicons-style outline icon treatment, smooth transitions, hover lift/scale, drag pulse states, skeleton-style loading, glass toolbars, image-card overlays, and modern admin/album layouts.
- User wants frontend refactor, frontend performance optimization, and a newer modern UI design.
- The installed `ui-aesthetics` skill exposes only `SKILL.md` in this environment; shared reference files are unavailable, matching prior memory for this checkout.
- Current frontend stack is Next.js 16.2.4 App Router, React 19, TypeScript, Tailwind CSS, Zustand, and react-hot-toast.
- Current UI already has shared primitives under `frontend/src/components/ui/` and shared app shell/header components.
- Current global palette is heavily warm/beige/amber, which risks reading as one-note.
- Current pages are functional but visually basic:
  - upload page has a dropzone, storage selector, progress/result panel, token panel, and recent uploads
  - history page lists IndexedDB uploads
  - API page renders endpoint cards and curl snippets
  - admin shell has status, images, settings, and login flows
- Current selected-storage upload contract must remain intact:
  - `GET /v1/storage-options`
  - optional multipart `storage_key` on upload
  - returned `storage_key` / `storage_backend`
  - history persists upload storage metadata
- Current `useUiPreferencesStore` uses Zustand `persist`; avoid inline object selectors because Zustand v5 selector instability caused previous runtime loops.
- `git status` is unavailable because this workspace has no `.git`.

## Requirements

- Preserve all frontend routes and external API contracts.
- Preserve language and theme switching across public and admin pages.
- Preserve selected-storage upload behavior from the current task.
- Modernize global visual system to match `img.html`:
  - dark mode is the first-run/default presentation while keeping existing theme switch architecture
  - use slate/cool-gray surfaces with violet-to-cyan gradients for primary action, logo, active states, badges, progress, and selected states
  - use translucent glass panels (`bg-white/60`, `dark:bg-slate-900/60` direction), `backdrop-blur-xl`, fine white/slate borders, and soft shadows
  - keep light and dark themes coherent through the existing root preference sync
  - keep text contrast and focus states clear
- Improve shared primitives:
  - `Button`, `Card`, `Input`, and form/select/table treatments should have consistent radii, borders, focus, disabled, hover, and active states
  - avoid oversized marketing composition; this is an operational tool
  - avoid nested card-heavy layouts where a simpler section/group works better
- Refactor main pages for clearer hierarchy:
  - global header: fixed glass navigation with gradient logo, responsive mobile menu, preference controls, upload action, and avatar/admin affordance
  - upload page: large dashed glass dropzone, drag violet glow/pulse, storage target card, progress/result panel, token panel, and recent uploads as image cards
  - history page: album-style local image cards with glass metadata/action areas and a strong empty state
  - API page: glass endpoint cards, gradient method badges, readable dark code blocks, and copy-friendly spacing
  - admin shell: glass side navigation, status cards, login card, image management toolbar with search/view controls, selected-state glow, and bottom bulk-action bar
  - storage settings: named storage instances as glass selectable cards and a dense but polished glass form
- Card fidelity requirement:
  - image cards should use the `img.html` anatomy: `group relative rounded-2xl overflow-hidden bg-white dark:bg-slate-800/60 border border-slate-200/60 dark:border-slate-700/40 shadow-sm hover:shadow-xl dark:hover:shadow-violet-500/10 transition-all duration-300 hover:-translate-y-1 cursor-pointer animate-fade-in`
  - image area should be `aspect-square overflow-hidden relative`
  - images should be `w-full h-full object-cover transition-transform duration-400 group-hover:scale-105`
  - hover overlay should fade from transparent to `bg-black/40` and reveal a centered magnifier icon with scale/opacity transition
  - bottom metadata should be an absolute glass strip: `absolute bottom-0 left-0 right-0 px-3 py-2 bg-white/60 dark:bg-slate-900/60 backdrop-blur-md border-t border-white/20 dark:border-slate-700/30 flex items-center justify-between text-xs`
  - do not add large always-visible details below these image cards unless the page needs separate action rows; keep the visible card silhouette close to `img.html`
- Image preview requirement:
  - clicking an image card opens an enlarged image preview/lightbox without navigating away
  - the preview should preserve card context with concise metadata/actions, support close via overlay click, close button, and `Escape`, and keep focus/ARIA behavior accessible
  - reuse the same preview pattern for recent upload, history, and admin image grids where image cards are clickable
- Improve frontend performance:
  - remove duplicated client fetch paths where practical
  - avoid unnecessary rerenders from unstable callbacks/selectors
  - keep server-state fetches localized and cancel-safe
  - avoid adding heavy dependencies
  - preserve or improve `npm run build` output without runtime warnings
- Improve accessibility:
  - visible focus states
  - semantic buttons/links
  - labels and `aria-describedby` for controls
  - status/error messaging not only via toast
  - `prefers-reduced-motion` coverage if motion is added
- Keep responsive layouts polished on mobile and desktop.

## Acceptance Criteria

- [ ] Public upload page has a modernized workflow layout and keeps selected-storage upload behavior.
- [ ] History page, API page, admin login, admin dashboard, image management, and storage settings use the updated visual system.
- [ ] The UI follows `img.html`'s minimal glassmorphism style across public, API, history, admin login, dashboard, image management, and settings pages.
- [ ] Recent upload, history, and admin image grid cards fully match the `img.html` card anatomy: square image, rounded-2xl card, white/slate card surface, light border, subtle shadow, hover lift, black overlay, centered magnifier, and bottom glass metadata strip.
- [ ] Clicking image cards in recent uploads, history, and admin image grids opens an accessible enlarged preview/lightbox and can be dismissed with overlay click, close button, or `Escape`.
- [ ] Dark mode is the default first-run presentation; light and dark themes are both coherent and preserve existing preference switching.
- [ ] Shared UI primitives provide consistent focus, hover, active, disabled, and danger states.
- [ ] No frontend API contract regressions for upload/history/admin/storage settings.
- [ ] Public storage selector remains accessible and does not lose current `aria-*` safeguards.
- [ ] Frontend performance is improved by removing obvious duplicate effects/fetches and avoiding unstable selectors/callbacks.
- [ ] `npm run lint`, `npm run typecheck`, and `npm run build` pass in `frontend/`.
- [ ] If backend files are untouched, backend tests are not required; if any backend contract is touched, `go test ./...` must pass.
- [ ] A local frontend dev server is started after implementation and the URL is reported.

## Definition of Done

- Implementation completed through `trellis-implement`.
- Independent `trellis-check` pass completed and fixed findings.
- Relevant frontend specs updated if new UI/performance conventions are introduced.
- Verification results reported clearly.

## Out of Scope

- Backend API redesign.
- New storage backend types.
- Replacing Tailwind or adding a new component library.
- Backend or storage contract changes.
- Full browser-based visual regression suite unless necessary to diagnose a rendering issue.

## Technical Approach

- Use existing Tailwind/CSS-variable design token setup instead of introducing a new theming library.
- Prefer reusable component polish and page-level layout improvements over decorative one-off CSS.
- Keep app-first density: dashboards and tools should remain compact and scannable.
- Treat storage options and admin data as server state, not long-lived global Zustand state.
- Keep Zustand selectors per-field and stable.

## Technical Notes

- Files inspected:
  - `frontend/package.json`
  - `frontend/src/app/globals.css`
  - `frontend/src/app/layout.tsx`
  - `frontend/src/components/shared/AppHeader.tsx`
  - `frontend/src/components/ui/Card.tsx`
  - `frontend/src/components/ui/Button.tsx`
  - `frontend/src/components/ui/Input.tsx`
  - `frontend/src/features/upload/UploadPageClient.tsx`
  - `frontend/src/features/history/HistoryPageClient.tsx`
  - `frontend/src/features/api/ApiPageContent.tsx`
  - `frontend/src/features/admin/AdminShell.tsx`
  - `frontend/src/features/admin/DashboardOverview.tsx`
  - `frontend/src/features/admin/ImageTable.tsx`
  - `frontend/src/features/admin/SettingsForm.tsx`
  - `frontend/src/features/admin/LoginForm.tsx`
  - `frontend/tailwind.config.ts`
- Existing verification commands:
  - `npm run lint`
  - `npm run typecheck`
  - `npm run build`
