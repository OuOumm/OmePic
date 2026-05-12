# 前端 Svelte 5 审查与整改建议

- 日期：2026-05-13
- 范围：`frontend/` 当前 SvelteKit 2 + Svelte 5 + Vite + TypeScript 前端
- 依据：Svelte 5 runes / event / `{@attach}` 最佳实践，项目 `.trellis/spec/frontend/*` 规范
- 产出性质：本文件最初为审查与整改建议文档；2026-05-13 已按本文件完成整改实现。

## 0. 整改完成记录

已完成整改项：

- P1-01：管理后台布局新增 `logged_out/checking/authenticated` 门禁状态，token 校验成功前不渲染受保护子路由。
- P1-02：删除 `frontend/next-env.d.ts`、`frontend/eslint.config.mjs`、未使用的 `frontend/src/lib/actions/click-outside.ts`，并清理 ESLint `.next/**` ignore。
- P2-01：公告弹窗在 detail 模式重新打开或最新公告变化时重置到最新公告。
- P2-02：文件选择器 `accept` 由运行时 effective MIME 列表生成，并过滤 SVG。
- P2-03：图片详情抽屉 IP 详情请求改为可 abort，关闭/切图时清理过期请求。
- P2-04：保留首屏主题脚本并为 `theme === 'system'` 增加运行时系统主题变化监听；测试继续覆盖主题决策 helper。
- P3-01：删除未使用旧式 Svelte action。
- P3-02：补齐 `image.copyUid` / `image.url` 双语键并替换图片详情硬编码字段标签。
- P3-03：`safeImageUrl` 支持传入运行时 `public_base_url` origin allow-list，历史/最近预览与表格缩略图已接入。

整改后验证：

| 命令 | 结果 |
|---|---:|
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run lint"` | 通过 |
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run typecheck"` | 通过 |
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run test"` | 通过，9 个测试文件、45 个测试 |
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run build:backend"` | 通过，仅 Rolldown 插件耗时提示 |

## 1. 自动化验证结果

| 命令 | 结果 | 备注 |
|---|---:|---|
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run lint"` | 通过 | 直接在 bash 中跑 `npm` 会因当前 shell 缺少 `sed/dirname/uname` 失败，改用 Windows `cmd.exe` 后通过。 |
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run typecheck"` | 通过 | `svelte-check found 0 errors and 0 warnings` |
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run test"` | 通过 | 9 个测试文件、41 个测试通过 |
| `cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run build:backend"` | 通过 | 静态文件已复制到 `backend/web/`；仅有 Rolldown 插件耗时提示 |

## 2. 总体结论

当前前端整体已经迁移到 SvelteKit/Svelte 5：

- 大多数组件使用 `$props()`、`$state`、`$derived`，未发现大面积 `export let`、`on:click`、`<slot>`、`use:` 等旧写法回流。
- API 请求基本集中在 `frontend/src/lib/api.ts`，符合项目规范。
- 共享偏好、toast 使用 runes store，上传历史封装在 IndexedDB helper 中。
- 多数列表已使用 keyed each block。

但仍有若干需要整改的风险点，优先级最高的是**管理员路由鉴权门禁**、**遗留 Next.js 文件**、**公告确认状态与显示对象可能错位**、**上传输入 accept 与 SVG 禁止策略不完全一致**。

## 3. 整改清单

### P1-01 管理后台子页面会在 token 校验完成前渲染

- 位置：`frontend/src/routes/admin/dashboard/+layout.svelte:44-66`、`frontend/src/routes/admin/dashboard/+layout.svelte:69-124`
- 现象：只要 `preferences.adminToken` 有值，布局就立即渲染侧栏与 `children`；异步 `adminGetStatus()` 失败后才清理 token 并跳转。
- 风险：
  - 过期/伪造 token 刷新 `/admin/dashboard/images`、`/settings`、`/security` 时，受保护子路由可能短暂渲染并发起管理 API 请求。
  - 与项目规范“Admin layout validates a stored token through the backend before rendering protected children”不一致。
- 建议整改：
  1. 在布局中新增显式状态，例如 `authState: 'logged_out' | 'checking' | 'authenticated'`。
  2. 有 token 时先进入 `checking`，只显示加载态；`adminGetStatus` 成功后再置为 `authenticated` 并渲染子页面。
  3. 校验失败时调用 `clearAdminToken()`，状态回到 `logged_out`，跳转 `/admin/dashboard`。
  4. 子路由的加载逻辑继续检查 `preferences.adminToken`，但 UI 渲染以布局门禁为准。
- 验收：
  - 手动写入无效 `omepic-admin-token` 后刷新 `/admin/dashboard/images`，应先看到校验/加载态，然后回登录页，不应展示图片表格或设置表单。
  - 有效 token 刷新深层管理路由，应先校验再显示目标页面。

### P1-02 SvelteKit 前端仍保留 Next.js 遗留文件

