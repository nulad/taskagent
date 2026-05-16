# Implementation Plan: T-050 Write End-to-End API Tests

## 1. Objective

Create end-to-end API tests that start the real HTTP stack with an isolated SQLite database, seed authentication, and verify the complete project and task workflow from HTTP request through database persistence and back.

## 2. Recommended Test Architecture

Use `internal/handler/integration_test.go` in package `handler`.

The test file should build the same stack as `cmd/server/main.go`:

1. Create a temporary SQLite database with `t.TempDir()`.
2. Open `store.NewStore(dbPath, logger)`.
3. Seed a user and API key directly through the store.
4. Construct `ProjectService`, `TaskService`, `ProjectHandler`, `TaskHandler`, and `AuthHandler`.
5. Register project, task, and auth routes on a protected `http.ServeMux`.
6. Wrap protected routes with `middleware.AuthMiddleware(appStore)`.
7. Start `httptest.NewServer(finalHandler)`.
8. Use `http.Client` requests with `X-API-Key` to exercise the real JSON API.

Do not modify production code for this task unless the tests expose a real defect. The first implementation should keep all setup and request helpers inside the integration test file.

## 3. Atomic Engineering Tasks

The granular task breakdown has been split into one file per task to keep the context small for implementation agents. Start from the lightweight index: [T-050-atomic-tasks.md](/home/nulad/repo/nulad/taskagent/docs/T-050-atomic-tasks.md).

## 4. Definition of Done Mapping

- Full server with `httptest.NewServer`: TASK-001
- Seed API key: TASK-001
- Create project and three tasks: TASK-003 and TASK-004
- Move task through statuses: TASK-004
- List and filter tasks: TASK-005
- Delete task and project: TASK-008
- Verify response status codes and body structure: TASK-002 through TASK-008
- Invalid transition returns `422` with parseable error: TASK-006
- List responses are arrays, never null: TASK-005
- Auth failures return `401` with JSON error body: TASK-007
- Tests run with `go test` and no external setup: TASK-009
