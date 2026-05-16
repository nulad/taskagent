# Atomic Tasks Index: T-050 Write End-to-End API Tests

Each atomic task lives in its own file so an implementation agent can load only the task it is executing.

1. [TASK-001: Create E2E Test Server Harness](/home/nulad/repo/nulad/taskagent/docs/T-050-task-001-e2e-test-server-harness.md)
2. [TASK-002: Add HTTP JSON Request Helpers](/home/nulad/repo/nulad/taskagent/docs/T-050-task-002-http-json-request-helpers.md)
3. [TASK-003: Test Project Lifecycle Over HTTP](/home/nulad/repo/nulad/taskagent/docs/T-050-task-003-project-lifecycle.md)
4. [TASK-004: Test Task Creation and Status Workflow](/home/nulad/repo/nulad/taskagent/docs/T-050-task-004-task-status-workflow.md)
5. [TASK-005: Test Task List Filtering and Empty Arrays](/home/nulad/repo/nulad/taskagent/docs/T-050-task-005-list-filtering-empty-arrays.md)
6. [TASK-006: Test Invalid Transition Error Contract](/home/nulad/repo/nulad/taskagent/docs/T-050-task-006-invalid-transition-error.md)
7. [TASK-007: Test Authentication Failure Contract](/home/nulad/repo/nulad/taskagent/docs/T-050-task-007-auth-failure-contract.md)
8. [TASK-008: Test Delete Workflow Cleanup](/home/nulad/repo/nulad/taskagent/docs/T-050-task-008-delete-workflow-cleanup.md)
9. [TASK-009: Run and Stabilize E2E Test Suite](/home/nulad/repo/nulad/taskagent/docs/T-050-task-009-run-and-stabilize.md)

## Execution Order

Run the tasks in numeric order. Later tasks assume the test harness and helper functions from earlier tasks exist.
