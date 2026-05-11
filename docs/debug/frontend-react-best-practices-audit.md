# 前端 React/Vercel Best Practices 适配审查报告

审查日期：2026-05-11  
审查范围：`frontend/` SvelteKit + Svelte 5 + TypeScript 前端项目  
参考依据：`vercel-react-best-practices` 技能中的 React/Next.js 性能最佳实践、项目 `.trellis/spec/frontend/` 前端规范

## 适用性说明

本项目当前前端不是 React 或 Next.js，而是 SvelteKit 2 + Svelte 5 + Vite + TypeScript。仓库检索未发现 React、Next.js、`useState`、`useEffect` 或 Zustand 等 React/Next 代码路径。

因此，本报告不建议直接引入 React 专用 API，例如 `React.cache()`、`next/dynamic`、React Suspense、`useMemo`、`useTransition`、SWR 或 React Server Components。报告仅将 Vercel React Best Practices 中的通用性能原则转译为当前 SvelteKit 项目可落地的审查项，重点覆盖：

- 消除请求竞态与不必要 waterfall
- 降低首屏和路由 bundle 成本
- 控制客户端状态更新频率
- 避免 hydration / 首屏主题闪烁
- 改善静态导出的压缩与资源加载策略

## 修复状态

已整改。本报告列出的 8 项问题均已处理：

- 首屏主题脚本、偏好 store 和旧 helper 已统一为默认 `light`。
- 管理端可切换/高频请求已增加 `AbortSignal` 与 stale response 防护，图片搜索改为 debounce 查询并在条件变化时回到第一页。
- 上传队列已增加并发上限、进度去重和批量完成后的合并刷新。
- 公告列表摘要不再批量执行 Markdown 解析，详情展示仍使用受控 Markdown token 渲染。
- SvelteKit 静态导出已启用 `precompress: true`。
- 客户端 token 已拆分到 `frontend/src/lib/client-token.ts`，`preferences.ts` 仅保留兼容 re-export。

## 总览

本次审查发现 8 项需要关注的问题：

| ID | 严重级别 | 类别 | 摘要 |
|----|----------|------|------|
| FE-RBP-001 | High | Rendering / Hydration | 首屏主题内联脚本默认 dark，与 store 默认 light 不一致，可能导致首屏闪烁 |
| FE-RBP-002 | Medium-High | Client-side Data Fetching | 多个路由加载请求缺少取消与 stale response 防护 |
| FE-RBP-003 | Medium-High | JavaScript / Re-render | 上传进度以 `tasks.map(...)` 高频更新整个队列，批量上传时写放大明显 |
| FE-RBP-004 | Medium | Client-side Data Fetching | 管理端图片搜索缺少 debounce、取消和请求状态隔离 |
| FE-RBP-005 | Medium | Bundle Size | Markdown 渲染依赖 `marked` / `dompurify` 顶层静态导入，可能进入非必要路由 chunk |
| FE-RBP-006 | Medium | JavaScript Performance | 上传 `Promise.all` 无并发上限，可能压满浏览器连接、带宽和后端处理能力 |
| FE-RBP-007 | Low-Medium | Bundle / Static Serving | SvelteKit 静态导出关闭预压缩，需确认 Go 后端是否补齐压缩 |
| FE-RBP-008 | Low-Medium | State / Bundle Hygiene | `preferences.ts` 保留旧主题/语言 helper，与 runes store 默认值不一致 |

## 正向观察

- 项目没有混入 React/Next.js 旧模式，符合 `.trellis/spec/frontend/index.md` 中“不应用 Next.js / React / Zustand 约定”的要求。
- 多处独立请求已使用 `Promise.all`，例如：
  - `frontend/src/routes/admin/dashboard/+page.svelte:35`
  - `frontend/src/routes/admin/dashboard/security/+page.svelte:30`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte:29`
  - `frontend/src/routes/history/+page.svelte:25`
- 图片列表缩略图已使用 `loading="lazy"` 与 `decoding="async"`，例如：
  - `frontend/src/lib/components/studio/ImageDataTable.svelte:33`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:118`
- 公共 API 请求逻辑集中在 `frontend/src/lib/api.ts`，管理端 auth header 也集中封装，符合项目规范。

