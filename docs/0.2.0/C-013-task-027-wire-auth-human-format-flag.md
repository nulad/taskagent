# TASK-027: Wire Auth Human Format Flag

## Token Safety Verification

- **Task ID & Title:** TASK-027: Wire Auth Human Format Flag
- **Objective:** Connect `--format human` to `task whoami` so it prints key label and user details without changing the default JSON output.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`, `cli/src/taskagent_cli/format.py`, `cli/tests/test_commands.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this touches one auth command and one small format helper.
  * Dependencies: `whoami` command from TASK-006; formatting conventions from TASK-023 and TASK-024.

## Step-by-step Instructions

1. Add `render_auth_key_detail(data)` to `format.py`.
2. Make the helper print key label plus user information in a concise human format.
3. Preserve full IDs and timestamps in human auth output.
4. Add optional `--format json|human` to `task whoami`.
5. Keep JSON as the default.
6. Reject unsupported format values through Click.
7. Pass successful `whoami` response data to `render_auth_key_detail` only when `--format human` is requested.

## Definition of Done

- `task whoami` still prints JSON.
- `task whoami --format human` prints key label and user details.
- Invalid format values exit 2.
