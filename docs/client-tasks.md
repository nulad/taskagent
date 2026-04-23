# TaskAgent Clients — Build Tasks

## Overview

**Project:** TaskAgent — an AI-agent-native task tracker
**Components covered here:** Shell CLI client + React dashboard
**Prerequisite:** HTTP API server from `docs/build-tasks.md` (must be deployable and usable end-to-end)
**Estimated effort:** ~22 hours across 5 weeks at ~5 hrs/week

### MVP Scope Guardrails

The API server is the source of truth. Both clients are thin consumers of the same REST surface. v1 scope:
- Shell CLI implements the full API surface an AI agent needs (projects + tasks + move)
- React dashboard is read-heavy: browse projects, see kanban, create tasks, change status via dropdown
- Auth is API-key only (`X-API-Key` header) for both clients — session/cookie auth for the dashboard is Phase 4+
- No drag-and-drop, no real-time updates, no MCP wrapper, no GitHub Issues sync in v1

The CLI is AI-first: JSON output is the default, human output is opt-in. The dashboard is human-first: it exists so you can glance at what the agents are doing.

---

## Part A: Shell CLI Client

A POSIX-sh-compatible CLI that wraps `curl` and `jq`. Written in shell so AI agents can read the source, audit behavior, and extend commands without a compile step. Lives in a separate repo subdirectory so it can be distributed independently of the Go server binary.

### C-000: Bootstrap CLI entry point and subcommand dispatch

**Description:**
Create `cli/task` as an executable shell script with a subcommand dispatch pattern (`task <subcommand> [args...]`). Subcommand implementations live in `cli/lib/cmd_<name>.sh` and are sourced on demand. Add a `cli/lib/common.sh` for shared helpers (config loading, HTTP request wrapper, error printing). Add a `task --help` that lists subcommands and exits 0. Add `task --version` that prints a version string.

**Definition of Done — Task:**
- `cli/task` is an executable script with `#!/usr/bin/env sh` and `set -eu` at the top
- Running `task` with no arguments prints usage and exits 2
- `task --help` and `task help` both print usage and exit 0
- `task --version` prints a version constant and exits 0
- Unknown subcommands print `unknown command: X` to stderr and exit 2
- `cli/lib/common.sh` contains `die()`, `log()`, and `require_cmd()` helpers
- `require_cmd curl jq` fails fast with a clear message if either is missing
- A `make -C cli check` target runs `shellcheck` on every `.sh` file with zero warnings

**Definition of Done — Learning Objective:**
- Understand `set -eu` and `set -o pipefail` (where supported): these turn silent shell failures into loud ones — unset variable reads abort, command failures abort, pipe failures propagate
- Understand why POSIX `sh` over `bash`: shell scripts written to the POSIX subset run on Alpine (busybox ash), macOS (old bash 3.2), and most container images without a bash dependency — this matters for AI agents running in minimal environments
- Practice the subcommand dispatch pattern: `case "$1" in add) shift; cmd_add "$@" ;; esac` — this is the shell idiom that mirrors how `git` and `kubectl` structure their CLIs
- Understand `shellcheck`: a static analyzer for shell that catches real bugs (unquoted expansions, subshell gotchas, portability issues) — treat its warnings as errors

---

### C-001: Configuration and credential storage

**Description:**
Load configuration from `$TASKAGENT_HOME/config` (default `~/.taskagent/config`), falling back to environment variables, falling back to defaults. Config format is `KEY=VALUE` lines — no TOML or YAML parser needed. Required values: `TASKAGENT_SERVER` (base URL, e.g. `http://localhost:8080`) and `TASKAGENT_API_KEY`. A `task login --server <url> --key <key>` subcommand writes the config file with mode `0600`.

**Definition of Done — Task:**
- `cli/lib/common.sh` has a `load_config()` function that sources the config file if it exists
- Environment variables override config file values; CLI flags override environment variables
- `task login --server URL --key KEY` creates `$TASKAGENT_HOME` with mode `0700` and writes the config with mode `0600`
- `task login --key -` reads the key from stdin so it doesn't appear in shell history or `ps` output
- Missing `TASKAGENT_SERVER` or `TASKAGENT_API_KEY` on any authenticated command prints a clear error pointing to `task login`
- `task whoami` hits `GET /auth/keys` with the configured key and prints the key label + user on success
- Tests (a `cli/test/test_config.sh` runnable script) verify precedence: flag > env > file > default

