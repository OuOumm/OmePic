# 任务：存储迁移与复制

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F9** — "存储迁移与复制"
- **§4.4 P2：存储迁移与复制** — "迁移任务、双写副本、校验、断点续传"

## 预估人日

10 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/migration_job.go`（新增）— 迁移任务模型
- `backend/internal/repository/migration_repo.go`（新增）— 迁移数据访问
- `backend/internal/service/migration_service.go`（新增）— 迁移服务
- `backend/internal/http/handler/admin_migration.go`（新增）— 迁移管理 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表

### 关键实现点

1. 迁移任务模型：`source_storage`、`target_storage`、`filter`（JSON）、`status`、`progress`、`total_count`、`migrated_count`、`error_count`。
2. 迁移流程：读取源存储 → 写入目标存储 → 校验完整性 → 更新记录 storage_key。
3. 断点续传：任务中断后可从上次进度继续。
4. 双写副本：上传时同时写入主存储和副本存储。
5. 校验：迁移完成后对比源和目标的文件大小、MD5。
6. 后台任务执行，支持暂停/恢复/取消。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] 迁移任务可创建和执行

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 迁移 10 张图片到新存储 | 全部成功，校验通过 |
| 正向 | 暂停并恢复迁移任务 | 从断点继续 |
| 正向 | 双写模式上传 | 两个存储都有文件 |
| 异常 | 目标存储不可用 | 任务暂停，记录错误 |
| 边界 | 迁移 0 张图片 | 返回提示 |
| 边界 | 迁移过程中源文件被删除 | 跳过并记录 |

### 必须覆盖的测试类型

- 单元测试：迁移逻辑、断点续传、校验
- 集成测试：完整迁移流程
- 容错测试：迁移过程中存储故障

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 迁移任务的并发控制
- 断点续传的可靠性
- 数据一致性校验

### 安全检查

- 迁移 API 需要管理员权限
- 迁移过程不暴露存储凭据

### 文档更新要求

- 管理员 API 文档说明迁移接口
- 运维文档说明迁移最佳实践

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
