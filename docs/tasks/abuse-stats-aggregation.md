# 任务：滥用统计聚合

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§3.1.3 数据库索引调优** — "滥用统计使用日/小时聚合表"
- **§4.3 P1：滥用统计聚合**

## 预估人日

4 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/abuse_stat.go`（新增）— 聚合统计模型
- `backend/internal/repository/abuse_repo.go`（新增）— 统计数据访问
- `backend/internal/service/abuse_aggregator.go`（新增）— 聚合任务
- `backend/internal/http/handler/admin_abuse.go`（新增）— 统计查询 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增聚合表
- `backend/cmd/server/main.go` — 启动聚合定时任务

### 关键实现点

1. 新增 `abuse_stats_hourly` 表：`hour_bucket`、`token_hash`、`ip_hash`、`upload_count`、`total_bytes`。
2. 新增 `abuse_stats_daily` 表：日级聚合。
3. 定时任务（每小时）从 `images` 表聚合到小时表；每天聚合到日表。
4. 管理员 API：
   - `GET /admin/abuse/top-tokens` — 热门 Token 排行
   - `GET /admin/abuse/top-ips` — 热门 IP 排行
   - `GET /admin/abuse/timeline` — 按时间维度的上传量趋势
5. 大数据量下避免每次全量 GROUP BY，使用预聚合表加速查询。
6. 支持时间范围筛选（最近 24 小时、7 天、30 天）。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传图片后等待聚合 | 聚合表出现对应记录 |
| 正向 | 查询热门 Token | 返回按上传量排序的 Token 列表 |
| 正向 | 查询时间趋势 | 返回按小时/日的上传量数据 |
| 边界 | 无上传数据时查询 | 返回空结果 |
| 边界 | 大量数据聚合 | 聚合任务在合理时间内完成 |

### 必须覆盖的测试类型

- 单元测试：聚合逻辑、时间分桶
- 集成测试：上传→聚合→查询完整链路
- 性能测试：大数据量聚合耗时

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 聚合任务的幂等性（重复执行不产生重复数据）
- 聚合任务的性能（批量 INSERT）
- 查询 API 的响应时间

### 安全检查

- 统计 API 需要管理员权限
- Token hash 和 IP hash 脱敏

### 文档更新要求

- 管理员 API 文档：统计查询接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
