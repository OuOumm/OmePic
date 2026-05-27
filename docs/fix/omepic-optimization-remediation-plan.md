# OmePic 项目优化与整改方案

> 审查时间：2026-05-28  
> 审查范围：当前工作区内的后端 Go/Gin、前端 SvelteKit/TypeScript、配置、文档、构建脚本；已排除 `frontend/node_modules/`、`frontend/out/`、`backend/web/` 等生成物。仓库实际前端为 **SvelteKit**，不是 React。根目录未发现已提交的 `Dockerfile`、`docker-compose.yml` 或 CI workflow，文档中仅提供了 Docker 示例片段。

## 1. 高危安全整改（必须最优先修复）

### H-01 默认管理员密码与默认签名密钥仅告警，不阻断生产启动

- **风险/缺陷描述**：首次部署会接受公开文档中的默认管理员密码 `admin123`；`JWT_SECRET`、`UID_ENCRYPTION_KEY` 也有固定默认值。代码仅记录 warning，不会阻断公网部署。攻击者可用默认密码登录管理后台，或在默认 JWT 密钥下伪造管理员 JWT。
- **具体定位**：
  - `backend/internal/service/admin_service.go:118`：`DefaultAdminPassword = "admin123"`。
  - `backend/internal/service/admin_service.go:270`：首次无密码哈希时写入默认密码 bcrypt 哈希。
  - `backend/internal/config/config.go:68-69`：默认 `UID_ENCRYPTION_KEY` / `JWT_SECRET`。
  - `backend/cmd/server/main.go:25-29,93-94`：仅打印默认密钥/密码告警。
  - `README.md:147,160-161`、`.env.example:8-9`：公开默认值。
- **推理依据**：管理接口包含图片删除、存储配置、Cloudflare token 配置、公告配置等高权限能力；默认口令和默认 JWT secret 是可远程利用的确定性凭据。
- **可操作整改方案**：
  1. 引入 `APP_ENV=production` 或 `OMEPIC_SETUP_REQUIRED` 启动模式；生产模式下发现默认密钥或默认密码时直接启动失败。
  2. 首次启动生成一次性 setup token（只打印到控制台或写入权限 `0600` 文件），通过 `/admin/setup` 设置强密码后禁用 setup。
  3. 对 `JWT_SECRET` 要求至少 32 字节随机值；`UID_ENCRYPTION_KEY` 同样要求强随机值。
  4. 保持向后兼容：开发模式仍可使用默认值，但管理后台持续显示高危告警；生产迁移时只需设置环境变量并修改密码。

```go
func enforceProductionSecrets(cfg config.AppConfig, repo *repository.Repository) error {
    if os.Getenv("APP_ENV") != "production" { return nil }
    if cfg.JWTSecret == "change-me-too" || len(cfg.JWTSecret) < 32 {
        return errors.New("JWT_SECRET must be a strong non-default value in production")
    }
    if cfg.UIDEncryptionKey == "change-me-uid-secret" || len(cfg.UIDEncryptionKey) < 32 {
        return errors.New("UID_ENCRYPTION_KEY must be a strong non-default value in production")
    }
    if usesDefaultAdminPassword(context.Background(), repo) {
        return errors.New("admin password must be changed before production start")
    }
    return nil
}
```

### H-02 HTTP 服务使用 `engine.Run`，缺少服务端超时和请求头限制

- **风险/缺陷描述**：`gin.Engine.Run` 使用默认 `http.Server`，缺少 `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout`、`IdleTimeout`、`MaxHeaderBytes`。公网环境可被 Slowloris、慢上传、超大请求头拖垮连接池和 goroutine。
- **具体定位**：`backend/cmd/server/main.go:130`。
- **推理依据**：Go `net/http` 默认超时为 0；项目是图片上传服务，长连接与慢请求更容易形成资源耗尽。
- **可操作整改方案**：
  1. 替换 `engine.Run(cfg.HTTPAddr)` 为显式 `http.Server`。
  2. 对上传、静态资源和普通 API 分别设置合理超时；上传可稍宽，普通 API 更严格。
  3. 增加优雅退出，避免部署重启时中断正在写入的存储对象。

```go
server := &http.Server{
    Addr:              cfg.HTTPAddr,
    Handler:           engine,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       60 * time.Second,
    WriteTimeout:      120 * time.Second,
    IdleTimeout:       60 * time.Second,
    MaxHeaderBytes:    1 << 20,
}
logger.Info("server starting", "addr", cfg.HTTPAddr)
if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
    return err
}
```

