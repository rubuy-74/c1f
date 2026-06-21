# Phase 2 Vertical Slices

This document breaks down the Phase 2 PRD into independently-implementable vertical slices (tracer bullets).

## Slice 1: TUI Foundation & Workflow List
**Type:** AFK  
**Blocked by:** None  
**User stories covered:** 1, 2, 3, 11, 12, 15, 16

### What to build
Scaffold the TUI architecture and implement the top-level Workflow List. This involves extending the API client to list workflows and creating the root Bubble Tea model that launches when `c1f` is run without arguments.

### Acceptance criteria
- [ ] `pkg/api` implements `ListWorkflows()` with existing retry logic.
- [ ] `pkg/ui` contains a root model and a `workflowlist` model using `bubbles/list`.
- [ ] Running `c1f` (no args) opens the TUI and displays a list of workflows (Name, Created At).
- [ ] `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` are correctly used from the environment.
- [ ] Window resizing works and preserves list layout.

---

## Slice 2: Instance List Drill-down & Custom Sorting
**Type:** AFK  
**Blocked by:** Slice 1  
**User stories covered:** 4, 5, 6, 7, 8, 9, 11, 12

### What to build
Implement the transition from the Workflow List to a detailed Instance List for a selected workflow. This includes the API work to fetch instances and the UI logic to display them with custom status colors and specific sorting.

### Acceptance criteria
- [ ] `pkg/api` implements `ListInstances(workflowName)`.
- [ ] Pressing `Enter` on a workflow navigates to the Instance List.
- [ ] Instance List displays Short ID, Status, Started, Duration, and Trigger.
- [ ] Statuses are color-coded (Success/Running = Green, Failure = Red).
- [ ] Sorting logic: Running instances first, then most recent by `CreatedAt`.

---

## Slice 3: Navigation Polish & Error Resilience
**Type:** AFK  
**Blocked by:** Slice 2  
**User stories covered:** 10, 13, 14

### What to build
Complete the navigation loop and implement robust error handling. This slice ensures the user can navigate back to the workflow list and that API failures are handled gracefully within the TUI.

### Acceptance criteria
- [ ] Pressing `Esc` or `b` in the Instance List returns to the Workflow List.
- [ ] API errors trigger an inline error overlay instead of crashing the app.
- [ ] Pressing `Enter` in the error overlay retries the last failed action.
- [ ] Pressing `Esc` in the error overlay dismisses it.
