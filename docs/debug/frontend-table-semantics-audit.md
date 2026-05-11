# 前端表格语义化调查报告

审查范围：`frontend/src` 下使用 `div`、`grid`、`studio-table-row`、`#each` 等结构展示多列数据的页面与组件。

审查依据：Web Interface Guidelines 中的可访问性规则：优先使用语义 HTML，包括 `<button>`、`<a>`、`<label>`、`<table>`，不要在有原生语义元素可用时用无语义 `div` 模拟。

## 修复状态

已按本报告完成主要处理：

- P1：后台图片管理桌面列表、安全页 Top IPs、已封禁 IP 列表已改为原生 `<table>` / `<thead>` / `<tbody>` / `<tr>` / `<th>` / `<td>`。
- P2：上传历史和首页最近上传已抽象为共享组件 `frontend/src/lib/components/studio/ImageDataTable.svelte`，使用语义表格结构。
- P3：图片详情键值信息已改为 `<dl>` / `<dt>` / `<dd>`；上传任务队列已改为 `<ul>` / `<li>`。
- 未强制修改移动端图片卡片、公告卡片、存储实例管理卡片和已正确使用 `<dl>` 的设置摘要。

## 结论

当前前端原本没有发现原生 `<table>` / `<thead>` / `<tbody>` / `<tr>` / `<td>` 结构，表格型数据主要由 `div + CSS grid` 模拟。

这不是所有场景都错误，但对于具备以下特征的结构，应优先改为原生 `<table>`：

- 有明确列头。
- 有多条同构记录。
- 每行由同一组字段组成。
- 用户需要理解单元格和列头之间的关系。

对于单个对象详情、任务队列、公告卡片、设置摘要等结构，不建议机械改为 `<table>`；更合适的语义通常是 `<dl>`、`<ul><li>`、`article` 或保留卡片布局。

## P1：建议优先改为 `<table>`

### frontend/src/routes/admin/dashboard/images/+page.svelte

frontend/src/routes/admin/dashboard/images/+page.svelte:110 - 已修复：桌面端图片管理列表由 `div.grid` 表头改为 `<table>` 表头。

frontend/src/routes/admin/dashboard/images/+page.svelte:121 - 已修复：图片记录行由 `div.grid` 改为 `<tr>` / `<td>` / `<th scope="row">`。

当前列结构：

- 选择
- 预览
- UID
- 大小
- 存储 Key
- IP
- 操作

判断：这是典型二维数据表，已改为 `<table>`、`<thead>`、`<tbody>`、`<tr>`、`<th scope="col">`、`<th scope="row">`、`<td>`。

移动端卡片结构位于同文件 81-108 行，视觉和交互更接近卡片列表，已按原建议保留。

### frontend/src/routes/admin/dashboard/security/+page.svelte

frontend/src/routes/admin/dashboard/security/+page.svelte:75 - 已修复：Top IPs 表头由 `div.grid` 改为 `<thead>`。

frontend/src/routes/admin/dashboard/security/+page.svelte:81 - 已修复：Top IPs 行由 `div.grid` 改为 `<tbody>` 中的 `<tr>`。

当前列结构：

- IP
- 上传数
- 大小
- 操作

判断：这是统计数据表，已改为原生 `<table>`。

frontend/src/routes/admin/dashboard/security/+page.svelte:99 - 已修复：已封禁 IP 表头由 `div.grid` 改为 `<thead>`。

frontend/src/routes/admin/dashboard/security/+page.svelte:106 - 已修复：已封禁 IP 行由 `div.grid` 改为 `<tbody>` 中的 `<tr>`。

当前列结构：

- IP
- 原因
- 时长
- 操作

判断：这是管理数据表，已改为原生 `<table>`。

## P2：建议抽象为语义化表格组件

### frontend/src/lib/components/studio/ImageDataTable.svelte

frontend/src/lib/components/studio/ImageDataTable.svelte:15 - 已新增：上传记录表格组件，内部渲染 `<table>`、`<thead>`、`<tbody>`。

调用位置：

- frontend/src/routes/history/+page.svelte:81
- frontend/src/routes/+page.svelte:308

当前列结构：

- 图片 / 文件
- 大小
- 存储
- 操作

