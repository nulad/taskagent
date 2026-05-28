# TASK-015: Add Project Command Tests

## Token Safety Verification

- **Task ID & Title:** TASK-015: Add Project Command Tests
- **Objective:** Add pytest command tests for all project subcommands against mocked HTTP responses.
- **Target Files/Scope:** `cli/tests/test_commands.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - one command test file can verify behavior through Click's test runner and HTTP mocks.
  * Dependencies: Project commands from TASK-011 through TASK-014.

## Step-by-step Instructions

1. Add or extend `cli/tests/test_commands.py`.
2. Use Click's `CliRunner` for command execution.
3. Use mocked HTTP responses or an injected client/transport.
4. Verify `projects list` sends `GET /projects`.
5. Verify `projects show` sends `GET /projects/{id}`.
6. Verify `projects create` sends `POST /projects` with the expected JSON body.
7. Verify `projects update` sends only fields provided by flags.
8. Verify `projects delete` sends `DELETE /projects/{id}` and keeps stdout empty on `204`.
9. Verify API 404 and 409 failures produce exit 3 and stderr output.
10. Verify invalid argument combinations exit 2.

## Definition of Done

- `make -C cli test` runs the project command tests.
- Project payloads are asserted as JSON.
- Command stdout and stderr behavior is covered.
