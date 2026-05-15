package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"omepic/backend/internal/repository"
)

func TestRuntimeSettingsLoadPersistsMissingDefaultsWithoutOverwritingExistingValues(t *testing.T) {
	ctx := context.Background()
	repo, err := repository.New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	if err := repo.SetConfigValue(ctx, "site_name", "Custom Site"); err != nil {
		t.Fatalf("SetConfigValue returned error: %v", err)
	}
	if err := repo.SetConfigValue(ctx, "avif_quality", "75"); err != nil {
		t.Fatalf("SetConfigValue avif_quality returned error: %v", err)
	}
	if err := repo.SetConfigValue(ctx, "avif_speed", "4"); err != nil {
		t.Fatalf("SetConfigValue avif_speed returned error: %v", err)
	}

	manager := NewRuntimeSettingsManager()
	if err := manager.Load(ctx, repo); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	values, err := repo.GetAllConfig(ctx)
	if err != nil {
		t.Fatalf("GetAllConfig returned error: %v", err)
	}
	defaultValues := runtimeConfigDefaultValues()
	preservedKeys := map[string]struct{}{
		"site_name":    {},
		"avif_quality": {},
		"avif_speed":   {},
	}
	for _, field := range GetAllFields() {
		got, ok := values[field.Key]
		if !ok {
			t.Fatalf("expected default runtime key %q to be persisted", field.Key)
		}
		if _, preserved := preservedKeys[field.Key]; !preserved && got != defaultValues[field.Key] {
			t.Fatalf("expected default runtime key %q to use field default %q, got %q", field.Key, defaultValues[field.Key], got)
		}
	}
	if values["site_name"] != "Custom Site" {
		t.Fatalf("expected existing site_name to remain unchanged, got %q", values["site_name"])
	}
	if values["avif_quality"] != "75" || values["avif_speed"] != "4" {
		t.Fatalf("expected existing avif settings to remain unchanged, got quality=%q speed=%q", values["avif_quality"], values["avif_speed"])
	}
	current := manager.Current()
	if current.SiteName != "Custom Site" {
		t.Fatalf("expected manager to load existing site name, got %q", current.SiteName)
	}
	if current.AvifQuality != 75 || current.AvifSpeed != 4 {
		t.Fatalf("expected manager to load existing avif settings, got quality=%d speed=%d", current.AvifQuality, current.AvifSpeed)
	}

	if err := repo.SetConfigValue(ctx, "site_tagline", "Custom Tagline"); err != nil {
		t.Fatalf("SetConfigValue tagline returned error: %v", err)
	}
	if err := manager.Load(ctx, repo); err != nil {
		t.Fatalf("second Load returned error: %v", err)
	}
	values, err = repo.GetAllConfig(ctx)
	if err != nil {
		t.Fatalf("GetAllConfig after reload returned error: %v", err)
	}
	if values["site_tagline"] != "Custom Tagline" {
		t.Fatalf("expected existing site_tagline to remain unchanged, got %q", values["site_tagline"])
	}
}

