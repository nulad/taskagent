# TASK-012: Implement Project Create Command

## Token Safety Verification

- **Task ID & Title:** TASK-012: Implement Project Create Command
- **Objective:** Add `task projects create --name NAME [--description TEXT]` with safe JSON payload construction.
- **Target Files/Scope:** `cli/src/taskagent_cli/cli.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - only the existing project command group is needed.
  * Dependencies: Project command group from TASK-011; `TaskAgentClient`.

## Step-by-step Instructions

1. Add `task projects create`.
2. Define `--name` as a required Click option.
3. Define optional `--description`.
4. Build a Python dict payload rather than string-concatenating JSON.
5. Include `description` only when the user provides it, unless the server contract explicitly requires an empty string.
6. Call `POST /projects`.
7. Print the created project JSON to stdout.
8. Let Click reject unknown flags and missing required values.

## Definition of Done

- `task projects create --name Demo` posts valid JSON to `/projects`.
- Descriptions with spaces and quotes are encoded correctly.
- Unknown flags and missing `--name` exit 2.
