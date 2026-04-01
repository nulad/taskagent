# TaskAgent API Server — Build Tasks

## Overview

**Project:** TaskAgent — an AI-agent-native task tracker
**Component:** HTTP API Server (Go, stdlib net/http, SQLite)
**Architecture:** handler → service → store, API key auth, REST endpoints
**Estimated effort:** ~20 hours across 4 weeks at ~5 hrs/week

---

## Phase 0: Project Bootstrapping

### T-000: Initialize Go module and directory structure

**Description:**
Create the Go module and the full directory structure as designed. This is the skeleton that every subsequent task builds on. The module path should reflect where the code will live (e.g. `github.com/nulad/taskagent`). Create all directories under `internal/` (handler, middleware, model, service, store) and the `cmd/server/` entry point. Add a `.gitignore` for Go (binaries, vendor, `.db` files). Add a `Makefile` with targets for `build`, `run`, `test`, and `lint`.

**Definition of Done — Task:**
- `go mod init` has been run and `go.mod` exists with the correct module path
- All directories exist: `cmd/server/`, `internal/handler/`, `internal/middleware/`, `internal/model/`, `internal/service/`, `internal/store/`, `migrations/`
- Each package directory contains at least a placeholder `.go` file with the correct `package` declaration (e.g. `package handler`)
- `Makefile` exists with working `build`, `run`, `test`, and `lint` targets
- `.gitignore` excludes binaries, `*.db`, and vendor/
- `go build ./...` succeeds with zero errors
- `go vet ./...` produces no warnings

**Definition of Done — Learning Objective:**
- Understand Go module path conventions: the module path should match the repository URL so that `go get` works correctly for external consumers
- Understand why Go uses `internal/` directories: packages under `internal/` cannot be imported by code outside the module root, enforcing encapsulation at the package level — this is a compiler-enforced boundary, not just convention
- Understand the `cmd/` convention: each subdirectory under `cmd/` produces a separate binary, keeping the `main` package small and focused on wiring dependencies together
- Understand why placeholder files are needed: Go will not compile a package with no `.go` files, so each directory needs at least one file to be valid

---

### T-001: Set up configuration management

**Description:**
Create a `Config` struct that holds all server configuration: listen address, database file path, log level. Load configuration from environment variables using only the standard library (`os.Getenv`). Provide sensible defaults for every field. Do not use a third-party config library (viper, envconfig, etc.) — the stdlib is sufficient and this is a learning exercise. Create a `internal/config/` package for this.

**Definition of Done — Task:**
- `Config` struct exists in `internal/config/config.go`
- Fields: `ListenAddr` (default `:8080`), `DatabasePath` (default `taskagent.db`), `LogLevel` (default `info`)
- A `Load() Config` function reads from env vars: `TASKAGENT_LISTEN_ADDR`, `TASKAGENT_DB_PATH`, `TASKAGENT_LOG_LEVEL`
- Missing env vars fall back to defaults without error
- Unit test verifies defaults are applied when env vars are unset
- Unit test verifies env var overrides work correctly

**Definition of Done — Learning Objective:**
- Understand why Go favors explicit configuration over implicit magic: unlike frameworks that auto-scan config files, Go convention is to make config loading an explicit function call in `main()` so the flow is traceable
- Practice returning structs by value: `Load() Config` returns a value, not a pointer, because Config is small and immutable after construction — understand when value vs pointer receivers/returns are appropriate
- Understand the stdlib approach to env vars: `os.Getenv` returns empty string for unset vars (not an error), and `os.LookupEnv` can distinguish between unset and empty — know when to use each

---

## Phase 1: Database and Store Layer

### T-010: Create SQLite database connection and migration runner

**Description:**
Write the database initialization code that opens a SQLite connection and runs SQL migration files. Migrations live as `.sql` files in the `migrations/` directory, embedded into the binary using `embed.FS`. The migration runner should be simple: execute each migration file in alphabetical order. Track which migrations have been applied in a `_migrations` table. Use `modernc.org/sqlite` as the SQLite driver (pure Go, no CGO dependency).

