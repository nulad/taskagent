# TASK-023: Add Project Human Format Helpers

## Token Safety Verification

- **Task ID & Title:** TASK-023: Add Project Human Format Helpers
- **Objective:** Add project-oriented human rendering helpers that read JSON from stdin.
- **Target Files/Scope:** `cli/lib/format.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this creates one rendering helper module.
  * Dependencies: `jq`; optional `column` command.

## Step-by-step Instructions

1. Add `cli/lib/format.sh`.
2. Implement color detection helpers that enable color only when `[ -t 1 ]` is true and `NO_COLOR` is unset.
3. Implement `render_project_table()` that reads a project JSON array from stdin.
4. Include project ID truncated to 8 chars, name, description, and timestamps where available.
5. Use `jq -r` to produce tab-separated rows.
6. Use `column -t -s "$(printf '\t')"` when `column` is available; otherwise print tab-separated output.
7. Implement a `render_project_detail()` helper for a single project object, preserving full UUID and full RFC3339 timestamps.
8. Do not wire command flags in this task; TASK-025 owns integration.

## Definition of Done

- `render_project_table` accepts `[]` and produces a stable empty or header-only output.
- Table view truncates IDs to 8 characters.
- Detail view preserves full IDs and timestamps.

