# 任务：URL 上传 SSRF 防护

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.3 安全短板与整改建议** — "URL 上传可能引入 SSRF"
- **§4.2 P0：URL 上传 SSRF 防护** — "禁止私网地址、限制跳转、下载大小、超时、DNS 重绑定防护"

## 预估人日

3 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/http/handler/upload_url.go`（新增）— `POST /v1/image/url` 处理器
- `backend/internal/util/url_safety.go`（新增）— URL 安全校验工具
- `backend/internal/http/router/router.go` — 注册新路由
- `frontend/src/lib/api.ts` — URL 上传改调后端安全接口
- `frontend/src/routes/+page.svelte` — URL 上传逻辑调整

### 关键实现点

1. 新增后端 `POST /v1/image/url`，前端 URL 上传改为调用后端，由后端统一复用上传策略。
2. URL 抓取保护：仅允许 `http` / `https` 协议。
3. 限制跳转次数（默认 5 次），每次跳转重新校验目标地址。
4. DNS 解析后拒绝私网（10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16）、环回（127.0.0.0/8, ::1）、链路本地（169.254.0.0/16, fe80::/10）、组播、未指定、CGNAT（100.64.0.0/10）地址。
5. 下载超时、Content-Length 限制和读取上限使用运行时最大上传大小。
6. 下载结果进入 `ImageService.Upload`，继续执行 Token、MIME、像素、去重和存储校验。
7. 使用 `http.Client` 自定义 `CheckRedirect` 和 `DialContext` 实现 DNS 后检查。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./internal/util/... ./internal/http/handler/...`
- [ ] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [ ] 前端 URL 上传切换到后端接口：`cd frontend && npm run build:backend`

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 从公网 URL 上传图片 | 正常下载并完成上传 |
| 正向 | URL 重定向到另一个公网图片 | 跟随重定向成功 |
| 异常 | URL 为 http://127.0.0.1/... | 拒绝，返回 SSRF 防护错误 |
| 异常 | URL 为 http://169.254.169.254/...（云元数据） | 拒绝 |
| 异常 | URL 为 http://192.168.1.1/...（内网） | 拒绝 |
| 异常 | URL 为 http://10.0.0.1/...（内网） | 拒绝 |
| 异常 | URL 重定向到内网地址 | 拒绝（跳转时重新检查） |
| 异常 | 下载超时（远端不响应） | 返回超时错误 |
| 异常 | Content-Length 超过限制 | 拒绝，提示文件过大 |
| 边界 | 跳转恰好 5 次 | 成功 |
| 边界 | 跳转第 6 次 | 拒绝，提示重定向过多 |
| 边界 | ftp:// 协议 URL | 拒绝，仅允许 http/https |

### 必须覆盖的测试类型

- 单元测试：URL 安全校验工具（IP 解析、私网判断、协议检查）
- 集成测试：完整 URL 上传链路（mock HTTP 服务器模拟各种场景）
- 安全测试：DNS rebinding 模拟、SSRF bypass 尝试

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 前端 URL 上传功能正常调用后端接口

## 复查

### 代码审查关注点

- DNS 解析后地址检查是否在每次拨号时执行（防止 DNS rebinding）
- 跳转次数限制是否正确实现
- 大文件下载是否有流式处理，避免内存溢出
- 错误消息不暴露内部网络拓扑

### 安全检查

- 所有私网/特殊地址段是否完整覆盖（IPv4 + IPv6）
- 是否存在 TOCTOU 竞态（DNS 解析与连接建立之间）
- 下载的文件是否经过与普通上传相同的完整性校验

### 文档更新要求

- 新增 API 文档：`POST /v1/image/url` 接口说明
- 安全文档说明 SSRF 防护策略

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
