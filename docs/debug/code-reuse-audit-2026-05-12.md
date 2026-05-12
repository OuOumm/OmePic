# 当前项目代码复用情况检查报告

- 日期：2026-05-12
- 范围：`backend/`、`frontend/src/`、`frontend/scripts/` 中已纳入 Git 的源码。
- 排除：`node_modules`、`frontend/out`、`backend/web` 等依赖与构建产物。
- 方法：基于 `git ls-files` 做源码范围枚举，使用归一化重复块扫描、关键模式搜索与人工阅读确认。
- 说明：本次仅生成整改报告，未修改业务代码。当前工作区已有 `backend/internal/http/router/frontend.go` 与 `backend/internal/http/router/frontend_test.go` 处于修改状态，本报告基于当前工作区内容分析。

## 总体结论

项目已经有一定公共层（如 `frontend/src/lib/utils.ts`、`frontend/src/lib/components/studio/*`、`backend/internal/response`、仓储扫描函数等），但仍存在以下复用问题：

1. 后端若干关键逻辑存在“双实现”，尤其是 IP 哈希/脱敏、API 路由清单、错误映射与存储配置字段转换，后续容易出现安全或行为漂移。
2. 前端管理页面存在大量相似的异步动作、弹窗外壳、表格外壳、复制到剪贴板与图片 URL 拼接逻辑，适合分层抽象为轻量工具或基础组件。
3. 部分已有工具/组件未被充分复用，例如 `getImageUrl`、`response.Success/Error`、`ImageDataTable`、`RuntimeStorageUpdate` 等。
4. 测试代码中也有重复 helper，可优先抽到 `_test.go` 测试工具文件，降低新增测试成本。

建议先处理会导致行为不一致的复用点（API 路由清单、IP 工具、错误映射、图片 URL/复制工具），再推进 UI 组件级抽象。

---

## 一、后端复用问题

### BE-01：前端 fallback 的 API 路由清单与真实路由注册重复维护

- 发现位置：
  - 真实路由注册：`backend/internal/http/router/router.go:63-95`
  - fallback API 404 判定：`backend/internal/http/router/frontend.go:110-162`
  - 相关测试：`backend/internal/http/router/frontend_test.go:68-145`
- 可复用代码/具体情况：
  - `router.go` 中已经声明了所有 API/admin 路由；`frontend.go` 又手动维护了一套 `shouldKeepAsAPI404`、`isAdminConfigMutation`、`isAdminAnnouncementRoute` 规则。
  - 这属于“已有实现未被复用”：真实路由表与 fallback 保护规则是同一事实来源，但当前分散在两处。
- 建议整改：
  - 抽出共享路由元数据，例如 `apiRoutePatterns` / `adminRoutePatterns`，由 `router.New` 注册路由和 `shouldKeepAsAPI404` 同时消费。
  - 如果直接共享 Gin 注册函数成本较高，至少将 path/method 常量集中到一个文件，避免新增 API 时漏改 fallback。
  - 为动态路由（如 `/admin/config/storage-instances/:storageKey`、`/admin/announcements/:id`）保留统一匹配函数，测试覆盖 GET/POST/PUT/DELETE/HEAD。
- 影响范围：
  - API 404 行为、SPA fallback、静态前端托管行为。
  - 新增 admin/API 路由时影响最大；漏改会导致 API 请求被前端 `index.html` shadow，或前端页面被错误返回 API 404。
- 优先级：高。

### BE-02：IP 哈希与 IP 脱敏逻辑在 repository 与 service 中重复实现

- 发现位置：
  - `backend/internal/service/ip_utils.go:11-29`
  - `backend/internal/repository/repository.go:1339-1357`
  - 调用点示例：`backend/internal/service/admin_service.go:198`、`backend/internal/service/admin_service.go:252`、`backend/internal/service/image_service.go:533`、`backend/internal/repository/repository.go:621`、`backend/internal/repository/repository.go:690`、`backend/internal/repository/repository.go:745`
- 可复用代码/具体情况：
  - `ipHash` 与 `ipHashValue` 都是 `strings.TrimSpace` 后做 SHA-256，再 hex 编码。
  - `maskIPAddress` 与 `maskIPValue` 都是 IPv4 保留前三段、IPv6 保留前两段。
  - 该逻辑涉及 IP-ban 去重、上传拦截、管理端展示与隐私脱敏，属于安全/隐私敏感逻辑。