## 详细问题

### FE-RBP-001：首屏主题脚本默认 dark，与偏好 store 默认 light 不一致

- 对应 Vercel 类别：Rendering Performance / Hydration no flicker
- 严重级别：High
- 位置：
  - `frontend/src/lib/components/studio/AppShell.svelte:42-54`
  - `frontend/src/lib/stores/preferences.svelte.ts:41-47`
  - `frontend/src/lib/preferences.ts:61-77`

证据：

```svelte
// AppShell.svelte
const raw = localStorage.getItem('omepic-ui-preferences');
const prefs = raw ? JSON.parse(raw) : { theme: 'dark' };
const theme = prefs.theme === 'system'
  ? (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
  : (prefs.theme || 'dark');
document.documentElement.classList.toggle('dark', theme === 'dark');
```

```ts
// stores/preferences.svelte.ts
const uiPrefs = readJSON(PREF_STORAGE_KEY, { language: detectLanguage(), theme: 'light' as Theme });

export const preferences = $state<PreferencesState>({
  theme: uiPrefs.theme === 'light' || uiPrefs.theme === 'dark' || uiPrefs.theme === 'system' ? uiPrefs.theme : 'light',
});
```

影响：

- 首次访问且没有本地偏好时，内联脚本会先把页面置为 dark；Svelte store 初始化后又按 light 纠正。
- 这会造成首屏主题闪烁，尤其在静态导出 + 后端 serving 场景下较明显。
- 与 Trellis 前端状态规范中“首次访问默认 light”的契约不一致。

建议：

1. 将 `AppShell.svelte` 的内联主题脚本默认值改为 `light`，并与 `preferences.svelte.ts` 共用同一套合法值判断逻辑。
2. `catch` 分支不要默认 `document.documentElement.classList.add('dark')`，应保持 light 或移除 dark class。
3. 清理 `frontend/src/lib/preferences.ts` 中旧的 `getTheme()` / `resolveTheme()` 默认 dark 逻辑，避免未来误用。
4. 增加一个轻量回归测试或手动检查项：清空 `omepic-ui-preferences` 后刷新首页，不应出现 dark-to-light 闪烁。

---

### FE-RBP-002：多个路由加载请求缺少取消与 stale response 防护

- 对应 Vercel 类别：Client-side Data Fetching / Request deduplication and cancellation
- 严重级别：Medium-High
- 位置：
  - `frontend/src/lib/api.ts:140-145`
  - `frontend/src/lib/api.ts:184-198`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:25-31`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:81`
  - `frontend/src/routes/admin/dashboard/security/+page.svelte:28-46`
  - `frontend/src/routes/admin/dashboard/security/+page.svelte:105`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte:27-31`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte:55`
  - `frontend/src/lib/components/studio/IPDetailPanel.svelte:26-41`
  - `frontend/src/lib/components/studio/IPDetailPanel.svelte:79`

证据：

```ts
// api.ts 仅公共 runtime / announcement helper 暴露 signal
export async function getRuntimeSettings(signal?: AbortSignal): Promise<PublicRuntimeSettings> {
  return apiFetch<PublicRuntimeSettings>("/v1/runtime-settings", { signal });
}

export async function getAnnouncements(signal?: AbortSignal): Promise<Announcement[]> {
  const data = await apiFetch<AnnouncementListResponse>("/v1/announcements", { signal });
  return data.items;
}
```

管理端 helpers 没有接收 `AbortSignal`：

```ts
export async function adminGetImages(
  token: string,
  page: number,
  pageSize: number,
  search?: string
): Promise<AdminImagesResponse> {
  return apiFetch<AdminImagesResponse>("/admin/images", { ... });
}
```

页面 effect 直接发起请求，没有取消或请求序号保护：

```svelte
// admin/dashboard/images/+page.svelte
async function load() {
  if (!preferences.adminToken) return;
  const data = await adminGetImages(preferences.adminToken, page, pageSize, search || undefined);
  images = data.items;
  total = data.total;
  selected = new Set();
}

$effect(() => { load(); });
```

影响：

