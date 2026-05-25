# TASK-008: Run and Stabilize T-052 Suite

## Token Safety Verification

- **Task ID & Title:** TASK-008: Run and Stabilize T-052 Suite
- **Objective:** Run all T-052 verification steps and make only minimal fixes needed for portable deployment.
- **Target Files/Scope:** Expected scope is `Dockerfile`, `.dockerignore`, `docker-compose.yml`, `Makefile`, `docs/deployment.md`, and optional smoke script if created.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this task reviews only deployment artifacts plus failing command output; it should not require reading broad application code.
  * Dependencies: All prior T-052 tasks, Docker or compatible runtime, `curl`, local Go toolchain for existing tests.

## Step-by-step Instructions

1. Run `go test ./...` to confirm deployment work did not break application behavior.
2. Run `docker compose config`.
3. Run the Docker image build target or `docker build -t taskagent:local .`.
4. Run the image contract checks from TASK-004.
5. Run the API round-trip checks from TASK-005.
6. Run the persistence checks from TASK-006.
7. Review `docs/deployment.md` and ensure command names, service names, image names, env vars, and paths match the implemented files.
8. Run `docker compose down` at the end. Use `--volumes` only if the task explicitly intends to remove local test state.
9. If a check fails, inspect the smallest implicated deployment artifact and make the minimal correction.
10. Do not alter API behavior, schema, or business logic unless a failing check exposes a confirmed production defect.

## Definition of Done

- `go test ./...` passes.
- `docker compose config` passes.
- Image builds successfully.
- Container health check succeeds.
- Authenticated API smoke test succeeds through Compose.
- SQLite persistence survives Compose restart.
- Deployment docs match the final commands and file names.
