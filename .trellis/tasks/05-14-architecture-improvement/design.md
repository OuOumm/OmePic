# Design: RuntimeSettings Table-Driven

## Scope

重构 `backend/internal/service/runtime_settings.go`，使用描述表驱动方式减少重复代码。

## Current State

- `RuntimeSettingsToConfigValues`: 手动映射每个字段到配置值
- `runtimeSettingsFromValues`: 14 个重复的 if 块
- `ValidateRuntimeSettingsInput`: 手动验证每个字段
- `setFieldValue`: switch 语句，不返回错误

## Target State

- 使用 `map[string]ConfigField` 存储字段定义
- `getFieldValue` 和 `setFieldValue` 通过 switch 访问字段
- `setFieldValue` 返回 `error`，避免静默失败
- 切片序列化使用 JSON 编码
- 所有序列化/反序列化/验证逻辑通过描述表驱动

## Files to Create/Modify

### New Files

1. `backend/internal/service/runtime_settings_fields.go`
   - `ConfigField` 结构体定义
   - `configFieldsMap` 变量（map 类型）
   - `GetFieldByKey` 和 `GetAllFields` 函数

2. `backend/internal/service/runtime_settings_accessors.go`
   - `getFieldValue` 函数
   - `setFieldValue` 函数（返回 error）

3. `backend/internal/service/runtime_settings_serialization.go`
   - `serializeValue` 函数（切片使用 JSON 编码）
   - `deserializeValue` 函数（支持 JSON 和逗号分隔）

### Modified Files

1. `backend/internal/service/runtime_settings.go`
   - `RuntimeSettingsToConfigValues`: 使用描述表
   - `runtimeSettingsFromValues`: 使用描述表
   - `ValidateRuntimeSettingsInput`: 使用描述表

2. `backend/internal/service/runtime_settings_test.go`
   - 添加字段 get/set 往返测试
   - 验证所有字段都能正确序列化/反序列化

## Implementation Details

### ConfigField 结构体

```go
type ConfigField struct {
    Key       string
    Type      FieldType
    Default   interface{}
    Validator func(interface{}) error
}
```

### FieldType 枚举

```go
type FieldType int

const (
    FieldTypeString FieldType = iota
    FieldTypeInt
    FieldTypeBool
    FieldTypeStringSlice
)
```

### configFieldsMap 变量

```go
var configFieldsMap = map[string]ConfigField{
    "site_name": {Key: "site_name", Type: FieldTypeString, Default: DefaultSiteName, ...},
    "site_tagline": {Key: "site_tagline", Type: FieldTypeString, Default: DefaultSiteTagline, ...},
    // ... 其他字段
}
```

### setFieldValue 函数

```go
func setFieldValue(settings *RuntimeSettings, key string, value interface{}) error {
    switch key {
    case "site_name":
        v, ok := value.(string)
        if !ok {
            return fmt.Errorf("site_name must be a string, got %T", value)
        }
        settings.SiteName = v
    // ... 其他字段
    default:
        return fmt.Errorf("unknown field: %s", key)
    }
    return nil
}
```

### serializeValue 函数

```go
func serializeValue(value interface{}, fieldType FieldType) string {
    switch fieldType {
    case FieldTypeStringSlice:
        // 使用 JSON 编码
        if v, ok := value.([]string); ok {
            if len(v) == 0 {
                return "[]"
            }
            jsonBytes, err := json.Marshal(v)
            if err != nil {
                return strings.Join(v, ",") // 回退
            }
            return string(jsonBytes)
        }
        return "[]"
    // ... 其他类型
    }
}
```

### deserializeValue 函数

```go
func deserializeValue(valueStr string, fieldType FieldType) (interface{}, error) {
    switch fieldType {
    case FieldTypeStringSlice:
        // 优先尝试 JSON 解析
        valueStr = strings.TrimSpace(valueStr)
        if strings.HasPrefix(valueStr, "[") {
            var result []string
            if err := json.Unmarshal([]byte(valueStr), &result); err == nil {
                return result, nil
            }
        }
        // 回退到逗号分隔
        return splitCSV(valueStr), nil
    // ... 其他类型
    }
}
```

## Migration Strategy

### Phase 1: Create new files
1. Create `runtime_settings_fields.go`
2. Create `runtime_settings_accessors.go`
3. Create `runtime_settings_serialization.go`

### Phase 2: Update existing code
1. Update `RuntimeSettingsToConfigValues`
2. Update `runtimeSettingsFromValues`
3. Update `ValidateRuntimeSettingsInput`

### Phase 3: Testing
1. Add unit tests for all fields
2. Run `go test ./...`
3. Verify behavior unchanged

## Compatibility

- JSON encoding for slices is backward compatible with comma-separated values
- `deserializeValue` handles both JSON and comma-separated formats
- Existing database values will be migrated on next save

## Risks

- Slice serialization format change may affect existing data
  - Mitigation: `deserializeValue` handles both formats
- Performance impact of JSON encoding/decoding
  - Mitigation: Negligible for small slices
