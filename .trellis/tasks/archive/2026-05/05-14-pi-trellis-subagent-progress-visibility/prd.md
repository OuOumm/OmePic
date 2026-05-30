# Pi Trellis subagent progress visibility

## Goal

修复 Pi 平台 Trellis 自定义 `subagent` 工具在执行 `trellis-implement` / `trellis-check` 等子代理时只显示 `working...` 的问题，让主会话能实时看到子代理正在读文件、执行命令、编辑文件以及阶段进度。

## Confirmed Facts

- 当前项目通过 `.pi/extensions/trellis/index.ts` 注册了 Trellis 专用 `subagent` 工具。
- 当前实现使用 `pi --mode text -p --no-session` 启动子 Pi 进程，并把 stdout/stderr 缓存在内存中，进程结束后才返回最终文本。
- 因为没有调用工具执行器提供的 `onUpdate`，Pi TUI 只能显示默认 `working...`。
- Pi 支持 `--mode json` 事件流；工具 `execute` 可通过 `onUpdate(partialResult)` 推送部分结果，`renderResult` 可自定义展示。
- 项目已安装官方 `pi-subagents` 包，但 Trellis 使用的是项目本地 `.pi/extensions/trellis/index.ts` 中的自定义工具。

## Requirements

- 子代理运行期间必须持续推送可见进度，而不是只显示默认 `working...`。
- 进度至少包含：代理名、运行状态、最近工具调用、最近助手文本片段、运行时长、turn 数和基本 token/费用统计（有则显示）。
- 支持 single / parallel / chain 三种已有模式。
- 最终返回内容仍应保持为子代理最终助手文本，避免破坏 Trellis 主流程。
- 继续保留 Trellis 当前上下文注入能力：active task、prd/design/implement、jsonl spec/research、`TRELLIS_CONTEXT_ID`。
- 失败或取消时仍返回可诊断错误，尽量保留 stderr / 事件尾部。
- 不引入新的全局 npm 依赖，不修改全局 Pi/Trellis 安装目录。

## Acceptance Criteria

- [ ] `.pi/extensions/trellis/index.ts` 的 `subagent` 工具执行时会调用 `onUpdate` 推送实时进度。
- [ ] 子 Pi 进程使用可解析的事件流（JSON mode）收集 message/tool/usage，而不是只等 text mode 最终输出。
- [ ] TUI 中能看到最近工具调用（如 read/bash/edit）和最近输出摘要。
- [ ] parallel / chain 能显示每个子任务或步骤的运行状态。
- [ ] 最终工具结果内容仍为子代理最终输出；错误时包含失败原因。
- [ ] TypeScript 语法检查或等效加载检查通过。

## Out of Scope

- 不替换为第三方 `pi-subagents` 的完整 runtime。
- 不实现后台异步 subagent 状态管理。
- 不改变 Trellis 任务生命周期或工作流规则。
