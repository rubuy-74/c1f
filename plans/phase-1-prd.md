# PRD: c1f Phase 1 — Hello World of the Cloudflare Workflows API

## Problem Statement

Cloudflare Workflows developers lack a terminal-native tool to quickly inspect the state and progress of a workflow instance. Currently, they must either navigate the Cloudflare web dashboard (breaking keyboard flow) or repeatedly run `wrangler workflows instances describe` (single-shot, no progress calculation). There is no lightweight CLI that fetches and displays the full raw JSON of a workflow instance, with the ability to calculate and report execution progress when a workflow is in progress.

## Solution

A Go CLI tool (`c1f`) that authenticates against the Cloudflare REST API and retrieves the state of a specific workflow instance. The tool outputs the raw JSON response with an injected `calculated_progress` field (e.g., `"40% (2/5 steps)"`) when the workflow is running, providing an immediate, scriptable view of workflow health without leaving the terminal.

## User Stories

1. As a Cloudflare Workflows developer, I want to query a specific workflow instance by name and ID, so that I can inspect its full execution state without opening the web dashboard.
2. As a Cloudflare Workflows developer, I want the tool to read my API token and account ID from environment variables, so that I can use it seamlessly in CI/CD pipelines without hardcoding credentials.
3. As a Cloudflare Workflows developer, I want to see the raw JSON output of the Cloudflare API response, so that I can pipe the output to `jq` or other tools for further processing.
4. As a Cloudflare Workflows developer, I want to see the execution progress of an in-progress workflow (e.g., "40% (2/5 steps)"), so that I can quickly gauge how far along a long-running workflow is without manually counting steps.
5. As a Cloudflare Workflows developer, I want a clear, human-readable error message when an API call fails, so that I can diagnose the problem without deciphering raw HTTP responses.
6. As a Cloudflare Workflows developer, I want to enable a `--debug` flag to see the full raw request/response JSON when troubleshooting, so that I can debug authentication issues or unexpected API behavior.
7. As a Cloudflare Workflows developer, I want the tool to handle API rate limiting (429 responses) gracefully by retrying automatically, so that I don't get spurious failures during high-frequency usage.
8. As a Cloudflare Workflows developer, I want the tool to exit with a non-zero status code on failure, so that scripts and CI pipelines can detect errors.
9. As a Cloudflare Workflows developer, I want the tool to have clear, discoverable `--help` output listing all available flags and their descriptions, so that I can learn how to use it without reading documentation.

## Implementation Decisions

### Architecture

The project follows a `pkg/` + `cmd/` directory layout. Reusable libraries live in `pkg/` and the CLI entry point lives in `cmd/c1f/`. This separation ensures that when Bubble Tea is introduced in Phase 2, the API client and models can be reused without modification.

### Modules

**API Client (`pkg/api`)** — A deep module that encapsulates all Cloudflare API interaction behind a single method: `GetWorkflowInstance(ctx, workflowName, instanceID) (*models.Instance, error)`. Internally it manages:
- HTTP client construction with configurable timeout
- `Authorization: Bearer <token>` and `Content-Type: application/json` headers
- URL construction from account ID, workflow name, and instance ID
- Retry logic for HTTP 429 (Too Many Requests) and 5xx server errors
- JSON unmarshaling of the Cloudflare API response envelope (`{ success, errors, messages, result }`)
- Parsing of Cloudflare error codes and messages into a structured error type

**Models (`pkg/models`)** — A shallow module containing type definitions and pure calculation logic:
- `Instance` struct with `ID`, `Status`, and `Steps` fields (minimal for Phase 1)
- `Step` struct with `Name` and `Status` fields
- `StepStatus` typed string constants for step states
- `Instance.CalculateProgress() (float64, string)` method that computes the ratio of successful steps to total steps

**CLI (`cmd/c1f`)** — A thin cobra-based wiring layer:
- Root command with a `describe` subcommand
- Required flags: `--workflow`, `--instance`
- Optional flag: `--debug` (enables raw request/response logging to stderr)

### Authentication

Authentication credentials are stored in environment variables for CI/CD compatibility:
- `CLOUDFLARE_API_TOKEN` — The Cloudflare API token with Workflows read permission
- `CLOUDFLARE_ACCOUNT_ID` — The Cloudflare account identifier

Workflow target parameters are passed via CLI flags:
- `--workflow` — The name of the workflow
- `--instance` — The ID of the specific workflow instance

### API

The tool calls the Cloudflare REST API v4 endpoint:
```
GET https://api.cloudflare.com/client/v4/accounts/{account_id}/workflows/{workflow_name}/instances/{instance_id}
```

The HTTP client uses Go's standard `net/http` package with a custom `http.Client` configured with a 30-second timeout.

### Output Format

The tool always outputs raw JSON to stdout. If the workflow instance status indicates it is in progress, a `calculated_progress` field (a string like `"40% (2/5 steps)"`) is injected into the root JSON object before output. Error messages are printed to stderr.

### Error Handling

The API client returns a structured error type (`APIError`) containing the Cloudflare error code and message. The CLI layer translates this into a human-readable message on stderr. When `--debug` is enabled, the full raw JSON request/response payload is also printed to stderr.

### Dependencies

- **`github.com/spf13/cobra`** — CLI framework for subcommands, flags, and help generation
- **Go standard library** — `net/http`, `encoding/json`, `context`, `os`, `fmt`, `time`

## Testing Decisions

Testing is deferred for Phase 1. The focus is on a rapid proof-of-concept that can be validated manually against a real Cloudflare Workflows account. When testing is implemented in later phases, the following approach is planned:
- **API Client**: Unit tests using `httptest.NewServer` to mock HTTP responses, verifying URL construction, auth headers, retry behavior, and JSON unmarshaling
- **Models**: Pure unit tests verifying `CalculateProgress()` for various step status combinations

## Out of Scope

- Bubble Tea TUI dashboard (Phase 2)
- Workflow listing and instance navigation (Phase 2)
- Step-by-step inspector view (Phase 3)
- Keyboard shortcuts and interactivity (Phase 4)
- Local development support with `--local` flag (Phase 4)
- Pause/Resume/Terminate workflow lifecycle operations (Phase 4)
- GraphQL Analytics integration (Phase 5)
- Cost monitoring and sparklines (Phase 5)
- Adaptive polling strategy (Phase 5)
- Automated test suites
- Configuration files (`.c1f.yaml` or similar)
- Multi-account support
- Log streaming via wrangler tail

## Further Notes

This phase intentionally keeps the scope minimal to prove the core API interaction pattern. The `pkg/api` client is designed as a deep module with a narrow, stable interface (`GetWorkflowInstance`) that will not change when additional API endpoints (workflow listing, instance listing, lifecycle operations) are added in later phases. The output format (raw JSON with optional injected fields) is designed to be compatible with both human consumption and programmatic parsing, making it suitable for use in CI/CD pipelines and as a building block for the future TUI.