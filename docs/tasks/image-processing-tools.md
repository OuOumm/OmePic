# 任务：图片处理工具

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F7** — "图片基础处理"
- **§4.4 P2：图片处理工具** — "缩放、裁剪、水印、EXIF 清理、预设模板"

## 预估人日

8 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/service/image_processor.go`（新增）— 图片处理引擎
- `backend/internal/http/handler/image_transform.go`（新增）— 处理 API
- `backend/internal/http/router/router.go` — 注册处理路由
- `backend/internal/cache/redis_cache.go` — 处理结果缓存
- 运行时配置 — 处理参数限制

### 关键实现点

1. URL 参数化处理：`/i/:uid?w=800&h=600&fit=crop&watermark=text`。
2. 支持操作：缩放（width/height/fit）、裁剪（crop）、旋转（rotate）、水印（text/image）、EXIF 清理。
3. 预设模板：`thumbnail`、`medium`、`social`（常用社交平台尺寸）。
4. 处理结果缓存：相同参数的请求复用已处理结果。
5. 安全限制：最大输出尺寸、最大处理时间、防 DoS。
6. EXIF 清理：默认移除 GPS 和相机信息（可配置保留）。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] 图片处理 API 可用

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | `/i/:uid?w=800` 缩放 | 返回缩放后的图片 |
| 正向 | `/i/:uid?w=200&h=200&fit=crop` | 返回裁剪后的正方形 |
| 正向 | `/i/:uid?watermark=test` | 返回带水印的图片 |
| 异常 | 参数超出限制 | 返回 400 错误 |
| 边界 | 相同参数多次请求 | 命中缓存 |
| 边界 | 处理超时 | 返回超时错误 |

### 必须覆盖的测试类型

- 单元测试：各处理操作的正确性
- 集成测试：上传→处理→缓存
- 安全测试：超大参数防护

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 处理参数的安全限制
- 缓存键设计（参数组合的唯一性）
- 内存使用（大图片处理）

### 安全检查

- 防止图片炸弹（限制输出尺寸）
- 参数注入防护

### 文档更新要求

- API 文档说明处理参数

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
