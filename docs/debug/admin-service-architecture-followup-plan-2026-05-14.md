# AdminService Architecture Follow-up Plan

Date: 2026-05-14
Status: proposal

## Why this follow-up exists

After stabilizing the upload pipeline and runtime-security contracts, `backend/internal/service/admin_service.go` remains the largest backend service and still carries several unrelated responsibilities:

- admin login / password rotation
- image listing / deletion
- storage instance CRUD / default switching
- runtime system settings
- IP ban workflows
- abuse statistics

This is workable today, but it increases review cost, test setup size, and the chance of cross-feature regressions.

## Recommended split

### 1. `AdminAuthService`
Owns:
- `Login`
- `ChangePassword`
- default-password bootstrap helpers
- security status helpers related to password / JWT / UID defaults

### 2. `AdminImageService`
Owns:
- `Status`
- `Images`
- `DeleteImages`
- image-centric admin views

### 3. `AdminSecurityService`
Owns:
- `CreateIPBan`
- `IPBans`
- `DeleteIPBan`
- `DeleteImagesByIPBan`
- `AbuseOverview`
- `AbuseIPDetail`

### 4. `AdminConfigService`
Owns:
- `GetConfig`
- `UpdateConfig`
- storage instance create/update/delete/default switch
- `GetSystemSettings`
- `UpdateSystemSettings`
- readonly runtime/security/storage environment views

## Migration strategy

### Phase 1: extract internal helpers only
- Move storage-config helpers into a dedicated file / struct.
- Move password helpers into a dedicated file / struct.
- Keep public handler wiring unchanged.

### Phase 2: introduce focused services behind current handler contract
- Construct the four focused services in `main.go`.
- Keep `AdminHandler` route surface unchanged.
- Route methods to the new services one group at a time.

### Phase 3: split tests
- Extract `newAdminServiceTestHarness` into reusable focused harness helpers.
- Separate auth/config/security/image admin tests to reduce setup noise.

## Guardrails

- Do not change JSON response shapes.
- Do not change admin routes.
- Keep runtime readonly security fields (`configured` / `using_default`) stable.
- Keep storage backend change protections stable.
- Preserve current IP ban and abuse semantics.

## Suggested task order

1. Extract `AdminAuthService`
2. Extract `AdminConfigService`
3. Extract `AdminSecurityService`
4. Leave `Status` / `Images` / `DeleteImages` in a slim `AdminImageService`

## Success criteria

- `AdminService` either disappears or becomes a thin compatibility facade.
- Public/admin route contracts remain unchanged.
- Existing backend tests still pass with smaller, more focused harnesses.
- No feature behavior regressions in storage/runtime/security flows.
