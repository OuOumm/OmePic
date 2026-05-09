# Pure-Go AVIF Options

## Question

Which AVIF approach best fits this repo after the user chose:

- pure-Go style deployment
- no old-link compatibility
- raster-only AVIF conversion

## Sources checked

- `github.com/gen2brain/avif` README / package docs
- `github.com/vegidio/avif-go` package docs
- older `go-avif` package docs

## Findings

### `github.com/gen2brain/avif`

- Described as CGO-free.
- Uses libavif/aom compiled to WASM through wazero, with optional dynamic-library acceleration.
- Supports encode/decode without requiring CGO at build time.

### `github.com/vegidio/avif-go`

- Described as having no external system dependency installs.
- Still requires CGO-enabled builds.
- Conflicts with the chosen lower-friction deployment direction for this repo.

### Older `go-avif` variants

- Depend on libaom or similar native dependencies.
- Not a fit for the chosen pure-Go / no-extra-runtime approach.

## Recommendation

Use `github.com/gen2brain/avif` for this task.

Why:

- Best fit for the chosen deployment constraint.
- Avoids CGO and external binary setup in the first implementation pass.
- Good enough for a backend upload-convert-store pipeline where simplicity matters more than maximal throughput.

## Implementation notes

- Deduplication should hash the transformed AVIF bytes, not the original upload bytes, because the stored physical object is AVIF.
- `svg` should be removed from the accepted upload list rather than forcing a vector-raster conversion path into this task.
- The new MIME contract should become `image/avif` for stored and served objects.
