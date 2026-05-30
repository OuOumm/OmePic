# 整改与扩展任务拆分与实施设计

## 1. 边界

- 源设计文档：`docs/debug/remediation-extension-design/rectification-and-extension-design.md`。
- 任务文档输出目录：`docs/tasks/remediation-extension/`。
- 本轮代码实现：仅 P0 安全与稳定性基线。
- P1/P2：仅输出后续任务文档，不实现代码。
- 不修改与本任务无关的前端错误页或既有未提交变更。

## 2. 任务文档结构

```text
docs/tasks/remediation-extension/
├── README.md                     # 总览、优先级、依赖图、总完成报告
├── p0/
│   ├── 01-upload-resource-guard.md
│   ├── 02-file-authenticity-validation.md
│   ├── 03-url-upload-ssrf-protection.md
│   ├── 04-token-governance.md
│   ├── 05-default-password-hardening.md
│   ├── 06-soft-delete-recycle-bin.md
│   ├── 07-sqlite-core-indexes.md
│   ├── 08-config-audit-log.md
│   └── 09-storage-health-check.md
├── p1/
│   └── *.md
└── p2/
    └── *.md
```

每个子任务文档包含：目标、范围、依赖、实施要点、验收标准、测试/复核要求、完成报告模板。本轮完成的 P0 文档会填充完成报告；P1/P2 保留模板。

## 3. P0 技术设计

### 3.1 上传资源保护

- 在运行时设置中增加上传保护字段：
  - `max_image_pixels`：原图像素上限，默认 40,000,000。
  - `avif_max_concurrency`：同步 AVIF 转换并发上限，默认 2。
  - `avif_conversion_timeout_seconds`：单次转换超时，默认 30 秒。
- `ImageService` 持有可调整信号量；运行时策略快照控制并发和超时。
- 转换前通过 `image.DecodeConfig` 校验尺寸，拒绝超限图片。
- 转换使用 context 超时；超时返回 `dependency_unavailable` 安全错误。

### 3.2 文件真实性校验

- 上传源准备完成后进行统一校验：
  - 读取文件头和完整可复读源。
  - 使用 magic bytes / `http.DetectContentType` 辅助识别。
  - 使用 `image.DecodeConfig` 验证真实图片格式与尺寸。
  - 将解码格式映射为 MIME，与请求 MIME 交叉校验。
- 上传允许列表仍以运行时 `allowed_mime_types` 为准，但判断对象改为真实 MIME。

### 3.3 URL 上传（已调整为前端下载）

- 2026-05-25 用户要求：URL 上传改为前端下载图片后上传到后端，不再由后端下载远端 URL。
- 移除后端 `POST /v1/image/url` 与远端抓取逻辑，消除后端 SSRF 攻击面。
- 前端 URL 上传流程：校验 `http` / `https` URL，浏览器 `fetch` 获取图片 Blob，转换为 `File`，再复用现有 `POST /v1/image` multipart 上传链路。
- 后端继续在标准文件上传链路执行 Token、真实 MIME、像素、大小、去重和存储校验。

### 3.4 Token 治理

- 新增 Token 治理表：
  - `token_usage`：按 token hash 聚合上传次数、大小、最近 IP、最近使用时间、预览值。
  - `token_controls`：记录禁用状态与原因。
- 上传前检查 Token 是否禁用；上传成功后记录使用统计。
- 新增管理员 API：列出 Token、禁用 Token、恢复 Token。

### 3.5 默认密码安全改造

- 默认密码引导仍保持兼容，但记录 `admin_password_uses_default=true`。
- 修改密码后写入 `admin_password_uses_default=false`。
- `GET /admin/system-settings` 的 `readonly.security.admin_password.using_default` 反映真实状态。
- 前端现有安全警告区域展示默认密码高危状态。

### 3.6 删除生命周期（已调整）

- 2026-05-25 用户确认不需要回收站功能。
- 不再新增删除元数据字段，不提供回收站列表或恢复 API。
- 删除操作直接删除数据库记录，并继续修复 UID / MD5 Redis 映射。

### 3.7 SQLite 核心索引

- 补齐复合索引：
  - `storage_key, md5_hash`
  - `token, created_at DESC`
  - `ip_address, created_at DESC`
  - `storage_key, created_at DESC`
- 迁移保持幂等。

### 3.8 配置审计日志

- 新增 `config_audit_logs` 表。
- runtime / storage 配置变更写入 before/after JSON 快照。
- 快照中的 secret 使用现有遮罩策略。
- 新增管理员 API 查询审计日志。

### 3.9 存储健康检查

- 新增 `storage_health_checks` 表。
- 管理员 API 支持手动检测单个/全部存储。
- 启动后台定时心跳，周期写入最新状态。
- 探测使用小对象写入、读取、删除，记录状态、延迟和错误。

## 4. 兼容与迁移

- SQLite 迁移通过 `ALTER TABLE ADD COLUMN` 与 `CREATE TABLE IF NOT EXISTS` 实现，保持幂等。
- 新增能力不破坏现有 `POST /v1/image` 协议。
- URL 上传由前端下载后复用标准文件上传；若浏览器下载或后端文件上传失败，显示已有错误提示。
- 图片删除后旧物理文件仍不主动删除，符合现有“请求路径不物理删除”约定。
- Token hash 与审计日志不会在公开 API 暴露秘密。

## 5. 运行与回滚

- 回滚代码后新增列/表可留存，不影响旧代码读取基础列。
- URL 上传不再保留后端抓取 API；如远端不允许浏览器跨域读取，URL 上传会按浏览器/CORS 策略失败。
- 存储健康后台任务失败仅记录日志，不阻断主服务。

## 6. 验证策略

- 后端：`cd backend && go test ./...`。
- 前端：`cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend`。
- 格式：`cd backend && gofmt -w <changed-go-files>` 后确认 `gofmt -l` 无输出。
- 文档：检查 `docs/tasks/remediation-extension/` 全量任务文档与 P0 完成报告。
