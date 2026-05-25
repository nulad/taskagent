# TASK-027: Wire Auth Human Format Flag

## Token Safety Verification

- **Task ID & Title:** TASK-027: Wire Auth Human Format Flag
- **Objective:** Connect `--format human` to `task whoami` so it prints key label and user details without changing the default JSON output.
- **Target Files/Scope:** `cli/lib/cmd_auth.sh`, `cli/lib/format.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this touches one small auth command module and one small format helper module.
  * Dependencies: `whoami` command from TASK-006; formatting conventions from TASK-023 and TASK-024.

## Step-by-step Instructions

1. Add a small `render_auth_key_detail()` helper to `cli/lib/format.sh`.
2. Make the helper read the `/auth/keys` JSON response from stdin and print key label plus user information in a concise human format.
3. Preserve full IDs and timestamps in the human auth output.
4. Source `cli/lib/format.sh` from `cmd_auth.sh`.
5. Add optional `--format json|human` parsing to `task whoami`.
6. Keep JSON as the default.
7. Reject unsupported format values with exit 2.
8. Pipe successful `whoami` JSON to `render_auth_key_detail` only when `--format human` is requested.

## Definition of Done

- `task whoami` still prints JSON.
- `task whoami --format human` prints key label and user details.
- Invalid format values exit 2.
