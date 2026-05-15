package service

import "fmt"

// FieldType represents the data type of a runtime config field.
type FieldType int

const (
	FieldTypeString FieldType = iota
	FieldTypeInt
	FieldTypeBool
	FieldTypeStringSlice
)

type runtimeFieldGetter func(*RuntimeSettings) (interface{}, error)
type runtimeFieldSetter func(*RuntimeSettings, interface{}) error

// ConfigField describes a single runtime config field definition.
//
// Keep the persistent key, default value and RuntimeSettings accessors together
// so adding a runtime setting automatically participates in default generation,
// missing-default persistence, serialization, and deserialization.
type ConfigField struct {
	Key     string
	Type    FieldType
	Default interface{}
	Get     runtimeFieldGetter
	Set     runtimeFieldSetter
}

var runtimeConfigFields = []ConfigField{
	{
		Key:     "site_name",
		Type:    FieldTypeString,
		Default: DefaultSiteName,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.SiteName, nil
		},
		Set: setStringRuntimeField("site_name", func(settings *RuntimeSettings, value string) {
			settings.SiteName = value
		}),
	},
	{
		Key:     "site_tagline",
		Type:    FieldTypeString,
		Default: DefaultSiteTagline,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.SiteTagline, nil
		},
		Set: setStringRuntimeField("site_tagline", func(settings *RuntimeSettings, value string) {
			settings.SiteTagline = value
		}),
	},
	{
		Key:     "public_base_url",
		Type:    FieldTypeString,
		Default: "",
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.PublicBaseURL, nil
		},
		Set: setStringRuntimeField("public_base_url", func(settings *RuntimeSettings, value string) {
			settings.PublicBaseURL = value
		}),
	},
	{
		Key:     "max_upload_size_mb",
		Type:    FieldTypeInt,
		Default: 20,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.MaxUploadSizeMB, nil
		},
		Set: setIntRuntimeField("max_upload_size_mb", func(settings *RuntimeSettings, value int) {
			settings.MaxUploadSizeMB = value
		}),
	},
	{
		Key:     "allowed_mime_types",
		Type:    FieldTypeStringSlice,
		Default: defaultAllowedMIMETypes,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.AllowedMIMETypes, nil
		},
		Set: setStringSliceRuntimeField("allowed_mime_types", func(settings *RuntimeSettings, value []string) {
			settings.AllowedMIMETypes = value
		}),
	},
	{
		Key:     "avif_quality",
		Type:    FieldTypeInt,
		Default: DefaultAVIFQuality,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.AvifQuality, nil
		},
		Set: setIntRuntimeField("avif_quality", func(settings *RuntimeSettings, value int) {
			settings.AvifQuality = value
		}),
	},
	{
		Key:     "avif_speed",
		Type:    FieldTypeInt,
		Default: DefaultAVIFSpeed,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.AvifSpeed, nil
		},
		Set: setIntRuntimeField("avif_speed", func(settings *RuntimeSettings, value int) {
			settings.AvifSpeed = value
		}),
	},
	{
		Key:     "allow_storage_selection",
		Type:    FieldTypeBool,
		Default: true,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.AllowStorageSelect, nil
		},
		Set: setBoolRuntimeField("allow_storage_selection", func(settings *RuntimeSettings, value bool) {
			settings.AllowStorageSelect = value
		}),
	},
	{
		Key:     "maintenance_mode",
		Type:    FieldTypeBool,
		Default: false,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.MaintenanceMode, nil
		},
		Set: setBoolRuntimeField("maintenance_mode", func(settings *RuntimeSettings, value bool) {
			settings.MaintenanceMode = value
		}),
	},
	{
		Key:     "maintenance_message",
		Type:    FieldTypeString,
		Default: "",
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.MaintenanceMessage, nil
		},
		Set: setStringRuntimeField("maintenance_message", func(settings *RuntimeSettings, value string) {
			settings.MaintenanceMessage = value
		}),
	},
	{
		Key:     "rate_limit_window_minutes",
		Type:    FieldTypeInt,
		Default: DefaultRateLimitWindowMinutes,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.RateLimitWindowMinutes, nil
		},
		Set: setIntRuntimeField("rate_limit_window_minutes", func(settings *RuntimeSettings, value int) {
			settings.RateLimitWindowMinutes = value
		}),
	},
	{
		Key:     "rate_limit_max_requests",
		Type:    FieldTypeInt,
		Default: DefaultRateLimitMaxRequests,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.RateLimitMaxRequests, nil
		},
		Set: setIntRuntimeField("rate_limit_max_requests", func(settings *RuntimeSettings, value int) {
			settings.RateLimitMaxRequests = value
		}),
	},
	{
		Key:     "upload_rate_limit_window_minutes",
		Type:    FieldTypeInt,
		Default: DefaultUploadRateLimitWindowMinutes,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.UploadRateLimitWindowMinutes, nil
		},
		Set: setIntRuntimeField("upload_rate_limit_window_minutes", func(settings *RuntimeSettings, value int) {
			settings.UploadRateLimitWindowMinutes = value
		}),
	},
	{
		Key:     "upload_rate_limit_max_requests",
		Type:    FieldTypeInt,
		Default: DefaultUploadRateLimitMaxRequests,
		Get: func(settings *RuntimeSettings) (interface{}, error) {
			return settings.UploadRateLimitMaxRequests, nil
		},
		Set: setIntRuntimeField("upload_rate_limit_max_requests", func(settings *RuntimeSettings, value int) {
			settings.UploadRateLimitMaxRequests = value
		}),
	},
}

