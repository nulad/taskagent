# TASK-005: Test Task List Filtering and Empty Arrays

- **Task ID & Title:** TASK-005: Test Task List Filtering and Empty Arrays
- **Objective:** Verify task list endpoints return JSON arrays and respect status/project filters.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the task only needs `GET /tasks` query behavior from `internal/handler/task.go` and helper functions from prior tasks.
  * Dependencies: Tasks created by TASK-004, `GET /tasks`, `project_id` query parameter, `status` query parameter.

## Step-by-step Instructions

1. Create a project and three tasks through the HTTP API.
2. Move one task to `done` and leave at least one task in `backlog`.
3. Call `GET /tasks?project_id={projectID}` and assert the response is a JSON array with all three project tasks.
4. Call `GET /tasks?project_id={projectID}&status=done` and assert the response is a JSON array containing only the done task.
5. Call `GET /tasks?project_id={projectID}&status=review` when no task is in review and assert the decoded response is a non-nil empty slice.
6. Call `GET /projects` in a fresh harness before creating projects and assert the decoded response is a non-nil empty slice.

## Definition of Done

- `GET /tasks?project_id={projectID}` returns all tasks for the project as a JSON array.
- `GET /tasks?project_id={projectID}&status=done` returns only done tasks.
- A task filter with no matches decodes as a non-nil empty slice, not `null`.
- An empty project list decodes as a non-nil empty slice, not `null`.
- All list response status codes are verified.
