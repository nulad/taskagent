# TASK-014: Implement Project Delete Command

## Token Safety Verification

- **Task ID & Title:** TASK-014: Implement Project Delete Command
- **Objective:** Add `task projects delete ID` and rely on API error mapping for conflicts and missing projects.
- **Target Files/Scope:** `cli/lib/cmd_projects.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this modifies one existing project command module.
  * Dependencies: Project command dispatch from TASK-011; `api_request` error mapping from TASK-009.

## Step-by-step Instructions

1. Extend `cmd_projects.sh` with a `projects_delete()` function.
2. Require exactly one project ID argument.
3. Call `api_request DELETE /projects/{id}`.
4. Print the API response body to stdout if the server returns one.
5. Let 404 and 409 errors flow through `api_request`, preserving exit code 3.
6. Reject unknown flags or extra arguments with exit 2.

## Definition of Done

- `./cli/task projects delete <id>` sends `DELETE /projects/<id>`.
- Missing ID exits 2.
- 409 conflict prints the server error to stderr and exits 3.

