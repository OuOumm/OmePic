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

### H-01 默认管理员密码与默认签名密钥未强制要求设置

- **风险**：当前接受默认密码 `admin123` 和固定默认密钥，攻击者可接管管理后台或伪造 JWT。
- **定位**：
  - `backend/internal/service/admin_service.go:118` — 默认密码
  - `backend/internal/config/config.go:68-69` — 默认密钥
  - `backend/cmd/server/main.go:25-29,93-94` — 仅打印告警
- **整改**：
  1. 启动时强制从 `.env` 读取 `ADMIN_PASSWORD`、`JWT_SECRET`、`UID_ENCRYPTION_KEY`。
  2. 任一未设置则直接启动失败，不再提供默认值。
  3. 删除代码中的 `DefaultAdminPassword = "admin123"` 和默认密钥常量。
  4. 不引入 setup token 或安装向导流程。

```go
func enforceRequiredSecrets(cfg config.AppConfig) error {
    if cfg.AdminPassword == "" {
        return errors.New("ADMIN_PASSWORD must be set in .env")
    }
    if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
        return errors.New("JWT_SECRET must be at least 32 bytes")
    }
    if cfg.UIDEncryptionKey == "" || len(cfg.UIDEncryptionKey) < 32 {
        return errors.New("UID_ENCRYPTION_KEY must be at least 32 bytes")
    }
    return nil
}
```

### H-02 HTTP 服务缺少超时和请求头限制

- **风险**：`gin.Engine.Run` 默认超时全为 0，公网可被 Slowloris、慢上传拖垮。
- **定位**：`backend/cmd/server/main.go:130`
- **整改**：替换为显式 `http.Server`，同时增加优雅退出。

```go
server := &http.Server{
    Addr:              cfg.HTTPAddr,
    Handler:           engine,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       60 * time.Second,
    WriteTimeout:      300 * time.Second,
    IdleTimeout:       60 * time.Second,
    MaxHeaderBytes:    1 << 20,
}
if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
    return err
}
```

### H-03 上传入口在 multipart 解析前缺少全局 body 限制

- **风险**：`FormFile` 内部可能已将超大文件写入临时文件，业务层 `LimitReader` 为时已晚。
- **定位**：
  - `backend/internal/http/handler/image_handler.go:32-33` — 先 `FormFile`
  - `backend/internal/service/image_service.go:413,427` — 服务层 LimitReader
- **整改**：路由层添加 `MaxBytesReader` 中间件，body 限制 = `MaxUploadSizeBytes + 1MiB`。

```go
func BodyLimit(settings *service.RuntimeSettingsManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        max := settings.Current().MaxUploadSizeBytes()
        if max > 0 {
            c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max+(1<<20))
        }
        c.Next()
    }
}
```

### H-04 CSP 允许 unsafe-inline + JWT TTL 过长

- **风险**：`unsafe-inline` 降低 CSP 对 XSS 的阻断能力。管理 JWT 存放在 localStorage，XSS 可直接读取。
- **定位**：`backend/internal/http/router/frontend.go:17` — CSP `unsafe-inline`
- **整改**：与 M-08 合并实施（同一个 PR）
  1. 移除 `'unsafe-inline'`，初始主题脚本移入静态 JS 或使用 nonce。
  2. 缩短管理员 JWT TTL（如 2-4 小时），降低泄漏窗口。
  3. 管理 token 保持 localStorage + Bearer token 模式，不迁移到 HttpOnly cookie（单管理员场景改动成本不匹配收益）。

### H-05 前端 X-Token 允许 Math.random 降级生成

- **风险**：X-Token 是匿名模式下用户删除图片的唯一授权凭据，前端 `generateToken()` 允许 `Math.random` 降级。弱随机数可被预测，攻击者可伪造 token 删除他人图片。
- **定位**：`frontend/src/lib/client-token.ts:1-26`
- **整改**：
  1. 移除 `Math.random` 分支；没有 `crypto.randomUUID` / `crypto.getRandomValues` 时直接报错，提示浏览器不受支持。
  2. 统一使用 `crypto.randomUUID()` 作为首选，`getRandomValues` 作为 fallback。