- 用户快速切换后台 tab、分页、搜索、打开/关闭详情时，旧请求可能晚于新请求返回并覆盖最新状态。
- 组件卸载后仍可能完成请求并写状态，导致无意义工作和潜在 UI 抖动。
- `ImageDetailDrawer.svelte` 已用 `loadedIpAddress` 做了一层结果校验，但其他路由没有统一模式。

建议：

1. 让高频或可切换页面的 API helper 接收 `signal?: AbortSignal`，例如 `adminGetImages(token, page, pageSize, search, signal)`。
2. 在 Svelte `$effect` 内创建 `AbortController`，cleanup 时 abort：

```ts
$effect(() => {
  const controller = new AbortController();
  load(controller.signal);
  return () => controller.abort();
});
```

3. 对无法取消的流程增加 `requestId` / `loadedKey` 检查，只允许最后一次请求写入状态。
4. 对后台 layout 的 token 校验也建议加请求序号，避免 logout/login 临界状态下旧校验写回。

---

### FE-RBP-003：上传进度高频更新时会重复遍历并替换整个任务数组

- 对应 Vercel 类别：JavaScript Performance / Re-render Optimization
- 严重级别：Medium-High
- 位置：
  - `frontend/src/routes/+page.svelte:132-144`
  - `frontend/src/routes/+page.svelte:169-178`

证据：

```svelte
async function uploadTask(task: UploadTask) {
  tasks = tasks.map((item) => (item.id === task.id ? { ...item, status: 'uploading', progress: 0 } : item));
  const result = await uploadImageWithProgress(
    task.file,
    token,
    (progress) => {
      tasks = tasks.map((item) => (item.id === task.id ? { ...item, progress } : item));
    },
    preferences.selectedStorageKey || undefined,
  );
  tasks = tasks.map((item) => (item.id === task.id ? { ...item, status: 'success', progress: 100, result } : item));
}
```

批量上传会并发启动所有任务：

```svelte
tasks = [...next, ...tasks];
await Promise.all(next.map(uploadTask));
```

影响：

- 每个 XHR progress 事件都会遍历整个 `tasks` 数组并创建新数组。
- 多文件、大文件上传时，progress 事件频率高，容易触发大量无意义响应式更新。
- 当前 `activeTasks` 是 `$derived(tasks.filter(...))`，每次任务数组替换也会重新计算。

建议：

1. 对 progress 更新做节流，例如只在百分比变化、`requestAnimationFrame` 或固定间隔内更新 UI。
2. 维护 `Map<taskId, index>` 或任务字典，避免每次 progress 全量 `map`。
3. 对单个任务封装局部状态组件，让一个任务进度变化不迫使整个队列重算。
4. 对批量上传增加并发上限，见 FE-RBP-006。

---

### FE-RBP-004：管理端图片搜索缺少 debounce、取消和请求状态隔离

- 对应 Vercel 类别：Client-side Data Fetching / Event-driven fetching
- 严重级别：Medium
- 位置：
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:17`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:25-31`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:81`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:92`

证据：

```svelte
let search = $state('');

async function load() {
  if (!preferences.adminToken) return;
  const data = await adminGetImages(preferences.adminToken, page, pageSize, search || undefined);
  images = data.items;
  total = data.total;
  selected = new Set();
}

$effect(() => { load(); });
```

搜索框只有按 Enter 触发：

```svelte
<input ... bind:value={search} ... onkeydown={(event) => event.key === 'Enter' && load()} />
```

影响：

- `$effect` 会订阅 `search`，搜索输入变化可能触发 `load()`；同时 Enter 又手动调用 `load()`，容易出现重复请求。
- 没有 debounce 或取消逻辑，连续输入会造成后台图片列表接口高频请求。
- 搜索条件改变时没有显式将 `page` 重置为 1，可能在高页码下搜索得到空结果，误以为无数据。

建议：

1. 明确搜索策略：要么“输入 debounce 自动搜索”，要么“输入不触发，点击/Enter 搜索”。不要同时让 `$effect` 与 Enter 都触发同一查询。
2. 如果保留自动搜索，使用 250-400ms debounce，并配合 `AbortController` 取消前一次查询。
3. 搜索条件变化时重置 `page = 1`。
4. 用 `{ page, pageSize, search }` 组成查询 key，只有 query key 变化时才发请求。

