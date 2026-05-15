# backend 缓存 Interface 过宽问题

## Files

- `backend/internal/cache/redis_cache.go`
- `backend/internal/service/image_service.go`
- `backend/internal/service/md5_mapping_flow.go`
- `backend/internal/http/handler/image_handler_test.go`
- `backend/internal/service/image_service_test.go`
- `backend/internal/cache/redis_cache_test.go`

## Problem

`cache.ImageCache` 这个 Interface 当前覆盖了两个不同领域：UID 图片缓存与原始 MD5 去重映射，还包含批量预热和健康检查。它对生产 Redis Adapter 方便，但对调用者和测试是过宽 Seam：

1. `ImageService.Resolve/Delete/Preheat` 主要使用 UID 缓存方法。
2. `md5MappingFlow` 主要使用 MD5 映射方法。
3. `HealthService` 或启动检查可能只需要 `Ping`。
4. Handler/Service 测试中的 fake cache 必须实现未使用的方法，增加噪音。
5. `md5MappingFlow` 已经是一个有 Depth 的 Module，但它的缓存 Adapter 仍暴露在总 `ImageCache` Interface 上，调用者容易继续把 UID 与 MD5 key 规则混在一起。

删除测试：如果删除 `ImageCache`，Redis 实现的方法不会减少，只是每个调用点都直接依赖 RedisCache；复杂度会外溢，说明需要缓存 Seam。但如果把 `Ping` 或批量 Set 与上传路径放在同一 Interface 中，删除这些方法对上传行为无影响，说明当前 Interface 比使用场景宽。

## Solution

先不写具体 Go interface。候选路线：

- 按调用意图拆分缓存 Seam 的语言：图片查找缓存、MD5 映射缓存、健康检查、预热批量写入。
- 保留一个 Redis Adapter 可以满足多个 Seam，但让不同 Module 只学习自己需要的 Interface。
- 让 `md5MappingFlow` 成为唯一了解 MD5 cache 行为的 Module，上传/删除/Resolve 只通过它表达领域意图。
- 在测试中先拆 fake：即使生产 Interface 还没拆，也先减少测试夹具对无关方法的关注。

## Benefits

- **Locality**：UID 缓存 key、MD5 映射 key、预热策略、健康检查错误处理分别集中，不会因为一个 Interface 变动影响所有测试。
- **Leverage**：Redis Adapter 仍复用同一 Implementation，但调用者只穿过更窄、更贴近意图的 Seam。
- **测试改善**：上传测试只需要 MD5 映射与 UID 写入 Adapter；Serve 测试只需要 UID 读取 Adapter；健康测试只需要 Ping Adapter。

## 删除测试

- 删除整个 `ImageCache`：复杂度会散到调用者，说明缓存 Seam 本身必要。
- 删除 `Ping` 对上传测试 fake 的要求：上传测试不受影响，说明 `Ping` 不属于上传路径 Interface。
- 删除 MD5 方法对 Serve-only 测试 fake 的要求：Serve 行为不受影响，说明 UID 与 MD5 应有不同使用视角。

## 证据引用

- `backend/internal/cache/redis_cache.go:13` — `ImageCache` 同时包含 UID、MD5、批量预热、Ping。
- `backend/internal/service/image_service.go:219` — `Resolve` 使用 `GetImage`。
- `backend/internal/service/image_service.go:251` — `Resolve` cache miss 后使用 `SetImage` 并调用 MD5 backfill。
- `backend/internal/service/image_service.go:266` — `Preheat` 使用 `SetImages` 后调用 MD5 preheat。
- `backend/internal/service/md5_mapping_flow.go:40` — `FindReusableObject` 使用 MD5 cache 与 SQLite fallback。
- `backend/internal/http/handler/image_handler_test.go:33` — handler 测试 fake cache 实现整套 `ImageCache` 方法。
- `backend/internal/cache/redis_cache_test.go:41` — Redis key shape 测试关注 MD5 key。

## 整改步骤

1. 统计每个 Module 实际使用的 cache 方法，形成方法-调用者矩阵。
2. 在测试 fake 中按意图区分 UID cache 与 MD5 mapping cache 的职责。
3. 将 `md5MappingFlow` 的外部 Interface 进一步固定为领域操作，不让其他调用者组合 MD5 cache 方法。
4. 再评估是否把生产 `ImageCache` 拆成多个小 Seam，由同一 Redis Adapter 满足。
5. 更新 backend quality spec 中“Prefer small interfaces”的具体缓存示例。

## 验证建议

- `cd backend && go test ./internal/service -run 'TestMD5|TestUpload|TestPreheat|TestResolve'`
- `cd backend && go test ./internal/http/handler -run TestServe`
- `cd backend && go test ./internal/cache`
