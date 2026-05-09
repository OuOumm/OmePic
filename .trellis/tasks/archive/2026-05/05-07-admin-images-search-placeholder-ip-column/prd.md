# 优化后台图片搜索提示和 IP 列

## Goal

让 `/admin/dashboard/images/` 的图片管理页更准确地提示后台接口支持的搜索范围，并在列表视图中直接展示上传 IP，方便管理员按 UID、Token、IP、MD5、存储键定位图片。

## What I already know

* 用户要求优化搜索框提示：增加 uid、ip、md5 等，并按照接口实际可用搜索字段提示。
* 用户要求在选择列表展示图片数据时增加 IP 列。
* 前端页面入口是 `frontend/src/app/admin/dashboard/images/page.tsx`，实际 UI 组件是 `frontend/src/features/admin/ImageTable.tsx`。
* 前端搜索调用 `adminGetImages(token, page, pageSize, search)`，通过 query 参数 `search` 传给 `/admin/images`。
* 后端 `Repository.SearchImages` 当前匹配字段为 `uid`、`token`、`ip_address`、`md5_hash`、`storage_key`。
* 前端 `AdminImage` 类型已经包含 `ip_address` 字段。

## Requirements

* 更新后台图片管理页搜索框 placeholder，使其明确提示可按 UID、Token、IP、MD5、Storage Key 搜索。
* 更新中英文 i18n 文案，避免硬编码单一语言。
* 在图片列表视图（list/table view）中增加 IP 列，展示 `img.ip_address`。
* 保持现有搜索、防抖、分页、批量选择、预览行为不变。
* 不修改后端搜索逻辑，因为接口已支持相关字段搜索。

## Acceptance Criteria

* [ ] `/admin/dashboard/images/` 搜索框提示包含 UID、Token、IP、MD5、Storage Key 等实际可搜索字段。
* [ ] 列表视图表头新增 IP 列。
* [ ] 列表视图每行展示对应图片的 `ip_address`。
* [ ] 网格视图、分页、选择、批量删除、预览功能保持现有行为。
* [ ] 前端 lint/type-check 通过可用的项目命令验证。

## Definition of Done

* Lint / typecheck green where project scripts are available.
* UI 变更通过本地页面检查；若无法启动浏览器或服务，明确说明未做手动 UI 验证。
* 不引入无关重构或后端行为变更。

## Technical Approach

* 修改 `frontend/src/lib/i18n.ts` 中 `admin.imagesSearch` 的英文和中文文案。
* 修改 `frontend/src/features/admin/ImageTable.tsx` 列表视图表头和行渲染，新增 IP 列，使用 `img.ip_address`。
* 根据表格宽度保持 IP 单元格为等宽小字号，必要时使用 `whitespace-nowrap` 防止 IP 换行。

## Decision (ADR-lite)

**Context**: 搜索框提示必须反映接口真实能力，避免提示 UID-only 但实际支持更多字段。

**Decision**: 仅更新前端提示和列表展示；搜索字段以当前后端 `SearchImages` 实际 SQL 条件为准：UID、Token、IP、MD5、Storage Key。

**Consequences**: 不改变 API 行为，风险低；如果未来后端搜索字段变更，需要同步更新 i18n placeholder。

## Out of Scope

* 不新增后端搜索字段。
* 不改搜索参数协议或高级筛选 UI。
* 不改网格卡片布局。
* 不改图片预览弹窗元数据，除非实现时发现已有一致性要求。

## Technical Notes

* `frontend/src/features/admin/ImageTable.tsx` 当前列表列为 UID、Type、Size、Token、MD5、Storage Key、Storage Backend、Created。
* `frontend/src/lib/i18n.ts` 当前 `admin.imagesSearch` 英文为 `Search by UID...`。
* `backend/internal/repository/repository.go` 中 `SearchImages` 的 where 条件：`uid LIKE ? OR token LIKE ? OR ip_address LIKE ? OR md5_hash LIKE ? OR storage_key LIKE ?`。
