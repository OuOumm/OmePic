# OmePic 项目优化与整改方案

> 审查时间：2026-05-28
> 项目阶段：**开发版本**
> 核心设计：**前端匿名生成 Token 上传 + 单密码管理员后台**

**两条用户路径**：

| 角色 | 认证方式 | 关键能力 |
|------|----------|----------|
| 匿名用户 | 前端自生成 X-Token（localStorage），**后端不签发用户 token** | 上传、凭 Token 删除自己的图片 |
| 管理员 | 单密码登录 → JWT | 图片管理、存储配置、IP 封禁、系统设置 |

**核心约束**：

- UID ≤ 30 字符、可逆解码、支持前缀检查；当前 Snowflake + XOR + Base62 方案满足要求，不替换为随机 ID。
- 后端不生成、不签发客户端 Token；X-Token 由前端 Web Crypto API 生成，后端仅存储和校验。

---

## 1. 高危安全整改

### H-04 CSP 允许 unsafe-inline + JWT TTL 过长

- **风险**：`unsafe-inline` 降低 CSP 对 XSS 的阻断能力。管理 JWT 存放在 localStorage，XSS 可直接读取。
- **定位**：`backend/internal/http/router/frontend.go:17` — CSP `unsafe-inline`
- **整改**：与 M-08 合并实施（同一个 PR）
  1. 移除 `'unsafe-inline'`，初始主题脚本移入静态 JS 或使用 nonce。
  2. 缩短管理员 JWT TTL（如 2-4 小时），降低泄漏窗口。
  3. 管理 token 保持 localStorage + Bearer token 模式，不迁移到 HttpOnly cookie（单管理员场景改动成本不匹配收益）。

---

## 2. 中低危安全与逻辑缺陷

### M-01 默认 CORS 允许任意 Origin

- **风险**：当前 CORS 对所有路由统一开放，管理 API 不应默认接受跨域请求。
- **定位**：`backend/internal/http/router/router.go:104-118`
- **整改**：
  1. 拆分 CORS 策略：公开 API（上传、删除、查看存储列表）支持跨域；管理 API 严格限制同源。
  2. CORS 允许的 Origin 列表通过后台运行时设置配置，支持热更新，无需重启服务。
  3. 开发模式可默认宽松，生产模式默认仅允许同源。

### M-04 存储与第三方凭据明文保存在 SQLite

- **定位**：`migration.go:25-37`；`storage_repository.go:20,188-205`
- **现状**：当前所有 SQL 均使用参数化查询，无已知注入点。此整改为纵深防御——即使未来出现 SQL 注入，凭据也不会明文暴露。
- **整改**：新增 `SECRET_ENCRYPTION_KEY`，使用 AES-GCM 信封加密。开发版直接纯密文写入，无需明文兼容读取逻辑。

### M-05 Cloudflare API Base URL 缺少配置提示

- **现状**：Cloudflare API Base URL 由单管理员在后台配置，不属于匿名用户输入；强制 allowlist 对当前单管理员模型收益有限。
- **定位**：`runtime_settings.go:270-273`
- **整改**：不做代码层面的强制锁定，仅在后台配置页提示：请确认该地址为 Cloudflare 官方 API 或可信自建代理。

### M-06 管理员 JWT 缺少会话撤销

- **风险**：修改密码后旧 token 在有效期内仍可用。
- **定位**：`backend/internal/auth/jwt.go:15-24`
- **现状**：JWT 当前为纯无状态校验，无 Redis 缓存依赖。
- **整改**：在 Redis 存储一个 `admin_revoked_before` 时间戳；修改密码时写入当前时间；JWT 验证时比对 claims 中的 `iat`（签发时间），早于撤销时间的 token 视为失效。不引入 SQLite session_version 或额外存储依赖。

### M-08 CSP 仅应用于前端兜底路由，安全头不统一

- **风险**：API 和图片响应缺少统一安全头。
- **定位**：`router/frontend.go:35-39,68-75`
- **整改**：与 H-04 合并实施（同一个 PR）。新增全局 `SecurityHeaders` middleware，根据响应类型分支设置：
  - 所有响应统一：`X-Content-Type-Options: nosniff`、`Referrer-Policy: strict-origin-when-cross-origin`
  - 前端 HTML 页面额外：CSP（移除 `unsafe-inline`）、`X-Frame-Options: DENY`
  - API JSON 响应（`/v1/*`）额外：`Cache-Control: no-store`
  - 图片响应（`/i/:uid`）额外：保持现有 `Cache-Control` 长缓存策略不变
  - 完成后删除旧的前端安全头设置代码

---

## 3. 代码质量与架构优化

### Q-01 缺少 CI 流水线

- **整改**：Go `vet` + `test` + `govulncheck`；Frontend `lint` + `typecheck` + `test` + `build:backend`；依赖扫描。

### Q-04 UID 混淆密钥命名与文档不准确

- **现状**：UID 使用 Snowflake + XOR + Base62，满足 ≤ 30 字符、可逆解码、前缀检查的设计要求。XOR 是混淆而非加密，但环境变量和代码注释中使用了"encryption"字样。
- **整改**：
  1. 环境变量 `UID_ENCRYPTION_KEY` 保留（避免破坏现有部署），文档和代码注释统一标注为"混淆密钥（obfuscation key），非加密安全边界"。
  2. 新增上下文说明：UID 的目的是避免暴露可预测的序列 ID，不提供密码学安全保证。

### Q-06 测试覆盖加强到安全边界

- **整改**：multipart 超大 body、默认密钥阻断、限流降级、X-Token 生成（无 Web Crypto 场景）、UID 前缀解码等补充表驱动测试。

---

## 4. 实施路线图

### 阶段 0：基础设施（1-2 天）

- [x] CI 流水线（Q-01）
- [x] Dockerfile + docker-compose（附带 `.env.production.example`）

### 阶段 2：管理员认证与浏览器安全（2-3 天）

- [x] CORS 拆分：公开 API 跨域 + 管理 API 同源 + 后台热更新（M-01）
- [x] JWT Redis 撤销机制（M-06）
- [x] 全局安全头 middleware + 移除 unsafe-inline + 缩短 JWT TTL（H-04 + M-08，同 PR）

### 阶段 3：部署与配置安全（2-3 天）

- [x] SQLite 凭据加密（M-04）
- [x] Cloudflare API Base URL 配置提示（M-05，低优先）

### 阶段 4：质量收尾（3-5 天）

- [x] UID 混淆密钥命名修正（Q-04）
- [x] 安全边界测试补充（Q-06）

---

## 5. 开发原则

1. **守住核心设计** — 匿名上传 + 单管理员，后端不签发用户 token。限流按 IP 维度实现，不引入账户体系。
2. **UID 约束不可破坏** — ≤ 30 字符、可逆解码、前缀可检查。Snowflake + XOR + Base62 方案保持不变，仅修正命名。
3. **安全分层** — 匿名侧：body 限制、限流、X-Token 客户端安全。管理侧：强密码、短 TTL、CORS 收窄。
4. **直接实施** — 开发版本无需迁移兼容层。
5. **测试伴随 + 文档同步** — 每项整改同步提交单测；API 字段、前端类型、README 随代码更新。