- 建议整改：
  - 新建独立包，例如 `backend/internal/iputil`，提供 `Hash(ip string) string`、`Mask(ip string) string`。
  - `repository` 不应反向依赖 `service`，因此不建议直接复用 `service/ip_utils.go`；应把工具下沉到 service/repository 都可依赖的位置。
  - 添加单元测试覆盖：空字符串、带空格、IPv4、IPv6、非法 IP、IPv4-mapped IPv6。
- 影响范围：
  - IP ban 创建/查询、上传/删除拦截、abuse overview、IP detail、admin 图片列表。
  - 抽象后需确认现有哈希值算法不变，否则会影响已保存 ban 记录匹配。
- 优先级：高。

### BE-03：HTTP 错误映射和错误消息清洗在多个 handler 中重复/分叉

- 发现位置：
  - Image handler 错误映射：`backend/internal/http/handler/image_handler.go:142-163`
  - Image handler 消息清洗：`backend/internal/http/handler/image_handler.go:199-206`
  - Admin handler 错误映射：`backend/internal/http/handler/admin_handler.go:252-269`
  - Admin handler 消息清洗：`backend/internal/http/handler/admin_handler.go:271-281`
  - Announcement handler 错误映射：`backend/internal/http/handler/announcement_handler.go:102-115`
  - Health handler 手写响应体：`backend/internal/http/handler/health_handler.go:21-40`
  - 公共响应工具：`backend/internal/response/response.go:10-25`
- 可复用代码/具体情况：
  - `ErrInvalidInput`、`ErrNotFound`、`ErrDependencyUnavailable` 等服务错误到 HTTP status/code 的映射在多个 handler 中重复。
  - `sanitizeMessage` 与 `sanitizeAdminMessage` 基本相同，仅空消息 fallback 略有差异。
  - `HealthHandler` 未复用 `response.Success/Error`，直接手写相同 JSON envelope。
- 建议整改：
  - 在 `handler` 或 `response` 层抽出统一 `ServiceErrorResponder`，支持 endpoint-specific override（如 `ErrMissingToken`、`ErrIPBanned`、admin forbidden 文案）。
  - 合并消息清洗函数，保留“默认 fallback 文案”参数。
  - `HealthHandler` 改为复用 `response.Success/Error`，避免 envelope 结构漂移。
- 影响范围：
  - 所有 API 错误响应结构、日志输出、前端 `ApiError` 解析。
  - 需要回归测试错误 code/status，尤其是上传、删除、admin 登录、公告管理。
- 优先级：中高。

### BE-04：repository 中 RowsAffected/NoRows、SQL 列清单和批量扫描模式重复

- 发现位置：
  - RowsAffected 判定重复：`backend/internal/repository/repository.go:340-347`、`backend/internal/repository/repository.go:355-362`、`backend/internal/repository/repository.go:384-390`、`backend/internal/repository/repository.go:454-460`、`backend/internal/repository/repository.go:610-616`、`backend/internal/repository/repository.go:865-871`、`backend/internal/repository/repository.go:880-886`、`backend/internal/repository/repository.go:901-907`
  - image select 列清单重复：`backend/internal/repository/repository.go:423`、`backend/internal/repository/repository.go:432`、`backend/internal/repository/repository.go:441`、`backend/internal/repository/repository.go:483`、`backend/internal/repository/repository.go:492`、`backend/internal/repository/repository.go:521`
  - scan slice 模式：`backend/internal/repository/repository.go:934-943`、`backend/internal/repository/repository.go:970-979`、`backend/internal/repository/repository.go:1010-1019`
- 可复用代码/具体情况：
  - 多处 `RowsAffected() == 0 -> sql.ErrNoRows` 完全一致。
  - `SELECT id, uid, token, storage_key...` 列清单多次手写，后续新增字段时容易漏改某个查询。
  - scan slice 模式相似，但 Go 泛型化收益有限，可按需处理。
