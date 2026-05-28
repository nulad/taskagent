# TASK-007: Add Config Precedence Tests

## Token Safety Verification

- **Task ID & Title:** TASK-007: Add Config Precedence Tests
- **Objective:** Add pytest coverage for config precedence, login persistence, and safe test isolation.
- **Target Files/Scope:** `cli/tests/test_config.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - one focused pytest file can cover the config module and login command.
  * Dependencies: Config loading from TASK-004; login command from TASK-005.

## Step-by-step Instructions

1. Create `cli/tests/` if it does not exist.
2. Add `cli/tests/test_config.py`.
3. Use `tmp_path` and monkeypatched `TASKAGENT_HOME` for all test cases.
4. Verify config file values are loaded when no env vars are set.
5. Verify env vars override config file values.
6. Verify explicit overrides intended for CLI flags override env values.
7. Verify missing API key fails with guidance that mentions `task login`.
8. Use Click's `CliRunner` to verify `task login` writes config.
9. Verify directory and file permission modes where POSIX mode checks are reliable.
10. Ensure tests never read or modify the user's real `~/.taskagent`.

## Definition of Done

- `make -C cli test` runs the config tests.
- Tests cover `flag > env > file > default` behavior.
- Tests cover `task login --key -`.
- Tests are isolated from real user config.
