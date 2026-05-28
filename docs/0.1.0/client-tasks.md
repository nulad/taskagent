# TaskAgent Dashboard — Build Tasks

## Overview

**Project:** TaskAgent — an AI-agent-native task tracker
**Components covered here:** React dashboard
**Prerequisite:** HTTP API server from `docs/0.1.0/build-tasks.md` (must be deployable and usable end-to-end)
**Estimated effort:** ~12 hours across 3 weeks at ~5 hrs/week

### MVP Scope Guardrails

The API server is the source of truth. The dashboard is a thin consumer of the same REST surface. v1 scope:
- React dashboard is read-heavy: browse projects, see kanban, create tasks, change status via dropdown
- Auth is API-key only (`X-API-Key` header) — session/cookie auth for the dashboard is Phase 4+
- No drag-and-drop, no real-time updates, no MCP wrapper, no GitHub Issues sync in v1

The dashboard is human-first: it exists so you can glance at what the agents are doing.

---

## Moved: CLI Client

The Python Click CLI client has moved to the 0.2.0 roadmap: [docs/0.2.0/cli-tasks.md](../0.2.0/cli-tasks.md).

---

## React Dashboard

Read-heavy kanban dashboard consuming the TaskAgent HTTP API. React Router v7 in framework mode (not Next.js — see PDF for why), Vite, Tailwind, deployed to Fly.io. v1 renders what the AI agents have done; humans can create tasks and change status via dropdown, but drag-and-drop and real-time updates are Phase 4+.

### D-000: Scaffold React Router v7 project

**Description:**
Scaffold a new React Router v7 project in framework mode with Vite and Tailwind. Lives in `web/` at the repo root so it shares the repo with the Go server but builds independently. Configure TypeScript in strict mode. Set up ESLint + Prettier with sensible defaults. Add a `make -C web dev` target that runs the Vite dev server.

**Definition of Done — Task:**
- `web/` contains a React Router v7 framework-mode app scaffolded via `npx create-react-router@latest`
- Tailwind CSS is configured and working (a root layout with a Tailwind utility class applied renders correctly)
- TypeScript `strict: true` in `tsconfig.json`
- ESLint config extends `@react-router/eslint-config` (or equivalent) with zero warnings on scaffold
- `make -C web dev` starts the Vite dev server on port 5173
- `make -C web build` produces a production build with no errors
- `make -C web typecheck` runs `tsc --noEmit` with zero errors
- README in `web/` documents the setup

**Definition of Done — Learning Objective:**
- Understand React Router v7 framework mode vs SPA mode: framework mode has a server component that runs loaders/actions (similar to Next.js App Router or Remix), SPA mode is pure client-side — framework mode is what maps cleanly to backend route handler thinking
- Understand why Vite over webpack/Next: Vite is the build tool, not the framework — it's unopinionated, fast in dev, and independent of any hosting vendor
- Practice strict TypeScript from the start: `strict: true` catches whole classes of bugs at compile time — turning it on later means backfilling dozens of type assertions

---

### D-001: API client module and typed models

**Description:**
Write `web/app/lib/api.ts` that wraps `fetch` with the `X-API-Key` header, base URL config, and typed request/response. Define TypeScript types that mirror the Go models (`Project`, `Task`, `TaskStatus`). Export typed functions: `listProjects()`, `createTask(input)`, `moveTask(id, status)`, etc. Every function takes the API key as an argument — no module-level state.

**Definition of Done — Task:**
- `web/app/lib/api.ts` exports typed functions for every endpoint the dashboard uses
- `Project`, `Task`, `TaskStatus`, `TaskFilter` types match the JSON contract from T-020
- `TaskStatus` is a union type: `'backlog' | 'todo' | 'in-progress' | 'review' | 'done'`
- A base `request<T>(method, path, body?, apiKey)` helper handles headers, JSON serialization, status code branching
- 4xx responses throw a typed `ApiError` with status, code, and message
- 401 specifically throws `AuthError` so the UI can redirect to login
- Base URL is read from `import.meta.env.VITE_TASKAGENT_API_BASE` at build time
- Unit tests (Vitest) with `msw` for HTTP mocking cover: success, 4xx, 5xx, 401

**Definition of Done — Learning Objective:**
- Understand `import.meta.env` in Vite: env vars prefixed with `VITE_` are embedded at build time — don't use for secrets (they end up in the bundle), but perfect for the API base URL
- Practice typed API clients: defining the response shape in TypeScript means every consumer gets autocomplete and compile-time safety — this is the React/TS equivalent of the Go model package
- Understand why no module-level auth state: passing the API key as an argument to each function keeps the module pure and testable — the auth context provider (D-002) is responsible for injecting it

---

### D-002: API key login and auth context

**Description:**
Build a login screen that accepts an API key (paste from `task add-key` output), validates it by calling `GET /auth/keys`, and stores it in `localStorage`. Expose the key via a React context so loaders/actions can read it. Provide a logout action that clears the key and redirects to login.