- 建议整改：
  - 抽 `ensureRowsAffected(result sql.Result) error` 或 `execExpectAffected(ctx, query, args...)`。
  - 抽 `const imageColumns = "id, uid, token, ..."`，查询通过 `SELECT ` + imageColumns + ` FROM images ...` 复用。
  - scan slice 可先保留，或在 Go 1.25 下用泛型 helper `scanRows[T]`，但应评估可读性。
- 影响范围：
  - repository 内部 CRUD、迁移兼容测试、admin 图片/公告/IP ban 管理。
  - 抽象主要影响仓储层，业务行为应保持不变。
- 优先级：中。

### BE-05：storage config 插入 SQL 与字段转换重复

- 发现位置：
  - 单条插入：`backend/internal/repository/repository.go:288-312`
  - 批量 seed 插入：`backend/internal/repository/repository.go:1075-1111`
  - storage config 扫描：`backend/internal/repository/repository.go:1022-1053`
- 可复用代码/具体情况：
  - `CreateStorageConfig` 与 `insertStorageConfigs` 使用相同 INSERT 列、相同参数顺序、相同 bool/string 转换。
  - `CreateStorageConfig` 中 `created_at` 和 `updated_at` 分别调用 `time.Now().UTC().Format(...)`，理论上可能出现极小时间差；批量插入则已经共用 `now`。
- 建议整改：
  - 抽 `storageConfigInsertSQL`、`storageConfigInsertArgs(cfg, now)`。
  - 为 `*sql.DB` 与 `*sql.Tx` 抽最小执行接口：`type execContexter interface { ExecContext(context.Context, string, ...any) (sql.Result, error) }`。
  - 单条与批量插入都调用同一个 helper，并共用同一个 `now`。
- 影响范围：
  - 存储实例创建、启动时 storage catalog 初始化/迁移。
  - 需覆盖 repository storage catalog 相关测试。
- 优先级：中。

### BE-06：storage config DTO/字段列表跨层重复，且 `RuntimeStorageUpdate` 未被复用

- 发现位置：
  - 核心配置：`backend/internal/config/config.go:47-84`
  - admin service DTO：`backend/internal/service/admin_service.go:16-87`
  - 前端类型：`frontend/src/lib/types/index.ts:111-132`
  - `RuntimeStorageUpdate` 仅定义未检索到实际使用：`backend/internal/config/config.go:70-84`
- 可复用代码/具体情况：
  - `RuntimeStorageConfig`、`AdminStorageConfigView`、`AdminConfigUpdateInput`、`AdminStorageConfigCreateInput`、`AdminStorageConfigUpdateInput`、前端 `StorageInstance` 均维护一套相似字段。
  - `RuntimeStorageUpdate` 与 `AdminStorageConfigUpdateInput` 字段几乎重合，但当前没有复用。
  - 字段重复会增加新增后端支持项时的漏改概率，例如新存储后端、新 S3 参数、新 WebDAV 参数。
- 建议整改：
  - 后端优先复用 `config.RuntimeStorageUpdate`：可使用类型别名或嵌入，避免同一 update payload 重复定义。
  - 对 create/view 继续保留独立 DTO 时，应集中 mapper：`AdminStorageConfigViewFromRuntimeConfig`、`RuntimeStorageConfigFromCreateInput`。
  - 前端 `StorageInstance` 可保留，但建议通过 OpenAPI/代码生成或至少在同一报告任务中维护字段对照测试/契约测试。
  - 注意 view 层必须继续做 secret masking，不能直接裸返回 `RuntimeStorageConfig`。
- 影响范围：
  - admin settings storage 管理、storage manager reload、前后端 JSON 契约。
  - 需要重点验证敏感字段不泄露。
- 优先级：中。

### BE-07：bool/CSV/后端名称归一化工具分散且语义略有差异

- 发现位置：
  - `backend/internal/repository/repository.go:1367-1383`
  - `backend/internal/service/runtime_settings.go:421-444`
  - `backend/internal/config/config.go:180-202`
  - `backend/internal/config/config.go:164-170`
  - `backend/internal/service/admin_service.go:774-780`
  - `backend/internal/storage/storage.go:352-369`
