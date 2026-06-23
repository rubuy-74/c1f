# Phase 5: Advanced Analytics

## Overview

Add on-demand analytics to the Workflow List via a persistent bottom panel, triggered by pressing `a` on a selected workflow. No background polling — data is fetched when the panel opens and on explicit refresh.

## Key Design Decisions

| Decision | Choice |
|----------|--------|
| Fetch mode | On-demand only (no auto-polling) |
| Display | Summary numbers + Unicode sparklines per metric |
| Cost estimation | Raw metrics + computed estimate, configurable via env var, hardcoded fallback |
| Entry point | `a` keybinding from Workflow List opens analytics panel below the selected row |
| Panel persistence | Persistent — stays open until `Esc` |
| Time ranges | 24h (default), 7d, 30d — cycled via `t` |
| GraphQL client | `graphql-go` with separate queries per metric |
| Error handling | Error message + retry via `r` |
| Empty state | 0 with flatline sparkline + "No invocations in last Nh" message |
| Analytics scope | Aggregate metrics only — no instance-level history |
| Cost field | `sum(wallTimeMs)` from `workflowsAdaptiveGroups` |
| GraphQL endpoint | `https://api.cloudflare.com/client/v4/graphql` |
| Credentials | Reuse `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID` required |
| Cost display | Single combined, human-readable (e.g., `~$0.12`), `$` symbol |
| Architecture | Analytics state as sub-state on Workflow List model — not a separate view |

## Implementation Plan

### 5.1 GraphQL Client Setup

- Add `github.com/graphql-go/graphql` dependency.
- Add `github.com/graphql-go/graphql-go` transport for HTTP.
- Create `internal/cloudflare/graphql.go` with a `GraphQLClient` wrapper reusing the existing HTTP client from REST calls.
- Define `WorkflowAnalyticsQuery` that accepts `workflowName`, `accountId`, `timeRange` (24h/7d/30d) and returns `InvocationMetric`, `WallTimeMetric`, `FailureRateMetric` — each with `sum`, `avg`, `buckets[]` (timestamp + value).

### 5.2 Time Range & Bucket Strategy

- `CLOUDFLARE_WORKFLOW_TIME_RANGE` env var: `24h`, `7d`, `30d` (default: `24h`).
- When `24h`: hourly buckets (`quantizedTimeBucket` with 1-hour granularity).
- When `7d` / `30d`: daily buckets.
- Bucket count is dynamic based on range — no fixed bucket count.

### 5.3 Pricing Configuration

- `CLOUDFLARE_CPU_COST_PER_100MS` env var (default: `0.000001` — $1 per 1M 100ms units).
- Fallback to hardcoded constant if env var absent.
- Cost formula: `(sum(wallTimeMs) / 100) * cpuCostPer100ms`.

### 5.4 Sparkline Renderer

- Add `internal/ui/sparkline.go` with `RenderSparkline(values []float64, width int) string`.
- Uses Unicode block characters: ` ▁▂▃▄▅▆▇█`.
- Values are normalized to 0–1 range, mapped to character index.
- Width defaults to 24 (one char per hour for 24h view).

### 5.5 Workflow List Model — Analytics Sub-State

Add to `WorkflowListModel`:

```go
type WorkflowListModel struct {
    // ... existing fields ...
    showAnalytics    bool
    analyticsTimeRange string  // "24h" | "7d" | "30d"
    analyticsData    *AnalyticsData
    analyticsError   error
    analyticsLoading  bool
}
```

- `AnalyticsData` holds invocation count, avg wall-time, fail ratio, CPU time sum, and raw bucket values for each metric.
- `analyticsLoading` gates re-fetching while a request is in-flight.

### 5.6 Keybinding — `a` Toggle

- `a` keybinding on Workflow List: toggles `showAnalytics`.
- When `showAnalytics` becomes `true`: trigger fetch (if not already loading).
- `Esc` keybinding: if `showAnalytics` is true, set `showAnalytics = false`; otherwise bubble up to parent/navigation.

### 5.7 Fetch Flow

