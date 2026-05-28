# TASK-023: Add Project Human Format Helpers

## Token Safety Verification

- **Task ID & Title:** TASK-023: Add Project Human Format Helpers
- **Objective:** Add project-oriented human rendering helpers that accept Python data.
- **Target Files/Scope:** `cli/src/taskagent_cli/format.py`, `cli/tests/test_format.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this extends one rendering helper module.
  * Dependencies: `print_json()` from TASK-002.

## Step-by-step Instructions

1. Extend `format.py`.
2. Implement color detection helpers that enable color only when stdout is a TTY and `NO_COLOR` is unset.
3. Implement `render_project_table(projects)` for a list of project objects.
4. Include project ID truncated to 8 chars, name, description, and timestamps where available.
5. Implement a `render_project_detail(project)` helper for a single project object.
6. Preserve full UUIDs and full RFC3339 timestamps in detail view.
7. Keep formatting dependency-free unless a small table dependency has already been justified in `pyproject.toml`.
8. Do not wire command flags in this task; TASK-025 owns integration.

## Definition of Done

- `render_project_table([])` produces stable empty or header-only output.
- Table view truncates IDs to 8 characters.
- Detail view preserves full IDs and timestamps.
- Unit tests cover empty and non-empty project rendering.