- 可复用代码/具体情况：
  - bool string 化/解析有 `boolString`、`parseBool`、`boolStringValue`、`parseBoolValue`、`envBool` 多套实现。
  - CSV split 在 `config` 与 `runtime_settings` 中重复，但一个过滤空值，一个保留后续再过滤，语义不同。
  - storage backend 归一化在 config/service/storage 中多次 `TrimSpace + ToLower`。
- 建议整改：
  - 不建议一次性做过度抽象；可先抽最稳定的 bool parse/string，例如 `internal/textconv.BoolString`、`ParseBoolLoose`。
  - 将 storage backend 归一化提升为 `config.NormalizeStorageBackend`，让 `admin_service` 与 `storage` 复用。
  - CSV helper 如要抽，应显式命名语义：`SplitCSVTrimmedKeepEmpty` / `SplitCSVTrimmedNonEmpty`。
- 影响范围：
  - 配置加载、运行时设置保存、存储后端匹配、迁移旧配置。
  - 因语义有差异，整改时应逐处确认不改变默认值与空值处理。
- 优先级：中低。

### BE-08：测试 helper 与测试 fixture 重复

- 发现位置：
  - PNG 生成 helper：`backend/internal/http/handler/image_handler_test.go:272-287`、`backend/internal/service/image_service_test.go:1022-1037`
  - AVIF 生成与 PNG 生成内层填充逻辑相似：`backend/internal/service/image_service_test.go:1039-1057`
  - discard writer：`backend/internal/http/handler/image_handler_test.go:266-270`、`backend/internal/service/image_service_test.go:1016-1020`
  - service 测试 harness：`backend/internal/service/admin_service_test.go:315-353`、`backend/internal/service/image_service_test.go:964-1013`
  - S3 secondary config fixture 多处重复：`backend/internal/service/admin_service_test.go:20-30`、`backend/internal/service/admin_service_test.go:65-75`、`backend/internal/service/admin_service_test.go:92-102`、`backend/internal/service/admin_service_test.go:281-291`、`backend/internal/service/image_service_test.go:887-897`
- 可复用代码/具体情况：
  - 测试里重复创建 temp repository、迁移、初始化 storage catalog、构造 logger、构造图片字节。
  - `admin_service_test.go` 使用了同包其他 `_test.go` 中的 `ioDiscard`，存在隐式测试文件耦合。
- 建议整改：
  - 在对应包内新增 `test_helpers_test.go`，集中放 `discardWriter`、`mustPNGBytes`、`mustAVIFBytes`、`newTestRepository`、`newTestStorageManager`、`testS3StorageConfig`。
  - 跨 package helper 如确实需要，可放 `backend/internal/testutil`，但要避免让生产代码依赖测试工具。
- 影响范围：
  - 仅测试代码；可读性和新增测试效率提升，业务行为风险低。
- 优先级：低到中。

---

## 二、前端复用问题

### FE-01：复制到剪贴板逻辑重复，且未统一处理失败场景

- 发现位置：
  - `frontend/src/lib/components/studio/ImageDetailDrawer.svelte:90-93`
  - `frontend/src/routes/+page.svelte:242-245`
  - `frontend/src/routes/api/+page.svelte:32-35`
  - `frontend/src/routes/history/+page.svelte:116-119`
- 可复用代码/具体情况：
  - 多处都是 `navigator.clipboard.writeText(value); toast.success(t(..., 'common.copied'))`。
  - 当前没有统一处理 `navigator.clipboard` 不可用、权限失败、非安全上下文失败等情况。
- 建议整改：
  - 新增 `frontend/src/lib/clipboard.ts` 或 `frontend/src/lib/ui-actions.ts`：`copyToClipboard(value, language): Promise<boolean>`。
  - 工具内部统一做：浏览器环境判断、`await navigator.clipboard.writeText`、成功/失败 toast、fallback 文案。
  - 组件中只保留 `onclick={() => copyToClipboard(value, preferences.language)}`。
- 影响范围：
  - 上传首页、历史页、API 页、admin 图片详情抽屉。
  - 用户可见行为：复制失败时将从静默失败变为错误提示。
- 优先级：高。

### FE-02：已有 `getImageUrl`/图片路由 helper 未被复用，图片 URL 拼接分散

