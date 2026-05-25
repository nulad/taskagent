# TASK-015: Add Project Command Tests

## Token Safety Verification

- **Task ID & Title:** TASK-015: Add Project Command Tests
- **Objective:** Add shell tests for all project subcommands against a fake or dev API server.
- **Target Files/Scope:** `cli/test/test_projects.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - one test script can treat the CLI as a black box.
  * Dependencies: Project commands from TASK-011 through TASK-014; `api_request` behavior.

## Step-by-step Instructions

1. Add `cli/test/test_projects.sh`.
2. Use a temporary `TASKAGENT_HOME` with test server config.
3. Prefer a fake server that records method, path, headers, and body. If using the real dev server, isolate test data.
4. Verify `projects list` maps to `GET /projects`.
5. Verify `projects show ID` maps to `GET /projects/{id}`.
6. Verify `projects create --name --description` maps to `POST /projects` with safe JSON.
7. Verify `projects update ID --name --description` maps to `PUT /projects/{id}` with only provided fields.
8. Verify `projects delete ID` maps to `DELETE /projects/{id}`.
9. Verify missing required arguments exit 2.
10. Verify a fake 409 response exits 3 and prints the server error.

## Definition of Done

- `sh cli/test/test_projects.sh` passes.
- Every project command path is covered.
- Tests verify JSON payload shape without brittle whitespace comparisons.

