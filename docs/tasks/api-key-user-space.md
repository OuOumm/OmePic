# 任务：API Key / 用户空间

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F10** — "API Key / 用户空间"
- **§4.4 P2：API Key / 用户空间** — "多用户、额度、权限、命名 Key、后台管理"

## 预估人日

12 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/api_key.go`（新增）— API Key 模型
- `backend/internal/model/user_space.go`（新增）— 用户空间模型
- `backend/internal/repository/api_key_repo.go`（新增）— API Key 数据访问
- `backend/internal/service/api_key_service.go`（新增）— API Key 业务逻辑
- `backend/internal/http/handler/admin_api_key.go`（新增）— 管理员 API Key 管理
- `backend/internal/http/handler/upload.go` — 支持 API Key 认证
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 新增表

### 关键实现点

1. API Key 模型：`id`、`name`、`key_hash`、`space_id`、`permissions`（JSON）、`quota_daily`、`quota_used`、`expires_at`、`created_at`。
2. 用户空间模型：`id`、`name`、`total_quota`、`used_quota`、`created_at`。
3. 上传支持 API Key 认证：`Authorization: Bearer <api_key>`。
4. 额度控制：每日/总量上传限制，超限返回 429。
5. 权限模型：`upload`、`delete`、`read`。
6. 管理员 API：创建/列表/禁用/删除 API Key；管理用户空间。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] API Key 认证可用

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 使用有效 API Key 上传 | 上传成功，额度扣减 |
| 正向 | 管理员创建 API Key | 返回 Key 明文（仅一次） |
| 异常 | 使用过期 API Key | 返回 401 |
| 异常 | 额度用尽后上传 | 返回 429 |
| 异常 | 权限不足的操作 | 返回 403 |
| 边界 | API Key 碰撞 | 以 key_hash 为准 |

### 必须覆盖的测试类型

- 单元测试：Key 生成、认证、额度、权限
- 集成测试：创建 Key→使用 Key→额度耗尽

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- API Key 不在数据库中存储明文
- 额度计算的线程安全
- 权限校验覆盖所有操作

### 安全检查

- Key 明文仅在创建时返回一次
- Key hash 使用安全哈希算法

### 文档更新要求

- API 文档说明 API Key 认证方式
- 管理员 API 文档说明 Key 管理接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