**Definition of Done — Task:**
- `internal/store/store.go` contains a `Store` struct that wraps `*sql.DB`
- `NewStore(dbPath string) (*Store, error)` opens the database and runs migrations
- WAL mode is enabled via `PRAGMA journal_mode=WAL` and foreign keys via `PRAGMA foreign_keys=ON`
- `migrations/001_initial.sql` is embedded using `//go:embed migrations/*.sql`
- A `_migrations` table tracks applied migration filenames
- Migrations are applied inside a transaction — if any migration fails, none are applied
- Unit test creates an in-memory database (`:memory:`), verifies tables exist after migration
- `Store` has a `Close() error` method that closes the underlying `*sql.DB`

**Definition of Done — Learning Objective:**
- Understand `embed.FS`: Go's `//go:embed` directive embeds files at compile time, producing a single binary with no external file dependencies — this is idiomatic for migrations, templates, and static assets
- Understand why `*sql.DB` is safe for concurrent use: it manages a connection pool internally, so a single `*sql.DB` instance should be shared across the application (not created per-request)
- Practice the constructor pattern: `NewStore()` is Go's convention for constructors — a function that returns an initialized struct, not a method — understand why Go doesn't have constructors on types
- Understand SQLite WAL mode: Write-Ahead Logging allows concurrent readers while a write is in progress, which matters when the API server handles multiple requests simultaneously
- Practice error wrapping: use `fmt.Errorf("opening database: %w", err)` to add context to errors while preserving the original error for `errors.Is()` and `errors.As()` checks

---

### T-011: Implement project store operations

**Description:**
Write the SQL-backed CRUD operations for projects. All functions are methods on the `Store` struct. Each function takes a `context.Context` as its first parameter. Use parameterized queries exclusively (never string interpolation for SQL). Generate UUIDs using `crypto/rand` or a minimal UUID library.

**Definition of Done — Task:**
- `internal/store/project.go` contains methods: `CreateProject(ctx, project) error`, `GetProject(ctx, id) (Project, error)`, `ListProjects(ctx) ([]Project, error)`, `UpdateProject(ctx, project) error`, `DeleteProject(ctx, id) error`
- All methods use `context.Context` and pass it to `db.QueryContext`/`db.ExecContext`
- All SQL uses `?` parameter placeholders, zero string concatenation in queries
- UUIDs are generated as v4 UUIDs
- Timestamps are RFC3339 strings generated via `time.Now().UTC().Format(time.RFC3339)`
- `GetProject` returns a defined sentinel error (e.g. `ErrNotFound`) when no row exists
- `DeleteProject` returns `ErrNotFound` if no row was affected
- Table-driven tests cover: create and retrieve, update existing, update non-existent, delete existing, delete non-existent, list empty, list multiple, create duplicate name returns error
- Tests use an in-memory SQLite database, not mocks

**Definition of Done — Learning Objective:**
- Understand context propagation: every function that does I/O (database, network, file) should accept `context.Context` as its first parameter — this enables request cancellation and timeout propagation from the HTTP layer down to the database query
- Practice sentinel errors: define package-level errors like `var ErrNotFound = errors.New("not found")` so callers can check with `errors.Is(err, store.ErrNotFound)` — understand why this is preferred over checking error strings
- Practice table-driven tests: Go convention is to define test cases as a slice of structs, then loop over them with `t.Run(tc.name, ...)` — this makes adding new cases trivial and test output clear
- Understand `sql.ErrNoRows`: `QueryRow().Scan()` returns `sql.ErrNoRows` when no rows match — wrap this into your own `ErrNotFound` so the store's callers don't depend on `database/sql` types
- Practice testing against real databases instead of mocks: for data access layers, testing against the actual database (even in-memory SQLite) catches SQL errors and constraint violations that mocks would miss

---

### T-012: Implement task store operations

**Description:**
Write the SQL-backed CRUD operations for tasks. Similar pattern to project store but with additional complexity: status validation, filtering by project and status on list operations, and the move operation that updates only the status field. The `ListTasks` function should accept a filter struct rather than many individual parameters.

