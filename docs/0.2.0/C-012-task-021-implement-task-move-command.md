# TASK-021: Implement Task Move Command

## Token Safety Verification

- **Task ID & Title:** TASK-021: Implement Task Move Command
- **Objective:** Add `task move ID STATUS` with local status validation and PATCH request behavior.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/models.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one Click command and reuses existing helpers.
  * Dependencies: `TaskAgentClient`; known status values from TASK-002.

## Step-by-step Instructions

1. Implement top-level `task move ID STATUS`.
2. Validate `STATUS` locally against the known status list before calling the network.
3. Invalid local status must exit 2 with `invalid status: X (must be one of: backlog, todo, in-progress, review, done)`.
4. Build the JSON body as `{"status": status}`.
5. Call `PATCH /tasks/{id}/move`.
6. Print the moved task JSON to stdout.
7. Let server-side 422 transition errors flow through shared API mapping as exit 3.
8. Expose status values for Click completion in TASK-028.

## Definition of Done

- `task move <id> todo` PATCHes `/tasks/<id>/move`.
- `task move <id> in_progress` exits 2 before hitting the network.
- Server 422 errors print the server message and exit 3.
