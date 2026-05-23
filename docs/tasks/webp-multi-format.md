# 任务：多格式输出 WebP

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F1** — "WebP / JPEG-XL / 原格式输出策略"
- **§2.1 F1 多格式输出策略**
- **§4.3 P1：多格式输出 WebP** — "derivative 模型、WebP 编码、固定格式 URL、缓存策略"

## 预估人日

7 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/image.go` — derivative 数据模型
- `backend/internal/service/transform_service.go`（新增）— 多格式转换服务
- `backend/internal/http/handler/image.go` — 支持格式协商和固定格式 URL
- `backend/internal/storage/provider.go` — 存储路径扩展
- `backend/internal/cache/redis_cache.go` — 缓存键设计
- `backend/internal/database/migration.go` — derivative 表迁移
- `backend/cmd/server/main.go` — WebP Worker 集成
- `frontend/src/lib/types/index.ts` — 前端类型更新

### 关键实现点

1. 存储模型：一张图片记录对应多个 derivative：`original`、`avif`、`webp`。
2. 新增 `image_derivatives` 表：`id`、`image_uid`、`format`、`storage_key`、`file_path`、`size_bytes`、`status`。
3. 访问策略：
   - 固定格式 URL：`/i/:uid.avif`、`/i/:uid.webp`。
   - 自动协商 URL：`/i/:uid` 根据 `Accept` 头返回最佳格式。
4. WebP 编码参数可配置（质量、速度）。
5. 缓存键设计：不同格式使用不同缓存键；CDN `Vary: Accept` 头。
6. 老链接兼容：原有 `/i/:uid` 仍返回 AVIF（向后兼容）。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 访问 `/i/:uid.avif` | 返回 AVIF 格式 |
| 正向 | 访问 `/i/:uid.webp` | 返回 WebP 格式 |
| 正向 | 访问 `/i/:uid`（Accept: image/webp） | 返回 WebP |
| 正向 | 访问 `/i/:uid`（Accept: image/avif） | 返回 AVIF |
| 正向 | 老链接 `/i/:uid` 无 Accept 头 | 返回 AVIF（兼容） |
| 异常 | 请求不支持的格式 | 返回最佳可用格式或 406 |
| 边界 | WebP 转换失败 | 回退到 AVIF |
| 边界 | 仅 AVIF 可用时请求 WebP | 返回 AVIF 或 404 |

### 必须覆盖的测试类型

- 单元测试：格式协商逻辑、Accept 头解析
- 集成测试：端到端上传→转换→多格式访问
- 缓存测试：不同格式缓存键隔离

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- Accept 头解析是否覆盖所有浏览器常见值
- 缓存键是否正确隔离不同格式
- derivative 状态机是否完整

### 安全检查

- 格式协商不引入路径遍历风险

### 文档更新要求

- API 文档说明格式 URL 和协商行为
- CDN 配置文档说明 Vary 头

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
