# TASK-032: Add CLI GitHub Actions Workflow

## Token Safety Verification

- **Task ID & Title:** TASK-032: Add CLI GitHub Actions Workflow
- **Objective:** Add CI coverage for CLI ShellCheck and smoke testing on pushes that affect CLI or server code.
- **Target Files/Scope:** `.github/workflows/cli.yml`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this creates one workflow file and can treat make targets as black-box commands.
  * Dependencies: `make -C cli check` from TASK-003; smoke test from TASK-031.

## Step-by-step Instructions

1. Create `.github/workflows/cli.yml`.
2. Trigger on pull requests and pushes that modify `cli/**`, `cmd/**`, `internal/**`, `migrations/**`, `go.mod`, `go.sum`, `Makefile`, or the workflow itself.
3. Check out the repository.
4. Set up Go using the version from `go.mod`.
5. Install ShellCheck.
6. Install `jq` and `curl` if the runner image does not already provide them.
7. Run `make -C cli check`.
8. Run `sh cli/test/smoke.sh`.
9. Keep the workflow focused on CLI validation; do not fold unrelated Docker or dashboard checks into this file.

## Definition of Done

- The workflow runs ShellCheck for CLI shell files.
- The workflow runs the CLI smoke test.
- The path filters include both CLI files and server files that can break the smoke test.
