# 前端原生浏览器弹窗审查报告

审查范围：`frontend/src` 下的 Svelte 页面、Svelte 组件与 TypeScript 代码。

审查依据：Vercel Web Interface Guidelines，重点参考以下规则：

- Destructive actions need confirmation modal or undo window—never immediate
- Async updates need `aria-live="polite"`
- Use semantic HTML before ARIA
- Interactive elements need visible focus and keyboard support

## 结论

本次未发现 `alert()` 或 `prompt()` 调用。

发现 11 处 `confirm()` 原生浏览器确认框，全部用于删除、清空、解封、清理图片等破坏性操作。虽然这些操作已经有确认步骤，但使用浏览器原生弹窗会带来以下问题：

- 视觉不可控，无法匹配当前 OmePic 手绘/纸张风格主题。
- 无法复用现有 `accessibleDialog` 的焦点进入、焦点陷阱、Escape 关闭与关闭后焦点恢复能力。
- 无法在确认文案中稳定呈现标题、风险说明、影响范围、按钮层级与危险态样式。
- 与项目已有 Toast、Dialog、Drawer 体系割裂，交互体验不一致。

建议统一替换为项目内自定义确认对话框组件，例如 `ConfirmDialog.svelte` 或共享确认服务，并接入 `frontend/src/lib/actions/accessible-dialog.ts`。

## 问题代码位置

## frontend/src/routes/admin/dashboard/images/+page.svelte

frontend/src/routes/admin/dashboard/images/+page.svelte:33 - `removeSelected()` 使用原生 `confirm()` 确认批量删除图片；应替换为自定义危险确认模态，显示选中数量，并使用危险主按钮。

frontend/src/routes/admin/dashboard/images/+page.svelte:41 - `removeOne(image)` 使用原生 `confirm()` 确认删除单张图片；应替换为自定义确认模态，显示目标 UID，避免浏览器默认弹窗割裂页面风格。

## frontend/src/routes/admin/dashboard/security/+page.svelte

frontend/src/routes/admin/dashboard/security/+page.svelte:44 - `unban(id)` 使用原生 `confirm()` 确认解封 IP；应替换为自定义确认模态，说明解封后该 IP 可再次上传。

frontend/src/routes/admin/dashboard/security/+page.svelte:51 - `purgeImages(id)` 使用原生 `confirm()` 确认清理封禁 IP 的图片；应替换为自定义危险确认模态，突出该操作会删除相关图片。

## frontend/src/lib/components/studio/ImageDetailDrawer.svelte

frontend/src/lib/components/studio/ImageDetailDrawer.svelte:69 - 图片详情抽屉内 `remove()` 使用原生 `confirm()` 确认删除当前图片；应替换为嵌套或提升层级的自定义确认模态，并保持抽屉焦点管理正确。

## frontend/src/routes/+page.svelte

frontend/src/routes/+page.svelte:206 - 首页最近上传 `removeRecent(record)` 使用原生 `confirm()` 确认删除历史图片；应替换为自定义确认模态或提供可撤销删除体验。

## frontend/src/routes/history/+page.svelte

frontend/src/routes/history/+page.svelte:29 - 历史页 `clearAll()` 使用原生 `confirm()` 确认清空全部上传历史；应替换为自定义危险确认模态，明确影响范围是“全部历史记录”。

frontend/src/routes/history/+page.svelte:37 - 历史页 `remove(record)` 使用原生 `confirm()` 确认删除单条历史图片；应替换为自定义确认模态或撤销窗口。

## frontend/src/lib/components/studio/StorageInstanceManager.svelte

frontend/src/lib/components/studio/StorageInstanceManager.svelte:122 - 存储实例管理 `remove(instance)` 使用原生 `confirm()` 确认删除存储配置；应替换为自定义危险确认模态，显示实例名称与 storage key，并说明默认实例不可删。

## frontend/src/lib/components/studio/AnnouncementManager.svelte

frontend/src/lib/components/studio/AnnouncementManager.svelte:109 - 公告管理 `remove(item)` 使用原生 `confirm()` 确认删除公告；应替换为自定义确认模态，显示公告标题并使用危险态按钮。

## frontend/src/lib/components/studio/IPDetailPanel.svelte

frontend/src/lib/components/studio/IPDetailPanel.svelte:62 - IP 详情面板 `purgeImages()` 使用原生 `confirm()` 确认删除该 IP 相关图片；应替换为自定义危险确认模态，配合已有面板/对话框焦点管理。

## 建议整改方案

1. 新增共享确认对话框组件，建议放在 `frontend/src/lib/components/studio/ConfirmDialog.svelte`。
2. 组件能力建议包括：`open`、`title`、`description`、`confirmLabel`、`cancelLabel`、`tone="danger"`、`busy`、`onConfirm`、`onClose`。
3. 对话框内部使用原生 `<button>`，接入 `use:accessibleDialog`，保留 Escape 关闭、焦点陷阱与关闭后焦点恢复。
4. 将 11 处 `confirm()` 改为状态驱动的确认流程：点击删除按钮只设置待确认目标；用户在自定义模态中确认后再执行真实删除。
5. 删除成功继续使用现有 `toast.success()`；删除失败继续使用 `toast.error()`，由 Toast live region 对辅助技术播报。

## 后续验证建议

- 全局搜索确认不再存在业务代码中的 `confirm(`、`alert(`、`prompt(`。
- 前端执行 `npm run lint` 与 `npm run typecheck`。
- 手动验证 11 个入口的取消、确认、Escape、Tab 焦点循环、确认中禁用按钮和删除成功 Toast。
