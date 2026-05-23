# 任务：默认密码安全改造

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.3 安全短板与整改建议** — "默认管理员密码存在部署风险"
- **§4.2 P0：默认密码安全改造** — "首次启动强制初始化或高危限制；后台明显告警"

## 预估人日

1.5 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/service/auth_service.go` — 默认密码状态追踪
- `backend/internal/repository/setting_repo.go` — 读写 `admin_password_uses_default` 设置
- `backend/internal/http/handler/admin_setting.go` — 系统设置返回默认密码状态
- `frontend/src/routes/admin/dashboard/settings/+page.svelte` — 安全区域展示默认密码高危告警

### 关键实现点

1. 默认密码引导保持兼容，但在数据库中记录 `admin_password_uses_default=true`。
2. 修改密码后写入 `admin_password_uses_default=false`。
3. `GET /admin/system-settings` 的 `readonly.security.admin_password.using_default` 反映真实状态。
4. 前端安全警告区域在默认密码状态下展示高危告警（红色醒目提示）。
5. 使用默认密码时限制部分管理能力（如禁止修改存储配置），仅允许修改密码。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./internal/service/...`
- [ ] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [ ] 前端构建通过：`cd frontend && npm run build:backend`

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 首次启动后查看后台设置 | 显示默认密码高危告警 |
| 正向 | 修改密码后查看后台设置 | 告警消失 |
| 异常 | 使用默认密码时修改存储配置 | 拒绝操作，提示需先修改密码 |
| 边界 | 修改密码后又改回默认密码 | 重新标记为使用默认密码 |

### 必须覆盖的测试类型

- 单元测试：默认密码状态读写、权限限制逻辑
- 集成测试：首次启动 → 告警 → 修改密码 → 告警消失
- 前端 E2E：告警显示与隐藏

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 手动验证告警显示与权限限制

## 复查

### 代码审查关注点

- 默认密码检测是否覆盖所有首次启动场景（Docker、二进制）
- 权限限制是否不影响密码修改操作本身
- 告警信息是否足够醒目

### 安全检查

- 默认密码不在日志中明文出现
- 密码修改接口防暴力破解（限流）

### 文档更新要求

- 部署文档说明首次启动密码修改流程

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
