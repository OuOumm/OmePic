# 完成报告：Cloudflare 图片 URL 缓存清理

> 2026-05-21 更新：Cloudflare Zone ID、API Token 与 API Base URL 已迁移为后台运行时设置，保存后热更新生效；不再通过 `CLOUDFLARE_*` 启动环境变量配置。

## 当前状态

OmePic 支持通过 Cloudflare API 清理图片 URL 的缓存，并保持以下核心要求不变：单密码登录管理后台、前端免登录生成 token 上传、Redis 缓存秒传。

## 当前实现要点

### 后端

- `CloudflareCachePurger` 调用 Cloudflare `POST /zones/{zone_id}/purge_cache`。
- 请求体使用 Cloudflare `files` 数组；手动 purge 为单 URL，后台批量删除会尽量一次提交多个图片 URL。
- 使用 `Authorization: Bearer <cloudflare_api_token>` 调用 Cloudflare。
- URL 清理前进行规范化：trim、只允许 `http/https`、必须有 host、移除 fragment。
- Runtime settings 包含：
  - `cloudflare_purge_enabled`
  - `cloudflare_zone_id`
  - `cloudflare_api_token`
  - `cloudflare_api_base_url`
- `cloudflare_api_token` 存储于 SQLite，作为 admin-only secret 处理：后台 GET 仅返回遮罩值，公开 API 不返回。
- 后台保存系统设置后，手动 purge 和删除图片自动 purge 都会使用最新 Cloudflare 配置，无需重启服务。
- 启用 Cloudflare 清理时要求 `public_base_url`、Zone ID 和 API Token 均可用。
- 删除图片时先清理 `{public_base_url}/i/{uid}.avif`，清理成功后才删除记录和 Redis 缓存；后台批量删除会先合并待删 URL 到一次 purge。
- Cloudflare 清理失败时删除流程停止，保留图片记录。
- 管理员接口：`POST /admin/cloudflare/purge-image-cache`。

### 前端

- `RuntimeSettings` 类型包含 Cloudflare Zone ID、API Token、API Base URL 字段。
- 管理后台设置页可编辑 Cloudflare runtime 配置。
- API Token 输入框显示后端遮罩值；保持遮罩值表示不变，清空表示清除 Token。
- 设置页保留手动清理 URL 表单。

### 文档

- 详细设计文档：`docs/cloudflare-single-url-cache-purge.md`。
- API 合同：`docs/api-reference.md`。
- 部署说明：`docs/running-and-deployment.md`。

## 设计约束

- 只清理图片 URL；不要扩展为全站、前缀或 `purge_everything`。
- Cloudflare API Token 不进入公开 API，不在前端明文展示，不写入日志。
- 已有依赖旧环境变量的部署升级后需要在后台设置页重新填写 Cloudflare 配置。
