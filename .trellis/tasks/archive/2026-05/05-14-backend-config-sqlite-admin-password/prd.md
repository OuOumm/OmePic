# Move mutable config to sqlite

## Goal

将可运行时修改的配置从 `.env` 收敛到 SQLite，减少部署环境变量数量，并让管理员密码支持通过管理端修改且不以明文保存。

## Confirmed Facts

- SQLite `config` 表已经存在，`RuntimeSettingsManager` 会从该表读取站点/上传策略/`public_base_url` 等运行时配置。
- 存储配置已迁移到 `storage_configs`，启动时通过 `InitializeStorageCatalog` 兜底初始化默认本地存储。
- 管理员密码当前已经存入 SQLite `config.admin_password_hash`，使用 bcrypt 哈希；首次登录若不存在会写入默认 `admin123` 哈希。
- `AdminService.ChangePassword` 已存在，但 HTTP 路由/前端 API 尚未暴露修改密码入口。
- 当前 `config.Load()` 仍从环境变量读取 `UID_PREFIX`、`UID_ENCRYPTION_KEY`、`JWT_SECRET`、`TRUSTED_PROXY_CIDRS`、`REAL_IP_HEADER` 等。
- 用户明确要求当前环境变量只保留：`HTTP_ADDR`、`DATABASE_PATH`、`REDIS_URL`、`UID_PREFIX`、`UID_ENCRYPTION_KEY`、`JWT_SECRET`。
- 用户确认不保留 `TRUSTED_PROXY_CIDRS` / `REAL_IP_HEADER` 环境变量。
- 用户确认管理员修改密码接口使用 `PUT /admin/password`。
- 用户确认首次启动仍允许默认管理员密码 `admin123` 写入 SQLite bcrypt 哈希。
- 用户要求同步提供前端“修改密码”入口。
- 用户要求程序首次运行时写入默认运行时配置；当前默认配置只在内存兜底，没有持久化到 SQLite。
- 用户要求删除修改密码表单中的说明/更新提示文案。
- 用户要求新增密码强度规则：新密码至少 8 位，且包含大小写字母和符号。
- 用户要求修改密码旧密码错误时返回/展示明确的密码错误提示，而不是泛化为 `forbidden`。

## Requirements

- `.env.example`、README/相关文档不得再要求存储配置、管理员密码、`PUBLIC_BASE_URL` 等已进入 SQLite 的环境变量。
- 后端 `config.Load()` 只从环境变量读取用户指定的保留项：
  - `HTTP_ADDR=:8080`
  - `DATABASE_PATH=data/omepic.db`
  - `REDIS_URL=redis://localhost:6379/0`
  - `UID_PREFIX=omeo_`
  - `UID_ENCRYPTION_KEY=change-me-uid-secret`
  - `JWT_SECRET=change-me-too`
- `PUBLIC_BASE_URL` 只通过 SQLite runtime settings 管理；管理端只展示 runtime/request_host 来源，不再展示 env public base url 状态。
- 管理员密码必须保存在 SQLite 中，使用不可逆安全哈希（当前 bcrypt 可继续使用），不得保存明文。
- 管理员必须能通过已认证的管理 API `PUT /admin/password` 修改密码；修改时需要校验旧密码，并校验新密码非空及强度。
- 新密码强度要求：至少 8 位，包含大写字母、小写字母和符号。
- 修改密码失败应使用现有 JSON 错误响应约定；不得泄露密码或哈希；旧密码错误应返回明确的密码错误消息而不是只显示 `forbidden`。
- 前端管理设置页需要提供修改密码入口，调用已认证的 `PUT /admin/password` API；不展示额外说明/更新提示文案。
- 程序首次运行时必须将默认 runtime settings 写入 SQLite `config` 表，至少覆盖站点信息、上传策略、`public_base_url`、维护模式、限流等现有 runtime settings 键；后续已有值不得被默认值覆盖。
- 更新或新增后端测试覆盖配置环境变量收敛、首次启动默认 runtime config 持久化、管理员密码修改 API/服务行为。

## Acceptance Criteria

- [ ] `.env.example` 只包含用户指定的 6 个环境变量及必要注释。
- [ ] `backend/internal/config` 不再读取 `TRUSTED_PROXY_CIDRS`、`REAL_IP_HEADER`、`PUBLIC_BASE_URL`、存储配置或管理员密码相关环境变量。
- [ ] `AdminEnvironmentStatus` 不再暴露 `env_public_base_url_set` 等已废弃 env 状态。
- [ ] 存储配置与 `public_base_url` 继续从 SQLite 读取并可由现有管理端设置保存。
- [ ] 提供认证后的管理员修改密码 API `PUT /admin/password`，请求包含 `old_password` 与 `new_password`。
- [ ] 前端管理设置页能输入旧密码/新密码并完成密码修改；不会在日志、状态或 UI 中展示哈希或额外说明提示。
- [ ] 新密码不足 8 位、缺少大写/小写/符号时修改失败并返回 `invalid_input`。
- [ ] 旧密码错误时返回明确的密码错误提示，不在前端只显示 `forbidden`。
- [ ] 修改密码后旧密码登录失败，新密码登录成功；密码在 SQLite 中为 bcrypt 哈希而非明文。
- [ ] 首次运行/空库迁移后，SQLite `config` 表存在默认 runtime settings；已有配置再次启动不被覆盖。
- [ ] `go test ./...` 在 `backend/` 通过；如改动前端，则运行可行的前端类型检查。

## Out of Scope

- 不重做完整前端设置页 UI，只新增必要的修改密码入口。
- 不引入独立迁移工具；继续使用现有幂等启动迁移。
