# backend 上传流程 Seam 整改上下文

## 当前状态机

1. **准备**：校验 token、IP ban、维护模式；获取 runtime policy 快照；读取上传源并计算原始字节 MD5；校验大小和 MIME；解析当前/选中存储实例；构造 storage-scoped MD5 key；获取该 key 的 hash lock。
2. **复用检查**：通过 `md5MappingFlow.FindReusableObject` 查询 Redis scoped MD5 映射，必要时回退 SQLite 并修复/清理 stale cache。
3. **重复提交**：生成新 UID，只写新的 SQLite 图片行与 UID cache；复用已有物理对象，不重新 AVIF 转码、不调用 `SaveStream`。
4. **新物理写入**：用新 UID 构造 `.avif` object key；以 runtime `AvifQuality/AvifSpeed` 流式编码 AVIF 并 `SaveStream`；再提交 SQLite 行、UID cache、MD5 映射。
5. **失败回滚**：上传源与 hash lock 由上传事务统一释放；新物理对象已保存但 SQLite 记录提交失败时删除；SQLite 记录已提交后的 Redis UID/MD5 发布失败应返回依赖错误而不删除已被 SQLite 引用的物理对象；编码/存储流任一侧失败时通过 pipe close 解锁另一侧，避免挂起。
6. **修复/预热**：MD5 映射 stale 修复、删除后重指向、预热批量写入归 `md5MappingFlow` 所有，调用者不拼 Redis key。

## Seam 边界

- `uploadFlow.Run` 只负责创建一次 `uploadTransaction` 并提交；准备细节、锁生命周期、复用/新写入分支、清理规则由 transaction 内聚。
- `uploadTransaction` 是上传流程的深模块边界：调用者不直接传递 `policy/source/storage/md5Key/unlockFn` 字段组合。
- UID cache 与 MD5 mapping cache 在测试夹具中分开；上传测试只通过聚合 fake 读取统计，避免把健康检查或预热细节混入普通上传语义。
- 上传视角的存储依赖重点是 `SaveStream` 与失败清理所需的 `Delete`；新增 `fakeStreamStorageProvider` 明确重复对象不应再次流式写入。

## 必须保持的契约

- MD5 基于原始上传字节，且在 AVIF 转换前用于 storage-scoped 去重。
- 重复上传只复用已有物理对象，不重新编码、不再次 `SaveStream`。
- 新物理对象编码参数来自当前 runtime settings 的 `AvifQuality` / `AvifSpeed`。
- SQLite 是源事实；Redis UID/MD5 映射写入失败仍应作为依赖错误暴露，避免默默破坏一致性。
- 新物理对象已经保存但 SQLite 记录提交失败时，必须尝试删除该新对象，且不得发布 MD5 映射。
