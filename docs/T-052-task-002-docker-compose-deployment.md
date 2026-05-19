# TASK-002: Create Docker Compose Deployment

## Token Safety Verification

- **Task ID & Title:** TASK-002: Create Docker Compose Deployment
- **Objective:** Add a Compose file that runs TaskAgent locally with persistent SQLite storage and the correct runtime environment.
- **Target Files/Scope:** `docker-compose.yml`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only needs the T-052 deployment contract and the known server env vars; no Go source changes are required.
  * Dependencies: Docker image name from TASK-001, existing env vars from `internal/config/config.go`, existing `/health` route.

## Step-by-step Instructions

1. Create `docker-compose.yml` with one service named `taskagent`.
2. Configure the service to build from the repository root and tag the image as `taskagent:local`.
3. Map host port `8080` to container port `8080`.
4. Set `TASKAGENT_LISTEN_ADDR=:8080`.
5. Set `TASKAGENT_DB_PATH=/data/taskagent.db`.
6. Set `TASKAGENT_LOG_LEVEL=info`.
7. Include `TASKAGENT_CORS_ORIGINS` with an empty default or environment interpolation that defaults to empty.
8. Bind-mount `./data` to `/data`.
9. Add a Compose healthcheck if the Dockerfile healthcheck cannot be relied on by the installed Compose version.
10. Avoid adding unrelated services, external networks, reverse proxies, or provider-specific configuration.

## Definition of Done

- `docker compose config` succeeds.
- `docker compose up --build` starts exactly one service.
- The service writes SQLite state under local `./data`.
- The service is reachable on `http://localhost:8080/health`.
