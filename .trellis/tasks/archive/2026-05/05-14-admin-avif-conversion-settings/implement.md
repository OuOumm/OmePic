# Implementation Plan

## Checklist

1. Backend runtime settings
   - Add `AvifQuality` / `AvifSpeed` JSON fields to runtime settings structs.
   - Add defaults `60` / `8`.
   - Persist/read config keys `avif_quality` / `avif_speed`.
   - Validate ranges `quality 0..100`, `speed 0..10`.
   - Update runtime settings tests.

2. AVIF conversion
   - Introduce AVIF conversion options/settings type.
   - Pass current runtime settings quality/speed into conversion for new physical uploads.
   - Preserve duplicate short-circuit behavior.
   - Update/add image service tests to assert configured options are used.

3. Frontend admin settings
   - Extend `frontend/src/lib/types/index.ts` `RuntimeSettings`.
   - Add number inputs for AVIF quality/speed in settings runtime/upload policy area.
   - Add English/Chinese i18n labels and hints.
   - Update API tests fixtures.

4. Docs/spec
   - Update `.trellis/spec/backend/runtime-settings.md` and frontend type spec if needed.
   - Update docs/README/API docs where runtime settings payload is shown.

5. Validation
   - `cd backend && go test ./...`
   - `cd frontend && npm run typecheck`
   - `cd frontend && npm test -- --run src/lib/api.test.ts src/lib/ui-errors.test.ts`
   - `cd frontend && npm run build:backend`
   - `git diff --check`

## Risk / Rollback

- AVIF speed semantics are inverse-intuitive: smaller is slower/better compression. UI hint should say this clearly.
- Do not re-encode existing files; rollback only removes new settings fields and restores hardcoded options.
