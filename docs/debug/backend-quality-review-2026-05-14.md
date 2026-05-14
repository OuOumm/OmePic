# Backend quality review — 2026-05-14

## 审查范围

本次围绕任务 `05-14-admin-avif-conversion-settings` 对 `backend/` 相关实现与未提交改动进行审查，重点覆盖：

- `backend/internal/service/image_service.go`
- `backend/internal/service/image_transform.go`
- `backend/internal/service/runtime_settings.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/http/handler/image_handler.go`
- `backend/internal/http/handler/admin_handler.go`
- `backend/internal/storage/storage.go`
- 相关测试：
  - `backend/internal/service/image_service_test.go`
  - `backend/internal/service/runtime_settings_test.go`
  - `backend/internal/service/admin_service_test.go`
  - `backend/internal/http/handler/image_handler_test.go`
  - `backend/internal/http/handler/admin_handler_test.go`

同时对 Trellis 任务文档与 backend spec 做了对照审查：

- `.trellis/tasks/05-14-admin-avif-conversion-settings/prd.md`
- `.trellis/tasks/05-14-admin-avif-conversion-settings/design.md`
- `.trellis/tasks/05-14-admin-avif-conversion-settings/implement.md`
- `.trellis/spec/backend/index.md`
- `.trellis/spec/backend/runtime-settings.md`
- `.trellis/spec/backend/database-guidelines.md`
- `.trellis/spec/backend/error-handling.md`
- `.trellis/spec/backend/quality-guidelines.md`

## 已执行检查

1. 审阅当前未提交 diff（backend + 相关文档）
2. 阅读任务 PRD / Design / Implement
3. 阅读适用 backend spec
4. 代码检索：AVIF 设置、上传链路、runtime settings、system-settings 更新路径
5. 运行后端测试：

```bash
cd backend && go test ./...
```

结果：**通过（114 passed in 16 packages）**

## 总结结论

- **本次未发现 AVIF 质量/速度配置需求本身的明显实现缺口**：runtime settings、范围校验、上传时使用当前配置、重复上传不重转，这些主路径从代码与测试上看基本成立。
- **但发现 1 个阻塞性质量问题**：AVIF 流式编码 + 存储写入的错误处理存在潜在死锁，属于上传链路稳定性问题，建议在继续合并/发布前修复。
- 另有若干 **中低优先级问题/覆盖缺口**，不会立即阻塞功能，但建议在本任务收尾前补强。

---

## 发现的问题

### P0 / 阻塞：`saveConvertedAVIF` 在存储端提前失败时可能死锁

**结论**

`ImageService.saveConvertedAVIF()` 当前用 `io.Pipe()` 把 AVIF 编码协程与 `provider.SaveStream()` 串起来；如果存储层在尚未消费 `pipeReader` 前就提前返回错误，编码协程可能阻塞在向 `pipeWriter` 写数据，主协程又会阻塞等待 `encodeErrCh`，导致整个上传请求卡死。

**证据**

- `backend/internal/service/image_service.go:613`
  - `saveConvertedAVIF()` 内创建 `pipeReader, pipeWriter := io.Pipe()`
- `backend/internal/service/image_service.go:627`
  - 同步调用：`provider.SaveStream(ctx, objectKey, pipeReader, -1, publicImageMIMEType)`
- `backend/internal/service/image_service.go:628`
  - 随后无条件等待：`encodeErr := <-encodeErrCh`
- `backend/internal/storage/storage.go:168`
  - `localProvider.SaveStream()` 在 `os.MkdirAll` / `os.Create` 失败时可在真正读取 `reader` 前直接返回
- `backend/internal/storage/storage.go:223`
  - `s3Provider.SaveStream()` 也可能在上传初始化失败时直接返回
- `backend/internal/storage/storage.go:264`
  - `webdavProvider.SaveStream()` 在 `MkdirAll` / `WriteStream*` 初期失败时也可能提前返回

**为什么这是阻塞问题**

典型死锁路径：

1. 编码 goroutine 开始向 `pipeWriter` 写入 AVIF 数据；
2. `provider.SaveStream()` 因目录创建/网络/鉴权等问题提前返回；
3. `pipeReader` 无人继续消费；
4. 编码 goroutine 卡在写 pipe；
5. 主 goroutine 等待 `<-encodeErrCh`，请求挂死。

这类问题在本地 happy-path 测试里不容易出现，所以 `go test ./...` 通过不能证明这里安全。

**整改建议**

- 在 `provider.SaveStream()` 返回后，若 `saveErr != nil`，应主动中断 pipe，使编码协程可退出，例如：
  - `pipeReader.Close()` / `pipeWriter.CloseWithError(...)` 配合 context/cancel；或
  - 将编码与存储都置于可取消上下文中，任何一端失败都广播取消；或
  - 改成显式 errgroup + 双向关闭协议。
