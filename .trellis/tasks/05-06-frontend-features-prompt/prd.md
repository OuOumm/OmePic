# Export frontend features as prompt

## Goal

Export the currently implemented frontend functionality into a root-level prompt document, without visual style or design descriptions.

## Requirements

* Inspect the current frontend implementation and summarize implemented product functionality.
* Output the prompt at the project root.
* Exclude style, visual appearance, palette, layout aesthetics, and component styling instructions.
* Keep API endpoints, state behavior, data persistence, user flows, and role-based capabilities.
* Preserve functionally relevant constraints such as request methods, headers, request bodies, upload sources, permission checks, and storage-management behavior.

## Acceptance Criteria

* [ ] Root-level prompt file exists.
* [ ] Prompt describes only frontend functionality.
* [ ] Prompt covers upload, local history, API examples, admin login/session, admin status, admin image management, storage settings, preferences, local persistence, and API client behavior.
* [ ] Prompt does not include visual or stylistic implementation guidance.

## Definition of Done

* Root prompt file updated.
* Relevant frontend code was inspected before writing the prompt.
* No application code changes are required.

## Technical Approach

Review `frontend/src` page, feature, store, IndexedDB, and API client files. Rewrite `FRONTEND_FEATURES_PROMPT.md` as a functional prompt only.

## Out of Scope

* Changing frontend code.
* Adding or modifying tests.
* Describing UI style, theme aesthetics, or visual design direction.

## Technical Notes

Inspected frontend feature files under `frontend/src/features`, shared navigation, API client, stores, and IndexedDB upload history implementation.
