# Implementation Plan: T-051 Input Validation, Error Consistency, and CORS

## 1. Objective

Review the API server’s request validation, error response contract, and CORS behavior so every endpoint returns predictable JSON errors, invalid input is rejected with the correct status code, and browser clients can be explicitly allowed through configuration.

## 2. Scope

T-051 covers three related but separable concerns:

1. Input validation for project, task, auth, and query payloads.
2. Consistent JSON error responses across handlers and middleware.
3. Configurable CORS support for the future React dashboard.

The implementation should keep validation in the handler layer. The store remains responsible for persistence constraints, and services remain responsible for business rules such as invalid task transitions.

## 3. Status Code Contract

- `400`: malformed JSON, missing required fields, invalid field values, invalid query parameters.
- `401`: missing or invalid API key.
- `404`: resource not found.
- `409`: conflict, such as deleting a project that still has tasks.
- `422`: valid request shape that violates a business rule, such as an invalid task status transition.
- `500`: unexpected server errors, always as `{"error":"..."}`.

## 4. Validation Rules

- Project `name` is required and must be at most 100 characters.
- Project `description` must be at most 5000 characters.
- Task `project_id` is required on create.
- Task `title` is required and must be at most 500 characters.
- Task `description` must be at most 5000 characters.
- Task `status`, when supplied by a client, must be one of the public statuses.
- Task `tags` must contain at most 20 items.
- Auth key `label` and `user_name` are required.

## 5. Atomic Engineering Tasks

The granular task breakdown has been split into one file per task to keep implementation context small. Start from the index: [T-051-atomic-tasks.md](/home/nulad/repo/nulad/taskagent/docs/T-051-atomic-tasks.md).

## 6. Definition of Done Mapping

- Validation helper that collects multiple field errors: TASK-001
- Project validation and error statuses: TASK-002
- Task create/update validation: TASK-003
- Task move/list validation: TASK-004
- Auth request validation: TASK-005
- No plain text or HTML error responses from middleware: TASK-006
- `TASKAGENT_CORS_ORIGINS` config: TASK-007
- CORS middleware behavior: TASK-008
- CORS applied to the server stack: TASK-009
- Error response coverage for each category: TASK-010
- Full suite verification and small stabilization fixes: TASK-011
