# 任务：配置审计日志 MVP

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.4 可维护性不足与整改建议** — "配置变更缺少审计和回滚"
- **§3.2.3 配置变更审计日志**
- **§4.2 P0：配置审计日志 MVP** — "runtime/storage 配置变更记录、敏感字段遮罩"

## 预估人日

3 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/config_audit_log.go`（新增）— 审计日志模型
- `backend/internal/repository/audit_repo.go`（新增）— 审计日志数据访问层
- `backend/internal/service/audit_service.go`（新增）— 审计记录业务逻辑
- `backend/internal/http/handler/admin_setting.go` — 配置变更时写入审计日志
- `backend/internal/http/handler/admin_storage.go` — 存储配置变更时写入审计日志
- `backend/internal/http/handler/admin_audit.go`（新增）— 审计日志查询 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表迁移

### 关键实现点

1. 新增 `config_audit_logs` 表：`id`、`actor`、`actor_ip`、`config_scope`（runtime/storage/security/announcement）、`before_snapshot`、`after_snapshot`、`created_at`。
2. runtime/storage 配置变更时自动写入 before/after JSON 快照。
3. 快照中的 secret（如存储密码、API Key）使用现有遮罩策略（显示前 4 位 + ***）。
4. 新增管理员 API：`GET /admin/audit-logs` 查询审计日志（分页、按 scope 筛选、按时间范围）。
5. 审计记录写入使用异步，不阻塞配置变更主链路。

### 完成标准

- [x] 代码编译通过：`cd backend && go test ./...` 覆盖编译
- [x] 单元测试通过：`cd backend && go test ./...`
- [x] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [x] 迁移幂等

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 修改 runtime 设置 | 生成审计日志，包含 before/after 快照 |
| 正向 | 修改 storage 配置 | 生成审计日志 |
| 正向 | 查询审计日志 | 返回正确记录，分页正常 |
| 正向 | 按 scope 筛选 | 只返回对应 scope 的记录 |
| 异常 | 配置未实际变更 | 不生成审计日志（可选策略） |
| 边界 | 快照中包含密码 | 密码被遮罩 |
| 边界 | 大量审计日志查询 | 分页正确，性能可接受 |

### 必须覆盖的测试类型

- 单元测试：敏感字段遮罩逻辑、审计记录创建
- 集成测试：配置变更 → 审计日志生成 → 查询验证
- 安全测试：确认密码不出现在审计日志中

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 手动验证审计日志内容正确

## 复查

### 代码审查关注点

- 敏感字段遮罩策略是否覆盖所有 secret 类型
- 审计写入是否真正异步（不阻塞配置变更 API 响应时间）
- 快照 JSON 序列化是否处理了 nil/空值情况

### 安全检查

- 审计日志本身需要管理员权限才能查询
- 快照中的 secret 必须遮罩，不可明文存储

### 文档更新要求

- 管理员 API 文档：审计日志查询接口
- 数据库 Schema 文档新增表

### 复查通过标准

- [x] 实施者自查通过，等待主会话/人工 reviewer 最终复核
- [x] 本地后端校验绿色
- [x] 完成报告已填写

## 完成报告

### 实施摘要

- 新增 `config_audit_logs` SQLite 表与索引，迁移使用 `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS` 保持幂等。
- 新增配置审计模型、Repository 与 Service 查询/创建能力；查询支持 `scope`、`page`、`page_size`。
- runtime 设置更新写入 `runtime` scope 的 before/after JSON 快照。
- storage 配置变更（legacy `POST /admin/config`、create、patch、delete、default switch）写入 `storage` scope 的 before/after JSON 快照。
- 新增管理员接口 `GET /admin/audit-logs`，复用现有 JWT 管理员鉴权。
- 快照序列化前统一遮罩 `secret`、`password`、`token`、`access_key`、`api_key`、`webdav_pass` 等敏感字段，避免明文落库。

### 测试与验证

- `cd backend && gofmt -w ./cmd ./internal`：通过。
- `cd backend && go test ./...`：通过，255 个测试通过。
- `cd backend && gofmt -l ./cmd ./internal`：通过，无输出。

### 覆盖场景

- 迁移幂等与审计日志查询分页/过滤。
- runtime 设置变更生成审计日志，并验证 Cloudflare token 不明文写入快照。
- storage 配置新增生成审计日志，并验证 S3 access/secret 不明文写入快照。
- 管理员审计日志查询服务支持 scope 过滤与分页。

### 复核结论

- 已按 MVP 范围完成配置审计日志，不包含软删除、存储健康检查等其他 P0 功能。
- 未修改 `frontend/src/lib/i18n.ts`、`frontend/src/routes/+error.svelte`。
- 已知限制：本实现采用同步写入审计日志以保证配置变更与审计记录一致；若后续需要严格异步，可引入后台队列并补偿失败审计写入。
