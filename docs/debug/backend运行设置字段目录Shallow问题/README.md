# backend 运行设置字段目录 Shallow 问题

## Files

- `backend/internal/service/runtime_settings.go`
- `backend/internal/service/runtime_settings_fields.go`
- `backend/internal/service/runtime_settings_accessors.go`
- `backend/internal/service/runtime_settings_serialization.go`
- `backend/internal/repository/config_repository.go`
- `.trellis/spec/backend/runtime-settings.md`

## Problem

当前 Runtime Settings 为 AVIF 参数新增了字段目录，这是为了减少默认值、持久化 key、序列化和反序列化的漂移。但字段目录 Module 目前仍偏 Shallow：

1. `ConfigField` 的 Interface 要求维护者理解 `Key/Type/Default/Get/Set` 五种信息，但 `defaultRuntimeSettings()` 仍单独维护默认值。
2. `RuntimeSettingsToConfigValues` 使用 `normalizeRuntimeSettings(settings)` 后再从字段目录读取；`runtimeSettingsFromValues` 又从 `defaultRuntimeSettings()` 开始反序列化。默认值来源仍有两处。
3. `runtime_settings_accessors.go` 只包装 `GetFieldByKey` 后调用 `Get/Set`，删除后复杂度不会显著外溢，是兼容测试用的浅 Module。
4. `ConfigField.Default` 当前主要是描述信息，没有成为 `defaultRuntimeSettings()` 的真正来源，容易给维护者造成“已经有单一来源”的错觉。

删除测试：如果删除 `runtime_settings_accessors.go`，调用方可直接使用 `GetFieldByKey`，复杂度几乎不变，说明该 Module 没有足够 Depth。如果删除 `ConfigField.Default`，当前运行时行为也基本不变，说明字段目录还没有真正形成 Deep 的单一来源。

## Solution

先制定整改路线，不直接改 Go 代码：

- 决定 Runtime Settings 的“字段目录”是否要成为真正 Deep 的 Module。
- 如果要保留字段目录，应让默认 runtime config、持久化值、反序列化、测试覆盖都从同一个目录得到 Leverage。
- 如果不准备让它承载默认值，就移除或弱化 `Default` 的语义，避免 Interface 承诺超过 Implementation。
- 合并或删除只用于转发的访问器 Module，让测试穿过真实的字段目录 Seam。

## Benefits

- **Locality**：新增类似 `avif_quality` / `avif_speed` 的字段时，默认值、SQLite key、序列化、反序列化、测试断言集中在一个位置。
- **Leverage**：字段目录一旦 Deep，新增字段只需少量声明即可自动获得持久化、读取、默认写入和测试枚举能力。
- **测试改善**：可以从“断言某几个 key”改为“字段目录中的每个 key 都可默认持久化、反序列化并通过验证”，减少漏测。

## 删除测试

- 删除 `runtime_settings_accessors.go`：复杂度不会显著增加，说明它当前可被合并或降级为测试 helper。
- 删除 `ConfigField.Default`：当前默认值仍来自 `defaultRuntimeSettings()`，说明 `Default` 没有承担关键 Implementation。
- 删除 `runtimeConfigFields`：序列化与读取循环会重新散落成 switch/手写 map，说明字段目录方向有价值，但需要 Deepen。

## 证据引用

- `backend/internal/service/runtime_settings_fields.go:23` — `ConfigField` 同时声明 key、类型、默认值、getter、setter。
- `backend/internal/service/runtime_settings_fields.go:31` — `runtimeConfigFields` 已包含 `avif_quality`、`avif_speed` 等配置。
- `backend/internal/service/runtime_settings.go:267` — `RuntimeSettingsToConfigValues` 遍历字段目录序列化。
- `backend/internal/service/runtime_settings.go:282` — `runtimeSettingsFromValues` 从 `defaultRuntimeSettings()` 而不是字段目录默认值开始。
- `backend/internal/service/runtime_settings.go:305` — `defaultRuntimeSettings` 仍独立维护所有默认字段。
- `backend/internal/service/runtime_settings_accessors.go:6` — `getFieldValue` 是窄转发 helper。
- `backend/internal/service/runtime_settings_accessors.go:15` — `setFieldValue` 是窄转发 helper。

## 整改步骤

1. 先列出每个 Runtime Settings 字段的唯一事实源：key、默认值、类型、验证、公开/后台可见性。
2. 决定字段目录承载哪些事实，不承载哪些事实，避免半单一来源。
3. 如果字段目录承载默认值，则让 `defaultRuntimeSettings()` 从字段目录生成或至少由字段目录测试强制一致。
4. 删除或内联纯转发访问器，让测试直接穿过真实字段目录 Seam。
5. 为字段目录增加表驱动验证：每个字段可 serialize/deserialize；默认持久化不会漏 key；新增字段若未覆盖测试则失败。

## 验证建议

- `cd backend && go test ./internal/service -run 'TestRuntimeSettings|TestDefault|TestMissing|TestAVIF'`
- 增加表驱动测试：`runtimeConfigFields` 中每个字段都能出现在 `RuntimeSettingsToConfigValues(defaultRuntimeSettings())`。
- 增加删除测试式审查：新增 Runtime Settings 字段时只允许编辑一个事实源和验证函数。
