# TASK-003: Validate Task Create and Update Requests

## Token Safety Verification

- **Task ID & Title:** TASK-003: Validate Task Create and Update Requests
- **Objective:** Enforce required fields, maximum lengths, public status values, and tag count limits for task create and update requests.
- **Target Files/Scope:** `internal/handler/task.go`, `internal/handler/task_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - `task.go` and `task_test.go` are under 800 LOC combined, and no service/store internals are required.
  * Dependencies: Validation helper from TASK-001, public status validation in `internal/model`, existing task handler tests.

## Step-by-step Instructions

1. Add validation for `POST /tasks`.
2. Require `project_id` and `title` for create requests.
3. Reject `title` longer than 500 characters.
4. Reject `description` longer than 5000 characters.
5. Reject more than 20 tags.
6. If create request includes a non-empty `status`, reject values that are not public valid task statuses.
7. Add validation for `PUT /tasks/{id}` using the same title, description, tag, and status rules.
8. Return `400 Bad Request` for missing required fields and invalid field values.
9. Add tests for missing create fields, too-long title, too-long description, too many tags, and invalid status.

## Definition of Done

- Task create/update validation failures return `400` with JSON error bodies.
- Valid task create/update behavior is unchanged.
- Store and service logic are not used as the primary validation mechanism for request shape.
