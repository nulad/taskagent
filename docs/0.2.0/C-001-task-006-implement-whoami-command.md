# TASK-006: Implement Whoami Command

## Token Safety Verification

- **Task ID & Title:** TASK-006: Implement Whoami Command
- **Objective:** Add `task whoami` to validate the configured API key and print key metadata.
- **Target Files/Scope:** `cli/lib/cmd_auth.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only adds a small auth command module and one dispatch branch.
  * Dependencies: `api_request` from TASK-008; config loading from TASK-004.

## Step-by-step Instructions

1. Add `cli/lib/cmd_auth.sh`.
2. Implement `cmd_whoami()` with no required arguments.
3. Load config and require authenticated config before making the request.
4. Call `api_request GET /auth/keys`.
5. Print the raw JSON response by default.
6. If the command supports `--format human` after TASK-025, defer that integration to TASK-025; do not implement table rendering here.
7. Wire `whoami` dispatch in `cli/task` to source `cmd_auth.sh` and call `cmd_whoami "$@"`.
8. Reject unknown whoami flags with usage error exit 2.

## Definition of Done

- `./cli/task whoami` sends `GET /auth/keys` through the shared request helper.
- Missing config produces the shared `task login` guidance.
- Successful responses are printed to stdout as JSON.

