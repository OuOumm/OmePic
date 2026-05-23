# 任务：存储回退策略

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§1.2 性能瓶颈与整改建议** — "WebDAV/S3 网络写入在上传主链路中失败即失败"
- **§3.2.2 存储回退策略**
- **§4.3 P1：存储回退策略** — "fallback group、健康状态联动、上传切换、读取副本策略设计"

## 预估人日

6 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/storage.go` — 存储实例增加回退配置字段
- `backend/internal/service/storage_router.go`（新增）— 存储路由策略服务
- `backend/internal/service/image_service.go` — 上传使用路由策略选择存储
- `backend/internal/http/handler/image.go` — 读取支持副本回退
- `backend/internal/http/handler/admin_storage.go` — 存储迁移工具 API
- `backend/internal/database/migration.go` — 迁移脚本

### 关键实现点

1. 存储实例配置增加：`priority`、`read_only`、`fallback_group`、`health_status`。
2. 上传选择策略：
   - 默认存储健康：写默认。
   - 默认存储降级：写同组备用实例。
   - 用户指定存储不可用：返回明确错误或询问是否切换。
3. 读取策略：
   - 单副本：按记录的 storage_key 读取。
   - 主失败后读副本，并异步修复主副本。
4. 后台提供迁移工具：按存储实例、日期、Token 分批迁移。
5. 可重试错误引入指数退避；区分幂等重试与不可重试错误。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过：`cd backend && go test ./...`
- [ ] gofmt 格式正确

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 默认存储健康时上传 | 写入默认存储 |
| 正向 | 默认存储降级时上传 | 写入备用存储 |
| 正向 | 读取主副本失败 | 自动读副本 |
| 异常 | 所有存储不可用 | 返回明确错误 |
| 异常 | 迁移任务目标存储不可用 | 暂停迁移，记录错误 |
| 边界 | 存储从降级恢复为健康 | 上传自动切回主存储 |

### 必须覆盖的测试类型

- 单元测试：存储路由策略、回退逻辑
- 集成测试：模拟存储故障→回退→恢复
- 迁移测试：跨存储迁移正确性

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 存储选择策略的线程安全
- 回退读取时的延迟影响
- 迁移任务的断点续传能力

### 安全检查

- 存储迁移 API 需要管理员权限
- 迁移过程不暴露存储凭据

### 文档更新要求

- 运维文档说明存储回退配置
- API 文档说明迁移工具接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
