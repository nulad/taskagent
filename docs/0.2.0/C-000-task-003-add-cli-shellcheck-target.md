# TASK-003: Add CLI Shellcheck Target

## Token Safety Verification

- **Task ID & Title:** TASK-003: Add CLI Shellcheck Target
- **Objective:** Add a `make -C cli check` target that runs ShellCheck against all CLI shell files.
- **Target Files/Scope:** `cli/Makefile`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only creates a small Makefile in the new CLI directory.
  * Dependencies: `shellcheck`; shell files created by TASK-001 and TASK-002.

## Step-by-step Instructions

1. Add `cli/Makefile`.
2. Define a `check` target.
3. Have `check` fail clearly if `shellcheck` is not installed.
4. Run ShellCheck on `task` and every `*.sh` file under `cli/lib/` and `cli/test/` when those directories exist.
5. Use POSIX shell compatibility flags where practical, such as `shellcheck -s sh`.
6. Avoid changing the root `Makefile` in this task.

## Definition of Done

- `make -C cli check` runs without Makefile syntax errors.
- When `shellcheck` is unavailable, the target fails with an actionable message.
- The target includes future `cli/lib/*.sh` and `cli/test/*.sh` files without needing edits.

