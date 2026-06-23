## Problem Statement

Developers currently lack a way to see the internal execution flow of a Cloudflare Workflow instance without using the web dashboard. While Phase 2 provides instance status, Phase 3 aims to provide the "k9s moment"—deep, step-level visibility that allows for rapid debugging of failed or stuck workflows directly in the terminal.

## Solution

Implement a split-pane "Step Inspector" view. Users can drill down from an Instance into its individual steps, seeing configuration (retries/timeouts), execution timestamps, outputs, and detailed error stack traces. The interface will prioritize density and readability, mimicking the inspection experience of high-end TUI tools like `k9s`.

## User Stories

1. As a developer, I want to press `Enter` on a workflow instance to see its internal execution steps.
2. As an operator, I want to see a chronological list of all steps executed so far.
3. As a developer, I want to see visual icons distinguishing between standard steps, sleeps, and event waits.
4. As an operator, I want to see the "wall time" and "CPU time" (if available) for each step to identify bottlenecks.
5. As a developer, I want to see the retry configuration and the current attempt count for a failed step.
6. As a user, I want a split-pane view where the right side updates instantly as I navigate the step list on the left.
7. As a developer, I want to see the full stack trace for a failed step in a scrollable viewport.
8. As a user, I want to toggle line-wrapping in the detail pane so I can read long error messages.
9. As a developer, I want to see the result/output payload of a successful step.
10. As a user, I want to press `r` to manually refresh the step data while monitoring a running workflow.
11. As a developer, I want to press `v` to view the raw JSON of the entire instance response for deep debugging.
12. As a user, I want to press `f` to cycle filters (e.g., show only failed steps) to quickly find issues in long workflows.
13. As a developer, I want to see a "Waiting for steps..." message if an instance has started but hasn't recorded any steps yet.
14. As a user, I want to see a non-intrusive error message if a refresh fails, allowing me to keep inspecting my current data.

## Implementation Decisions

- **Expanded Models**: Update `pkg/models/workflow.go` to include comprehensive `Step` metadata (Config, Error, Timestamps, Attempts, Output).
- **Split-Pane UI**: Use a custom Bubble Tea model for the Inspector that manages two distinct layout areas.
- **Header Section**: The detail pane will feature a fixed header showing Instance ID, Version, Trigger, and status.
- **Viewport Integration**: Use the `viewport` bubble for the right-hand detail pane to handle large text efficiently.
- **Navigation Map**:
    - `j/k`: Navigate step list.
    - `Tab`: Switch focus to/from the detail viewport.
    - `r`: Trigger manual API fetch.
    - `v`: Toggle full-screen raw JSON view.
    - `f`: Cycle status filters.
    - `w`: Toggle line-wrapping in the detail viewport.
    - `Esc/b`: Return to Instance List.
- **Visual Style**: Use Lipgloss to create distinct "cards" or "sections" in the detail pane for Config, Error, and Output.
- **Sorting**: Maintain strict chronological order for steps in Phase 3.

## Testing Decisions

- **Mock Data**: Create complex mock API responses containing mixed step types, failures with stack traces, and large output payloads.
- **Filter Logic**: Unit test the status filtering logic to ensure it correctly handles edge cases (e.g., filtering a list with no failures).
- **Metadata Calculation**: Verify duration calculations between `StartedAt` and `FinishedAt`.

## Out of Scope

- Automatic background polling (deferred to Phase 4).
- Sending events to `waitForEvent` steps (Phase 4).
- Parallel step tree visualization (linear chronological list only for now).
- Searching within the stack trace (deferred to Phase 4).
