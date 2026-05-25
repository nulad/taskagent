# TASK-006: Test Invalid Transition Error Contract

- **Task ID & Title:** TASK-006: Test Invalid Transition Error Contract
- **Objective:** Verify invalid task status transitions return `422` with a parseable JSON error body.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the task only needs `PATCH /tasks/{id}/move` behavior and the service transition rule already exposed through the handler.
  * Dependencies: A task created through `POST /tasks`, invalid direct transition such as `backlog` to `done`, JSON error assertion helper from TASK-002.

## Step-by-step Instructions

1. Start a fresh e2e harness.
2. Create a project and one task through the HTTP API.
3. Attempt to move the task directly from `backlog` to `done`.
4. Assert the response status is `422 Unprocessable Entity`.
5. Decode the body as a JSON object and assert it contains a non-empty `error` field.
6. Fetch the task and assert its status remains `backlog`.

## Definition of Done

- An invalid direct transition such as `backlog -> done` is covered by an e2e test.
- The invalid transition returns HTTP `422`.
- The error response is parseable JSON with a non-empty `error` field.
- The task status remains unchanged after the failed transition.