func TestValidateRuntimeSettingsInputRejectsInvalidAVIFSettings(t *testing.T) {
	base := RuntimeSettingsUpdateInput(defaultRuntimeSettings())
	cases := []struct {
		name   string
		mutate func(*RuntimeSettingsUpdateInput)
	}{
		{name: "quality below min", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifQuality = -1 }},
		{name: "quality above max", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifQuality = 101 }},
		{name: "speed below min", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifSpeed = -1 }},
		{name: "speed above max", mutate: func(input *RuntimeSettingsUpdateInput) { input.AvifSpeed = 11 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.mutate(&input)
			if _, err := ValidateRuntimeSettingsInput(input); err == nil || !containsError(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestGetFieldByKeyReturnsAllDefinedFields(t *testing.T) {
	for _, field := range GetAllFields() {
		retrieved, ok := GetFieldByKey(field.Key)
		if !ok {
			t.Fatalf("GetFieldByKey(%q) returned not found", field.Key)
		}
		if retrieved.Key != field.Key || retrieved.Type != field.Type {
			t.Fatalf("GetFieldByKey(%q) returned mismatched field: %+v", field.Key, retrieved)
		}
	}
	if _, ok := GetFieldByKey("nonexistent"); ok {
		t.Fatal("GetFieldByKey should return false for unknown keys")
	}
}

func TestGetAllFieldsReturnsDeterministicOrder(t *testing.T) {
	first := GetAllFields()
	second := GetAllFields()
	if len(first) != len(second) {
		t.Fatalf("GetAllFields returned different lengths: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Key != second[i].Key {
			t.Fatalf("GetAllFields order mismatch at index %d: %q vs %q", i, first[i].Key, second[i].Key)
		}
	}
}

func TestRuntimeConfigFieldsDefaultDirectoryDrivesSettingsAndPersistence(t *testing.T) {
	settings := defaultRuntimeSettings()
	defaults := runtimeConfigDefaultValues()
	serializedSettings := RuntimeSettingsToConfigValues(settings)

	if len(defaults) != len(GetAllFields()) {
		t.Fatalf("expected %d default values, got %d", len(GetAllFields()), len(defaults))
	}
	if len(serializedSettings) != len(GetAllFields()) {
		t.Fatalf("expected %d serialized settings, got %d", len(GetAllFields()), len(serializedSettings))
	}

	for _, field := range GetAllFields() {
		t.Run(field.Key, func(t *testing.T) {
			defaultValue := serializeValue(field.Default, field.Type)
			if got, ok := defaults[field.Key]; !ok || got != defaultValue {
				t.Fatalf("runtimeConfigDefaultValues mismatch: got %q present=%v want %q", got, ok, defaultValue)
			}
			if got, ok := serializedSettings[field.Key]; !ok || got != defaultValue {
				t.Fatalf("defaultRuntimeSettings serialization mismatch: got %q present=%v want %q", got, ok, defaultValue)
			}
			current, err := field.Get(&settings)
			if err != nil {
				t.Fatalf("field.Get failed: %v", err)
			}
			if got := serializeValue(current, field.Type); got != defaultValue {
				t.Fatalf("field default mismatch: got %q want %q", got, defaultValue)
			}
		})
	}
}

func TestAllFieldsRoundTripFromValues(t *testing.T) {
	values := runtimeConfigDefaultValues()

	settings, err := runtimeSettingsFromValues(values)
	if err != nil {
		t.Fatalf("runtimeSettingsFromValues failed: %v", err)
	}

	result := RuntimeSettingsToConfigValues(settings)
	for _, field := range GetAllFields() {
		original, ok := values[field.Key]
		if !ok {
			t.Fatalf("missing key %q in original values", field.Key)
		}
		got, ok := result[field.Key]
		if !ok {
			t.Fatalf("missing key %q in round-trip result", field.Key)
		}
		// String slices may differ in serialization (CSV vs JSON) and get sorted by normalizeMIMETypes;
		// verify deserialized content equality regardless of order.
		if field.Type == FieldTypeStringSlice {
			origSlice := deserializeValueOrFail(t, original, field.Type).([]string)
			gotSlice := deserializeValueOrFail(t, got, field.Type).([]string)
			if len(origSlice) != len(gotSlice) {
				t.Fatalf("field %q round-trip length mismatch: %v vs %v", field.Key, origSlice, gotSlice)
			}
			origSet := make(map[string]struct{}, len(origSlice))
			for _, v := range origSlice {
				origSet[v] = struct{}{}
			}
			for _, v := range gotSlice {
				if _, ok := origSet[v]; !ok {
					t.Fatalf("field %q round-trip: unexpected value %q in result", field.Key, v)
				}
			}
			continue
		}
		if original != got {
			t.Fatalf("field %q round-trip mismatch: %q vs %q", field.Key, original, got)
		}
	}
}

func TestRuntimeConfigFieldsOwnAccessors(t *testing.T) {
	settings := defaultRuntimeSettings()

	for _, field := range GetAllFields() {
		val, err := field.Get(&settings)
		if err != nil {
			t.Fatalf("field.Get(%q) failed: %v", field.Key, err)
		}
		if val == nil {
			t.Fatalf("field.Get(%q) returned nil", field.Key)
		}
		serialized := serializeValue(val, field.Type)
		deserialized, err := deserializeValue(serialized, field.Type)
		if err != nil {
			t.Fatalf("deserializeValue(%q, %v) failed: %v", serialized, field.Type, err)
		}
		if err := field.Set(&settings, deserialized); err != nil {
			t.Fatalf("field.Set(%q) failed: %v", field.Key, err)
		}
		if _, err := field.Get(&settings); err != nil {
			t.Fatalf("field.Get(%q) after set failed: %v", field.Key, err)
		}
	}
}

func TestRuntimeConfigFieldsRejectWrongSetterType(t *testing.T) {
	cases := []struct {
		key   string
		value interface{}
	}{
		{key: "site_name", value: 42},
		{key: "max_upload_size_mb", value: "not-a-number"},
		{key: "allow_storage_selection", value: "true"},
		{key: "allowed_mime_types", value: "image/png"},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			settings := defaultRuntimeSettings()
			field, ok := GetFieldByKey(tc.key)
			if !ok {
				t.Fatalf("missing field %q", tc.key)
			}
			if err := field.Set(&settings, tc.value); err == nil {
				t.Fatalf("expected error when setting %q with %T", tc.key, tc.value)
			}
		})
	}
}

func TestSerializeDeserializeStringSliceCSV(t *testing.T) {
	input := []string{"image/png", "image/jpeg", "image/webp"}
	serialized := serializeValue(input, FieldTypeStringSlice)
	if strings.HasPrefix(serialized, "[") {
		t.Fatalf("expected CSV, got JSON array %q", serialized)
	}
	deserialized, err := deserializeValue(serialized, FieldTypeStringSlice)
	if err != nil {
		t.Fatalf("deserializeValue failed: %v", err)
	}
	result, ok := deserialized.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", deserialized)
	}
	if len(result) != len(input) {
		t.Fatalf("length mismatch: %d vs %d", len(result), len(input))
	}
	for i := range input {
		if result[i] != input[i] {
			t.Fatalf("index %d mismatch: %q vs %q", i, result[i], input[i])
		}
	}
}

