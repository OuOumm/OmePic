# 后端全面审查报告：AVIF runtime settings

- 日期：2026-05-14
- 任务：`.trellis/tasks/05-14-admin-avif-conversion-settings`
- 审查范围：`backend/` 全部 Go 代码，重点审查本次 AVIF 质量/速度运行时配置改动；同时回看 API、持久化、上传去重、错误处理、安全和测试覆盖。
- 依据：任务 `prd.md` / `design.md` / `implement.md`，以及 `.trellis/spec/backend/`、`.trellis/spec/frontend/` 相关规范。

## 审查结论

本次 AVIF 配置任务的后端主链路已满足 PRD，且本轮已完成 B01-B07 后端整改；复核阶段新增完成 C01-C02 小修：

- `RuntimeSettings` / `RuntimeSettingsUpdateInput` 已包含 `avif_quality`、`avif_speed`。
- 默认配置通过 `RuntimeSettingsToConfigValues(defaultRuntimeSettings())` 进入 `InsertMissingConfigValues`，缺失 key 会写入 `60` / `8`，已有 key 不会被覆盖。
- `ValidateRuntimeSettingsInput` 拒绝 `avif_quality` 越界 `0..100`、`avif_speed` 越界 `0..10`。
- 上传在原始字节 MD5 去重之后才调用 transformer；新物理对象转换时传入当前 runtime settings 的质量/速度。
- 重复上传路径仅插入新逻辑记录并复用已有 `file_path`，不调用 AVIF transformer。
- 后端测试覆盖默认持久化、不覆盖已有值、范围校验、无部分保存、转换参数透传、重复上传不重转。

## 已修复项

### F01 / 低 / 文档函数名过期

- 文件：`docs/CODE_WIKI.md`
- 问题：文档仍写 `convertToAVIF(data)`，与当前代码 `convertToAVIFWithSettings(data, settings)` 不一致。
- 处理：已更新为“按 runtime AVIF 质量/速度参数转换图片”。

### B01 / 高 / 上传与删除全局互斥锁影响吞吐

- 文件：`backend/internal/service/image_service.go`、`backend/internal/service/image_service_test.go`
- 处理：移除覆盖整个 `Upload()` / `Delete()` 的 `operationMux`；上传改为按 `storage_key + 原始字节 MD5` 的 keyed mutex 保护去重/写入窗口，避免不同 MD5 上传和删除被单把全局锁串行化。
- 测试：新增并发测试，验证同一 storage/MD5 不会并发转换、不同 MD5 可在一个转换阻塞时继续进入转换。
- 剩余风险：该 keyed mutex 仍是进程内保护，不提供跨进程/多实例互斥；如后续支持多实例高并发去重，应继续下沉到 SQLite 显式事务/唯一约束或分布式锁，并设计孤儿对象清理。

### B02 / 中 / Repository 文件过大、职责聚合

- 文件：`backend/internal/repository/*.go`
- 处理：保持 `Repository` 对外方法不变，按领域拆分为 `migration.go`、`config_repository.go`、`storage_repository.go`、`image_repository.go`、`ip_ban_repository.go`、`announcement_repository.go`、`helpers.go`；`repository.go` 仅保留构造、连接、Ping 和 SQLite PRAGMA。
- 测试：现有 repository/service/handler 测试全量通过，证明 API 兼容。

### B03 / 中 / HTTP handler 直接持有 storage manager 打开文件

- 文件：`backend/internal/service/image_service.go`、`backend/internal/http/handler/image_handler.go`
- 处理：新增 `ImageService.Open(ctx, uid)`，在 service 层完成 UID 解析、storage provider 解析和对象打开；`ImageHandler.Serve()` 只负责状态映射、响应头和流式输出。
- 测试：现有 `TestServeStreamsStoredAVIFByUIDRoute` / `TestServeRejectsBareUIDRoute` 继续覆盖公共图片读取行为。

### B04 / 中 / 错误消息清洗依赖字符串切分