**Definition of Done — Task:**
- `internal/store/task.go` contains methods: `CreateTask(ctx, task) error`, `GetTask(ctx, id) (Task, error)`, `ListTasks(ctx, filter) ([]Task, error)`, `UpdateTask(ctx, task) error`, `DeleteTask(ctx, id) error`, `UpdateTaskStatus(ctx, id, status) error`
- `TaskFilter` struct has optional fields: `ProjectID *string`, `Status *string`, `Limit int`, `Offset int`
- `ListTasks` dynamically builds WHERE clauses based on which filter fields are non-nil
- All the same context, parameterized query, and error handling patterns from T-011 apply
- `tags` field is stored as JSON text and marshalled/unmarshalled via `json.Marshal`/`json.Unmarshal` in the store layer
- Table-driven tests cover: full CRUD cycle, list with various filter combinations, status update, create task with non-existent project_id returns foreign key error

**Definition of Done — Learning Objective:**
- Practice building dynamic SQL safely: construct WHERE clauses by appending conditions and args to slices (`conditions = append(conditions, "status = ?")`, `args = append(args, filter.Status)`) then joining with AND — never use fmt.Sprintf for SQL
- Understand pointer fields for optional values: `*string` fields in `TaskFilter` use nil to mean "not specified" vs empty string — this is Go's way of representing optional fields without a separate "was this set" boolean
- Practice JSON marshalling with `database/sql`: understand the difference between Go struct fields and database columns, and how to manually handle the serialization boundary (store layer converts `[]string` tags to JSON text for storage and back on retrieval)
- Understand how `sql.NullString` works and when to use it vs pointer types for nullable database columns

---

### T-013: Implement API key store operations

**Description:**
Write the store operations for API key management. Keys are hashed with SHA-256 before storage. The raw key is generated using `crypto/rand`, formatted as `ta_` followed by 32 hex characters, and returned exactly once on creation. The store never sees or returns the raw key — only the hash. A `ValidateKey(ctx, rawKey) (ApiKey, error)` method hashes the provided key and looks it up.

**Definition of Done — Task:**
- `internal/store/apikey.go` contains methods: `CreateApiKey(ctx, label, userName) (rawKey string, error)`, `ValidateKey(ctx, rawKey) (ApiKey, error)`, `ListApiKeys(ctx) ([]ApiKey, error)`, `DeleteApiKey(ctx, id) error`
- Raw key format: `ta_` prefix + 32 hex characters from `crypto/rand`
- Key is hashed with `sha256.Sum256()` before storage
- `ValidateKey` hashes the input and queries by hash — returns `ErrNotFound` if invalid
- `CreateApiKey` returns the raw key string; the `ApiKey` struct in the database never contains the raw key
- `ListApiKeys` returns id, label, user_name, created_at — never the hash
- Tests cover: create and validate, validate with wrong key, delete and then validate fails, list returns metadata only

**Definition of Done — Learning Objective:**
- Understand cryptographic key generation in Go: `crypto/rand.Read()` produces cryptographically secure random bytes — never use `math/rand` for security-sensitive values
- Practice the principle of least information: the store returns the raw key exactly once (on creation) and never stores or returns it again — this pattern is used by GitHub, Stripe, and most API key systems
- Understand hashing vs encryption: SHA-256 is a one-way hash (you can verify but not decrypt), which is correct for API keys — if the database is compromised, the raw keys are not exposed
- Understand `encoding/hex`: use `hex.EncodeToString()` to convert random bytes to a hex string for the key format

---

## Phase 2: Models and Business Logic

### T-020: Define domain models

**Description:**
Create the model structs that are used across layers. These are plain data structs with JSON tags. Define the status type as a custom string type with constants, not bare strings. Keep models in `internal/model/` and ensure they have no dependencies on database or HTTP packages.

**Definition of Done — Task:**
- `internal/model/models.go` contains: `Project`, `Task`, `ApiKey`, `TaskStatus` (custom type), `TaskFilter`
- `TaskStatus` is a custom `type TaskStatus string` with constants: `StatusBacklog`, `StatusTodo`, `StatusInProgress`, `StatusReview`, `StatusDone`
- A `ValidStatus(s string) bool` function validates whether a string is a known status
- All structs have `json:"field_name"` tags using snake_case (matching the JSON output contract for AI agents)
- `Task.Tags` is typed as `[]string` with `json:"tags"`
- Models have no imports from `database/sql`, `net/http`, or any external package
- `TaskStatus` has a `String() string` method (implements `fmt.Stringer`)

