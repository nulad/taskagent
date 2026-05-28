# TASK-017: Implement Task List Filters

## Token Safety Verification

- **Task ID & Title:** TASK-017: Implement Task List Filters
- **Objective:** Add `task list` with project, status, limit, and offset query parameter support.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/models.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends top-level task commands.
  * Dependencies: `TaskAgentClient`; status constants from TASK-002.

## Step-by-step Instructions

1. Implement top-level `task list`.
2. Add optional `--project`, `--status`, `--limit`, and `--offset`.
3. Validate `--status` against known task statuses.
4. Define `--limit` and `--offset` as non-negative integer Click options.
5. Construct query parameters from provided flags as a Python dict.
6. Let `httpx` handle URL encoding.
7. Call `GET /tasks`.
8. Print the JSON array response to stdout.
9. Do not implement project name resolution in this task; TASK-019 owns it.

## Definition of Done

- `task list` calls `GET /tasks`.
- `task list --project P --status todo --limit 10 --offset 20` sends matching query params.
- Invalid status and invalid numeric filters exit 2.
