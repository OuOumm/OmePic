# Implementation Plan

## Checklist

1. Update `.pi/extensions/trellis/index.ts` imports to include Pi markdown/TUI helpers if needed.
2. Add subagent progress data types and helper functions:
   - JSON event parsing
   - final assistant text extraction from message content
   - tool-call formatting
   - usage aggregation
   - output tail truncation
3. Replace `runPi` internals with JSON-mode child process streaming:
   - spawn `pi --mode json -p --no-session`
   - write prompt to stdin as before
   - parse stdout lines incrementally
   - update run state and call `onUpdate`
   - resolve with final text plus state
4. Thread `onUpdate` through `runSubagent`, `parallel`, and `chain` orchestration.
5. Add `renderCall` / `renderResult` to the registered `subagent` tool for compact/expanded progress rendering.
6. Validate:
   - `cd .pi/npm && npm install` only if dependencies are missing (avoid unless needed)
   - `node --check` is not enough for TS; use `npx tsc --noEmit --allowImportingTsExtensions` only if a tsconfig exists or use Pi reload smoke where possible.
   - At minimum run `npx jiti .pi/extensions/trellis/index.ts` or a focused TypeScript transpile smoke test.
   - `git diff --check`.

## Notes

- Do not edit global Pi installation or upstream Trellis package.
- Keep responses/final output concise and Chinese.
