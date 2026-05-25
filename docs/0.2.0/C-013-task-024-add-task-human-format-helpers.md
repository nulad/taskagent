# TASK-024: Add Task Human Format Helpers

## Token Safety Verification

- **Task ID & Title:** TASK-024: Add Task Human Format Helpers
- **Objective:** Add task-oriented human rendering helpers with status color support.
- **Target Files/Scope:** `cli/lib/format.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends one formatting helper module that should remain under 400 LOC.
  * Dependencies: Project format helper scaffolding from TASK-023; `jq`.

## Step-by-step Instructions

1. Extend `cli/lib/format.sh`.
2. Implement `render_task_table()` that reads a task JSON array from stdin.
3. Include ID truncated to 8 chars, status, title, project ID or project name when present, and shortened updated timestamp.
4. Implement `render_task_detail()` for a single task object with full UUIDs and full RFC3339 timestamps.
5. Colorize the status column only when color is enabled.
6. Use status colors: backlog=gray, todo=yellow, in-progress=blue, review=magenta, done=green.
7. Ensure redirected output contains no ANSI escape sequences.
8. Handle empty tag arrays and missing optional fields without `jq` errors.
9. Do not wire command flags in this task; TASK-025 owns integration.

## Definition of Done

- Task table rendering works for empty and non-empty arrays.
- Status colors appear only on TTY output when `NO_COLOR` is unset.
- Detail rendering keeps full IDs and timestamps.

