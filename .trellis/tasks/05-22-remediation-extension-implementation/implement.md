# 整改与扩展任务拆分与实施计划

## 1. 执行顺序

1. 输出 `docs/tasks/remediation-extension/` 全量任务文档。
2. 补齐后端 P0 数据模型与迁移：运行时保护字段、存储健康表、核心索引。
3. 实现上传安全基线：资源保护、真实文件校验、URL 上传前端下载改造。
4. 实现运营治理：默认密码状态、存储健康检查。
5. 补齐前端跨层类型/API/最小 UI：URL 上传由前端下载后复用文件上传接口，运行时设置显示新增保护字段，默认密码告警。
6. 为每个 P0 子任务补充或更新测试。
7. 运行验证命令并将结果写入各 P0 子任务完成报告。
8. 输出 `docs/tasks/remediation-extension/README.md` 总完成报告。

## 2. 待改文件范围

- 后端：
  - `backend/internal/model/*`
  - `backend/internal/repository/*`
  - `backend/internal/service/*`
  - `backend/internal/http/handler/*`
  - `backend/internal/http/router/*`
  - `backend/internal/storage/*`
  - `backend/cmd/server/main.go`
- 前端：
  - `frontend/src/lib/api.ts`
  - `frontend/src/lib/types/index.ts`
  - `frontend/src/routes/+page.svelte`
  - `frontend/src/routes/admin/dashboard/settings/+page.svelte`
  - `frontend/src/lib/i18n.ts`
- 文档：
  - `docs/tasks/remediation-extension/**`
  - 必要时同步 `docs/api-reference.md` / `.trellis/spec/**`

## 3. 验证命令

```bash
cd backend && gofmt -w ./cmd ./internal
cd backend && go test ./...
cd frontend && npm run lint
cd frontend && npm run typecheck
cd frontend && npm run test
cd frontend && npm run build:backend
```

## 4. 风险与回滚点

- SQLite 迁移新增列/表需幂等，失败时回滚代码即可；新增列可留存。
- URL 上传改为前端下载后需确保公开文件上传接口仍兼容，并接受浏览器下载得到的 `File`。
- 回收站功能已按用户确认取消；删除保持直接删除数据库记录并修复缓存映射。
- 存储健康检查不得阻断主上传链路；后台定时任务仅记录日志。

## 5. 复核清单

- [ ] P0/P1/P2 全量文档存在且可映射源设计文档。
- [ ] P0 每个子任务完成报告包含实现、测试、复核结论。
- [ ] 后端测试通过。
- [ ] 前端 lint/typecheck/test/build 通过。
- [ ] 未混入与本任务无关的文件。