**Definition of Done — Learning Objective:**
- Understand custom types for domain safety: `type TaskStatus string` creates a distinct type that prevents accidentally passing any string where a status is expected — the compiler catches `DoSomething("random string")` if the function expects `TaskStatus`
- Practice the `fmt.Stringer` interface: implementing `String()` means your type automatically formats correctly in `fmt.Printf("%s", status)` and log output
- Understand why models should be dependency-free: models are shared across layers (store, service, handler), so if they import `database/sql`, the handler layer transitively depends on `database/sql` — keeping models pure prevents unwanted coupling
- Understand JSON struct tags: `json:"project_id"` controls serialization to JSON, `json:"id,omitempty"` omits the field when empty — this is the contract your AI agent consumers will rely on

---

### T-021: Implement task service with status transition rules

**Description:**
Create the service layer that enforces business rules. The primary rule is status transitions: define which status moves are allowed (e.g. backlog→todo is allowed, backlog→done is not). The service sits between handler and store — handlers call service methods, services call store methods. The service takes a `Store` dependency via its constructor.

**Definition of Done — Task:**
- `internal/service/task.go` contains a `TaskService` struct with a store dependency injected via `NewTaskService(store *store.Store) *TaskService`
- Allowed transitions are defined as a `map[TaskStatus][]TaskStatus` (e.g. `StatusBacklog: {StatusTodo}`, `StatusInProgress: {StatusReview, StatusTodo}`)
- `MoveTask(ctx, id, newStatus) error` validates the transition before calling `store.UpdateTaskStatus`
- If the transition is invalid, return a well-typed error: `type InvalidTransitionError struct { From, To TaskStatus }` with an `Error() string` method
- `CreateTask` validates that the project exists before creating the task
- All other methods (Get, List, Update, Delete) delegate to the store with no additional logic for now
- Unit tests for `MoveTask` cover every valid and invalid transition using table-driven tests
- Tests use a real in-memory store (not a mock interface) since the store is already tested

**Definition of Done — Learning Objective:**
- Practice dependency injection via constructors: `NewTaskService(store)` makes the dependency explicit and testable — understand why this is preferred over global variables or `init()` functions in Go
- Understand custom error types: `InvalidTransitionError` implements the `error` interface and carries structured data — callers can use `errors.As()` to extract it, and the handler can map it to a 422 HTTP status
- Understand when NOT to use interfaces: injecting `*store.Store` directly (not an interface) is pragmatic when there's only one implementation — extract an interface later if you need a second implementation (e.g. for a mock in a different package)
- Practice the thin service layer pattern: not every method needs business logic — it's fine for service methods to delegate directly to the store; the layer exists so logic has a place to go when it's needed, not to force abstraction prematurely

---

### T-022: Implement project service

**Description:**
Create the project service layer. This is thin for now — mostly delegation to the store. The main business rule is: a project cannot be deleted if it has tasks. This prevents orphaned tasks in the database.

**Definition of Done — Task:**
- `internal/service/project.go` contains `ProjectService` struct with store dependency
- `DeleteProject(ctx, id) error` checks for existing tasks before deletion — if tasks exist, return a descriptive error
- All other methods delegate to the store
- Tests verify: delete empty project succeeds, delete project with tasks returns error

**Definition of Done — Learning Objective:**
- Practice composing checks from existing store methods: `DeleteProject` calls `store.ListTasks` with a project filter, then conditionally calls `store.DeleteProject` — this is business logic composition without adding new database queries
- Understand error message quality: the error for "project has tasks" should include the count and project ID so the AI agent or human gets actionable feedback (e.g. `"cannot delete project abc: 5 tasks still exist"`)

---

## Phase 3: HTTP Handlers and Middleware

### T-030: Implement API key authentication middleware

**Description:**
Create middleware that extracts the `X-API-Key` header, validates it against the store, and either allows the request to proceed or returns 401. Inject the authenticated user info into the request context so handlers can access it. Use Go's `http.Handler` middleware pattern (a function that takes a handler and returns a handler).

**Definition of Done — Task:**
- `internal/middleware/apikey.go` contains `func AuthMiddleware(store *store.Store) func(http.Handler) http.Handler`
- Extracts `X-API-Key` from request header
- Returns 401 with a JSON error body if header is missing or key is invalid
- On success, injects the `ApiKey` model into the request context using a custom context key
- A `GetApiKey(ctx context.Context) (ApiKey, bool)` helper retrieves the key from context
- Tests: missing header returns 401, invalid key returns 401, valid key passes through and context contains the key

