# TASK-004: Test Task Creation and Status Workflow

- **Task ID & Title:** TASK-004: Test Task Creation and Status Workflow
- **Objective:** Verify that an authenticated client can create tasks and move one task through the valid status sequence.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - the task only needs task route behavior from `internal/handler/task.go`, task status constants from `internal/model/models.go`, and the existing e2e helpers.
  * Dependencies: Project created by TASK-003, `POST /tasks`, `PATCH /tasks/{id}/move`, `GET /tasks/{id}`, valid statuses `backlog`, `todo`, `in-progress`, `review`, `done`.

## Step-by-step Instructions

1. Create a project through the HTTP API.
2. Create three tasks in that project with distinct titles.
3. Assert each `POST /tasks` returns `201`, a non-empty `id`, matching `project_id`, matching `title`, and an initial status of `backlog`.
4. Pick one task and move it sequentially with `PATCH /tasks/{id}/move` to `todo`, `in-progress`, `review`, and `done`.
5. After each move, assert `200` and verify the response body reports the requested status.
6. Fetch the moved task with `GET /tasks/{id}` and assert its final status is `done`.

## Definition of Done

- The e2e suite creates three tasks through the HTTP API.
- Each task creation response validates status code, id, project id, title, and initial `backlog` status.
- One task is moved through `backlog -> todo -> in-progress -> review -> done`.
- Each valid move returns `200` and the requested status in the response body.
- Fetching the moved task confirms its final status is `done`.
