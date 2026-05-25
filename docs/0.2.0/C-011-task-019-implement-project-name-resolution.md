# TASK-019: Implement Project Name Resolution For Tasks

## Token Safety Verification

- **Task ID & Title:** TASK-019: Implement Project Name Resolution For Tasks
- **Objective:** Let task commands accept a project name as well as a project ID for `--project`.
- **Target Files/Scope:** `cli/lib/cmd_tasks.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this modifies one command module that should remain under 500 LOC.
  * Dependencies: Task add/list from TASK-016 and TASK-017; `GET /projects` response shape.

## Step-by-step Instructions

1. Add a helper in `cmd_tasks.sh` such as `resolve_project_id()`.
2. Treat values that look like UUIDs as IDs and return them unchanged.
3. For non-UUID values, call `api_request GET /projects`.
4. Use `jq` to find exactly one project whose `.name` matches the provided value.
5. If no project matches, exit 2 or 3 with `project not found: <name>`; choose the code consistently with usage vs API failure semantics.
6. If multiple projects match, fail with `ambiguous project name: <name>` and ask for the ID.
7. Use the resolver in `task add --project` and `task list --project`.
8. Do not change server behavior or add a new endpoint.

## Definition of Done

- `./cli/task add --project <uuid>` still sends the UUID directly.
- `./cli/task add --project "Project Name"` resolves the project ID via `/projects`.
- Ambiguous or missing names fail before creating or listing tasks.

