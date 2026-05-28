# TASK-010: Add API Client Error Tests

## Token Safety Verification

- **Task ID & Title:** TASK-010: Add API Client Error Tests
- **Objective:** Add pytest coverage for every `TaskAgentClient` success and error class.
- **Target Files/Scope:** `cli/tests/test_api.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is one focused test file plus the client API.
  * Dependencies: `TaskAgentClient` from TASK-008 and TASK-009; HTTP mocking dependency from TASK-003.

## Step-by-step Instructions

1. Add `cli/tests/test_api.py`.
2. Use `pytest-httpx`, `respx`, or an injected `httpx.MockTransport`.
3. Verify 200 JSON responses return parsed data.
4. Verify 204 responses return `None`.
5. Verify request headers include `X-API-Key` and `Accept`.
6. Verify JSON request bodies are sent correctly.
7. Verify 400, 401, 404, 409, and 422 responses raise `ApiClientError`.
8. Verify 500 responses raise `ApiServerError`.
9. Verify non-JSON error responses use a readable fallback.
10. Verify timeout and connection failures raise `NetworkError`.

## Definition of Done

- `make -C cli test` runs the API tests.
- Tests do not require the real Go API server.
- Tests assert stdout/stderr separation at the command layer where relevant.
