# TASK-018: Implement Task Update Payloads

## Token Safety Verification

- **Task ID & Title:** TASK-018: Implement Task Update Payloads
- **Objective:** Add `task update ID [--title T] [--description D] [--tags a,b]` with partial JSON body construction.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/models.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends top-level task commands.
  * Dependencies: Payload helpers from TASK-002; task commands from TASK-016.

## Step-by-step Instructions

1. Implement top-level `task update ID`.
2. Define optional `--title`, `--description`, and `--tags`.
3. Require at least one update field.
4. Convert comma-separated tags to a JSON array when `--tags` is provided.
5. Build the JSON payload as a Python dict including only provided fields.
6. Call `PUT /tasks/{id}`.
7. Print the updated task JSON to stdout.

## Definition of Done

- `task update <id> --title T` sends only `title`.
- `task update <id> --tags a,b` sends a JSON array for `tags`.
- Missing ID or missing update fields exits 2.
