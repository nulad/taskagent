# Atomic Tasks Index: T-052 Portable Docker Deployment

Each atomic task lives in its own file so an implementation agent can load only the task it is executing.

1. [TASK-001: Create Docker Image Definition](/home/nulad/repo/nulad/taskagent/docs/T-052-task-001-docker-image-definition.md)
2. [TASK-002: Create Docker Compose Deployment](/home/nulad/repo/nulad/taskagent/docs/T-052-task-002-docker-compose-deployment.md)
3. [TASK-003: Add Container Make Targets](/home/nulad/repo/nulad/taskagent/docs/T-052-task-003-container-make-targets.md)
4. [TASK-004: Verify Docker Image Contract](/home/nulad/repo/nulad/taskagent/docs/T-052-task-004-verify-docker-image-contract.md)
5. [TASK-005: Verify Container API Round Trip](/home/nulad/repo/nulad/taskagent/docs/T-052-task-005-container-api-round-trip.md)
6. [TASK-006: Verify Compose SQLite Persistence](/home/nulad/repo/nulad/taskagent/docs/T-052-task-006-compose-sqlite-persistence.md)
7. [TASK-007: Write Generic VPS Deployment Docs](/home/nulad/repo/nulad/taskagent/docs/T-052-task-007-generic-vps-deployment-docs.md)
8. [TASK-008: Run and Stabilize T-052 Suite](/home/nulad/repo/nulad/taskagent/docs/T-052-task-008-run-and-stabilize.md)

## Execution Order

Start with TASK-001 through TASK-003 to create the deployment artifacts. TASK-004 through TASK-006 verify the artifacts in increasing scope. TASK-007 can run after TASK-002 and may run in parallel with TASK-004. TASK-008 must run last.

## Parallelization

- TASK-001, TASK-002, and TASK-003 can run in parallel if agents coordinate the image tag `taskagent:local` and Compose service name `taskagent`.
- TASK-004 can run in parallel with TASK-007 after TASK-001 exists.
- TASK-005 must wait for TASK-001 and TASK-002.
- TASK-006 must wait for TASK-005.
- TASK-008 must wait for all previous tasks.
