# 替换 gen2brain/avif 为 discord/lilliput 实现高性能图片处理

## Goal

用 `discord/lilliput` 替换 `gen2brain/avif`，获得：
- **GIF 转动画 AVIF**：保留多帧动图
- **C 原生性能**：比 WASM 解释执行更快
- **更低内存占用**：lilliput 专为高吞吐设计

## 现状分析

| 项目 | 当前值 |
|------|--------|
| 图片处理库 | `gen2brain/avif` v0.4.4（wazero WASM 解码） |
| 输出格式 | AVIF（后缀 `.avif`，MIME `image/avif`） |
| Docker 基础镜像 | `scratch`（~18 MB） |
| CGO 依赖 | 零（`CGO_ENABLED=0`） |
| 编码入口 | `image_transform.go:encodeAVIFToWriter` |
| 编码流程 | `io.Pipe()` 流式：`image.Decode` → `avif.Encode` |
| 并发控制 | `avifLimiter`（channel semaphore） |
| 编码参数 | `Quality`(0-100), `Speed`(0-10) 通过 `AVIFConversionSettings` 传递 |
| GIF 行为 | 接受上传，但只取第一帧转静态 AVIF |

## Requirements

### R1: 替换 AVIF 编码器
- 将 `ImageService.encoder` 从 `encodeAVIFToWriter` 替换为基于 lilliput 的实现
- 保持调用签名不变：`func(io.Reader, io.Writer, AVIFConversionSettings) error`
- 映射编码参数：`Quality` → `lilliput.AvifQuality`，`Speed` → `lilliput.AvifSpeed`

### R2: GIF 动图保留（核心需求）
- 动画 GIF 上传后输出动画 AVIF（保留所有帧）
- 静态图片行为不变（输出静态 AVIF）
- 上传时自动检测图片是否为动图，无需用户指定
- **帧数上限 300 帧**：超过上限直接拒接上传（用户友好提示）

### R3: Docker 镜像适配
- 从 `scratch` 切换为 `alpine` 基础镜像
- 安装 lilliput 所需的 C 运行时库（libavif、libwebp 等）
- 构建阶段安装 lilliput 依赖的 C 开发头文件

### R4: 兼容性
- 保持现有配置项不变：`AVIFMaxConcurrency`、`AVIFConversionTimeoutSeconds`、`AvifQuality`、`AvifSpeed`
- 保持 `publicImageExtension = ".avif"` 和 `publicImageMIMEType = "image/avif"`
- 现有测试全部通过（或适配 lilliput 行为差异）

## Acceptance Criteria

- [ ] GIF 上传后生成动画 AVIF 文件（多帧保留）
- [ ] JPEG/PNG/WebP/BMP/静态AVIF 上传行为不变
- [ ] `go test ./...` 全部通过
- [ ] Docker 构建成功，镜像能正常启动
- [ ] `go vet ./...` 无新增警告

## Out of Scope

- 调整输出格式（仍然统一输出 AVIF）
- WebP 动画输出（保持 AVIF）
- 图片缩放/裁剪功能
- 前端改动
- 动画帧数/时长限制（后续任务）

## Open Questions

1. **是否需要保留原始 GIF 不转码的选项？** 当前不保留，所有图片统一转 AVIF。