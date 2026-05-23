# 任务：请求追踪与结构化日志

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§3.3.2 请求追踪**

## 预估人日

3 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/middleware/request_id.go`（新增）— 请求 ID 中间件
- `backend/internal/middleware/structured_logger.go`（新增）— 结构化日志中间件
- `backend/internal/http/router/router.go` — 注册中间件
- `backend/internal/service/image_service.go` — 日志字段注入
- `backend/internal/http/handler/*.go` — 错误响应携带 request_id

### 关键实现点

1. 为每个请求生成 `request_id`（UUID），响应头返回 `X-Request-ID`。
2. 结构化日志包含：`request_id`、`uid`、`storage_key`、`token_hash`、`ip_hash`、耗时、错误码。
3. 后台任务继承 trace context：上传请求和转换任务可串联排查。
4. 错误响应体增加 `request_id` 字段，方便用户反馈。
5. 可选接入 OpenTelemetry，输出到 Jaeger/Tempo（后期扩展）。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确
- [ ] 所有响应包含 `X-Request-ID` 头

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 发送请求 | 响应包含 `X-Request-ID` 头 |
| 正向 | 请求失败 | 错误响应体包含 `request_id` |
| 正向 | 查看日志 | 包含 request_id 和相关上下文 |
| 边界 | 客户端提供 X-Request-ID | 使用客户端提供的值或生成新值 |

### 必须覆盖的测试类型

- 单元测试：request_id 生成与传播
- 集成测试：请求→日志→错误响应验证

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- request_id 格式是否统一（UUID v4）
- 日志字段是否脱敏（token_hash, ip_hash）
- 结构化日志库选择是否合理

### 安全检查

- 日志不包含明文 Token 或 IP

### 文档更新要求

- API 文档说明 `X-Request-ID` 头
- 运维文档说明日志格式与查询方式

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
