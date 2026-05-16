# TASK-001: Create E2E Test Server Harness

- **Task ID & Title:** TASK-001: Create E2E Test Server Harness
- **Objective:** Add reusable test setup helpers that start the full API stack with an isolated database and seeded API key.
- **Target Files/Scope:** `internal/handler/integration_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - the agent only needs `cmd/server/main.go`, `internal/handler/project.go`, `internal/handler/task.go`, `internal/handler/auth.go`, and the store API used for user/key seeding; each file is small and the relevant setup path is under 1,000 LOC combined.
  * Dependencies: `store.NewStore`, `Store.CreateUser`, `Store.CreateApiKey`, `service.NewProjectService`, `service.NewTaskService`, handler constructors, route registration functions, `middleware.AuthMiddleware`, `httptest.NewServer`.

## Step-by-step Instructions

1. Create or update `internal/handler/integration_test.go` using package `handler`.
2. Define a test harness struct containing `server *httptest.Server`, `client *http.Client`, `apiKey string`, and any cleanup state needed by the test.
3. Implement `newE2EServer(t *testing.T)` and mark it with `t.Helper()`.
4. Inside the helper, create a temp database path using `filepath.Join(t.TempDir(), "test.db")`.
5. Create a quiet test logger, open the store, and register `t.Cleanup` to close the store and test server.
6. Seed an admin user and raw API key directly through the store.
7. Build services, handlers, muxes, auth middleware, request ID middleware, and request logging middleware to match `cmd/server/main.go`.
8. Return the harness with an `httptest.NewServer` URL and authenticated API key.
9. Add a minimal smoke test, such as authenticated `GET /projects`, to verify the harness starts and returns `200` with an array body.

## Definition of Done

- `internal/handler/integration_test.go` exists and compiles.
- `newE2EServer(t)` starts a real `httptest.NewServer` using the real store, services, handlers, and auth middleware.
- The test database is isolated through `t.TempDir()` and requires no external setup.
- A user and API key are seeded inside the helper.
- A smoke test proves an authenticated request reaches the protected API and receives `200`.
