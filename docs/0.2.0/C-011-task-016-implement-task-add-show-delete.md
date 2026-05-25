# TASK-016: Implement Task Add Show Delete Commands

## Token Safety Verification

- **Task ID & Title:** TASK-016: Implement Task Add Show Delete Commands
- **Objective:** Add `task add`, `task show ID`, and `task delete ID` for core task CRUD access.
- **Target Files/Scope:** `cli/lib/cmd_tasks.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one command module and dispatch branches.
  * Dependencies: `api_request`; task endpoint contract from the API server.

## Step-by-step Instructions

1. Add `cli/lib/cmd_tasks.sh`.
2. Implement `cmd_task_add()`, `cmd_task_show()`, and `cmd_task_delete()`.
3. Wire top-level `add`, `show`, and `delete` dispatch branches in `cli/task`.
4. For `add`, parse `--title`, required `--project`, optional `--description`, and optional `--tags`.
5. Build the create payload with `jq`.
6. For this task, `--project` may be treated as a project ID only; name resolution is TASK-019.
7. Convert comma-separated tags to a JSON array.
8. POST task creation to `/tasks`.
9. Implement `show` as `GET /tasks/{id}`.
10. Implement `delete` as `DELETE /tasks/{id}`.
11. Print raw JSON to stdout.

## Definition of Done

- `./cli/task add --title T --project P` posts to `/tasks`.
- Tags such as `bug,urgent` become `["bug","urgent"]`.
- `./cli/task show <id>` calls `GET /tasks/<id>`.
- `./cli/task delete <id>` calls `DELETE /tasks/<id>`.

