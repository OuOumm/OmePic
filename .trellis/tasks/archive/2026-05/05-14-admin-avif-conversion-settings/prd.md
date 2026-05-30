# Admin configurable AVIF conversion settings

## Goal

允许管理员在后台运行设置中调整上传图片转换为 AVIF 时使用的质量和速度参数，让部署者可以在画质、文件大小和转换耗时之间按需取舍。

## Confirmed Facts

- 当前 AVIF 转换在 `backend/internal/service/image_transform.go` 中硬编码：`Quality: 60`、`Speed: 8`。
- 使用的 `github.com/gen2brain/avif` 版本中：
  - `Quality` 范围为 `0..100`，`100` 表示 lossless，默认值为 `60`。
  - `Speed` 范围为 `0..10`，值越小通常更慢但压缩/质量表现更好，库默认值为 `10`。
- 上传流程会先检查原始字节 MD5 去重；只有新物理文件需要转换 AVIF。
- Runtime settings 已通过 SQLite `config` 表持久化，并由后台 `GET|PUT /admin/system-settings` 编辑。
- 前端后台设置页已使用 `system.runtime` 编辑站点、上传策略、公开 URL、维护和密码相关设置。

## Requirements

- 在 SQLite runtime settings 中新增 AVIF 转换配置：
  - `avif_quality`：整数，范围 `0..100`，默认值 `60` 写入初始化/默认 runtime config。
  - `avif_speed`：整数，范围 `0..10`，默认值 `8` 写入初始化/默认 runtime config。
- 不在 AVIF 转换层另设质量/速度硬编码兜底；转换参数必须来自 runtime settings / 初始化数据库默认配置。
- 程序首次运行或缺失 key 时必须写入上述默认值，且不得覆盖已有配置。
- `PUT /admin/system-settings` 必须校验 AVIF 参数范围；越界或非合理值返回 `invalid_input`，不保存部分配置。
- 上传新图片时，AVIF 转换必须使用当前 runtime settings 的 `avif_quality` / `avif_speed`。
- 去重命中的重复上传不重新转换历史文件；更改 AVIF 参数只影响之后新写入的物理 AVIF 对象。
- 后台设置页需要新增质量、速度输入控件，并随现有保存按钮一起提交。
- TypeScript 类型、API 测试、后端服务测试和文档/spec 需同步更新。

## Acceptance Criteria

- [ ] `RuntimeSettings` / `RuntimeSettingsUpdateInput` / 前端 `RuntimeSettings` 类型包含 `avif_quality` 与 `avif_speed`。
- [ ] 默认 runtime config 持久化包含 `avif_quality=60`、`avif_speed=8`，且不会覆盖已有值。
- [ ] 后端拒绝 `avif_quality < 0`、`avif_quality > 100`、`avif_speed < 0`、`avif_speed > 10`。
- [ ] 新上传的非重复图片转换 AVIF 时使用管理员配置的质量和速度。
- [ ] 重复上传仍按现有去重流程复用已有物理文件，不因设置变化触发重转。
- [ ] 后台设置页可编辑 AVIF 质量/速度并保存。
- [ ] `cd backend && go test ./...` 通过；前端改动后 `cd frontend && npm run typecheck` 与相关测试通过。

## Out of Scope

- 不批量重转已上传图片。
- 不新增每个存储实例独立的 AVIF 参数。
- 不修改输出格式，仍统一输出 `image/avif` 与 `.avif` URL。
