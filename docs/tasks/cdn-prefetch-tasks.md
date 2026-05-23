# 任务：CDN 预热任务

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§3.1.2 图片 CDN 预热**

## 预估人日

4 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/prefetch_job.go`（新增）— 预热任务模型
- `backend/internal/repository/prefetch_repo.go`（新增）— 预热数据访问
- `backend/internal/service/prefetch_service.go`（新增）— CDN 预热服务
- `backend/internal/http/handler/admin_prefetch.go`（新增）— 预热管理 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表

### 关键实现点

1. 上传成功后可异步触发 CDN 预热：`/i/:uid.avif`、`/i/:uid.webp`。
2. 对批量迁移、格式重建、热门图片支持批量预热。
3. 管理后台展示预热状态：未预热、预热中、成功、失败。
4. Cloudflare 场景复用现有 purge 配置扩展出 prefetch 任务。
5. 预热失败自动重试（指数退避，最多 3 次）。
6. 管理员 API：触发预热、查看状态、重试失败任务。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] CDN 预热可触发和查看状态

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传后自动预热 | 预热任务创建并执行 |
| 正向 | 批量预热 | 批量任务创建 |
| 正向 | 查看预热状态 | 显示正确状态 |
| 异常 | CDN API 不可用 | 标记失败，可重试 |
| 边界 | 预热失败后重试 | 按指数退避重试 |

### 必须覆盖的测试类型

- 单元测试：预热逻辑、重试策略
- 集成测试：上传→预热→状态查看

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 预热任务不阻塞上传主链路
- Cloudflare API 调用的限流处理
- 重试策略的合理性

### 安全检查

- CDN API 凭据安全存储
- 预热 API 需要管理员权限

### 文档更新要求

- 管理员 API 文档说明预热接口
- 运维文档说明 CDN 预热配置

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
