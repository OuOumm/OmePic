# 前端硬编码与多语言适配审查报告

审查范围：`frontend/src` 下的 Svelte 页面、组件、i18n 与工具函数。

审查重点：依据 Web Interface Guidelines 的 Locale & i18n 规则，重点检查用户可见文案是否绕过 `t(...)` 多语言字典、日期/数字/单位展示是否使用显式 `Intl.*` 或语言偏好。

## 修复状态

已修复本报告列出的主要问题：

- 用户可见硬编码文案已补齐 i18n key，并在英文、中文字典中提供对应翻译。
- 文件大小、MB 单位与日期展示已统一改为显式 `Intl.NumberFormat` / `Intl.DateTimeFormat`，并接收当前界面语言。
- 现有文件大小和日期格式化调用已改为传入 `preferences.language` 或组件 `language`。

## 总览

本次审查共发现 22 项问题，主要集中在：

- 用户可见英文文案直接写在 Svelte 模板或组件默认值中，切换中文时不会翻译。
- API 示例标题、按钮文案和页面 eyebrow 仍为硬编码英文。
- 上传队列、预览标签、管理侧边栏标题等核心界面文案未进入 i18n 字典。
- 文件大小、日期时间、上传限制单位等格式化逻辑没有显式绑定当前语言环境。

以下条目已排除品牌名、技术缩写和数据字段标签，例如 `OmePic`、`OP`、`URL`、`MD5`、`Token`、`IP`、`API`、`DB`、`Redis`、`S3`、`WebDAV` 等。

## frontend/src/lib/components/studio/AppShell.svelte

frontend/src/lib/components/studio/AppShell.svelte:83 - 已修复：主题切换按钮 `Light` / `Dark` 改为 `t(preferences.language, 'common.light'/'common.dark')`。

## frontend/src/lib/components/studio/BlueprintFlow.svelte

frontend/src/lib/components/studio/BlueprintFlow.svelte:2 - 已修复：组件默认标题 `Pipeline` 改为 i18n key `studio.pipeline`。

frontend/src/lib/components/studio/BlueprintFlow.svelte:10 - 已修复：用户可见标签 `live sketch` 改为 i18n key `studio.liveSketch`。

## frontend/src/lib/components/studio/CanvasDropzone.svelte

frontend/src/lib/components/studio/CanvasDropzone.svelte:65 - 已修复：装饰卡片文案 `Paste` 改为 i18n key `upload.sourcePaste`。

frontend/src/lib/components/studio/CanvasDropzone.svelte:71 - 已修复：装饰卡片文案 `Host` 改为 i18n key `upload.sourceHost`。

## frontend/src/lib/components/studio/ImagePreviewDialog.svelte

frontend/src/lib/components/studio/ImagePreviewDialog.svelte:18 - 已修复：图片标题回退文案 `image` 改为 i18n key `common.fallbackImage`。

frontend/src/lib/components/studio/ImagePreviewDialog.svelte:26 - 已修复：预览标签 `preview` 改为 i18n key `studio.preview`。

## frontend/src/lib/components/studio/StorageInspector.svelte

frontend/src/lib/components/studio/StorageInspector.svelte:46 - 已修复：上传大小展示改为 `formatMegabytes(settings.upload.max_upload_size_mb, language)`。

## frontend/src/lib/utils.ts

frontend/src/lib/utils.ts:8 - 已修复：`formatBytes(bytes)` 改为 `formatBytes(bytes, language)`，支持按当前语言格式化。

frontend/src/lib/utils.ts:13 - 已修复：文件大小数字格式化改为 `Intl.NumberFormat`，不再使用 `toFixed()`。

frontend/src/lib/utils.ts:16 - 已修复：`formatDate(dateStr)` 改为 `formatDate(dateStr, language)`，支持按当前语言格式化。

frontend/src/lib/utils.ts:17 - 已修复：日期格式化改为显式 `Intl.DateTimeFormat(locale(language), ...)`，不再依赖浏览器默认语言。

## frontend/src/routes/+page.svelte

frontend/src/routes/+page.svelte:282 - 已修复：上传队列标题 `Upload queue` 改为 i18n key `upload.queueTitle`。

frontend/src/routes/+page.svelte:294 - 已修复：空队列提示 `No active uploads.` 改为 i18n key `upload.queueEmpty`。

frontend/src/routes/+page.svelte:304 - 已修复：最近上传区域 eyebrow `File desk` 改为 i18n key `admin.fileDeskEyebrow`。

## frontend/src/routes/history/+page.svelte

frontend/src/routes/history/+page.svelte:62 - 已修复：历史页 `PageTitle` 的 eyebrow `File desk` 改为 i18n key `admin.fileDeskEyebrow`。

## frontend/src/routes/api/+page.svelte

frontend/src/routes/api/+page.svelte:12 - 已修复：API 示例标题 `Upload` 改为 i18n key `api.exampleUpload`。

frontend/src/routes/api/+page.svelte:16 - 已修复：API 示例标题 `Delete` 改为 i18n key `api.exampleDelete`。

frontend/src/routes/api/+page.svelte:20 - 已修复：API 示例标题 `Response` 改为 i18n key `api.exampleResponse`。

frontend/src/routes/api/+page.svelte:41 - 已修复：API 页 eyebrow `Developer notes` 改为 i18n key `api.eyebrow`。

frontend/src/routes/api/+page.svelte:59 - 已修复：复制按钮 `Copy` 改为 i18n key `common.copy`。

## frontend/src/routes/admin/dashboard/+layout.svelte

frontend/src/routes/admin/dashboard/+layout.svelte:47 - 已修复：管理侧边栏标题 `Admin Blueprint` 改为 i18n key `admin.blueprintTitle`。

## frontend/src/routes/admin/dashboard/settings/+page.svelte

frontend/src/routes/admin/dashboard/settings/+page.svelte:91 - 已修复：表单标签不再硬编码 `/ MB`，当前运行时限制通过 `formatMegabytes(system.runtime.max_upload_size_mb, preferences.language)` 展示。

## 后续建议

1. 后续新增用户可见文案时同步补齐英文、中文字典，不直接写入 Svelte 模板。
2. 日期、数字、文件大小、单位展示统一使用 `frontend/src/lib/utils.ts` 中的本地化格式化 helper。
3. 技术缩写可保留原样，但按钮、提示、标题、ARIA 文案应继续使用 i18n。
4. 如果后续引入测试框架，建议加入语言切换快照或 DOM 查询，覆盖英文/中文两种语言。