**Definition of Done — Learning Objective:**
- Understand Unix file permissions for secrets: `chmod 600` on the config file and `chmod 700` on the directory is the minimum bar for storing an API key — `ssh` and `gnupg` use the same pattern and will refuse to read world-readable key files
- Practice safe input handling: reading the key from stdin with `--key -` avoids leaking it to shell history, `ps aux`, and process argument logs on shared systems
- Understand config precedence layers: flag → env → file → default is the standard precedence for CLIs (matches `kubectl`, `aws`, `gh`) — higher specificity wins
- Practice POSIX sourcing: `. "$config_file"` (dot, not `source`) is the POSIX-portable way to load config; `source` is a bash-ism

---

### C-002: HTTP request wrapper with error mapping

**Description:**
Write a single `api_request METHOD PATH [BODY]` helper in `cli/lib/common.sh` that wraps `curl`, injects the `X-API-Key` header, captures both status code and body, and maps HTTP errors to exit codes + human-readable stderr messages. All subcommands call this helper; none invoke `curl` directly. On 4xx/5xx, extract the `error` field from the JSON response body and print it.

**Definition of Done — Task:**
- `api_request` takes: method, path (without base URL), optional JSON body (read from stdin or a file)
- Sets `X-API-Key`, `Content-Type: application/json`, `Accept: application/json` headers
- Captures status code via `curl -w '%{http_code}'` or `--write-out`, separates it from the body
- 2xx: prints body to stdout, exits 0
- 4xx: prints `error` field from body to stderr, exits 3
- 5xx: prints `error` field (or raw body if not JSON) to stderr, exits 4
- Network errors (curl exit != 0): prints `network error: <message>` to stderr, exits 5
- Respects `TASKAGENT_TIMEOUT` (default 10s) via `curl --max-time`
- Integration tests using a fake server (`python3 -m http.server` or `nc` one-shot) verify each error path

**Definition of Done — Learning Objective:**
- Understand `curl --write-out` vs parsing headers: `--write-out '%{http_code}'` gives you the status code as a separate stream from the body, which is the clean way to branch on status in shell
- Practice distinct exit codes per error class: 0 = success, 2 = usage error, 3 = API client error (4xx), 4 = API server error (5xx), 5 = network error — this lets agents and scripts branch on exit code without parsing output
- Understand why stderr matters for CLIs: stdout is for machine-parseable output (JSON responses), stderr is for human messages (errors, progress) — agents piping `task list | jq` should never see error text in the pipeline
- Understand `curl -f` vs manual status checking: `-f` makes curl exit non-zero on 4xx/5xx but discards the response body — not useful here because we need the body to extract the error message

---

### C-010: Project subcommands

**Description:**
Implement `task projects list`, `task projects create`, `task projects show`, `task projects update`, `task projects delete`. Each maps one-to-one to a server endpoint from `docs/build-tasks.md` (T-032). Output is raw JSON by default. Accept `--format human` to pretty-print via `jq` + `column`.

**Definition of Done — Task:**
- `cli/lib/cmd_projects.sh` implements all five subcommands
- `task projects create --name NAME [--description TEXT]` POSTs to `/projects` and prints the created project as JSON
- `task projects list` GETs `/projects` and prints the array as JSON (`[]` if empty, never `null`)
- `task projects show ID` GETs `/projects/{id}`; 404 exits 3 with `not found: <id>` on stderr
- `task projects update ID [--name NAME] [--description TEXT]` PUTs `/projects/{id}` with a merged body
- `task projects delete ID` DELETEs `/projects/{id}`; 409 (conflict — tasks exist) exits 3 with the server's error message
- `--format human` on `list` and `show` renders a table via `jq -r` + `column -t -s $'\t'`
- Shell tests verify each path against a running dev server

**Definition of Done — Learning Objective:**
- Practice building JSON payloads safely in shell: use `jq -n --arg name "$NAME" '{name: $name}'` instead of string concatenation — this handles quoting, escaping, and nulls correctly without shell injection
- Understand `jq -r` (raw output): `-r` strips JSON string quotes so `jq -r '.id'` prints `abc123` instead of `"abc123"` — essential when piping IDs into other commands
- Practice machine-first, human-opt-in output: AI agents get JSON by default so they never need to parse human-formatted tables; humans add `--format human` when they want readability

---