```ts
function generateToken(): string {
  const crypto = globalThis.crypto;
  if (!crypto) {
    throw new Error('Web Crypto API is required. Please use a modern browser.');
  }
  if (crypto.randomUUID) return crypto.randomUUID();
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}
```

---

## 2. 中低危安全与逻辑缺陷

### M-01 默认 CORS 允许任意 Origin

- **风险**：当前 CORS 对所有路由统一开放，管理 API 不应默认接受跨域请求。
- **定位**：`backend/internal/http/router/router.go:104-118`
- **整改**：
  1. 拆分 CORS 策略：公开 API（上传、删除、查看存储列表）支持跨域；管理 API 严格限制同源。
  2. CORS 允许的 Origin 列表通过后台运行时设置配置，支持热更新，无需重启服务。
  3. 开发模式可默认宽松，生产模式默认仅允许同源。

### M-02 代理 IP 信任配置不可配置

- **风险**：反向代理后所有用户共享代理 IP，限流/IP 封禁全部失效或误伤。
- **定位**：`backend/cmd/server/main.go:109`；`clientip/resolver.go:32-40`
- **整改**：在后台运行时设置中新增"真实 IP 来源"配置项（如 X-Forwarded-For、X-Real-IP、CF-Connecting-IP、直连 RemoteAddr），管理员在设置页选择，无需修改 `.env` 或重启服务。

### M-03 上传 URL 基于 Host 头推导

- **风险**：未配置 `public_base_url` 时，上传返回的 URL 基于请求 Host 头拼接，可被恶意构造。
- **定位**：`backend/internal/http/handler/image_handler.go:132-138`
- **整改**：与 H-01 联动——生产模式下 `public_base_url` 必填，URL 不再依赖 Host 头。开发模式保留 Host 回退。无需额外 Host allowlist。

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

### M-07 限流器 Redis 故障时 fail-open

- **风险**：Redis 故障期间上传接口失去限流保护。匿名上传是高频入口，影响尤为严重。
- **定位**：`rate_limit_middleware.go:28-35`
- **整改**：上传、登录等受限接口在 Redis 故障时直接拒绝请求（fail-closed），不做本机内存兜底。普通 GET 可 fail-open。

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

### Q-02 数据库 Schema 缺少版本号

- **现状**：开发版无需兼容迁移，但仍需要防止开发/测试环境 schema 漂移。
- **整改**：采用最低成本方案：使用 `PRAGMA user_version` 标记 schema 版本，不引入完整 migration 框架。每次 schema 结构变化必须 bump 版本号，并在测试中校验。

### Q-03 后端缺少优雅关闭

- **整改**：根 `context.WithCancel` + 信号监听；`Server.Shutdown` → 停止 heartbeat → 关闭 Redis/SQLite。

### Q-04 UID 混淆密钥命名与文档不准确

- **现状**：UID 使用 Snowflake + XOR + Base62，满足 ≤ 30 字符、可逆解码、前缀检查的设计要求。XOR 是混淆而非加密，但环境变量和代码注释中使用了"encryption"字样。
- **整改**：
  1. 环境变量 `UID_ENCRYPTION_KEY` 保留（避免破坏现有部署），文档和代码注释统一标注为"混淆密钥（obfuscation key），非加密安全边界"。
  2. 新增上下文说明：UID 的目的是避免暴露可预测的序列 ID，不提供密码学安全保证。

### Q-05 上传同步阻塞：AVIF 转换 + 存储写入在单次请求内完成

