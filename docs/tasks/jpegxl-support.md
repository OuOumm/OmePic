# 任务：JPEG-XL 支持

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F1** — "WebP / JPEG-XL / 原格式输出策略"
- **§4.4 P2：JPEG-XL 支持** — "编码库评估、兼容性开关、CDN 缓存策略"

## 预估人日

5 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/service/transform_service.go` — 新增 JXL 编码支持
- `backend/internal/model/image.go` — derivative 模型支持 JXL
- `backend/internal/http/handler/image.go` — 支持 `.jxl` 格式 URL
- `backend/internal/cache/redis_cache.go` — JXL 缓存键
- 运行时配置 — JXL 编码开关与参数

### 关键实现点

1. 编码库评估：选择 Go 原生或 CGO 绑定的 JPEG-XL 编码器。
2. 兼容性开关：运行时配置 `jxl_enabled`，默认关闭。
3. derivative 扩展：新增 `jxl` 格式。
4. 固定格式 URL：`/i/:uid.jxl`。
5. Accept 头协商：`image/jxl` 优先级可配置。
6. CDN 缓存策略：`Vary: Accept` 头需包含 `jxl`。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] JXL 格式图片可正确编码和访问

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 启用 JXL 后上传图片 | 生成 JXL derivative |
| 正向 | 访问 `/i/:uid.jxl` | 返回 JXL 格式 |
| 异常 | JXL 编码失败 | 回退到 AVIF |
| 边界 | JXL 关闭时访问 `.jxl` | 返回 404 或回退 |

### 必须覆盖的测试类型

- 单元测试：JXL 编码逻辑
- 集成测试：上传→JXL 转换→访问

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- JXL 编码器的性能和内存使用
- 编码器依赖是否可选（CGO 依赖）

### 安全检查

- JXL 解码不引入新的安全风险

### 文档更新要求

- 运行时配置文档说明 JXL 选项

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
