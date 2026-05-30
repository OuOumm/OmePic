# 后台 IP 取消脱敏展示

## Goal

后台管理端不再对 IP 做脱敏展示，统一直接显示真实 `ip_address`，减少排查滥用与封禁时的信息损耗。

## Confirmed Facts

- 当前后台前后端同时存在 `ip_address` 与 `ip_address_masked` 两套字段。
- 后台图片列表、图片详情、滥用分析、IP 封禁列表、IP 详情等界面当前主要展示 `ip_address_masked`。
- 后端安全分析与 IP 封禁流程会生成/返回 `ip_address_masked`。
- SQLite `ip_bans` 表当前持久化 `ip_address_masked` 列。
- `backend/internal/repository/repository_test.go` 已明确：对已废弃列不做兼容重建，旧开发库应手动重置。
- 现有 `.trellis/spec` 中的 backend security/database 与 frontend type-safety 文档仍将 `ip_address_masked` 视为有效契约。

## Requirements

- 后台管理端所有 IP 展示统一改为 `ip_address`。
- 移除后台前后端响应模型中的 `ip_address_masked` 使用。
- 不再依赖 `ip_address_masked` 完成后台展示、封禁、滥用分析与详情交互。
- 彻底删除后端 `ip_address_masked` 生成逻辑、模型字段、repository 读写以及 SQLite `ip_bans` 持久化列。
- 变更后后台相关操作（查看、封禁、解封、按封禁删除图片）保持可用。

## Acceptance Criteria

- [ ] 后台图片列表中的 IP 列展示 `ip_address`，不再展示 `ip_address_masked`。
- [ ] 后台图片详情、IP 详情、滥用排行、IP 封禁列表、相关确认弹窗均展示 `ip_address`。
- [ ] 后台前端类型与后端 JSON 响应不再要求 `ip_address_masked` 才能正常工作。
- [ ] 后端 `model` / `service` / `repository` / API JSON 中不再保留 `ip_address_masked`。
- [ ] SQLite `ip_bans` 表不再包含 `ip_address_masked` 持久化列，且相关流程在改动后仍正常。
- [ ] 如变更影响现有 spec，需同步更新 `.trellis/spec` 中相关约定。

## Out of Scope

- 公共上传/访客侧页面的 IP 展示策略调整。
- IP 存储来源、可信代理解析、封禁判定逻辑调整。
- 非 IP 字段的脱敏策略调整。
- 对旧版 SQLite `ip_bans` 结构提供原地兼容迁移；旧库由开发/部署侧按需重建。

## Open Questions

- 无。
