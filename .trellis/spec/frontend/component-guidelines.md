# Component Guidelines

> Component rules for the current SvelteKit + Svelte 5 frontend.

---

## Current State

- Shared UI components live under `frontend/src/lib/components/studio/`.
- Route pages live under `frontend/src/routes/` and should orchestrate data loading, local state, and component composition.
- Components use Svelte component files, Svelte 5 runes where needed, Tailwind utility classes, and shared `studio-*` CSS classes from `frontend/src/app.css`.
- Do not apply React Client Component, shadcn React, or Next.js component patterns to the active frontend.

---

## Component Boundaries

- Route components own page-level orchestration:
  - loading server data through `frontend/src/lib/api.ts`
  - holding page-local form state
  - handling route-level success/error state
  - composing shared studio components
- Studio components own reusable UI behavior and markup:
  - `AppShell.svelte`
  - `CanvasDropzone.svelte`
  - `ImageDataTable.svelte`
  - `ImagePreviewDialog.svelte`
  - `ImageDetailDrawer.svelte`
  - `ImageSwitchButton.svelte`
  - `BanIPDialog.svelte`
  - `IPDetailPanel.svelte`
  - `StorageInspector.svelte`
  - `StorageInstanceManager.svelte`
  - `AnnouncementManager.svelte`
  - `AnnouncementDialog.svelte`
  - `ConfirmDialog.svelte`
  - `ToastViewport.svelte`
  - `MarkdownContent.svelte`
  - `PageTitle.svelte`
  - `MetricStrip.svelte`
  - `LineChart.svelte`
- Svelte actions own reusable DOM behavior:
  - `accessible-dialog.ts` — focus trap and ARIA attributes for modal dialogs
  - `viewport-portal.ts` — body-level modal portal attachment for viewport-wide fixed backdrops
- API calls should remain centralized in `frontend/src/lib/api.ts` and be passed into components through route orchestration or component callbacks where appropriate.

---

## UI Composition Rules

- Prefer project studio classes such as `studio-panel`, `studio-button`, `studio-input`, and `studio-table-row` for consistency.
- Reuse existing studio components before creating a new component.
- Keep destructive actions behind an explicit confirmation flow.
- Use shared toast state for success/error feedback, but do not rely on toast alone for destructive confirmations.
- Use accessible labels for icon-only buttons.
- Use `loading`, `empty`, and `error` states for async tables and panels.
- For image-heavy UI, use lazy loading where appropriate and eager loading only for focused previews.

---

## Image Preview and Details

- Public upload/history previews use `ImagePreviewDialog.svelte`.
- Admin image details use `ImageDetailDrawer.svelte` because it includes admin-only actions such as delete, IP detail loading, and IP ban creation.
- Admin image tables use `ImageDataTable.svelte` for list/grid display concerns.
- Do not duplicate image detail metadata rendering in route pages when an existing studio component already owns it.

---

## Admin Security Components

- Use `BanIPDialog.svelte` for creating an IP ban from a UID or IP address.
- Use `IPDetailPanel.svelte` for IP-specific abuse details.
- Admin security components should consume typed payloads from `frontend/src/lib/types/index.ts`.
- Admin security route pages should call typed helpers from `frontend/src/lib/api.ts` rather than embedding fetch calls.

---

## Component Props

- Keep props explicit and typed with `type Props = { ... }` and `$props()` for larger Svelte components.
- Prefer callback props for route-owned mutations, for example `onDeleted`, `onClose`, `onNavigate`, or `onConfirm`.
- Keep presentational components free of unrelated persistence concerns.
- For language (`preferences.language`): any component may read it directly, since nearly all components need it for `t()` translation calls. Passing it through props would add boilerplate with no practical benefit.
- For admin token (`preferences.adminToken`): prefer reading `preferences` only in components that already sit at an admin-specific boundary (e.g. admin dashboard layout, admin route pages). Pass the token through callbacks or props to deeper presentational components.
- For other global values: pass through props when the component is reused in varied contexts; read directly from `preferences` when the component is always rendered inside the app shell.

