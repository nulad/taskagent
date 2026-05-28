# TASK-011: Implement Project List And Show Commands

## Token Safety Verification

- **Task ID & Title:** TASK-011: Implement Project List And Show Commands
- **Objective:** Add `task projects list` and `task projects show ID` using the shared Python API client.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one Click command group and two command implementations.
  * Dependencies: `TaskAgentClient` from TASK-008 and TASK-009; project endpoints from the API server.

## Step-by-step Instructions

1. Implement a `projects` Click group.
2. Add `task projects list`.
3. Add `task projects show ID`.
4. For `list`, call `GET /projects`.
5. For `show`, call `GET /projects/{id}`.
6. Print JSON to stdout by default through `print_json()`.
7. Rely on shared API error mapping for 404 and other failures.
8. Leave create, update, and delete placeholders in place until later tasks.
9. Keep `--format human` out of scope until TASK-025.

## Definition of Done

- `task projects list` calls `GET /projects`.
- `task projects show <id>` calls `GET /projects/<id>`.
- Missing or extra `show` arguments exit 2 through Click.
- Output is JSON by default.
