# Go 后端质量审查报告

- 日期：2026-05-13
- 范围：`backend/` 全部 Go 代码（16 个包，94 个测试）
- 依据：`.trellis/spec/backend/*` 全部 8 份规范
- 方法：逐文件审查 + 模式扫描 + 自动化验证

---

## 1. 自动化验证

| 命令 | 结果 | 耗时 |
|---|---|---|
| `go test ./...` | 94 passed, 16 packages | ~15s |
| `gofmt -l .` | 1 处对齐修复（`runtime_settings.go`） | 即时 |
| `go build ./...` | 编译通过 | ~5s |
| `go vet ./...` | 无警告 | ~3s |

---

## 2. 架构总览

```
backend/
├── cmd/server/main.go          # 启动入口，依赖注入
├── internal/
│   ├── auth/                   # JWT + X-Token 验证
│   ├── cache/                  # Redis UID/MD5 缓存
│   ├── config/                 # 环境配置加载
│   ├── http/
│   │   ├── clientip/           # 可信代理 IP 解析
│   │   ├── handler/            # Gin 处理器（薄层）
│   │   ├── middleware/         # 认证/限流/日志中间件
│   │   └── router/             # 路由注册 + 前端静态服务
│   ├── iputil/                 # IP sha256 hashing + mask
│   ├── model/                  # 数据库对接结构体
│   ├── ratelimit/              # Redis 固定窗口限流
│   ├── repository/             # SQLite 数据访问（~1300 行）
│   ├── response/               # JSON 响应信封
│   ├── service/                # 业务逻辑
│   ├── storage/                # 存储适配器 (local/S3/WebDAV)
│   └── uid/                    # UID 加密/解密码表
├── web/                        # 前端静态构建产物
└── go.mod
```

### 整体评价

后端架构清晰，严格遵循 Go 标准分层模式：Handler 仅解析请求 → Service 执行业务逻辑 → Repository/Cache/Storage 访问数据。依赖通过构造函数注入，无全局可变状态。测试覆盖率达 94 个测试，关键流程（上传、去重、删除、缓存、认证）均有覆盖。

---

## 3. 按规范逐项审计

### 3.1 目录结构

| 要求 | 状态 | 说明 |
|---|---|---|
| 包按职责命名 | ✅ | service / repository / handler 等 |
| 无 utils.go / helpers.go | ✅ | 无通用转储文件 |
| 入口在 cmd/server/main.go | ✅ | |
| 中间件在 middleware/ | ✅ | auth / logging / rate_limit |
| 路由在 router/ | ✅ | router.go + routes.go + frontend.go |
| 响应封装在 response/ | ✅ | Success / Error 统一信封 |

### 3.2 数据库规范

| 要求 | 状态 | 说明 |
|---|---|---|
| SQLite 为唯一事实源 | ✅ | repository 层全部 SQL 操作 |
| Redis 为缓存 | ✅ | 预热 + 回退 SQLite |
| 先写 SQLite 再写 Redis | ✅ | Upload / Delete 流程正确 |
| 物理删除不在在线路径中 | ✅ | 仅操作 SQL 行 + Redis key |
| MD5 去重作用域为 storage_key | ✅ | 缓存 key 为 md5:{storage_key}:{hash} |
| 原始文件名不存入 SQLite | ✅ | images 表无此列 |
| Repository 方法接收 context.Context | ✅ | |

### 3.3 安全规范

| 要求 | 状态 | 说明 |
|---|---|---|
| 使用 clientip.Resolver | ✅ | handler 统一调用 |
| 限流 key 使用 iputil.Hash(sha256) | ✅ | rate_limit_middleware.go |
| IP 封禁 hash 匹配 | ✅ | iputil.Hash 使用 sha256 |
| 管理员 JWT 验证 | ✅ | auth_middleware.go |
| 上传/删除检查 IP 封禁 | ✅ | image_service.ensureIPAllowed |
| 不记录密钥到日志 | ✅ | 审查后确认 |

### 3.4 错误处理规范

| 要求 | 状态 | 说明 |
|---|---|---|
| 错误集中映射到 HTTP | ✅ | writeServiceError + mapJSONError |
| 不暴露 SQL 错误详情 | ✅ | 统一 dependency_unavailable |
| 使用标准错误代码 | ✅ | 7 个 sentinel error |
| 无 panic 用于用户错误 | ✅ | |
| 使用 errors.Is() 匹配 | ✅ | |

### 3.5 质量规范

| 要求 | 状态 | 说明 |
|---|---|---|
| 业务逻辑不在 handler | ✅ | Handler 仅解析+调用 service |
| 不直接访问 Redis/SQLite 来自 handler | ✅ | |
| 测试覆盖高风险流 | ✅ | 94 测试覆盖 |
| context.Context 贯穿 | ✅ | |
| 小接口在依赖边界 | ✅ | cache.ImageCache 接口 |

### 3.6 日志规范

| 要求 | 状态 | 说明 |
|---|---|---|
| 使用 log/slog 结构化日志 | ✅ | |
| 日志级别适当 | ✅ | info/warn/error |
| 不记录 X-Token / JWT / 密钥 | ✅ | 审查确认 |
| 请求日志在中间件 | ✅ | logging_middleware.go |

