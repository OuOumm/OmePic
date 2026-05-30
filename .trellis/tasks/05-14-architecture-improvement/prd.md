# Architecture Improvement: RuntimeSettings Table-Driven Design

## Goal

重构 RuntimeSettings 模块，使用描述表驱动方式减少重复代码，提高可维护性和扩展性。

## Confirmed Facts

- 当前添加一个新配置项需要修改 7+ 个位置
- `runtimeSettingsFromValues` 包含 14 个重复的 if 块
- 使用切片存储配置字段定义，查找效率为 O(n)
- `setFieldValue` 不返回错误，可能导致静默失败
- 切片序列化使用逗号分隔，可能与配置值中的逗号冲突

## Requirements

- 使用 `map[string]ConfigField` 而非切片，提高查找效率到 O(1)
- `setFieldValue` 必须返回 `error`，避免静默失败
- 切片序列化使用 JSON 编码，更可靠地处理特殊字符
- 添加单元测试验证所有字段都能通过 get/set 往返
- 保持向后兼容，不破坏现有 API 和数据库

## Acceptance Criteria

- [ ] `configFields` 改为 `map[string]ConfigField` 类型
- [ ] `setFieldValue` 返回 `error` 类型
- [ ] 切片序列化使用 JSON 编码
- [ ] 所有字段都有 get/set 往返测试
- [ ] `cd backend && go test ./...` 通过
- [ ] 现有功能行为不变

## Out of Scope

- 不使用反射
- 不修改前端代码（后续任务）
- 不引入代码生成（后续任务）
- 不拆分 AdminService（后续任务）
