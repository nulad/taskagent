# TASK-006: Verify Compose SQLite Persistence

## Token Safety Verification

- **Task ID & Title:** TASK-006: Verify Compose SQLite Persistence
- **Objective:** Prove SQLite data stored through Compose survives a service restart.
- **Target Files/Scope:** Verification commands; optional `scripts/container-smoke.sh` if extending automation from TASK-005.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only uses Compose, the public API, and the `./data` bind mount.
  * Dependencies: TASK-002, TASK-005, seeded API key, created project ID.

## Step-by-step Instructions

1. Reuse the Compose environment and API key from TASK-005.
2. Record the project ID created during the container API round trip.
3. Confirm `./data/taskagent.db` exists on the host after the API write.
4. Restart the service with `docker compose restart taskagent`.
5. Poll `GET /health` until the restarted service is ready.
6. Send `GET /projects/{id}` with the same `X-API-Key`.
7. Assert the response status is `200`.
8. Assert the response body still contains the same project ID and name.
9. If persistence fails, inspect only `docker-compose.yml`, Dockerfile user/data permissions, and the container logs.

## Definition of Done

- Host `./data/taskagent.db` is created.
- Data written before restart is readable after restart.
- The Compose bind mount is confirmed as the persistence boundary.