// configFieldsMap is the descriptor table for all runtime config fields.
// It is generated from runtimeConfigFields so lookup and iteration share the
// same source of truth.
var configFieldsMap = buildConfigFieldsMap(runtimeConfigFields)

func buildConfigFieldsMap(fields []ConfigField) map[string]ConfigField {
	result := make(map[string]ConfigField, len(fields))
	for _, field := range fields {
		result[field.Key] = field
	}
	return result
}

// GetFieldByKey returns the ConfigField definition for the given key.
func GetFieldByKey(key string) (ConfigField, bool) {
	field, ok := configFieldsMap[key]
	return field, ok
}

// GetAllFields returns all config field definitions in a deterministic order.
func GetAllFields() []ConfigField {
	fields := make([]ConfigField, len(runtimeConfigFields))
	copy(fields, runtimeConfigFields)
	return fields
}

func runtimeSettingsFromFieldDefaults() RuntimeSettings {
	var settings RuntimeSettings
	for _, field := range runtimeConfigFields {
		if err := field.Set(&settings, cloneConfigFieldValue(field.Default)); err != nil {
			panic(fmt.Sprintf("invalid runtime config default for %s: %v", field.Key, err))
		}
	}
	return normalizeRuntimeSettings(settings)
}

func runtimeConfigDefaultValues() map[string]string {
	result := make(map[string]string, len(runtimeConfigFields))
	for _, field := range runtimeConfigFields {
		result[field.Key] = serializeValue(cloneConfigFieldValue(field.Default), field.Type)
	}
	return result
}

func cloneConfigFieldValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []string:
		return append([]string(nil), v...)
	default:
		return v
	}
}

func setStringRuntimeField(key string, assign func(*RuntimeSettings, string)) runtimeFieldSetter {
	return func(settings *RuntimeSettings, value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string, got %T", key, value)
		}
		assign(settings, v)
		return nil
	}
}

func setIntRuntimeField(key string, assign func(*RuntimeSettings, int)) runtimeFieldSetter {
	return func(settings *RuntimeSettings, value interface{}) error {
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("%s must be an int, got %T", key, value)
		}
		assign(settings, v)
		return nil
	}
}

func setBoolRuntimeField(key string, assign func(*RuntimeSettings, bool)) runtimeFieldSetter {
	return func(settings *RuntimeSettings, value interface{}) error {
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%s must be a bool, got %T", key, value)
		}
		assign(settings, v)
		return nil
	}
}

func setStringSliceRuntimeField(key string, assign func(*RuntimeSettings, []string)) runtimeFieldSetter {
	return func(settings *RuntimeSettings, value interface{}) error {
		v, ok := value.([]string)
		if !ok {
			return fmt.Errorf("%s must be a []string, got %T", key, value)
		}
		assign(settings, v)
		return nil
	}
}
