# TASK-009: Run and Stabilize E2E Test Suite

- **Task ID & Title:** TASK-009: Run and Stabilize E2E Test Suite
- **Objective:** Ensure the completed e2e tests run reliably with `go test ./...` and require no external services.
- **Target Files/Scope:** `internal/handler/integration_test.go`; only touch production files if a test reveals a confirmed defect needed to satisfy T-050.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is a verification and small-fix task; the expected working set is the new integration test file and any single production file implicated by a failing assertion.
  * Dependencies: All prior T-050 tasks, local Go toolchain, isolated temp SQLite database.

## Step-by-step Instructions

1. Run `go test ./...`.
2. If the e2e tests fail because of helper mistakes, fix only `internal/handler/integration_test.go`.
3. If the e2e tests reveal an API contract defect already covered by T-050, inspect the smallest relevant production file and make the minimal fix.
4. Re-run `go test ./...`.
5. Confirm that tests do not depend on a pre-existing database, environment variable, port, or manually seeded API key.

## Definition of Done

- `go test ./...` passes.
- The e2e tests require no external service, existing database, fixed port, or manual API key.
- Any production change made during stabilization is minimal and directly tied to a failing T-050 assertion.
- The final test suite validates the full T-050 workflow end to end.
