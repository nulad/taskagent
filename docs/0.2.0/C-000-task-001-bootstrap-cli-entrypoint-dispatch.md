# TASK-001: Bootstrap Python Click Entrypoint

## Token Safety Verification

- **Task ID & Title:** TASK-001: Bootstrap Python Click Entrypoint
- **Objective:** Create the Python package skeleton and Click-based `task` console entrypoint.
- **Target Files/Scope:** `cli/pyproject.toml`, `cli/src/taskagent_cli/__init__.py`, `cli/src/taskagent_cli/__main__.py`, `cli/src/taskagent_cli/cli.py`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - the Python CLI tree is new and the agent only needs this plan plus the command list in `cli-tasks.md`.
  * Dependencies: CLI version value from `docs/0.2.0/cli-tasks.md`; command names from the CLI plan.

## Step-by-step Instructions

1. Create the `cli/` Python package layout under `cli/src/taskagent_cli/`.
2. Add `cli/pyproject.toml` with project metadata, Python version support, dependencies on `click` and `httpx`, and a `[project.scripts] task = "taskagent_cli.cli:main"` entry.
3. Define `__version__ = "0.2.0"` in `taskagent_cli/__init__.py`.
4. Add `taskagent_cli/__main__.py` that calls `main()` so `python -m taskagent_cli` works.
5. In `taskagent_cli/cli.py`, create a Click top-level group with `--version`.
6. Add placeholder command groups/commands for `login`, `whoami`, `projects`, `add`, `list`, `show`, `update`, `delete`, `move`, and `completion`.
7. Make placeholders fail with Click usage errors or `not implemented` until later tasks fill them in.
8. Configure the top-level command so no subcommand exits 2 after printing help.
9. Keep stdout reserved for command data; help and usage may use Click defaults.

## Definition of Done

- `python -m taskagent_cli --help` exits 0 and prints usage.
- An editable install exposes `task --help`.
- `task --version` exits 0 and prints `0.2.0`.
- Running `task` with no subcommand exits 2.
- The package can be imported without side effects.