- 发现位置：
  - 既有工具：`frontend/src/lib/utils.ts:44-47`、`frontend/src/lib/utils.ts:116-117`
  - 手写路径：`frontend/src/lib/components/studio/ImageDetailDrawer.svelte:31`
  - 手写路径：`frontend/src/routes/admin/dashboard/images/+page.svelte:145`
  - 手写路径：`frontend/src/routes/admin/dashboard/images/+page.svelte:153`
  - API delete path：`frontend/src/lib/api.ts:149-153`
- 可复用代码/具体情况：
  - `utils.ts` 已有 `getImageUrl(uid)` 和 `getAbsoluteImageUrl(uid)`，但 admin 图片预览/打开链接仍手写 `/i/${uid}.avif`。
  - `api.ts` 的 delete endpoint 也手写 `/i/${uid}.avif`，虽然它作为 API path 不能直接复用带 base 的 `getImageUrl`。
- 建议整改：
  - 抽更底层的 `getImagePath(uid): string`，返回 `/i/${uid}.avif`。
  - `getImageUrl` 基于 `getImagePath` 加 API base；`api.deleteImageByUid`、admin 图片列表、详情抽屉都复用 `getImagePath` 或 `getImageUrl`。
  - 如远程 API base 场景需要跨域展示图片，应明确组件使用 `getImageUrl` 而不是相对路径。
- 影响范围：
  - 所有图片预览、打开、删除 API route 构造。
  - 后续若图片扩展名或路由改变，只需改一处。
- 优先级：高。

### FE-03：admin 页面异步动作模式重复，错误提示与 busy/loading 状态处理不一致

- 发现位置示例：
  - `frontend/src/lib/components/studio/AnnouncementManager.svelte:54-66`、`frontend/src/lib/components/studio/AnnouncementManager.svelte:94-110`、`frontend/src/lib/components/studio/AnnouncementManager.svelte:113-126`
  - `frontend/src/lib/components/studio/StorageInstanceManager.svelte:93-108`、`frontend/src/lib/components/studio/StorageInstanceManager.svelte:110-136`
  - `frontend/src/lib/components/studio/IPDetailPanel.svelte:26-43`、`frontend/src/lib/components/studio/IPDetailPanel.svelte:50-79`
  - `frontend/src/lib/components/studio/ImageDetailDrawer.svelte:55-73`、`frontend/src/lib/components/studio/ImageDetailDrawer.svelte:95-134`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:41-83`
  - `frontend/src/routes/admin/dashboard/security/+page.svelte:28-49`、`frontend/src/routes/admin/dashboard/security/+page.svelte:69-99`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte:27-59`
- 可复用代码/具体情况：
  - 重复结构：检查 `preferences.adminToken`、设置 busy/loading、try 调 API、toast success、catch `err instanceof Error ? err.message : t(...common.error)`、finally 重置。
  - 部分动作没有 catch/finally 或 busy 保护，例如 `security/+page.svelte:101-115` 的 unban/purgeImages，失败时可能出现未处理错误且确认弹窗状态不一致。
- 建议整改：
  - 先抽小工具而不是复杂状态框架：
    - `errorMessage(err, language, fallbackKey = 'common.error')`
    - `toastApiError(err, language, fallbackKey?)`
    - `runAsyncAction({ setBusy, action, onSuccess, successMessage, fallbackErrorKey })`
  - 对带 AbortSignal 的 loader，可抽 `runAbortableLoad` 或统一封装 abort error 判断。
  - 保持 Svelte runes 状态由调用方传 setter，避免工具直接持有组件状态。
- 影响范围：
  - admin dashboard 图片、security、settings、公告、存储实例、IP detail。
  - 可减少错误处理漏网，但需要逐个动作确认 success toast、reload、关闭弹窗时机。
- 优先级：高。

### FE-04：AbortController + `$effect` 加载模式重复

