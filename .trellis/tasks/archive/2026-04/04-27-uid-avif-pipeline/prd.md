# Change UID Format and AVIF Upload Pipeline

## Goal

Replace the current public image UID format with an encrypted `prefix + sid` scheme and convert uploaded images to AVIF before writing them to storage, while preserving the existing upload, deduplication, cache, serve, and delete flows.

## What I already know

- Current upload flow is implemented in `backend/internal/service/image_service.go`.
- Current public UID is stored and served as plain text; upload URLs are built as `/i/{uid}`.
- Current storage deduplication is keyed by MD5 of the original uploaded bytes.
- Current storage object naming uses `storage.BuildObjectKey(md5Hash, extension)`.
- Current serve path in `backend/internal/http/handler/image_handler.go` directly resolves the route parameter as the DB/cache UID.
- Current SQLite schema stores `images.uid` as `TEXT UNIQUE NOT NULL`.
- The requested change is:
  - use a snowflake SID
  - compose `custom-prefix + short base62 snowflake id`
  - encrypt the raw uid with a simple XOR cipher
  - convert the XOR output into the same base64-byte-stream then base62 flow used by `test.js`
  - expose links as `/i/{uid}.avif`
  - decrypt on read and return `404` immediately when the decrypted prefix is absent or decode/decrypt fails
  - convert uploaded images to AVIF before saving to storage

## Assumptions (temporary)

- The encrypted public UID is the canonical UID stored in DB/cache; this task does not need compatibility storage for raw SID/plaintext values in local development.
- AVIF conversion is intended for raster formats in the current allowlist and may need special handling for SVG input.
- Existing admin search and delete flows should continue to operate against the new UID representation.

## Requirements (evolving)

- Replace current UUID-based public UID generation with a snowflake-based SID pipeline.
- The UID prefix must be configured through environment variable `UID_PREFIX`.
- `UID_PREFIX` value is `omeo_`.
- Generated public UID total length should stay around 24 characters and must not exceed 30 characters, including the visible prefix.
- UID generation and decryption must follow the same algorithm shape as `test.js`, implemented in Go.
- SQLite does not need to persist the original upload filename for image rows.
- Public image URLs must end in `.avif`.
- This is a breaking switch: old plaintext `/i/{uid}` links do not need compatibility.
- `/i/:uid` handling must reject malformed or prefix-missing decrypted values with `404`.
- Uploads must be converted to AVIF before persistence to the configured storage backend.
- Newly stored AVIF objects must use the generated UID as the saved filename/key basis instead of the original filename or MD5 hash.
- AVIF conversion must use a pure-Go / CGO-free backend choice for this repository.
- `svg` upload support is removed as part of this change because the chosen AVIF path targets raster images only.
- Deduplication, Redis cache preheat, serve, and delete semantics must remain correct after the UID and format change.

## Acceptance Criteria (evolving)

- [ ] Upload returns a new encrypted UID and a public URL ending in `.avif`.
- [ ] Generated public UID length stays within the target budget and never exceeds 30 characters.
- [ ] `UID_PREFIX` is loaded from environment and enforced after decrypting the raw UID.
- [ ] `images` persistence, queries, and admin listing/search no longer depend on an `original_filename` database column.
- [ ] `GET /i/{uid}.avif` decrypts the UID, validates the prefix, and serves the stored AVIF file.
- [ ] Invalid ciphertext or wrong-prefix links return `404` without falling through to normal lookup behavior.
- [ ] Old plaintext `/i/{uid}` links are no longer supported and return `404`.
- [ ] Stored files are AVIF objects rather than original-format files.
- [ ] The first physical stored object for a unique upload uses a UID-based filename/key ending in `.avif`.
- [ ] `svg` is rejected from upload validation and documentation/UI no longer advertise it as a supported type.
- [ ] Duplicate uploads still reuse the same physical stored file.
- [ ] Redis `uid:{uid}` and `md5:{hash}` behavior remains correct for upload, preheat, serve, and delete flows.
- [ ] Tests cover UID decode validation and AVIF-oriented upload/serve behavior.

## Definition of Done

- Tests added or updated for backend service/handler behavior
- Lint / typecheck / build checks green
- Docs/config updated for any new env vars or operational dependencies

## Out of Scope (explicit)

- Unrelated admin UI redesign
- Changing auth model
- Non-image media support
- Backward compatibility for old plaintext image links
- Legacy SQLite compatibility work to preserve throwaway development data; delete/reset the stale local DB instead (for the default `cd backend && go run ./cmd/server` flow, this is typically `backend/data/omepic.db`)

## Technical Approach

- Introduce a backend UID codec that:
  - generates a Twitter Snowflake SID
  - base62-encodes the SID and normalizes it to an 8-character short id using the same truncation/padding rule as `test.js`
  - builds the raw uid as `{UID_PREFIX}_{shortId}`
  - XOR-encrypts the raw uid with the configured secret key
  - converts the XOR output to base64, reinterprets that string as bytes, then base62-encodes the resulting integer
  - enforces the same max-30-character output rule as `test.js`
- Normalize public URLs to `/i/{uid}.avif`.
- On image serve and delete lookups, strip the `.avif` suffix, base62-decode, reverse the XOR/base64 transform, validate the decrypted `omeo_` prefix, and return `404` immediately on any invalid token or prefix mismatch.
- Remove `original_filename` from SQLite image persistence and repository search criteria; keep filename usage only in request-local or client-local contexts where it is still needed for upload UX.
- Generate the public UID before saving a non-duplicate upload and use that UID to build the physical object key/path for the stored AVIF file.
- Convert raster uploads to AVIF before storage, then deduplicate and cache based on the original uploaded bytes while reusing the first UID-based stored object for duplicate uploads.
- Remove SVG from validation and update frontend copy/accept lists to match the new raster-only AVIF pipeline.

## Decision (ADR-lite)

**Context**: The task changes both the public identifier contract and the stored image format. The repo currently has no image-processing dependency, and the user explicitly wants the new URL contract to be a hard break from the old plaintext links.

**Decision**:
- break compatibility with old plaintext `/i/{uid}` links
- use `UID_PREFIX=omeo_`
- use a pure-Go / CGO-free AVIF path
- remove `svg` upload support instead of introducing heavier runtime dependencies for vector-to-raster conversion

**Consequences**:
- deployment stays simpler than CGO or external-binary approaches
- existing public links stop working by design
- upload validation, docs, and frontend hints must all be updated together
- AVIF conversion and UID codec behavior become critical regression-test targets

## Technical Notes

- Inspected files:
  - `backend/internal/service/image_service.go`
  - `backend/internal/http/handler/image_handler.go`
  - `backend/internal/model/image.go`
  - `backend/internal/repository/repository.go`
  - `backend/internal/config/config.go`
- Chosen inputs so far:
  - no backward compatibility for old plaintext links
  - `UID_PREFIX=omeo_`
  - pure-Go / CGO-free AVIF conversion
  - remove `svg` upload support
- Likely impacted areas:
  - upload ID generation
  - object-key generation
  - MIME/type metadata
  - route parameter parsing and lookup
  - DB/cache key contract
  - frontend display and delete URLs
