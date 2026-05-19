# Atomic Tasks Index: T-051 Input Validation, Error Consistency, and CORS

Each atomic task lives in its own file so an implementation agent can load only the task it is executing.

1. [TASK-001: Create Validation Error Helper](/home/nulad/repo/nulad/taskagent/docs/T-051-task-001-validation-error-helper.md)
2. [TASK-002: Validate Project Requests](/home/nulad/repo/nulad/taskagent/docs/T-051-task-002-validate-project-requests.md)
3. [TASK-003: Validate Task Create and Update Requests](/home/nulad/repo/nulad/taskagent/docs/T-051-task-003-validate-task-create-update.md)
4. [TASK-004: Validate Task Move and List Inputs](/home/nulad/repo/nulad/taskagent/docs/T-051-task-004-validate-task-move-list.md)
5. [TASK-005: Validate Auth Key Requests](/home/nulad/repo/nulad/taskagent/docs/T-051-task-005-validate-auth-key-requests.md)
6. [TASK-006: Normalize Middleware Error Responses](/home/nulad/repo/nulad/taskagent/docs/T-051-task-006-normalize-middleware-errors.md)
7. [TASK-007: Add CORS Configuration](/home/nulad/repo/nulad/taskagent/docs/T-051-task-007-add-cors-configuration.md)
8. [TASK-008: Implement CORS Middleware](/home/nulad/repo/nulad/taskagent/docs/T-051-task-008-implement-cors-middleware.md)
9. [TASK-009: Wire CORS Into Server Stack](/home/nulad/repo/nulad/taskagent/docs/T-051-task-009-wire-cors-server-stack.md)
10. [TASK-010: Add Error Contract Integration Tests](/home/nulad/repo/nulad/taskagent/docs/T-051-task-010-error-contract-integration-tests.md)
11. [TASK-011: Run and Stabilize T-051 Suite](/home/nulad/repo/nulad/taskagent/docs/T-051-task-011-run-and-stabilize.md)

## Execution Order

Run the tasks in numeric order. Later tasks assume the validation helper and CORS config/middleware from earlier tasks exist.
