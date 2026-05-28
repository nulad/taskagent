# TASK-022: Add Task Move Tests

## Token Safety Verification

- **Task ID & Title:** TASK-022: Add Task Move Tests
- **Objective:** Add pytest command tests for valid move requests, local status typos, server transition errors, and missing tasks.
- **Target Files/Scope:** `cli/tests/test_commands.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends the command test file.
  * Dependencies: Move command from TASK-021; API error mapping.

## Step-by-step Instructions

1. Extend `cli/tests/test_commands.py`.
2. Verify a valid status sends `PATCH /tasks/{id}/move` with `{"status": "..."}`.
3. Verify a typo such as `in_progress` exits 2 and does not hit the mocked transport.
4. Verify a fake 422 response exits 3 and prints the server error.
5. Verify a fake 404 response exits 3 and prints the server error.
6. Verify missing arguments and extra arguments exit 2.

## Definition of Done

- `make -C cli test` runs the move tests.
- The valid request body is asserted as JSON.
- Local validation is proven to avoid the network.
