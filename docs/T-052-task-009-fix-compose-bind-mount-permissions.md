# TASK-009: Fix Compose Bind-Mount Permission Failure

## Token Safety Verification

- **Task ID & Title:** TASK-009: Fix Compose Bind-Mount Permission Failure
- **Objective:** Ensure `docker compose up --build` starts successfully when using the required `./data:/data` bind mount, without requiring manual ad-hoc permission repair.
- **Target Files/Scope:** `Makefile`, `docker-compose.yml`, `docs/deployment.md` (and optional tiny helper script if needed).
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task should only touch deployment artifacts and command documentation.
  * Dependencies: Existing TASK-001 Dockerfile non-root runtime, existing TASK-002 bind mount requirement, existing `/health` route.

## Step-by-step Instructions

1. Reproduce the failure from a clean state where `./data` is missing or not writable by the container runtime user.
2. Keep the TASK-002 requirement to bind-mount `./data` to `/data`; do not replace it with a named volume.
3. Add a deterministic preflight step (preferably in `Makefile`) that prepares `./data` before Compose startup.
4. The preflight must make `./data` writable by the process that runs inside the container while preserving TASK-001 non-root runtime behavior.
5. Wire the preflight into the normal Compose-up workflow so users do not need to run manual one-off `chown` commands.
6. Preserve existing Compose contract details (single service `taskagent`, `taskagent:local` image, required env vars, port mapping, healthcheck).
7. Update deployment documentation so first-time setup and troubleshooting clearly explain why `./data` permissions matter and how the workflow handles it.
8. Validate with `docker compose config` and `docker compose up --build` on a clean local state.

## Definition of Done

- `docker compose config` succeeds.
- `docker compose up --build` starts exactly one `taskagent` service without `unable to open database file` errors.
- SQLite file creation succeeds under local `./data` via bind mount.
- The container still runs with non-root execution semantics from TASK-001.
- Updated docs and/or Make targets document the final startup flow.
