# TASK-007: Test Authentication Failure Contract

- **Task ID & Title:** TASK-007: Test Authentication Failure Contract
- **Objective:** Verify missing and invalid API keys return `401` with JSON error bodies.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the task only needs auth middleware behavior from `internal/middleware/apikey.go` and the e2e harness.
  * Dependencies: `GET /projects`, request helper that can omit or override `X-API-Key`, JSON error assertion helper from TASK-002.

## Step-by-step Instructions

1. Extend the request helper or add a narrow helper that can send unauthenticated requests.
2. Send `GET /projects` with no `X-API-Key` header.
3. Assert status `401` and a JSON object with a non-empty `error` string.
4. Send `GET /projects` with an invalid `X-API-Key`.
5. Assert status `401` and a JSON object with a non-empty `error` string.

## Definition of Done

- Missing API key requests to a protected route return `401`.
- Invalid API key requests to a protected route return `401`.
- Both auth failure responses are JSON objects with a non-empty `error` field.
- The test can send requests with no API key and with an overridden API key.