func TestDeserializeStringSliceJSONCompatibility(t *testing.T) {
	jsonArray := `["image/png","image/jpeg","image/webp"]`
	deserialized, err := deserializeValue(jsonArray, FieldTypeStringSlice)
	if err != nil {
		t.Fatalf("deserializeValue JSON failed: %v", err)
	}
	result, ok := deserialized.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", deserialized)
	}
	if len(result) != 3 || result[0] != "image/png" || result[1] != "image/jpeg" || result[2] != "image/webp" {
		t.Fatalf("unexpected JSON deserialization: %v", result)
	}
}

func TestDeserializeStringSliceBackwardCompatibleCSV(t *testing.T) {
	csv := "image/png,image/jpeg,image/webp"
	deserialized, err := deserializeValue(csv, FieldTypeStringSlice)
	if err != nil {
		t.Fatalf("deserializeValue CSV failed: %v", err)
	}
	result, ok := deserialized.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", deserialized)
	}
	if len(result) != 3 || result[0] != "image/png" || result[1] != "image/jpeg" || result[2] != "image/webp" {
		t.Fatalf("unexpected CSV deserialization: %v", result)
	}
}

func TestDeserializeStringSliceEmptyJSON(t *testing.T) {
	deserialized, err := deserializeValue("[]", FieldTypeStringSlice)
	if err != nil {
		t.Fatalf("deserializeValue empty JSON failed: %v", err)
	}
	result, ok := deserialized.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", deserialized)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
}

