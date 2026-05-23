# 任务：上传资源保护

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.2 性能瓶颈与整改建议** — "AVIF 转换在上传请求内同步完成""编码参数全局生效，缺少资源保护"
- **§4.2 P0：上传资源保护** — "AVIF 转换并发上限、任务超时、最大像素数限制、极端参数保护"

## 预估人日

2 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/runtime_setting.go` — 新增运行时保护字段模型
- `backend/internal/repository/setting_repo.go` — 读写新增设置字段
- `backend/internal/service/image_service.go` — 增加信号量、超时控制
- `backend/internal/http/handler/admin_setting.go` — 暴露保护字段给管理后台
- `backend/internal/http/handler/upload.go` — 上传前校验像素上限
- `frontend/src/lib/types/index.ts` — 前端类型新增保护字段
- `frontend/src/routes/admin/dashboard/settings/+page.svelte` — 设置页展示保护字段

### 关键实现点

1. 运行时设置新增字段：`max_image_pixels`（默认 40,000,000）、`avif_max_concurrency`（默认 2）、`avif_conversion_timeout_seconds`（默认 30s）。
2. `ImageService` 持有可调整信号量；运行时策略快照控制并发和超时。
3. 上传流程在 AVIF 转换前通过 `image.DecodeConfig` 校验原图像素数，超过 `max_image_pixels` 返回 413 错误。
4. AVIF 转换使用 `context.WithTimeout`；超时返回 `dependency_unavailable` 安全错误，避免泄漏内部信息。
5. 管理后台设置页可查看/编辑保护字段，包含安全档位提示。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./internal/service/...`
- [ ] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出
- [ ] 前端类型与 UI 可展示新增字段：`cd frontend && npm run typecheck`

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传像素数 < max_image_pixels 的图片 | 正常上传并转换成功 |
| 正向 | 设置 avif_max_concurrency=2 并发上传 | 信号量生效，最多 2 个并发转换 |
| 异常 | 上传像素数超过 max_image_pixels | 返回 413 错误，明确提示图片过大 |
| 异常 | AVIF 转换超过 timeout | 返回 dependency_unavailable，不泄漏内部信息 |
| 边界 | max_image_pixels 设为 0 | 使用默认值或拒绝配置（文档记录行为） |
| 边界 | avif_max_concurrency 设为 1 | 严格串行转换 |

### 必须覆盖的测试类型

- 单元测试：像素校验逻辑、信号量获取/释放、超时逻辑
- 集成测试：运行时设置变更后上传行为变化
- 本地手动验证：修改后台设置 → 上传测试图片 → 观察行为

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 手动验证上述场景均符合预期

## 复查

### 代码审查关注点

- 信号量初始化与运行时热更新的线程安全
- 超时 context 是否正确传播到存储写入层
- 错误返回是否统一使用项目错误码体系
- 像素校验是否覆盖 BMP、PNG、JPEG、GIF 等格式

### 安全检查

- 超时/并发/像素限制默认值是否合理（防 DoS）
- 错误消息不暴露内部路径或参数

### 文档更新要求

- 运行时设置字段文档说明
- API 错误码新增 413 / dependency_unavailable 说明

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
