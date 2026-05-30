# Design: Move mutable config to sqlite

## Scope

Backend-focused cleanup across config loading, admin settings status, admin password change route, docs, and tests.

## Environment Contract

`config.AppConfig` remains the startup-only config object. It should contain only values that cannot be read before opening SQLite or that are security bootstrap secrets:

- `HTTPAddr`
- `DatabasePath`
- `RedisURL`
- `UIDPrefix`
- `UIDEncryptionKey`
- `JWTSecret`

Remove trusted proxy env fields from `AppConfig` for this task. For now, construct the client IP resolver with no trusted proxies and default header behavior, preserving safe local behavior. `TRUSTED_PROXY_CIDRS` and `REAL_IP_HEADER` are intentionally not retained.

## SQLite Runtime Config

- Existing `config` key/value table remains the store for runtime settings.
- On first run or empty/missing config rows, default runtime settings must be persisted into SQLite with upsert semantics that only fills missing keys and does not overwrite existing user-configured values.
- `RuntimeSettings.PublicBaseURL` remains the single configured public URL source.
- `AdminEnvironmentStatus` should report `public_base_url_source` and `runtime_public_base_url_set`; remove env-specific fields.

## Admin Password

- Keep `admin_password_hash` in SQLite `config`.
- Use bcrypt for hashing and comparison.
- Expose a protected admin route:
  - `PUT /admin/password`
  - Body: `{ "old_password": string, "new_password": string }`
  - Success: JSON success with empty object.
- Handler delegates to `AdminService.ChangePassword`.
- `ChangePassword` verifies old password before writing the new hash.
- New password must be at least 8 characters and include uppercase, lowercase, and symbol characters.
- Wrong old password should map to a clear password error message in the existing JSON envelope instead of a bare `forbidden` UI message.
- Avoid using `Login` internally if that makes unnecessary JWTs; a private verification helper is preferred.

## Routing/Auth

- Add route constants/spec so frontend fallback and API 404 classification stay consistent.
- Route lives under admin group and uses existing `AdminAuth` middleware.

## Frontend

- Add a small change-password form in the admin settings area.
- Add a typed API helper for `PUT /admin/password`.
- Do not show extra explanatory/update hint text inside the change-password block.
- Do not store or display hashes; clear password fields after success.

## Compatibility

- First boot behavior must keep default `admin123` bootstrap hash if no password exists, so existing dev setup still works.
- First boot must also persist default runtime settings into SQLite without clobbering existing rows on later starts.
- Existing JWT tokens remain signed by env `JWT_SECRET`; password change does not rotate JWT secret.

## Tests

- Update config tests for removed env fields and explicit UID key behavior.
- Add service or handler tests for password change:
  - missing/empty new password rejected
  - weak new password rejected
  - wrong old password rejected with clear message
  - old password fails after change
  - new password succeeds after change
  - stored value starts with bcrypt marker and is not equal to plaintext
- Update router route tests if route list asserts all admin API paths.
- Add repository/service test for persisting missing default runtime settings without overwriting existing config values.
- Add frontend type/API tests if practical for the new password helper.