func TestSerializeDeserializeIntRoundTrip(t *testing.T) {
	serialized := serializeValue(42, FieldTypeInt)
	deserialized, err := deserializeValue(serialized, FieldTypeInt)
	if err != nil {
		t.Fatalf("deserializeValue failed: %v", err)
	}
	result, ok := deserialized.(int)
	if !ok {
		t.Fatalf("expected int, got %T", deserialized)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}

	// Edge cases
	serialized = serializeValue(0, FieldTypeInt)
	if serialized != "0" {
		t.Fatalf("expected \"0\", got %q", serialized)
	}
	deserialized, err = deserializeValue("0", FieldTypeInt)
	if err != nil {
		t.Fatalf("deserializeValue 0 failed: %v", err)
	}
	if deserialized.(int) != 0 {
		t.Fatal("expected 0")
	}
}

func TestSerializeDeserializeBoolRoundTrip(t *testing.T) {
	serialized := serializeValue(true, FieldTypeBool)
	if serialized != "true" {
		t.Fatalf("expected \"true\", got %q", serialized)
	}
	result, err := deserializeValue(serialized, FieldTypeBool)
	if err != nil {
		t.Fatalf("deserializeValue true failed: %v", err)
	}
	if !result.(bool) {
		t.Fatal("expected true")
	}

	serialized = serializeValue(false, FieldTypeBool)
	if serialized != "false" {
		t.Fatalf("expected \"false\", got %q", serialized)
	}
	result, err = deserializeValue(serialized, FieldTypeBool)
	if err != nil {
		t.Fatalf("deserializeValue false failed: %v", err)
	}
	if result.(bool) {
		t.Fatal("expected false")
	}
}

func TestSerializeDeserializeStringRoundTrip(t *testing.T) {
	serialized := serializeValue("hello world", FieldTypeString)
	deserialized, err := deserializeValue(serialized, FieldTypeString)
	if err != nil {
		t.Fatalf("deserializeValue failed: %v", err)
	}
	if deserialized.(string) != "hello world" {
		t.Fatal("round-trip mismatch")
	}

	serialized = serializeValue("", FieldTypeString)
	deserialized, err = deserializeValue(serialized, FieldTypeString)
	if err != nil {
		t.Fatalf("deserializeValue empty string failed: %v", err)
	}
	if deserialized.(string) != "" {
		t.Fatal("empty string round-trip mismatch")
	}
}

func TestSerializeValueNonMatchingType(t *testing.T) {
	// Passing wrong Go type for FieldTypeString should return zero value
	result := serializeValue(42, FieldTypeString)
	if result != "" {
		t.Fatalf("expected empty string for int passed as string field, got %q", result)
	}
	result = serializeValue("hello", FieldTypeInt)
	if result != "0" {
		t.Fatalf("expected \"0\" for string passed as int field, got %q", result)
	}
	result = serializeValue("hello", FieldTypeBool)
	if result != "false" {
		t.Fatalf("expected \"false\" for string passed as bool field, got %q", result)
	}
	result = serializeValue("hello", FieldTypeStringSlice)
	if result != "" {
		t.Fatalf("expected empty string for string passed as string-slice field, got %q", result)
	}
}

func TestRuntimeSettingsToConfigValuesIncludesAllFields(t *testing.T) {
	settings := defaultRuntimeSettings()
	result := RuntimeSettingsToConfigValues(settings)
	fields := GetAllFields()

	for _, field := range fields {
		if _, ok := result[field.Key]; !ok {
			t.Fatalf("RuntimeSettingsToConfigValues missing key %q", field.Key)
		}
	}
	if len(result) != len(fields) {
		t.Fatalf("expected %d keys, got %d: %v", len(fields), len(result), result)
	}
}

func deserializeValueOrFail(t *testing.T, valueStr string, fieldType FieldType) interface{} {
	t.Helper()
	result, err := deserializeValue(valueStr, fieldType)
	if err != nil {
		t.Fatalf("deserializeValue(%q) failed: %v", valueStr, err)
	}
	return result
}
