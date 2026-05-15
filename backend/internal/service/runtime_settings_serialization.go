package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// serializeValue serializes a Go value to a string for database persistence.
// String slices keep the existing comma-separated SQLite representation.
func serializeValue(value interface{}, fieldType FieldType) string {
	switch fieldType {
	case FieldTypeString:
		v, ok := value.(string)
		if !ok {
			return ""
		}
		return v
	case FieldTypeInt:
		v, ok := value.(int)
		if !ok {
			return "0"
		}
		return strconv.Itoa(v)
	case FieldTypeBool:
		v, ok := value.(bool)
		if !ok {
			return "false"
		}
		return boolStringValue(v)
	case FieldTypeStringSlice:
		v, ok := value.([]string)
		if !ok || len(v) == 0 {
			return ""
		}
		return strings.Join(v, ",")
	default:
		return ""
	}
}

// deserializeValue deserializes a string value from database persistence back to a Go value.
// String slices primarily use CSV and also accept JSON arrays for compatibility
// with databases touched by earlier development builds.
func deserializeValue(valueStr string, fieldType FieldType) (interface{}, error) {
	switch fieldType {
	case FieldTypeString:
		return valueStr, nil
	case FieldTypeInt:
		parsed, err := strconv.Atoi(strings.TrimSpace(valueStr))
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %w", err)
		}
		return parsed, nil
	case FieldTypeBool:
		return parseBoolValue(valueStr), nil
	case FieldTypeStringSlice:
		valueStr = strings.TrimSpace(valueStr)
		if strings.HasPrefix(valueStr, "[") {
			var result []string
			if err := json.Unmarshal([]byte(valueStr), &result); err == nil {
				return result, nil
			}
		}
		return splitCSV(valueStr), nil
	default:
		return nil, fmt.Errorf("unknown field type: %v", fieldType)
	}
}
