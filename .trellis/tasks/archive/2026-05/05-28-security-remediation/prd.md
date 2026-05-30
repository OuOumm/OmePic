# 安全与质量整改实施

## Goal

根据 `docs/fix/omepic-optimization-remediation-plan.md` 实施安全与质量整改，按路线图阶段 0 → 阶段 5 顺序推进。

## Core Design Constraints

- **前端匿名生成 Token 上传 + 单密码管理员后台**
- 后端不生成、不签发客户端 Token；X-Token 由前端 Web Crypto API 生成
- UID ≤ 30 字符、可逆解码、支持前缀检查；保持 Snowflake + XOR + Base62
- 开发版本，无需迁移兼容层

## Phase 0: 基础设施（1-2 天）

### Q-02 SQLite PRAGMA user_version schema 版本号
- 使用 `PRAGMA user_version` 标记 schema 版本
- 不引入完整 migration 框架
- 每次 schema 变化 bump 版本号，测试中校验

## Phase 1: 入口安全与抗 DoS（3-5 天）

### H-01 强制要求 .env 配置管理员密码和密钥
- 启动时强制读取 `ADMIN_PASSWORD`、`JWT_SECRET`、`UID_ENCRYPTION_KEY`
- 任一未设置则直接启动失败
- 删除 `DefaultAdminPassword = "admin123"` 和默认密钥常量

### H-02 显式 http.Server + 优雅关闭
- ReadHeaderTimeout: 5s, ReadTimeout: 60s, WriteTimeout: 300s, IdleTimeout: 60s
- MaxHeaderBytes: 1MB
- 配合 Q-03 优雅退出

### H-03 上传 MaxBytesReader body limit
- 路由层添加 `MaxBytesReader` 中间件
- body 限制 = MaxUploadSizeBytes + 1MiB

### M-07 限流 Redis 故障 fail-closed
- 上传、登录等受限接口 Redis 故障时直接拒绝请求
- 普通 GET 可 fail-open

### H-05 X-Token 移除 Math.random fallback
- 移除 `Math.random` 分支
- 没有 Web Crypto API 时直接 throw Error

### M-02 后台真实 IP 来源配置
- 运行时设置新增"真实 IP 来源"配置项
- 可选: X-Forwarded-For, X-Real-IP, CF-Connecting-IP, 直连 RemoteAddr
- 管理员在设置页选择，支持热更新

### M-03 强制 public_base_url 配置
- 与 H-01 联动，生产模式下 public_base_url 必填
- URL 不再依赖 Host 头

## Acceptance Criteria

- [ ] Phase 0 全部完成
- [ ] Phase 1 全部完成
- [ ] `cd backend && go test ./... && go vet ./...` 通过
- [ ] `cd frontend && npm run lint && npm run typecheck && npm run test` 通过
- [ ] `cd frontend && npm run build:backend` 通过
- [ ] 无回归问题

## Notes

- 详细方案见 `docs/fix/omepic-optimization-remediation-plan.md`
- Phase 2-5 后续任务在本任务完成后按需创建新任务
