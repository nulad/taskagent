# TASK-017: Implement Task List Filters

## Token Safety Verification

- **Task ID & Title:** TASK-017: Implement Task List Filters
- **Objective:** Add `task list` with project, status, limit, and offset query parameter support.
- **Target Files/Scope:** `cli/lib/cmd_tasks.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends the single task command module and one dispatch branch.
  * Dependencies: `api_request`; add/show/delete task command module from TASK-016.

## Step-by-step Instructions

1. Implement `cmd_task_list()` in `cli/lib/cmd_tasks.sh`.
2. Wire top-level `list` dispatch in `cli/task`.
3. Parse optional `--project`, `--status`, `--limit`, and `--offset`.
4. Validate `--limit` and `--offset` as non-negative integers.
5. Construct the query string from provided flags.
6. URL-encode query values safely. Use a small helper if needed, but keep it in the task command module unless it is generally useful.
7. Call `GET /tasks` with the query string.
8. Print the JSON array response to stdout.
9. Do not implement project name resolution in this task; TASK-019 owns it.

## Definition of Done

- `./cli/task list` calls `GET /tasks`.
- `./cli/task list --project P --status todo --limit 10 --offset 20` calls `/tasks` with matching query params.
- Invalid numeric filters exit 2.