**Definition of Done — Task:**
- `/login` route renders a form with an API key textarea and a "Sign in" button
- Submit calls `GET /auth/keys` with the provided key; on 200, stores `taskagent.apiKey` in `localStorage` and redirects to `/`
- On 401, renders `Invalid API key` inline
- An `AuthContext` provider reads the key from `localStorage` on mount and exposes `{ apiKey, signIn, signOut }`
- A `requireAuth()` helper for loaders redirects to `/login` if no key is present
- Logout button in the header clears `localStorage` and navigates to `/login`
- Security note documented in code: `localStorage` is readable by any JS on the origin — this is acceptable for a single-user dashboard on a trusted device, NOT acceptable for multi-tenant deployments (v2 swaps to HttpOnly session cookies)

**Definition of Done — Learning Objective:**
- Understand React context correctly: context is for values that many components need and that change rarely (auth, theme, locale) — overusing context for state that changes often causes re-render cascades
- Understand `localStorage` tradeoffs: synchronous, simple, persists across reloads, but accessible to any script on the origin (XSS risk) — HttpOnly cookies are safer but require server-side session handling
- Practice the protected-route pattern: calling `requireAuth()` at the top of each loader is more explicit than a wrapper component — it mirrors how Go middleware feels (check at the boundary, short-circuit if missing)

---

### D-003: Projects list page

**Description:**
`/projects` lists all projects via a loader. Clicking a project navigates to `/projects/:id` (the kanban board in D-004). Add a "New project" button that opens a modal/form using an action to `POST /projects`. Show empty state when there are no projects.

**Definition of Done — Task:**
- `web/app/routes/projects._index.tsx` defines a loader that calls `listProjects(apiKey)`
- Renders a grid/list of project cards with name, description, task count (count can be "-" for v1 if no endpoint exists yet)
- "New project" opens an inline form with `name` and `description`; submit is a Remix-style action
- On successful create, action redirects to `/projects/:newId`
- Empty state shows "No projects yet" with a prominent "Create your first project" button
- Error boundary renders API errors readably (not a white screen)

**Definition of Done — Learning Objective:**
- Understand loaders vs useEffect + fetch: loaders run on the server (framework mode) or at route transition time, before the component renders — no loading flicker, no useEffect race conditions, and they integrate with React Router's deferred/streaming features
- Understand Remix-style actions: form submissions POST to the same route, the action runs, the loader re-runs, the page re-renders with fresh data — this is the "web-native form" pattern that skips most client state management
- Practice the error boundary: every route should have an error boundary so API failures render a useful message instead of crashing the whole app

---

### D-004: Kanban board for a single project

**Description:**
`/projects/:id` renders the project's tasks grouped by status into five columns: Backlog, Todo, In Progress, Review, Done. Each task is a card with title, short description, tags, and a status dropdown. Changing the dropdown triggers an action that calls `PATCH /tasks/:id/move`. No drag-and-drop in v1.

**Definition of Done — Task:**
- Loader fetches the project and its tasks (`listTasks({ projectId: params.id })`) in parallel
- Tasks are grouped client-side by status into five columns with consistent ordering
- Each card shows title, truncated description (first 100 chars), tag pills, a status `<select>` dropdown, and an open-detail link
- Changing the dropdown submits an action that calls `moveTask(id, newStatus)`
- 422 (invalid transition) renders an inline toast/banner on the card with the server error message
- Column headers show count badges
- Empty columns render a subtle "No tasks" placeholder

**Definition of Done — Learning Objective:**
- Understand optimistic vs pessimistic updates: v1 uses pessimistic (wait for server, then re-render) — simpler and safer; optimistic updates are Phase 4+ when you add drag-and-drop and the latency becomes visible
- Practice composing loaders: one route can load multiple resources in parallel with `Promise.all` — the loader blocks until both resolve, then renders with complete data
- Understand why server-side grouping isn't always better: grouping by status in the client is cheap (5 buckets, small arrays) and keeps the API endpoint generic — server-side grouping would require a new endpoint for dashboard-specific shape

---

### D-005: Task detail and edit

**Description:**
`/tasks/:id` shows full task details (title, full description, all metadata, tags) with an edit mode. Editing submits an action to `PUT /tasks/:id`. Delete button triggers `DELETE /tasks/:id` with a confirmation prompt, then redirects to the parent project board.

**Definition of Done — Task:**
- Loader fetches the task by ID; 404 renders a "Task not found" page with a back link
- Read mode shows all task fields plus created_at / updated_at in a readable format
- Edit button toggles to a form with every editable field
- Save submits to a PUT action, invalidates the loader, re-renders in read mode
- Delete button uses a native `confirm()` (or simple modal) then calls the DELETE action and redirects to `/projects/:projectId`
- Cancel button in edit mode returns to read mode without saving

**Definition of Done — Learning Objective:**
- Understand route-level data invalidation: after an action runs, React Router re-runs the loader for that route automatically — you don't manage cache invalidation manually like you would with TanStack Query
- Practice progressive enhancement: the edit form should work without JS (pure HTML form submission to the action) — framework-mode forms submit over standard POST and upgrade to client-side navigation when JS loads

---

### D-006: Create task form

