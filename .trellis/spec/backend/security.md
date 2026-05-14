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

- Environment keys: none for trusted proxy behavior. `config.Load()` must not read `TRUSTED_PROXY_CIDRS` or `REAL_IP_HEADER`.
- Constructor: `clientip.NewResolver(trustedProxyCIDRs []string, realIPHeader string)`; production startup currently passes `nil, ""`, so no forwarded header is trusted by default.
- Resolver: `Resolver.Resolve(req *http.Request) string`.
- Public routes using resolved IP:
  - `POST /v1/image`
  - `DELETE /i/:uid`
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
  - `ip_bans.ip_address_masked`
  - `ip_bans.expires_at`

### 3. Contracts

- Never trust `X-Forwarded-For` or `X-Real-IP` directly from untrusted remote peers.
- The default application wiring has no trusted proxies and therefore always ignores forwarded headers.
- The resolver may still be unit-tested with explicit trusted CIDRs, but those values must come from an intentional future configuration source, not environment variables removed from `AppConfig`.
- For `X-Forwarded-For`, use the first syntactically valid IP only when the remote peer is trusted by the resolver instance.
- Store `images.ip_address` from `clientip.Resolver.Resolve`, not `c.ClientIP()` and not raw headers.
- Rate-limit keys must hash the resolved client IP: `ratelimit:{scope}:ip:{sha256(client_ip)}`.
- IP-ban lookup must use `sha256(trimmed_ip)` against `ip_bans.ip_hash`.
- Active ban means `expires_at IS NULL OR expires_at = '' OR expires_at > now`.
- Public upload/delete should return HTTP 403 with error code `ip_banned` when the resolved IP is actively banned.
- Admin-created bans may be created from `uid` or explicit `ip_address`.
- UI/API display should use `ip_address_masked` where full IP display is unnecessary.
- CORS defaults to `AllowAllOrigins=true` only when runtime `public_base_url` is unset. When runtime `public_base_url` is configured, CORS must narrow to that exact origin (trim trailing slash) instead of remaining fully open.
- Startup should warn when `JWT_SECRET` or `UID_ENCRYPTION_KEY` still use their documented default values.

### 4. Validation & Error Matrix

- Default startup with any spoofed real-IP header -> ignore header, use remote IP.
- Resolver constructed with explicit trusted proxy and valid `X-Forwarded-For` -> use first valid forwarded IP.
- Resolver constructed with explicit trusted proxy and missing/invalid real-IP header -> use remote IP.
- Upload/delete from active banned IP -> `ip_banned`, HTTP 403.
- `POST /admin/ip-bans` without both `uid` and `ip_address` -> `invalid_input`.
- `GET /admin/abuse/ip` with empty IP -> `invalid_input`.
- Abuse range with `from >= to` or range > 90 days -> `invalid_input`.

### 5. Good/Base/Bad Cases

- Good: default local and production startup use `RemoteAddr`, keeping rate limits and bans safe even when clients spoof forwarded headers.
- Base: a future SQLite-backed trusted-proxy setting may call `clientip.NewResolver` with explicit trusted CIDRs, then forwarded headers are read only from those peers.
- Bad: reintroduce `TRUSTED_PROXY_CIDRS` / `REAL_IP_HEADER` environment variables or trust `X-Forwarded-For` from every request; attackers can bypass rate limits or bans by spoofing the header.

### 6. Tests Required

- Client IP resolver tests for trusted proxy, untrusted proxy, invalid header, and X-Forwarded-For first-valid-IP behavior.
- Config/startup tests must assert `TRUSTED_PROXY_CIDRS` and `REAL_IP_HEADER` are not part of `AppConfig` / startup environment contract.
- Rate-limit middleware tests asserting key derivation uses resolver output.
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
- Prefer masked IPs in UI-oriented responses where full IP is not required.
- Admin audit-style logs may include ban IDs and masked IPs, but should not include secrets.
