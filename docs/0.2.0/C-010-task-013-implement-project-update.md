# TASK-013: Implement Project Update Command

## Token Safety Verification

- **Task ID & Title:** TASK-013: Implement Project Update Command
- **Objective:** Add `task projects update ID [--name NAME] [--description TEXT]` with partial JSON fields.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/models.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this modifies the existing project command group.
  * Dependencies: Project command group from TASK-011; payload helpers from TASK-002.

## Step-by-step Instructions

1. Add `task projects update ID`.
2. Define optional `--name` and `--description` options.
3. Require at least one update field.
4. Build a Python dict containing only fields the user provided.
5. Call `PUT /projects/{id}`.
6. Print the updated project JSON to stdout.
7. Let Click reject unknown flags and extra positional arguments.

## Definition of Done

- `task projects update <id> --name New` sends only `name`.
- `task projects update <id> --description Text` sends only `description`.
- Calling update with no fields exits 2.
