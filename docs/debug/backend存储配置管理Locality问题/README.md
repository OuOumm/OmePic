# backend 存储配置管理 Locality 问题

## Files

- `backend/internal/service/admin_service.go`
- `backend/internal/repository/storage_repository.go`
- `backend/internal/storage/storage.go`
- `backend/internal/config/config.go`
- `.trellis/spec/backend/database-guidelines.md`

## Problem

存储配置管理涉及 RuntimeStorageConfig、SQLite `storage_configs`、StorageManager 热重载、后台 masked view、默认实例切换、禁止 in-use backend 变更等规则。当前这些规则分散在多个 Module 中，Locality 不足：

1. `AdminService.UpdateConfig` 兼容旧 flat route，并手动决定 patch target/default switch 顺序。
2. `AdminService.UpdateStorageConfig` 负责 merge、in-use 检查、名称验证、Provider 验证、SQLite 更新、StorageManager reload。
3. `Repository.InitializeStorageCatalog` 又负责 legacy config 迁移、默认项归一、image storage_key 回填。
4. `storage.Manager.Reconfigure` 与 `storage.ValidateConfig` 负责另一部分 normalization/validation。
5. secret mask/保留逻辑在 `AdminService.mergeStorageConfig` 和 `maskStorageConfig`，与持久化 Module 相隔较远。

这些规则都围绕“存储实例目录”这个领域概念，但没有一个 Deep Module 把它们收束。维护者修改后台存储配置时，必须同时理解 Admin、Repository、Storage、Config 四处 Implementation。

删除测试：如果删除 `AdminService.UpdateStorageConfig` 的若干 helper，复杂度不会消失，只会散回 `UpdateStorageConfig`；如果删除 `Repository.InitializeStorageCatalog`，迁移和回填规则会散到启动 wiring，说明它有价值但承担了过多非 SQLite 语义。

## Solution

先提出候选整改路线，不设计具体 Go interface：

- 将“存储实例目录”作为一个候选 Deep Module 来梳理：它的 Interface 应表达创建、patch、删除、设默认、初始化、masked view、热重载这些意图，而不是暴露后台 route 的 patch 细节。
- 把 legacy flat route 的兼容规则保留为 Adapter，避免它污染主要存储实例目录 Interface。
- 把 secret mask/保留作为存储实例目录的 Implementation 规则，而不是由后台调用者记住。
- 把 Provider 构造验证与配置 normalization 的关系写清：`storage.ValidateConfig` 是存储 Adapter 能力验证，还是目录规则验证；两者不要混在同一个 Seam。

## Benefits

- **Locality**：默认实例、backend 不可变、secret 保留、StorageManager reload、SQLite 写入顺序集中，减少“先写库再失败”的部分保存风险。
- **Leverage**：后台旧 route、新 storage routes、启动初始化都可复用同一个存储实例目录 Module 的行为。
- **测试改善**：可以通过存储实例目录 Interface 做高层测试：patch + missing default 不保存、in-use backend 变更拒绝、masked secret 不覆盖、删除 default 拒绝、reload 失败行为一致。

## 删除测试

- 删除 `UpdateConfig` 的兼容 target 选择逻辑：复杂度会重新出现在 handler 或前端兼容层，说明需要一个 Adapter 承载 legacy 输入转换。
- 删除 `Repository.InitializeStorageCatalog`：启动初始化、legacy seed、storage_key 回填会散落，说明初始化 Module 有 Depth，但应与普通 CRUD 规则协调。
- 删除 `storage.ValidateConfig`：Provider 构造错误会进入更晚阶段，说明它有 Leverage；但它不应独自承担目录级规则。

## 证据引用

- `backend/internal/service/admin_service.go:439` — `AdminService.UpdateConfig` 处理 legacy flat route、patch target、default switch 顺序。
- `backend/internal/service/admin_service.go:501` — `UpdateStorageConfig` 合并 patch、检查 backend 变更、验证、持久化、reload。
- `backend/internal/service/admin_service.go:539` — `DeleteStorageConfig` 同时检查 default、in-use、删除、reload。
- `backend/internal/service/admin_service.go:569` — `SetDefaultStorageConfig` 更新默认实例后 reload。
- `backend/internal/service/admin_service.go:624` — `reloadStorageManager` 从 SQLite 重读后调用 `storage.Reconfigure`。
- `backend/internal/repository/storage_repository.go:16` — `InitializeStorageCatalog` 处理初始化、默认归一、历史图片回填。
- `backend/internal/repository/storage_repository.go:258` — `normalizeDefaultStorageConfig` 修改默认标记。
- `backend/internal/repository/storage_repository.go:292` — `backfillImageStorageKeys` 处理历史记录 storage_key。
- `backend/internal/storage/storage.go:95` — `Manager.Reconfigure` normalize 并重建 providers。
- `backend/internal/storage/storage.go:114` — `ValidateConfig` 通过构造 Provider 验证配置。

## 整改步骤

1. 梳理存储实例目录的领域规则，分出目录规则、Provider Adapter 规则、legacy 输入 Adapter 规则。
2. 给现有测试加上“部分保存”断言，覆盖 patch + default switch、reload 失败、masked secret 等组合。
3. 把 legacy flat route 输入转换为单独整改目标，先让主流程只处理规范化后的存储目录意图。
4. 再决定是否提取 Deep Module，将 SQLite 写入和 StorageManager reload 的成功/失败语义集中。
5. 更新 `.trellis/spec/backend/database-guidelines.md`，明确普通编辑与启动初始化各自的 Seam。

## 验证建议

- `cd backend && go test ./internal/service -run 'Test.*Storage|Test.*Config'`
- `cd backend && go test ./internal/repository -run 'TestInitializeStorageCatalog|Test.*Storage'`
- 增加 focused tests：StorageManager reload 失败时后台响应与 SQLite 状态是否符合预期；masked secret 输入不会覆盖原 secret。
