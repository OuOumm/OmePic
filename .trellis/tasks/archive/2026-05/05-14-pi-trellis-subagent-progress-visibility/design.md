# Design: Pi Trellis subagent progress visibility

## Scope

Modify the project-local Pi Trellis extension at `.pi/extensions/trellis/index.ts` so the custom Trellis `subagent` tool streams child Pi JSON events into Pi tool partial updates.

## Approach

1. Switch child invocation from `--mode text` to `--mode json`.
2. Parse newline-delimited JSON events from child stdout.
3. Maintain a per-child `SubagentRunState` with:
   - agent / prompt preview / status / startedAt / finishedAt
   - final assistant text
   - recent assistant text tail
   - recent tool calls
   - stderr tail
   - usage totals and turn count
4. Call `onUpdate` whenever a meaningful event arrives:
   - `message_update` text deltas
   - `message_end` assistant messages and usage
   - `tool_execution_start` / `tool_execution_update` / `tool_execution_end`
5. Add `details` to partial/final results so `renderResult` can display compact live status.
6. Add custom `renderCall` / `renderResult` for the Trellis `subagent` tool.

## Mode Handling

- `single`: one state.
- `parallel`: create placeholder states for all prompts and update each independently as its child events arrive.
- `chain`: run sequentially and keep completed states plus the currently running step visible.

## Compatibility

- Keep the public tool schema unchanged (`agent`, `prompt`, `mode`, `prompts`, `model`, `thinking`).
- Keep final `content[0].text` as the final child assistant output / joined outputs, so the parent LLM receives the same kind of result as before.
- Preserve current Trellis prompt construction and context key propagation.

## Risks

- Child Pi JSON event shapes can evolve. Parser should ignore unknown/malformed events and fall back to raw stdout/stderr diagnostics.
- Frequent text deltas can cause excessive UI updates. Throttle updates to a modest interval while still emitting immediately for tool start/end and completion.
