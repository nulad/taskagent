# TASK-006: Normalize Middleware Error Responses

## Token Safety Verification

- **Task ID & Title:** TASK-006: Normalize Middleware Error Responses
- **Objective:** Ensure middleware never emits plain text or HTML error bodies.
- **Target Files/Scope:** `internal/middleware/request.go`, `internal/middleware/request_test.go`, `internal/middleware/apikey.go`, `internal/middleware/apikey_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - these middleware files and tests are under 450 LOC combined, and only error response behavior is in scope.
  * Dependencies: Existing middleware `ErrorResponse`, auth middleware tests, request ID middleware tests.

## Step-by-step Instructions

1. Replace any `http.Error` usage in middleware with a JSON error response.
2. Keep the response contract as `{"error":"..."}` and `Content-Type: application/json`.
3. Confirm missing API key still returns `401`.
4. Confirm invalid API key still returns `401`.
5. Add or update tests to assert middleware error bodies decode as JSON objects with a non-empty `error`.
6. Do not move handler helper functions into middleware unless needed; a small local middleware helper is acceptable to avoid package cycles.

## Definition of Done

- Middleware error paths return JSON, not plain text.
- Existing auth behavior and request ID behavior continue to pass tests.
- No package cycle is introduced between handler and middleware.