### H-03 上传入口在 multipart 解析前没有全局 body 限制

- **风险/缺陷描述**：`ImageHandler.Upload` 调用 `c.FormFile("file")` 后才基于 `fileHeader.Size` 判断大小；Gin/Go 在 `FormFile` 内部可能已经解析 multipart 并把大文件写入临时文件。攻击者可绕过业务层 `MaxUploadSizeMB`，通过超大 multipart 消耗磁盘、内存和 CPU。
- **具体定位**：
  - `backend/internal/http/handler/image_handler.go:32-33`：先 `FormFile`。
  - `backend/internal/service/image_service.go:413,427`：服务层 `LimitReader` 只限制已进入业务层的流。
- **推理依据**：业务限制发生在 multipart 解析之后，不等价于 HTTP body 限制；图片站上传接口是高频、高成本入口。
- **可操作整改方案**：
  1. 在路由中给 `/v1/image` 添加 upload body limit 中间件，先执行 `http.MaxBytesReader`。
  2. body 限制应为 `MaxUploadSizeBytes + multipartOverhead`，例如额外 1 MiB。
  3. 继续保留服务层 `LimitReader` 作为二次防线。

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
// engine.POST("/v1/image", uploadLimiter, BodyLimit(deps.Settings), deps.ImageHandler.Upload)
```

### H-04 管理员 JWT 持久化在 `localStorage`，且 CSP 允许内联脚本

- **风险/缺陷描述**：管理员 token 保存于浏览器 `localStorage`；CSP 中 `script-src` 包含 `'unsafe-inline'`。一旦任意页面发生 XSS 或第三方脚本/浏览器扩展注入，攻击者可读取 `omepic-admin-token` 并调用所有管理 API。
- **具体定位**：
  - `frontend/src/lib/stores/preferences.svelte.ts:7,60,94`：`omepic-admin-token` 读写 localStorage。
  - `backend/internal/http/router/frontend.go:17`：`script-src 'self' 'unsafe-inline'`。
  - `backend/internal/http/middleware/auth_middleware.go:14-21`：管理接口只依赖 bearer token。
- **推理依据**：localStorage 对同源脚本完全可读；`unsafe-inline` 会显著降低 CSP 对 XSS 的阻断能力。管理 token 有 24 小时有效期（`backend/internal/service/admin_service.go:200`）。
- **可操作整改方案**：
  1. 将管理员会话迁移到 `HttpOnly + SameSite=Lax/Strict` cookie；生产 HTTPS 时启用 `Secure`。
  2. 对 cookie 鉴权的状态变更接口增加 CSRF token 或至少强制自定义 header + Origin 校验。
  3. 如需保持前后端分离兼容，可短期改为内存保存 access token + 短有效期 refresh cookie。
  4. 消除内联脚本：把初始主题脚本移入静态 JS，或为动态响应生成 nonce，并移除 `'unsafe-inline'`。

```go
http.SetCookie(w, &http.Cookie{
    Name: "omepic_admin_session", Value: signedToken,
    Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode,
    Secure: cfg.CookieSecure,
    MaxAge: 86400,
})
```

## 2. 中低危安全与逻辑缺陷修复

### M-01 默认 CORS 允许任意 Origin

- **风险/缺陷描述**：未配置 `PublicBaseURL` 时，CORS `AllowAllOrigins=true`，且允许 `Authorization`、`X-Token` 请求头。虽然当前未启用 cookies，仍会放大泄漏 token 后的跨站调用面，也不符合最小权限原则。
- **具体定位**：`backend/internal/http/router/router.go:104-118`；测试还固定了该行为：`backend/internal/http/router/frontend_test.go:195-198`。
- **推理依据**：管理 API 是 bearer token 模式，CORS 不应默认为全网开放；公开上传 API 可按需要开放，但管理 API 应严格同源或 allowlist。
- **整改方案**：
  - 默认仅允许同源；配置 `CORS_ALLOWED_ORIGINS` 或复用 `PublicBaseURL`。
  - 将公开 API 与管理 API CORS 策略拆分，管理 API 禁止 `*`。
  - 平滑升级：开发模式保留 localhost 常见端口白名单。

### M-02 代理 IP 信任配置不可配置，限流/IP 封禁易失效或误伤

- **风险/缺陷描述**：入口创建 `clientip.NewResolver(nil, "")`，没有读取可信代理 CIDR。部署在 Nginx/Cloudflare/容器网关后，应用会把反向代理 IP 当作客户端 IP，导致所有用户共享限流、IP 封禁误伤，滥用审计失真。
- **具体定位**：`backend/cmd/server/main.go:109`；解析器只有可信代理时才读取 `X-Forwarded-For`：`backend/internal/http/clientip/resolver.go:32-40`。
- **推理依据**：限流键依赖 IP hash（`backend/internal/http/middleware/rate_limit_middleware.go:28-35`），IP 封禁也基于解析后的 IP。
- **整改方案**：
  - 新增 `TRUSTED_PROXY_CIDRS`、`REAL_IP_HEADER` 环境变量。
  - 只在 `RemoteAddr` 命中可信 CIDR 时采信 `X-Forwarded-For` 或 `X-Real-IP`。
  - 文档给出 Cloudflare、Nginx、Docker bridge 示例。

### M-03 上传 URL 基于 `Host` 头推导，存在 Host Header Poisoning 风险

- **风险/缺陷描述**：未配置 `public_base_url` 时，上传返回 URL 使用请求 `Host`。攻击者可构造恶意 Host，让响应中生成钓鱼域名图片 URL，用户复制分享后进入错误域名。
- **具体定位**：`backend/internal/http/handler/image_handler.go:132-138`；`backend/internal/service/runtime_settings.go:207-214`。
- **推理依据**：`Host` 是客户端可控输入，不能作为可信外部 URL 的来源。
- **整改方案**：
  - 生产模式强制配置 `public_base_url` 或 `PUBLIC_BASE_URL_ALLOWED_HOSTS`。
  - 对 `c.Request.Host` 做 allowlist 校验，失败时回退到配置值或拒绝请求。

### M-04 存储与第三方凭据明文保存在 SQLite

- **风险/缺陷描述**：S3 access key/secret、WebDAV 密码、Cloudflare API token 存储在 SQLite 明文字段中。数据库文件泄露即可接管外部存储或 Cloudflare 缓存清理权限。
- **具体定位**：
  - `backend/internal/repository/migration.go:25-37`：凭据字段。
  - `backend/internal/repository/storage_repository.go:20,188-205`：明文写入。
  - `backend/internal/service/runtime_settings_fields.go:77-83`：Cloudflare token 字段。
- **推理依据**：UI 返回时做了遮罩（`backend/internal/service/admin_service.go:631-651`），但持久化层仍为明文。
- **整改方案**：
  - 新增 `SECRET_ENCRYPTION_KEY`，使用 AES-GCM 对 secret 字段信封加密。
  - 增加密钥版本字段，支持后续轮换。
  - 迁移策略：启动时检测明文字段，后台任务逐条加密；读取逻辑兼容明文/密文，写入只写密文。

### M-05 Cloudflare API Base URL 可配置为任意 HTTP/HTTPS 地址

- **风险/缺陷描述**：`cloudflare_api_base_url` 仅校验协议为 http/https，之后会把 `Authorization: Bearer <apiToken>` 发往该地址。配置被误改、低权限管理员误操作或 XSS 修改配置时，可导致 token 外泄；也会造成服务端向任意地址发请求。
- **具体定位**：
  - `backend/internal/service/runtime_settings.go:270-273`：只校验 URL 格式。
  - `backend/internal/service/cloudflare_cache.go:71-78`：向 `apiBaseURL` 发送 bearer token。
- **推理依据**：Cloudflare token 是高价值凭据；默认不应允许任意出站目标。
- **整改方案**：
  - 默认固定 `https://api.cloudflare.com/client/v4`。
  - 如需私有代理，增加 `ALLOW_CUSTOM_CLOUDFLARE_API_BASE=true` 和 host allowlist。
  - 禁止明文 `http`，除非显式 dev 模式。

