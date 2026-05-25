# TASK-013: Implement Project Update Command

## Token Safety Verification

- **Task ID & Title:** TASK-013: Implement Project Update Command
- **Objective:** Add `task projects update ID [--name NAME] [--description TEXT]` with partial JSON fields.
- **Target Files/Scope:** `cli/lib/cmd_projects.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this modifies one existing command module.
  * Dependencies: Project command dispatch from TASK-011; `jq`; `api_request`.

## Step-by-step Instructions

1. Extend `cmd_projects.sh` with a `projects_update()` function.
2. Require the project ID as the first positional argument.
3. Parse optional `--name` and `--description` flags.
4. Require at least one update field.
5. Build the JSON body with `jq`, including only fields that were provided.
6. PUT the payload to `/projects/{id}` using `api_request`.
7. Print the updated project JSON to stdout.
8. Reject unknown flags and extra positional arguments with exit 2.

## Definition of Done

- `./cli/task projects update <id> --name New` sends only `name`.
- `./cli/task projects update <id> --description Text` sends only `description`.
- Calling update with no fields exits 2.