- `fetchAnalytics(workflowName, accountId, timeRange)` — runs the three GraphQL queries in goroutines concurrently.
- On success: populate `analyticsData`, clear `analyticsError`.
- On failure: set `analyticsError`, leave `analyticsData` unchanged (preserve last valid data).
- `r` keybinding: re-triggers fetch (same time range).

### 5.8 Time Range Cycling — `t`

- `t` cycles: `24h` → `7d` → `30d` → `24h`.
- On cycle: if `showAnalytics` is true, immediately re-fetch with new range.
- No persistence — resets to `24h` on app restart.

### 5.9 Analytics Panel UI (Render Function)

Render below the workflow list (in the same view's `Render` function, after the list):

```
┌─ analytics: my-workflow ────────────────────────────── [24h ▼] ─┐
│                                                              │
│  Invocations      Avg Wall-time    Failure Rate              │
│  1,234            230ms             2.1%                      │
│  ▁▂▃▄▅▇▇▇▅▃▂▁▁▃▅▇█▇▅▃▂▁   (sparkline per metric)            │
│                                                              │
│  CPU Time (estimated)                                        │
│  12,340,000ms total  ·  ~$0.12 estimated cost                 │
│                                                              │
│  [t] cycle range  [r] refresh  [Esc] close                   │
└──────────────────────────────────────────────────────────────┘
```

- If `analyticsLoading`: show spinner in place of metrics.
- If `analyticsError`: show error message + "Press r to retry".
- If workflow has 0 invocations: show `0` with flatline sparkline + "No invocations in last 24h".
- Sparklines rendered via `RenderSparkline` with width = terminal width / 3 minus padding.
- Time range badge in top-right of panel (`[24h]`, `[7d]`, `[30d]`).
- Panel height is dynamic — 8 rows base + sparkline row.

### 5.10 Three GraphQL Queries

Each query targets `workflowsAdaptiveGroups` with appropriate filters:

**Invocations query** — `sum(invocationCount)` grouped by `quantizedTimeBucket`.
**Wall-time query** — `avg(wallTimeMs)` and `sum(wallTimeMs)` grouped by `quantizedTimeBucket`.
**Failure rate query** — `avg(failRatio)` grouped by `quantizedTimeBucket`.

All three queries include:
```graphql
filter: {
  workflowName: $workflowName
  dateRange: $dateRange  # ONE_DAY, SEVEN_DAYS, THIRTY_DAYS
}
```

### 5.11 Cost Estimation

- `estimatedCost = (analyticsData.sumWallTimeMs / 100.0) * cpuCostPer100ms`
- Display: `~$0.12` format — truncate to 2 decimal places, prefix with `~$`.
- If total is < $0.01: show `~$0.00` (don't show negative or zero unexpectedly).

## File Changes

```
internal/cloudflare/
  + graphql.go          # GraphQL client, query definitions, fetch functions
  + analytics.go        # AnalyticsData struct, cost calculation

internal/ui/
  + sparkline.go        # Unicode sparkline renderer

internal/tui/
  ~ workflow_list.go    # Add analytics sub-state, keybindings, panel rendering
```

## Keybindings Summary (Analytics Panel)

| Key | Action |
|-----|--------|
| `a` | Toggle analytics panel |
| `t` | Cycle time range (24h → 7d → 30d → 24h) |
| `r` | Refresh analytics |
| `Esc` | Close analytics panel |

Global (unchanged):
| Key | Action |
|-----|--------|
| `j/k` | Navigate workflow list |
| `Enter` | Drill into Instance List |
| `q` | Quit |

## Out of Scope

- Auto-refresh polling for analytics.
- Instance-level history in analytics view.
- Config file for time range persistence.
- Storage cost breakdown.
- Customizable currency symbol.
- Count prefixes or half-page scrolling in analytics.
- `h`/`l` pane switching.

## Testing Decisions

- Unit test sparkline normalization: [0, 1, 0.5] → correct char mapping.
- Unit test cost estimation formula with known inputs.
- Unit test `t` cycle logic (24h→7d→30d→24h).
- Verify analytics panel renders (or is hidden) based on `showAnalytics` state.
- Manual QA: fetch data, cycle time ranges, verify error state shows on bad credentials.
