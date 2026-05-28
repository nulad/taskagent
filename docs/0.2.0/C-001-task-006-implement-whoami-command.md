# TASK-006: Implement Whoami Command

## Token Safety Verification

- **Task ID & Title:** TASK-006: Implement Whoami Command
- **Objective:** Add `task whoami` to validate the configured API key and print key metadata.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task adds one Click command.
  * Dependencies: `TaskAgentClient` from TASK-008; config loading from TASK-004.

## Step-by-step Instructions

1. Implement `task whoami` with no required arguments.
2. Load config and require authenticated config before making the request.
3. Create a `TaskAgentClient` using configured server, API key, and timeout.
4. Call `GET /auth/keys`.
5. Print JSON to stdout by default via `print_json()`.
6. Accept `--format json|human` only if the shared format option already exists; otherwise leave human rendering to TASK-027.
7. Reject unknown flags through Click usage errors.

## Definition of Done

- `task whoami` sends `GET /auth/keys` through the shared client.
- Missing config produces the shared `task login` guidance.
- Successful responses are printed to stdout as JSON.
- API failures use the shared error-to-exit-code behavior.
