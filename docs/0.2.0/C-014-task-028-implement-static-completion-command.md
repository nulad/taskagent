# TASK-028: Implement Static Completion Command

## Token Safety Verification

- **Task ID & Title:** TASK-028: Implement Static Completion Command
- **Objective:** Add `task completion bash` and `task completion zsh` that print sourceable static completion scripts.
- **Target Files/Scope:** `cli/lib/cmd_completion.sh`, `cli/task`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this adds one command module containing static script templates and one dispatch branch.
  * Dependencies: Final command and flag names from previous tasks; known task status values.

## Step-by-step Instructions

1. Add `cli/lib/cmd_completion.sh`.
2. Implement `cmd_completion()` requiring one shell argument: `bash` or `zsh`.
3. For `bash`, print a complete script using `complete -F`.
4. For `zsh`, print a complete script using `compdef`.
5. Include static completion for top-level subcommands.
6. Include relevant long flags for each command group.
7. Complete task statuses for `task move <id> <TAB>` as `backlog`, `todo`, `in-progress`, `review`, and `done`.
8. Do not implement dynamic task ID or project name completion unless it fits without adding API calls.
9. Wire top-level `completion` dispatch in `cli/task`.
10. Unknown shell names exit 2.

## Definition of Done

- `./cli/task completion bash` prints a sourceable Bash completion script.
- `./cli/task completion zsh` prints a sourceable Zsh completion script.
- Completion scripts include subcommands, flags, and status values.
- Unsupported shells exit 2.
