# Phase 5 Implementation Slices (Tracer Bullets)

## Slice 1: GraphQL Client + Invocation Analytics Panel
**Type**: AFK
**Blocked by**: None
**User stories covered**: (Phase 5 — Analytics panel, on-demand metrics)

### What to build
Add `github.com/graphql-go/graphql` dependency and create the GraphQL client in `pkg/api/graphql.go` (reusing the existing HTTP client from the REST API). Define and execute the invocation count query (`sum(invocationCount)` grouped by `quantizedTimeBucket`) for the `workflowsAdaptiveGroups` dataset. Extend `pkg/ui/workflowlist/model.go` with analytics sub-state (`showAnalytics`, `analyticsLoading`, `analyticsError`, `analyticsData`). Wire `a` keybinding to toggle the panel, `r` to refresh, `Esc` to close. Render a basic panel showing the invocation count with a flatline sparkline placeholder (value `0` for all buckets) until wall-time and failure rate are available.

### Acceptance criteria
- [ ] `graphql-go` dependency added to `go.mod`
- [ ] `pkg/api/graphql.go` created with `GraphQLClient` and invocation count query
- [ ] Query uses `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` (same env vars as REST)
- [ ] `WorkflowListModel` gains `showAnalytics bool`, `analyticsLoading bool`, `analyticsError error`, `analyticsData *AnalyticsData`
- [ ] `a` keybinding toggles analytics panel open/closed
- [ ] `r` keybinding re-triggers fetch when panel is open
- [ ] `Esc` closes analytics panel (panel state takes precedence over navigation)
- [ ] Panel renders invocation count with flatline sparkline (24 buckets, all 0)
- [ ] Loading spinner shown while fetching
- [ ] `root.go` updated to pass analytics key events through when panel is open

---

## Slice 2: Wall-Time + Failure Rate + Sparkline Renderer
**Type**: AFK
**Blocked by**: Slice 1
**User stories covered**: (Phase 5 — sparklines, all three metrics)

### What to build
Add the two remaining GraphQL queries: wall-time (`avg(sum wallTimeMs)` per bucket) and failure rate (`avg(failRatio)` per bucket). Extend `AnalyticsData` to hold all three metrics plus their per-bucket values. Create `pkg/ui/common/sparkline.go` with `RenderSparkline(values []float64, width int) string` using Unicode block characters ` ▁▂▃▄▅▆▇█`. Extend the analytics panel to render all three metrics with sparklines below each. Handle empty state: show `0` with flatline + "No invocations in last 24h". Handle error state: show "Failed to load — press r to retry".

### Acceptance criteria
- [ ] `pkg/api/graphql.go` extended with wall-time and failure rate queries
- [ ] `AnalyticsData` holds `InvocationCount`, `AvgWallTimeMs`, `FailRatio`, `SumWallTimeMs`, and bucket values for all three
- [ ] `pkg/ui/common/sparkline.go` with `RenderSparkline` using Unicode blocks
- [ ] All three metrics render with summary number + sparkline in the panel
- [ ] Empty state: `0` invocations + flatline + "No invocations in last 24h"
- [ ] Error state: "Failed to load — press r to retry"
- [ ] Sparklines normalize to 0–1 range, map to correct Unicode chars

---

## Slice 3: Time Range Cycling + Cost Estimation
**Type**: AFK
**Blocked by**: Slice 2
**User stories covered**: (Phase 5 — time range, cost monitoring)

### What to build
Implement `t` keybinding to cycle time range: `24h` → `7d` → `30d` → `24h`. Each cycle re-fetches with the appropriate Cloudflare `dateRange` filter (`ONE_DAY`, `SEVEN_DAYS`, `THIRTY_DAYS`). Native bucket granularity: hourly for 24h (24 buckets), daily for 7d (7 buckets), daily for 30d (30 buckets). Add `CLOUDFLARE_CPU_COST_PER_100MS` env var with hardcoded fallback (`0.000001`). Compute `estimatedCost = (sumWallTimeMs / 100.0) * cpuCostPer100ms` and display as `~$0.12` (human-readable). Add time range badge `[24h]` in panel top-right. Add hint footer `[t] cycle range [r] refresh [Esc] close`.

### Acceptance criteria
- [ ] `t` cycles: `24h` → `7d` → `30d` → `24h`, re-fetches on cycle
- [ ] 24h: hourly buckets, 7d/30d: daily buckets
- [ ] `CLOUDFLARE_CPU_COST_PER_100MS` env var with fallback constant
- [ ] Cost displayed as `~$0.12` format (2 decimal places, `$` prefix)
- [ ] Time range badge shown in panel header
- [ ] Hint footer: `[t] cycle range [r] refresh [Esc] close`
- [ ] No persistence — resets to 24h on app restart
