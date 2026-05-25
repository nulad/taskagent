# TASK-029: Document Completion Installation

## Token Safety Verification

- **Task ID & Title:** TASK-029: Document Completion Installation
- **Objective:** Document how users install Bash and Zsh completion scripts for the CLI.
- **Target Files/Scope:** `cli/README.md`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this creates or updates one small README.
  * Dependencies: Completion command from TASK-028.

## Step-by-step Instructions

1. Create `cli/README.md` if it does not exist.
2. Add a short CLI overview and installation notes only as needed for completion documentation.
3. Document Bash installation with `task completion bash > ~/.local/share/bash-completion/completions/task`.
4. Document Zsh installation with a standard completion directory such as `~/.zsh/completions/_task` and the necessary `fpath` note.
5. Mention that completion is static in v0.2.0 if dynamic task/project completion was skipped.
6. Keep the README concise and focused on commands users can run.

## Definition of Done

- `cli/README.md` documents Bash completion installation.
- `cli/README.md` documents Zsh completion installation.
- Known limitations are explicit.
