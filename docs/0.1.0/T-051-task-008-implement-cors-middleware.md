# TASK-008: Implement CORS Middleware

## Token Safety Verification

- **Task ID & Title:** TASK-008: Implement CORS Middleware
- **Objective:** Add middleware that allows configured origins, handles preflight requests, and rejects disallowed origins.
- **Target Files/Scope:** `internal/middleware/cors.go`, `internal/middleware/cors_test.go`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this is a new isolated middleware plus focused tests.
  * Dependencies: CORS origins parsed by TASK-007, existing middleware style.

## Step-by-step Instructions

1. Create `CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler`.
2. For requests without an `Origin` header, pass through without CORS headers.
3. For requests with an allowed `Origin`, set `Access-Control-Allow-Origin` to that exact origin.
4. Include `Vary: Origin` when evaluating origins.
5. Support allowed methods `GET, POST, PUT, PATCH, DELETE`.
6. Include allowed headers `X-API-Key` and `Content-Type`.
7. For allowed preflight `OPTIONS` requests, return `204 No Content`.
8. For disallowed preflight origins, return `403 Forbidden` with a JSON error body.
9. For disallowed actual requests, do not add `Access-Control-Allow-Origin`; return `403 Forbidden` with a JSON error body.
10. Add tests for no origin, allowed actual request, disallowed actual request, allowed preflight, and disallowed preflight.

## Definition of Done

- Allowed origins receive the correct CORS headers.
- Disallowed origins are rejected consistently.
- Preflight requests do not require API key auth when CORS is wired before auth.
