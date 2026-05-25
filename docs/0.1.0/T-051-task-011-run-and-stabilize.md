# TASK-011: Run and Stabilize T-051 Suite

## Token Safety Verification

- **Task ID & Title:** TASK-011: Run and Stabilize T-051 Suite
- **Objective:** Run the full Go test suite and make only minimal fixes needed to satisfy T-051.
- **Target Files/Scope:** Expected scope is tests touched by T-051; production files may be touched only when a failing test exposes a confirmed T-051 defect.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is a verification task; inspect only failing test output and the smallest implicated file.
  * Dependencies: All prior T-051 tasks, local Go toolchain, isolated SQLite test databases.

## Step-by-step Instructions

1. Run `gofmt` on modified Go files.
2. Run `go test ./...`.
3. If tests fail because expected status codes changed from pre-T-051 behavior, update the relevant test assertions only when the new expectation matches the T-051 contract.
4. If tests fail because a handler or middleware violates the T-051 contract, inspect the smallest relevant production file and make the minimal fix.
5. Re-run `go test ./...`.
6. Confirm no endpoint in the covered stack returns `text/plain` or an HTML error page for expected errors.

## Definition of Done

- `go test ./...` passes.
- All modified Go files are formatted.
- Any stabilization change is directly tied to input validation, JSON error consistency, or CORS behavior.
