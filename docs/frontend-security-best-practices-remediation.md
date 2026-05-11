# 前端安全最佳实践审查与整改建议

审查日期：2026-05-11  
审查范围：`frontend/` SvelteKit + Svelte 5 + TypeScript 前端项目  
参考规范：`security-best-practices` 技能中的 JavaScript/TypeScript Web Frontend Security Spec、项目现有 Trellis 前端质量约定

## 执行摘要

本次审查未发现已证实的 Critical 级 DOM XSS、任意代码执行、`postMessage` 滥用或明文硬编码密钥问题。项目在公告 Markdown 展示处避免了 `{@html}`，并使用 tokenized 渲染方式输出文本节点，这是较好的 XSS 防护实践。

主要整改重点集中在三类：

1. 管理员 JWT 当前存储在浏览器可读存储中，若未来出现任意 XSS，会扩大管理端账号接管风险。
2. 生产安全响应头（尤其 CSP、`frame-ancestors`、`X-Content-Type-Options`、`Referrer-Policy`、`Permissions-Policy`）在前端与后端代码中未见显式配置，需要在 Go 后端或部署层补齐。
3. 上传历史中持久化服务端返回的 `url`、`markdown`、`bbcode`，这些 IndexedDB 数据被当作后续 DOM URL 使用，需增加 URL 协议/同源校验。客户端 `client_token` 是匿名删除凭证且必须随历史记录保存，以确保用户重置当前 Token 后仍可删除旧图。

## 审查依据与方法

重点检索与人工核查了以下类别：

- DOM XSS sink：`innerHTML`、`outerHTML`、`insertAdjacentHTML`、Svelte `{@html}`、`document.write`。
- 字符串执行：`eval`、`new Function`、字符串形式 `setTimeout` / `setInterval`。
- 事件处理器字符串注入：`setAttribute("on...")`、内联事件属性。
- URL 与导航 sink：`window.location`、`location.href`、动态 `href` / `src`。
- 浏览器存储：`localStorage`、IndexedDB、Token / session / auth 类 key。
- CSP 与安全响应头：`Content-Security-Policy`、Trusted Types、frame 防护、MIME sniffing 防护。
- 第三方脚本、远程资源、远程图片下载流程。

## 高风险发现

### FE-SEC-001：管理员 JWT 保存在 `localStorage`，XSS 后可被直接读取

