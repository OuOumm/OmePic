# 任务：图片访问统计

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F8** — "图片访问统计"
- **§4.4 P2：图片访问统计** — "PV/UV/Referer/流量统计、热门排行、导出"

## 预估人日

6 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/access_log.go`（新增）— 访问日志模型
- `backend/internal/repository/access_repo.go`（新增）— 访问数据访问
- `backend/internal/service/access_aggregator.go`（新增）— 聚合统计任务
- `backend/internal/http/handler/admin_stats.go`（新增）— 统计查询 API
- `backend/internal/http/handler/image.go` — 记录访问日志
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表

### 关键实现点

1. 访问日志表：`image_uid`、`ip_hash`、`referer`、`user_agent`、`bytes_served`、`accessed_at`。
2. 异步记录访问日志，不阻塞图片读取。
3. 定时聚合到 `access_stats_daily` 表：PV、UV、流量、热门 Referer。
4. 管理员 API：
   - `GET /admin/stats/overview` — 总体统计
   - `GET /admin/stats/top-images` — 热门图片排行
   - `GET /admin/stats/top-referers` — 热门来源
   - `GET /admin/stats/timeline` — 时间趋势
5. 支持导出统计数据（CSV/JSON）。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] 访问统计正确记录和展示

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 访问图片 | 访问日志被记录 |
| 正向 | 查看热门图片排行 | 返回正确排序 |
| 正向 | 查看时间趋势 | 返回正确的 PV/UV 数据 |
| 边界 | 无访问数据 | 返回空结果 |
| 边界 | 高并发访问 | 异步记录不阻塞读取 |

### 必须覆盖的测试类型

- 单元测试：日志记录、聚合逻辑
- 集成测试：访问→日志→聚合→查询

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 访问日志异步写入的可靠性
- 聚合任务性能
- 日志表增长控制策略

### 安全检查

- 统计 API 需要管理员权限
- IP hash 脱敏

### 文档更新要求

- 管理员 API 文档说明统计接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
