# TASK-002: Add HTTP JSON Request Helpers

- **Task ID & Title:** TASK-002: Add HTTP JSON Request Helpers
- **Objective:** Add small helper functions for authenticated JSON requests, response decoding, and JSON error assertions.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only depends on the harness created in TASK-001 and standard library HTTP/JSON behavior.
  * Dependencies: Harness fields from TASK-001, `encoding/json`, `net/http`, `strings` or `bytes`, `testing.T.Helper()`.

## Step-by-step Instructions

1. Add a helper that builds JSON requests against `harness.server.URL + path`.
2. Automatically set `Content-Type: application/json` when a request body is present.
3. Automatically set `X-API-Key` for authenticated helper calls.
4. Add a helper that executes a request, verifies the expected status code, and decodes the JSON response into a caller-provided destination.
5. Add a helper for `204 No Content` responses that verifies the status code and does not decode a body.
6. Add a helper that asserts JSON error responses contain an `error` string field.
7. Ensure every helper calls `t.Helper()` so failures point to the test workflow step.

## Definition of Done

- Authenticated JSON request helpers are available to later e2e tests.
- Helpers set headers consistently, including `Content-Type` and `X-API-Key`.
- JSON responses can be decoded into typed structs, slices, or maps.
- Error responses can be asserted as JSON objects with an `error` string.
- All helper functions call `t.Helper()`.
