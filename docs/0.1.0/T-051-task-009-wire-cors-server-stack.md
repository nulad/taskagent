# TASK-009: Wire CORS Into Server Stack

## Token Safety Verification

- **Task ID & Title:** TASK-009: Wire CORS Into Server Stack
- **Objective:** Apply CORS middleware to the server’s final HTTP stack using configured origins.
- **Target Files/Scope:** `cmd/server/main.go`; update `internal/handler/integration_test.go` only if its test harness needs the same wrapper.
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - `cmd/server/main.go` is about 200 LOC; the integration harness is larger but only the server construction helper should need attention.
  * Dependencies: TASK-007 config field, TASK-008 CORS middleware.

## Step-by-step Instructions

1. Wrap the server handler with `middleware.CORSMiddleware(cfg.CORSOrigins)`.
2. Place CORS outside auth so browser preflight `OPTIONS` requests do not require `X-API-Key`.
3. Ensure `/health` and protected API routes both pass through the CORS layer.
4. If the integration test harness builds an equivalent full server stack, update only the harness construction to include the CORS wrapper.
5. Do not change route registration or business handler behavior.

## Definition of Done

- The production server applies CORS to all routes.
- Preflight requests can be answered before auth.
- Existing server startup and graceful shutdown behavior are unchanged.
