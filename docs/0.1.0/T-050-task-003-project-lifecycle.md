# TASK-003: Test Project Lifecycle Over HTTP

- **Task ID & Title:** TASK-003: Test Project Lifecycle Over HTTP
- **Objective:** Verify that projects can be created, listed, fetched, and deleted through the real authenticated API.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the task only needs the project route contract from `internal/handler/project.go` plus the helpers from TASK-001 and TASK-002.
  * Dependencies: `POST /projects`, `GET /projects`, `GET /projects/{id}`, `DELETE /projects/{id}`, project JSON fields.

## Step-by-step Instructions

1. Add a test or helper-backed section that creates a project with `name` and `description`.
2. Assert `POST /projects` returns `201` and a JSON object containing non-empty `id`, matching `name`, matching `description`, and timestamp fields.
3. Call `GET /projects/{id}` and assert the returned project matches the created project.
4. Call `GET /projects` and assert the response decodes as a JSON array containing the created project.
5. Delete the project with `DELETE /projects/{id}` after dependent tasks are removed by later workflow tests.
6. Assert project deletion returns `204`.

## Definition of Done

- The e2e suite creates a project through the HTTP API.
- The created project response validates status code, id, name, description, and timestamps.
- Fetching the project by id returns the expected JSON object.
- Listing projects returns a JSON array containing the created project.
- Deleting a project without tasks returns `204`.