- 位置：`frontend/next-env.d.ts`、`frontend/eslint.config.mjs`、`frontend/eslint.config.js` 中的 `.next/**` ignore
- 现象：仓库中仍存在 `next-env.d.ts`，内容引用 `next` 与 `./.next/types/routes.d.ts`；同时存在一个简化版旧 `eslint.config.mjs`。
- 风险：
  - 与项目规范“Do not apply old Next.js App Router, React component, or Zustand conventions”冲突。
  - 新成员或工具可能误判该前端仍使用 Next.js。
  - 双 ESLint 配置增加维护歧义。
- 建议整改：
  1. 删除 `frontend/next-env.d.ts`。
  2. 删除不用的 `frontend/eslint.config.mjs`，保留当前 `eslint.config.js`。
  3. 如 `.next/**` ignore 已无意义，可从 `eslint.config.js` 中移除；若保留也应在注释中说明只是历史兼容。
  4. 重新运行 `npm run lint`、`npm run typecheck`，确认无 Next 类型依赖。

### P2-01 公告弹窗的确认对象可能与最新公告不一致

- 位置：`frontend/src/lib/components/studio/AnnouncementDialog.svelte:41-44`、`frontend/src/routes/+page.svelte:109-114`
- 现象：`AnnouncementDialog` 在打开时只设置 `mode = initialMode`，没有在打开新一轮详情弹窗时重置 `index`；而首页 `acknowledgeAnnouncementDialog()` 总是把 `announcements[0]` 的时间戳写入 `omepic:announcement:lastSeen`。
- 风险：
  - 用户曾在历史模式中切到旧公告后，下次 detail 模式打开仍可能显示旧公告。
  - 点击“Got it”时可能确认的是最新公告时间戳，但视觉上看到的是旧公告。
- 建议整改：
  1. `open && !wasOpen` 时，若 `initialMode === 'detail'`，将 `index = 0`。
  2. 当 `announcements[0]?.id` 变化时，自动重置详情索引。
  3. 可考虑在 history 模式中隐藏或改文案说明“确认最新公告”，避免语义混淆。
- 验收：
  - 手动打开公告历史并选择旧公告，关闭后触发最新公告自动弹出，应显示 `announcements[0]`。
  - 点击确认后写入的 `lastSeen` 与当前显示公告一致。

### P2-02 文件选择 accept 过宽，未与后端/运行时 MIME allow-list 完全对齐

- 位置：`frontend/src/lib/components/studio/CanvasDropzone.svelte:83`、`frontend/src/routes/+page.svelte:53-55`
- 现象：文件输入使用 `accept="image/*"`，浏览器选择器仍可能展示 SVG；项目规范要求客户端 accept list 与后端 raster 类型保持一致，且不得加入 SVG。
- 风险：
  - 虽然后续 `isAllowedImageMimeType()` 会拒绝 `image/svg+xml`，但原生选择器层面仍给用户错误暗示。
  - 若运行时配置变化，UI 展示和 file picker allow-list 可能漂移。
- 建议整改：
  1. 从 `runtimeSettings.upload.effective_allowed_mime_types` 生成 `accept` 字符串，并过滤 `image/svg+xml`。
  2. 给 `CanvasDropzone` 增加 `accept?: string` prop，默认使用安全 raster MIME 列表而不是 `image/*`。
  3. 拖拽和粘贴路径继续复用 `isAllowedImageMimeType()`，保持单一判断逻辑。
- 验收：
  - 默认文件选择器不应展示 SVG 为可选类型。
  - 运行时 allow-list 改变后，展示文本、accept、实际校验三者一致。

### P2-03 图片详情抽屉的 IP 详情请求不可取消，关闭/切图时可能留下悬挂加载态

- 位置：`frontend/src/lib/components/studio/ImageDetailDrawer.svelte:58-80`
- 现象：`loadIPDetail()` 调用 `adminGetAbuseIPDetail()` 时未传 `AbortSignal`；组件关闭或切图时仅靠 `loadedIpAddress` 避免旧结果覆盖。
- 风险：
  - 快速切图/关闭时仍会保留无用网络请求。
  - 关闭后旧请求完成时，`ipDetailLoading` 可能不能在所有路径下及时复位。
- 建议整改：
  1. 将 IP 详情加载改成 effect 内部创建 `AbortController`，并传入 `adminGetAbuseIPDetail(token, ip, signal)`。
  2. effect cleanup 中 abort。
  3. `!image` 分支显式 `ipDetailLoading = false`。
- 验收：
  - 快速切换图片时只保留最新 IP 的详情。
  - 关闭抽屉后不再出现旧 IP 加载错误 toast。

### P2-04 主题初始化脚本与已测试 helper 存在重复逻辑

- 位置：`frontend/src/lib/components/studio/AppShell.svelte:39-58`、`frontend/src/lib/utils.ts:90-100`
- 现象：`AppShell.svelte` 的 `<svelte:head><script>` 内联解析 `omepic-ui-preferences`；`utils.ts` 中另有 `getInitialThemeScriptTheme()` 并有测试覆盖，但生产内联脚本未复用该 helper。
- 风险：
  - 测试覆盖的是 helper，不一定覆盖实际首屏主题脚本。
  - 未来修改 theme 默认值或 system 逻辑时容易两处漂移。
