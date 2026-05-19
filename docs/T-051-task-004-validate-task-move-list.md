# TASK-004: Validate Task Move and List Inputs

## Token Safety Verification

- **Task ID & Title:** TASK-004: Validate Task Move and List Inputs
- **Objective:** Make task move and list parameter validation consistent with the T-051 status code contract.
- **Target Files/Scope:** `internal/handler/task.go`, `internal/handler/task_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - the task stays within the existing task handler and task handler tests, under 1,000 LOC combined.
  * Dependencies: TASK-003 task validation conventions, `model.ValidStatus`, existing list and move tests.

## Step-by-step Instructions

1. Validate `PATCH /tasks/{id}/move` request bodies before calling the service.
2. Return `400 Bad Request` when `status` is missing or not a valid public status.
3. Preserve `422 Unprocessable Entity` only for service-level invalid transition errors.
4. Preserve `404 Not Found` for moving a missing task.
5. Keep `GET /tasks` query validation for `limit`, `offset`, and `status`.
6. Confirm invalid `limit`, invalid `offset`, and invalid query `status` all return `400` with JSON error bodies.
7. Add or update tests for missing move status, invalid move status, invalid transition, invalid query status, invalid limit, and invalid offset.

## Definition of Done

- Invalid move/list inputs return `400`.
- Invalid status transitions still return `422`.
- All move/list errors use the `{"error":"..."}` response contract.
