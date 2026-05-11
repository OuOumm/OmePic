# 前端 Web Interface Guidelines 审查报告

审查范围：`frontend/src` 下的 Svelte 页面、Svelte 组件与全局 CSS。

审查依据：Vercel Web Interface Guidelines，重点覆盖可访问性、焦点状态、表单、对话框、导航状态、图片/动画/性能与内容处理规则。

## 修复状态

已修复本报告列出的主要问题，并通过以下前端验证：

- `npm run lint`
- `npm run typecheck`
- `npm run build:backend`

构建仅输出插件耗时提示，无 lint、类型或构建错误。

## 总览

本次审查共发现 39 项问题，主要集中在以下类别：

- 焦点状态与键盘可达性：全局输入/按钮焦点样式不足，多处对话框缺少焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复。
- 可访问状态通知：Toast、运行时错误、登录错误、上传进度缺少 `aria-live`、`role="alert"` 或进度条 ARIA 属性。
- 表单语义：URL 输入、密码输入、搜索输入、文件输入和复选框存在类型、自动完成或可访问名称缺失。
- 导航状态：桌面/移动主导航和管理后台导航缺少 `aria-current`。
- 交互语义与安全：文件上传区域存在嵌套交互控件，外链缺少 `rel` 防护。

## frontend/src/app.css

frontend/src/app.css:132 - 已修复：`.studio-button` 补充显式 `:focus-visible` 样式。

frontend/src/app.css:186 - 已修复：`.studio-input` 移除无替代的 `outline: none`。

frontend/src/app.css:190 - 已修复：`.studio-input` 改为 `:focus-visible` 并增强焦点指示。

## frontend/src/lib/components/studio/AppShell.svelte

frontend/src/lib/components/studio/AppShell.svelte:63 - 已修复：主导航链接补充 `aria-current="page"`。

frontend/src/lib/components/studio/AppShell.svelte:81 - 已修复：移动菜单按钮补充本地化 `aria-label`、`aria-expanded` 与 `aria-controls`。

frontend/src/lib/components/studio/AppShell.svelte:89 - 已修复：移动导航链接补充 `aria-current`。

## frontend/src/lib/components/studio/ToastViewport.svelte

frontend/src/lib/components/studio/ToastViewport.svelte:6 - 已修复：Toast 容器补充 `role="status"`、`aria-live="polite"` 与 `aria-atomic="true"`。

## frontend/src/lib/components/studio/CanvasDropzone.svelte

frontend/src/lib/components/studio/CanvasDropzone.svelte:34 - 已修复：上传区域改为非交互容器，选择文件使用内部按钮触发，避免嵌套交互控件。

frontend/src/lib/components/studio/CanvasDropzone.svelte:56 - 已修复：允许文件类型改为页面可见文本，并通过 `aria-describedby` 关联。

frontend/src/lib/components/studio/CanvasDropzone.svelte:72 - 已修复：文件输入补充可访问名称与说明关联。

## frontend/src/routes/+page.svelte

frontend/src/routes/+page.svelte:244 - 已修复：运行时错误信息补充 `role="alert"`。

frontend/src/routes/+page.svelte:263 - 已修复：URL 输入补充 `type="url"`、`name`、`autocomplete="url"` 与 `inputmode="url"`。

frontend/src/routes/+page.svelte:288 - 已修复：上传进度条补充 `role="progressbar"` 与 `aria-valuemin`、`aria-valuemax`、`aria-valuenow`。

## frontend/src/lib/components/studio/ImagePreviewDialog.svelte

frontend/src/lib/components/studio/ImagePreviewDialog.svelte:21 - 已修复：Escape 处理迁移到可聚焦对话框节点。

frontend/src/lib/components/studio/ImagePreviewDialog.svelte:22 - 已修复：对话框接入共享焦点管理，支持焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复。

## frontend/src/lib/components/studio/AnnouncementDialog.svelte

frontend/src/lib/components/studio/AnnouncementDialog.svelte:46 - 已修复：Escape 处理迁移到可聚焦对话框节点。

frontend/src/lib/components/studio/AnnouncementDialog.svelte:47 - 已修复：对话框接入共享焦点管理，支持焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复。

