# TASK-010: Add API Request Error Tests

## Token Safety Verification

- **Task ID & Title:** TASK-010: Add API Request Error Tests
- **Objective:** Add shell tests that exercise every `api_request` success and error class against a fake local server.
- **Target Files/Scope:** `cli/test/test_api_request.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is one focused test script plus the common helper API.
  * Dependencies: `api_request` from TASK-008 and TASK-009; `python3` or `nc` for a fake server.

## Step-by-step Instructions

1. Add `cli/test/test_api_request.sh`.
2. Use a temporary `TASKAGENT_HOME` with a valid config file.
3. Start a tiny fake HTTP server that can return: 200 JSON, 400 JSON error, 404 JSON error, 500 JSON error, and invalid non-JSON 500 body.
4. Verify 2xx responses print the body to stdout and exit 0.
5. Verify 4xx responses exit 3 and print the `.error` message to stderr.
6. Verify 5xx JSON responses exit 4 and print the `.error` message to stderr.
7. Verify 5xx non-JSON responses exit 4 and print a useful fallback.
8. Verify a connection to an unused port exits 5 and prints `network error:`.
9. Stop the fake server and clean up temporary files on exit.

## Definition of Done

- `sh cli/test/test_api_request.sh` passes consistently.
- Tests do not require the real Go API server.
- Tests assert stdout and stderr are separated correctly.