---

### FE-RBP-005：Markdown 渲染依赖顶层静态导入，可能增加非必要路由 bundle

- 对应 Vercel 类别：Bundle Size Optimization / Conditional module loading / Dynamic imports
- 严重级别：Medium
- 位置：
  - `frontend/src/lib/components/studio/MarkdownContent.svelte:2-18`
  - `frontend/src/lib/components/studio/AnnouncementDialog.svelte:6`
  - `frontend/src/lib/components/studio/AnnouncementDialog.svelte:63-74`
  - `frontend/src/lib/components/studio/AnnouncementManager.svelte:13`
  - `frontend/src/routes/+page.svelte:3`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte:5`

证据：

```svelte
// MarkdownContent.svelte
import DOMPurify from 'dompurify';
import { marked, type Token } from 'marked';

marked.use({
  async: false,
  gfm: true,
  breaks: true
});

const tokens = $derived(marked.lexer(DOMPurify.sanitize(content)));
```

首页静态导入公告弹窗，而公告弹窗静态导入 Markdown 渲染组件：

```svelte
// routes/+page.svelte
import AnnouncementDialog from '@/components/studio/AnnouncementDialog.svelte';
```

```svelte
// AnnouncementDialog.svelte
import MarkdownContent from './MarkdownContent.svelte';
```

影响：

- `marked` 与 `dompurify` 是相对重的文本处理依赖；即使公告弹窗未打开，也可能被打进对应路由 chunk。
- 公告历史模式中每条公告都渲染 `MarkdownContent`，多公告时会批量 lexer/sanitize。
- 当前 `MarkdownContent` 没有使用 `{@html}`，最终只输出文本节点和受控元素；`DOMPurify.sanitize(content)` 的必要性可以重新评估，避免不必要处理成本。

建议：

1. 用构建分析确认 `marked` / `dompurify` 是否进入首页首屏 chunk。
2. 对公告弹窗或 `MarkdownContent` 使用 Svelte 可用的动态导入/条件加载，只在需要打开公告或管理公告时加载 Markdown 解析逻辑。
3. 对 history 列表的摘要模式可先用纯文本截断，点击详情时再解析完整 Markdown。
4. 如果继续不使用 `{@html}`，可评估是否用 `marked.lexer` 的 token 白名单替代 DOMPurify，减少重复 sanitize 成本；如果未来改回 HTML 输出，则必须保留严格 sanitize。

---

### FE-RBP-006：批量上传 `Promise.all` 没有并发上限

- 对应 Vercel 类别：Eliminating Waterfalls / JavaScript Performance
- 严重级别：Medium
- 位置：
  - `frontend/src/routes/+page.svelte:169-178`
  - `frontend/src/lib/api.ts:78-131`

证据：

```svelte
async function handleFiles(files: File[]) {
  const accepted = validateFiles(files);
  const next = accepted.map((file) => ({ id: `task-${++counter}`, file, progress: 0, status: 'pending' as const }));
  tasks = [...next, ...tasks];
  await Promise.all(next.map(uploadTask));
}
```

影响：

- `Promise.all` 对独立请求是好实践，但上传是高带宽、高内存、高后端成本操作；无上限并发会让大量文件同时启动 XHR。
- 可能导致浏览器连接排队、进度 UI 抖动、后端瞬时压力升高、用户无法判断哪些任务真正开始。
- 和 FE-RBP-003 的 progress 高频更新叠加后，批量上传场景成本更高。

建议：

1. 增加上传并发池，例如同时上传 2-4 个文件，其余保持 pending。
2. 并发数可根据运行时设置或浏览器网络情况调节。
3. 在队列 UI 中区分 pending、uploading、success、error，并允许后续扩展暂停/取消。
4. 对大批量任务的 `loadRecent()` 刷新做合并；当前每个成功任务都会调用一次 `loadRecent()`（`frontend/src/routes/+page.svelte:160-162`），可改为批量上传结束后刷新一次，或节流刷新。

---

### FE-RBP-007：静态导出关闭预压缩，需确认后端 serving 是否补齐压缩

- 对应 Vercel 类别：Bundle Size Optimization / Static asset delivery
- 严重级别：Low-Medium
- 位置：
  - `frontend/svelte.config.js:7-13`
  - `frontend/package.json:8-12`

证据：

```js
adapter: adapter({
  pages: 'out',
  assets: 'out',
  fallback: 'index.html',
  precompress: false,
  strict: true,
}),
```

`build:backend` 会把静态产物复制到 Go 后端：

```json
"build:backend": "vite build && node scripts/copy-static-to-backend.mjs"
```

影响：

- 构建输出显示了 gzip size，但 `precompress: false` 不会生成 `.gz` / `.br` 静态文件。
- 如果 Go 后端没有动态 gzip/brotli middleware，生产环境可能直接传输未压缩 JS/CSS。
- 对移动网络下首屏 JS/CSS 下载时间影响较明显。

建议：

1. 检查 Go 后端静态文件 serving 是否启用了 gzip/brotli。
2. 如果后端不做压缩，可考虑将 adapter-static `precompress` 改为 `true`，并让后端优先 serving `.br` / `.gz`。
3. 把压缩策略写入 `.trellis/spec/frontend/quality-guidelines.md` 或后端静态资源规范，避免部署路径不一致。

---

### FE-RBP-008：`preferences.ts` 保留旧主题/语言 helper，与当前 runes store 不一致

- 对应 Vercel 类别：Bundle Hygiene / State management consistency
- 严重级别：Low-Medium
- 位置：
  - `frontend/src/lib/preferences.ts:37-97`
  - `frontend/src/lib/stores/preferences.svelte.ts:41-91`
  - `frontend/src/routes/+page.svelte:10`
  - `frontend/src/routes/api/+page.svelte:4`
  - `frontend/src/routes/history/+page.svelte:9`

证据：

`preferences.ts` 中仍保留旧的语言/主题 helper：

```ts
export function getTheme(): Theme {
  if (typeof window === "undefined") return "dark";
  ...
  return "dark";
}

