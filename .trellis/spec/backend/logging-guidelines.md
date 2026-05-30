# Logging Guidelines

> Logging rules for the current backend implementation.

---

## Current State

- The backend is already using structured logging via Go's standard `log/slog`.
- Keep new code on the same structured logging path unless a stronger project-wide requirement is introduced.

---

## Log Levels

- `debug`: optional local-only detail such as Redis preheat counts by batch or SQL timing diagnostics. Do not depend on debug logs for normal operations.
- `info`: startup milestones, selected storage backend, HTTP listen address, successful Redis preheat summary, and high-level admin actions.
- `warn`: recoverable abnormal behavior such as invalid tokens, rejected uploads, Redis misses that fall back to SQLite, or failed attempts to delete an image not owned by the caller.
- `error`: request failures caused by dependency outages, failed storage operations, migration failures, or invariant breaks.

Use log levels consistently so production filtering stays useful.

---

## Structured Logging

Prefer key/value structured logs with fields such as:

- `request_id`
- `method`
- `path`
- `status`
- `uid`
- `storage_backend`
- `cache_hit`
- `duration_ms`

Bootstrap example:

```go
logger.Info(
    "image served",
    "uid", uid,
    "storage_backend", record.StorageBackend,
    "cache_hit", cacheHit,
)
```

Keep request logging in middleware such as `backend/internal/http/middleware/logging_middleware.go` once that file exists.

---

## What To Log

- Process startup and shutdown.
- Migration success or failure.
- Redis preheat totals and any skipped or failed records.
- Upload success with metadata safe to expose internally:
  - `uid`
  - `size`
  - `mime_type`
  - `storage_backend`
  - `duplicate`
- Admin configuration changes, but only after masking secrets.
- Storage backend failures with enough context to triage which adapter failed.

---

## What NOT To Log

Never log:

- raw `X-Token` values
- raw JWTs
- admin passwords
- `JWT_SECRET`
- Redis URLs containing credentials
- S3 secret keys
- WebDAV passwords
- full unredacted config payloads
- image binary contents

If a log line needs to mention a secret-bearing field, log that it was present or changed, not the value itself.
