# TASK-011: Implement Project List And Show Commands

## Token Safety Verification

- **Task ID & Title:** TASK-011: Implement Project List And Show Commands
- **Objective:** Add `task projects list` and `task projects show ID` using the shared API request helper.
- **Target Files/Scope:** `cli/lib/cmd_projects.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one command module and one dispatch branch.
  * Dependencies: `api_request` from TASK-008 and TASK-009; project endpoints from the API server.

## Step-by-step Instructions

1. Add `cli/lib/cmd_projects.sh`.
2. Implement `cmd_projects()` as a sub-dispatch for `list`, `show`, `create`, `update`, and `delete`.
3. In this task, implement only `list` and `show`; make other project actions return `not implemented` exit 2 until later tasks fill them in.
4. Implement `task projects list` as `GET /projects`.
5. Implement `task projects show ID` as `GET /projects/{id}`.
6. Require exactly one ID argument for `show`.
7. Print raw JSON to stdout.
8. Rely on `api_request` for 404 and other error mapping.
9. Wire `projects` dispatch in `cli/task`.

## Definition of Done

- `./cli/task projects list` calls `GET /projects`.
- `./cli/task projects show <id>` calls `GET /projects/<id>`.
- Missing or extra `show` arguments exit 2.
- Output is JSON by default.

