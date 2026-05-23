# 任务：短链服务

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F3** — "短链服务"
- **§2.1 F3 短链服务**
- **§4.4 P2：短链服务** — "slug、过期、访问统计、禁用、管理后台"

## 预估人日

5 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/short_link.go`（新增）— 短链模型
- `backend/internal/repository/short_link_repo.go`（新增）— 短链数据访问
- `backend/internal/service/short_link_service.go`（新增）— 短链业务逻辑
- `backend/internal/http/handler/short_link.go`（新增）— 短链访问与管理 API
- `backend/internal/http/router/router.go` — 注册短链路由
- `backend/internal/database/migration.go` — 新增表迁移

### 关键实现点

1. 短链记录字段：`slug`、`target_uid`、`expires_at`、`disabled_at`、`visit_count`、`created_at`。
2. 访问短链 `/s/:slug` → 302 重定向到图片 URL。
3. 可选能力：一次性链接、密码访问、Referer 白名单。
4. 风险控制：slug 枚举防护（不存在的 slug 返回 404，不泄露是否存在）、访问限流、后台禁用。
5. 管理员 API：创建短链、列表、禁用、删除。
6. slug 自动生成（6-8 位 base62），支持自定义。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] 短链访问正确重定向

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 创建短链并访问 | 302 重定向到图片 |
| 正向 | 访问已禁用短链 | 返回 410 Gone |
| 正向 | 访问已过期短链 | 返回 410 Gone |
| 异常 | 访问不存在的 slug | 返回 404 |
| 边界 | 自定义 slug 冲突 | 返回 409 Conflict |
| 边界 | 并发访问同一短链 | visit_count 准确递增 |

### 必须覆盖的测试类型

- 单元测试：slug 生成、过期判断、禁用逻辑
- 集成测试：创建→访问→统计→禁用

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- slug 不可预测性（随机性）
- 不存在的 slug 不泄露信息
- visit_count 原子递增

### 安全检查

- 短链管理 API 需要管理员权限
- 防止 slug 枚举攻击

### 文档更新要求

- API 文档说明短链接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
