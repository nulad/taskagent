# TASK-004: Verify Docker Image Contract

## Token Safety Verification

- **Task ID & Title:** TASK-004: Verify Docker Image Contract
- **Objective:** Validate that the Docker image satisfies the size, static build, non-root, env, exposed port, and health check requirements.
- **Target Files/Scope:** `Dockerfile` only if fixes are needed; otherwise verification commands only.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - inspect only Docker build output, `docker image inspect`, and the Dockerfile if a check fails.
  * Dependencies: TASK-001, local Docker or compatible container runtime.

## Step-by-step Instructions

1. Build the image with `docker build -t taskagent:local .`.
2. Inspect the image size with `docker image inspect taskagent:local`.
3. Confirm the compressed or local image size target is under 20MB where the chosen runtime base makes that practical; if Alpine pushes it slightly over, document the reason and switch to a smaller runtime if feasible.
4. Confirm the image config exposes port `8080`.
5. Confirm the image config includes a health check.
6. Run `docker run --rm taskagent:local id` only if the image contains `id`; otherwise use `docker image inspect` to verify the configured user is not root.
7. Start the container with `TASKAGENT_DB_PATH=/data/taskagent.db` and a temporary volume or bind mount.
8. Confirm the server logs show it is listening on `:8080`.
9. Confirm `GET /health` returns HTTP `200` and `{"status":"ok"}` from outside the container.
10. If any check fails, make the minimal Dockerfile fix and rerun the relevant checks.

## Definition of Done

- Docker image builds successfully.
- Image runs as non-root.
- Image exposes `8080`.
- Image declares a working health check.
- `/health` succeeds against a running container.
- Image size target is met or a justified exception is documented in the final T-052 verification notes.