**Definition of Done — Learning Objective:**
- Understand the middleware pattern in Go: `func(http.Handler) http.Handler` is the standard signature — middleware wraps a handler, optionally modifying the request/response or short-circuiting the chain
- Practice context values correctly: use an unexported custom type for context keys (`type contextKey string`) to avoid collisions — never use bare strings as context keys
- Understand why context values should be request-scoped: the authenticated user is per-request data, which is exactly what `context.Value` is designed for — don't use context for passing dependencies like database connections
- Practice writing JSON error responses: create a helper function `writeError(w, statusCode, message)` that sets `Content-Type: application/json` and writes a consistent error JSON structure — AI agents need parseable errors, not plain text

---

### T-031: Implement request/response helpers

**Description:**
Create shared helper functions for common HTTP handler operations: reading JSON request bodies, writing JSON responses, writing error responses, and parsing URL path parameters. These helpers reduce boilerplate in every handler and enforce consistent response formatting.

**Definition of Done — Task:**
- `internal/handler/helpers.go` contains: `readJSON(r *http.Request, dst any) error`, `writeJSON(w http.ResponseWriter, status int, data any)`, `writeError(w http.ResponseWriter, status int, message string)`, `pathParam(r *http.Request, name string) string`
- `readJSON` enforces a max body size (e.g. 1MB) using `http.MaxBytesReader`
- `readJSON` decodes with `json.Decoder` and calls `DisallowUnknownFields()` to reject unexpected JSON keys
- `writeJSON` sets `Content-Type: application/json` and handles marshal errors gracefully
- `writeError` produces consistent JSON: `{"error": "message here"}`
- `pathParam` wraps `r.PathValue(name)` (Go 1.22+)
- Unit tests for `readJSON`: valid body, oversized body, malformed JSON, unknown fields

**Definition of Done — Learning Objective:**
- Understand `json.Decoder` vs `json.Unmarshal`: Decoder reads from a stream (io.Reader) and is more memory efficient for HTTP bodies; `Unmarshal` requires the full byte slice in memory first
- Practice defensive input handling: `DisallowUnknownFields` rejects payloads with typos or extra fields, which is important for AI agents that might hallucinate field names — fail loudly rather than silently ignore
- Understand `http.MaxBytesReader`: it wraps the request body and returns an error if the client sends more than the limit, preventing memory exhaustion from large payloads
- Understand Go 1.22+ routing: `r.PathValue("id")` extracts named path parameters from patterns like `GET /tasks/{id}` — this is the stdlib alternative to chi/gin parameter extraction

---

### T-032: Implement project handlers

**Description:**
Write the HTTP handlers for project CRUD. Each handler is a method on a handler struct that holds the service dependency. Register routes in a `RegisterRoutes(mux *http.ServeMux)` method. All responses use the JSON helpers from T-031.

**Definition of Done — Task:**
- `internal/handler/project.go` contains `ProjectHandler` struct with `ProjectService` dependency
- Handlers: `handleCreate`, `handleGet`, `handleList`, `handleUpdate`, `handleDelete`
- `RegisterRoutes(mux)` registers: `POST /projects`, `GET /projects/{id}`, `GET /projects`, `PUT /projects/{id}`, `DELETE /projects/{id}`
- Create returns 201 with the created project
- Get returns 200 or 404
- List returns 200 with an array (empty array if no projects, never null)
- Update returns 200 or 404
- Delete returns 204 (no content) or 404 or 409 (conflict, if tasks exist)
- Service errors are mapped to appropriate HTTP status codes (ErrNotFound→404, InvalidTransitionError→422, etc.)
- Integration tests using `httptest.NewServer`: full request/response cycle for each endpoint

**Definition of Done — Learning Objective:**
- Understand `httptest` package: `httptest.NewServer` creates a real HTTP server on a random port for testing — this tests the full HTTP stack including routing, middleware, and serialization without needing an external server
- Practice the handler struct pattern: grouping related handlers as methods on a struct with dependencies is more composable than standalone functions with closure-captured dependencies
- Understand HTTP status code semantics: 201 for resource creation (with Location header), 204 for successful deletion (no body), 409 for conflict — AI agents rely on status codes for control flow, not just error messages
- Practice returning empty arrays: `json.Marshal([]Project{})` produces `[]` while `json.Marshal(nil)` for a nil slice produces `null` — initialize slices to empty when returning lists so agents get consistent types

