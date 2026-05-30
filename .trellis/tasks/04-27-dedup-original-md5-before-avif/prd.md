# Deduplicate On Original Upload Bytes Before AVIF Conversion

## Goal

Change the AVIF upload pipeline so deduplication keys off the original uploaded file bytes before any AVIF conversion work happens, avoiding unnecessary AVIF conversion for duplicate uploads while preserving the current public UID and stored-AVIF contract.

## What I already know

- Current upload flow in `backend/internal/service/image_service.go` converts to AVIF first and then computes `md5_hash` from the converted bytes.
- That means duplicate uploads still pay the AVIF conversion cost before deduplication can short-circuit.
- Current public contract already hard-switched to encrypted public UID tokens and `/i/{uid}.avif`.
- Current storage contract still requires stored objects and served MIME type to be AVIF.

## Requirements

- Compute the deduplication MD5 from the original uploaded bytes before AVIF conversion.
- Check Redis/SQLite duplicate mappings using that original-bytes MD5.
- If a duplicate exists, skip AVIF conversion and physical storage writes, and just insert the new logical UID row that reuses the existing stored AVIF object.
- If no duplicate exists, then convert to AVIF and store the AVIF object as before.
- Keep stored objects, `mime_type`, public URLs, and public route semantics unchanged from the current AVIF contract.
- Update tests and specs to reflect that `md5_hash` now represents the original uploaded bytes, not the transformed AVIF bytes.

## Acceptance Criteria

- [ ] Duplicate uploads are detected from the original uploaded bytes before AVIF conversion.
- [ ] Duplicate uploads do not invoke AVIF conversion again.
- [ ] New uploads still store AVIF objects and serve `image/avif`.
- [ ] Redis `md5:{hash}` and SQLite `md5_hash` continue to work for upload, preheat, serve, and delete flows.
- [ ] Backend tests cover the new dedup semantics and the “no conversion on duplicate” behavior.

## Definition of Done

- Backend tests updated and passing
- Backend build passing
- Relevant spec text updated for the changed meaning of `md5_hash`

## Out of Scope

- Changing the encrypted UID contract
- Restoring old plaintext link compatibility
- Changing frontend behavior unrelated to upload result display

## Technical Approach

- Hash `input.Bytes` first and use that hash for duplicate lookup.
- Only call the AVIF transformer on the non-duplicate path.
- Keep storing AVIF bytes and `image/avif` metadata on the non-duplicate path.
- Use tests to verify duplicate uploads share the same stored file while conversion runs only once.

## Technical Notes

- Inspected:
  - `backend/internal/service/image_service.go`
  - `backend/internal/service/image_service_test.go`
  - `.trellis/spec/backend/database-guidelines.md`
- Likely impacted:
  - upload flow
  - dedup tests
  - database/cache spec wording