- 建议整改：
  - 方案 A：抽出一个生成内联脚本文本的 helper，并测试生成内容中的关键分支。
  - 方案 B：删除未被生产使用的 helper 测试，直接增加面向 `AppShell` 行为的端到端/组件测试。
  - 同时考虑为 `prefers-color-scheme` 的变化添加监听，使 `theme === 'system'` 时能跟随系统变化。

### P3-01 存在未使用的旧式 Svelte action

- 位置：`frontend/src/lib/actions/click-outside.ts`
- 现象：该文件导出旧式 action，但当前代码没有使用；项目已采用 `{@attach attachAccessibleDialog(...)}` 模式。
- 风险：
  - 后续开发者可能重新引入 `use:clickOutside`，与 Svelte 5 指南“use `{@attach ...}` instead of `use:action`”冲突。
- 建议整改：
  - 若确认无用途，删除该文件。
  - 若仍需点击外部关闭能力，提供 `attachClickOutside()` 包装，并在组件中使用 `{@attach ...}`。

### P3-02 i18n 缺失键与硬编码可见文案

- 位置：`frontend/src/lib/components/studio/ImageDetailDrawer.svelte:160-164`、`frontend/src/lib/components/studio/ImageDataTable.svelte:83-85`、`frontend/src/lib/i18n.ts`
- 现象：
  - `ImageDetailDrawer` 使用 `t(preferences.language, 'image.copyUid')`，但 `i18n.ts` 中未定义该键，aria-label 会回退为 `image.copyUid`。
  - 详情字段显示 `UID`、`URL`、`MD5`、`Token`、`IP` 等硬编码文本。部分是通用缩写，但已有 `image.uid`、`image.token`、`image.md5`、`image.ip`，应优先复用。
  - 表格复制按钮显示 `URL`、`MD`、`BB`，可考虑使用 title/aria-label 与翻译键保持一致。
- 风险：
  - 可访问性文本不友好。
  - 新增语言时难以覆盖所有可见文案。
- 建议整改：
  1. 在英/中文字典中补齐 `image.copyUid`，必要时增加 `image.url`。
  2. 详情字段用 `t(...)` 输出。
  3. 缩写按钮至少保证 `aria-label` 完整翻译，必要时加 `title`。

### P3-03 历史预览的图片 URL 只允许同源，可能与绝对 public URL 合同不一致

- 位置：`frontend/src/lib/components/studio/ImagePreviewDialog.svelte:24`、`frontend/src/lib/utils.ts:54-66`
- 现象：`safeImageUrl()` 拒绝所有非当前 origin 的 URL；但 `UploadResult.url` 合同允许返回绝对公开 URL。
- 风险：
  - 如果后端 `public_base_url` 配到 CDN 或独立域名，历史页预览/下载入口可能不可用。
- 建议整改：
  1. 明确产品是否允许跨域 public URL。
  2. 若允许，`safeImageUrl` 应接受当前 origin 加运行时 `public_base_url` origin 的白名单。
  3. 若不允许，类型/文档应收紧，明确历史预览只支持同源 URL。

## 4. Svelte 5 专项检查记录

| 检查项 | 当前情况 | 结论 |
|---|---|---|
| `$props()` | 组件普遍使用 `$props()`，未发现 `export let` 回流 | 良好 |
| 事件语法 | 使用 `onclick` / `onchange` 等现代事件属性，未发现大量 `on:` | 良好 |
| `$derived` | 计算值多数使用 `$derived`，例如 nav、分页、运行时派生值 | 良好 |
| `$state.raw` | API 响应列表、历史列表等大数组多处使用 `$state.raw` | 良好 |
| `$effect` | 数据加载/DOM 同步场景仍使用 `$effect`，大多合理；个别需要 abort 或避免状态错位 | 需局部整改 |
| `{@attach}` | 弹窗可访问性逻辑已通过 `attachAccessibleDialog` 包装 | 良好 |
| keyed each | 表格/列表大多使用稳定 key | 良好 |
| i18n | 大部分文案集中在 `i18n.ts`；仍有缺失键/缩写硬编码 | 需局部整改 |

## 5. 推荐整改顺序

1. **先做 P1**：管理后台 token 校验门禁、删除 Next.js 遗留文件。
2. **再做 P2**：公告确认索引、上传 accept、IP 详情请求取消、主题初始化逻辑去重。
3. **最后做 P3**：未用 action 清理、i18n 补键、跨域 public URL 策略确认。

每批整改后建议执行：

```bash
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run lint"
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run typecheck"
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run test"
cmd.exe //c "cd /d D:\\Works\\MyProject\\OmePic\\frontend && npm run build:backend"
```

## 6. 手动验收建议

- 无 token / 无效 token / 有效 token 分别刷新 `/admin/dashboard`、`/admin/dashboard/images`、`/admin/dashboard/settings`。
- 公告：关闭不确认、确认最新公告、查看历史公告后再确认。
- 上传：文件选择、拖拽、粘贴、URL 上传；特别验证 SVG 被客户端拒绝且提示一致。
- 图片详情：快速打开/关闭、上一张/下一张切换、IP ban、删除后列表刷新。
- 主题/语言：刷新后首屏主题无闪烁，语言和 theme 持久化正常。