---

### T-033: Implement task handlers

**Description:**
Write the HTTP handlers for task CRUD and the status move endpoint. Parse query parameters for list filtering. The move endpoint is `PATCH /tasks/{id}/move` with a body of `{"status": "in-progress"}`.

**Definition of Done — Task:**
- `internal/handler/task.go` contains `TaskHandler` struct with `TaskService` dependency
- Handlers: `handleCreate`, `handleGet`, `handleList`, `handleUpdate`, `handleDelete`, `handleMove`
- `RegisterRoutes(mux)` registers all endpoints including `PATCH /tasks/{id}/move`
- `handleList` parses query params: `project_id`, `status`, `limit` (default 50), `offset` (default 0) into a `TaskFilter`
- `handleMove` returns 422 with a descriptive error body when the status transition is invalid
- `handleCreate` returns 422 if required fields (title, project_id) are missing
- Integration tests cover: full CRUD, list with filters, valid move, invalid move, create with missing fields

**Definition of Done — Learning Objective:**
- Practice parsing query parameters safely: use `r.URL.Query().Get("param")` and handle missing/invalid values explicitly with defaults — don't assume query params are present
- Understand PATCH vs PUT semantics: PUT replaces the full resource, PATCH modifies specific fields — the move endpoint is PATCH because it only changes the status, not the entire task
- Practice mapping service errors to HTTP responses: use `errors.As()` to check for `InvalidTransitionError` and return 422, `errors.Is()` for `ErrNotFound` and return 404 — this is the typed error pattern paying off
- Understand the `strconv` package: use `strconv.Atoi` to parse limit/offset from query strings, and handle the error case (return 400 for non-numeric values)

---

### T-034: Implement auth handlers

**Description:**
Write handlers for API key management. Creating a key requires an existing valid key (bootstrap problem is solved by a seed command — see T-040). Listing keys shows metadata only (never hashes). Deleting a key requires a valid key and cannot delete itself.

**Definition of Done — Task:**
- `internal/handler/auth.go` contains `AuthHandler` struct with store dependency
- `POST /auth/keys` creates a new key, returns 201 with `{"key": "ta_...", "id": "...", "label": "..."}` — this is the only time the raw key is returned
- `GET /auth/keys` returns 200 with a list of key metadata (id, label, user_name, created_at)
- `DELETE /auth/keys/{id}` returns 204 or 404; returns 400 if the key being deleted is the same as the one making the request
- All endpoints require valid API key (enforced by middleware, not by the handler)
- Integration tests cover: create key, list keys, delete key, prevent self-deletion

**Definition of Done — Learning Objective:**
- Understand separation of concerns for auth: the handler doesn't check the API key — the middleware already did that and put the identity in context; the handler just retrieves it to check for self-deletion
- Practice the "return once, never again" secret pattern: the response for key creation includes the raw key, but no subsequent API call can retrieve it — understand why this is important for security

---

## Phase 4: Server Wiring and Operations

### T-040: Implement server entry point and dependency wiring

**Description:**
Write `cmd/server/main.go` that wires everything together: load config, open database, create store, create services, create handlers, register routes, apply middleware, start HTTP server. This is where dependency injection happens — manually, without a framework. Also add a `seed` subcommand that creates the first API key for bootstrapping.

**Definition of Done — Task:**
- `cmd/server/main.go` has a clean `main()` that follows the pattern: config → store → services → handlers → mux → middleware → server
- `go run ./cmd/server` starts the server and logs the listen address
- `go run ./cmd/server seed --user "admin" --label "bootstrap"` creates and prints the first API key
- If the database file doesn't exist, it's created automatically (via store/migrations)
- Subcommand parsing uses `os.Args` or `flag` package — no third-party CLI framework needed for this binary
- The mux applies auth middleware to all routes except a `GET /health` endpoint

