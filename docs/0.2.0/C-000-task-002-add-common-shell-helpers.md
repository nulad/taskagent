# TASK-002: Add Common Shell Helpers

## Token Safety Verification

- **Task ID & Title:** TASK-002: Add Common Shell Helpers
- **Objective:** Add shared shell helper functions used by later CLI modules.
- **Target Files/Scope:** `cli/lib/common.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is one new helper file expected to stay below 200 LOC.
  * Dependencies: POSIX `sh`; required external commands `curl` and `jq`.

## Step-by-step Instructions

1. Create `cli/lib/` if it does not exist.
2. Add `cli/lib/common.sh` with `#!/usr/bin/env sh` style-compatible content, but design it to be sourced rather than executed.
3. Add `die()` that prints its message to stderr and exits with an optional exit code, defaulting to 1.
4. Add `log()` that prints human messages to stderr.
5. Add `require_cmd()` that accepts one or more command names and fails fast with `missing required command: <name>` if any is absent.
6. Add a best-effort `enable_pipefail()` helper that enables `set -o pipefail` only where the shell supports it.
7. Do not add HTTP or config logic in this task.
8. Keep all helpers POSIX-compatible; do not use Bash arrays, `[[ ... ]]`, or `source`.

## Definition of Done

- `cli/lib/common.sh` can be sourced by `sh` without executing CLI behavior.
- `require_cmd curl jq` succeeds when both are installed.
- `require_cmd definitely_missing_command` exits non-zero with a clear stderr message.

