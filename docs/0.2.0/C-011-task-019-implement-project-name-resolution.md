# TASK-019: Implement Project Name Resolution For Tasks

## Token Safety Verification

- **Task ID & Title:** TASK-019: Implement Project Name Resolution For Tasks
- **Objective:** Let task commands accept a project name as well as a project ID for `--project`.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/models.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this adds a resolver used by multiple task commands.
  * Dependencies: Task add/list from TASK-016 and TASK-017; `GET /projects` response shape.

## Step-by-step Instructions

1. Add a helper such as `resolve_project_id(client, value)`.
2. Treat values that look like UUIDs as IDs and return them unchanged.
3. For non-UUID values, call `GET /projects`.
4. Find projects whose `name` exactly matches the provided value.
5. If no project matches, fail before making the task request with `project not found: <name>`.
6. If multiple projects match, fail with `ambiguous project name: <name>` and ask for the ID.
7. Use the resolver in `task add --project` and `task list --project`.
8. Do not change server behavior or add a new endpoint.

## Definition of Done

- `task add --project <uuid>` still sends the UUID directly.
- `task add --project "Project Name"` resolves the project ID via `/projects`.
- `task list --project "Project Name"` resolves before listing.
- Ambiguous or missing names fail before creating or listing tasks.
