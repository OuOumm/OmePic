# Backend Upload Pipeline Remediation Summary

日期：2026-05-14  
范围：上传链路、AVIF 转码输出、默认安全告警、后台安全提示、相关 spec / debug 文档

---

## 目标

本轮整改围绕后端系统性质量审查报告，重点处理以下问题：

1. 上传链路中的额外内存拷贝与重复扫描
2. AVIF 编码结果先落整块内存再写存储
3. 默认密钥 / 默认管理员密码缺少可见告警
4. CORS 在配置 `public_base_url` 后仍然过宽
5. 后台设置页缺少安全状态可视提示
6. 相关 spec / debug 文档与实现脱节

---

## 已完成整改

### 1. 上传链路改为 service 统一准备上传源

当前生产路径不再以整块 `[]byte` 作为首选输入，而是：

- handler 仅把 multipart 文件句柄以 `io.Reader` 形式传入 service
- service 使用 `prepareUploadSource()`：
  - 对 reader-backed 输入按需落临时文件
  - 同步计算原始字节 MD5
  - 记录实际大小
  - 在请求结束后统一清理临时文件

效果：

- 原图不再要求完整常驻内存
- 临时文件生命周期不再散落在 handler
- 生产契约明确收敛到 `Source + DeclaredSize`

### 2. `Bytes` 输入降级为兼容 / 测试入口

`service.UploadInput` 仍保留 `Bytes []byte`，但已明确标注：

- 仅用于兼容 / 测试路径
- 生产路径优先使用 `Source + DeclaredSize`

同时新增：

- `NewUploadInputFromBytes(...)`

便于未来继续收敛测试写法。

### 3. 原始字节 MD5 只计算一次

现在 reader-backed 生产路径在 service 准备上传源时就完成：

- 原始字节 MD5 计算
- 后续 dedup 直接使用该 MD5

避免 service 再次对整块原图做重复扫描。

### 4. AVIF 编码结果改为流式写入 storage

storage provider 新增：

- `SaveStream(ctx, objectKey, reader, size, contentType)`

并为三种 provider 实现：

- local
- s3
- webdav

AVIF 编码新增：

- `encodeAVIFToWriter(source, target, settings)`

ImageService 通过：

- `io.Pipe()`
- `countingWriter`

把 AVIF 编码结果直接流向 storage provider，避免：

- 先编码到完整 `[]byte`
- 再整块调用 `Save([]byte)`

### 5. 默认安全配置启动告警

`backend/cmd/server/main.go` 现在会在启动时对以下状态输出 `WARN`：

- `JWT_SECRET` 仍是默认值
- `UID_ENCRYPTION_KEY` 仍是默认值
- 管理员密码仍可通过首次启动默认密码引导登录

### 6. CORS 收敛

当前逻辑：

- 未配置 runtime `public_base_url`：维持 `AllowAllOrigins=true`
- 已配置 runtime `public_base_url`：自动收紧到该精确 Origin

这样既兼容默认部署，又避免在明确配置公开访问域名后依然完全开放跨域。

### 7. 后台设置页新增安全警告面板

前端设置页现在会基于 `AdminSystemSettings.readonly.security` 展示：

- JWT 默认值警告
- UID 加密密钥默认值警告
- 管理员密码哈希尚未初始化警告

---

## 影响的主要文件

### Backend

- `backend/cmd/server/main.go`
- `backend/internal/http/handler/image_handler.go`
- `backend/internal/http/router/router.go`
- `backend/internal/http/router/frontend_test.go`
- `backend/internal/repository/image_repository.go`
- `backend/internal/service/image_service.go`
- `backend/internal/service/image_service_test.go`
- `backend/internal/service/image_transform.go`
- `backend/internal/service/admin_service_test.go`
- `backend/internal/storage/storage.go`

### Frontend

- `frontend/src/routes/admin/dashboard/settings/+page.svelte`
- `frontend/src/lib/i18n.ts`

### Docs / Spec

- `.trellis/spec/backend/directory-structure.md`
- `.trellis/spec/backend/database-guidelines.md`
- `.trellis/spec/backend/security.md`
- `.trellis/spec/backend/runtime-settings.md`
- `docs/debug/backend-systematic-quality-check-2026-05-14.md`

---

## 当前上传链路（整改后）

### 原图输入

`multipart file -> service.Upload(Source, DeclaredSize, ...)`

### 上传源准备

`Source -> temp file + original-byte md5 + size`

### 去重

`original md5 -> dedup before AVIF conversion`

### 转码

`temp file reader -> image.Decode -> avif.Encode(writer)`

### 持久化

`writer -> io.Pipe -> storage.SaveStream`

---

## 仍未处理但已明确保留的事项

1. `AdminService` 职责拆分
2. 限流 `fail-open / fail-close` 策略化
3. Snowflake 时钟回退策略优化
4. 完全移除 `Bytes` 测试兼容入口（当前保留是为降低回归风险）

---

## 验证结果

- `cd backend && go test ./...` ✅ `114 passed in 16 packages`
- `cd frontend && npm run typecheck` ✅
- `cd frontend && npm test -- --run src/lib/api.test.ts src/lib/ui-errors.test.ts` ✅
- `cd frontend && npm run build:backend` ✅
- `git diff --check` ✅

---

## 结论

本轮整改已将上传链路从“整块原图驻内存 + 编码结果整块驻内存”推进到：

- 原图 reader 优先
- service 统一临时文件管理
- 原始字节 MD5 单次计算
- AVIF 编码结果流式写 storage

这已经是当前架构下兼顾稳定性、性能、可维护性的一种高质量实现。