# TASK-024: Add Task Human Format Helpers

## Token Safety Verification

- **Task ID & Title:** TASK-024: Add Task Human Format Helpers
- **Objective:** Add task-oriented human rendering helpers with status color support.
- **Target Files/Scope:** `cli/src/taskagent_cli/format.py`, `cli/tests/test_format.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends the formatting module.
  * Dependencies: Project format helper scaffolding from TASK-023; status constants from TASK-002.

## Step-by-step Instructions

1. Extend `format.py`.
2. Implement `render_task_table(tasks)` for a list of task objects.
3. Include ID truncated to 8 chars, status, title, project ID or project name when present, and shortened updated timestamp.
4. Implement `render_task_detail(task)` for a single task object with full UUIDs and full RFC3339 timestamps.
5. Colorize the status column only when color is enabled.
6. Use status colors: backlog=gray, todo=yellow, in-progress=blue, review=magenta, done=green.
7. Ensure redirected output contains no ANSI escape sequences.
8. Handle empty tag arrays and missing optional fields without errors.
9. Do not wire command flags in this task; TASK-026 owns integration.

## Definition of Done

- Task table rendering works for empty and non-empty arrays.
- Status colors appear only on TTY output when `NO_COLOR` is unset.
- Detail rendering keeps full IDs and timestamps.
- Unit tests cover color and no-color paths.
