# TASK-004: Implement Config Loading Precedence

## Token Safety Verification

- **Task ID & Title:** TASK-004: Implement Config Loading Precedence
- **Objective:** Implement JSON config loading with file, environment, and explicit override precedence.
- **Target Files/Scope:** `cli/src/taskagent_cli/config.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is one focused Python module.
  * Dependencies: Shared error constants from TASK-002; config contract from `docs/0.2.0/cli-tasks.md`.

## Step-by-step Instructions

1. Add a `Config` dataclass with `server`, `api_key`, `timeout`, and `home` fields.
2. Add `default_taskagent_home()` returning `$TASKAGENT_HOME` or `~/.taskagent`.
3. Load config from `$TASKAGENT_HOME/config.json` by default.
4. Support JSON keys for `server`, `api_key`, and `timeout`.
5. Apply precedence in this order: explicit overrides, environment variables, config file, defaults.
6. Support environment variables `TASKAGENT_SERVER`, `TASKAGENT_API_KEY`, and `TASKAGENT_TIMEOUT`.
7. Default server to `http://localhost:8080` and timeout to 10 seconds.
8. Add `require_auth_config(config)` that raises a config/usage error when server or API key is missing and points users to `task login`.
9. Keep file writing for `task login` out of this task.

## Definition of Done

- File values load when env vars are absent.
- Env vars override config file values.
- Explicit overrides override env vars.
- Defaults are applied only after file and env values are considered.
- Missing auth config produces `task login` guidance.