处理结果：原 `ImageDataRow.svelte` 已不再使用并删除；历史页和首页最近上传统一使用 `ImageDataTable.svelte`。

## P3：可考虑，但不是必须改为 table

### frontend/src/lib/components/studio/StorageInstanceManager.svelte

frontend/src/lib/components/studio/StorageInstanceManager.svelte:142 - 保留：存储实例列表继续使用 `div.grid` + `article` 展示多列管理项。

当前字段结构：

- 名称 / Storage Key
- 后端类型
- 操作

判断：它有表格特征，但当前视觉上更像管理卡片列表。优先级低于图片管理列表和安全页列表，本轮未强制改为 table。

## 不建议改为 table 的位置

### frontend/src/lib/components/studio/ImageDetailDrawer.svelte

frontend/src/lib/components/studio/ImageDetailDrawer.svelte:122 - 已修复：图片详情键值信息由多行 `div.grid` 改为 `<dl>`、`<dt>`、`<dd>`。

当前字段包括：

- URL
- MD5
- Token
- IP
- 安全状态
- 大小
- 类型
- 存储
- 创建时间

判断：这是单个对象的属性说明，不是多条同构记录组成的二维数据表，不应改为 `<table>`。

### frontend/src/routes/+page.svelte

frontend/src/routes/+page.svelte:281 - 已修复：上传队列由 `div` 列表改为 `<ul>` / `<li>`，内部保留现有 `role="progressbar"`。

判断：这是任务队列/进度列表，不是数据表，不应改 table。

### frontend/src/lib/components/studio/AnnouncementManager.svelte

frontend/src/lib/components/studio/AnnouncementManager.svelte:139 - 保留：公告管理列表继续使用 `article` 卡片结构。

判断：这是内容卡片列表，不是二维数据表，不建议改 table。

### frontend/src/lib/components/studio/AnnouncementDialog.svelte

frontend/src/lib/components/studio/AnnouncementDialog.svelte:54 - 保留：公告历史继续使用按钮卡片列表。

判断：这是可点击公告条目列表，不是数据表，不建议改 table。

### frontend/src/lib/components/studio/StorageInspector.svelte

frontend/src/lib/components/studio/StorageInspector.svelte:46 - 保留：设置摘要已经使用 `<dl>`、`<dt>`、`<dd>`。

判断：当前语义合适，不需要改 table。

## 修复优先级汇总

| 优先级 | 位置 | 建议 | 状态 |
| --- | --- | --- | --- |
| P1 | `frontend/src/routes/admin/dashboard/images/+page.svelte:110-134` | 桌面端图片管理列表改为 `<table>` | 已完成 |
| P1 | `frontend/src/routes/admin/dashboard/security/+page.svelte:75-90` | Top IPs 列表改为 `<table>` | 已完成 |
| P1 | `frontend/src/routes/admin/dashboard/security/+page.svelte:99-116` | 已封禁 IP 列表改为 `<table>` | 已完成 |
| P2 | `frontend/src/lib/components/studio/ImageDataRow.svelte:15-40` | 新增/重构为 `ImageDataTable.svelte`，父级提供表头和表结构 | 已完成，旧组件已删除 |
| P3 | `frontend/src/lib/components/studio/StorageInstanceManager.svelte:142-159` | 可改 table，也可保留卡片/列表语义 | 保留 |
| P3 | `frontend/src/lib/components/studio/ImageDetailDrawer.svelte:122-132` | 不改 table，建议改 `<dl>` | 已完成 |
| P3 | `frontend/src/routes/+page.svelte:281-292` | 不改 table，建议改 `<ul><li>` | 已完成 |

## 后续注意事项

- 移动端卡片布局不必强制改 table；可以保留卡片，用桌面端 table + 移动端 card 的响应式策略。
- `<th>` 应使用 `scope="col"`，如果第一列是行标题，可使用 `scope="row"`。
- 操作列可使用 `<th scope="col">操作</th>`，按钮仍保留明确 `aria-label`。
- 表格外层可保留 `overflow-x-auto`，避免小屏横向溢出。
- 避免用 `role="table"` 代替原生 `<table>`；除非布局限制极大，否则原生语义更可靠。
