# TASK-003: Add CLI Quality Targets

## Token Safety Verification

- **Task ID & Title:** TASK-003: Add CLI Quality Targets
- **Objective:** Add Python CLI quality tooling and `make -C cli` targets for lint, format check, typecheck, and tests.
- **Target Files/Scope:** `cli/Makefile`, `cli/pyproject.toml`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only updates CLI package config and a small Makefile.
  * Dependencies: Python package skeleton from TASK-001.

## Step-by-step Instructions

1. Add dev dependencies for `pytest`, `pytest-httpx` or `respx`, `ruff`, `mypy`, and `build`.
2. Configure Ruff in `pyproject.toml` with a conservative rule set.
3. Configure mypy for the `taskagent_cli` package.
4. Add `cli/Makefile`.
5. Define `lint`, `format-check`, `typecheck`, `test`, `check`, `build`, and `smoke` targets.
6. Make `check` run lint, format check, typecheck, and unit tests.
7. Avoid changing the root `Makefile` in this task.

## Definition of Done

- `make -C cli check` is defined.
- `make -C cli test` runs pytest.
- `make -C cli build` produces Python distribution artifacts once package code exists.
- Missing tooling fails through normal Python packaging dependency installation, not ad hoc checks.
