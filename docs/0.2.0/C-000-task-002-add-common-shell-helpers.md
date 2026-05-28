# TASK-002: Add Python CLI Shared Utilities

## Token Safety Verification

- **Task ID & Title:** TASK-002: Add Python CLI Shared Utilities
- **Objective:** Add shared Python helpers for CLI errors, output, context, and small model utilities used by later commands.
- **Target Files/Scope:** `cli/src/taskagent_cli/errors.py`, `cli/src/taskagent_cli/models.py`, `cli/src/taskagent_cli/format.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - these are small new modules.
  * Dependencies: Package skeleton from TASK-001.

## Step-by-step Instructions

1. Add `errors.py` with exception types for usage/config errors, API client errors, API server errors, and network errors.
2. Centralize exit code constants: 0 success, 2 usage/config, 3 API client error, 4 API server error, 5 network error.
3. Add `models.py` with valid task statuses and helpers for status validation.
4. Add a `parse_tags()` helper that converts comma-separated tags into a trimmed list and drops empty entries.
5. Add a helper for partial update payloads that includes only explicitly provided fields.
6. Add `format.py` with a `print_json()` helper using `json.dumps`.
7. Keep helpers framework-light: no server calls and no Click command definitions in this task.

## Definition of Done

- Shared exit codes are importable from one module.
- Valid task statuses are defined once.
- `parse_tags("bug, urgent,,")` returns `["bug", "urgent"]`.
- `print_json()` writes JSON to stdout and is easy to assert in tests.
