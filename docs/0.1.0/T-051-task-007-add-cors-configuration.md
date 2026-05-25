# TASK-007: Add CORS Configuration

## Token Safety Verification

- **Task ID & Title:** TASK-007: Add CORS Configuration
- **Objective:** Load allowed CORS origins from `TASKAGENT_CORS_ORIGINS` with an empty default.
- **Target Files/Scope:** `internal/config/config.go`, `internal/config/config_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - config source and tests are small and isolated.
  * Dependencies: Existing `Config` struct and `Load()` environment variable pattern.

## Step-by-step Instructions

1. Add a `CORSOrigins []string` field to `config.Config`.
2. Add an env var constant for `TASKAGENT_CORS_ORIGINS`.
3. Parse the env var as a comma-separated allowlist.
4. Trim surrounding whitespace from each origin.
5. Ignore empty entries.
6. Keep the default as an empty slice or nil slice, meaning no cross-origin access.
7. Add config tests for default empty origins, one origin, multiple origins, and whitespace/empty entry cleanup.

## Definition of Done

- `config.Load()` exposes parsed CORS origins.
- Empty configuration does not allow any origin by default.
- Config tests cover the parsing behavior.
