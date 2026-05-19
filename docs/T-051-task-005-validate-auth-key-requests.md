# TASK-005: Validate Auth Key Requests

## Token Safety Verification

- **Task ID & Title:** TASK-005: Validate Auth Key Requests
- **Objective:** Make API key creation and deletion input validation return consistent JSON errors and status codes.
- **Target Files/Scope:** `internal/handler/auth.go`, `internal/handler/auth_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - `auth.go` and `auth_test.go` are about 450 LOC combined, so the task remains within the limit.
  * Dependencies: Validation helper from TASK-001, existing auth handler test setup, API key store methods.

## Step-by-step Instructions

1. Validate `POST /auth/keys` with the validation helper.
2. Require `label` and `user_name`.
3. Return `400 Bad Request` for missing `label` or `user_name`.
4. Preserve `404 Not Found` when `user_name` does not refer to an existing user.
5. Keep malformed JSON as `400 Bad Request`.
6. Confirm `DELETE /auth/keys/{id}` returns `400` for a non-integer ID and for attempts to delete the caller’s own key.
7. Add or update tests for missing label, missing user name, unknown user, invalid key ID, and own-key deletion.

## Definition of Done

- Auth handler validation errors return `400` with JSON error bodies.
- Unknown users still return `404`.
- Successful create/list/delete behavior is unchanged.
