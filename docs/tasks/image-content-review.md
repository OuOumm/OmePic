# 任务：图片审核

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F4** — "图片鉴黄/内容审核"
- **§4.3 P1：图片审核** — "审核状态、第三方 API 接入、后台人工复核、拒绝访问策略"

## 预估人日

8 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/image.go` — 新增审核状态字段
- `backend/internal/service/review_service.go`（新增）— 审核服务
- `backend/internal/http/handler/admin_review.go`（新增）— 审核管理 API
- `backend/internal/http/handler/image.go` — 公开访问审核状态检查
- `backend/internal/http/router/router.go` — 注册新路由
- `backend/internal/database/migration.go` — 迁移脚本
- `frontend/src/routes/admin/dashboard/reviews/+page.svelte`（新增）— 审核管理页面

### 关键实现点

1. 状态流：`pending_review` → `approved` / `rejected` / `manual_review`。
2. 上传后异步调用内容安全服务或本地 NSFW 模型。
3. 处理模式：
   - 公开图床：审核通过前不公开访问（返回 403）。
   - 私有图床：先公开但后台标记风险。
4. 管理后台人工复核：查看待审核列表、批准/拒绝、批量操作。
5. 接入选项：可配置第三方内容安全 API 或本地模型。
6. 拒绝的图片保留记录但公开访问返回 404。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传正常图片 | 自动审核通过，公开可访问 |
| 正向 | 管理员手动批准图片 | 公开可访问 |
| 正向 | 管理员手动拒绝图片 | 公开访问返回 404 |
| 异常 | 上传违规图片 | 自动拒绝或标记待人工审核 |
| 异常 | 公开访问待审核图片 | 返回 403（公开模式） |
| 边界 | 审核服务不可用 | 标记为 manual_review，不阻塞上传 |
| 边界 | 批量审核 | 批量操作正确执行 |

### 必须覆盖的测试类型

- 单元测试：审核状态机、状态转换合法性
- 集成测试：上传→审核→访问控制完整链路
- 容错测试：审核服务故障时不阻塞上传

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 审核服务调用是否异步且不阻塞上传
- 状态机转换是否严密（不允许非法状态跳转）
- 审核服务故障时的降级策略

### 安全检查

- 审核管理 API 需要管理员权限
- 审核结果不泄露给非管理员用户

### 文档更新要求

- 管理员 API 文档：审核管理接口
- 运维文档说明审核服务接入配置

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
