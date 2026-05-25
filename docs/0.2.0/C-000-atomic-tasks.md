# Atomic Tasks Index: C-000 TaskAgent CLI

Each atomic task lives in its own file so an implementation agent can load only the task it is executing. These tasks break down [cli-tasks.md](cli-tasks.md) into small scopes that should fit inside a 32,000-token agent runtime.

## Task Files

1. [TASK-001: Bootstrap CLI Entrypoint Dispatch](C-000-task-001-bootstrap-cli-entrypoint-dispatch.md)
2. [TASK-002: Add Common Shell Helpers](C-000-task-002-add-common-shell-helpers.md)
3. [TASK-003: Add CLI Shellcheck Target](C-000-task-003-add-cli-shellcheck-target.md)
4. [TASK-004: Implement Config Loading Precedence](C-001-task-004-implement-config-loading-precedence.md)
5. [TASK-005: Implement Login Command](C-001-task-005-implement-login-command.md)
6. [TASK-006: Implement Whoami Command](C-001-task-006-implement-whoami-command.md)
7. [TASK-007: Add Config Precedence Tests](C-001-task-007-add-config-precedence-tests.md)
8. [TASK-008: Implement API Request Success Path](C-002-task-008-implement-api-request-success-path.md)
9. [TASK-009: Implement API Request Error Mapping](C-002-task-009-implement-api-request-error-mapping.md)
10. [TASK-010: Add API Request Error Tests](C-002-task-010-add-api-request-error-tests.md)
11. [TASK-011: Implement Project List And Show Commands](C-010-task-011-implement-project-list-show.md)
12. [TASK-012: Implement Project Create Command](C-010-task-012-implement-project-create.md)
13. [TASK-013: Implement Project Update Command](C-010-task-013-implement-project-update.md)
14. [TASK-014: Implement Project Delete Command](C-010-task-014-implement-project-delete.md)
15. [TASK-015: Add Project Command Tests](C-010-task-015-add-project-command-tests.md)
16. [TASK-016: Implement Task Add Show Delete Commands](C-011-task-016-implement-task-add-show-delete.md)
17. [TASK-017: Implement Task List Filters](C-011-task-017-implement-task-list-filters.md)
18. [TASK-018: Implement Task Update Payloads](C-011-task-018-implement-task-update-payloads.md)
19. [TASK-019: Implement Project Name Resolution For Tasks](C-011-task-019-implement-project-name-resolution.md)
20. [TASK-020: Add Task Command Tests](C-011-task-020-add-task-command-tests.md)
21. [TASK-021: Implement Task Move Command](C-012-task-021-implement-task-move-command.md)
22. [TASK-022: Add Task Move Tests](C-012-task-022-add-task-move-tests.md)
23. [TASK-023: Add Project Human Format Helpers](C-013-task-023-add-project-human-format-helpers.md)
24. [TASK-024: Add Task Human Format Helpers](C-013-task-024-add-task-human-format-helpers.md)
25. [TASK-025: Wire Project Human Format Flags](C-013-task-025-wire-project-human-format-flags.md)
26. [TASK-026: Wire Task Human Format Flags](C-013-task-026-wire-task-human-format-flags.md)
27. [TASK-027: Wire Auth Human Format Flag](C-013-task-027-wire-auth-human-format-flag.md)
28. [TASK-028: Implement Static Completion Command](C-014-task-028-implement-static-completion-command.md)
29. [TASK-029: Document Completion Installation](C-014-task-029-document-completion-installation.md)
30. [TASK-030: Add CLI Distribution Packaging](C-015-task-030-add-cli-distribution-packaging.md)
31. [TASK-031: Add CLI Smoke Test](C-015-task-031-add-cli-smoke-test.md)
32. [TASK-032: Add CLI GitHub Actions Workflow](C-015-task-032-add-cli-github-actions-workflow.md)

## Execution Order

1. Bootstrap foundation: TASK-001, TASK-002, TASK-003.
2. Add configuration and auth commands: TASK-004, TASK-005, TASK-006, TASK-007.
3. Add the HTTP transport layer: TASK-008, TASK-009, TASK-010.
4. Add project commands: TASK-011, TASK-012, TASK-013, TASK-014, TASK-015.
5. Add task commands: TASK-016, TASK-017, TASK-018, TASK-019, TASK-020.
6. Add status move: TASK-021, TASK-022.
7. Add human formatting: TASK-023, TASK-024, TASK-025, TASK-026, TASK-027.
8. Add shell completion: TASK-028, TASK-029.
9. Add release packaging and final verification: TASK-030, TASK-031, TASK-032.

## Parallelization

- TASK-001, TASK-002, and TASK-003 are foundational and should be done first. TASK-002 and TASK-003 can run in parallel after the directory layout from TASK-001 is agreed.
- TASK-005 and TASK-006 can run in parallel after TASK-004 and TASK-008 exist, because they touch separate command modules.
- TASK-011 through TASK-014 must be sequential if one agent owns `cli/lib/cmd_projects.sh`; otherwise they can be split only with explicit coordination because they share the same file.
- TASK-016 through TASK-019 must be sequential if one agent owns `cli/lib/cmd_tasks.sh`; otherwise they require explicit merge coordination because they share the same file.
- TASK-015 can run after TASK-011 through TASK-014 are complete. TASK-020 can run after TASK-016 through TASK-019 are complete.
- TASK-023 and TASK-024 can run in parallel because they add separate functions in `cli/lib/format.sh`. TASK-025, TASK-026, and TASK-027 can then run in parallel because they wire separate command modules.
- TASK-028 and TASK-030 can run in parallel after the command surface is stable.
- TASK-031 must wait for the command implementation and packaging entrypoints. TASK-032 should run last because it wires the completed checks into CI.
