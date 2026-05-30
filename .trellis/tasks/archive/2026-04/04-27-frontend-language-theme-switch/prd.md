# Add Frontend Language And Theme Switching

## Goal

Add global frontend preferences for bilingual UI (`zh` / `en`) and theme mode (`light` / `dark` / `system`) across the existing Next.js app without introducing new runtime dependencies or route restructuring.

## What I already know

- The frontend uses Next.js App Router with a single root layout in `frontend/src/app/layout.tsx`.
- Global chrome currently comes from `frontend/src/components/shared/AppHeader.tsx`.
- Styling is driven by `frontend/src/app/globals.css` plus Tailwind utility classes.
- There is no existing i18n library, theme library, or global UI-preference store.
- The app already uses Zustand for other client state, but most visible UI copy is hardcoded in components/pages.
- Current root `<html>` is hardcoded to `lang="en"` and the CSS root hardcodes `color-scheme: light`.

## Requirements

- The frontend must support switching visible UI language between Chinese and English.
- The frontend must support switching theme mode between `light`, `dark`, and `system`.
- Language and theme preferences must apply globally across public and admin pages.
- Preferences must persist across reloads in the browser.
- Initial language may default from browser language detection when no saved preference exists.
- Initial theme may default to `system` when no saved preference exists.
- `system` theme mode must follow the current OS/browser color-scheme preference.
- The active document language should reflect the selected UI language.
- Do not add third-party dependencies for i18n or theme switching.
- Keep the implementation lightweight and codebase-consistent with the existing repo patterns.

## Acceptance Criteria

- [ ] Header exposes controls to switch language (`中文` / `English`) and theme (`Light` / `Dark` / `System`).
- [ ] The selected language updates visible UI copy on the current page without a full page reload.
- [ ] The selected theme updates the UI on the current page without a full page reload.
- [ ] Public pages and admin pages read the same shared preferences.
- [ ] Reloading the browser preserves both preferences.
- [ ] With no saved theme preference, the UI follows the system theme.
- [ ] With no saved language preference, the UI chooses a sensible browser-language default.
- [ ] The root document language updates to `zh-CN` or `en` to match the active UI language.
- [ ] Frontend lint, typecheck, and build pass after the change.

## Out Of Scope

- Server-side locale routing or localized URLs
- Backend-driven translation management
- More than two UI languages
- User-profile syncing of preferences across devices

## Technical Approach

- Introduce a small shared frontend preference layer for `language` and `theme`.
- Store preferences in browser storage and hydrate them on the client.
- Apply theme mode at the document/root level using a stable attribute or class strategy that works with Tailwind and existing global CSS.
- Add a minimal translation dictionary and helper hook/function for the existing visible UI strings.
- Update the key public/admin surfaces so all user-facing copy shown in the current app can switch between Chinese and English.
- Keep the preference controls in the shared header so they affect the whole app from one place.

## Notes

- Prefer a small in-repo solution over pulling in `next-themes` or a full i18n framework.
- Avoid hydration flicker as much as practical for theme application.
