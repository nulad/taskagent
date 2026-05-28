# TASK-009: Implement API Client Error Mapping

## Token Safety Verification

- **Task ID & Title:** TASK-009: Implement API Client Error Mapping
- **Objective:** Map HTTP and network failures from `TaskAgentClient` to stable CLI exceptions and exit codes.
- **Target Files/Scope:** `cli/src/taskagent_cli/api.py`, `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/errors.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this changes the API client and top-level error handling.
  * Dependencies: Successful client wrapper from TASK-008; JSON error contract from the HTTP API server.

## Step-by-step Instructions

1. Extend `TaskAgentClient.request()` to handle non-2xx responses.
2. On 4xx responses, raise `ApiClientError`.
3. On 5xx responses, raise `ApiServerError`.
4. On `httpx.TimeoutException` and transport errors, raise `NetworkError`.
5. Extract useful error text from server JSON fields such as `error`, `message`, and `code` when present.
6. Fall back to response text for malformed or non-JSON error bodies.
7. Keep stdout empty for all error paths.
8. Add top-level Click error handling that maps usage/config to exit 2, 4xx to exit 3, 5xx to exit 4, and network failures to exit 5.
9. Preserve 2xx behavior from TASK-008.

## Definition of Done

- 4xx API errors exit 3.
- 5xx API errors exit 4.
- Network failures exit 5.
- Error text is written to stderr only.
- Malformed JSON error bodies still produce useful messages.
