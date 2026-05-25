# Implementation Plan: T-052 Portable Docker Deployment

## 1. Objective

Package the TaskAgent API server as a small, vendor-neutral OCI container with a local `docker-compose.yml` deployment path, persistent SQLite storage, health checks, and concise generic VPS deployment documentation.

## 2. Scope

T-052 covers deployment packaging only:

1. A production Docker image for `cmd/server`.
2. A Compose file for local and single-host operation.
3. Repository commands that make build, run, smoke, and cleanup workflows repeatable.
4. Documentation for running the image on any Docker-compatible host.

T-052 should not change API behavior, database schema, auth semantics, CORS behavior, or application routing except for the smallest fix required if container verification exposes a real deployment bug.

## 3. Deployment Contract

- The container listens on `:8080` by default.
- Runtime configuration comes from environment variables: `TASKAGENT_LISTEN_ADDR`, `TASKAGENT_DB_PATH`, `TASKAGENT_LOG_LEVEL`, and `TASKAGENT_CORS_ORIGINS`.
- SQLite state lives outside the container at `/data/taskagent.db`.
- `GET /health` is the container health check endpoint.
- The image runs as a non-root user.
- The final runtime image contains only the server binary and runtime necessities such as CA certificates.
- The image remains vendor-neutral: no provider-specific files are required to run it.

## 4. Go Version Note

The original T-052 task text mentions `golang:1.22-alpine`, but this repository currently declares `go 1.25.6` in `go.mod`. The Dockerfile task should use a builder image that matches the repository's declared Go version, or first make an explicit, tested decision to adjust `go.mod`. Do not silently build with an older Go image that cannot satisfy the module directive.

## 5. Atomic Engineering Tasks

The granular task breakdown has been split into one file per task to keep implementation context small. Start from the index: [T-052-atomic-tasks.md](/home/nulad/repo/nulad/taskagent/docs/T-052-atomic-tasks.md).

## 6. Definition of Done Mapping

- Multi-stage Dockerfile with static Go build: TASK-001
- Minimal build context and no accidental database/source noise in the image build: TASK-001
- Non-root runtime user, `EXPOSE 8080`, and `HEALTHCHECK`: TASK-001
- Compose service with port mapping, env vars, and SQLite bind mount: TASK-002
- Repeatable local commands for container workflows: TASK-003
- Container build validation, non-root validation, health check validation, and image size check: TASK-004
- API round-trip through the container using seed + authenticated request: TASK-005
- SQLite persistence after Compose restart: TASK-006
- Generic VPS deployment documentation: TASK-007
- Final stabilization and full verification: TASK-008

## 7. Execution Order

Tasks that author runtime artifacts should run first:

1. TASK-001: Create Docker Image Definition.
2. TASK-002: Create Docker Compose Deployment.
3. TASK-003: Add Container Make Targets.

Verification and documentation can then proceed:

4. TASK-004 must run after TASK-001.
5. TASK-005 must run after TASK-001 and TASK-002.
6. TASK-006 must run after TASK-002 and TASK-005.
7. TASK-007 can run after TASK-002 and may run in parallel with TASK-004.
8. TASK-008 must run last.

## 8. Parallelization

Can run in parallel:

- TASK-001 and TASK-002 can be implemented by separate agents because they write different files. TASK-002 should assume the image name and service command from the plan, then reconcile with TASK-001 during TASK-008.
- TASK-003 can run in parallel with TASK-001 and TASK-002 if it limits itself to `Makefile` targets that call standard Docker commands.
- TASK-007 can run in parallel with TASK-004 because it only writes documentation and uses the deployment contract above.

Must be sequential:

- TASK-004 depends on TASK-001 because it validates the Docker image.
- TASK-005 depends on TASK-001 and TASK-002 because it exercises the containerized API.
- TASK-006 depends on TASK-005 because persistence should be checked after the API round-trip exists.
- TASK-008 depends on every prior task because it reconciles names, commands, docs, and verification output.
