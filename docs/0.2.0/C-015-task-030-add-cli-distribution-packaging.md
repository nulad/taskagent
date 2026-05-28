# TASK-030: Add Python Distribution Packaging

## Token Safety Verification

- **Task ID & Title:** TASK-030: Add Python Distribution Packaging
- **Objective:** Add Python wheel/source distribution packaging and local install documentation.
- **Target Files/Scope:** `cli/Makefile`, `cli/pyproject.toml`, `cli/README.md`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this updates package metadata, Makefile targets, and README notes.
  * Dependencies: CLI package layout from implementation tasks; version constant from `taskagent_cli.__version__`.

## Step-by-step Instructions

1. Ensure `cli/pyproject.toml` includes build-system metadata.
2. Ensure package discovery includes `src/taskagent_cli`.
3. Extend `cli/Makefile` with a `build` target that runs `python -m build`.
4. Produce wheel and source distributions under `cli/dist/`.
5. Document `pipx install ./cli/dist/*.whl` as the local install path.
6. Document editable development install with the chosen package manager.
7. Verify the installed wheel exposes the `task` console script.
8. Avoid requiring sudo or global site-packages installs.

## Definition of Done

- `make -C cli build` creates wheel and sdist artifacts under `cli/dist/`.
- The wheel contains the Python package and console script metadata.
- `pipx install ./cli/dist/*.whl` installs a working `task` command.