### M-06 JWT 缺少 issuer/audience/jti 与会话撤销机制

- **风险/缺陷描述**：JWT 只有 `iat`/`exp`，密码修改后旧 token 仍在 24 小时内有效；不能按设备撤销会话，也不能区分签发方/受众。
- **具体定位**：`backend/internal/auth/jwt.go:15-24`；`backend/internal/service/admin_service.go:200`。
- **推理依据**：管理后台高权限操作需要可撤销会话；密码泄露后的止血能力不足。
- **整改方案**：
  - Claims 增加 `Issuer: "omepic"`、`Audience: ["omepic-admin"]`、`ID`。
  - 在 SQLite/Redis 保存 `admin_session_version` 或 jti denylist；修改密码时递增版本或清空会话。
  - token TTL 缩短到 15-60 分钟，配合 refresh token/cookie。

### M-07 管理图片列表直接返回上传者 X-Token

- **风险/缺陷描述**：管理图片列表将每张图片的 `Token` 返回前端。管理员本身可删除图片，但把用户删除凭据展示在 UI/API 中会扩大泄露面；浏览器插件、截图或前端 XSS 都可获得用户 token，并在管理会话失效后继续删除对应图片。
- **具体定位**：`backend/internal/service/admin_service.go:316-321`。
- **推理依据**：`X-Token` 是公开删除接口的授权凭据（`backend/internal/http/handler/image_handler.go:83-86`）。
- **整改方案**：
  - Admin API 默认返回 token hash 或末 6 位遮罩值。
  - 如确需排查，新增单独的“查看凭据”接口，要求二次确认/重新输入密码并写审计日志。