- Rule ID：JS-STORAGE-001
- Severity：High
- 位置：
  - [preferences.svelte.ts:L4-L7](file:///d:/Works/MyProject/OmePic/frontend/src/lib/stores/preferences.svelte.ts#L4-L7)
  - [preferences.svelte.ts:L36-L45](file:///d:/Works/MyProject/OmePic/frontend/src/lib/stores/preferences.svelte.ts#L36-L45)
  - [preferences.svelte.ts:L70-L78](file:///d:/Works/MyProject/OmePic/frontend/src/lib/stores/preferences.svelte.ts#L70-L78)
  - [preferences.ts:L89-L109](file:///d:/Works/MyProject/OmePic/frontend/src/lib/preferences.ts#L89-L109)
- 证据：

```ts
const ADMIN_TOKEN_KEY = 'omepic-admin-token';
const adminPrefs = readJSON(ADMIN_TOKEN_KEY, { token: null as string | null });
adminToken: adminPrefs.token ?? null,

export function setAdminToken(token: string) {
  preferences.adminToken = token;
  writeJSON(ADMIN_TOKEN_KEY, { token });
}
```

以及旧 helper 中同名 key：

```ts
export function getAdminToken(): string | null {
  const raw = localStorage.getItem(ADMIN_TOKEN_KEY);
  if (raw) {
    const data = JSON.parse(raw);
    return data.token ?? null;
  }
  return null;
}

export function setAdminToken(token: string) {
  localStorage.setItem(ADMIN_TOKEN_KEY, JSON.stringify({ token }));
}
```

- 影响：一旦站点存在任何 XSS 或恶意同源脚本，攻击者可读取管理员 JWT，并调用图片删除、封禁 IP、存储配置、公告管理等管理端接口。
- 整改建议：
  1. 首选将管理端会话迁移为后端设置的 `HttpOnly` Cookie，会话校验由后端完成，前端不再接触 JWT 明文。
  2. 如果短期无法改后端，至少改为内存态保存，并缩短 JWT 有效期，刷新页面后要求重新登录。
  3. 配合 FE-SEC-004 增加严格 CSP，降低 XSS 成功率。
  4. 删除或合并 [preferences.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/preferences.ts) 中旧的 admin token helper，避免双实现造成维护遗漏。
- 缓解措施：上线前确认管理端路由仅 HTTPS 可访问；设置短 TTL；后端对高危操作加入二次确认或审计日志。
- 误报说明：如果管理员 JWT 后端已极短期且权限受限，风险可降低，但 JS 可读 Token 的根本风险仍存在。

### FE-SEC-002：客户端删除 Token 随上传历史持久化是产品必需设计，需作为匿名删除凭证进行边界说明

- Rule ID：Product security boundary / JS-STORAGE-001 reviewed exception
- Severity：Informational
- 位置：
  - [preferences.ts:L4-L24](file:///d:/Works/MyProject/OmePic/frontend/src/lib/preferences.ts#L4-L24)
  - [+page.svelte:L127-L154](file:///d:/Works/MyProject/OmePic/frontend/src/routes/+page.svelte#L127-L154)
  - [index.ts:L232-L246](file:///d:/Works/MyProject/OmePic/frontend/src/lib/types/index.ts#L232-L246)
  - [upload-history.ts:L21-L29](file:///d:/Works/MyProject/OmePic/frontend/src/lib/indexeddb/upload-history.ts#L21-L29)
- 证据：

```ts
const TOKEN_STORAGE_KEY = "omepic-client-token";

export function getClientToken(): string {
  let token = localStorage.getItem(TOKEN_STORAGE_KEY);
  if (!token) {
    token = generateToken();
    localStorage.setItem(TOKEN_STORAGE_KEY, token);
  }
  return token;
}
```

上传成功后将当次 token 写入 IndexedDB 记录：

```ts
await saveUploadToHistory({
  uid: result.uid,
  url: result.url,
  markdown: result.markdown,
  bbcode: result.bbcode,
  client_token: token,
  original_filename: task.file.name,
  saved_at: new Date().toISOString(),
});
```

历史删除时当前代码使用当前客户端 token：

```ts
await deleteImageByUid(record.uid, getClientToken());
```

- 影响：根据产品约束，客户端 Token 是匿名访问与删除图片的唯一凭证，不是管理员账号、JWT、刷新令牌或可访问全局资源的登录凭据。将上传时的 `client_token` 随历史记录重复保存是必要设计：用户未来可能重置当前 Token，如果不保存历史记录对应的原 Token，将无法删除旧图。
- 整改建议：
  1. 保留 `client_token` 随上传历史持久化的设计，不再将其列为需要删除的安全问题。
  2. 删除旧图时应优先使用历史记录里的 `record.client_token`，而不是始终使用 `getClientToken()`，以满足“重置当前 Token 后仍可删除旧图”的产品目标。
  3. 在 API / 帮助文案中明确说明：客户端 Token 是匿名删除凭证，清除浏览器数据或历史记录后将失去对应图片的本地删除能力。
  4. 仍建议使用 `crypto.getRandomValues` / `crypto.randomUUID` 生成客户端 token，替代 `Math.random()`，这是凭证随机性加固，不改变持久化策略。
- 缓解措施：提供清空本地历史和重置当前 Token 的用户操作；清空历史时删除 IndexedDB 中对应记录即可。
- 误报说明：此前将其按敏感授权材料列为高风险不符合当前产品安全边界；在本项目语境中它是匿名删除凭证，重复保存是功能正确性要求。

## 中风险发现

### FE-SEC-003：服务端返回 URL 被持久化并直接用于 `img src` / `a href`，缺少前端 URL 协议与来源校验

- Rule ID：JS-URL-002
- Severity：Medium
- 位置：
  - [+page.svelte:L140-L154](file:///d:/Works/MyProject/OmePic/frontend/src/routes/+page.svelte#L140-L154)
  - [ImageDataTable.svelte:L31-L50](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/ImageDataTable.svelte#L31-L50)
  - [ImagePreviewDialog.svelte:L33-L47](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/ImagePreviewDialog.svelte#L33-L47)
- 证据：

```svelte
<img src={record.url} alt={record.original_filename || record.uid} class="h-full w-full object-cover" loading="lazy" />
<a class="studio-button p-2" href={record.url} target="_blank" rel="noreferrer" ...>
```

```svelte
<img src={record.url} alt={filename} class="max-h-[62dvh] max-w-full object-contain" />
<a class="studio-button" href={record.url} download={filename} target="_blank" rel="noreferrer">
```

- 影响：当前 `record.url` 主要来自后端上传响应，正常应为 `/i/{uid}.avif` 或可信公开 URL；但它被持久化在 IndexedDB，任何同源脚本或手动篡改都可使 UI 加载非预期远程资源。对于 `<a href>`，若出现 `javascript:`、`data:` 等 scheme，可能触发 URL 类安全问题；对 `<img>` 则可能造成隐私泄露和内容混淆。
- 整改建议：
  1. 新增统一 URL 规范化函数，例如 `safeImageUrl(value)`：仅允许相对路径、同源 URL、或后端配置的公开图片域名；只允许 `http:` / `https:`。
  2. 渲染 `img src`、`a href` 前统一调用该函数，校验失败时显示占位状态并禁用打开/下载链接。
  3. IndexedDB 读取后对历史记录做轻量 schema 校验；不可信字段不要直接进入 DOM URL context。
  4. 优先持久化 `uid`，展示时重新构造 `/i/${uid}.avif`，减少对持久化 URL 的信任。
- 缓解措施：CSP 增加 `img-src 'self' data: blob: https:` 或更严格的图片域名白名单；`connect-src` 限制 API 域。
- 误报说明：如果后端严格保证 `url` 永远同源且 IndexedDB 仅本地可信，风险较低；但浏览器存储不能作为可信边界。

### FE-SEC-004：未见显式 CSP 与关键安全响应头配置

- Rule ID：JS-CSP-001 / JS-CSP-002 / JS-TT-001
- Severity：Medium
- 位置：
  - [app.html:L1-L11](file:///d:/Works/MyProject/OmePic/frontend/src/app.html#L1-L11)
  - [svelte.config.js:L7-L13](file:///d:/Works/MyProject/OmePic/frontend/svelte.config.js#L7-L13)
  - 全仓库检索未发现 `Content-Security-Policy`、`X-Frame-Options`、`frame-ancestors`、`X-Content-Type-Options`、`Referrer-Policy`、`Permissions-Policy`。
- 证据：

```html
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  %sveltekit.head%
</head>
```

SvelteKit 使用静态适配器输出：

```js
adapter({
  pages: 'out',
  assets: 'out',
  fallback: 'index.html',
  precompress: false,
  strict: true,
})
```

- 影响：如果未来出现 DOM XSS、依赖投毒或不可信内容渲染问题，缺少 CSP 会显著增加脚本执行与数据外传风险；缺少 `frame-ancestors` / `X-Frame-Options` 会增加点击劫持风险；缺少 `X-Content-Type-Options` 会增加 MIME sniffing 风险。
- 整改建议：
  1. 优先在 Go 后端静态资源响应中设置 HTTP Header，而不是仅在 HTML meta 中配置。
  2. 推荐生产基线：

```http
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: https:; font-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

  3. 若 Tailwind/Svelte 运行结果可去除 inline style 依赖，再逐步收紧 `style-src`。
  4. 中长期评估 Trusted Types：`require-trusted-types-for 'script'`，但需要先验证 SvelteKit 与第三方依赖兼容性。
- 缓解措施：若短期只能改 HTML，可添加 meta CSP，但必须了解 meta CSP 不支持 `frame-ancestors`、`report-uri`、`sandbox`，点击劫持仍需后端/部署层 Header。
- 误报说明：安全头可能由生产反向代理、CDN 或容器平台注入；当前仓库内未见配置，建议运行时用浏览器网络面板或 `curl -I` 验证。

### FE-SEC-005：远程 URL 上传由浏览器直接 `fetch` 用户输入 URL，缺少下载体积、协议解析和重定向后的最终 URL 策略

- Rule ID：JS-URL-001 / JS-URL-002
- Severity：Medium
- 位置：[+page.svelte:L175-L200](file:///d:/Works/MyProject/OmePic/frontend/src/routes/+page.svelte#L175-L200)
- 证据：

```ts
const url = urlInput.trim();
if (!/^https?:\/\//i.test(url)) {
  toast.error(t(preferences.language, 'upload.invalidUrl'));
  return;
}
const response = await fetch(url);
if (!response.ok) throw new Error('Download failed');
const blob = await response.blob();
const mimeType = response.headers.get('Content-Type') || blob.type;
if (!mimeType.startsWith('image/')) {
  toast.error(t(preferences.language, 'upload.urlNotImage'));
  return;
}
```

- 影响：该流程通常受浏览器 CORS 限制，不是服务端 SSRF；但用户可触发浏览器下载超大文件、重定向到非预期资源、或使用 `image/svg+xml` 等后端不应接受的类型。当前先下载完整 blob 再校验大小，可能造成内存与带宽浪费。
- 整改建议：
  1. 用 `new URL(url)` 解析并显式检查 `protocol` 为 `http:` / `https:`，避免仅正则判断。
  2. 读取 `Content-Length` 并在下载前按运行时最大上传大小拒绝；下载后仍保留 blob size 校验。
  3. 使用现有 `validateFiles` 的 MIME 白名单，而不是仅 `mimeType.startsWith('image/')`。
  4. 明确拒绝 SVG：项目约定后端 raster-only allowlist 不应重新引入 `svg`。
  5. 考虑对重定向后的 `response.url` 再做协议校验。
- 缓解措施：保留后端 MIME、大小、解码校验作为最终安全边界。
- 误报说明：这是用户浏览器侧下载，不会直接让服务器访问内网；不能按 SSRF 定性。

## 低风险 / 防御纵深发现

### FE-SEC-006：Markdown 组件目前安全姿态较好，但 DOMPurify 使用位置可简化并增加回归保护

- Rule ID：JS-XSS-001
- Severity：Low
- 位置：[MarkdownContent.svelte:L1-L56](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/MarkdownContent.svelte#L1-L56)
- 证据：

```ts
import DOMPurify from 'dompurify';
import { marked, type Token } from 'marked';

const tokens = $derived(marked.lexer(DOMPurify.sanitize(content)));
```

组件随后用 Svelte 文本插值输出 token 文本，没有使用 `{@html}`：

```svelte
<p>{textFromTokens(token.tokens)}</p>
<pre><code>{token.text}</code></pre>
```

- 影响：当前未发现可执行 HTML 注入路径；但 `DOMPurify.sanitize(content)` 在进入 `marked.lexer` 前执行，主要降低输入中的 HTML 干扰，而真正安全边界来自“不把 HTML 输出为 HTML”。未来若改为 `{@html marked(...)}`，风险会显著上升。
- 整改建议：
  1. 保持“不使用 `{@html}` 渲染公告内容”的约束。
  2. 增加最小测试/审查样例：`<img src=x onerror=alert(1)>`、`[x](javascript:alert(1))`、内联 HTML、代码块等输入应只显示文本或被忽略。
  3. 若未来确需富文本 HTML，必须先用 DOMPurify allowlist 清洗最终 HTML，再考虑 Trusted Types。
- 缓解措施：在代码审查清单中加入“Markdown 渲染不得直接输出 raw HTML”。
- 误报说明：当前发现是防御纵深建议，不是已证实漏洞。

### FE-SEC-007：`target="_blank"` 链接使用 `rel="noreferrer"`，建议统一为 `noopener noreferrer`

- Rule ID：JS-SUPPLY-001 / Browser navigation hardening
- Severity：Low
- 位置：
  - [ImageDataTable.svelte:L50](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/ImageDataTable.svelte#L50)
  - [ImagePreviewDialog.svelte:L46-L47](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/ImagePreviewDialog.svelte#L46-L47)
  - [ImageDetailDrawer.svelte:L137-L138](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/ImageDetailDrawer.svelte#L137-L138)
  - [+page.svelte:L126](file:///d:/Works/MyProject/OmePic/frontend/src/routes/admin/dashboard/images/+page.svelte#L126)
- 证据：

```svelte
<a class="studio-button p-2" href={record.url} target="_blank" rel="noreferrer" ...>
```

- 影响：现代浏览器中 `noreferrer` 通常也隐含 `noopener`，但显式写出 `noopener noreferrer` 更清晰、更兼容，也便于静态审计。
- 整改建议：将所有新窗口链接统一改为 `rel="noopener noreferrer"`，并结合 FE-SEC-003 做 URL 校验。
- 缓解措施：保留当前 `noreferrer` 已能降低 referrer 泄漏并通常阻断 opener。
- 误报说明：这是加固建议，当前不是高危漏洞。

### FE-SEC-008：管理端存储密钥表单会在编辑时把密钥字段放入前端状态，需确认后端不会回传明文密钥

- Rule ID：Secrets handling / JS-STORAGE-001
- Severity：Low
- 位置：
  - [StorageInstanceManager.svelte:L22-L38](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/StorageInstanceManager.svelte#L22-L38)
  - [StorageInstanceManager.svelte:L60-L64](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/StorageInstanceManager.svelte#L60-L64)
  - [StorageInstanceManager.svelte:L76-L89](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/StorageInstanceManager.svelte#L76-L89)
  - [StorageInstanceManager.svelte:L216-L227](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/StorageInstanceManager.svelte#L216-L227)
- 证据：

```ts
function startEdit(instance: StorageInstance) {
  editingKey = instance.storage_key;
  form = { ...blank, ...instance };
  editorOpen = true;
}
```

```ts
base.s3_access_key = form.s3_access_key?.trim();
base.s3_secret_key = form.s3_secret_key?.trim();
base.webdav_pass = form.webdav_pass?.trim();
```

```svelte
<input class="studio-input" type="password" autocomplete="new-password" bind:value={form.s3_secret_key} />
<input class="studio-input" type="password" autocomplete="new-password" bind:value={form.webdav_pass} />
```

- 影响：如果后端 `adminGetConfig` 返回明文 `s3_secret_key` / `webdav_pass`，这些秘密会进入浏览器内存并可被 XSS 或浏览器扩展读取；若回传掩码值，前端保存时也可能把掩码当作真实密钥提交。
- 整改建议：
  1. 后端配置读取接口只返回 `configured` / `masked` 状态，不返回明文 secret。
  2. 前端编辑时 secret 字段默认空；空值表示“不修改原 secret”。
  3. 保存 payload 时仅在用户输入新 secret 时发送 secret 字段。
  4. 在类型上区分 `StorageInstanceRead` 与 `StorageInstanceWrite`，避免读模型和写模型混用。
- 缓解措施：当前 `autocomplete="new-password"` 是正确方向，但不能替代“不向前端回传明文密钥”。
- 误报说明：需结合后端接口确认是否已经做了掩码；仅从前端类型无法证明明文泄露。

## 正向安全实践

1. [MarkdownContent.svelte](file:///d:/Works/MyProject/OmePic/frontend/src/lib/components/studio/MarkdownContent.svelte) 未使用 `{@html}`，公告 Markdown 通过 token 转文本/结构化节点渲染，降低 XSS 风险。
2. 未发现 `eval`、`new Function`、`document.write`、`postMessage('*')` 等高危模式。
3. API helper 使用 `URLSearchParams` 构造查询参数，避免手写拼接查询字符串。
4. 新窗口链接已设置 `rel="noreferrer"`，具备基本 opener/referrer 风险缓解。
5. 管理端请求集中使用 `Authorization: Bearer`，客户端上传/删除使用 `X-Token`，两类凭证逻辑相对清晰。

## 优先级整改路线

### P0 / 近期必须完成

1. 为生产静态资源响应补齐安全响应头，优先在 Go 后端或部署层设置 CSP、`frame-ancestors`、`X-Content-Type-Options`、`Referrer-Policy`、`Permissions-Policy`。
2. 审核并整改管理员 JWT 存储方式，至少缩短有效期并从 `localStorage` 迁移到内存态；理想方案为后端 `HttpOnly` Cookie 会话。

### P1 / 下一轮迭代

1. 增加 `safeImageUrl` / `safeHref` 工具，所有 `record.url`、动态图片 URL、动态下载链接统一校验。
2. 远程 URL 上传改用 `new URL`、下载前 `Content-Length` 校验、下载后 `blob.size` 校验、MIME 白名单校验，并拒绝 SVG。
3. 删除旧图时优先使用历史记录中保存的 `record.client_token`，确保用户重置当前 Token 后仍可删除旧图。
4. 存储密钥编辑流程改成 read/write 类型分离，确保后端不回传明文 secret，前端只在用户输入新值时提交 secret。

### P2 / 防御纵深

1. Markdown 组件增加恶意输入回归测试或最小审查脚本。
2. 将所有 `target="_blank"` 链接统一为 `rel="noopener noreferrer"`。
3. 评估 Trusted Types 与更严格 CSP 的兼容性。
4. 清理 [preferences.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/preferences.ts) 与 [preferences.svelte.ts](file:///d:/Works/MyProject/OmePic/frontend/src/lib/stores/preferences.svelte.ts) 中重叠的 admin token 逻辑，减少安全修复遗漏。

## 建议验证命令

本次任务仅生成审查文档，未修改前端运行代码。后续实施整改后，建议在 `frontend/` 目录运行：

```bash
npm run lint
npm run typecheck
npm run build:backend
```

并在部署环境额外验证：

```bash
curl -I https://<your-domain>/
```

确认响应中存在预期的 CSP 与安全响应头。