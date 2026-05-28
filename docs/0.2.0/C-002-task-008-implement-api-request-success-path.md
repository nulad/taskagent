# TASK-008: Implement API Client Success Path

## Token Safety Verification

- **Task ID & Title:** TASK-008: Implement API Client Success Path
- **Objective:** Add the successful 2xx path for the shared `TaskAgentClient`.
- **Target Files/Scope:** `cli/src/taskagent_cli/api.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this adds the transport module used by all commands.
  * Dependencies: Config model from TASK-004; `httpx` dependency from TASK-001.

## Step-by-step Instructions

1. Add `TaskAgentClient` to `api.py`.
2. Accept base URL, API key, timeout, and optional injected `httpx.Client`.
3. Add a `request(method, path, json=None, params=None)` method.
4. Build URLs from the configured base URL and path without duplicate slashes.
5. Send `X-API-Key` and `Accept: application/json` headers.
6. Send `Content-Type: application/json` when a JSON body exists.
7. Respect the configured timeout.
8. For 2xx responses with a body, return parsed JSON.
9. For `204` or empty bodies, return `None`.
10. Add convenience methods only if they stay thin wrappers around `request`.
11. Leave non-2xx mapping to TASK-009 except for a clear placeholder failure.

## Definition of Done

- Successful JSON responses return Python data.
- Empty `204` responses return `None`.
- Headers include API key and accept.
- Timeout is passed to `httpx`.
- Commands do not call `httpx` directly once the client exists.
