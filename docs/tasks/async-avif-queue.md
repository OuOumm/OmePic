# 任务：AVIF 异步转换队列

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.2 性能瓶颈与整改建议** — "AVIF 转换在上传请求内同步完成"
- **§1.5.1 AVIF 转换异步队列 + 任务重试**
- **§4.3 P1：AVIF 异步队列** — "image_jobs、Worker、重试、失败重跑、后台队列页"

## 预估人日

8 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/image_job.go`（新增）— 异步任务模型
- `backend/internal/repository/job_repo.go`（新增）— 任务数据访问层
- `backend/internal/service/transform_worker.go`（新增）— 转换 Worker
- `backend/internal/service/image_service.go` — 上传流程改为写入任务队列
- `backend/internal/http/handler/admin_job.go`（新增）— 队列管理 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表迁移
- `backend/cmd/server/main.go` — 启动 Worker Pool
- `frontend/src/routes/admin/dashboard/jobs/+page.svelte`（新增）— 队列管理页面

### 关键实现点

1. 新增 `image_jobs` 表：`id`、`uid`、`source_path`、`target_format`、`status`（queued/converting/failed/completed）、`retry_count`、`last_error`、`created_at`、`updated_at`。
2. 上传接口增加本进程内 Worker Pool 与并发信号量，限制同步转换并发。
3. Worker 支持并发配置：`transform_worker_count`、`transform_timeout_seconds`。
4. 失败策略：指数退避，默认最多 3 次；超过后进入 `failed`。
5. 管理后台支持：查看队列状态、失败任务列表、重试单个/批量任务、调整队列并发。
6. MVP 阶段上传接口仍同步返回，内部只增加并发限制和重试。
7. 后续可升级为：上传返回 `job_id`，前端轮询或 SSE/WebSocket 推送。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./internal/service/... ./internal/repository/...`
- [ ] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [ ] Worker Pool 优雅退出
- [ ] 迁移幂等

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传图片触发异步转换 | 转换完成，AVIF 可访问 |
| 正向 | 并发上传多张图片 | 信号量生效，不超过配置的并发数 |
| 正向 | 转换失败后自动重试 | 按指数退避重试，最多 3 次 |
| 正向 | 管理员手动重试失败任务 | 任务重新进入队列 |
| 异常 | 转换超过 timeout | 标记失败，记录错误信息 |
| 异常 | 重试 3 次仍失败 | 进入 failed 状态 |
| 边界 | Worker Pool 大小为 1 | 严格串行转换 |
| 边界 | 服务重启时队列中有待处理任务 | 重启后继续处理 |

### 必须覆盖的测试类型

- 单元测试：Worker 逻辑、重试策略、指数退避计算
- 集成测试：上传 → 任务创建 → 转换 → 完成
- 压力测试：高并发上传下的队列积压与处理
- 容错测试：Worker panic 恢复

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 高并发场景下 CPU 使用率在可控范围内

## 复查

### 代码审查关注点

- Worker Pool 的 goroutine 生命周期管理
- 指数退避实现是否正确
- 任务状态机是否完整（queued→converting→completed/failed）
- 信号量与运行时设置热更新的兼容性

### 安全检查

- 队列管理 API 需要管理员权限
- 任务错误信息不暴露内部路径

### 文档更新要求

- 管理员 API 文档：队列管理接口
- 运维文档说明 Worker 配置参数
- 数据库 Schema 文档新增表

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
