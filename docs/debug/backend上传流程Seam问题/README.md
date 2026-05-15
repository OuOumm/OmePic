# backend 上传流程 Seam 问题

## Files

- `backend/internal/service/image_service.go`
- `backend/internal/service/upload_flow.go`
- `backend/internal/service/md5_mapping_flow.go`
- `backend/internal/service/image_transform.go`
- `backend/internal/cache/redis_cache.go`
- `backend/internal/storage/storage.go`
- `backend/internal/model/md5_mapping.go`

## Problem

上传链路已经从单个大 Implementation 拆出 `uploadFlow` 与 `md5MappingFlow`，这是明显进步；但当前 Module 的 Seam 仍然泄漏较多上传顺序知识：

1. `uploadFlow.Run` 需要按「准备上传源 → MD5 映射查找 → UID 生成 → 重复提交或物理写入 → SQLite/Redis 写入 → MD5 映射写入」的顺序手动编排。
2. `uploadPreparedContext` 暴露 `policy/source/storage/md5Key/unlockFn`，调用者仍要知道锁、清理、AVIF 参数、存储实例和 MD5 映射如何组合。
3. `storage.Provider` Interface 同时提供 `Save` 与 `SaveStream`，上传流式写入只需要后者，但测试和调用者仍要满足完整 Interface。
4. `cache.ImageCache` Interface 同时包含 UID 缓存、MD5 映射、批量预热和健康检查，导致上传测试需要构造过宽 Adapter。

删除测试：如果删除 `uploadFlow`，复杂度会基本回到 `ImageService.Upload`，说明它开始有 Depth；但如果删除 `uploadPreparedContext`，许多字段只是换地方传参，复杂度没有明显消失，说明这个中间 Module 偏 Shallow。

## Solution

先不要新增 Go interface 代码。建议把上传流程继续 Deepen 成更贴近领域概念的少数 Module：

- 将「上传准备 + 策略快照 + 存储解析 + 原始 MD5 scoped key + 清理/解锁」收束成一个更 Deep 的上传上下文 Module，让调用者只表达“创建一次可提交的上传事务”。
- 将「重复对象复用」与「新物理对象写入」作为上传事务背后的 Implementation 细节，而不是让 `Run` 直接携带所有顺序知识。
- 将 UID 缓存与 MD5 映射缓存的 Seam 拆清楚：不是立即设计新 Interface，而是先在文档和测试中区分“图片查找缓存”和“原始 MD5 映射缓存”。
- 将 `storage.Provider` 的流式持久化需求从上传视角描述清楚，避免所有上传测试都必须理解 `Open/Delete/Save` 等不相关能力。

## Benefits

- **Locality**：上传顺序、锁生命周期、清理规则、Redis/SQLite 一致性规则集中在一个更 Deep 的 Module 中，后续修改去重或 AVIF 转换参数时不需要跨 `ImageService`、`uploadFlow`、`md5MappingFlow` 和测试 Adapter 同步理解。
- **Leverage**：调用者只穿过一个更小的 Seam 即可覆盖上传、重复复用、新物理写入、失败回滚等行为。
- **测试改善**：上传测试可以少构造 Redis/Storage 的过宽 Adapter，更多通过上传 Module 的 Interface 验证“重复不转码”“新对象用 runtime AVIF 参数”“失败清理”这些高价值行为。

## 删除测试

- 删除 `uploadFlow`：复杂度会重新散落回 `ImageService.Upload`，说明它有保留价值。
- 删除 `uploadPreparedContext`：复杂度主要变成 `prepare` 返回更多局部变量，说明它当前 Depth 不足，应 Deepen 而不是继续平铺字段。
- 删除 `md5MappingFlow`：Redis stale 修复、SQLite fallback、删除后修复会散落到上传、Resolve、Delete、Preheat，说明该 Module 有较高 Leverage，适合继续保留并强化其 Interface。

## 证据引用

- `backend/internal/service/image_service.go:168` — `ImageService.Upload` 只委托给 `newUploadFlow(...).Run()`，外部 Seam 已很小。
- `backend/internal/service/upload_flow.go:53` — `uploadFlow.Run` 仍手动编排去重、UID、写物理对象、提交记录。
- `backend/internal/service/upload_flow.go:81` — `uploadFlow.prepare` 同时处理 token、IP ban、runtime policy、temp file、MIME、storage、MD5 key、hash lock。
- `backend/internal/service/upload_flow.go:150` — `commitDuplicate` 直接写 SQLite 与 UID cache。
- `backend/internal/service/upload_flow.go:187` — `commitNewPhysical` 直接写 SQLite、UID cache、MD5 cache，并处理物理删除回滚。
- `backend/internal/service/md5_mapping_flow.go:40` — `FindReusableObject` 同时处理 Redis 命中、stale UID、SQLite fallback、cache repair。
- `backend/internal/cache/redis_cache.go:13` — `ImageCache` 同时覆盖 UID 缓存、MD5 映射、批量预热、Ping。
- `backend/internal/storage/storage.go:24` — `Provider` 同时覆盖 Save、SaveStream、Open、Delete。

## 整改步骤

1. 为上传流程画出当前状态机：准备、复用、写入、提交、修复、失败回滚。
2. 标注哪些知识应留在外部 Interface，哪些应成为内部 Implementation。
3. 先收敛测试命名与夹具：把 UID cache Adapter、MD5 mapping Adapter、stream storage Adapter 的职责分开描述。
4. 再逐步 Deepen 上传上下文 Module，使 `Run` 的实现更接近领域步骤，而不是资源字段搬运。
5. 最后再考虑是否拆分实际 Go seam；只有当测试出现两个真实 Adapter 或生产路径需要替换行为时再落地。

## 验证建议

- `cd backend && go test ./internal/service -run 'TestUpload|TestMD5|TestPreheat'`
- `cd backend && go test ./internal/http/handler -run TestRuntimeSettingsReturnsSafePublicCatalog`
- 增加聚焦测试：上传源准备失败时 hash lock 不泄漏、重复上传不触发 stream storage Adapter、Redis stale MD5 自动修复。
