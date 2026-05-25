# TASK-026: Wire Task Human Format Flags

## Token Safety Verification

- **Task ID & Title:** TASK-026: Wire Task Human Format Flags
- **Objective:** Connect `--format human` to task list and show commands while keeping JSON as the default.
- **Target Files/Scope:** `cli/lib/cmd_tasks.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this touches only the task command module, which should remain under the 1,000 LOC limit.
  * Dependencies: Task format helpers from TASK-024; task commands from TASK-016 through TASK-019.

## Step-by-step Instructions

1. Source `cli/lib/format.sh` from `cmd_tasks.sh`.
2. Add optional `--format json|human` parsing to `task list` and `task show`.
3. Keep JSON as the default.
4. Reject unsupported format values with exit 2.
5. For `task list --format human`, pipe the JSON API response to `render_task_table`.
6. For `task show ID --format human`, pipe the JSON API response to `render_task_detail`.
7. Avoid changing add, update, delete, and move output in this task.
8. Preserve machine-readable stdout and human errors on stderr.

## Definition of Done

- `task list` still prints JSON.
- `task list --format human` prints a task table.
- `task show ID --format human` prints task detail.
- Invalid format values exit 2.

