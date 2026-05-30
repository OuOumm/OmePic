# Security Guidelines

> Backend security, trusted client IP, rate limit, IP-ban, and abuse-analysis contracts.

---

## Current State

- Client IP resolution is implemented under `backend/internal/http/clientip/` (`Resolver.Resolve`).
- IP utility functions (SHA256 hashing, IP masking) are implemented under `backend/internal/iputil/`.
- Redis-backed rate limiting is implemented under `backend/internal/ratelimit/` and `backend/internal/http/middleware/rate_limit_middleware.go`.
- CORS is wired in `backend/internal/http/router/router.go` using `gin-contrib/cors`.
- IP-ban and abuse workflows are implemented through `AdminHandler`, `AdminService`, `Repository`, and `ImageService`.
- Upload and delete requests must use the trusted client IP resolver output.

---

## Scenario: Trusted Client IP, Rate Limit, and IP Ban

### 1. Scope / Trigger

- Trigger: Upload/delete security, rate limiting, IP bans, and abuse analytics depend on one trusted client IP contract.

### 2. Signatures

- Environment keys: none for trusted proxy behavior. `config.Load()` must not read `TRUSTED_PROXY_CIDRS` or `REAL_IP_HEADER`; real-IP header selection is a SQLite runtime setting.
- Runtime setting: `real_ip_source` controls the source used by uploads, deletes, rate limits, IP bans, and abuse analytics. Valid values are `remote-addr` (default), `x-forwarded-for`, `x-real-ip`, and `cf-connecting-ip`.
- Constructor: `clientip.NewResolver(trustedProxyCIDRs []string, realIPSourceFunc func() string)`.
- Resolver: `Resolver.Resolve(req *http.Request) string`.
- Public routes using resolved IP:
  - `POST /v1/image`
  - `DELETE /i/:uid.avif`
- Admin security routes:
  - `GET /admin/ip-bans`
  - `POST /admin/ip-bans`
  - `DELETE /admin/ip-bans/:id`
  - `DELETE /admin/ip-bans/:id/images`
  - `GET /admin/abuse/overview`
  - `GET /admin/abuse/ip`
- DB tables/columns:
  - `images.ip_address`
  - `ip_bans.ip_hash`
  - `ip_bans.ip_address`
  - `ip_bans.expires_at`

### 3. Contracts

- Default `real_ip_source` is `remote-addr`, which ignores all forwarding headers.
- Admins may switch `real_ip_source` at runtime to `x-forwarded-for`, `x-real-ip`, or `cf-connecting-ip`; `Resolver.Resolve` must read the setting dynamically so changes are hot-reloaded.
- When a forwarded-header source is enabled without explicit trusted proxy CIDRs, the deployment must ensure the edge proxy strips untrusted client-supplied copies of that header before forwarding requests.
- If trusted proxy CIDRs are provided to the resolver in tests or future wiring, forwarded headers are honored only when the remote peer is in those CIDRs.
- For `X-Forwarded-For`, use the first syntactically valid IP only for the selected runtime source.
- Store `images.ip_address` from `clientip.Resolver.Resolve`, not `c.ClientIP()` and not raw headers.
- Rate-limit keys must hash the resolved client IP: `ratelimit:{scope}:ip:{sha256(client_ip)}`.
- IP-ban lookup must use `sha256(trimmed_ip)` against `ip_bans.ip_hash`.
- Active ban means `expires_at IS NULL OR expires_at = '' OR expires_at > now`.
- Public upload/delete should return HTTP 403 with error code `ip_banned` when the resolved IP is actively banned.
- Admin-created bans may be created from `uid` or explicit `ip_address`.
- Admin security UI/API display uses `ip_address` directly for image lists, ban lists, abuse ranking, and IP detail flows.
- CORS defaults to `AllowAllOrigins=true` only when runtime `public_base_url` is unset. When runtime `public_base_url` is configured, CORS must narrow to that exact origin (trim trailing slash) instead of remaining fully open.
- Startup must fail before dependency wiring when any required startup secret is missing or too short: `ADMIN_PASSWORD`, `JWT_SECRET` (32+ chars), `UID_ENCRYPTION_KEY` (32+ chars). Do not reintroduce default admin passwords or default signing/UID obfuscation secrets.

### 4. Validation & Error Matrix

- Default startup with `real_ip_source=remote-addr` and any spoofed real-IP header -> ignore header, use remote IP.
- Runtime `real_ip_source=x-forwarded-for` with a sanitized proxy header -> use first valid forwarded IP.
- Runtime `real_ip_source=x-real-ip` or `cf-connecting-ip` with a syntactically valid selected header -> use that IP.
- Resolver constructed with explicit trusted proxy and valid selected header -> use header only for trusted proxy peers.
- Resolver constructed with explicit trusted proxy and untrusted peer, missing header, or invalid selected header -> use remote IP.
- Upload/delete from active banned IP -> `ip_banned`, HTTP 403.
- `POST /admin/ip-bans` without both `uid` and `ip_address` -> `invalid_input`.
- `GET /admin/abuse/ip` with empty IP -> `invalid_input`.
- Abuse range with `from >= to` or range > 90 days -> `invalid_input`.

### 5. Good/Base/Bad Cases

- Good: default local and production startup use `RemoteAddr`, keeping rate limits and bans safe when clients spoof forwarded headers.
- Base: an admin behind a reverse proxy that strips spoofed headers can switch `real_ip_source` to `X-Forwarded-For`, `X-Real-IP`, or `CF-Connecting-IP` without restarting.
- Bad: reintroduce `TRUSTED_PROXY_CIDRS` / `REAL_IP_HEADER` environment variables or hard-code a forwarded header source outside runtime settings.

### 6. Tests Required

- Client IP resolver tests for default remote-addr behavior, runtime source hot reload, trusted proxy restriction, invalid header fallback, and `X-Forwarded-For` first-valid-IP behavior.
- Config/startup tests must assert `TRUSTED_PROXY_CIDRS` and `REAL_IP_HEADER` are not part of `AppConfig` / startup environment contract.
- Config/startup tests must assert `ADMIN_PASSWORD`, `JWT_SECRET`, and `UID_ENCRYPTION_KEY` have no defaults; startup secret enforcement rejects empty values and secrets shorter than 32 characters where required.
- Rate-limit middleware tests must assert key derivation uses resolver output and fail-closed policies return HTTP 503 with `dependency_unavailable` on Redis errors, while fail-open policies continue.
- Upload tests asserting banned IP returns `ip_banned` and does not insert an image row.
- Admin IP-ban tests for create-by-uid, create-by-ip, duplicate active ban, delete ban, and delete images by ban.
- Abuse tests for default 24-hour range, invalid ranges, top IP aggregation, top token aggregation, and active-ban annotation.

### 7. Wrong vs Correct

#### Wrong

```go
ip := c.GetHeader("X-Forwarded-For")
record.IPAddress = ip
```

#### Correct

```go
ip := resolver.Resolve(c.Request)
record.IPAddress = ip
```

---

## Logging and Redaction

- Do not log JWTs, `X-Token`, storage secrets, WebDAV passwords, S3 secrets, or full request bodies.
- Do not introduce extra IP masking fields into admin UI/API contracts unless a future product requirement explicitly restores them.
- Admin audit-style logs may include ban IDs and IP metadata needed for operations, but should not include secrets.
