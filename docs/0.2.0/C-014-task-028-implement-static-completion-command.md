# TASK-028: Implement Click Completion Command

## Token Safety Verification

- **Task ID & Title:** TASK-028: Implement Click Completion Command
- **Objective:** Add `task completion bash|zsh|fish` helpers around Click's built-in shell completion support.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one command and status completion metadata.
  * Dependencies: Final command and flag names from previous tasks; known task status values.

## Step-by-step Instructions

1. Add `task completion SHELL` where `SHELL` is one of `bash`, `zsh`, or `fish`.
2. Use Click's documented completion environment variable mechanism for the installed `task` command.
3. Print a sourceable completion script or exact setup command for the requested shell.
4. Include static completion for top-level commands and long flags through Click metadata.
5. Add shell completion for task status values on `task move`.
6. Do not implement dynamic task ID or project name completion unless it fits without extra API calls.
7. Unsupported shell names must exit 2.

## Definition of Done

- `task completion bash` prints Bash completion setup output.
- `task completion zsh` prints Zsh completion setup output.
- `task completion fish` prints Fish completion setup output.
- Completion includes commands, flags, and status values.
- Unsupported shells exit 2.
