# 修复模态框遮罩层未覆盖视口

## Goal

修复网站中除首页外其他页面打开模态框/对话框时，遮罩层未完整覆盖浏览器视口的问题，确保所有页面均具备标准模态框遮罩效果。

## Problem

当前部分页面点击按钮弹出模态框时，背后的 backdrop/overlay 只覆盖局部容器，导致遮罩之外的内容仍可见或可能交互。

## User Value

- 弹窗体验一致，符合用户对模态对话框的预期。
- 防止遮罩外内容被误点或误操作。
- 降低不同页面布局对弹窗组件的影响。

## Requirements

- 找出项目中模态框/对话框遮罩层实现或样式不正确的根因。
- 修复遮罩层定位/层级/挂载逻辑，使其覆盖完整浏览器视口。
- 修复范围应覆盖首页以外页面，不得只针对单个页面硬编码。
- 保持现有弹窗外观、动画和交互语义不被破坏。
- 不混入与本问题无关的改动。

## Acceptance Criteria

- [x] 在首页以外页面打开模态框时，遮罩层覆盖整个视口。
- [x] 遮罩层外内容不可见为未遮罩状态，且不可被交互。
- [x] 首页现有模态框/对话框行为不回退。
- [x] 前端 lint/typecheck/test 至少运行与改动范围匹配的校验并通过，或记录非代码原因。
- [x] 完成复核并记录验证结果。

## Validation Plan

- 代码检查：确认 overlay/backdrop 使用 viewport 固定定位或等价全屏挂载方式。
- 前端校验：`cd frontend && npm run lint && npm run typecheck`。
- 如可行，补充/运行相关测试或构建命令。

## Completion Report

- 实现摘要：新增 `frontend/src/lib/actions/viewport-portal.ts`，将共享模态框根节点挂载到 `document.body`，避免被页面/布局容器的 `overflow`、层叠上下文或局部布局约束；已接入 ImagePreview、Confirm、BanIP、Announcement、ImageDetail、IPDetail、Storage/Announcement 管理等共享对话框。同步修复一个阻塞前端 lint 的历史页局部去重写法（行为等价）。
- 验证命令与结果：`cd frontend && npm run lint && npm run typecheck` 通过；`cd frontend && npm run test && npm run build:backend` 通过（10 个测试文件、55 个测试通过；静态资源已复制到 backend/web）。
- 复核结论：代码检查确认所有 `frontend/src/lib/components/studio` 中的 `fixed inset-0` 模态根节点均使用 `attachViewportPortal()`，backdrop 仍保持固定定位、原有层级、动画、Esc/点击遮罩关闭与焦点陷阱语义。
- 风险/后续：未做浏览器截图级手动验证；建议在 `/history`、`/admin/dashboard/images`、`/admin/dashboard/security|settings` 实机打开弹窗复查遮罩覆盖视口。

