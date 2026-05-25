# TASK-002: Validate Project Requests

## Token Safety Verification

- **Task ID & Title:** TASK-002: Validate Project Requests
- **Objective:** Enforce required and maximum-length validation for project create and update requests.
- **Target Files/Scope:** `internal/handler/project.go`, `internal/handler/project_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - the two files are roughly 540 LOC combined, and the task only needs the project handler plus its existing tests.
  * Dependencies: Validation helper from TASK-001, `model.Project`, existing project service/store behavior.

## Step-by-step Instructions

1. Add project request validation in `handleCreate` and `handleUpdate`.
2. Require `name` to be non-empty after normal request decoding.
3. Reject `name` longer than 100 characters.
4. Reject `description` longer than 5000 characters.
5. Return `400 Bad Request` for all validation failures, including missing `name`.
6. Keep malformed JSON responses as `400 Bad Request`.
7. Preserve `404` for missing project IDs on update/delete/get.
8. Preserve `409 Conflict` for deleting a project that still has tasks.
9. Add or update handler tests for missing name, too-long name, and too-long description.

## Definition of Done

- Invalid project create/update inputs return `400` with `{"error":"..."}`.
- Valid project create/update behavior is unchanged.
- Existing not-found and conflict behavior remains intact.
