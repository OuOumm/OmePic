# Frontend Clipboard And URL Upload

## Goal

Add two new frontend upload sources to the existing homepage upload workflow: paste an image from the clipboard, and enter an image URL so the frontend downloads the image data and uploads it through the current image upload API.

## What I Already Know

- The homepage route renders `frontend/src/features/upload/UploadPageClient.tsx`.
- File selection and drag/drop are handled by `frontend/src/features/upload/UploadDropzone.tsx`.
- Uploads already go through `uploadImageWithProgress(file, token, setProgress, selectedStorageKey)`.
- The existing API accepts a `File` through multipart form data and supports optional `storage_key`.
- Current UI text says URL upload is unavailable, so this task should replace that state with a working URL upload affordance.
- Upload success already updates toast, recent upload state, IndexedDB history, selected storage behavior, and duplicate messaging.

## Requirements

- Support clipboard image upload on the homepage.
  - Users can paste image data copied from screenshots, image editors, or browser image copy operations while focused on the upload page/dropzone.
  - The first supported image item in the clipboard should be converted to a `File` and sent through the existing upload flow.
  - If the clipboard has no image, show a localized error toast instead of failing silently.
- Support image URL upload on the homepage.
  - Users can enter an HTTP or HTTPS URL.
  - The frontend downloads the URL with `fetch`, validates that the response is an image, converts it to a `File`, and sends it through the existing upload flow.
  - URL upload must reuse the existing client token, selected storage target, progress state, success handling, duplicate handling, and local history save behavior.
  - Handle invalid URLs, non-image responses, and download failures with localized user-facing errors.
- Preserve existing local file, drag/drop, storage selection, recent uploads, and history behavior.
- Keep the backend upload contract unchanged for this task.
  - The unchanged contract is still multipart `POST /v1/image` with field `file`, optional `storage_key`, current token/progress/history behavior, and AVIF output URLs.
  - Source validation may include AVIF as an accepted raster input because the existing backend conversion pipeline can decode and re-encode AVIF; successful uploads still return `mime_type: image/avif` and `/i/{uid}.avif`.
- Keep bilingual UI strings in `frontend/src/lib/i18n.ts`.

## Acceptance Criteria

- [ ] Selecting a local file still uploads successfully through the existing workflow.
- [ ] Drag/drop still uploads successfully and drag hover behavior remains stable.
- [ ] Pasting a clipboard image while on the upload page triggers upload.
- [ ] Pasting non-image clipboard content shows a localized error.
- [ ] Entering a valid image URL downloads the image and uploads it through the existing API.
  - CDN image URLs with transform-style names such as `*.jpg@...avif` are accepted when the response is a supported image MIME type, or when the server returns a generic binary content type and the URL's final path segment ends in a supported image extension.
- [ ] Entering an invalid URL, a non-image URL, or an unreachable URL shows a localized error.
- [ ] URL upload controls are disabled while token preparation or upload is in progress.
- [ ] Recent uploads and IndexedDB history include uploads from file, paste, and URL sources with a useful filename.
- [ ] Lint/typecheck/build pass for the frontend.

## Definition Of Done

- Tests added or updated where practical for new helper logic.
- Lint, typecheck, and build are run for the frontend.
- No backend API contract changes are introduced.
- New UI states are accessible, localized, and consistent with existing shadcn-style frontend components.

## Technical Approach

- Convert every upload source to a browser `File` and call the existing `handleUpload(file)` path.
- Add paste handling in the upload page/dropzone area using browser clipboard `DataTransferItem` image items.
- Add a compact URL input and action button near the upload quick actions, replacing the existing "URL upload unavailable" messaging.
- Add small helper logic for filename derivation and URL image download, using structured URL parsing and `Blob` validation rather than ad hoc string handling.

## Decision (ADR-lite)

**Context**: URL uploads could be implemented either by adding a backend fetch-by-URL endpoint or by letting the browser download the image and submit it as a file.

**Decision**: For this task, use frontend download-to-File and keep the backend upload API unchanged.

**Consequences**: This keeps the change small and preserves existing upload/storage contracts. Browser CORS can block some remote image URLs; those failures should be surfaced as download/upload-source errors. A future backend URL-import endpoint can be considered if server-side fetch, private network rules, or CORS avoidance becomes necessary.

## Out Of Scope

- Adding a backend URL-import endpoint.
- Uploading multiple pasted files or multiple URL files at once.
- Supporting non-image URL content.
- Changing AVIF conversion, deduplication, UID, or storage behavior.

## Technical Notes

- Relevant files inspected:
  - `frontend/src/features/upload/UploadPageClient.tsx`
  - `frontend/src/features/upload/UploadDropzone.tsx`
  - `frontend/src/lib/api.ts`
  - `frontend/src/lib/i18n.ts`
  - `frontend/src/stores/upload-store.ts`
- Memory notes indicate homepage upload hierarchy should keep the upload workbench primary and preserve existing upload/history/admin/storage contracts.
