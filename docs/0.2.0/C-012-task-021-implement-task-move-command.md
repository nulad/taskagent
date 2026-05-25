# TASK-021: Implement Task Move Command

## Token Safety Verification

- **Task ID & Title:** TASK-021: Implement Task Move Command
- **Objective:** Add `task move ID STATUS` with local status validation and PATCH request behavior.
- **Target Files/Scope:** `cli/lib/cmd_move.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one small command module and one dispatch branch.
  * Dependencies: `api_request`; known status values from the API model.

## Step-by-step Instructions

1. Add `cli/lib/cmd_move.sh`.
2. Define a single status list: `backlog`, `todo`, `in-progress`, `review`, `done`.
3. Implement `cmd_move()` requiring exactly `ID` and `STATUS`.
4. Validate `STATUS` locally against the known list before calling the network.
5. Invalid local status must exit 2 with `invalid status: X (must be one of: backlog, todo, in-progress, review, done)`.
6. Build the JSON body with `jq -n --arg status "$status" '{status: $status}'`.
7. PATCH `/tasks/{id}/move` through `api_request`.
8. Print the moved task JSON to stdout.
9. Let server-side 422 transition errors flow through `api_request` as exit 3.
10. Wire top-level `move` dispatch in `cli/task`.

## Definition of Done

- `./cli/task move <id> todo` PATCHes `/tasks/<id>/move`.
- `./cli/task move <id> in_progress` exits 2 before hitting the network.
- Server 422 errors print the server message and exit 3.

