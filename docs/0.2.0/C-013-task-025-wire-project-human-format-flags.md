# TASK-025: Wire Project Human Format Flags

## Token Safety Verification

- **Task ID & Title:** TASK-025: Wire Project Human Format Flags
- **Objective:** Connect `--format human` to project list and show commands while keeping JSON as the default.
- **Target Files/Scope:** `cli/lib/cmd_projects.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this touches only the project command module.
  * Dependencies: Project format helpers from TASK-023; project commands from TASK-011 through TASK-014.

## Step-by-step Instructions

1. Source `cli/lib/format.sh` from `cmd_projects.sh`.
2. Add optional `--format json|human` parsing to `projects list` and `projects show`.
3. Keep JSON as the default.
4. Reject unsupported format values with exit 2.
5. For `projects list --format human`, pipe the JSON API response to `render_project_table`.
6. For `projects show ID --format human`, pipe the JSON API response to `render_project_detail`.
7. Avoid changing create, update, and delete output in this task.
8. Preserve machine-readable stdout and human errors on stderr.

## Definition of Done

- `task projects list` still prints JSON.
- `task projects list --format human` prints a human table.
- `task projects show ID --format human` prints project detail.
- Invalid format values exit 2.

