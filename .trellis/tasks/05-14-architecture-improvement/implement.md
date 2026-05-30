# Implementation Plan

## Checklist

1. Create `runtime_settings_fields.go`
   - Define `FieldType` enum
   - Define `ConfigField` struct
   - Define `configFieldsMap` variable
   - Define `GetFieldByKey` and `GetAllFields` functions

2. Create `runtime_settings_accessors.go`
   - Implement `getFieldValue` function
   - Implement `setFieldValue` function (returns error)

3. Create `runtime_settings_serialization.go`
   - Implement `serializeValue` function (JSON for slices)
   - Implement `deserializeValue` function (JSON + comma fallback)

4. Update `runtime_settings.go`
   - Update `RuntimeSettingsToConfigValues` to use descriptor table
   - Update `runtimeSettingsFromValues` to use descriptor table
   - Update `ValidateRuntimeSettingsInput` to use descriptor table

5. Add tests
   - Add field get/set round-trip tests
   - Add serialization/deserialization tests
   - Run `go test ./...`

6. Validation
   - `cd backend && go test ./...`
   - `gofmt` clean
   - Verify behavior unchanged

## Risk / Rollback

- Slice serialization format change may affect existing data
  - Mitigation: `deserializeValue` handles both JSON and comma-separated formats
  - Rollback: Revert to comma-separated serialization