- 发现位置：
  - `frontend/src/lib/components/studio/AnnouncementManager.svelte:128-132`
  - `frontend/src/lib/components/studio/IPDetailPanel.svelte:81-85`
  - `frontend/src/routes/+page.svelte:263-285`
  - `frontend/src/routes/admin/dashboard/+layout.svelte:47-50`
  - `frontend/src/routes/admin/dashboard/+page.svelte:47-51`
  - `frontend/src/routes/admin/dashboard/images/+page.svelte:104-108`
  - `frontend/src/routes/admin/dashboard/security/+page.svelte:117-121`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte:61-65`
- 可复用代码/具体情况：
  - 多处都是 `$effect(() => { const controller = new AbortController(); void load(controller.signal); return () => controller.abort(); })`。
- 建议整改：
  - 可抽 `abortableEffect((signal) => load(signal))` 工具函数，返回 Svelte `$effect` 的 cleanup。
  - 对有额外事件监听的页面（如上传页 paste handler）仍保留本地逻辑，或只抽 controller 部分。
- 影响范围：
  - 页面加载、路由切换时取消请求的行为。
  - 需要确认 Svelte 5 runes 下工具调用不会破坏依赖追踪。
- 优先级：中。

### FE-05：图片预览/详情中的线性导航和方向键逻辑重复

- 发现位置：
  - 上传历史预览：`frontend/src/lib/components/studio/ImagePreviewDialog.svelte:25-43`
  - admin 图片详情：`frontend/src/lib/components/studio/ImageDetailDrawer.svelte:35-53`
  - 已复用按钮组件：`frontend/src/lib/components/studio/ImageSwitchButton.svelte`
- 可复用代码/具体情况：
  - 两个组件都计算 `currentIndex`、`hasNavigation`、previous/next，并处理 ArrowLeft/ArrowRight。
  - 数据类型不同（`UploadHistoryRecord` vs `AdminImage`），但都以 `uid` 定位。
- 建议整改：
  - 抽泛型 helper：`createLinearNavigation(items, current, getKey, onNavigate)`，返回 previous/next/handleKeydown。
  - 或抽更窄的 `handleArrowNavigation(event, previous, next, navigateTo)`。
  - 保留 `ImageSwitchButton` 现有复用。
- 影响范围：
  - 上传历史图片预览、admin 图片详情抽屉。
  - 需要验证键盘导航、disabled 状态、图片加载状态重置。
- 优先级：中。

### FE-06：弹窗/抽屉外壳重复，`accessibleDialog` 已复用但 modal shell 未组件化

- 发现位置：
  - 通用确认弹窗：`frontend/src/lib/components/studio/ConfirmDialog.svelte:25-46`
  - 公告详情/编辑弹窗：`frontend/src/lib/components/studio/AnnouncementManager.svelte:187-208`
  - 存储实例编辑弹窗：`frontend/src/lib/components/studio/StorageInstanceManager.svelte:177-180`
  - admin 图片详情抽屉：`frontend/src/lib/components/studio/ImageDetailDrawer.svelte:137-143`
  - IP detail 弹窗：`frontend/src/lib/components/studio/IPDetailPanel.svelte:88-97`
  - 上传历史图片预览：`frontend/src/lib/components/studio/ImagePreviewDialog.svelte:50-53`
- 可复用代码/具体情况：
  - overlay、panel、ARIA attributes、close button、backdrop click、`accessibleDialog` 使用方式高度相似。
  - 各组件手写 z-index、padding、panel class，后续可访问性修复需要改多处。
- 建议整改：
  - 新增 `ModalShell.svelte` / `StudioDialog.svelte`：接收 `open`、`titleId`/`ariaLabel`、`onClose`、`size`、`zIndex`、`closeLabel`、`busy`、`panelClass`，通过 snippet/render children 渲染内容。
  - `ConfirmDialog` 可逐步迁移到同一个 shell，保证视觉和 a11y 一致。
  - 不建议一次性把所有业务内容也抽走；先抽外壳。
- 影响范围：
  - 所有弹窗/抽屉交互、焦点管理、关闭行为、移动端布局。
  - 需要做浏览器回归或 Playwright 交互验证。
- 优先级：中。

### FE-07：表格外壳与行样式重复，已有 `ImageDataTable` 未覆盖 admin 图片表

- 发现位置：
  - 已有上传历史表：`frontend/src/lib/components/studio/ImageDataTable.svelte:24-82`
  - admin 图片表手写：`frontend/src/routes/admin/dashboard/images/+page.svelte:124-158`
  - 公告表：`frontend/src/lib/components/studio/AnnouncementManager.svelte:149-183`
  - 存储实例表：`frontend/src/lib/components/studio/StorageInstanceManager.svelte:145-174`
  - security top IP / ban 表：`frontend/src/routes/admin/dashboard/security/+page.svelte:187-212`、`frontend/src/routes/admin/dashboard/security/+page.svelte:221-247`
- 可复用代码/具体情况：
  - 多处重复 table wrapper、thead class、`studio-table-row`、横向滚动容器、actions 列。
  - `ImageDataTable` 已解决上传记录表的 preview/copy/delete/select，但 admin 图片表因类型和动作不同重新实现了一张相似表。
- 建议整改：
  - 短期抽 `StudioTable.svelte` 或 class 常量，统一 wrapper/header/empty 状态。
  - 中期考虑让 `ImageDataTable` 支持 adapter：`getUid`、`getImageUrl`、`getTitle`、`actions` snippet，使它也能承载 `AdminImage`。
  - 如果泛化过度影响可读性，可只抽 `TableShell` + `ActionCell`。
- 影响范围：
  - admin 图片、历史、公告、存储、安全页面的表格布局与可访问性。
  - 泛化 `ImageDataTable` 风险中等，需覆盖选择、预览、删除、复制动作。
- 优先级：中。

### FE-08：详情字段行和 copyable row 在 `ImageDetailDrawer` 内高度重复

- 发现位置：
  - `frontend/src/lib/components/studio/ImageDetailDrawer.svelte:157-168`
- 可复用代码/具体情况：
  - 多个 `<div><dt><dd>...</dd></div>` 使用相同 grid/border class。
  - UID、URL、MD5、Token 四行还有相同 copy button 结构。
- 建议整改：
  - 抽 `DetailRow.svelte` 和 `CopyableDetailRow.svelte`，或用本地数组驱动渲染。
  - 若后续 IP detail、storage detail 也使用相同样式，可放入 `components/studio` 公共组件。
- 影响范围：
  - admin 图片详情抽屉 UI；风险低。
- 优先级：低到中。

### FE-09：localStorage 安全读写 helper 没有统一暴露，公告已读状态直接访问 storage

- 发现位置：
  - 私有 JSON helper：`frontend/src/lib/stores/preferences.svelte.ts:22-35`
  - client token 直接访问：`frontend/src/lib/client-token.ts:21-29`
  - 公告已读状态直接访问：`frontend/src/routes/+page.svelte:94`、`frontend/src/routes/+page.svelte:111`
  - 主题首屏脚本直接访问：`frontend/src/lib/components/studio/AppShell.svelte:43-56`
  - 既有主题解析工具：`frontend/src/lib/utils.ts:79-88`
- 可复用代码/具体情况：
  - `preferences.svelte.ts` 已有安全 `readJSON/writeJSON`，但作用域私有，其他模块无法复用。
  - 公告 lastSeen、client token、主题脚本都手写 localStorage 访问。
  - AppShell 的内联脚本有首屏防闪烁需求，不能简单 import 工具，但主题解析规则仍与 `getInitialThemeScriptTheme` 存在重复。
- 建议整改：
  - 新增 `browser-storage.ts`：`safeGetItem`、`safeSetItem`、`readJSON`、`writeJSON`，集中 SSR/try-catch。
  - 将 storage key 常量集中定义，避免字符串散落。
  - AppShell 内联脚本可继续保留，但建议由同一规则生成字符串或在注释/测试中确保与 `getInitialThemeScriptTheme` 一致。
- 影响范围：
  - UI preferences、client token、公告已读状态、主题首屏体验。
  - 需要注意不要把 admin token 持久化；当前 in-memory admin token 设计应保持。
- 优先级：中。

### FE-10：storage instance 表单默认值与 payload 构造可抽为工具

- 发现位置：
  - 默认对象：`frontend/src/lib/components/studio/StorageInstanceManager.svelte:22-38`
  - payload 构造：`frontend/src/lib/components/studio/StorageInstanceManager.svelte:66-90`
  - 类型定义：`frontend/src/lib/types/index.ts:111-127`
- 可复用代码/具体情况：
  - storage instance 默认值和按 backend 裁剪 payload 的逻辑都在组件内部，后续若其他页面/测试也需要创建存储实例，会复制这段逻辑。
  - 该逻辑与后端 `AdminStorageConfigCreateInput` / `AdminStorageConfigUpdateInput` 字段强相关。
- 建议整改：
  - 抽到 `frontend/src/lib/storage-instance.ts`：`createBlankStorageInstance()`、`storageInstancePayload(form, editingKey)`、`isStorageInstanceValid(form)`。
  - 组件只负责 UI 状态和调用 API。
- 影响范围：
  - admin settings/storage 实例创建编辑。
  - 后续新增存储后端字段时可集中修改。
- 优先级：中低。

### FE-11：runtime/system settings 的加载与保存流程在 settings 与 security/rate-limit 中重复

- 发现位置：
  - settings 加载/保存：`frontend/src/routes/admin/dashboard/settings/+page.svelte:27-59`
  - security rate-limit 加载/保存：`frontend/src/routes/admin/dashboard/security/+page.svelte:41-84`
- 可复用代码/具体情况：
  - 两处都调用 `adminGetSystemSettings`、修改 `system.runtime`、调用 `adminUpdateSystemSettings`、toast 成功/失败。
  - security 页面额外做 rate limit 数值归一化，settings 页面额外做 MIME types 文本解析。
- 建议整改：
  - 抽 `useAdminSystemSettings` 风格的轻量 helper/store，提供 `loadSystemSettings`、`saveSystemSettings(runtime)`、统一错误提示。
  - 各页面保留各自的表单归一化逻辑。
- 影响范围：
  - admin runtime settings 与 rate limit 页面。
  - 可减少 API 错误处理重复，但需避免不同 tab 加载条件互相影响。
- 优先级：中低。

---

## 三、建议整改顺序

### 第一阶段：低耦合、高收益

1. 抽后端 IP 工具包，替换 repository/service 双实现。
2. 抽前端 clipboard helper，统一复制成功/失败提示。
3. 抽 `getImagePath`，让图片展示、打开、删除路径统一来源。
4. 让 `HealthHandler` 复用 `response.Success/Error`，合并 handler 消息清洗函数。
5. 将测试图片生成、discard writer、常用 storage fixture 抽到测试 helper。

### 第二阶段：行为一致性与维护成本

1. 统一后端 router API 路由元数据，消除 fallback 与真实路由重复清单。
2. 抽 repository `RowsAffected` helper、image column 常量、storage config insert helper。
3. 复用或删除未使用的 `RuntimeStorageUpdate`，集中 storage config DTO mapper。
4. 前端抽 admin async action/error toast helper。

### 第三阶段：UI 组件化与结构优化

1. 抽 `StudioDialog/ModalShell`，逐步迁移自定义弹窗。
2. 抽 `StudioTable`/`TableShell`，再评估是否泛化 `ImageDataTable`。
3. 抽 `DetailRow/CopyableDetailRow`、线性导航 helper。
4. 抽 browser storage helper 与 system settings helper。

---

## 四、整改验证建议

- 后端：
  - `go test ./...`。
  - 补充 IP 工具测试，确认哈希算法与脱敏输出不变。
  - 对 fallback/API 404 路由增加表驱动测试，覆盖新增共享路由元数据。
  - 对 storage config 插入/迁移测试确认 `created_at/updated_at`、bool 字段、默认 storage 行为不变。
- 前端：
  - `npm run typecheck`、`npm run lint`、`npm run test`（在 `frontend/` 下）。
  - 对 `getImagePath/getImageUrl`、clipboard helper、browser storage helper 增加单元测试。
  - 对 dialog/table 抽象做浏览器交互回归：打开、关闭、Esc、backdrop click、焦点、移动端横向滚动。

---

## 五、暂不建议立即抽象的内容

- 单文件内少量重复 Tailwind class，如表单 label class，若没有跨组件维护痛点，可等 `StudioFormField` 组件出现实际需求后再抽。
- repository 的所有 scan slice 函数不必强行泛型化；只有在新增更多表模型后，泛型 helper 的收益才明显。
- 前端 admin async action 不宜一次性抽成复杂状态机；先抽错误消息/clipboard/URL 等稳定小工具，再根据重复形态演进。
