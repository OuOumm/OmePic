# Implementation Plan

## Checklist

1. Config contract
   - Remove obsolete env fields from `backend/internal/config/config.go`.
   - Update `backend/internal/config/config_test.go`.
   - Adjust `backend/cmd/server/main.go` client IP resolver construction.

2. Runtime config bootstrap/status
   - Persist default runtime settings into SQLite on first run / missing keys without overwriting existing values.
   - Remove env public URL fields from `AdminEnvironmentStatus` and `loadSystemSettingsView`.
   - Update frontend TypeScript types/tests if they include these fields.

3. Admin password route
   - Add service helper if needed to verify password without issuing JWT.
   - Add `AdminHandler.ChangePassword`.
   - Add password strength validation: 8+ chars, uppercase, lowercase, symbol.
   - Map wrong old password to a clear password error message instead of bare `forbidden`.
   - Add `/admin/password` route constants/spec/router registration.
   - Add/update API client function.

4. Frontend password UI
   - Add settings-page password form for old/new password.
   - Remove explanatory/update hint text from the password form.
   - Submit through `PUT /admin/password` with admin auth header.
   - Clear password inputs after success and use existing toast/error utilities.

5. Docs/env cleanup
   - Update `.env.example` to the requested six variables.
   - Update README references to removed env keys.

6. Tests/format
   - `gofmt` changed Go files.
   - Run `cd backend && go test ./...`.
   - Run targeted frontend type/test checks if practical after UI/API type changes.

## Validation Commands

```bash
cd backend && go test ./...
```

Frontend if changed:

```bash
cd frontend && npm run typecheck
```

## Risk / Rollback

- Risk: removing trusted proxy env disables proxy IP forwarding unless a future SQLite setting is added. This is safe by default but may change production IP behavior.
- Risk: changing admin system-settings JSON shape requires frontend type/test sync.
- Risk: default runtime config persistence must not overwrite existing admin settings; test missing-key-only behavior.
- Rollback: restore removed env fields and route changes if compatibility issues appear.
