# TASK-005: Implement Login Command

## Token Safety Verification

- **Task ID & Title:** TASK-005: Implement Login Command
- **Objective:** Add `task login --server URL --key KEY` credential persistence with secure file permissions.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/config.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this touches one command and config persistence helpers.
  * Dependencies: Config conventions from TASK-004; Click entrypoint from TASK-001.

## Step-by-step Instructions

1. Add a `write_config(config)` helper that creates `$TASKAGENT_HOME` with mode `0700`.
2. Write `$TASKAGENT_HOME/config.json` with mode `0600`.
3. Persist only server and API key values needed by authenticated commands.
4. Implement `task login --server URL --key KEY`.
5. Support `task login --server URL --key -` by reading the API key from stdin.
6. Require both server and key after applying config/env fallbacks where appropriate.
7. Print a short success message to stderr, not stdout.
8. Keep stdout empty on success.
9. Use Click options and errors instead of manual argument parsing.

## Definition of Done

- `task login --server http://localhost:8080 --key test` creates `config.json`.
- The config directory has mode `0700` where the platform supports POSIX modes.
- The config file has mode `0600` where the platform supports POSIX modes.
- `task login --server http://localhost:8080 --key -` reads the key from stdin.
- Successful login does not print secrets.
