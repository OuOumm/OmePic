# 任务：存储健康检查 MVP

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.4 可维护性不足与整改建议** — "多存储实例缺少统一运行状态模型"
- **§1.5.2 多存储实例心跳健康检查**
- **§4.2 P0：存储健康检查 MVP** — "后台手动检测 + 定时心跳 + 状态展示"

## 预估人日

4 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/storage_health.go`（新增）— 健康检查模型
- `backend/internal/repository/storage_health_repo.go`（新增）— 健康检查数据访问层
- `backend/internal/service/storage_health_service.go`（新增）— 健康检查业务逻辑
- `backend/internal/http/handler/admin_storage.go` — 新增健康检查 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表迁移
- `backend/cmd/server/main.go` — 启动后台心跳定时任务
- `backend/internal/storage/provider.go` — 存储 Provider 接口增加健康探测方法

### 关键实现点

1. 新增 `storage_health_checks` 表：`id`、`storage_key`、`status`（healthy/degraded/unavailable）、`last_check_at`、`latency_ms`、`error_message`、`consecutive_failures`。
2. 存储 Provider 接口增加 `HealthCheck(ctx) error` 方法，执行小对象写入→读取→删除。
3. 管理员 API：
   - `POST /admin/storage/:key/health-check` — 手动触发单个存储检测
   - `POST /admin/storage/health-check-all` — 手动触发全部存储检测
   - `GET /admin/storage/health` — 查看所有存储健康状态
4. 启动后台定时心跳（默认每 5 分钟），周期执行健康检查并写入最新状态。
5. 探测使用小对象（1KB 随机数据），记录状态、延迟和错误信息。
6. 后台心跳失败仅记录日志，不阻断主服务。

### 完成标准

- [x] 代码编译通过：`cd backend && go test ./...`（覆盖编译）
- [x] 单元测试通过：`cd backend && go test ./internal/service/... ./internal/repository/...`（已由 `go test ./...` 覆盖）
- [x] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [x] 迁移幂等

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 手动检测健康存储 | 返回 healthy，记录延迟 |
| 正向 | 手动检测不可用存储 | 返回 unavailable，记录错误信息 |
| 正向 | 查看所有存储健康状态 | 返回每个存储的最新状态 |
| 正向 | 等待定时心跳执行 | 健康状态自动更新 |
| 异常 | 存储写入超时 | 标记为 degraded 或 unavailable |
| 异常 | 后台心跳任务 panic | 仅记录日志，不影响主服务 |
| 边界 | 首次检查（无历史记录） | 返回最新检测结果 |
| 边界 | 连续失败 N 次 | consecutive_failures 递增 |

### 必须覆盖的测试类型

- 单元测试：健康检查逻辑、状态判断、延迟记录
- 集成测试：手动触发 → 查看状态 → 等待心跳 → 状态更新
- 容错测试：存储 Provider 模拟故障

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 手动验证后台心跳正常运行

## 复查

### 代码审查关注点

- 健康探测是否使用独立的短超时 context
- 后台心跳 goroutine 的优雅退出（shutdown 时）
- 探测对象是否在检查后清理（不污染存储）
- 并发检测多个存储时的资源控制

### 安全检查

- 健康检查 API 需要管理员权限
- 探测对象不包含敏感信息

### 文档更新要求

- 管理员 API 文档：健康检查接口
- 运维文档说明心跳配置

### 复查通过标准

- [x] 自检通过；待主会话/人工 reviewer 最终批准
- [x] 本地验证命令全部通过；待 CI 运行
- [x] 完成报告已填写

## 完成报告

### 实现摘要

- 新增 `storage_health_checks` 表及索引，迁移通过 `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS` 保持幂等。
- 新增存储健康模型、Repository 与 Service：支持查询全部存储最新状态、手动检测单个存储、手动检测全部存储。
- 健康探测复用现有 Provider 的 `Save` → `Open` → `Delete` 能力，写入 1KB 随机探测对象到 `.omepic-health/` 前缀；读取校验后清理，失败路径也执行清理。
- 新增管理员 API：
  - `GET /admin/storage/health`
  - `POST /admin/storage/:key/health-check`
  - `POST /admin/storage/health-check-all`
- `cmd/server/main.go` 启动后台 5 分钟周期心跳；心跳错误仅记录日志，不阻断服务；Service 暴露可停止的 heartbeat 便于 shutdown 与测试。

### 测试记录

- `cd backend && gofmt -w ./cmd ./internal`：通过。
- `cd backend && go test ./...`：通过（265 passed）。
- `cd backend && gofmt -l ./cmd ./internal`：通过，无输出。

### 覆盖场景

- 迁移幂等与字段存在性。
- 健康状态记录、失败状态记录、连续失败递增、恢复 healthy 后计数清零。
- 手动检测单个存储、查询健康列表、手动检测全部存储 API。
- 后台心跳短周期启动/停止并写入状态。
- 探测对象读后关闭并清理，避免污染存储。

### 复核结论

- 需求、设计、代码与测试一致；本次仅实现“存储健康检查 MVP”，未触碰无关前端文件。
- 当前状态满足 P0 存储健康检查 MVP 验收；后续可在 P1 存储回退/指标任务中扩展 degraded 判定、Prometheus 指标和更精细的心跳配置。
