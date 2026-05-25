# TASK-030: Add CLI Distribution Packaging

## Token Safety Verification

- **Task ID & Title:** TASK-030: Add CLI Distribution Packaging
- **Objective:** Add `make -C cli dist` packaging and an idempotent CLI installer script.
- **Target Files/Scope:** `cli/Makefile`, `cli/install.sh`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this updates one small Makefile and adds one installer script.
  * Dependencies: CLI file layout from implementation tasks; version constant from `cli/task`.

## Step-by-step Instructions

1. Extend `cli/Makefile` with a `dist` target.
2. Produce `dist/task-0.2.0.tar.gz` or derive the version from `cli/task` if practical.
3. Include `task`, `lib/*.sh`, `install.sh`, and CLI README files in the tarball.
4. Add `cli/install.sh` with POSIX `sh` compatibility.
5. Make the installer copy `task` to `~/.local/bin/task`.
6. Make the installer copy `lib/*.sh` to `~/.local/share/taskagent/lib`.
7. Create target directories if missing.
8. Make installation idempotent: re-running updates existing files without duplication.
9. Preserve executable mode for the installed `task` script.
10. Avoid requiring sudo.

## Definition of Done

- `make -C cli dist` creates a tarball under `cli/dist/`.
- The tarball contains the executable, library files, installer, and README.
- Running `cli/install.sh` twice succeeds and leaves one updated install.