- **现状**：`POST /v1/image` 同步完成 multipart 解析 → MD5 → 去重 → AVIF 转换 → 存储写入 → 返回 URL。大图或慢存储会导致请求超时。
- **整改**：改为轻量异步：
  1. 接收 multipart → 立即存原始文件 → 返回 UID + `status: "processing"`
  2. 后台协程执行 AVIF 转换 + 存储写入
  3. 新增 `GET /v1/image/:uid/status` 端点，前端短轮询直到 `status: "ready"`
  4. 处理中的图片访问返回占位图或 loading 状态
  5. 前端 `uploadImageWithProgress` 在收到 `processing` 后启动轮询，完成后 resolve URL
- **失败处理**：
  - 后台转换或存储写入失败时，标记 `status: "failed"`，前端展示错误提示。不做自动重试（转换失败通常是永久性的）。
  - 服务启动时扫描卡在 `processing` 状态的孤儿记录，标记为 `failed`。

### Q-06 测试覆盖加强到安全边界

- **整改**：multipart 超大 body、默认密钥阻断、限流降级、X-Token 生成（无 Web Crypto 场景）、UID 前缀解码、异步上传状态转换等补充表驱动测试。

---

> **Acceptance Criteria 追踪**（2026-05-29 更新）
>
> - [x] 阶段 1 全部完成（7/7 项已落地）
> - [x] `cd backend && go test ./... && go vet ./...` — 248 passed
> - [x] `cd frontend && npm run lint && npm run typecheck && npm run test` — 57 passed
> - [x] `cd frontend && npm run build:backend` — 通过
> - [x] 无回归问题

## 4. 实施路线图

### 阶段 0：基础设施（1-2 天）

- [x] SQLite `PRAGMA user_version` schema 版本号（Q-02）
- [ ] CI 流水线（Q-01）
- [ ] Dockerfile + docker-compose（附带 `.env.production.example`）

### 阶段 1：入口安全与抗 DoS（3-5 天）

- [x] 强制要求 .env 配置管理员密码和密钥（H-01）
- [x] 显式 `http.Server` + 优雅关闭（H-02, Q-03）
- [x] 上传 `MaxBytesReader` body limit（H-03）
- [x] 限流 Redis 故障改为 fail-closed（M-07）
- [x] X-Token 移除 Math.random fallback（H-05）
- [x] 后台真实 IP 来源配置（M-02）
- [x] 强制 `public_base_url` 配置（M-03）

### 阶段 2：管理员认证与浏览器安全（2-3 天）

- [ ] CORS 拆分：公开 API 跨域 + 管理 API 同源 + 后台热更新（M-01）
- [ ] JWT Redis 撤销机制（M-06）
- [ ] 全局安全头 middleware + 移除 unsafe-inline + 缩短 JWT TTL（H-04 + M-08，同 PR）

### 阶段 3：部署与配置安全（2-3 天）

- [ ] SQLite 凭据加密（M-04）
- [ ] Cloudflare API Base URL 配置提示（M-05，低优先）

### 阶段 4：质量收尾（3-5 天）

- [ ] UID 混淆密钥命名修正（Q-04）
- [ ] 安全边界测试补充（Q-06）

### 阶段 5：异步上传（1-2 周）

- [ ] 上传 handler 拆分为接收 + 后台转换（Q-05）
- [ ] 新增 `GET /v1/image/:uid/status` 端点
- [ ] 前端轮询逻辑 + 处理中图片占位展示
- [ ] WriteTimeout 放宽到 300s 作为兜底（H-02）

---

## 5. 开发原则

1. **守住核心设计** — 匿名上传 + 单管理员，后端不签发用户 token。限流按 IP 维度实现，不引入账户体系。
2. **UID 约束不可破坏** — ≤ 30 字符、可逆解码、前缀可检查。Snowflake + XOR + Base62 方案保持不变，仅修正命名。
3. **安全分层** — 匿名侧：body 限制、限流、X-Token 客户端安全。管理侧：强密码、短 TTL、CORS 收窄。
4. **直接实施** — 开发版本无需迁移兼容层。
5. **测试伴随 + 文档同步** — 每项整改同步提交单测；API 字段、前端类型、README 随代码更新。
