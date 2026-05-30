# 补充 API 页面接口示例

## Goal

让前端 `/api` 页面不再只展示零散 curl 片段，而是为页面中展示的接口补充更完整的“如何调用 + 响应示例”，方便开发者直接照抄联调。

## Confirmed Facts

- 当前 API 页面文件为 `frontend/src/routes/api/+page.svelte`。
- 现在页面只展示少量示例卡片，已包含上传、删除、运行时设置与部分响应 JSON。
- 仓库内更完整的接口合同已写在 `docs/api-reference.md`。
- 用户希望“补充 API 页面每个接口的示例”，并明确举例上传区域需要补充调用方式与响应示例。

## Requirements

- API 页面中的接口说明应比当前更完整，至少包含请求示例与成功响应示例。
- 示例内容应与现有后端接口合同保持一致，避免和 `docs/api-reference.md` 冲突。
- 页面仍应保持可复制、适合开发者阅读的前端展示形式。
- 本轮范围限定为 3 类公开接口示例：上传、删除、获取存储选项。
- “获取存储选项”按现有后端合同映射为 `GET /v1/runtime-settings`，并重点展示 `data.storage.options` 的读取方式。
- 本轮不扩展管理员接口示例。

## Acceptance Criteria

- [ ] `/api` 页面补充 `POST /v1/image` 的请求示例与成功响应示例。
- [ ] `/api` 页面补充 `DELETE /i/:uid.avif` 的请求示例与成功响应示例。
- [ ] `/api` 页面补充 `GET /v1/runtime-settings` 的请求示例与成功响应示例，并清楚体现 `storage.options`。
- [ ] 示例与当前实际 API 合同一致。
- [ ] 页面示例仍支持复制，展示风格与现有 API 页面一致。