### M-08 限流器 Redis 故障时 fail-open

- **风险/缺陷描述**：Redis 限流失败时仅记录 warning 并继续处理请求。Redis 抖动期间，上传、登录、API 均失去限流保护。
- **具体定位**：`backend/internal/http/middleware/rate_limit_middleware.go:28-35`。
- **推理依据**：上传转换 AVIF 成本高，限流是关键防滥用控制。
- **整改方案**：
  - 登录和上传接口采用 fail-closed 或本机内存降级限流。
  - 普通 GET API 可 fail-open，但记录指标告警。

### M-09 客户端 X-Token 长期保存在 localStorage，且允许 Math.random 降级

- **风险/缺陷描述**：删除授权 token 长期保存在 `localStorage`；极旧环境下会降级到 `Math.random`。一旦 XSS/扩展读取 token，攻击者可删除用户历史图片。
- **具体定位**：`frontend/src/lib/client-token.ts:1-26`。
- **推理依据**：公开删除接口以 `X-Token` 判定所有权，token 是敏感凭据。
- **整改方案**：
  - 移除 `Math.random` fallback；没有 Web Crypto 时提示浏览器不受支持。
  - 后端可签发 scoped delete token，支持轮换和吊销。
  - 前端保留兼容读取旧 token，首次使用后迁移到新凭据格式。

### M-10 CSP 仅应用于前端静态兜底路由，API/图片响应安全头不统一

- **风险/缺陷描述**：`setFrontendSecurityHeaders` 在 NoRoute 静态前端路径设置；API 路由和图片响应没有统一安全头。虽然图片响应风险较低，但统一基线更便于安全审计。
- **具体定位**：`backend/internal/http/router/frontend.go:35-39,68-75`；`backend/internal/http/handler/image_handler.go:104-107`。
- **推理依据**：安全头在中心化中间件中更容易覆盖所有浏览器响应。
- **整改方案**：新增全局 `SecurityHeaders` middleware；对 API 设置 `X-Content-Type-Options`、`Referrer-Policy`，前端页面额外设置 CSP/frame-ancestors。

## 3. 代码质量与架构优化

### Q-01 缺少真实提交的 Docker 部署文件

- **缺陷描述**：`docs/wiki/running-and-deployment.md:185-253` 仅说明“创建 Dockerfile / docker-compose.yml”，但仓库根目录未发现实际 `Dockerfile` 或 `docker-compose.yml`。
- **影响**：用户按开源项目部署时需要手工复制文档片段，容易遗漏 secret、数据卷、健康检查和非 root 用户配置。
- **整改方案**：提交生产级多阶段 Dockerfile 与 compose：
  - 前端 `npm ci && npm run build`；后端复制 `backend/web` 后 `go build`。
  - 运行镜像使用非 root 用户、只挂载 `/data`、内置 `HEALTHCHECK`。
  - compose 管理 Redis、app、持久卷和 `.env`。

### Q-02 缺少 CI 安全与质量流水线

- **缺陷描述**：未发现 `.github/workflows/*`。当前质量命令散落在文档：`go test`, `go vet`, `npm run lint/typecheck/test/build:backend`。
- **影响**：依赖漏洞、类型错误、迁移回退和前端构建问题可能进入主分支。
- **整改方案**：增加 CI：
  1. Go：`gofmt -w` 检查、`go vet ./...`、`go test ./...`、`govulncheck ./...`。
  2. Frontend：`npm ci`、`npm run lint`、`npm run typecheck`、`npm run test`、`npm run build:backend`。
  3. 依赖：`npm audit --omit=dev` 或 Dependabot/Renovate；Go module 自动更新 PR。

