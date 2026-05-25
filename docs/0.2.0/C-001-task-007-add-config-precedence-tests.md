# TASK-007: Add Config Precedence Tests

## Token Safety Verification

- **Task ID & Title:** TASK-007: Add Config Precedence Tests
- **Objective:** Add a runnable shell test that verifies CLI config precedence and login file permissions.
- **Target Files/Scope:** `cli/test/test_config.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - one focused shell test file can source the small CLI helper modules.
  * Dependencies: Config loading from TASK-004; login command from TASK-005.

## Step-by-step Instructions

1. Create `cli/test/` if it does not exist.
2. Add `cli/test/test_config.sh` with `#!/usr/bin/env sh` and `set -eu`.
3. Use a temporary `TASKAGENT_HOME` for all test cases.
4. Verify config file values are loaded when no env vars are set.
5. Verify env vars override config file values.
6. Verify explicit override hooks intended for CLI flags override env values.
7. Verify missing `TASKAGENT_SERVER` or `TASKAGENT_API_KEY` fails with guidance that mentions `task login`.
8. Verify `task login` writes a `0700` config directory and a `0600` config file.
9. Clean up the temporary directory on exit.
10. Keep the test runnable directly with `sh cli/test/test_config.sh`.

## Definition of Done

- `sh cli/test/test_config.sh` passes on a clean checkout after TASK-004 and TASK-005.
- The test does not read or modify the user's real `~/.taskagent`.
- The test covers `flag > env > file > default` behavior.

