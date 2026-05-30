# Bootstrap Discovery

## Repository state

- The repository is in a bootstrap state and currently contains only a few root files:
  - `AGENTS.md`
  - `README.md`
  - `LICENSE`
  - `start.md`
  - `.trellis/`
- There is no `backend/` or `frontend/` implementation yet.
- `git status --short --branch` failed because no `.git` directory is present in the current workspace snapshot.

## Existing convention sources

- `AGENTS.md` only contains the Trellis managed bootstrap instructions.
- There are no other convention files such as `CLAUDE.md`, `.cursorrules`, `CONTRIBUTING.md`, or `.editorconfig`.

## Product baseline from start.md

`start.md` defines the initial product architecture that future implementation should follow:

- Backend: Go, Gin, SQLite, Redis, local filesystem storage, S3-compatible storage, WebDAV storage
- Frontend: Next.js 16.2.4 App Router, TypeScript, Tailwind CSS, shadcn/ui, Zustand, react-hot-toast
- Authentication split:
  - user operations use `X-Token`
  - admin operations use `Authorization: Bearer <jwt>`
- Monorepo target shape:
  - `backend/`
  - `frontend/`
  - optional `docker-compose.yml`

## Guidance for bootstrap spec writing

- Because the codebase is still empty, the first spec pass cannot quote real source examples from app code.
- The bootstrap spec should therefore document:
  - current repository reality
  - conventions implied by `start.md`
  - explicit reminders that the spec must be updated once real code establishes stronger patterns
- Avoid pretending that unimplemented patterns already exist in code.