### Q-03 数据库迁移缺少版本化演进机制

- **缺陷描述**：`backend/internal/repository/migration.go:9-113` 主要使用 `CREATE TABLE IF NOT EXISTS` 和 `CREATE INDEX IF NOT EXISTS`，未见 `schema_version` 或 `PRAGMA user_version` 版本迁移。
- **影响**：后续新增列、改字段类型、回填数据时难以保证幂等和可回滚。
- **整改方案**：
  - 使用 `PRAGMA user_version` 或引入轻量迁移表 `schema_migrations`。
  - 每个迁移有 `up`、回填、校验步骤；启动前备份 SQLite。
  - 保持现有表结构不变，逐版本平滑升级。

### Q-04 后端服务缺少优雅关闭和后台任务生命周期管理

- **缺陷描述**：主进程直接 `engine.Run`；存储健康检查 heartbeat 用 `context.Background()` 启动。
- **影响**：部署重启时上传/AVIF 转换/SQLite 写入可能被中断，后台任务难以统一停止。
- **整改方案**：
  - 使用根 `context.WithCancel`，监听 SIGINT/SIGTERM。
  - `http.Server.Shutdown(ctx)` 后停止 storage heartbeat，再关闭 Redis/SQLite。

### Q-05 UID 使用 XOR 混淆，不应标注为“加密”或“不可预测”安全保证

- **缺陷描述**：`backend/internal/uid/codec.go:78-89,179-185` 使用 Snowflake + XOR + Base64/Base62。XOR 不是现代加密；如果 secret 默认或泄露，可反推出时间序列信息。
- **影响**：安全文档中把它作为“不可预测 UID”容易误导；虽然当前仅作为对象 ID，不直接导致鉴权绕过，但应加强。
- **整改方案**：
  - 新生成 UID 改用 `crypto/rand` 128-bit/160-bit 随机 ID，保留旧 UID 解码兼容。
  - 或使用 HMAC/AEAD 封装 Snowflake payload。
  - 文档将“UID_ENCRYPTION_KEY”改名为“UID_OBFUSCATION_KEY”或迁移到新随机 UID 后废弃。

### Q-06 文档路径与实际结构不一致

- **缺陷描述**：`README.md:199` 链接 `docs/api-reference.md`，实际文件位于 `docs/wiki/api-reference.md`。
- **影响**：开源用户无法从 README 跳转完整 API 文档。
- **整改方案**：修正 README 链接；增加文档链接检查 CI。

### Q-07 运行时配置缺少操作审计

- **缺陷描述**：管理端可以修改存储、Cloudflare、限流、公告等高风险配置，但未见审计日志表。
- **影响**：误操作或账号被盗后难以定位变更来源、时间和内容。
- **整改方案**：新增 `audit_logs` 表：记录 actor、IP、action、resource、diff、request_id、created_at；敏感字段只记录“changed=true”。

### Q-08 测试覆盖还可加强到安全边界与部署边界

- **当前优点**：仓库已有大量 Go/Vitest 单测，覆盖上传、配置、缓存、前端工具等。
- **缺口**：缺少 multipart 超大 body、默认密钥生产阻断、CORS 策略、可信代理、JWT 撤销、Docker 构建等测试。
- **整改方案**：为上述整改新增表驱动测试和 e2e smoke test；至少覆盖“旧配置可启动、新生产配置强制安全”的升级路径。

## 4. 功能扩展建议

### F-01 多用户/API Key 与配额体系

- **价值**：当前 `X-Token` 是匿名客户端凭据，不适合多人或团队使用。引入用户/API Key 可支持配额、审计、禁用、个人空间。
- **难度**：高。
- **优先级**：P1。
- **实现思路**：新增 `users`、`api_keys`、`quotas` 表；上传记录关联 `owner_id`；旧 `X-Token` 自动映射为 legacy owner，保证历史删除不破坏。

### F-02 相册/目录与批量管理

- **价值**：图床常见需求，便于公开分享一组图片、按项目组织素材。
- **难度**：中。
- **优先级**：P2。
- **实现思路**：新增 `albums`、`album_images`；前端历史/管理页支持批量加入相册；公开相册页可设置访问密码/过期时间。

### F-03 原图备份与多格式输出

