# Implementation Plan: 后台 IP 取消脱敏展示

## Ordered checklist

1. 复读本任务 `prd.md`、`design.md` 与相关 backend/frontend/spec 文档。
2. 修改 backend model/service：删除 `ip_address_masked` 字段与后台脱敏生成逻辑。
3. 修改 backend repository/migration：删除 `ip_bans` 表的 `ip_address_masked` 读写与 schema 列。
4. 修改 frontend types 与后台管理页面/组件：统一展示 `ip_address`。
5. 更新相关测试：移除对 `ip_address_masked` 的断言，补充新 schema / 新响应契约断言。
6. 更新 `.trellis/spec` 文档，删除后台仍依赖 masked IP 的描述。
7. 运行 backend / frontend 验证命令并修正问题。

## Validation commands

后端：

```bash
cmd.exe //c "cd /d D:\Works\MyProject\OmePic\backend && go test ./..."
```

前端：

```bash
cmd.exe //c "cd /d D:\Works\MyProject\OmePic\frontend && npm run lint"
cmd.exe //c "cd /d D:\Works\MyProject\OmePic\frontend && npm run typecheck"
cmd.exe //c "cd /d D:\Works\MyProject\OmePic\frontend && npm run build:backend"
```

## Risky files / rollback points

- `backend/internal/repository/migration.go`
  - 风险：本地旧库 schema 不兼容；本任务不做兼容迁移。
- `backend/internal/repository/ip_ban_repository.go`
  - 风险：SQL 字段数与 scan 目标不一致会直接导致接口失败。
- `backend/internal/service/security_analysis.go`
  - 风险：IP 封禁、滥用排行、IP 详情都会受影响。
- `frontend/src/lib/types/index.ts`
  - 风险：会联动多个后台页面与组件的类型检查。
- `frontend/src/routes/admin/dashboard/security/+page.svelte`
  - 风险：封禁列表、排行列表、确认弹窗会同时受影响。

## Review gates before start

- PRD 已明确为“彻底删除，包含后端字段与数据库列”。
- 用户已确认 **不兼容旧库，允许重建表/删库**。
- 该任务同时跨 backend/frontend/spec，实施前后都需要做完整验证。
