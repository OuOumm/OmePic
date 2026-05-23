# 任务：文件真实性校验

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.3 安全短板与整改建议** — "MIME 校验可能只依赖请求头"
- **§4.2 P0：文件真实性校验** — "magic bytes、图片解码校验、MIME 与扩展名交叉验证"

## 预估人日

2 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/service/image_service.go` — 统一真实 MIME 校验流程
- `backend/internal/util/mime_detect.go`（新增）— magic bytes 检测 + `http.DetectContentType` 辅助
- `backend/internal/http/handler/upload.go` — 上传源准备完成后调用统一校验

### 关键实现点

1. 上传源准备完成后进行统一校验：读取文件头，使用 magic bytes / `http.DetectContentType` 辅助识别。
2. 使用 `image.DecodeConfig` 验证真实图片格式与尺寸，将解码格式映射为 MIME。
3. 将真实 MIME 与请求 MIME 交叉校验；不匹配时拒绝上传。
4. 上传允许列表仍以运行时 `allowed_mime_types` 为准，但判断对象改为真实 MIME。
5. 拒绝无法解码为图片的文件（如伪装为 PNG 的可执行文件）。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./internal/service/... ./internal/util/...`
- [ ] gofmt 格式正确：`cd backend && gofmt -l ./cmd ./internal` 无输出

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 上传真实 JPEG，Content-Type 为 image/jpeg | 校验通过 |
| 正向 | 上传真实 PNG，Content-Type 为 image/png | 校验通过 |
| 异常 | 修改扩展名为 .jpg 但文件实际为 exe | 拒绝，提示文件类型不匹配 |
| 异常 | 伪造 Content-Type 为 image/png 但文件头为 GIF | 拒绝或接受为 GIF（取决于策略） |
| 边界 | 空文件 / 0 字节 | 拒绝，提示文件为空 |
| 边界 | 超小有效图片（1x1 像素） | 校验通过 |
| 边界 | 损坏的图片文件（有效头但截断数据） | 拒绝，提示文件损坏 |

### 必须覆盖的测试类型

- 单元测试：各种格式的 magic bytes 检测、MIME 映射、交叉校验
- 集成测试：通过 API 上传各种文件验证端到端校验
- 安全测试：恶意文件伪装上传

### 测试通过标准

- `cd backend && go test ./...` 全部通过
- 手动上传伪装文件均被正确拒绝

## 复查

### 代码审查关注点

- magic bytes 检测是否覆盖所有允许的图片格式
- 真实 MIME 与请求 MIME 的交叉校验逻辑是否严密
- 文件头读取是否使用 `io.LimitReader` 防止大文件内存问题
- 错误提示是否对用户友好

### 安全检查

- 防止 zip bomb / 解析炸弹（限制解码维度和帧数）
- 确保不信任任何客户端提供的元数据

### 文档更新要求

- 上传 API 文档说明 MIME 校验行为变更
- 允许的 MIME 类型列表文档

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
