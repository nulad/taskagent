# TASK-022: Add Task Move Tests

## Token Safety Verification

- **Task ID & Title:** TASK-022: Add Task Move Tests
- **Objective:** Add shell tests for valid move requests, local status typos, server transition errors, and missing tasks.
- **Target Files/Scope:** `cli/test/test_move.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is one focused black-box test file.
  * Dependencies: Move command from TASK-021; `api_request` error mapping.

## Step-by-step Instructions

1. Add `cli/test/test_move.sh`.
2. Use a temporary `TASKAGENT_HOME`.
3. Use a fake HTTP server to record PATCH requests and return fixture responses.
4. Verify a valid status sends `PATCH /tasks/{id}/move` with `{"status":"..."}`.
5. Verify a typo such as `in_progress` exits 2 and does not hit the fake server.
6. Verify a fake 422 response exits 3 and prints the server error.
7. Verify a fake 404 response exits 3 and prints the server error.
8. Verify missing arguments and extra arguments exit 2.

## Definition of Done

- `sh cli/test/test_move.sh` passes.
- The valid request body is asserted as JSON.
- Local validation is proven to avoid the network.

