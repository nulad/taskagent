# TASK-001: Create Validation Error Helper

## Token Safety Verification

- **Task ID & Title:** TASK-001: Create Validation Error Helper
- **Objective:** Add a small handler-layer validation helper that collects multiple field errors and writes them as one JSON error response.
- **Target Files/Scope:** `internal/handler/helpers.go`, `internal/handler/helpers_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - `helpers.go` is about 50 LOC and `helpers_test.go` is about 100 LOC, so the full working set is far below 1,000 LOC.
  * Dependencies: Existing `writeJSON`, `writeError`, `readJSON`, and handler test conventions.

## Step-by-step Instructions

1. Add a lightweight validation type in `internal/handler/helpers.go`, such as `validationErrors`, that can collect field-specific messages.
2. Provide methods to add an error, check whether any errors exist, and build one stable message string.
3. Keep the response shape as `{"error":"..."}`; do not introduce a new top-level `fields` JSON contract for T-051.
4. Add a helper that writes validation errors with `400 Bad Request`.
5. Preserve existing `writeError` behavior for non-validation errors.
6. Add unit tests covering empty validation errors, one field error, multiple field errors, and the JSON response produced by the helper.

## Definition of Done

- Handlers can collect multiple validation failures before returning.
- Validation failures still return exactly one JSON object with an `error` field.
- Tests verify the helper output is deterministic enough for assertions.
