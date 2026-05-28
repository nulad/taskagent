# TASK-016: Implement Task Add Show Delete Commands

## Token Safety Verification

- **Task ID & Title:** TASK-016: Implement Task Add Show Delete Commands
- **Objective:** Add `task add`, `task show ID`, and `task delete ID` for core task CRUD access.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/models.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds three Click commands.
  * Dependencies: `TaskAgentClient`; task endpoint contract from the API server.

## Step-by-step Instructions

1. Implement top-level `task add`.
2. Implement top-level `task show ID`.
3. Implement top-level `task delete ID`.
4. For `add`, define required `--title`, required `--project`, optional `--description`, and optional `--tags`.
5. Build the create payload as a Python dict.
6. For this task, `--project` may be treated as a project ID only; name resolution is TASK-019.
7. Convert comma-separated tags to a JSON array using `parse_tags()`.
8. Call `POST /tasks` for creation.
9. Implement `show` as `GET /tasks/{id}`.
10. Implement `delete` as `DELETE /tasks/{id}`.
11. Print JSON to stdout for responses with bodies; keep stdout empty on `204`.

## Definition of Done

- `task add --title T --project P` posts to `/tasks`.
- Tags such as `bug,urgent` become `["bug", "urgent"]`.
- `task show <id>` calls `GET /tasks/<id>`.
- `task delete <id>` calls `DELETE /tasks/<id>`.
