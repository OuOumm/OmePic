# Design: Admin configurable AVIF conversion settings

## Scope

Cross-layer runtime setting addition for AVIF encoder quality/speed, covering backend runtime settings, image conversion, admin system settings API, frontend settings UI, tests, and docs/spec.

## Data Model / Persistence

Add two keys to the existing SQLite `config` key/value table through runtime settings persistence:

- `avif_quality`: integer string, default `60`, valid `0..100`.
- `avif_speed`: integer string, default `8`, valid `0..10`.

No schema migration is needed because `config` is key/value. `RuntimeSettingsManager.Load()` already persists missing defaults through insert-missing semantics; include these two keys there so existing databases get defaults without overwriting user values.

## Backend Contracts

Extend:

- `RuntimeSettings`
- `RuntimeSettingsUpdateInput`
- `RuntimeSettingsToConfigValues`
- `runtimeSettingsFromValues`
- `ValidateRuntimeSettingsInput`

Validation should fail the whole settings update before any DB write when values are out of range.

## AVIF Conversion Flow

Current `ImageService` transformer is `func([]byte) ([]byte, error)` with hardcoded `convertToAVIF` options. Change it to pass AVIF options from the current runtime settings at upload time.

Recommended shape:

- Define `AVIFConversionSettings{Quality int; Speed int}` or reuse validated values from `RuntimeSettings`.
- Add `convertToAVIFWithOptions(payload, settings)` and require callers to pass settings from `RuntimeSettings`.
- Do not add a separate AVIF conversion-layer default/hardcoded quality/speed fallback; defaults belong to SQLite/default runtime config initialization.
- Update `ImageService` transformer signature to accept options, or wrap call in `Upload` with current settings.

Deduplication should remain before conversion. Existing duplicate paths must continue not to invoke the transformer.

## Frontend

Extend shared `RuntimeSettings` type and settings page with two number inputs in the upload policy/runtime area:

- AVIF quality: `min=0`, `max=100`, integer.
- AVIF speed: `min=0`, `max=10`, integer.

Use existing `adminUpdateSystemSettings` request path. Add i18n strings for labels/hints.

## Compatibility

- Existing SQLite databases get default keys on next startup/load.
- Existing API consumers that send full runtime payloads must include the new fields after fetching `GET /admin/system-settings`; older clients that omit them may cause zero values if they submit stale payloads. The current admin UI fetches before submit and will include values.
- Public runtime settings do not need to expose AVIF encoder values unless there is a public UX need. This task only requires admin settings.

## Tests

Backend:

- Runtime defaults include `avif_quality=60` and `avif_speed=8`.
- Missing-default persistence writes new keys without overwriting existing ones.
- Validation rejects out-of-range values.
- Upload/new conversion uses configured values (test by injecting transformer and observing options, or by a focused helper test).
- Duplicate upload still skips conversion.

Frontend:

- Typecheck after adding fields.
- API test fixture updated with new fields.
- Optional helper/component test if existing test infrastructure covers settings forms.
