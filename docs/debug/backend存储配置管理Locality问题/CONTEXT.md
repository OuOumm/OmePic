# backend 存储配置管理 Locality 整改上下文

## 领域边界

存储实例目录（storage catalog）负责后台可运行时编辑的 `storage_configs` 集合，而不是单个 route 的请求形状。核心规则包括：

- 创建、更新、删除、默认实例切换都必须围绕完整目录校验。
- 已被图片引用的 `storage_key` 不能切换 `storage_backend`。
- 后台视图返回的 masked secret 原样提交时必须保留旧密钥。
- SQLite 是持久化 source of truth；成功写入后再 reload `StorageManager`。
- reload 失败要向上返回 `dependency_unavailable`，但已写入 SQLite 的状态保留，等待后续 reload / 重启恢复。
- legacy flat `/admin/config` 只是兼容 Adapter：将旧 payload 解析为“patch 目标 + 可选默认切换 + patch 字段”，再交给目录规则处理。

## 当前实现

- `backend/internal/service/storage_catalog.go` 收束 service-layer 目录规则：
  - `View`
  - `Create`
  - `Patch`
  - `Delete`
  - `SetDefault`
  - `ApplyLegacyPatch`
- `backend/internal/service/admin_service.go` 仅保留 HTTP service entrypoint，委托给 storage catalog。
- `backend/internal/storage/storage.go` 新增 `ValidateCatalog`，复用 `StorageManager.Reconfigure` 的 normalize + provider validation 逻辑，用于写 SQLite 前验证完整 post-change catalog。
- `backend/internal/repository/storage_repository.go` 仍保留启动初始化/legacy seed/image storage_key backfill 的 SQLite 初始化职责，不承载后台编辑规则。

## 关键语义

### 写入顺序

1. 从 SQLite 读取当前目录。
2. 构造 post-change catalog。
3. 校验目录规则与 provider 配置。
4. 写入 SQLite。
5. 从 SQLite 重新读取并 reload `StorageManager`。
6. 返回 masked admin view。

### Legacy Adapter

`UpdateConfig` 不再自己实现 target/default 顺序逻辑，只构造 `legacyStorageConfigPatch`。兼容性规则保持：

- 有 `storage_key` 时 patch 该实例。
- 无 `storage_key` 且有 `default_storage_key` 和 patch 字段时 patch default target。
- 两者都无但有 patch 字段时 patch 当前默认实例。
- `default_storage_key` 非空时先校验存在，避免 patch 后才失败。

## 测试补充

- masked `s3_access_key` / `s3_secret_key` 原样提交不覆盖旧 secret。
- invalid post-change catalog 在 SQLite 写入前失败，不保存 patch / default switch。
- reload 失败返回依赖错误，同时确认 SQLite 已成为新的 source of truth、内存 manager 未被错误更新。