- 增加回归测试：构造一个 `SaveStream` 在未读流前立即报错的 fake provider，断言 `Upload()` 会快速返回而不是阻塞。

**建议验证**

- 新增单测：`SaveStream` 立即返回 error
- 新增单测：`SaveStream` 读一小段后返回 error
- 必要时加超时保护，避免测试自身卡死

---

### P1 / 重要：缺少对上述“存储失败路径”的单元测试覆盖

**结论**

当前测试覆盖了 AVIF 转换主链路、去重、配置传递、URL 规则，但没有覆盖最危险的“流式转换 + 存储失败”路径。

**证据**

- `backend/internal/service/image_service_test.go`
  - 已覆盖：
    - 新上传转 AVIF
    - 重复上传跳过重转
    - 读取源数据到临时文件
    - 配置化 `AvifQuality` / `AvifSpeed`
  - **未见**：
    - fake provider 提前失败
    - `saveConvertedAVIF()` 错误传播与 goroutine 退出
    - 上传接口在存储失败场景下不会悬挂

**整改建议**

补 2 类测试：

1. `provider.SaveStream()` 立即失败；
2. `provider.SaveStream()` 读到部分字节后失败。

同时建议把 `saveConvertedAVIF()` 抽成更易注入 fake provider / fake encoder 的可测单元。

---

### P1 / 重要：HTTP 层对 runtime settings AVIF 参数的契约测试仍偏弱

**结论**

服务层已经有 AVIF 范围校验与“不部分保存”的测试，但 HTTP handler 层没有看到对 `/admin/system-settings` 的 JSON 契约进行覆盖，尤其是无效 `avif_quality` / `avif_speed` 时返回 `400 invalid_input` 的断言。

**证据**

- `backend/internal/service/admin_service_test.go:380`
  - 已有 `TestUpdateSystemSettingsRejectsInvalidAVIFSettingsWithoutPartialSave`
- `backend/internal/http/handler/admin_handler.go:260`
  - `UpdateSystemSettings()` 负责 HTTP 入口绑定与错误映射
- `backend/internal/http/handler/admin_handler_test.go`
  - 当前文件主要覆盖改密，未见 system-settings 的 AVIF 参数错误映射测试

**风险**

如果后续 handler 绑定、错误映射、响应 message/code 发生回归，服务层测试无法及时发现。

**整改建议**

增加 handler 级测试，至少覆盖：

- `PUT /admin/system-settings` 提交 `avif_quality: 101` -> `400 invalid_input`
- `avif_speed: -1` -> `400 invalid_input`
- 成功更新后响应体包含当前 runtime settings 的 `avif_quality` / `avif_speed`

---

### P2 / 中：Reader 上传路径始终落盘，性能上是“稳妥优先”而非“轻量优先”

**结论**

当前 `UploadInput.Source` 路径无论文件大小都会先完整写入临时文件，再二次打开进行解码和转换。这符合当前“先算原始 MD5、后做去重、再转换”的契约，也便于控制内存，但会带来额外磁盘 IO。

**证据**

- `backend/internal/service/image_service.go:571`
  - `prepareUploadSource()` 对 `input.Source` 使用 `os.CreateTemp()`
- 同函数内部
  - `io.Copy(writer, io.LimitReader(input.Source, readLimit))`
- 后续 `prepared.Open()` 再次读取该临时文件参与 AVIF 转换

**影响**

- 对小文件：多一次磁盘写入与读回
- 对高并发上传：临时文件 IO 与临时目录压力会上升
- 但优点是：不会把原图完整驻留内存，且满足“原始字节 MD5 去重先于转换”

**整改建议**

这个问题**不建议在本任务内强改**，但可以记录为后续性能优化项：

- 小文件走内存阈值，大文件落盘；
- 或引入“单次流读取 + tee 到 hash + 可重读缓存”的分层策略；
- 若未来接入高并发对象存储上传，需复查临时目录和磁盘瓶颈。

---

### P2 / 中：上传 MIME 校验仍主要依赖 multipart header / 文件名回退，兼容性风险存在

**结论**

当前 handler 将 `fileHeader.Header.Get("Content-Type")` 作为 MIME 主来源，仅在为空时用文件扩展名回退。服务层会先按 runtime MIME 白名单拦截，再进入真正的图片解码。这意味着：某些客户端若上传真实图片但带了通用或错误的 Content-Type（如 `application/octet-stream`），可能在解码前就被拒绝。

**证据**

- `backend/internal/http/handler/image_handler.go:41-45`
  - 先取 multipart header 的 `Content-Type`
- `backend/internal/http/handler/image_handler.go:46-55`
  - 仅为空时回退 `detectContentType(fileHeader.Filename)`
- `backend/internal/service/image_service.go`
  - `runtimeSettingsAllowsMIME(runtimeSettings, input.MIMEType)` 在真正解码前执行

