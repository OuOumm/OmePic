# 优化全局错误页

## Goal

优化 `frontend/src/routes/+error.svelte` 的内容与视觉表现，让全局错误页能按状态码提供更明确的信息，并保持与当前前端风格一致的高质量界面。

## Confirmed Facts

- 当前全局错误页位于 `frontend/src/routes/+error.svelte`。
- 当前实现只做了非常简单的错误展示：标题固定为 `common.error`，404 时正文直接复用 `history.noMatches`。
- 当前页面没有针对 404 / 403 / 500 等常见状态做差异化文案与视觉层次。
- 当前项目已有统一视觉基础设施：`studio-panel`、`studio-button`、纸张/荧光笔风格 CSS 变量、双语 i18n 字典。
- `App.Page` 中可直接读取 `page.status` 与 `page.error`，因此无需新增接口即可完成状态驱动展示。

## Requirements

- 优化 `frontend/src/routes/+error.svelte` 的布局与样式，使其更美观、更有层次。
- 按状态码提供差异化错误展示，至少覆盖 `404` 与通用服务错误场景。
- 页面文案不能再复用历史页的 `history.noMatches`，应使用专门的错误页文案。
- 保持与现有前端视觉语言一致，继续使用当前项目的 sketch / paper / studio 风格。
- 提供明确的恢复操作入口，例如返回首页。
- 如新增可见文案，需同步补齐 `frontend/src/lib/i18n.ts` 中英文词条。

## Acceptance Criteria

- [ ] 访问不存在的路由时，错误页明确展示 `404` 与“页面不存在”语义，而不是历史记录空状态文案。
- [ ] 非 404 场景下，错误页展示通用错误标题/说明，并保留可恢复入口。
- [ ] 错误页视觉风格明显优于当前版本，具备更完整的排版、层次和状态识别。
- [ ] 新增文案已同步覆盖 `en` / `zh`。
- [ ] 前端校验命令通过：`npm run lint`、`npm run typecheck`、`npm run build:backend`。

## Out of Scope

- 后端错误码体系调整。
- 新增专门的错误页路由树或 server hook 错误处理逻辑。
- 针对每一种 HTTP 状态码做完整独立插画或复杂交互。

## Open Questions

- 无。
