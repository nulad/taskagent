# TASK-008: Test Delete Workflow Cleanup

- **Task ID & Title:** TASK-008: Test Delete Workflow Cleanup
- **Objective:** Verify tasks can be deleted before deleting their project, and deleted resources are no longer retrievable.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the task only needs delete route behavior from `internal/handler/task.go` and `internal/handler/project.go` plus the e2e helpers.
  * Dependencies: Project and tasks created through HTTP, `DELETE /tasks/{id}`, `GET /tasks/{id}`, `DELETE /projects/{id}`, `GET /projects/{id}`.

## Step-by-step Instructions

1. Create a project and three tasks through the HTTP API.
2. Delete each task with `DELETE /tasks/{id}` and assert `204`.
3. Fetch one deleted task and assert `404` with a JSON error body.
4. Delete the project with `DELETE /projects/{id}` and assert `204`.
5. Fetch the deleted project and assert `404` with a JSON error body.

## Definition of Done

- The e2e suite deletes tasks through the HTTP API and verifies `204`.
- Fetching a deleted task returns `404` with a JSON error body.
- The project can be deleted after its tasks are deleted.
- Fetching a deleted project returns `404` with a JSON error body.
