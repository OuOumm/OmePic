# 前端架构详解

## 目录

1. [技术栈](#1-技术栈)
2. [目录结构](#2-目录结构)
3. [页面路由](#3-页面路由)
4. [核心组件](#4-核心组件)
5. [API 调用层](#5-api-调用层)
6. [状态管理](#6-状态管理)
7. [关键模块](#7-关键模块)
8. [构建配置](#8-构建配置)

---

## 1. 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| SvelteKit | ^2.59.1 | 应用框架 |
| Svelte | ^5.55.5 | 前端组件框架（runes API） |
| TypeScript | ^5.9.3 | 类型安全 |
| Tailwind CSS | ^3.4.17 | 样式框架 |
| Vite | ^8.0.10 | 构建工具 |
| Vitest | ^4.0.16 | 单元测试 |
| lucide-svelte | ^0.468.0 | 图标库 |
| clsx + tailwind-merge | 2.x / 3.x | CSS 类名合并 |
| dompurify | ^3.4.2 | HTML 安全过滤 |
| marked | ^18.0.3 | Markdown 渲染 |

## 2. 目录结构

```
frontend/
├── src/
│   ├── app.css                # 全局样式（Tailwind + 自定义变量）
│   ├── app.html               # HTML 模板
│   ├── routes/                # SvelteKit 页面路由
│   │   ├── +layout.svelte     # 根布局（AppShell）
│   │   ├── +page.svelte       # 首页（上传页面）
│   │   ├── api/+page.svelte   # 开发者 API 页面
│   │   ├── history/+page.svelte # 上传历史页面
│   │   └── admin/dashboard/   # 管理后台（含子页面）
│   │       ├── +layout.svelte # 管理后台布局（侧边栏）
│   │       ├── +page.svelte   # 登录 / 总览
│   │       ├── images/+page.svelte     # 图片管理
│   │       ├── security/+page.svelte   # 安全面板
│   │       └── settings/+page.svelte   # 设置面板
│   └── lib/                   # 共享代码
│       ├── api.ts             # 后端 API 封装
│       ├── types/index.ts     # TypeScript 类型定义
│       ├── i18n.ts            # 国际化
│       ├── utils.ts           # 工具函数
│       ├── client-token.ts    # 客户端 Token 管理
│       ├── clipboard.ts       # 剪贴板操作
│       ├── upload-queue.ts    # 上传队列控制
│       ├── ui-errors.ts       # UI 错误处理
│       ├── performance-utils.ts # 性能工具
│       ├── preferences.ts     # 废弃的偏好设置（旧版）
│       ├── components/studio/ # UI 组件
│       ├── stores/            # Svelte 5 stores (runes)
│       ├── actions/           # Svelte actions
│       └── indexeddb/         # IndexedDB 上传历史
├── static/favicon.svg
├── svelte.config.js           # SvelteKit 配置（adapter-static）
├── tailwind.config.ts         # Tailwind 配置
├── vite.config.ts             # Vite 配置
├── tsconfig.json
└── package.json
```

## 3. 页面路由

### 3.1 首页 — `+page.svelte`

**路径**: [src/routes/+page.svelte](file:///d:/Works/MyProject/OmePic/frontend/src/routes/+page.svelte)

主要的图片上传页面，包含：
- **CanvasDropzone**: 拖拽/点击上传区域
- **URL 上传**: 通过 URL 抓取远程图片
- **StorageInspector**: 存储实例选择器
- **上传队列**: 实时进度显示
- **RecentUploads**: 近期上传记录表格
- **公告系统**: 最新公告弹窗提示

核心状态：
```typescript
let tasks = $state<UploadTask[]>([]);              // 上传任务队列
let recentUploads = $state<UploadHistoryRecord[]>([]); // 历史记录
let announcements = $state<Announcement[]>([]);    // 公告
let previewRecord = $state<UploadHistoryRecord | null>(null); // 预览
```

上传体验特性：
- 拖拽上传（CanvasDropzone）
- 粘贴上传（ClipboardEvent 监听）
- URL 抓取上传
- 并发上传（默认 3 个并发）
- 上传进度条
- IndexedDB 本地持久化历史

### 3.2 上传历史 — `history/+page.svelte`

查看完整的上传历史记录列表

### 3.3 开发者 API — `api/+page.svelte`

开发者文档页面，展示 API 使用说明

### 3.4 管理后台

**路由**: `/admin/dashboard/*`

**布局**: [admin/dashboard/+layout.svelte](file:///d:/Works/MyProject/OmePic/frontend/src/routes/admin/dashboard/+layout.svelte)

- 侧边栏导航：Status、Images、Security、Settings
- 全局登录状态管理
- 自动获取并显示总览统计

#### 子页面

| 路由 | 用途 |
|------|------|
| `/admin/dashboard` | 总览面板（含登录表单） |
| `/admin/dashboard/images` | 图片管理（搜索、分页、批量删除） |
| `/admin/dashboard/security?tab=abuse` | 滥用监控面板 |
| `/admin/dashboard/security?tab=rate-limit` | 速率限制配置 |
| `/admin/dashboard/settings?tab=runtime` | 运行时设置 |
| `/admin/dashboard/settings?tab=storage` | 存储实例管理 |
| `/admin/dashboard/settings?tab=announcements` | 公告管理 |

## 4. 核心组件

**路径**: [src/lib/components/studio/](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/)

| 组件 | 用途 |
|------|------|
| **AppShell.svelte** | 应用主壳层（导航栏、页脚） |
| **CanvasDropzone.svelte** | 文件拖拽/点击上传区域 |
| **ImageDataTable.svelte** | 图片数据表格（含复制链接、预览、删除） |
| **ImagePreviewDialog.svelte** | 图片预览对话框 |
| **ImageSwitchButton.svelte** | 切换视图模式按钮 |
| **StorageInspector.svelte** | 存储实例选择器 |
| **StorageInstanceManager.svelte** | 存储实例管理（Admin） |
| **MetricStrip.svelte** | 统计数据显示条 |
| **PageTitle.svelte** | 页面标题组件 |
| **AnnouncementDialog.svelte** | 公告显示对话框 |
| **AnnouncementManager.svelte** | 公告管理（Admin CRUD） |
| **ConfirmDialog.svelte** | 确认操作对话框 |
| **BanIPDialog.svelte** | IP 封禁对话框 |
| **IPDetailPanel.svelte** | IP 详情面板 |
| **BlueprintFlow.svelte** | 蓝图流程图组件 |
| **MarkdownContent.svelte** | Markdown 安全渲染 |
| **ToastViewport.svelte** | Toast 通知容器 |

## 5. API 调用层

**路径**: [src/lib/api.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/api.ts)

### 核心设计

```typescript
// 通用 API 调用
async function apiFetch<T>(path, options) → T

// 上传（使用 XMLHttpRequest 支持进度回调）
function uploadImageWithProgress(file, token, onProgress, storageKey?) → Promise<UploadResult>

// 错误处理
class ApiError extends Error {
    code?: string;
    status?: number;
    retryAfter?: number;
}
```

### 公开 API

| 函数 | 端点 |
|------|------|
| `getRuntimeSettings()` | `GET /v1/runtime-settings` |
| `getAnnouncements()` | `GET /v1/announcements` |
| `deleteImageByUid(uid, token)` | `DELETE /i/:uid` |
| `uploadImageWithProgress(file, token, onProgress, storageKey?)` | `POST /v1/image` |

### 管理 API

| 函数 | 端点 |
|------|------|
| `adminLogin(password)` | `POST /admin/login` |
| `adminGetStatus(token)` | `GET /admin/status` |
| `adminGetImages(token, page, pageSize, search?)` | `GET /admin/images` |
| `adminDeleteImages(token, uids)` | `DELETE /admin/images` |
| `adminCreateIPBan(token, input)` | `POST /admin/ip-bans` |
| `adminGetIPBans(token)` | `GET /admin/ip-bans` |
| `adminDeleteIPBan(token, id)` | `DELETE /admin/ip-bans/:id` |
| `adminDeleteIPBanImages(token, id)` | `DELETE /admin/ip-bans/:id/images` |
| `adminGetAbuseOverview(token, from?, to?)` | `GET /admin/abuse/overview` |
| `adminGetAbuseIPDetail(token, ip)` | `GET /admin/abuse/ip` |
| `adminGetConfig(token)` | `GET /admin/config` |
| `adminCreateStorageInstance(token, instance)` | `POST /admin/config/storage-instances` |
| `adminUpdateStorageInstance(token, key, instance)` | `PUT /admin/config/storage-instances/:storageKey` |
| `adminDeleteStorageInstance(token, key)` | `DELETE /admin/config/storage-instances/:storageKey` |
| `adminSetDefaultStorage(token, key)` | `POST /admin/config/default` |
| `adminGetSystemSettings(token)` | `GET /admin/system-settings` |
| `adminUpdateSystemSettings(token, settings)` | `PUT /admin/system-settings` |
| `adminGetAnnouncements(token)` | `GET /admin/announcements` |
| `adminCreateAnnouncement(token, input)` | `POST /admin/announcements` |
| `adminUpdateAnnouncement(token, id, input)` | `PUT /admin/announcements/:id` |
| `adminDeleteAnnouncement(token, id)` | `DELETE /admin/announcements/:id` |
| `adminArchiveAnnouncement(token, id)` | `POST /admin/announcements/:id/archive` |

## 6. 状态管理

**路径**: [src/lib/stores/preferences.svelte.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/stores/preferences.svelte.ts)

使用 Svelte 5 的 runes API（`$state`）管理：

```typescript
// 偏好设置（持久化到 localStorage）
preferences:
  - language: 'en' | 'zh'          # 语言
  - theme: 'light' | 'dark' | 'system' # 主题
  - viewMode: 'grid' | 'list'      # 视图模式
  - adminToken: string | null       # 管理 JWT
  - runtimeSettings: PublicRuntimeSettings | null  # 运行时配置
  - selectedStorageKey: string      # 用户选择的存储实例
```

**路径**: [src/lib/stores/toast.svelte.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/stores/toast.svelte.ts)

- `toast.success(msg)`, `toast.error(msg)`, `toast.info(msg)`

## 7. 关键模块

### 7.1 客户端 Token

**路径**: [src/lib/client-token.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/client-token.ts)

- `getClientToken()`: 获取/生成客户端标识 token（localStorage 持久化）
- 用于 `X-Token` 请求头标识上传者

### 7.2 上传队列

**路径**: [src/lib/upload-queue.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/upload-queue.ts)

- `runWithConcurrency(tasks, concurrency)`: 并发控制执行器
- `createProgressReporter(onProgress)`: 进度变化去重报告器

### 7.3 IndexedDB 上传历史

**路径**: [src/lib/indexeddb/upload-history.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/indexeddb/upload-history.ts)

- `saveUploadToHistory(record)`: 保存上传记录
- `getRecentUploads(limit)`: 获取最近上传
- `deleteUploadFromHistory(uid)`: 删除上传记录
- `getAllHistory()`: 获取全部历史

### 7.4 国际化

**路径**: [src/lib/i18n.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/i18n.ts)

- 支持 `en`（英文）和 `zh`（中文）
- `t(lang, key, params?)`: 翻译函数

### 7.5 剪贴板

**路径**: [src/lib/clipboard.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/clipboard.ts)

- `copyToClipboard(text, lang)`: 复制文本到剪贴板（含反馈 Toast）

### 7.6 UI 错误处理

**路径**: [src/lib/ui-errors.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/ui-errors.ts)

- `errorMessage(err, lang, fallback?)`: 统一错误消息格式化

### 7.7 类型定义

**路径**: [src/lib/types/index.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/types/index.ts)

完整的 TypeScript 类型定义，涵盖：
- `UploadResult`, `StorageOption`
- `AdminStatus`, `AdminImage`, `AdminIPBan`
- `AdminAbuseOverview`, `AdminAbuseIPRankItem`, `AdminAbuseTokenRankItem`
- `AdminConfig`, `StorageInstance`, `RuntimeSettings`
- `PublicRuntimeSettings`, `AdminSystemSettings`
- `Announcement`, `UploadHistoryRecord`
- `Language`, `Theme`, `ViewMode`

## 8. 构建配置

### SvelteKit 配置

**路径**: [svelte.config.js](file:///d:/Works/MyProject/OmePic/frontend/svelte.config.js)

使用 `@sveltejs/adapter-static` 进行静态导出：
```javascript
adapter: adapter({
    pages: 'out',     // 导出到 frontend/out/
    assets: 'out',
    fallback: 'index.html',  // SPA 降级
})
```

### 构建脚本

| 命令 | 说明 |
|------|------|
| `npm run dev` | 开发服务器（port 3000） |
| `npm run build` | Vite 构建 |
| `npm run build:backend` | 构建并复制到 `backend/web/` |
| `npm run lint` | ESLint 检查 |
| `npm run typecheck` | TypeScript 类型检查 |
| `npm run test` | Vitest 测试 |
