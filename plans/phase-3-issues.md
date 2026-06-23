# Phase 3 Implementation Slices (Tracer Bullets)

## Slice 1: Minimal Inspector (The Navigation Bullet)
**Type**: AFK  
**Blocked by**: None  
**User stories covered**: 1, 2, 6, 13  

### What to build
Establish the end-to-end path from the Instance List to the Step Inspector. This includes the model updates for basic step fetching, the navigation transition in `root.go`, and a basic split-pane UI that displays a chronological list of steps or a "Waiting for steps..." placeholder.

### Acceptance criteria
- [ ] `pkg/models` updated with basic `Step` fields (`Name`, `Status`, `StartedAt`).
- [ ] `pkg/api` client supports fetching single instance details with steps.
- [ ] Pressing `Enter` on an Instance transitions to the new `StepInspector` view.
- [ ] UI displays a split-pane with a scrollable list of steps on the left.
- [ ] Empty step arrays show the "Waiting for steps..." placeholder.

---

## Slice 2: Error & Failure Details (The Debugger Bullet)
**Type**: AFK  
**Blocked by**: Slice 1  
**User stories covered**: 5, 7, 8  

### What to build
Implement the deep-dive capability for failures. This adds the `viewport` component to the right pane to handle large stack traces and expands the data model to capture error messages and retry configurations.

### Acceptance criteria
- [ ] `Step` model includes `Error` (Message/Stack) and `Config` (Retries/Timeout).
- [ ] Right pane features a scrollable `viewport` for detailed information.
- [ ] `Tab` toggles focus between the step list and the detail viewport.
- [ ] `w` key toggles line-wrapping in the detail viewport.
- [ ] Stack traces are clearly legible and scrollable in the right pane.

---

## Slice 3: Step Metadata & Outputs (The Context Bullet)
**Type**: AFK  
**Blocked by**: Slice 1  
**User stories covered**: 3, 4, 9, 10  

### What to build
Enhance the visual context of the workflow. Add icons to distinguish step types, implement the metadata header for the instance, and render the successful output of steps.

### Acceptance criteria
- [ ] Step list displays icons/prefixes for `step.do`, `sleep`, and `waitForEvent`.
- [ ] Header section in the right pane shows Instance ID, Version, and Trigger.
- [ ] Successful step outputs/results are rendered in the detail pane.
- [ ] Duration for each step is calculated and displayed (StartedAt vs FinishedAt).

---

## Slice 4: Interactive Management (Refresh & Filtering)
**Type**: AFK  
**Blocked by**: Slice 1  
**User stories covered**: 10, 12, 14  

### What to build
Add interactivity to the inspector. This allows users to manually refresh data and filter the step list to find failures in long-running or complex workflows.

### Acceptance criteria
- [ ] `r` key triggers a manual API refresh of the current instance.
- [ ] `f` key cycles through status filters (All -> Failed -> Running -> Success).
- [ ] API failures during refresh display a non-intrusive error message (toast) rather than crashing the view.
- [ ] Filtered state is visually indicated in the UI.

---

## Slice 5: Raw Data Escape Hatch
**Type**: AFK  
**Blocked by**: Slice 1  
**User stories covered**: 11  

### What to build
Provide the ultimate fallback for debugging by allowing the user to view the entire raw JSON response from the Cloudflare API.

### Acceptance criteria
- [ ] `v` key toggles a full-screen raw JSON view of the Instance.
- [ ] The raw view uses a formatted/indented JSON representation.
- [ ] Pressing `v` or `Esc` returns to the standard split-pane view.
