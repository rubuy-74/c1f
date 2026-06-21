## Problem Statement

As a developer or operator using Cloudflare Workflows, I currently rely on the web dashboard or static `wrangler` CLI commands to monitor my workflows. The web dashboard requires context-switching out of the terminal, and `wrangler workflows instances list` provides a static, non-interactive output that is difficult to navigate or refresh quickly. I need a terminal-native, interactive way to browse my workflows and monitor their instances in real-time.

## Solution

Implement a `k9s`-like terminal user interface (TUI) dashboard. This "Readonly Dashboard" will allow users to launch the app with a single command, see a list of all their workflows, and drill down into the history of specific instances. The interface will be responsive, support standard TUI navigation (j/k/Enter/Esc), and provide clear visual feedback on instance statuses.

## User Stories

1. As a developer, I want to run `c1f` without arguments to quickly open my workflow dashboard.
2. As an operator, I want to see a list of all my workflows by name so I can find the one I'm interested in.
3. As a developer, I want to see when each workflow was created so I can distinguish between old and new projects.
4. As an operator, I want to press `Enter` on a workflow to see its instance history.
5. As a developer, I want to see the status of instances (Running, Completed, Failed) color-coded so I can identify issues at a glance.
6. As an operator, I want to see how long an instance has been running or took to complete.
7. As a developer, I want to see what triggered an instance (e.g., HTTP, Cron) to understand the context of the run.
8. As a user, I want the most recent instances to be at the top of the list so I don't have to scroll to see the latest activity.
9. As a user, I want running instances to be pinned above completed ones so I can monitor active work easily.
10. As a user, I want to use `Esc` or `b` to navigate back from the instance list to the workflow list.
11. As a user, I want to see a loading spinner when data is being fetched so I know the app hasn't hung.
12. As a user, I want to see an informative message if a workflow has no instances yet.
13. As a developer, I want to see error messages inline if an API call fails (e.g., network error or expired token).
14. As a user, I want to press `Enter` to retry a failed API call and `Esc` to dismiss the error message.
15. As a user, I want the UI to adjust automatically when I resize my terminal window.
16. As a developer, I want to provide my Cloudflare API token via an environment variable for security and convenience.

## Implementation Decisions

- **Framework**: Use Charmbracelet's Bubble Tea for the TUI loop, Lipgloss for styling, and the `list` bubble for efficient rendering of data.
- **Project Structure**: Unify all core logic in the `pkg/` directory (`pkg/api`, `pkg/models`, `pkg/ui`) to maintain a consistent library-like structure.
- **Application Entry Point**: Modify the root Cobra command to launch the TUI if no subcommands are provided.
- **State Management**: Implement a Root Model that manages the transition between `WorkflowList` and `InstanceList` using a simple switch-case update loop.
- **API Client**: Extend the existing `pkg/api.Client` with `ListWorkflows` and `ListInstances` methods, ensuring they inherit the established retry and 429-handling logic.
- **Data Models**: Introduce a `Workflow` model and enhance the `Instance` model to include trigger information and truncated IDs (8 characters).
- **Navigation**: Map `Enter` to "drill down" and `Esc` to "go back" or "dismiss error".
- **Sorting**: Implement a custom sort for the Instance List: `Running` status first, followed by `CreatedAt` descending (most recent first).
- **Pagination**: Support only the first page (50 items) of results from the Cloudflare API for Phase 2 to keep implementation simple.
- **Error Propagation**: Use a dedicated `ErrorMsg` type to communicate API failures from background commands to the UI models.

## Testing Decisions

- **What makes a good test**: Tests should focus on the correctness of data transformation (e.g., sorting logic, relative time formatting) and API client behavior (e.g., retry logic, JSON unmarshaling) rather than UI pixel-perfect rendering.
- **API Client Testing**: Unit tests for `ListWorkflows` and `ListInstances` using mock HTTP servers to verify correct header usage and error handling.
- **Sorting Logic**: Unit tests for the custom "Running-first" sort to ensure instances are ordered correctly regardless of their original API response order.
- **Model Tests**: Extend `pkg/models/workflow_test.go` to cover new fields and methods.

## Out of Scope

- Searching, filtering, or sorting toggles within the TUI (reserved for Phase 4).
- Background polling or automatic refreshing (data is fetched only on view entry for Phase 2).
- Support for Cloudflare GraphQL Analytics (Phase 5).
- Local development support / `--local` flag (Phase 4).
- Writing configuration to disk (environment variables only for now).

## Further Notes

- The binary will remain `c1f`. Running `c1f describe` will continue to work as a standalone CLI tool, while `c1f` will launch the new interactive dashboard.
- We will use standard ANSI colors via Lipgloss to ensure compatibility across most modern terminals.
