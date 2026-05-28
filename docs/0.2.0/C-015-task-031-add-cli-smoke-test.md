# TASK-031: Add CLI Smoke Test

## Token Safety Verification

- **Task ID & Title:** TASK-031: Add CLI Smoke Test
- **Objective:** Add an end-to-end smoke test that exercises the Python CLI against a freshly started TaskAgent API server.
- **Target Files/Scope:** `cli/tests/test_smoke.py`, `cli/Makefile`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this test needs command contracts plus server startup and seed behavior, but can treat server internals as a black box.
  * Dependencies: Completed CLI commands; Go server `go run ./cmd/server`; API key seed command if available.

## Step-by-step Instructions

1. Add `cli/tests/test_smoke.py` or an equivalent pytest-driven smoke test.
2. Create isolated temporary directories for `TASKAGENT_HOME` and server database state.
3. Start the API server with `go run ./cmd/server` on an available local port.
4. Seed or create an API key using the existing server-supported mechanism.
5. Run the installed or module-invoked `task login` against the local server.
6. Create one project.
7. Create three tasks in that project.
8. Move one task through valid statuses until `done`.
9. List tasks with project and status filters.
10. Delete one task.
11. Delete or clean up remaining tasks, then delete the project.
12. Stop the server on exit, including failure paths.
13. Keep assertions machine-readable and progress logs concise.
14. Wire `make -C cli smoke` to run this smoke test.

## Definition of Done

- `make -C cli smoke` passes on a clean checkout with Go and Python CLI dependencies installed.
- The test starts and stops its own server.
- The workflow covers login, project CRUD, task CRUD, list filters, and move.