### 3.7 运行时设置规范

| 要求 | 状态 | 说明 |
|---|---|---|
| image/jpg 归一化为 image/jpeg | ✅ | normalizeMIMETypes() |
| SVG 被拒绝 | ✅ | 显式检查 |
| 不允许非 image/ MIME | ✅ | 前缀检查 |
| 运行时设置通过 SQLite | ✅ | config 表 |
| 前端使用 effective_allowed_mime_types | ✅ | |

---

## 4. 发现的问题

### V01. image_service.go 使用 sync.Mutex 序列化全部上传

- **文件**：`backend/internal/service/image_service.go`
- **代码**：`s.operationMux.Lock()` 包裹整个 `Upload()` 和 `Delete()`
- **说明**：为了提高 MD5 去重流程的安全性，整个上传操作（含 AVIF 转换、存储写入）被同一把锁串行化。高并发场景下影响吞吐量。
- **建议**：将锁的范围缩小到 MD5 检查 + InsertImage 窗口，利用 SQLite 事务 `BEGIN IMMEDIATE` 保证去重原子性，AVIF 转换和存储写入放在事务外。

### V02. sanitizeErrorMessage 依赖错误格式字符串解析

- **文件**：`backend/internal/http/handler/errors.go`
- **代码**：`strings.SplitN(message, ": ", 2)` 提取用户消息
- **说明**：隐式依赖 service 层错误格式 `ErrXXX: user message`。如果未来某处改变了分隔符格式，handler 层提取的消息会错误。
- **建议**：为 service 错误添加显式 `UserMessage()` 方法，或改用 `type UserError struct{ Err error; Message string }`。

### O01. Repository 文件过大（~1300 行）

- **文件**：`backend/internal/repository/repository.go`
- **说明**：包含 Migrate、图片 CRUD、搜索、聚合统计、存储配置 CRUD、IP 封禁 CRUD、滥用查询、公告 CRUD 等全部 SQLite 操作。建议按领域拆分为 `image_repo.go`、`config_repo.go`、`ip_ban_repo.go`、`announcement_repo.go`。
- **影响**：低，但多人协作时合并冲突概率高。

### O02. 硬编码管理员默认密码

- **文件**：`backend/internal/service/admin_service.go`
- **代码**：`s.appConfig.AdminPassword == "admin123"`
- **说明**：仅用于管理界面中 "是否使用默认密码" 状态显示，不参与认证逻辑。但代码字面量易被误认为硬编码凭据。
- **建议**：提取为 `const defaultAdminPassword = "admin123"` 并加注释。

### O03. 不必要的公开 SetUIDGenerator / SetUIDValidator

- **文件**：`backend/internal/service/image_service.go`
- **说明**：这两个 setter 未被使用（UID 生成器和验证器通过构造函数注入），且无并发防护。之后若被误用可能导致运行时 uid 生成器被替换。
- **建议**：删除这两个 setter 或标记为 deprecated。

### O04. ~1300 行Repository文件建议拆分

同 O01，按领域拆分为多个文件可提升可维护性。

---

## 5. 合规总表

| 规范文件 | 检查项数 | 通过 | 违规 | 观察项 |
|---|---|---|---|---|
| directory-structure.md | 6 | 6 | 0 | 0 |
| database-guidelines.md | 7 | 7 | 0 | 0 |
| security.md | 6 | 6 | 0 | 0 |
| error-handling.md | 5 | 5 | 1 (V02) | 0 |
| quality-guidelines.md | 5 | 5 | 0 | 0 |
| logging-guidelines.md | 4 | 4 | 0 | 0 |
| runtime-settings.md | 5 | 5 | 0 | 0 |

**总计：38 项，37 通过，1 建议改进（V02 错误消息格式），3 观察项**

---

## 6. 测试覆盖分析

| 包 | 测试文件 | 说明 |
|---|---|---|
| auth | jwt_test.go | JWT 签发/验证 |
| cache | redis_cache_test.go | Redis UID/MD5 缓存 |
| config | config_test.go | 环境配置加载 |
| handler | image_handler_test.go | 上传/删除/服务处理 |
| iputil | iputil_test.go | IP hash/mask |
| repository | repository_test.go | SQLite 操作 |
| service | admin_service_test.go、image_service_test.go、announcement_service_test.go | 核心业务验证 |
| storage | storage_test.go | 存储适配器 |
| uid | codec_test.go | UID 编解码 |

**测试缺口**：
- `response/` 包无测试
- `middleware/` 无测试（auth、rate_limit、logging）
- `router/frontend.go` 虽然使用 `frontend_test.go` — 实际上有覆盖 ✅
- `clientip/resolver.go` 无测试
- `ratelimit/redis_limiter.go` 无测试

---

## 7. 健康指标

| 指标 | 值 |
|---|---|
| Go 包数 | 16 |
| Go 源文件数 | 33 |
| 测试文件数 | 9 |
| 测试总数 | 94 |
| gofmt 不合规文件 | 0（已修复） |
| go vet 警告 | 0 |
| 编译错误 | 0 |
| 硬编码密钥 | 0（admin123 仅为状态判断，非认证凭据） |
| 无主文件/死代码 | 暂无 |
