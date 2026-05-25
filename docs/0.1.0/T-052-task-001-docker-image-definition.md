# TASK-001: Create Docker Image Definition

## Token Safety Verification

- **Task ID & Title:** TASK-001: Create Docker Image Definition
- **Objective:** Add a production Dockerfile and build-context ignore file that produce a small non-root TaskAgent server image.
- **Target Files/Scope:** `Dockerfile`, `.dockerignore`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only needs `go.mod`, `cmd/server/main.go`, and root build context awareness; combined context stays well under 1,000 LOC.
  * Dependencies: Existing `cmd/server` entry point, `go.mod` Go version, existing `/health` route.

## Step-by-step Instructions

1. Create `.dockerignore` to exclude `.git`, local database files, `data/`, build outputs, logs, editor files, and documentation artifacts that are not required for `go build`.
2. Create a multi-stage `Dockerfile`.
3. Use a Go Alpine builder image that matches the Go version declared in `go.mod`; do not use a builder image older than the module directive.
4. In the builder stage, copy `go.mod` and `go.sum` first, download modules, then copy the remaining source.
5. Build `./cmd/server` with `CGO_ENABLED=0` and output a single binary such as `/out/taskagent`.
6. Use a minimal runtime stage such as Alpine, distroless static, or scratch. If Alpine is used, install only `ca-certificates`.
7. Create and run as a non-root user in the runtime stage.
8. Set default env vars for container operation: `TASKAGENT_LISTEN_ADDR=:8080` and `TASKAGENT_DB_PATH=/data/taskagent.db`.
9. Create `/data` and ensure the non-root user can write to it.
10. Add `EXPOSE 8080`.
11. Add a `HEALTHCHECK` that calls `GET /health`. If the runtime image does not contain a shell or HTTP client, use a tiny purpose-built healthcheck approach or choose Alpine with `wget`.
12. Set the default command to run the server binary.

## Definition of Done

- `docker build -t taskagent:local .` succeeds.
- The image starts the API server without extra command-line arguments.
- The container runs as a non-root user.
- `/data/taskagent.db` is writable by the process.
- The final image includes a health check and exposes port 8080.