**Definition of Done — Learning Objective:**
- Understand manual dependency injection in Go: constructing each dependency in `main()` and passing it to the next layer is explicit, traceable, and requires no framework — this is the idiomatic Go approach and the default choice for most Go projects
- Understand why `main()` should be thin: `main` is hard to test because it's the entry point — keep it as a wiring function that creates dependencies and starts the server, with all logic living in testable packages
- Practice the `flag` package for CLI arguments: `flag.String`, `flag.Parse()` is sufficient for simple subcommands — understand when this is enough vs when to reach for cobra
- Understand the health check pattern: an unauthenticated `GET /health` endpoint returning `{"status": "ok"}` is standard for load balancers and monitoring — it should not hit the database

---

### T-041: Implement structured logging

**Description:**
Add structured JSON logging using Go's `log/slog` package (stdlib, available since Go 1.21). Create a request logging middleware that logs method, path, status code, duration, and request ID for every request. Generate a request ID using `crypto/rand` and inject it into the context and response headers.

**Definition of Done — Task:**
- A `slog.Logger` is created in `main()` with `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: configuredLevel})`
- Request logging middleware logs: method, path, status code, duration_ms, request_id
- Request ID is generated as 8 hex characters, added to context and `X-Request-ID` response header
- All store and service errors include the request ID in log output via context
- Log level is configurable via `TASKAGENT_LOG_LEVEL` env var
- The logger is passed as a dependency (not a global) — handlers/services receive it via constructors

**Definition of Done — Learning Objective:**
- Understand `log/slog`: Go 1.21+ provides structured logging in the stdlib — `slog.Info("msg", "key", value)` produces JSON log entries; understand why structured logging (key-value pairs) is preferred over `log.Printf` string formatting for production services
- Practice wrapping `http.ResponseWriter` to capture status codes: create a `responseWriter` struct that embeds `http.ResponseWriter` and overrides `WriteHeader` to record the status code — this is the only way to log the status code in middleware since the stdlib writer doesn't expose it after writing
- Understand logger-as-dependency: passing `*slog.Logger` through constructors instead of using `slog.Default()` makes it possible to inject a test logger or a no-op logger in tests — avoid global loggers

---

### T-042: Implement graceful shutdown

**Description:**
Handle `SIGINT` and `SIGTERM` signals to gracefully shut down the HTTP server. This means: stop accepting new connections, wait for in-flight requests to complete (with a timeout), close the database connection, then exit.

**Definition of Done — Task:**
- `main()` starts the server in a goroutine
- `signal.Notify` listens for `SIGINT` and `SIGTERM`
- On signal: call `server.Shutdown(ctx)` with a 10-second timeout context
- After shutdown: call `store.Close()` to close the database
- Log messages indicate: "shutting down...", "shutdown complete" or "shutdown timed out"
- If shutdown times out, call `server.Close()` to force-kill connections and exit with code 1

**Definition of Done — Learning Objective:**
- Understand graceful shutdown in Go: `http.Server.Shutdown()` stops accepting new connections and waits for active requests to finish — this prevents dropped requests during deployments
- Practice goroutine coordination: the server runs in one goroutine while `main` blocks on signal — this is a fundamental Go concurrency pattern for long-running processes
- Understand `signal.Notify`: Go doesn't handle OS signals by default — you must explicitly register which signals to catch, otherwise SIGTERM kills the process immediately without cleanup
- Understand resource cleanup ordering: close the HTTP server first (stop accepting work), then close the database (release resources) — reverse order of initialization

---

## Phase 5: Testing and Quality

### T-050: Write end-to-end API tests

**Description:**
Create a test file that starts the full server (in-memory database, real HTTP) and exercises the complete workflow an AI agent would perform: create a project, create tasks, move tasks through statuses, list and filter, delete. These tests validate the full stack from HTTP request to database and back.

**Definition of Done — Task:**
- `e2e_test.go` (or `internal/handler/integration_test.go`) starts a full server with `httptest.NewServer`
- Test workflow: seed an API key → create project → create 3 tasks → move task through backlog→todo→in-progress→review→done → list tasks filtered by status → delete task → delete project
- Verify every response status code and body structure
- Verify that invalid transitions return 422 with a parseable error
- Verify that list responses are arrays (never null)
- Verify that auth failures return 401 with a JSON error body
- Tests run with `go test` and require no external setup

