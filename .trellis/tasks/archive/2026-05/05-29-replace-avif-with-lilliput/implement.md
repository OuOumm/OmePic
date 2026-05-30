# 实施计划：替换为 discord/lilliput

## 执行顺序

### Step 1: 添加 lilliput 依赖
```bash
cd backend && go get github.com/discord/lilliput@latest
```
- 验证：`go mod tidy && go build ./...`

### Step 2: 新增 `image_transform_lilliput.go`
- 实现 `encodeAVIFToWriterLilliput` 函数
- 实现帧数检测 `countFrames`
- 常量 `maxAnimationFrames = 300`
- 新增错误类型 `ErrTooManyFrames`
- 验证：`go build ./internal/service/`

### Step 3: 修改 `image_service.go`
- 构造函数中将 `encoder: encodeAVIFToWriter` 改为 `encoder: encodeAVIFToWriterLilliput`
- 注入 `lilliput` 全局初始化（如果有的话）
- 验证：`go build ./...`

### Step 4: 删除/清理旧代码
- 保留 `image_transform.go` 不动（回滚用）
- 删除 `gen2brain/avif` 依赖：`go mod tidy`
- 验证：`go build ./... && go vet ./...`

### Step 5: 适配测试
- `backend/internal/service/image_service_test.go`：avif.Decode 验证 → lilliput 或标准库解码
- `backend/internal/http/handler/image_handler_test.go`：同适配
- 新增 GIF 动画测试用例
- 新增帧数上限测试
- 验证：`go test -race ./...`

### Step 6: 更新 Dockerfile
- 构建阶段添加 C 开发库
- 运行阶段安装 C 运行时库
- `CGO_ENABLED=1`，移除 `-tags nodynamic`
- 最终基础镜像从 `scratch` 改为 `alpine:3.21`
- 验证：本地 `docker build -t omepic .`

### Step 7: 集成验证
```bash
go test -race ./...     # 全部测试
go vet ./...            # 静态检查
docker build -t omepic . # Docker 构建
```

## 验证命令

| 步骤 | 命令 |
|------|------|
| 编译 | `go build ./backend/...` |
| 测试 | `go test -race ./backend/...` |
| 静态检查 | `go vet ./backend/...` |
| Docker | `docker build -t omepic . && docker run --rm omepic` |

## 回滚点

每个 Step 都是独立的回滚点。Step 2-3 之间可以通过切换 encoder 赋值回滚。