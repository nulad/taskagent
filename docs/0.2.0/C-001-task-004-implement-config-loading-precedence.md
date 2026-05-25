# TASK-004: Implement Config Loading Precedence

## Token Safety Verification

- **Task ID & Title:** TASK-004: Implement Config Loading Precedence
- **Objective:** Implement config loading in `common.sh` with file, environment, and explicit override precedence.
- **Target Files/Scope:** `cli/lib/common.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - `common.sh` should still be a small helper module under 300 LOC.
  * Dependencies: Helpers from TASK-002; config contract from `docs/0.2.0/cli-tasks.md`.

## Step-by-step Instructions

1. Add `default_taskagent_home()` that returns `${TASKAGENT_HOME:-$HOME/.taskagent}`.
2. Add `load_config()` that reads `${TASKAGENT_HOME:-$HOME/.taskagent}/config` if it exists.
3. Source the config file with POSIX dot syntax: `. "$config_file"`.
4. Support `KEY=VALUE` config lines for `TASKAGENT_SERVER`, `TASKAGENT_API_KEY`, and `TASKAGENT_TIMEOUT`.
5. Preserve environment variable overrides over config file values.
6. Apply default values after file and env handling, including `TASKAGENT_TIMEOUT=10`.
7. Add helper functions or variables that allow later command-specific CLI flags to override loaded values without reparsing the config file.
8. Add `require_auth_config()` that errors clearly when `TASKAGENT_SERVER` or `TASKAGENT_API_KEY` is missing and points users to `task login`.
9. Do not implement `task login` in this task.

## Definition of Done

- Authenticated commands can call `load_config` followed by `require_auth_config`.
- File values are loaded when env vars are absent.
- Env vars override config file values.
- Defaults are only applied after file and env values are considered.

