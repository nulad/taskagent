# TASK-012: Implement Project Create Command

## Token Safety Verification

- **Task ID & Title:** TASK-012: Implement Project Create Command
- **Objective:** Add `task projects create --name NAME [--description TEXT]` with safe JSON payload construction.
- **Target Files/Scope:** `cli/lib/cmd_projects.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - only the existing project command module is needed.
  * Dependencies: Project command dispatch from TASK-011; `jq`; `api_request`.

## Step-by-step Instructions

1. Extend `cmd_projects.sh` with a `projects_create()` function.
2. Parse `--name` and optional `--description` using a POSIX `while` and `case` loop.
3. Require `--name`; missing name exits 2 with a usage message.
4. Reject unknown flags with exit 2.
5. Build the JSON payload with `jq -n --arg`, never with string concatenation.
6. Include `description` only when the user provides it, unless the server contract explicitly requires an empty string.
7. POST the payload to `/projects` using `api_request`.
8. Print the created project JSON to stdout.

## Definition of Done

- `./cli/task projects create --name Demo` POSTs valid JSON to `/projects`.
- Descriptions with spaces and quotes are encoded correctly.
- Unknown flags and missing `--name` exit 2.

