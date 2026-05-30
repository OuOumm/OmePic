# Backend Development Guidelines

> Current backend conventions for the checked-in Go service.

---

## Current State

- The repository contains a checked-in `backend/` service implementation.
- These docs now describe the current contract for backend changes, with concrete
  source paths such as:
  - `backend/internal/service/image_service.go`
  - `backend/internal/http/handler/image_handler.go`
  - `backend/internal/repository/repository.go`
- Keep the behavioral scenarios in these docs synchronized with the real code
  and route contracts.

---

## Pre-Development Checklist

Before touching backend code, read:

1. [Directory Structure](./directory-structure.md)
2. [Database Guidelines](./database-guidelines.md)
3. [Security Guidelines](./security.md)
4. [Error Handling](./error-handling.md)
5. [Logging Guidelines](./logging-guidelines.md)
6. [Quality Guidelines](./quality-guidelines.md)
7. [Runtime Settings Guidelines](./runtime-settings.md)

Also confirm the documented contracts still match the repo:

- Re-read [README.md](../../../README.md) if route shape, auth headers, or storage requirements change.
- Update the relevant scenario sections when changes affect upload, deduplication, cache, serve/delete behavior, runtime settings, storage health, Cloudflare purge, or admin security APIs.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | Module organization and file layout | Active |
| [Database Guidelines](./database-guidelines.md) | SQLite + Redis persistence rules, upload storage pipeline contracts | Active |
| [Security Guidelines](./security.md) | Trusted client IP, rate limit, IP-ban, and abuse-analysis contracts | Active |
| [Error Handling](./error-handling.md) | Error types, mapping, and response rules | Active |
| [Quality Guidelines](./quality-guidelines.md) | Testing and review expectations | Active |
| [Logging Guidelines](./logging-guidelines.md) | Structured logging and redaction rules | Active |
| [Runtime Settings Guidelines](./runtime-settings.md) | Runtime site metadata, upload policy, and public/admin settings contracts | Active |

---

## Maintenance Rule

These docs intentionally describe:

- the current repository reality
- the architecture committed in [README.md](../../../README.md)
- the checked-in backend package boundaries and behavioral contracts

When implementation details or route contracts change, update these docs in the same task instead of leaving bootstrap-era text behind.
