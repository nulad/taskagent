# TASK-026: Wire Task Human Format Flags

## Token Safety Verification

- **Task ID & Title:** TASK-026: Wire Task Human Format Flags
- **Objective:** Connect `--format human` to task list and show commands while keeping JSON as the default.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/tests/test_commands.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Medium - this touches task command options and tests.
  * Dependencies: Task format helpers from TASK-024; task commands from TASK-016 through TASK-019.

## Step-by-step Instructions

1. Add `--format json|human` to `task list` and `task show`.
2. Keep JSON as the default.
3. Reject unsupported format values through Click.
4. For `task list --format human`, pass response data to `render_task_table`.
5. For `task show ID --format human`, pass response data to `render_task_detail`.
6. Avoid changing add, update, delete, and move output in this task.
7. Preserve machine-readable stdout and human errors on stderr.

## Definition of Done

- `task list` still prints JSON.
- `task list --format human` prints a task table.
- `task show ID --format human` prints task detail.
- Invalid format values exit 2.