## frontend/src/lib/components/studio/BanIPDialog.svelte

frontend/src/lib/components/studio/BanIPDialog.svelte:57 - 已修复：模态层接入 Escape 关闭处理。

frontend/src/lib/components/studio/BanIPDialog.svelte:58 - 已修复：对话框接入共享焦点管理，支持焦点进入、焦点陷阱与关闭后焦点恢复。

## frontend/src/lib/components/studio/ImageDetailDrawer.svelte

frontend/src/lib/components/studio/ImageDetailDrawer.svelte:55 - 已修复：模态对话框接入共享焦点管理，支持焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复。

frontend/src/lib/components/studio/ImageDetailDrawer.svelte:71 - 已修复：URL 复制图标按钮补充可访问名称。

frontend/src/lib/components/studio/ImageDetailDrawer.svelte:72 - 已修复：MD5 复制图标按钮补充可访问名称。

frontend/src/lib/components/studio/ImageDetailDrawer.svelte:73 - 已修复：Token 复制图标按钮补充可访问名称。

## frontend/src/lib/components/studio/IPDetailPanel.svelte

frontend/src/lib/components/studio/IPDetailPanel.svelte:79 - 已修复：模态对话框接入共享焦点管理，支持焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复。

## frontend/src/lib/components/studio/StorageInstanceManager.svelte

frontend/src/lib/components/studio/StorageInstanceManager.svelte:163 - 已修复：`role="dialog"` 补充 `aria-labelledby`。

frontend/src/lib/components/studio/StorageInstanceManager.svelte:163 - 已修复：模态编辑器接入共享焦点管理，支持焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复。

frontend/src/lib/components/studio/StorageInstanceManager.svelte:202 - 已修复：密钥密码输入补充 `autocomplete="new-password"`。

frontend/src/lib/components/studio/StorageInstanceManager.svelte:212 - 已修复：WebDAV 密码输入补充 `autocomplete="new-password"`。

## frontend/src/lib/components/studio/AnnouncementManager.svelte

frontend/src/lib/components/studio/AnnouncementManager.svelte:164 - 已修复：详情模态接入共享焦点管理，支持焦点进入、焦点陷阱与关闭后焦点恢复。

frontend/src/lib/components/studio/AnnouncementManager.svelte:182 - 已修复：编辑模态 `role="dialog"` 补充 `aria-labelledby`。

frontend/src/lib/components/studio/AnnouncementManager.svelte:182 - 已修复：编辑模态接入共享焦点管理，支持焦点进入、焦点陷阱与关闭后焦点恢复。

## frontend/src/routes/admin/dashboard/+page.svelte

frontend/src/routes/admin/dashboard/+page.svelte:55 - 已修复：管理员密码输入补充 `autocomplete="current-password"`。

frontend/src/routes/admin/dashboard/+page.svelte:57 - 已修复：登录错误信息补充 `role="alert"`。

## frontend/src/routes/admin/dashboard/+layout.svelte

frontend/src/routes/admin/dashboard/+layout.svelte:47 - 已修复：管理侧边栏链接补充 `aria-current`。

frontend/src/routes/admin/dashboard/+layout.svelte:54 - 已修复：设置子导航补充 `aria-current`。

## frontend/src/routes/admin/dashboard/images/+page.svelte

frontend/src/routes/admin/dashboard/images/+page.svelte:50 - 已修复：搜索输入补充屏幕阅读器标签、`name` 与 `autocomplete="off"`。

frontend/src/routes/admin/dashboard/images/+page.svelte:63 - 已修复：图片选择复选框补充可访问名称。

frontend/src/routes/admin/dashboard/images/+page.svelte:86 - 已修复：图片选择复选框补充可访问名称。

frontend/src/routes/admin/dashboard/images/+page.svelte:94 - 已修复：`target="_blank"` 链接补充 `rel="noreferrer"`。

## 后续建议

- 当前共享焦点管理已覆盖本次审查中发现的模态/抽屉；后续新增对话框时建议统一接入 `frontend/src/lib/actions/accessible-dialog.ts`。
- 如果后续引入浏览器端测试框架，建议补充对键盘 Tab 循环、Escape 关闭、`aria-current` 与 Toast live region 的自动化回归测试。