---

## Anti-patterns

- Creating new React/TSX components in the active Svelte frontend.
- Putting duplicated fetch logic inside multiple Svelte components.
- Building another table/detail/preview component for admin images without first extending `ImageDataTable.svelte` or `ImageDetailDrawer.svelte`.
- Hardcoding visible user-facing copy without updating both language dictionaries.
- Adding unlabelled icon-only buttons.
- Calling destructive admin APIs without confirm UI.
- Rendering a `fixed inset-0` modal/backdrop directly inside route or layout containers without `attachViewportPortal()`; container transforms, overflow, or stacking contexts can make the backdrop cover only part of the viewport.

---

## Scenario: Viewport-Wide Modal Backdrops

### 1. Scope / Trigger

- Trigger: any shared dialog, drawer, or modal renders a viewport overlay/backdrop such as `fixed inset-0`.
- Affected components live under `frontend/src/lib/components/studio/` and may be rendered from public, history, or admin routes with different layout containers.

### 2. Signatures

- Action: `frontend/src/lib/actions/viewport-portal.ts`
- Import: `import { attachViewportPortal } from '@/actions/viewport-portal';`
- Usage on the overlay root: `{@attach attachViewportPortal()}`

### 3. Contracts

- Modal overlay roots with `fixed inset-0` must attach `attachViewportPortal()` so the DOM node is moved to `document.body` at runtime.
- Keep `attachAccessibleDialog()` on the focusable dialog element or the same modal root as appropriate; portal attachment does not replace focus trapping or `aria-modal`.
- The action must remain SSR-safe by checking for `document` before touching `document.body`.
- Existing overlay z-index, click-to-dismiss, Esc behavior, and backdrop styling should remain owned by the component.

### 4. Validation & Error Matrix

- Missing portal in nested/admin route -> backdrop may be clipped by route shell, overflow, transform, or containment styles.
- Portal action touches `document` during SSR -> SvelteKit build/typecheck risk.
- Portal replaces accessible-dialog behavior -> keyboard/focus regression.

### 5. Good/Base/Bad Cases

- Good: a shared modal root uses `fixed inset-0` plus `{@attach attachViewportPortal()}` and keeps the inner dialog focus trap.
- Base: a non-modal inline panel does not use the portal because it is not a viewport overlay.
- Bad: fixing one route with local CSS while other shared dialogs remain nested under constrained containers.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend` for modal/backdrop changes.
- Code-review scan all `fixed inset-0` modal roots and confirm portal attachment or an explicit non-modal reason.
- Manually verify at least one public/history route and one admin route when browser access is available.

### 7. Wrong vs Correct

#### Wrong

```svelte
<div class="fixed inset-0 z-50 bg-black/50">
  <section role="dialog" aria-modal="true">...</section>
</div>
```

#### Correct

```svelte
<script lang="ts">
  import { attachViewportPortal } from '@/actions/viewport-portal';
  import { attachAccessibleDialog } from '@/actions/accessible-dialog';
</script>

<div class="fixed inset-0 z-50 bg-black/50" {@attach attachViewportPortal()}>
  <section role="dialog" aria-modal="true" tabindex="-1" {@attach attachAccessibleDialog(() => ({ onClose }))}>...</section>
