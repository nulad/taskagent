# TASK-020: Add Task Command Tests

## Token Safety Verification

- **Task ID & Title:** TASK-020: Add Task Command Tests
- **Objective:** Add pytest command tests for task CRUD commands, filters, tags, and project name resolution.
- **Target Files/Scope:** `cli/tests/test_commands.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - one command test file can verify behavior through mocked HTTP requests.
  * Dependencies: Task commands from TASK-016 through TASK-019.

## Step-by-step Instructions

1. Extend `cli/tests/test_commands.py`.
2. Use `CliRunner` with isolated environment and config.
3. Use mocked HTTP responses that record request method, path, params, and JSON bodies.
4. Verify `add` sends a `POST /tasks` payload with title, project ID, optional description, and tags array.
5. Verify `list` sends query parameters for project, status, limit, and offset.
6. Verify `show` sends `GET /tasks/{id}`.
7. Verify `update` sends only fields provided by flags.
8. Verify `delete` sends `DELETE /tasks/{id}`.
9. Verify project name resolution succeeds for one matching project.
10. Verify missing and ambiguous project names fail without posting a task.
11. Verify invalid argument combinations exit 2.

## Definition of Done

- `make -C cli test` runs the task command tests.
- Tags-to-JSON-array conversion is covered.
- Query parameter construction is covered.
- Name resolution edge cases are covered.
