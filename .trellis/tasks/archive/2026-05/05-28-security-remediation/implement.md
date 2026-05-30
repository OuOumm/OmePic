# Implementation Plan

## Execution Order

### Phase 0: 基础设施
1. **Q-02**: SQLite `PRAGMA user_version` schema 版本号
   - 修改 `backend/internal/repository/migration.go`
   - 添加 version 管理逻辑
   - 添加测试校验

### Phase 1: 入口安全与抗 DoS
2. **H-01**: 强制 .env 配置
   - 修改 `backend/internal/config/config.go`
   - 修改 `backend/cmd/server/main.go` 启动检查
   - 删除 `DefaultAdminPassword` 常量

3. **H-02 + Q-03**: http.Server + 优雅关闭
   - 修改 `backend/cmd/server/main.go`
   - 添加 signal 监听和 graceful shutdown

4. **H-03**: MaxBytesReader body limit
   - 新增 `backend/internal/http/middleware/body_limit.go`
   - 注册到上传路由

5. **M-07**: 限流 Redis fail-closed
   - 修改 `backend/internal/http/middleware/rate_limit_middleware.go`

6. **H-05**: X-Token 移除 Math.random
   - 修改 `frontend/src/lib/client-token.ts`

7. **M-02**: 后台真实 IP 来源配置
   - 修改 `backend/internal/service/runtime_settings_fields.go`
   - 修改 `backend/internal/http/clientip/resolver.go`
   - 修改前端设置页

8. **M-03**: 强制 public_base_url
   - 修改启动配置校验逻辑

## Validation

- 每个完成后运行相关测试
- 最终运行完整测试套件:
  - `cd backend && go test ./... && go vet ./...`
  - `cd frontend && npm run lint && npm run typecheck && npm run test && npm run build:backend`

## Dependencies

- Phase 1 items 互相独立，可并行
- M-03 依赖 H-01 的配置校验机制
