# Plan Duplicate UID Deletion Semantics

## Goal

Clarify and, if needed, adjust duplicate-upload deletion behavior so repeated uploads of the same image create new logical UIDs while deletion does not break other UIDs that share the same physical image object.

## What I already know

- Current upload flow already creates a new logical `uid` row for duplicate image content.
- Current delete flow in `backend/internal/service/image_service.go` deletes the SQL row first.
- Current delete flow only removes the physical object when `CountByStoredFile(storage_backend, file_path) == 0`.
- That means the current implementation already protects other duplicate UIDs from losing access when one UID is deleted.
- The remaining product decision is the last-reference policy:
  - delete the physical file immediately when the final UID is deleted
  - or keep the physical file and rely on a later cleanup mechanism

## Requirements (evolving)

- Duplicate image uploads create a new logical UID record.
- Deleting one duplicate UID must not make other UIDs that point to the same physical file inaccessible.
- SQL deletion and physical-file deletion policy must be explicit and documented.
- Even when the deleted UID was the last remaining logical reference, the delete path should still only remove SQL/cache state and should not delete the physical file immediately.
- Physical files with zero remaining logical references become orphaned assets and must be handled by a later cleanup mechanism rather than the online delete path.

## Acceptance Criteria

- [ ] Duplicate uploads create a new UID row that can coexist with other rows pointing at the same stored file.
- [ ] Deleting one UID removes only that SQL/cache record and never breaks access for sibling UIDs.
- [ ] Deleting the last remaining UID still does not delete the physical file in the request path.
- [ ] The system clearly distinguishes logical deletion from physical garbage collection in docs/specs and service behavior.

## Technical Approach

- Keep the current duplicate-upload behavior of inserting a new logical UID row for the same stored object.
- Change delete semantics so the online delete path:
  - validates ownership
  - deletes the SQL row
  - clears `uid:{uid}` cache
  - repairs or clears `md5:{hash}` cache as needed
  - never deletes the physical object from storage
- Treat unreferenced files as deferred-cleanup candidates for a future maintenance job.
- Update service tests, README/spec wording, and any admin assumptions so the retained-physical-file policy is explicit.

## Technical Notes

- Inspected:
  - `backend/internal/service/image_service.go`
  - `backend/internal/service/image_service_test.go`
  - `.trellis/spec/backend/database-guidelines.md`
  - `README.md`