- 文件：`backend/internal/service/errors.go`、`backend/internal/http/handler/errors.go` 及相关 service 调用点
- 处理：新增显式 `UserError` / `WithUserMessage()` / `UserMessage()`；handler 不再使用 `strings.SplitN` 从错误文本中切消息，而是仅返回显式标记为用户安全的消息或 fallback。
- 测试：现有密码修改、运行设置、上传/存储错误测试继续验证客户端消息。

### B05 / 中 / 健康检查 handler 直接 Ping repository/cache

- 文件：`backend/internal/service/health_service.go`、`backend/internal/http/handler/health_handler.go`、`backend/cmd/server/main.go`
- 处理：新增 `HealthService.Check()` 封装 SQLite/Redis 健康检查；handler 只调用服务并返回 JSON 响应。

### B06 / 低 / 默认管理员密码字面量散落

- 文件：`backend/internal/service/admin_service.go`、`backend/internal/service/admin_service_test.go`、`backend/internal/http/handler/admin_handler_test.go`
- 处理：提取 `DefaultAdminPassword = "admin123"`，注释说明仅 first-boot 兼容默认值且管理员应立即修改；测试改为复用该常量。

### B07 / 低 / UID generator/validator setter 暴露且无并发保护

- 文件：`backend/internal/service/image_service.go`、`docs/backend-architecture.md`
- 处理：删除未使用的 `SetUIDGenerator` / `SetUIDValidator`，继续通过构造函数注入 UID 生成与校验函数；文档同步移除 setter 表述。

### C01 / 低 / Handler 对 wrapped not_found 判断不稳健

- 文件：`backend/internal/http/handler/image_handler.go`
- 问题：`ImageHandler.Serve()` / `mapJSONError()` 使用 `err == service.ErrNotFound`，若后续 service 返回 wrapped `ErrNotFound` 会被误映射为 503 或通用错误。
- 处理：改为 `errors.Is(err, service.ErrNotFound)`；JSON not_found 消息仍通过 `UserMessage` 安全回退。

### C02 / 低 / Content-Disposition 文件名未集中转义

- 文件：`backend/internal/service/image_service.go`
- 问题：`ImageService.Open()` 生成 `Content-Disposition` 时直接拼接 basename。当前 object key 来自 UID，风险很低，但集中在 service 层后应显式处理响应头值。
- 处理：新增 `contentDispositionForPath()`，去除 CR/LF 并转义反斜杠与双引号，避免未来存储 key 来源变化时产生响应头注入风险。

## 测试与覆盖观察

- 已覆盖：本次 AVIF 配置新增单元测试覆盖核心验收项；B01 新增并发上传去重范围测试。
- 仍建议后续补充：
  - `runtimeSettingsFromValues` 对已存在但越界的 SQLite 配置值返回 `invalid_input` 的直接测试。
  - `PUT /admin/system-settings` handler 层对 AVIF 越界返回 JSON `invalid_input` 的集成/handler 测试。
  - 若支持多实例部署，围绕跨进程并发重复上传、对象写入冲突与孤儿对象清理新增专门测试。

## 自动化验证结果

| 命令 | 结果 |
|---|---|
| `cd backend && go test ./...` | 通过，108 passed / 16 packages；复核后再次通过 |
| `cd backend && go test -race ./internal/service` | 通过，39 passed / 1 package |
| `cd backend && go vet ./...` | 通过，无警告；复核后再次通过 |
| `cd backend && go build ./...` | 通过 |
| `cd backend && gofmt -l .` | 通过，无输出 |
| `cd frontend && npm run typecheck` | 通过，0 errors / 0 warnings |
| `git diff --check` | 通过，无空白问题；复核后再次通过 |

## 剩余风险

- B01 已消除全局进程锁导致的跨 MD5 上传/删除串行化，但 keyed mutex 仍不是跨进程一致性机制；多实例部署仍需 repository 事务/唯一约束/分布式锁级别方案。
- Repository 拆文件仅做物理拆分，未改变 SQL 行为；更深层事务边界优化应作为独立任务处理。
