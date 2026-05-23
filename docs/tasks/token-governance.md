# 任务：Token 治理基础

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.3 安全短板与整改建议** — "客户端 Token 只标识上传者，不具备账户级生命周期"
- **§4.2 P0：Token 治理基础** — "Token 使用记录、Token 禁用、按 Token 限流/封禁"

## 预估人日

3 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/token_usage.go`（新增）— `token_usage` 表模型
- `backend/internal/model/token_control.go`（新增）— `token_controls` 表模型
- `backend/internal/repository/token_repo.go`（新增）— Token 治理数据访问层
- `backend/internal/service/token_service.go`（新增）— Token 治理业务逻辑
- `backend/internal/http/handler/admin_token.go`（新增）— 管理员 Token 治理 API
- `backend/internal/http/router/router.go` — 注册新管理路由
- `backend/internal/service/image_service.go` — 上传前检查 Token 禁用状态，上传后记录使用统计
- `backend/internal/database/migration.go` — 新增表迁移

### 关键实现点

1. 新增 `token_usage` 表：按 token hash 聚合上传次数、总大小、最近 IP、最近使用时间、预览值。
2. 新增 `token_controls` 表：记录禁用状态与原因。
3. 上传前检查 Token 是否被禁用；禁用则返回 403 错误。
4. 上传成功后异步记录使用统计到 `token_usage`。
5. 新增管理员 API：
   - `GET /admin/tokens` — 列出所有 Token 及使用统计
   - `POST /admin/tokens/:token/disable` — 禁用 Token
   - `POST /admin/tokens/:token/enable` — 恢复 Token
6. Token hash 使用 SHA-256，不在数据库中存储明文 Token。

### 完成标准

- [x] 代码编译通过：`cd backend && go test ./...`
- [x] 单元测试通过：`cd backend && go test ./...`
- [x] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [x] 数据库迁移幂等：重复执行 `Repository.Migrate` 不报错

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 使用正常 Token 上传图片 | 上传成功，token_usage 记录更新 |
| 正向 | 管理员禁用 Token | token_controls 写入禁用记录 |
| 正向 | 管理员恢复 Token | 禁用记录清除 |
| 异常 | 使用已禁用 Token 上传 | 返回 403，提示 Token 已禁用 |
| 异常 | 使用不存在的 Token 上传 | 按现有逻辑拒绝 |
| 边界 | 同一 Token 并发多次上传 | token_usage 原子更新，不丢失计数 |
| 边界 | Token hash 碰撞 | 以 hash 为准，不影响功能 |

### 必须覆盖的测试类型

- 单元测试：Token 禁用/恢复逻辑、使用统计更新、hash 计算
- 集成测试：上传 → 检查使用统计 → 禁用 → 再次上传被拒 → 恢复 → 上传成功
- 并发测试：多线程上传同一 Token 的统计准确性

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 管理员 API 响应正确

## 复查

### 代码审查关注点

- Token hash 是否使用 SHA-256 且不可逆
- 使用统计更新是否异步且不阻塞上传主链路
- 禁用检查是否在上传链路最早期执行
- 管理员 API 是否需要额外权限校验

### 安全检查

- Token 明文不出现在日志或 API 响应中
- 禁用/恢复操作记录审计日志

### 文档更新要求

- 管理员 API 文档：Token 治理接口
- 数据库 Schema 文档新增表

### 复查通过标准

- [x] 实施代理自查通过
- [x] 本地后端测试全部绿色
- [x] 完成报告已填写

## 完成报告

- 完成时间：2026-05-24
- 实施范围：仅实现 P0「Token 治理基础」，未触碰软删除、审计、健康检查等其他 P0 子任务。
- 代码变更：
  - 新增 `token_usage`、`token_controls` 幂等迁移与索引。
  - 新增 Token SHA-256 hash、预览遮罩、使用统计聚合、禁用/恢复控制。
  - 上传前检查禁用 Token；上传成功后记录使用次数、总字节、最近 IP、最近使用时间。
  - 新增管理员接口：`GET /admin/tokens`、`POST /admin/tokens/:token_hash/disable`、`POST /admin/tokens/:token_hash/enable`，路由参数仅接受 token hash。
- 安全结论：数据库、API 与日志路径均不新增明文 Token 暴露；API 响应只返回 `token_hash` 与 `token_preview`。
- 验证命令：
  - `cd backend && gofmt -w ./cmd ./internal`：通过
  - `cd backend && go test ./...`：通过（17 packages，252 tests）
  - `cd backend && gofmt -l ./cmd ./internal`：通过（无输出）
- 复核结论：实现、测试与本任务需求一致；未修改 `frontend/src/lib/i18n.ts`、`frontend/src/routes/+error.svelte`。
