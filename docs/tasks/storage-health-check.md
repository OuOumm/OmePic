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

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./internal/service/... ./internal/repository/...`
- [ ] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [ ] 迁移幂等

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

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