### C-011: Task subcommands (CRUD)

**Description:**
Implement `task add`, `task list`, `task show`, `task update`, `task delete`. The `add` and `update` commands accept structured flags (`--title`, `--description`, `--project`, `--tags`). The `list` command accepts `--project`, `--status`, `--limit`, `--offset` and maps them to query parameters. Tags are passed as comma-separated (`--tags bug,urgent`) and converted to a JSON array via `jq`.

**Definition of Done — Task:**
- `cli/lib/cmd_tasks.sh` implements all five subcommands
- `task add --title T --project P [--description D] [--tags a,b,c]` POSTs to `/tasks` and prints the created task
- `task list [--project P] [--status S] [--limit N] [--offset N]` GETs `/tasks` with query params
- `task show ID` GETs `/tasks/{id}`
- `task update ID [--title T] [--description D] [--tags a,b]` PUTs `/tasks/{id}`; only flags that are provided are included in the body
- `task delete ID` DELETEs `/tasks/{id}`
- `--project` accepts either a project ID or a project name (resolves name to ID via `/projects` lookup when it doesn't look like a UUID)
- `--format human` renders a one-line-per-task table with id (short), status, title, project
- Tests verify every flag combination and the tags-to-JSON-array conversion

**Definition of Done — Learning Objective:**
- Practice argument parsing in POSIX shell: a `while [ $# -gt 0 ]; do case "$1" in --title) TITLE=$2; shift 2 ;; esac; done` loop is the portable pattern — `getopts` is POSIX but doesn't support long options
- Understand conditional JSON construction: only include fields the user specified on update, so partial updates don't wipe out unchanged columns — build the body incrementally with `jq --arg` and merge with `+`
- Practice UX polish without scope creep: accepting project name OR ID is a small convenience that makes the CLI dramatically more pleasant without adding server complexity

---

### C-012: Task move (status transition)

**Description:**
Implement `task move ID STATUS` which PATCHes `/tasks/{id}/move` with `{"status": "..."}`. Validate the status locally against the known list (`backlog`, `todo`, `in-progress`, `review`, `done`) before sending the request — fail fast with a helpful error if the user types `in_progress` instead of `in-progress`. Surface the server's 422 transition error clearly.

**Definition of Done — Task:**
- `task move ID STATUS` validates STATUS against a shell constant list before hitting the network
- Invalid status exits 2 with `invalid status: X (must be one of: backlog, todo, in-progress, review, done)`
- Valid status: PATCHes `/tasks/{id}/move` with the JSON body
- 422 response (invalid transition) prints the server's error message to stderr and exits 3
- `task move` tab-completes status values (see C-014)
- Tests cover: valid transition, typo in status, invalid transition (server-side 422), non-existent task (404)

**Definition of Done — Learning Objective:**
- Understand client-side validation as UX, not security: the server is still authoritative (it will reject invalid transitions), but catching obvious typos locally saves a network round-trip and gives a faster, clearer error
- Practice respecting the server contract: the CLI doesn't encode the transition graph (e.g. "backlog → todo only") — that's server business logic and duplicating it here would mean two places to update when rules change

---

### C-013: Human-format rendering helpers

**Description:**
Extract the `--format human` rendering logic into `cli/lib/format.sh` so the projects, tasks, and auth commands share it. Provide `render_task_table` and `render_project_table` functions that read a JSON array from stdin and emit a columnar layout. Use color codes when stdout is a TTY (`[ -t 1 ]`), plain text when redirected to a file or pipe.

**Definition of Done — Task:**
- `cli/lib/format.sh` exports `render_task_table`, `render_project_table`, `render_task_detail`
- Each function reads JSON from stdin and writes formatted text to stdout
- Color is enabled only when `[ -t 1 ]` is true and `NO_COLOR` is unset (respect the `NO_COLOR` convention)
- Status column is colorized: backlog=gray, todo=yellow, in-progress=blue, review=magenta, done=green
- Timestamps are shortened in table view (`2026-04-22 14:30`) but full RFC3339 in detail view
- ID is truncated to first 8 chars in table view, full UUID in detail view

**Definition of Done — Learning Objective:**
- Understand the `NO_COLOR` convention: users set `NO_COLOR=1` globally to disable color across well-behaved CLIs (documented at no-color.org) — CI logs and screen readers both benefit from honoring it
- Practice TTY detection: `[ -t 1 ]` tests whether stdout is a terminal — writing color codes into a piped-to-file output pollutes the file with escape sequences, which breaks downstream parsing
- Understand separation of rendering from data: the same JSON flows out of `api_request` regardless of output format; rendering is a terminal concern layered on top, not baked into the request layer

---

### C-014: Shell completion for bash and zsh

**Description:**
Generate tab completion for subcommands, flags, and dynamic values (project names, status values). Install via `task completion bash` / `task completion zsh` which print a completion script to stdout for the user to source. Dynamic completion (e.g. `task move <TAB>` showing task IDs) is a stretch goal — static completion for subcommands and status values is the minimum.

**Definition of Done — Task:**
- `task completion bash` and `task completion zsh` print a complete, sourceable completion script
- Static completion: subcommands (`add`, `list`, etc.), long flags (`--project`, `--status`), known status values
- `task move <id> <TAB>` completes the five status values
- README documents: `task completion bash > ~/.local/share/bash-completion/completions/task`
- Tested manually in both bash and zsh; document known limitations if dynamic completion is skipped

**Definition of Done — Learning Objective:**
- Understand the difference between `complete -F` (bash) and `compdef` (zsh): the underlying mechanisms differ, which is why most CLIs generate per-shell scripts rather than trying to write one portable file
- Understand why completion is AI-first too: even though agents don't tab-complete, having the completion script also serves as a machine-readable inventory of the CLI's flag surface — it's a fallback introspection mechanism

---

### C-015: Release packaging and smoke test

**Description:**
Package the CLI as a single tarball (`task-<version>.tar.gz`) containing `task`, `lib/*.sh`, and an `install.sh`. The install script copies files into `~/.local/bin` and `~/.local/share/taskagent/lib`. Write a `cli/test/smoke.sh` that spins up the server (via `go run ./cmd/server`), seeds an API key, configures the CLI, and exercises the full CRUD-and-move workflow end-to-end.

**Definition of Done — Task:**
- `make -C cli dist` produces `dist/task-<version>.tar.gz`
- `install.sh` is idempotent (re-running it updates in place, doesn't duplicate)
- `cli/test/smoke.sh` runs end-to-end against a freshly-started server and passes with exit 0
- Smoke test workflow: seed key → login → create project → create 3 tasks → move one through every status → list with filters → delete one → delete project
- A GitHub Actions workflow runs `shellcheck` and `smoke.sh` on every push

**Definition of Done — Learning Objective:**
- Understand smoke tests vs unit tests: a smoke test proves the thing runs end-to-end against a real server — it catches integration bugs (wrong endpoint, wrong JSON shape) that unit tests miss
- Practice idempotent installers: re-running an installer should never leave the system in a broken state — this is what `curl | sh` installers get right and what package-manager-free tools must replicate
- Understand the `~/.local/bin` + XDG convention: installing into the user's home (not `/usr/local/bin`) avoids sudo and matches where modern CLIs deploy (gh, fly, etc.)

---

## Part B: React Dashboard

Read-heavy kanban dashboard consuming the same HTTP API the CLI uses. React Router v7 in framework mode (not Next.js — see PDF for why), Vite, Tailwind, deployed to Fly.io. v1 renders what the AI agents have done; humans can create tasks and change status via dropdown, but drag-and-drop and real-time updates are Phase 4+.

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
| A | C-000 – C-015 | Shell CLI client | 10 |
| B | D-000 – D-008 | React dashboard | 12 |
| **Total** | **17 tasks** | | **~22 hours** |

At ~5 hours/week, ~5 weeks. Sequencing options:
- **Strict:** finish Part A before Part B (most honest — you use the CLI with Claude Code for weeks, then build the dashboard with real data to display)
- **Interleaved:** build C-000–C-012 (core CLI) first, then jump to D-000–D-004 (minimum viable dashboard), then return to C-013–C-015 and D-005–D-008 for polish

The PDF conversation argues for strict sequencing: ship the CLI, use it yourself for a few weeks with Claude Code, then build the dashboard on top of the proven API. That discipline is what keeps the scope honest.

**Phase 4+ — explicitly not planned in this roadmap:** MCP server wrapper, GitHub Issues sync, multi-user permissions refinement, real-time updates (WebSocket/SSE), drag-and-drop on the board, notifications. Revisit these only after Part A + Part B are deployed and you've used the tool yourself.
