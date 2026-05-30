# Error Handling

> Error handling contract for the current backend implementation.

---

## Current State

- Backend domain errors are defined under `backend/internal/service/`.
- JSON response helpers live under `backend/internal/response/`.
- The rules below are the active contract for keeping error shapes and route behavior consistent across endpoints.

---

## Error Types

Define typed or sentinel errors in the service layer and map them once in HTTP handlers or middleware.

Recommended bootstrap categories:

| Code | When to use | HTTP status |
|------|-------------|-------------|
| `invalid_input` | Missing file, unsupported type, oversize upload, malformed JSON | `400` |
| `missing_token` | `X-Token` header missing on user routes | `401` |
| `invalid_admin_token` | Missing or invalid `Authorization: Bearer` JWT | `401` |
| `forbidden` | Token mismatch on delete, admin action without enough scope | `403` |
| `not_found` | Unknown `uid`, missing config entry, missing image record | `404` |
| `conflict` | Duplicate writes that cannot be resolved safely | `409` |
| `dependency_unavailable` | SQLite, Redis, storage backend, or WebDAV/S3 dependency failure | `503` |
| `internal_error` | Unexpected bug or invariant break | `500` |

Keep storage-driver-specific and database-driver-specific errors wrapped, then translate them into one of the categories above before they reach the client.

---

## Error Handling Patterns

- Services return domain errors; handlers map domain errors to HTTP.
- Middleware owns cross-cutting failures such as malformed auth headers.
- Do not `panic` for expected user-caused errors.
- Log the detailed underlying error once, near the boundary that has full context.
- Return sanitized client messages. The response body should never expose:
  - raw SQL errors
  - filesystem paths outside the intended public contract
  - JWT secrets or Redis connection details

Actual handler pattern:

```go
if errors.Is(err, service.ErrTokenMismatch) {
    response.Error(c, http.StatusForbidden, "forbidden", "token does not own this image")
    return
}
```

---

## API Error Responses

For JSON APIs, use one envelope shape everywhere:

```json
{
  "success": false,
  "error": {
    "code": "invalid_input",
    "message": "file type is not allowed"
  }
}
```

Notes:

- `POST /v1/image`, `POST /admin/login`, `/admin/*`, and `/health` should use JSON.
- `GET /i/:uid.avif` serves bytes, so it may return plain HTTP statuses without the JSON envelope when the file cannot be served.
- `DELETE /i/:uid.avif` should still use JSON because the route is an API-like mutation even though it shares the `/i/:uid` path family.

---

## Scenario: XOR-Obfuscated Public UID Route Validation

### 1. Scope / Trigger

- Trigger: public image routes now carry XOR-obfuscated public UID tokens plus a required `.avif` suffix.

### 2. Signatures

- `GET /i/{uid}.avif`
- `DELETE /i/{uid}.avif`

### 3. Contracts

- Public routes must strip `.avif`, decode/deobfuscate the remaining token, and validate the configured `UID_PREFIX` before SQLite or Redis lookup.
- Admin-internal delete flows may still operate on canonical bare XOR-obfuscated public UIDs because they are not public route inputs.

### 4. Validation & Error Matrix

- Missing `.avif` suffix -> `404`
- Base62 decode or base64 unpack failure -> `404`
- XOR deobfuscation / plaintext validation failure -> `404`
- Decrypted prefix mismatch -> `404`
- Valid XOR-obfuscated public UID with no backing row -> `404`

### 5. Good / Base / Bad Cases

- Good: `/i/<valid-token>.avif` -> resolve and serve/delete normally
- Base: valid token with missing DB row -> `404`
- Bad: log or expose whether the failure was malformed base encoding, bad obfuscated payload, or wrong prefix

### 6. Tests Required

- Unit test for malformed token -> `ErrNotFound`
- Unit test for prefix mismatch -> `ErrNotFound`
- Service or handler test for bare `/i/{uid}` route -> `ErrNotFound` / `404`

### 7. Wrong vs Correct

#### Wrong

- Attempt Redis/SQLite lookup first and only later discover the route token was invalid

#### Correct

- Reject malformed, undecodable, or wrong-prefix route tokens immediately as `404` before normal lookup behavior

---

## Common Mistakes To Avoid

- Returning `500` for validation failures or auth failures.
- Logging the submitted admin password, JWT, `X-Token`, or storage credentials.
- Deleting the physical file in the online request path instead of limiting the request to logical deletion and cache repair.
- Mixing multiple error body shapes across admin and user APIs.
- Treating a Redis miss as a fatal error when SQLite can still answer the request.
