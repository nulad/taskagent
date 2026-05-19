# TASK-005: Verify Container API Round Trip

## Token Safety Verification

- **Task ID & Title:** TASK-005: Verify Container API Round Trip
- **Objective:** Prove the containerized server can seed an API key and complete an authenticated project creation request through the published port.
- **Target Files/Scope:** Verification commands; optional `scripts/container-smoke.sh` if the project chooses to automate the sequence.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - use the public HTTP contract and existing `cmd/server seed` behavior; inspect no production Go source unless a real container-only failure appears.
  * Dependencies: TASK-001, TASK-002, running Docker Compose service, `curl`.

## Step-by-step Instructions

1. Start a clean Compose environment with an empty `./data` directory.
2. Run the seed command inside the container, for example with `docker compose run --rm taskagent seed --user admin --label bootstrap`.
3. Capture the generated raw API key from the seed command output.
4. Start or restart the main Compose service if the seed command was run as a one-off container.
5. Poll `http://localhost:8080/health` until it returns `200`.
6. Send `POST /projects` with `X-API-Key` and a JSON body such as `{"name":"container-smoke"}`.
7. Assert the response status is `201`.
8. Assert the response body contains a non-empty project `id` and the requested `name`.
9. Keep any automation script small and shell-only; do not introduce a new test framework for this task.

## Definition of Done

- The API key can be seeded using the container image.
- The running Compose service accepts the seeded key.
- An authenticated project creation request succeeds through `localhost:8080`.
- Smoke verification can be rerun from a clean `./data` directory.