export function resolveTheme(theme: Theme): "light" | "dark" {
  if (theme === "system") {
    if (typeof window === "undefined") return "dark";
    return window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }
  return theme;
}
```

当前 store 的契约是默认 light：

```ts
const uiPrefs = readJSON(PREF_STORAGE_KEY, { language: detectLanguage(), theme: 'light' as Theme });
```

但 `preferences.ts` 仍被多个路由导入，主要为了 `getClientToken()`：

```svelte
import { getClientToken } from '@/preferences';
```

影响：

- 旧 helper 当前大多未使用，但仍存在未来误用风险。
- 文件同时承载 client token 与旧 UI preference helper，不符合当前 runes store 作为偏好状态唯一入口的规范。
- 与 FE-RBP-001 一样，旧默认 dark 逻辑会加剧首屏主题一致性问题。

建议：

1. 将 `getClientToken()` 拆到更明确的文件，例如 `frontend/src/lib/client-token.ts`。
2. 删除或废弃 `preferences.ts` 中 `getLanguage`、`setLanguage`、`getTheme`、`setTheme`、`resolveTheme` 等旧 helper。
3. 所有 UI 语言/主题读写只通过 `frontend/src/lib/stores/preferences.svelte.ts`。
4. 如需保留兼容导出，内部也必须委托当前 store，并统一默认 light。

## 建议整改顺序

1. 先修 FE-RBP-001：这是用户可见的首屏问题，且修复范围小。
2. 再处理 FE-RBP-002 与 FE-RBP-004：为管理端查询引入取消、请求 key 和 debounce，降低竞态风险。
3. 处理 FE-RBP-003 与 FE-RBP-006：上传队列是核心路径，建议一并做并发控制和 progress 节流。
4. 处理 FE-RBP-005：先用 bundle 分析确认 `marked` / `dompurify` 影响，再决定动态加载或轻量摘要渲染。
5. 最后处理 FE-RBP-007 与 FE-RBP-008：分别沉淀部署压缩策略和偏好状态边界。

## 建议验证命令

完成整改后建议运行：

```bash
cd frontend
npm run lint
npm run typecheck
npm run test
npm run build:backend
```

如涉及资源体积优化，建议额外增加一次 Vite bundle 分析或至少对比 `build:backend` 输出 chunk 体积。