**风险说明**

这不是本任务新引入的问题，但和上传链路契约直接相关。若未来接第三方客户端、移动端或脚本上传，可能出现“文件本身可解码，但因为 header 不规范被拒绝”的兼容性投诉。

**整改建议**

后续可考虑：

- 对 `application/octet-stream` 等泛型值做内容 sniff；
- 或在服务层允许“header 不可靠时，以实际解码结果兜底”；
- 但要确保仍符合 spec 中“按 runtime MIME 白名单校验”的约束，避免引入新的扩展名/内容漂移问题。

---

## 正向结论 / 已确认较好的点

以下方面本次未发现明显问题：

1. **AVIF 配置已进入 runtime settings 主数据源**
   - `backend/internal/service/runtime_settings.go` 已包含：
     - `AvifQuality`
     - `AvifSpeed`
   - 默认值：`60 / 8`
   - `Load()` 通过 `InsertMissingConfigValues()` 持久化缺失默认值，符合 spec

2. **范围校验位置正确**
   - `ValidateRuntimeSettingsInput()` 对：
     - `avif_quality 0..100`
     - `avif_speed 0..10`
     做了统一校验
   - `AdminService.UpdateSystemSettings()` 先校验，再 `UpsertConfigValues()`，能避免部分保存

3. **上传链路已按当前 runtime settings 传递 AVIF 参数**
   - `image_transform.go` 中 `avifConversionSettingsFromRuntime()`
   - `image_service.go` 上传新物理文件时调用：
     - `saveConvertedAVIF(..., avifConversionSettingsFromRuntime(runtimeSettings))`

4. **重复上传不会因设置变化触发重转**
   - 去重在转换前完成；命中后直接复用已有 `FilePath` / `MD5Hash`
   - 相关测试已覆盖重复上传跳过转换

5. **后端测试总体健康**
   - `cd backend && go test ./...` 全绿
   - 已有测试对当前任务核心需求提供了较好基础覆盖

## 可选的后续验证步骤

建议在修复 P0 后补跑：

```bash
cd backend && go test ./...
```

并新增/补跑以下聚焦用例：

1. `ImageService.Upload` + fake provider 立即失败
2. `ImageService.Upload` + fake provider 部分读取后失败
3. `PUT /admin/system-settings` handler 级无效 AVIF 参数测试
4. 如后续改 MIME 判定策略，再补上传 Content-Type 兼容性测试

## 最终判断

- **是否发现阻塞性问题：是**
- **阻塞项**：`saveConvertedAVIF()` 在存储端提前失败时存在潜在死锁风险
- **是否建议继续发布当前后端实现：不建议在未修复该问题前直接发布上传链路相关改动**

在修复上述阻塞项前，其余 AVIF runtime settings 功能实现整体方向是对的，测试基础也不错。

---

## 整改结果 / 复验结果（2026-05-15）

### 已修复项

1. **`saveConvertedAVIF()` 提前失败挂起问题已修复**
   - 文件：`backend/internal/service/image_service.go`
   - 调整点：
     - 将 `provider.SaveStream(...)` 放入独立 goroutine；
     - 主流程先等待保存结果；
     - 一旦保存失败，主动执行 `pipeReader.Close()` 与 `pipeWriter.CloseWithError(saveErr)`，确保编码协程从 pipe 写入阻塞中退出；
     - 随后再等待编码协程结束并统一返回错误。
   - 结果：Upload 在存储层即时失败、部分读取后失败两种场景下都能快速返回，不再卡死。

2. **补充了上传失败路径回归测试**
   - 文件：`backend/internal/service/image_service_test.go`
   - 新增覆盖：
     - fake provider **立即失败** 时，`Upload()` 不挂起并返回错误；
     - fake provider **读取部分流后失败** 时，`Upload()` 不挂起并返回错误；
     - 两种失败场景都断言不会写入图片记录。

3. **补充了 `/admin/system-settings` handler 契约测试**
   - 文件：`backend/internal/http/handler/admin_handler_test.go`
   - 新增覆盖：
     - `avif_quality=101` 返回 `400 invalid_input`；
     - `avif_speed=11` 返回 `400 invalid_input`；
     - 成功更新时响应体包含 `runtime.avif_quality` / `runtime.avif_speed`。

### 复验结果

已执行：

```bash
cd backend && go test ./...
```

结果：**通过（all packages passed）**

### 最新结论

- 本文前述 **P0 阻塞问题已关闭**。
- 与本任务相关的 backend 风险点已得到最小范围修复，并补齐关键回归测试。
- 当前关于“管理员配置 AVIF 质量/速度 + 上传链路使用当前设置 + 非法值返回 invalid_input”的后端实现，结合本次补测后，**可以继续进入后续校验/收尾流程**。
