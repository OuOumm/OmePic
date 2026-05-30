# Implementation Plan: Frontend Svelte 5 audit remediation

## Ordered checklist

1. Re-read frontend specs and audit doc before editing.
2. Remove legacy files:
   - `frontend/next-env.d.ts`
   - `frontend/eslint.config.mjs`
   - `frontend/src/lib/actions/click-outside.ts`
   - clean `.next/**` ignore from `frontend/eslint.config.js` if safe.
3. Implement admin gate in `frontend/src/routes/admin/dashboard/+layout.svelte`.
4. Fix announcement detail reset in `AnnouncementDialog.svelte`.
5. Add MIME/accept helpers and wire `CanvasDropzone` / home route.
6. Refactor `ImageDetailDrawer.svelte` IP-detail fetch to abort stale requests.
7. Reduce theme-script duplication and add system theme listener; update helper tests.
8. Add i18n keys and replace image detail hardcoded labels.
9. Extend `safeImageUrl()` with runtime public URL origin allow-list and update preview/table call sites.
10. Update tests:
    - utility tests for MIME accept and safe public URL allow-list
    - theme helper/script tests as needed
    - API/admin gate tests if lightweight helper extraction is introduced
11. Update `docs/debug/frontend-svelte5-audit-remediation-2026-05-13.md` with completion status.
12. Run validation commands.

## Validation commands

Run from repository root using Windows `cmd.exe` invocation because direct bash `npm` is broken in this environment:

```bash
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run lint"
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run typecheck"
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run test"
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run build:backend"
```

## Risky files / rollback points

- `frontend/src/routes/admin/dashboard/+layout.svelte`: rollback if admin route rendering loops or login entry is hidden.
- `frontend/src/lib/utils.ts`: rollback URL allow-list or MIME helper changes if tests reveal contract drift.
- `frontend/src/lib/components/studio/AppShell.svelte`: rollback system theme listener if it causes hydration issues.
- Generated `backend/web/` changes may appear after build; do not manually edit generated output.

## Review gates before start

- PRD, design, and implementation plan exist.
- User already approved implementing the remediation doc.
