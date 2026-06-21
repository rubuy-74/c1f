# Phase 1 — Vertical Slice Breakdown

Parent PRD: `plans/phase-1-prd.md`

---

## Slice 1: Describe command — happy path (AFK)

### What to build

The minimal end-to-end `c1f describe` command. Initialize the Go project with cobra, set up the CLI with `--workflow` and `--instance` flags, read `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` from env vars, build an API client that makes an authenticated GET request to the Cloudflare REST API, parse the response envelope, and print the raw instance JSON to stdout. Basic error handling only: print Cloudflare error JSON to stderr and exit non-zero.

### Acceptance criteria

- [ ] `go mod init github.com/c1f/c1f` and `pkg/` + `cmd/c1f/` directory structure exist
- [ ] `c1f describe --help` prints flag descriptions for `--workflow` and `--instance`
- [ ] `c1f describe --workflow <name> --instance <id>` prints the raw instance JSON to stdout when auth is valid
- [ ] Missing `CLOUDFLARE_API_TOKEN` or `CLOUDFLARE_ACCOUNT_ID` prints a clear error to stderr and exits non-zero
- [ ] API error response (e.g. invalid workflow name) prints the error JSON to stderr and exits non-zero
- [ ] HTTP client uses a 30-second timeout

### Blocked by

None — can start immediately.

### User stories addressed

- 1 (query instance by name and ID)
- 2 (read token and account ID from env vars)
- 3 (raw JSON output for piping)
- 9 (discoverable `--help` output)

---

## Slice 2: Progress calculation (AFK)

### What to build

Add `CalculateProgress()` to the `Instance` model and inject a `calculated_progress` field into the JSON output when the instance status indicates it is in progress. The field value is a string like `"40% (2/5 steps)"` computed from the ratio of successful steps to total steps.

### Acceptance criteria

- [ ] `Instance.CalculateProgress()` returns `(float64, string)` with correct ratio and formatted string
- [ ] When instance status is `"running"`, `calculated_progress` appears in the JSON output
- [ ] When instance status is `"complete"`, `calculated_progress` is NOT injected
- [ ] When instance has 0 steps, progress displays `"0% (0/0 steps)"` without division by zero
- [ ] When all steps are successful, progress displays `"100% (N/N steps)"`

### Blocked by

Slice 1

### User stories addressed

- 4 (execution progress for in-progress workflows)

---

## Slice 3: Robust error handling + debug (AFK)

### What to build

Add a structured `APIError` type, a simple retry loop for HTTP 429 and 5xx responses, a `--debug` flag that prints raw request/response JSON to stderr, and friendly human-readable error messages to stderr instead of raw JSON on failure.

### Acceptance criteria

- [ ] `--debug` flag is available on the `describe` command and prints raw request URL/headers/body to stderr when set
- [ ] On 429 response, the client retries after a brief delay (up to 3 attempts)
- [ ] On 5xx response, the client retries after a brief delay (up to 3 attempts)
- [ ] On 4xx response (except 429), the client does NOT retry
- [ ] API errors print a friendly message to stderr: `"Error: <message> (Code <code>)"`
- [ ] With `--debug`, the raw error response JSON is also printed to stderr
- [ ] Network errors (DNS, connection refused) print a friendly message to stderr
- [ ] All error paths exit with non-zero status code

### Blocked by

Slice 1 (can run in parallel with Slice 2)

### User stories addressed

- 5 (human-readable error messages)
- 6 (`--debug` flag for raw request/response)
- 7 (automatic retry on rate limiting)
- 8 (non-zero exit on failure)

---

## Dependency Graph

```
Slice 1 (happy path)
├──→ Slice 2 (progress)
└──→ Slice 3 (error handling + debug)
```

Slices 2 and 3 are independent of each other and can be implemented in parallel after Slice 1 is complete.