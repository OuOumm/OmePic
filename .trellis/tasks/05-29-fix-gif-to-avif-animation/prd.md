# 修复 GIF 转 AVIF 后丢失动画

## Goal

修复 GIF 上传后转换为 AVIF 仍然变成静态图的问题，确保 Linux/Docker 环境下使用 lilliput 时能输出动画 AVIF。

## Confirmed Facts

- 当前 Windows 本地开发环境通过平台 build tag 回退到 `gen2brain/avif`，该路径只支持单帧编码，因此 Windows 本地 `go run` 下 GIF 转 AVIF 仍会是静态图。
- Linux/macOS 构建路径使用 `discord/lilliput`。
- lilliput 的 `ImageOptions.MaxEncodeFrames` 默认值 `0` 表示无限制，不适合作为 300 帧上限。
- lilliput 的 `ImageOptions.DisableAnimatedOutput` 默认 false，理论上应支持多帧输出。

## Requirements

- Linux/Docker 环境下 GIF 转 AVIF 必须保留动画。
- Windows 本地环境不能再悄悄走静态 fallback 导致误判；应明确限制或提示。
- 保持现有上传接口、输出后缀 `.avif`、MIME `image/avif` 不变。
- 保持最大动画帧数 300 帧限制。

## Acceptance Criteria

- [ ] Docker/Linux 构建路径使用 lilliput 动画编码。
- [ ] GIF 转 AVIF 的 lilliput 配置显式启用动画输出并限制最多 300 帧。
- [ ] Windows fallback 行为有明确说明，避免误以为 Windows 本地支持 GIF 动画转码。
- [ ] `GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build ./...` 通过。
- [ ] Windows 本地 `go build ./...` 通过。
