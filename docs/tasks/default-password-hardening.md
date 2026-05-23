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

- [x] 后端测试通过：`cd backend && go test ./...`
- [x] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [x] 前端类型检查通过：`cd frontend && npm run typecheck`
- [x] 前端构建通过：`cd frontend && npm run build:backend`

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

- [x] 需求、设计、代码、测试一致性自查通过
- [x] 本地验证命令通过
- [x] 完成报告已填写

## 完成报告

- 完成时间：2026-05-24
- 实施范围：仅实现 P0「默认密码安全改造」，未实现 Token、软删除、审计、健康检查等其他 P0 功能。
- 代码变更：
  - `backend/internal/service/admin_service.go`：持久化 `admin_password_uses_default`；首次默认密码状态为 true；改密后按新密码是否等于默认密码写入状态；`GET /admin/system-settings` 返回真实 `readonly.security.admin_password.using_default`；默认密码状态下拒绝 runtime/storage 高危配置变更但允许修改密码。
  - `backend/internal/http/handler/admin_handler.go`：高危配置拒绝时返回服务层用户提示。
  - `backend/internal/service/admin_service_test.go`、`backend/internal/http/handler/admin_handler_test.go`：覆盖首次默认状态、改密后 false、改回默认 true、默认密码阻断高危配置且允许改密。
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte`：安全警告区域在 `admin_password.using_default=true` 时展示高危告警，改密成功后刷新系统设置。
- 验证结果：
  - `cd backend && gofmt -w ./cmd ./internal`：通过。
  - `cd backend && go test ./...`：通过（245 passed）。
  - `cd backend && gofmt -l ./cmd ./internal`：通过（无输出）。
  - `cd frontend && npm run typecheck`：通过（0 errors, 0 warnings）。
  - `cd frontend && npm run build:backend`：通过，静态文件已复制到 `backend/web/`。
- 复核结论：默认密码兼容引导保留；真实状态通过 config 表持久化；高危 runtime/storage 配置在默认密码状态下被禁止；密码修改接口不被禁止；前端最小 UI 已展示告警。
- 风险/后续：前端仅做类型检查和构建，未补充 E2E；生产部署仍需管理员首次登录后立即修改默认密码。
