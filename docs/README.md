# OmePic Code Wiki

OmePic 是一个自托管的图片托管服务（Image Hosting Service），采用单体仓库（Monorepo）结构，包含 Go 后端和 SvelteKit 前端。

## 目录

- [项目架构总览](architecture-overview.md) — 整体架构、层次划分、数据流
- [后端架构详解](backend-architecture.md) — Go 后端各模块详细说明
- [前端架构详解](frontend-architecture.md) — SvelteKit 前端结构与组件
- [API 参考](api-reference.md) — 全部 API 端点与请求/响应格式
- [运行与部署](running-and-deployment.md) — 本地开发、生产构建、环境变量
- [Cloudflare 图片 URL 缓存清理](cloudflare-single-url-cache-purge.md) — Cloudflare purge 目标、配置、流程与测试覆盖
