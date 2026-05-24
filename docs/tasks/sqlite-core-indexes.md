# 任务：SQLite 核心索引

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.2 性能瓶颈与整改建议** — "SQLite 图片列表与滥用统计可能随数据量退化"
- **§3.1.3 数据库索引调优**
- **§4.2 P0：SQLite 核心索引** — "为 UID、MD5、时间、Token、IP、存储查询补齐索引与迁移测试"

## 预估人日

1.5 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/database/migration.go` — 新增索引迁移
- `backend/internal/repository/image_repo.go` — 确认查询利用索引

### 关键实现点

1. 补齐复合索引：
   - `CREATE INDEX IF NOT EXISTS idx_images_uid ON images(uid)`
   - `CREATE INDEX IF NOT EXISTS idx_images_storage_md5 ON images(storage_key, md5_hash)`
   - `CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at DESC)`
   - `CREATE INDEX IF NOT EXISTS idx_images_token_created_at ON images(token, created_at DESC)`
   - `CREATE INDEX IF NOT EXISTS idx_images_ip_created_at ON images(ip_address, created_at DESC)`
   - `CREATE INDEX IF NOT EXISTS idx_images_storage_created_at ON images(storage_key, created_at DESC)`
2. 迁移保持幂等：`CREATE INDEX IF NOT EXISTS`。
3. 确认现有查询的 WHERE/ORDER BY 能利用新索引。
4. 大表场景验证索引对查询性能的提升。

### 完成标准

- [x] 代码编译通过：`cd backend && go build ./...`
- [x] 单元测试通过：`cd backend && go test ./...`
- [x] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [x] 迁移幂等：重复启动不报错

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 首次启动，索引创建成功 | 日志无错误 |
| 正向 | 重启，索引已存在 | `IF NOT EXISTS` 保证无报错 |
| 正向 | 按 Token 查询图片列表 | 查询使用索引，响应快速 |
| 正向 | 按 IP 查询上传历史 | 查询使用索引 |
| 边界 | 空表时创建索引 | 正常创建 |
| 边界 | 大数据量下查询 | 索引显著降低查询时间 |

### 必须覆盖的测试类型

- 单元测试：迁移脚本幂等性
- 集成测试：迁移后查询正确性
- 性能测试：大数据量下有无索引的查询对比（可选）

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 数据库升级无报错

## 复查

### 代码审查关注点

- 索引是否覆盖所有高频查询模式
- 是否存在冗余索引（已包含在复合索引中的前缀列）
- 索引对写入性能的影响评估

### 安全检查

- 无安全敏感项

### 文档更新要求

- 数据库 Schema 文档更新索引列表

### 复查通过标准

- [x] 本地实现复核通过
- [x] 本地验证命令通过
- [x] 完成报告已填写

## 完成报告

### 完成时间

2026-05-24

### 实现内容

- 在 `backend/internal/repository/migration.go` 的启动迁移中补齐 SQLite 图片核心索引：
  - `idx_images_uid ON images(uid)`
  - `idx_images_storage_md5 ON images(storage_key, md5_hash)`
  - `idx_images_created_at ON images(created_at DESC)`
  - `idx_images_token_created_at ON images(token, created_at DESC)`
  - `idx_images_ip_created_at ON images(ip_address, created_at DESC)`
  - `idx_images_storage_created_at ON images(storage_key, created_at DESC)`
- 迁移保持幂等，并能将同名旧定义索引安全替换为本任务要求的新定义。
- 用户确认取消回收站功能后，已移除 deleted_at 相关索引要求与实现。
- 在 `backend/internal/repository/repository_test.go` 增加迁移测试，覆盖：重复迁移幂等、核心索引存在且定义正确。

### 测试与校验记录

- `cd backend && gofmt -w ./cmd ./internal`：通过。
- `cd backend && go test ./...`：通过（16 个 package，224 个测试）。
- `cd backend && gofmt -l ./cmd ./internal`：通过，无输出。
- `cd backend && go build ./...`：通过。

### 复核结论

- 需求一致性：已按 P0「SQLite 核心索引」范围补齐要求索引；未实现其他 P0 子任务。
- Schema 边界：用户确认取消回收站功能后，当前 schema 不包含删除元数据字段。
- 测试一致性：新增测试明确验证迁移幂等与核心索引存在。
- 风险/后续：保留旧单列索引以避免影响既有查询；后续可结合真实慢查询再评估冗余索引清理。
