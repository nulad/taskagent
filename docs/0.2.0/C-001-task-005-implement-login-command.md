# TASK-005: Implement Login Command

## Token Safety Verification

- **Task ID & Title:** TASK-005: Implement Login Command
- **Objective:** Add `task login --server URL --key KEY` credential persistence with secure file permissions.
- **Target Files/Scope:** `cli/lib/cmd_login.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the agent only needs the dispatch file and one new command module.
  * Dependencies: `load_config` conventions from TASK-004; `die` helper from TASK-002.

## Step-by-step Instructions

1. Add `cli/lib/cmd_login.sh`.
2. Implement `cmd_login()` with long-option parsing for `--server` and `--key`.
3. Support `--key -` by reading the API key from stdin.
4. Require both server URL and key after parsing, allowing the server to come from existing env/config only if this is already supported cleanly by `load_config`.
5. Create `$TASKAGENT_HOME` or `~/.taskagent` with mode `0700`.
6. Write the config file with only `TASKAGENT_SERVER=...` and `TASKAGENT_API_KEY=...` lines.
7. Set the config file mode to `0600`.
8. Print a short success message to stderr, not stdout.
9. Wire `login` dispatch in `cli/task` to source `cmd_login.sh` and call `cmd_login "$@"`.
10. Keep all implementation POSIX `sh` compatible.

## Definition of Done

- `./cli/task login --server http://localhost:8080 --key test` creates the config file.
- The config directory has mode `0700`.
- The config file has mode `0600`.
- `./cli/task login --server http://localhost:8080 --key -` reads the key from stdin.

