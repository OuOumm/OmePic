# 任务：批量打包下载

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F5** — "批量打包下载"
- **§4.3 P1：批量打包下载** — "异步导出任务、ZIP 生成、临时下载链接、清理任务"

## 预估人日

5 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/export_job.go`（新增）— 导出任务模型
- `backend/internal/repository/export_repo.go`（新增）— 导出任务数据访问
- `backend/internal/service/export_service.go`（新增）— 批量导出服务
- `backend/internal/http/handler/admin_export.go`（新增）— 导出 API
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 迁移脚本
- `frontend/src/routes/admin/dashboard/images/+page.svelte` — 批量操作 UI

### 关键实现点

1. 支持选择来源：后台勾选、按 Token、按 IP、按日期、按存储实例。
2. 新增 `export_jobs` 表：`id`、`filter_criteria`（JSON）、`status`、`file_path`、`file_size`、`expires_at`、`created_at`。
3. 大批量场景使用异步导出任务：创建任务 → 后台生成 ZIP → 完成后提供临时下载链接。
4. ZIP 内包含：图片文件、元数据 JSON、Markdown 清单。
5. 临时下载链接有效期（默认 24 小时），过期自动清理。
6. 清理任务定期删除过期的导出文件。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 勾选 10 张图片导出 | 生成 ZIP，下载成功 |
| 正向 | 按 Token 筛选导出 | 包含该 Token 的所有图片 |
| 正向 | 按日期范围导出 | 包含日期范围内的图片 |
| 异常 | 导出 0 张图片 | 返回错误提示 |
| 异常 | 下载过期链接 | 返回 410 Gone |
| 边界 | 导出 1000 张图片 | 异步任务，进度可查 |
| 边界 | 磁盘空间不足 | 任务失败，记录错误 |

### 必须覆盖的测试类型

- 单元测试：ZIP 生成、筛选逻辑、链接过期
- 集成测试：创建导出任务→等待完成→下载→验证内容
- 清理测试：过期文件自动删除

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 大批量导出的内存使用（流式写入 ZIP）
- 临时文件清理机制
- 并发导出任务的资源控制

### 安全检查

- 导出下载链接不可预测
- 导出 API 需要管理员权限

### 文档更新要求

- 管理员 API 文档：导出接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
