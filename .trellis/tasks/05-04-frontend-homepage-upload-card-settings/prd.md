# frontend homepage upload card settings

## Goal

Rebuild the OmePic home upload page into a single-column, full-screen upload-first experience: remove the homepage upload-flow intro/status/result rails, move storage selection into the header action area, keep upload progress and post-upload copy actions in the recent-upload card area, and consolidate client token, language, and theme controls inside a settings popover.

## What I already know

- The homepage route renders `frontend/src/features/upload/UploadPageClient.tsx`.
- Recent uploads are displayed by `frontend/src/features/upload/RecentUploads.tsx` using the shared `ImgStyleImageCard` and `ImageLightbox`.
- Header language and theme preferences are centralized in `frontend/src/components/shared/AppHeader.tsx`, `frontend/src/stores/ui-preferences-store.ts`, and `frontend/src/lib/i18n.ts`.
- Existing uncommitted work already changed admin header links from `/admin/login` to `/admin/dashboard`; preserve that change.
- User wants the homepage to follow a full-height single-column responsive layout with a fixed top nav, a dominant drag-upload core, quick action links, a recent image grid/empty state, and the existing lightbox modal layer.

## Requirements

- Remove the homepage upload-flow introduction/header section entirely.
- Remove the homepage side information rail for upload status, latest result, and client token.
- Move storage selection out of the homepage content and into the navigation action area beside the settings button, as a dropdown selector.
- Make the drag-and-drop upload card the full-screen primary focus below the fixed header, with vertically centered icon/title/help text and decorative gradient/light layers that do not disrupt content flow.
- Keep the main homepage content as a single-column vertical scrolling flow with max-width constraints and mobile-first spacing.
- Add a quick action row beneath the upload area with short text links and a small separator. A paste-URL panel can be toggled, but backend URL-upload behavior is out of scope unless an endpoint already exists.
- Keep the image section title row as icon + title on the left and count badge on the right.
- Recent upload cards must remain a responsive CSS grid: 2 columns by default, 4 columns on medium screens, 5 columns on large screens, with square thumbnails and bottom glass metadata strips.
- Empty state must be mutually exclusive with the image grid.
- Keep the shared full-screen lightbox preview layer for enlarged images.
- Keep upload progress visible by adding progress UI to the top/active card in the recent uploads area while an upload is running.
- After upload success, show three copy buttons directly below the successful recent-upload card: URL, MD, and BB.
- Remove delete-related progress UI from the homepage upload experience.
- Move client token, language switching, and theme switching into a settings button in the navigation bar that opens a settings panel/popover.
- Preserve existing upload, selected storage, IndexedDB recent-history, image preview, copy, language, and theme behavior.
- Keep visible text localized in both English and Chinese.

## Acceptance Criteria

- [ ] Homepage no longer renders standalone cards titled upload status, latest result, or client token.
- [ ] Homepage no longer renders the old upload-flow intro/header section.
- [ ] Storage selection is available in the header next to the settings button as a dropdown.
- [ ] Upload dropzone is the dominant full-screen main card under the fixed nav.
- [ ] Homepage content follows a single-column vertical flow.
- [ ] Upload progress is visible inside the recent-upload section/card during upload.
- [ ] The most recent successful upload card shows URL, MD, and BB copy buttons immediately below it.
- [ ] Header shows a settings button/panel containing client token, language controls, and theme controls.
- [ ] Header language/theme controls are no longer separate always-visible groups outside the settings panel.
- [ ] `npm run lint`, `npm run typecheck`, and `npm run build` pass in `frontend/`.

## Out of Scope

- Backend API changes.
- Upload/delete contract redesign.
- Implementing URL-paste upload if no backend endpoint exists.
- Admin storage management changes.

## Technical Notes

- Relevant specs loaded from `.trellis/spec/frontend/`.
- Visual approach: medium-to-large homepage restructure, reduce competing side panels, keep operational shadcn-like surfaces and shared image-card grammar while making the upload card the first-screen focus.
