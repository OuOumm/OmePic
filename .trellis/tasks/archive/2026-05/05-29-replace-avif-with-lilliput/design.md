# 技术设计：替换为 discord/lilliput

## 架构对比

### 当前（gen2brain/avif）

```
io.Reader → image.Decode() → avif.Encode() → io.Pipe → storage.SaveStream()
                ↑ Go标准库              ↑ WASM编码               ↑ 流式上传
```

- **优点**：纯 Go + 流式，内存友好
- **缺点**：不支持动画编码，WASM 性能一般

### 新方案（discord/lilliput）

```
io.Reader → ioutil.ReadAll() → lilliput.Transform() → io.Writer
                ↑ 全量读入             ↑ C原生编解码           ↑ 写入结果
```

- **优点**：支持动画，C 原生高性能
- **缺点**：全量内存缓冲（上传限制 20MB 内可接受），CGO 依赖

## 模块设计

### 1. 新增文件：`backend/internal/service/image_transform_lilliput.go`

```go
// 替换 image_transform.go 中的 encodeAVIFToWriter
// 保持签名：func(io.Reader, io.Writer, AVIFConversionSettings) error

func encodeAVIFToWriterLilliput(source io.Reader, target io.Writer, settings AVIFConversionSettings) error {
    // 1. 读取全部源数据
    data, _ := io.ReadAll(source)
    
    // 2. 创建 lilliput decoder
    decoder, _ := lilliput.NewDecoder(data)
    defer decoder.Close()
    
    // 3. 检测动图，检查帧数
    frameCount := countFrames(decoder)
    if frameCount > maxAnimationFrames {
        return ErrTooManyFrames
    }
    
    // 4. 创建 ImageOps + Transform
    ops := lilliput.NewImageOps(0) // 0 = no resize
    defer ops.Close()
    
    opts := &lilliput.ImageOptions{
        FileType:             ".avif",
        Width:                0,              // 不缩放
        Height:               0,
        ResizeMethod:         lilliput.ImageOpsNoResize,
        NormalizeOrientation: true,
        EncodeOptions: map[int]int{
            lilliput.AvifQuality: settings.Quality,
            lilliput.AvifSpeed:   settings.Speed,
        },
    }
    
    // 5. 输出
    buf := make([]byte, len(data)*2) // 预估输出 ≤ 源数据 2 倍
    result, err := ops.Transform(decoder, opts, buf)
    target.Write(result)
}
```

### 2. 帧数检测

lilliput 的 Decoder 逐帧解码：
```go
func countFrames(decoder lilliput.Decoder, maxFrames int) int {
    h, _ := decoder.Header()
    if !isAnimated(h) { return 1 } // 静态图
    
    fb := lilliput.NewFramebuffer(h.Width(), h.Height())
    defer fb.Close()
    
    count := 0
    for count < maxFrames+1 {
        err := decoder.DecodeTo(fb)
        if err == io.EOF { break }
        count++
    }
    return count
}
```

**注意**：帧数检测会消耗 decoder 状态，需要重新创建 decoder 做实际转换。或者先检测帧数再决定是否创建新的 decoder。

### 3. 动图检测策略

lilliput 提供了 `Duration()` 方法：
- 静态图：`Duration() == 0`
- 动图 GIF：`Duration() > 0`（但可能不准确，取决于实现）
- 动图 WebP/AVIF：`Duration() > 0`

但更可靠的方式是检查帧数 > 1。

### 4. ImageService 集成

```go
// image_service.go 构造函数中
encoder: encodeAVIFToWriterLilliput,  // 替换原来的 encodeAVIFToWriter
```

保留 `ImageService.encoder` 字段类型不变，只需替换赋值。

保留 `avifStreamConversion` 流式管道机制——虽然 lilliput 内部不流式，但对外接口保持一致，storage 层仍可以流式上传。

### 5. 并发安全

- lilliput `ImageOps` 实例**不可跨 goroutine 使用**（C 库内部状态）
- 当前 `avifLimiter`（channel semaphore）控制并发度，每个 goroutine 内创建独立的 lilliput 对象
- 无需改为对象池（首次实现保持简单，后续优化）

## Docker 变更

### C 依赖（Alpine）

**构建阶段**（golang:1.25-alpine）：
```dockerfile
RUN apk add --no-cache \
    libavif-dev libwebp-dev libjpeg-turbo-dev \
    libpng-dev giflib-dev pkgconf build-base
```

**运行阶段**（alpine:3.21）：
```dockerfile
RUN apk add --no-cache ca-certificates \
    libavif libwebp libjpeg-turbo libpng giflib
```

### 构建参数

```dockerfile
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -o /server ./cmd/server/
```

- 无需 `-tags nodynamic`（lilliput 不走 gen2brain/avif 路径）
- CGO_ENABLED=1 必须
- CGO_LDFLAGS 由 pkgconf 自动处理

## 涉及文件

| 文件 | 变更 |
|------|------|
| `backend/internal/service/image_transform.go` | 删除或保留为 fallback |
| `backend/internal/service/image_transform_lilliput.go` | **新增**：lilliput 编码实现 |
| `backend/internal/service/image_service.go` | 修改构造函数中 encoder 赋值 + 新增注入点 |
| `backend/internal/service/image_service_test.go` | 适配测试：avif.Decode → lilliput 解码验证 |
| `backend/go.mod` | 添加 `github.com/discord/lilliput` 依赖 |
| `backend/go.sum` | 自动更新 |
| `Dockerfile` | C 依赖 + CGO_ENABLED=1，基础镜像 alpine |
| `docker-compose.yml` | 不变（镜像结构兼容） |

## 风险点

| 风险 | 缓解 |
|------|------|
| lilliput C 库与 Alpine 版本不兼容 | 锁定 Alpine 3.21，lilliput 已支持 |
| 帧数检测消耗大（解码两遍） | 首帧检测无开销，动图才需全量解码；300 帧上限可接受 |
| 输出 buffer 预估不足 | 初始源文件 2 倍，不足时 lilliput 返回 `io.ErrShortBuffer`，重试 |
| 内存翻倍（全量读入 vs 流式） | 上传上限 20MB，内存占用 ~40MB，可接受 |
| lilliput API 变更 | 锁定版本 |

## 回滚方案

保留 `image_transform.go` 不改动，新增文件 `_lilliput.go`。如需回滚：
1. 将构造函数 encoder 恢复为 `encodeAVIFToWriter`
2. 删除 lilliput 依赖重跑 `go mod tidy`
3. Docker 切换回 `CGO_ENABLED=0` + `FROM scratch`