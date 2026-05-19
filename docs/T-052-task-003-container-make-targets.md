# TASK-003: Add Container Make Targets

## Token Safety Verification

- **Task ID & Title:** TASK-003: Add Container Make Targets
- **Objective:** Add Makefile targets that make Docker build, Compose startup, shutdown, logs, and smoke verification discoverable.
- **Target Files/Scope:** `Makefile`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the current Makefile is small and this task only appends container workflow targets.
  * Dependencies: `Dockerfile` from TASK-001, `docker-compose.yml` from TASK-002, existing `build`, `run`, `test`, and `lint` targets.

## Step-by-step Instructions

1. Add container-related targets without changing the behavior of existing `build`, `run`, `test`, or `lint`.
2. Add a target such as `docker-build` that runs `docker build -t taskagent:local .`.
3. Add a target such as `compose-up` that runs `docker compose up --build`.
4. Add a target such as `compose-down` that runs `docker compose down`.
5. Add a target such as `compose-logs` that follows logs for the `taskagent` service.
6. Add a target such as `compose-smoke` that delegates to the smoke verification command created or documented by TASK-005.
7. Mark all new targets as `.PHONY`.
8. Keep commands POSIX-shell compatible and avoid requiring nonstandard Make features.

## Definition of Done

- `make docker-build` builds the image.
- `make compose-up` starts the Compose service.
- `make compose-down` stops the service.
- Existing `make test` and `make lint` still behave as before.