</div>
```


## Scenario: Admin Image Details With Security Actions

### 1. Scope / Trigger

- Trigger: admin image preview/details includes delete, IP detail, and IP-ban operations.

### 2. Signatures

- Component: `frontend/src/lib/components/studio/ImageDetailDrawer.svelte`
- Props:
  - `image: AdminImage | null`
  - `images?: AdminImage[]`
  - `onClose: () => void`
  - `onDeleted: () => void`
  - `onNavigate?: (image: AdminImage) => void`
- APIs used through `frontend/src/lib/api.ts`:
  - `adminDeleteImages(token, [uid])`
  - `adminCreateIPBan(token, input)`
  - `adminGetAbuseIPDetail(token, ip)`

### 3. Contracts

- Detail drawer shows image metadata and admin-only actions.
- Delete action must show `ConfirmDialog` before calling the API.
- On successful delete, call `onDeleted()` so the route refreshes its table data.
- If neighboring images exist, navigate to the next/previous image after delete; otherwise close the drawer.
- IP ban action should not be offered as active when the image IP is already banned.
- Arrow-left and arrow-right navigation should work when `images` and `onNavigate` are provided.

### 4. Validation & Error Matrix

- No admin token -> do not call admin APIs.
- `image === null` -> render nothing.
- IP detail load failure -> show error toast and keep drawer usable.
- Delete failure -> keep drawer open and show error toast.
- Ban creation failure -> keep dialog state recoverable and show error toast.

### 5. Good/Base/Bad Cases

- Good: route owns data refresh, drawer owns focused image actions and calls `onDeleted()` after successful deletion.
- Base: drawer can render one image without navigation.
- Bad: route page duplicates drawer metadata and deletes without confirmation.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend`.
- Verify admin image deletion from drawer refreshes table data.
- Verify IP-ban action updates displayed ban state.
- Verify keyboard navigation works when multiple images are passed.

### 7. Wrong vs Correct

#### Wrong

```svelte
<button onclick={() => adminDeleteImages(token, [image.uid])}>Delete</button>
```

#### Correct

```svelte
<button onclick={() => (deleteOpen = true)}>{t(language, 'common.delete')}</button>
<ConfirmDialog open={deleteOpen} onConfirm={remove} />
```

---

## Scenario: Public Announcement Acknowledgement

### 1. Scope / Trigger

- Trigger: public announcements must keep reappearing until the visitor explicitly acknowledges the latest announcement.

### 2. Signatures

- Component: `frontend/src/lib/components/studio/AnnouncementDialog.svelte`
- Props:
  - `open: boolean`
  - `onClose: () => void`
  - `onAcknowledge: () => void`
- Route state: `frontend/src/routes/+page.svelte`
- Storage key: `omepic:announcement:lastSeen`

### 3. Contracts

- `onClose` only closes the visible dialog; it must not write `omepic:announcement:lastSeen`.
- `onAcknowledge` is the only callback that writes the latest announcement timestamp to `omepic:announcement:lastSeen`.
- Overlay click, close button, and Esc dismissal use `onClose`.
- The "got it" action uses `onAcknowledge`.
- Manual announcement history viewing must not accidentally mark a newly published announcement as seen unless the user clicks the acknowledgement action.

### 4. Validation & Error Matrix

- No announcements -> do not open the dialog.
- Latest announcement timestamp equals stored `lastSeen` -> do not auto-open.
- Latest announcement timestamp differs from stored `lastSeen` -> auto-open in detail mode.
- Dismiss without acknowledgement -> keep stored `lastSeen` unchanged so the dialog can appear again.
- Acknowledge -> update `lastSeen` and close the dialog.

### 5. Good/Base/Bad Cases

- Good: the visitor clicks the close icon and refreshes; the latest announcement opens again.
- Base: the visitor clicks "got it"; the same announcement no longer auto-opens.
- Bad: wiring the primary acknowledgement button to the same callback used by overlay dismissal.

### 6. Tests Required

- Run `npm run lint`, `npm run typecheck`, and `npm run build:backend`.
- Verify closing by overlay, close button, or Esc does not update `omepic:announcement:lastSeen`.
- Verify clicking the acknowledgement button writes the latest timestamp.
- Verify a newer announcement opens even after an older announcement was acknowledged.

### 7. Wrong vs Correct

#### Wrong

```svelte
<button onclick={onClose}>{t(language, 'announcement.gotIt')}</button>
```

#### Correct

```svelte
<button onclick={onAcknowledge}>{t(language, 'announcement.gotIt')}</button>
```
