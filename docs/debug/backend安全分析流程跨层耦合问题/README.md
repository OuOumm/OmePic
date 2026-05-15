# backend 安全分析流程跨层耦合问题

## Files

- `backend/internal/service/admin_service.go`
- `backend/internal/service/abuse.go`
- `backend/internal/repository/ip_ban_repository.go`
- `backend/internal/model/abuse.go`
- `backend/internal/model/ip_ban.go`
- `backend/internal/service/ip_utils.go`
- `backend/internal/repository/helpers.go`
- `.trellis/spec/backend/security.md`

## Problem

IP-ban 与 abuse 分析流程目前跨 AdminService、Repository、Model、IP utility 分散实现，Seam 有泄漏：

1. `AdminService.CreateIPBan` 同时处理 UID 查 IP、默认 reason、expires_at、重复 active ban、summary。
2. `Repository.TopAbuseIPs` 查询上传聚合，同时调用 `ActiveIPBansByHash` 并组装 masked IP、ban 状态。
3. `Repository.IPDetail` 又组合 `ImageSummaryByIP`、mask、active ban。
4. IP hash/mask 在 service 与 repository 各有 helper 包装，维护者必须知道哪些层可计算 hash/mask，哪些层只做 SQL。
5. `normalizeAbuseRange` 是独立小函数，有测试价值，但它只覆盖时间范围；真正容易错的 abuse 聚合与 ban 标注仍在 Repository Implementation 内。

结果是“安全分析”这个领域 Module 没有集中 Interface。Repository 不只是 SQLite Adapter，还知道 abuse view 的展示字段、masked IP、active ban 标注等领域组合。

删除测试：删除 `normalizeAbuseRange` 只会把几行时间逻辑放回 `AdminService.AbuseOverview`，复杂度变化不大；删除 `Repository.TopAbuseIPs` 则会迫使 SQL 聚合与 active ban 标注在上层重写，说明这里混入了有价值但位置不清的领域 Implementation。

## Solution

建议把“安全分析/封禁工作流”作为候选 Deep Module：

- 将 IP-ban 创建、active ban 判断、abuse overview、IP detail 看作同一领域流程，而不是由 AdminService 与 Repository 各自拼装。
- Repository 作为 SQLite Adapter 尽量返回持久化事实或聚合事实；masked view、active ban 标注、默认 reason 等由安全分析 Module 的 Implementation 统一处理。
- IP hash/mask 的 Interface 应在一个地方表达，避免 service/repository helper 同时持有同一知识。
- `normalizeAbuseRange` 可并入安全分析 Module 的内部 Implementation，让测试穿过更高 Leverage 的 Seam。

## Benefits

- **Locality**：active ban 语义、hash/mask、默认 reason、时间范围、abuse 标注集中，安全规则变化不会跨 SQL Adapter 与 AdminService 同步修改。
- **Leverage**：后台 IP-ban 列表、创建封禁、abuse overview、IP detail 都可复用同一 Module 行为。
- **测试改善**：测试可以直接验证安全分析 Module 的 Interface：同一 IP 被 active ban 标注、过期 ban 不标注、UID 创建 ban 的 summary 正确、时间窗口非法返回 `invalid_input`。

## 删除测试

- 删除 `normalizeAbuseRange`：复杂度只回到调用点，说明该函数偏 Shallow。
- 删除 `TopAbuseIPs` 中 active ban 标注：上层必须知道 hash 与 active ban 规则，说明当前 Repository Implementation 包含领域深度但放错了 Seam。
- 删除 service/repository 中任一套 hash/mask helper：另一层仍有相似知识，说明知识重复。

## 证据引用

- `backend/internal/service/admin_service.go:287` — `CreateIPBan` 同时负责 UID 查找、IP 选择、reason、expires、重复 active ban 与 summary。
- `backend/internal/service/admin_service.go:410` — `DeleteImagesByIPBan` 从 ban 查图片并逐个调用 Image 删除流程。
- `backend/internal/service/abuse.go:5` — `normalizeAbuseRange` 只封装时间范围。
- `backend/internal/repository/ip_ban_repository.go:70` — `FindActiveIPBanByIP` 在 Repository 内计算 IP hash。
- `backend/internal/repository/ip_ban_repository.go:106` — `TopAbuseIPs` 聚合图片并标注 active ban。
- `backend/internal/repository/ip_ban_repository.go:188` — `IPDetail` 组合 summary、mask、active ban。
- `backend/internal/repository/helpers.go:63` — `ipHashValue` 在 Repository helper 中包装 hash。
- `backend/internal/repository/helpers.go:67` — `maskIPValue` 在 Repository helper 中包装 mask。

## 整改步骤

1. 先列出安全分析用例：create ban by UID/IP、list active bans、overview、IP detail、delete ban images。
2. 标注哪些输出是持久化事实，哪些是展示/安全分析派生事实。
3. 调整测试，把 active ban 标注、mask、默认 reason 从 Repository 测试提升到安全分析 Module 测试。
4. Repository 保留 SQL 查询与基础聚合，减少对 masked view 和 active ban 标注的了解。
5. 统一 IP hash/mask 的使用位置，避免跨层重复 helper。

## 验证建议

- `cd backend && go test ./internal/service -run 'Test.*IPBan|Test.*Abuse'`
- `cd backend && go test ./internal/repository -run 'Test.*IPBan|Test.*Abuse'`
- 增加组合测试：过期 ban 不标注、重复 active ban 返回原 ban 和 summary、delete images by ban 保持 Image 删除语义。
