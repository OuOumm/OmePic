# 任务：Prometheus 指标

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§3.3 可观测性优化 §3.3.1 Prometheus 指标**

## 预估人日

4 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/metrics/collector.go`（新增）— 指标注册与采集
- `backend/internal/metrics/middleware.go`（新增）— HTTP 请求指标中间件
- `backend/internal/http/router/router.go` — 注册 `/metrics` 端点
- `backend/internal/service/image_service.go` — 上传/转换/去重指标
- `backend/internal/storage/provider.go` — 存储请求指标
- `backend/cmd/server/main.go` — 指标初始化

### 关键实现点

1. 暴露 `/metrics` 端点（Prometheus exposition format）。
2. 核心指标：
   - `omepic_upload_total`（Counter）— 上传总数，按结果、存储、MIME 分类
   - `omepic_upload_bytes_total`（Counter）— 上传原始字节数
   - `omepic_transform_duration_seconds`（Histogram）— 图片转换耗时
   - `omepic_storage_request_duration_seconds`（Histogram）— 存储请求耗时
   - `omepic_storage_errors_total`（Counter）— 存储错误数
   - `omepic_dedup_hits_total`（Counter）— 去重命中次数
   - `omepic_cache_hit_total`（Counter）— 缓存命中
   - `omepic_rate_limited_total`（Counter）— 限流次数
   - `omepic_ip_banned_total`（Counter）— IP 封禁命中
   - `omepic_job_queue_depth`（Gauge）— 队列积压
3. 提供 Grafana 面板样例 JSON。
4. 指标采集对请求延迟影响最小化。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确
- [ ] `/metrics` 端点可访问

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 访问 `/metrics` | 返回 Prometheus 格式指标 |
| 正向 | 上传图片后检查指标 | `omepic_upload_total` 递增 |
| 正向 | 去重命中后检查指标 | `omepic_dedup_hits_total` 递增 |
| 异常 | 存储错误后检查指标 | `omepic_storage_errors_total` 递增 |
| 边界 | 高并发上传 | 指标采集不成为瓶颈 |

### 必须覆盖的测试类型

- 单元测试：指标注册、计数器递增、Histogram 记录
- 集成测试：端到端操作→指标验证

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 指标命名是否符合 Prometheus 命名规范
- 标签基数是否可控（避免高基数标签）
- 指标采集对性能的影响

### 安全检查

- `/metrics` 端点是否需要访问控制

### 文档更新要求

- 运维文档说明指标采集与 Grafana 配置

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
