# Quality Guidelines

> Backend quality bar for the current implementation and future changes.

---

## Current State

- The backend source tree and test suite exist.
- These rules define the minimum acceptable local quality bar and review expectations for incremental changes.

---

## Forbidden Patterns

- Business logic in Gin handlers.
- Direct Redis or SQLite access from handlers or middleware.
- Generic dumping-ground files such as `utils.go` or `misc.go`.
- Global mutable singletons for request-scoped state.
- Swallowing errors and returning success anyway.
- Diverging response shapes between `/v1/*` and `/admin/*` without a documented reason.
- Storage adapters that know about Gin request objects.

---

## Required Patterns

- Keep HTTP, service, repository, cache, and storage concerns separated as described in [directory-structure.md](./directory-structure.md).
- Pass `context.Context` through repository, cache, and storage boundaries.
- Centralize auth header parsing for:
  - `X-Token`
  - `Authorization: Bearer <jwt>`
- Prefer small interfaces at dependency boundaries, especially for storage adapters, repositories, and cache seams.
- Cache call sites must depend on intent-scoped seams instead of one broad adapter view: UID image lookup/write, batch preheat writes, MD5 mapping operations, and health checks are separate interfaces even when one Redis adapter implements all of them.
- Keep cache-key formats in one package so `uid:{uid}` and scoped `md5:{storage_key}:{hash}` do not drift; MD5 mapping behavior belongs behind the `md5MappingFlow` domain module rather than direct service/handler key composition.

---

## Testing Requirements

The backend test suite already exists under `backend/internal/` with ~14 test files covering:

| File | Coverage |
|------|----------|
| `image_service_test.go` | Upload success, deduplication by MD5, AVIF conversion, delete token checks, IP ban enforcement, storage selection, and duplicate skip behavior |
| `image_handler_test.go` | HTTP handler upload / delete with various input validations |
| `admin_handler_test.go` | Admin login, status, image management, settings update, password change |
| `admin_service_test.go` | Admin service: image listing, config updates, runtime settings, IP bans, abuse statistics |
| `runtime_settings_test.go` | Default key persistence, missing-key insert, AVIF/image-limit range validation, MIME normalization |
| `announcement_service_test.go` | Create, list, update, delete announcements |
| `repository_test.go` | SQLite insert/query/delete for images, config, storage_configs, ip_bans, storage_health_checks |
| `frontend_test.go` / `router` | Frontend fallback serving, API path preservation, static asset serving |
| `cache/redis_cache_test.go` | Redis client configuration |
| `iputil/iputil_test.go` | IP hash and mask helpers; client IP resolver tests belong under `http/clientip` |
| `auth/jwt_test.go` | JWT generation and validation |
| `uid/codec_test.go` | UID encode/decode round-trip with prefix verification and opaque head property |
| `storage/storage_test.go` | Storage adapter factory, route resolution, and stream save behavior |
| `config/config_test.go` | Environment config loading |

Minimum local checks:

- `go test ./...` from `backend/`
- `gofmt`-clean source

If linting such as `staticcheck` is added later, document the exact command here.

---

## Code Review Checklist

- Does the change preserve the auth split from [README.md](../../../README.md)?
- Are SQLite and Redis responsibilities clearly separated?
- Does the delete path stay limited to logical deletion and cache repair, leaving physical-file cleanup to a deferred maintenance flow?
- Are secrets masked in logs and API responses?
- Are new packages placed under intentional paths rather than catch-all folders?
- Did the author add or update tests for deduplication, cache, auth, storage health, runtime limits, or delete semantics when touching those flows?
