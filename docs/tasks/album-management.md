# 任务：相册/集合管理

## 关联的整改/扩展条目

- 源设计文档 `docs/debug/remediation-extension-design/rectification-and-extension-design.md` **§2 核心功能扩展提案 F6** — "相册/集合管理"
- **§4.4 P2：相册/集合** — "相册 CRUD、公开分享页、批量加入、权限控制"

## 预估人日

7 人日

## 开发

### 需要修改/新增的模块或文件

- `backend/internal/model/album.go`（新增）— 相册模型
- `backend/internal/model/album_image.go`（新增）— 相册-图片关联
- `backend/internal/repository/album_repo.go`（新增）— 相册数据访问
- `backend/internal/service/album_service.go`（新增）— 相册业务逻辑
- `backend/internal/http/handler/album.go`（新增）— 相册 API（公开+管理）
- `backend/internal/http/router/router.go` — 注册相册路由
- `backend/internal/database/migration.go` — 新增表迁移

### 关键实现点

1. 相册模型：`id`、`title`、`description`、`visibility`（public/private）、`cover_uid`、`created_at`。
2. 相册-图片关联表：`album_id`、`image_uid`、`sort_order`。
3. 公开分享页：`/a/:album_id` 展示相册图片。
4. 管理 API：创建、编辑、删除相册；批量添加/移除图片；排序。
5. 权限控制：私有相册仅管理员可见。
6. 相册封面自动选择第一张图片或手动设置。

### 完成标准

- [ ] 代码编译通过：`cd backend && go build ./...`
- [ ] 单元测试通过
- [ ] 公开相册页面可访问

## 测试

### 测试场景

| 类型 | 场景 | 预期结果 |
|------|------|----------|
| 正向 | 创建公开相册并添加图片 | 相册页面展示图片 |
| 正向 | 创建私有相册 | 非管理员不可见 |
| 正向 | 批量添加图片到相册 | 图片正确关联 |
| 异常 | 访问不存在的相册 | 返回 404 |
| 边界 | 相册为空 | 展示空状态 |
| 边界 | 删除相册 | 关联关系清除，图片不受影响 |

### 必须覆盖的测试类型

- 单元测试：相册 CRUD、权限判断
- 集成测试：创建→添加图片→访问分享页

### 测试通过标准

- `cd backend && go test ./...` 全部通过

## 复查

### 代码审查关注点

- 删除相册时是否正确清理关联关系
- 公开分享页性能（图片列表分页）

### 安全检查

- 私有相册访问控制
- 相册管理 API 权限

### 文档更新要求

- API 文档说明相册接口

### 复查通过标准

- [ ] 至少一位 reviewer 批准
- [ ] CI 全部绿色
- [ ] 完成报告已填写
