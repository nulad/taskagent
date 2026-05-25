# TASK-008: Implement API Request Success Path

## Token Safety Verification

- **Task ID & Title:** TASK-008: Implement API Request Success Path
- **Objective:** Add the successful 2xx path for the shared `api_request METHOD PATH [BODY]` helper.
- **Target Files/Scope:** `cli/lib/common.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this extends the common helper module, but it should remain well below 500 LOC.
  * Dependencies: `load_config` and `require_auth_config` from TASK-004; `curl` and `jq`.

## Step-by-step Instructions

1. Add `api_request()` to `cli/lib/common.sh`.
2. Accept method and path as required arguments.
3. Accept an optional JSON body argument; if omitted, send no request body.
4. Load and validate config before the request.
5. Build the URL from `TASKAGENT_SERVER` plus the path argument, avoiding duplicate slashes.
6. Send `X-API-Key`, `Content-Type: application/json`, and `Accept: application/json` headers.
7. Respect `TASKAGENT_TIMEOUT`, defaulting to 10 seconds.
8. Capture response body and HTTP status separately using `curl --write-out` or an equivalent robust approach.
9. For 2xx responses, print only the response body to stdout and return success.
10. Do not implement non-2xx error mapping in this task beyond returning a placeholder failure; TASK-009 owns that behavior.

## Definition of Done

- A successful API response prints only JSON body to stdout.
- Headers include API key, content type, and accept.
- `TASKAGENT_TIMEOUT` is passed to curl.
- No command module calls `curl` directly for authenticated API requests.

