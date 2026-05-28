# TASK-014: Implement Project Delete Command

## Token Safety Verification

- **Task ID & Title:** TASK-014: Implement Project Delete Command
- **Objective:** Add `task projects delete ID` and rely on API error mapping for conflicts and missing projects.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this modifies one existing project command group.
  * Dependencies: Project command group from TASK-011; API error mapping from TASK-009.

## Step-by-step Instructions

1. Add `task projects delete ID`.
2. Require exactly one project ID argument through Click.
3. Call `DELETE /projects/{id}`.
4. Print no output on `204` or `None`.
5. If the server returns a JSON body, print it as JSON.
6. Let 404 and 409 errors flow through shared API mapping as exit code 3.

## Definition of Done

- `task projects delete <id>` sends `DELETE /projects/<id>`.
- Missing ID exits 2.
- `204` success keeps stdout empty.
- 409 conflict prints the server error to stderr and exits 3.
