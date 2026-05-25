# TASK-001: Bootstrap CLI Entrypoint Dispatch

## Token Safety Verification

- **Task ID & Title:** TASK-001: Bootstrap CLI Entrypoint Dispatch
- **Objective:** Create the executable `cli/task` shell entrypoint with usage, version output, and subcommand dispatch stubs.
- **Target Files/Scope:** `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the `cli/` tree does not exist yet, so the agent only needs this plan and the original CLI command list.
  * Dependencies: CLI version value from `docs/0.2.0/cli-tasks.md`; command names from the CLI plan.

## Step-by-step Instructions

1. Create the `cli/` directory if it does not exist.
2. Add `cli/task` with `#!/usr/bin/env sh` and `set -eu` at the top.
3. Define a version constant for `0.2.0`.
4. Implement a `usage()` function that lists the planned subcommands: `help`, `login`, `whoami`, `projects`, `add`, `list`, `show`, `update`, `delete`, `move`, and `completion`.
5. Make running `task` with no arguments print usage to stderr and exit 2.
6. Make `task --help` and `task help` print usage to stdout and exit 0.
7. Make `task --version` print only the version string and exit 0.
8. Add a `case` dispatch that routes known command names to command functions that will be provided by later tasks.
9. For command modules that do not exist yet, fail with a clear `not implemented: <command>` message and exit 2.
10. Make unknown subcommands print `unknown command: X` to stderr and exit 2.
11. Mark `cli/task` executable.

## Definition of Done

- `./cli/task --help` exits 0 and prints usage.
- `./cli/task --version` exits 0 and prints `0.2.0`.
- `./cli/task` exits 2.
- `./cli/task nope` exits 2 and writes `unknown command: nope` to stderr.

