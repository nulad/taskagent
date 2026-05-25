# TASK-018: Implement Task Update Payloads

## Token Safety Verification

- **Task ID & Title:** TASK-018: Implement Task Update Payloads
- **Objective:** Add `task update ID [--title T] [--description D] [--tags a,b]` with partial JSON body construction.
- **Target Files/Scope:** `cli/lib/cmd_tasks.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends one task command module and one dispatch branch.
  * Dependencies: `jq`; task command module from TASK-016.

## Step-by-step Instructions

1. Implement `cmd_task_update()` in `cli/lib/cmd_tasks.sh`.
2. Wire top-level `update` dispatch in `cli/task`.
3. Require task ID as the first positional argument.
4. Parse optional `--title`, `--description`, and `--tags`.
5. Require at least one update field.
6. Convert comma-separated tags to a JSON array when `--tags` is provided.
7. Build the JSON payload with `jq`, including only provided fields.
8. PUT the payload to `/tasks/{id}` using `api_request`.
9. Print the updated task JSON to stdout.

## Definition of Done

- `./cli/task update <id> --title T` sends only `title`.
- `./cli/task update <id> --tags a,b` sends a JSON array for `tags`.
- Missing ID or missing update fields exits 2.