**Description:**
From the kanban board, a "New task" button per column opens a form pre-filled with that status. The form collects title (required), description, tags (comma-separated). Project is inferred from the current route. Submit posts to `POST /tasks`. Server 422 errors render inline on the relevant field.

**Definition of Done — Task:**
- "New task" button in each column header opens a modal form
- Title field is required (HTML `required`) and `maxLength={500}` matching server validation
- Description is a textarea with `maxLength={5000}`
- Tags input accepts comma-separated values, split client-side into an array before submission
- Status defaults to the column's status; user can change it in the form
- Submit uses a Remix action that calls `createTask`
- Server validation errors render inline: the 422 response's field-level errors map to the matching input
- Success closes the modal, loader re-runs, new card appears in the correct column

**Definition of Done — Learning Objective:**
- Understand native form validation + server validation: browser `required` / `maxLength` give instant feedback but are bypassable — server validation is authoritative — the UI shows both (instant for typos, server for business rules)
- Practice the modal-as-route pattern: the modal can be its own route (`/projects/:id/new-task`) so it's shareable/bookmarkable and back-button-friendly — this is a React Router idiom

---

### D-007: Dockerfile and portable deployment

**Description:**
Multi-stage Dockerfile for the React Router v7 app: build stage runs `pnpm build`, runtime stage runs the RR7 server via Node. Extend the root `docker-compose.yml` (from T-052) with a second service for the dashboard, wired to the same Docker network as the API. `VITE_TASKAGENT_API_BASE` is set at build time via a build arg. The dashboard image is vendor-neutral — same packaging story as the API server (runs on any OCI host, no provider-specific config).

**Definition of Done — Task:**
- `web/Dockerfile` multi-stage: `node:22-alpine` for build, `node:22-alpine` for runtime
- `pnpm install --frozen-lockfile` → `pnpm build` → copy `build/` and `node_modules/` (production only) into runtime image
- Runtime stage runs as a non-root user
- `VITE_TASKAGENT_API_BASE` is a build arg baked into the bundle
- `docker-compose.yml` adds a `web` service mapping port 3000, alongside the existing `api` service, so `docker compose up` brings the whole stack live locally
- The dashboard, hit at `http://localhost:3000`, logs in with a key from the local API and round-trips a create → move → see-on-board flow
- `docs/deployment.md` covers the generic two-service VPS deploy (same compose file works remotely) and calls out that the API and dashboard can be split onto separate hosts or registries if needed
- Final runtime image is under 200MB

**Definition of Done — Learning Objective:**
- Understand build-time vs runtime config for SPAs and frameworks: `import.meta.env.VITE_*` is baked into the JS bundle at build time — runtime env changes require a rebuild, which is a tradeoff vs server-rendered frameworks where env can be injected per-request
- Understand why two images (API + web) instead of one: keeping them separate means the API can be scaled, deployed, and updated independently of the dashboard, and the dashboard can be torn down entirely without affecting the CLI — this separation holds regardless of where they're hosted
- Practice cross-origin considerations: when the dashboard and API are served from different hostnames in production, CORS must be configured on the API server (covered in T-051) — the specific origin values are deploy-time config, not code
- Understand multi-service docker-compose: the `web` service can reference `api` by service name on the internal Docker network, which is the same shape as Kubernetes service discovery — learn the primitive once, it transfers everywhere

---

### D-008: End-to-end smoke test in CI

**Description:**
Add a Playwright smoke test that boots the API server locally, seeds a key, builds the React app in preview mode, opens the dashboard, logs in, creates a project and task, moves it through every status, and asserts the final state. Runs in GitHub Actions on every push to the `web/` directory.

**Definition of Done — Task:**
- `web/e2e/smoke.spec.ts` is a Playwright test that exercises the full login → create → move → delete flow
- A `make -C web e2e` target starts the API (via `go run`), seeds a key, starts the RR7 preview server, runs Playwright, tears everything down
- GitHub Actions workflow runs the smoke test on pushes that modify `web/` or `internal/`
- Test failures upload Playwright traces as artifacts
- Passes green on a clean checkout

**Definition of Done — Learning Objective:**
- Understand why E2E tests are worth their weight: they catch the integration bugs that unit + integration tests miss (CORS misconfigs, env var mismatches, auth flow regressions) — but they're slow and flaky, so keep them focused on golden-path workflows
- Practice test artifacts in CI: uploading Playwright traces on failure means debugging a flaky test doesn't require reproducing locally — you can step through the exact failed run in the trace viewer

---

## Summary

| Part | Tasks | Focus | Est. Hours |
|------|-------|-------|------------|
| B | D-000 – D-008 | React dashboard | 12 |
| **Total** | **9 tasks** | | **~12 hours** |

At ~5 hours/week, this is roughly 3 weeks.

The Python Click CLI plan now belongs to the 0.2.0 roadmap: [docs/0.2.0/cli-tasks.md](../0.2.0/cli-tasks.md).

**Phase 4+ — explicitly not planned in this roadmap:** MCP server wrapper, GitHub Issues sync, multi-user permissions refinement, real-time updates (WebSocket/SSE), drag-and-drop on the board, notifications. Revisit these only after the dashboard is deployed and you've used the tool yourself.
