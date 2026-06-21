# c1f Specification: Cloudflare Workflows TUI

## Project Goal
Implement a `k9s`-like terminal user interface (TUI) for monitoring and debugging Cloudflare Workflows, aimed at replacing the static `wrangler` CLI commands and reducing reliance on the web dashboard during development and operations.

## Architecture & Tech Stack
- **Language:** Go
- **Framework:** Bubble Tea (Charmbracelet)
- **Data Source:** Cloudflare REST API (for granular step data) and GraphQL Analytics API.
- **Authentication:** Cloudflare API Token (`CLOUDFLARE_API_TOKEN`).

## Implementation Roadmap (5 Phases)

### Phase 1: Hello World of the Cloudflare's API
- **Objective:** Establish connectivity and data retrieval.
- **Functionality:** A CLI tool that performs a raw call to the Workflow Instances API.
- **Deliverable:** A command that outputs the Raw JSON of a workflow state. If the workflow is in progress, it should calculate and display success progress based on the `steps[]` array.
- **Focus:** PoC of API interaction, authentication, and JSON parsing.

### Phase 2: Readonly Dashboard
- **Objective:** Basic navigation and status monitoring.
- **Functionality:** 
    - **Workflow List:** View all defined workflows in the account.
    - **Instance List:** Pressing `Enter` on a workflow shows historical and currently running instances.
- **Framework:** Implementation of basic Bubble Tea models and message loops.

### Phase 3: The Step Inspector (The "k9s" Moment)
- **Objective:** Deep debugging capabilities.
- **Functionality:** 
    - **Step-by-Step View:** Display workflow progression using the `steps[]` data from the REST API.
    - **Detailed Inspection:** See surface-level configuration (retries, timeouts) and raw stack traces for failed steps.
- **Use Case:** Debugging failed workflows directly in the terminal without clicking through the web UI.

### Phase 4: Interactivity & Local Development Support
- **Objective:** Performance and local-first iteration.
- **4.1 Interactivity:** Keyboard shortcuts (e.g., `j/k` for navigation, `/` for filtering, `q` for quit) to match the `k9s` experience.
- **4.2 Local Development Support:**
    - Support for the `--local` flag to connect to `http://localhost:8787`.
    - **Goal:** Ditch the use of `wrangler workflows instances describe` for local development.
- **Controls:** Add ability to `Pause`, `Resume`, or `Terminate` instances via shortcuts.

### Phase 5: Advanced Analytics
- **Objective:** Production-grade monitoring and cost visibility.
- **Functionality:** 
    - **GraphQL Analytics:** Integrate the `workflowsAdaptiveGroups` dataset.
    - **Metrics:** Display 24-hour invocation counts, average wall-time, and failure rate sparklines.
    - **Cost Monitoring:** Show estimated costs based on CPU time and storage metrics.
- **Outcome:** A comprehensive tool suitable for enterprise-level workflow management.

## Technical Considerations
- **Polling Strategy:** Implement adaptive polling to stay within Cloudflare's API rate limits (1,200 requests / 5 min).
- **Concurrency:** Use Go's goroutines to handle background API updates without blocking the UI.
