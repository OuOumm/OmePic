# 任务：URL 上传前端下载改造（取消后端下载）

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` 原 **§4.2 P0：URL 上传 SSRF 防护**。
- 2026-05-25 用户调整：**URL 上传改为前端下载图片然后上传到后端，而不是在后端进行下载**。

## 预估人日

3 人日（原估算；实现方向已调整）

## 当前结论

- 后端不再提供 `POST /v1/image/url`，也不再主动请求用户提供的远端 URL。
- URL 上传由浏览器端执行 `fetch(url)`，将响应转换为 `File` 后复用现有 `POST /v1/image` multipart 上传链路。
- 后端仍通过现有文件上传链路执行 Token、真实 MIME、像素、大小、AVIF 转换、去重、存储与缓存修复等校验。
- 因后端不再访问用户提供的 URL，后端 SSRF 攻击面被移除；浏览器端下载仍受浏览器 CORS/网络策略约束。

## 开发

### 已移除范围

- 后端 `POST /v1/image/url` 路由与 handler。
- 后端 `RemoteImageFetcher`、`UploadRemoteURL`、URL 下载超时/跳转/DNS 解析逻辑。
- 后端 `backend/internal/util/url_safety.go` 与对应测试。
- 后端 URL 上传 handler 集成测试。
- 前端 `uploadImageURL()` API helper。

### 当前实现范围

- `frontend/src/routes/+page.svelte`：
  - 校验 URL scheme 仅允许 `http` / `https`。
  - 前端 `fetch` 远端图片。
  - 校验响应状态、Content-Length / Blob 大小、图片 MIME 与 SVG 阻断规则。
  - 根据 URL path 与 MIME 生成临时文件名。
  - 将 Blob 包装为 `File` 后调用 `enqueueFiles([file], maintenanceMode)`，复用现有上传队列与历史记录逻辑。
- `frontend/src/lib/api.ts`：移除后端 URL 上传 helper，仅保留 multipart 文件上传。
- `backend/internal/http/router/routes.go` / `router.go`：移除 `/v1/image/url` 注册。

### 完成标准

- [x] 后端代码中不存在 `/v1/image/url`、`UploadRemoteURL`、`RemoteImageFetcher` 等后端下载入口。
- [x] 前端 URL 上传不再调用后端 URL 接口。
- [x] URL 下载成功后进入现有上传队列，并复用文件上传接口。
- [x] 本地验证通过。

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 输入可由浏览器访问且 CORS 允许的图片 URL | 浏览器下载为 File，并通过 multipart 上传成功 |
| 异常 | URL 非 http/https | 前端提示 URL 无效 |
| 异常 | 远端响应非 2xx | 前端提示上传失败 |
| 异常 | 响应体超过运行时上传大小限制 | 前端拒绝并提示错误 |
| 异常 | 响应 MIME 非图片或为 SVG | 前端拒绝并提示错误 |
| 回归 | 后端访问 `/v1/image/url` | 不再注册该接口 |

### 验证命令

- `cd backend && gofmt -w ./cmd ./internal && go test ./...`
- `cd frontend && npm run lint`
- `cd frontend && npm run typecheck`
- `cd frontend && npm run test`
- `cd frontend && npm run build:backend`

## 复查

### 代码审查关注点

- 后端不得再主动请求用户输入的 URL。
- 前端 URL 上传必须复用现有文件上传队列与历史记录逻辑，避免出现第二套上传结果处理。
- 文件大小/MIME 的前端预校验不得替代后端真实文件校验；后端仍是最终安全边界。
- 删除 `/v1/image/url` 后，路由列表与 API helper 不得残留旧接口。

### 安全检查

- 后端 SSRF 面已通过移除后端 URL 下载能力消除。
- 浏览器端下载受同源/CORS/浏览器网络策略约束。
- 后端仍对最终上传文件执行真实 MIME、像素、大小与允许类型校验。

## 完成报告

### 实施摘要

- 按用户要求，URL 上传已改为前端下载图片后上传到后端。
- 移除了后端 URL 下载 API、远端抓取服务、URL 安全工具和相关测试。
- 前端 URL 输入现在通过浏览器 `fetch` 获取图片 Blob，转换为 `File` 后进入现有上传队列。
- 文件最终仍通过 `POST /v1/image` multipart 上传，复用后端既有安全校验与存储链路。

### 修改文件

- `backend/internal/http/handler/image_handler.go`
- `backend/internal/http/router/router.go`
- `backend/internal/http/router/routes.go`
- `backend/internal/service/image_service.go`
- `backend/internal/service/url_upload.go`（删除）
- `backend/internal/http/handler/url_upload_handler_test.go`（删除）
- `backend/internal/util/url_safety.go`（删除）
- `backend/internal/util/url_safety_test.go`（删除）
- `frontend/src/lib/api.ts`
- `frontend/src/routes/+page.svelte`
- `docs/tasks/url-upload-ssrf-protection.md`
- `docs/status/progress-report.md`
- `.trellis/tasks/05-22-remediation-extension-implementation/design.md`
- `.trellis/tasks/05-22-remediation-extension-implementation/implement.md`

### 验证记录

- `cd backend && gofmt -w ./cmd ./internal && go test ./...`：通过，16 个 package / 242 项后端测试通过。
- `cd frontend && npm run lint`：通过，无 ESLint 问题。
- `cd frontend && npm run typecheck`：通过，0 errors / 0 warnings。
- `cd frontend && npm run test`：通过，10 个测试文件 / 55 项测试通过。
- `cd frontend && npm run build:backend`：通过，前端构建并复制到 `backend/web/`。

### 复核结论

- 需求已更新为“前端下载图片然后上传到后端”，当前实现与需求一致。
- 代码侧已确认后端不再存在 URL 下载接口、远端抓取服务或 URL 安全抓取工具；前端不再调用 `/v1/image/url`。
