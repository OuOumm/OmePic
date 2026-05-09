# fix frontend homepage styling

## Goal

Repair the homepage upload screen so it feels structurally coherent again instead of visually scattered. Preserve all upload/history/business behavior while improving layout hierarchy, responsive balance, and shared visual consistency with the existing OmePic frontend language.

## What I already know

* The homepage route `frontend/src/app/page.tsx` renders `frontend/src/features/upload/UploadPageClient.tsx`.
* The current design system is shared across public/admin pages through glass panels, slate surfaces, violet-to-cyan accents, and shared primitives such as `Button`, `Card`, `Badge`, `Select`, and `PageLayout`.
* The current homepage mixes a large hero, storage selection card, dropzone, token card, status card, result card, and recent uploads grid, but the hierarchy is noisy and the page feels over-segmented.
* The repo already contains an accepted visual reference in `img.html` for the current frontend direction.
* There are invalid Tailwind utility usages in shared UI (`px-4.5`, `h-4.5`, `w-4.5`) that can cause missing styles and contribute to a visually broken header/button presentation.

## Assumptions (temporary)

* No backend contract, routing, storage-selection behavior, or upload-state behavior should change.
* Existing locale/theme support must remain intact.
* The requested fix is limited to homepage and directly related shared visual primitives, not a full-site redesign.

## Open Questions

* None blocking. The request is specific enough to proceed with a focused homepage polish task.

## Requirements (evolving)

* Rework the homepage upload page so the primary action is visually clear and the page no longer feels cluttered or misaligned.
* Preserve existing homepage capabilities: storage target selection, upload dropzone, status/progress feedback, latest upload result, client token visibility, and recent uploads.
* Keep the visual direction aligned with the current project language from `img.html` and shared frontend primitives instead of inventing a new style system.
* Improve responsive behavior so the homepage remains readable and balanced on narrow screens and wide desktop layouts.
* Fix directly related shared style defects that make the homepage/header appear broken, including invalid utility-class usage when encountered.
* Keep accessibility intact for upload controls, status/error feedback, and keyboard navigation.

## Acceptance Criteria (evolving)

* [ ] Homepage has a clear visual hierarchy on mobile and desktop, with the upload area remaining the focal point.
* [ ] Storage selection, upload status, result card, and recent uploads are grouped more coherently and do not appear visually chaotic.
* [ ] Shared visual fixes applied for homepage-adjacent primitives do not break existing route behavior.
* [ ] No new hardcoded single-language copy is introduced outside the translation dictionary contract.
* [ ] `npm run lint`, `npm run typecheck`, and `npm run build` pass in `frontend/`.

## Definition of Done (team quality bar)

* Tests added/updated where appropriate
* Lint / typecheck / CI green
* Docs/notes updated if behavior changes
* Rollout/rollback considered if risky

## Out of Scope (explicit)

* Backend API or storage behavior changes
* Rewriting history, API, or admin pages beyond incidental shared primitive alignment
* Introducing new homepage product features unrelated to layout/style repair

## Technical Notes

* Primary implementation target: `frontend/src/features/upload/UploadPageClient.tsx`
* Likely supporting files: `frontend/src/components/ui/Button.tsx`, `frontend/src/components/shared/AppHeader.tsx`, and other homepage-adjacent shared UI only if needed
* Relevant references: `.trellis/spec/frontend/*`, `img.html`, and the homepage audit under `research/homepage-audit.md`