- **价值**：当前统一转换 AVIF；部分用户需要保留原图或输出 WebP/JPEG fallback。
- **难度**：中高。
- **优先级**：P2。
- **实现思路**：存储 `original_path`、`variants`；异步生成 AVIF/WebP/JPEG；URL 形如 `/i/:uid.avif`、`/i/:uid.webp`，旧 AVIF URL 不变。

### F-04 内容审核与安全扫描

- **价值**：公开上传服务需要防止违法/敏感/恶意图片滥用。
- **难度**：高。
- **优先级**：P2。
- **实现思路**：上传后进入 `pending/approved/rejected` 状态；接入本地模型或第三方审核；高风险内容仅管理员可见，保留申诉/复核。

### F-05 回收站与软删除

- **价值**：避免误删图片无法恢复，管理后台删除更安全。
- **难度**：中。
- **优先级**：P2。
- **实现思路**：`images` 增加 `deleted_at`、`deleted_by`；物理对象延迟清理；公开访问返回 404，管理员可恢复。

### F-06 存储健康、故障切换与迁移任务

- **价值**：项目已有存储健康检查雏形，可扩展为自动故障切换和迁移。
- **难度**：高。
- **优先级**：P3。
- **实现思路**：健康状态影响默认存储选择；后台任务按批次复制对象并校验 hash；切换时保持旧 URL 不变。

### F-07 Prometheus 指标与结构化追踪

- **价值**：便于生产运维观察上传量、错误率、AVIF 耗时、Redis/SQLite 状态。
- **难度**：中。
- **优先级**：P3。
- **实现思路**：`/metrics` 默认只监听内网或需 token；日志加入 request_id、client_ip_hash、storage_key、duration。

### F-08 短链和访问统计

- **价值**：提升分享体验，分析图片访问趋势和热门资源。
- **难度**：中。
- **优先级**：P3。
- **实现思路**：新增短链表和访问事件聚合；默认只统计匿名聚合数据，避免保存完整 IP。

## 5. 实施路线图

### 第 0 阶段：不破坏现有功能的安全护栏（1-2 天）

1. 修复 README 错误链接，补充“当前前端为 SvelteKit”。
2. 新增生产模式默认密钥/默认密码启动阻断，但默认开发模式保持兼容。
3. 提交 Dockerfile、docker-compose.yml、`.env.production.example`。
4. 建立 CI：Go/Frontend 测试、构建和基础漏洞扫描。

### 第 1 阶段：公网入口抗 DoS（2-4 天）

1. 显式 `http.Server` 超时、Header 限制、优雅关闭。
2. `/v1/image` 增加 `MaxBytesReader` body limit。
3. Redis 限流故障降级为内存限流；上传/登录接口优先 fail-closed。
4. 增加对应单元测试和压测脚本。

### 第 2 阶段：管理认证与浏览器安全（3-7 天）

1. 管理 token 从 localStorage 迁移到 HttpOnly cookie 或内存 access token + refresh cookie。
2. 增加 CSRF/Origin 校验，拆分管理 API CORS。
3. JWT 增加 issuer/audience/jti/session_version；修改密码后撤销旧 token。
4. 移除 CSP `unsafe-inline` 或改为 nonce/hash 方案。

### 第 3 阶段：部署与运行时配置安全（3-5 天）

1. 配置可信代理 CIDR，修正限流/IP 封禁真实客户端 IP。
2. 强制生产 `public_base_url` 或 Host allowlist，避免 Host header poisoning。
3. Cloudflare API Base URL 默认锁定官方域名；自定义代理需显式 allowlist。
4. SQLite secret 字段 AES-GCM 加密并提供明文兼容迁移。

### 第 4 阶段：数据模型与审计（1-2 周）

1. 引入版本化迁移系统。
2. 新增审计日志表与后台审计页面。
3. 管理图片列表默认遮罩 `X-Token`，敏感查看需二次确认。
4. UID 新生成策略迁移到随机 ID/HMAC，旧 UID 保持可读。

### 第 5 阶段：产品扩展（持续迭代）

1. P1：多用户/API Key/配额体系。
2. P2：相册、软删除、原图备份、多格式输出、内容审核。
3. P3：存储故障切换、Prometheus 指标、访问统计、短链。

### 升级兼容原则

- 所有安全开关先以“开发兼容、生产强制”的方式落地。
- 数据库迁移必须可重复执行；涉及 secret 加密时先读兼容、再写新格式。
- 公开图片 URL `/i/:uid.avif`、历史 `X-Token` 删除能力和已有存储路径保持兼容，避免现有用户链接失效。