**Definition of Done — Learning Objective:**
- Practice test helpers: create functions like `createProject(t, client, name) Project` that reduce boilerplate in integration tests — use `t.Helper()` so test failures report the caller's line number, not the helper's
- Understand `testing.T.Helper()`: marking a function as a test helper changes where failures are reported — without it, every failure points to the helper function instead of the test that called it
- Understand test build tags: if integration tests are slow, use `//go:build integration` to separate them from unit tests — `go test` skips them by default, `go test -tags integration` includes them
- Practice testing the JSON contract: unmarshal responses into maps or structs and assert field presence and types — this is what your AI agent consumers will rely on

---

### T-051: Add input validation and error consistency review

**Description:**
Review all endpoints for consistent error handling and input validation. Every error response should be a JSON object with an `error` field. Add validation for string lengths, required fields, and invalid enum values. Create a validation helper that collects multiple field errors into a single response.

**Definition of Done — Task:**
- Every error response from every endpoint is `{"error": "description"}` with appropriate HTTP status
- 400 for malformed JSON, missing required fields, invalid field values
- 401 for missing/invalid API key
- 404 for resource not found
- 409 for conflict (delete project with tasks)
- 422 for business rule violations (invalid status transition)
- Validation: title max 500 chars, description max 5000 chars, project name max 100 chars, tags max 20 items
- No endpoint returns a plain text error or an HTML error page
- Tests verify error response format for each error category

**Definition of Done — Learning Objective:**
- Understand HTTP status code contracts: AI agents use status codes programmatically (if 404, create the resource; if 409, resolve the conflict; if 422, fix the input) — consistent and correct status codes are more important for agent consumers than for humans
- Practice the error response contract: a consistent `{"error": "message"}` structure means agents can always parse the error the same way — inconsistency here forces agents to handle special cases
- Understand idempotency considerations: `PUT` should be idempotent (putting the same data twice produces the same result) — think about whether `DELETE` on a non-existent resource should return 404 or 204

---

### T-052: Add Dockerfile and Fly.io deployment config

**Description:**
Create a multi-stage Dockerfile that builds the Go binary and produces a minimal runtime image. Add a `fly.toml` configuration for deployment to Fly.io. The SQLite database file should be stored on a Fly volume for persistence.

**Definition of Done — Task:**
- `Dockerfile` uses a multi-stage build: `golang:1.22-alpine` for build, `alpine:latest` for runtime
- Final image contains only the binary and ca-certificates (for any future HTTPS calls)
- `fly.toml` configures: app name, internal port 8080, auto-stop/auto-start machines, a mounted volume for the database file
- `DATABASE_PATH` env var points to the volume mount path
- `fly deploy` succeeds and the health endpoint responds
- Image size is under 20MB

**Definition of Done — Learning Objective:**
- Understand multi-stage Docker builds: the build stage includes the full Go toolchain (~1GB), but the final image copies only the compiled binary (~10-15MB) — this reduces image size, attack surface, and pull time
- Understand CGO_ENABLED=0: setting this ensures a fully static binary (especially important since `modernc.org/sqlite` is pure Go) — the binary runs on `scratch` or `alpine` without glibc
- Understand Fly volumes: Fly.io machines are ephemeral by default — without a volume, the SQLite database is lost on every deploy; a volume provides persistent block storage mounted into the container filesystem
- Practice the `COPY --from=builder` pattern: selectively copy artifacts from the build stage to keep the final image minimal

---

## Summary

| Phase | Tasks | Focus | Est. Hours |
|-------|-------|-------|------------|
| 0 | T-000, T-001 | Project setup | 2 |
| 1 | T-010 – T-013 | Database and store layer | 6 |
| 2 | T-020 – T-022 | Models and business logic | 3 |
| 3 | T-030 – T-034 | HTTP handlers and middleware | 6 |
| 4 | T-040 – T-042 | Server wiring and operations | 3 |
| 5 | T-050 – T-052 | Testing, validation, deployment | 4 |
| **Total** | **18 tasks** | | **~24 hours** |

At ~5 hours/week, this is approximately **5 weeks** of work to a deployed, functional API server. Add 1 week buffer for the unexpected. Phase 3 (React dashboard) and CLI client are separate planning exercises after this is shipped.
