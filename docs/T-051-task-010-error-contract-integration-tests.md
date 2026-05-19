# TASK-010: Add Error Contract Integration Tests

## Token Safety Verification

- **Task ID & Title:** TASK-010: Add Error Contract Integration Tests
- **Objective:** Add end-to-end tests proving every T-051 error category returns the expected status and JSON error format.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this task only extends the existing integration test file, which is under 700 LOC, and relies on its existing server helpers.
  * Dependencies: Validation changes from TASK-002 through TASK-005, middleware normalization from TASK-006, CORS wiring from TASK-009.

## Step-by-step Instructions

1. Add an assertion helper that verifies a response body is a JSON object with a non-empty `error` field.
2. Add a `400` malformed JSON test against a protected route.
3. Add a `400` validation test, such as creating a project with no name.
4. Add a `401` test for missing or invalid API key if not already covered by T-050.
5. Add a `404` test for a missing task or project.
6. Add a `409` test for deleting a project that still has tasks.
7. Add a `422` test for an invalid task transition if not already covered by T-050.
8. Add CORS integration assertions for allowed preflight, disallowed preflight, and allowed actual request headers.
9. Keep tests independent: each test should create its own project/task data through helpers or store setup.

## Definition of Done

- Integration tests cover `400`, `401`, `404`, `409`, and `422` JSON error responses.
- CORS behavior is verified through the full HTTP stack.
- Tests do not require external services, a fixed port, or manually seeded credentials.
