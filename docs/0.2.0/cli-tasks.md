# TaskAgent CLI - Build Tasks

## Overview

**Project:** TaskAgent - an AI-agent-native task tracker
**Version:** 0.2.0
**Component covered here:** Python CLI client using Click
**Prerequisite:** HTTP API server from `docs/0.1.0/build-tasks.md` (must be deployable and usable end-to-end)
**Estimated effort:** ~10 hours across 2 weeks at ~5 hrs/week

### MVP Scope Guardrails

The API server is the source of truth. The CLI is a thin consumer of the same REST surface. v0.2.0 scope:
- Python CLI implements the full API surface an AI agent needs (projects + tasks + move)
- Auth is API-key only (`X-API-Key` header)
- JSON output is the default; human output is opt-in
- No dashboard, no real-time updates, no MCP wrapper, no GitHub Issues sync in 0.2.0

The CLI is AI-first: stdout stays machine-parseable by default, while errors and progress go to stderr.

---

## Python Click CLI Client

A Python package built with Click for command structure, `httpx` for HTTP, and standard-library JSON/config handling where possible. It lives in `cli/` so it can be packaged and released independently of the Go server binary.

Target layout:

```text
cli/
  pyproject.toml
  src/taskagent_cli/
    __init__.py
    __main__.py
    cli.py
    api.py
    config.py
    format.py
    models.py
  tests/
    test_config.py
    test_api.py
    test_commands.py
    test_smoke.py
```

Use the console script name `task` so the user-facing command remains:

```bash
task projects list
task add --title "Write docs" --project <project-id>
```

---

### C-000: Bootstrap Python package and Click entry point

**Description:**
Create a Python package under `cli/` with a Click command group exposed as a `task` console script. Add the top-level command, version flag, help text, and command groups for auth, projects, tasks, completion, and diagnostics. Use `uv` or `pipx`-compatible packaging via `pyproject.toml`.

**Definition of Done - Task:**
- `cli/pyproject.toml` defines the package, dependencies, and a `[project.scripts] task = "taskagent_cli.cli:main"` entry point
- Runtime dependencies include `click` and `httpx`
- Dev dependencies include `pytest`, `pytest-httpx` or `respx`, `ruff`, and `mypy`
- `python -m taskagent_cli --help` and installed `task --help` both work
- `task --version` prints the CLI version from one package constant
- Running `task` with no subcommand prints help and exits 2
- Unknown commands are handled by Click with a readable usage error
- `make -C cli check` runs `ruff check`, `ruff format --check`, `mypy`, and `pytest`

**Definition of Done - Learning Objective:**
- Understand Click command groups, decorators, options, arguments, and context objects
- Understand Python packaging basics: `pyproject.toml`, console scripts, editable installs, and dependency groups
- Practice keeping CLI wiring thin: Click parses input, then calls small functions in API/config modules

---

### C-001: Configuration and credential storage

**Description:**
Load configuration from `$TASKAGENT_HOME/config.json` (default `~/.taskagent/config.json`), falling back to environment variables, falling back to defaults. Required values are `TASKAGENT_SERVER` (base URL, e.g. `http://localhost:8080`) and `TASKAGENT_API_KEY`. Implement `task login --server URL --key KEY` and `task login --key -` for stdin-based key entry.

**Definition of Done - Task:**
- `taskagent_cli/config.py` contains a typed `Config` dataclass
- Config precedence is: CLI flags > environment variables > config file > defaults
- Default `TASKAGENT_SERVER` is `http://localhost:8080`
- `task login --server URL --key KEY` creates `$TASKAGENT_HOME` with mode `0700` and writes `config.json` with mode `0600`
- `task login --key -` reads the key from stdin without echoing or requiring it in shell history
- Missing server or API key on authenticated commands exits 2 with a clear message pointing to `task login`
- `task whoami` calls `GET /auth/keys` with the configured key and prints key metadata as JSON by default
- Unit tests verify config precedence, permission mode setting where supported, stdin key handling, and missing-config errors

**Definition of Done - Learning Objective:**
- Understand config precedence layers common to CLIs
- Practice safe local credential storage with strict file permissions
- Understand why JSON config is used here: Python can parse it with the standard library, it is explicit, and it avoids executable config files

---

### C-002: HTTP client and error mapping

**Description:**
Write a single `TaskAgentClient` in `taskagent_cli/api.py` that wraps `httpx`, injects the `X-API-Key` header, serializes JSON request bodies, parses JSON responses, and maps HTTP/network failures into CLI-friendly exceptions.

**Definition of Done - Task:**
- `TaskAgentClient` accepts base URL, API key, timeout, and optional injected `httpx.Client`
- All requests set `X-API-Key`, `Content-Type: application/json` when a body exists, and `Accept: application/json`
- 2xx responses return parsed JSON, or `None` for empty `204` responses
- 4xx responses raise `ApiClientError` and CLI exits 3
- 5xx responses raise `ApiServerError` and CLI exits 4
- Network errors and timeouts raise `NetworkError` and CLI exits 5
- Error messages prefer the server JSON error message/code when present, falling back to response text
- Respects `TASKAGENT_TIMEOUT` with a default of 10 seconds
- Tests with mocked HTTP responses cover success, 204, 400/401/404/409/422, 500, timeout, and connection failure

**Definition of Done - Learning Objective:**
- Understand why an explicit API client boundary is better than calling `httpx` directly from every command
- Practice stable process exit codes: 0 success, 2 usage/config, 3 API client error, 4 API server error, 5 network error
- Understand stdout vs stderr for scriptable CLIs

---

### C-003: Typed request/response models

**Description:**
Define lightweight Python models for projects, tasks, task filters, and task statuses. Use dataclasses or `typing.TypedDict` to keep the schema visible without adding a heavy validation framework.

**Definition of Done - Task:**
- `taskagent_cli/models.py` defines `TaskStatus`, `Project`, `Task`, and request payload types
- Valid statuses are centralized: `backlog`, `todo`, `in-progress`, `review`, `done`
- JSON keys match the Go API contract (`project_id`, `created_at`, `updated_at`, etc.)
- Helper functions parse comma-separated tags into a clean list
- Helper functions build update payloads that include only explicitly provided fields
- Unit tests cover status validation, tag parsing, and partial update payload construction

**Definition of Done - Learning Objective:**
- Understand Python's practical type tools for CLI programs: dataclasses, `Literal`, `TypedDict`, and `Mapping`
- Practice keeping API schema knowledge in one module instead of scattering string literals across commands

---

### C-010: Project subcommands

**Description:**
Implement `task projects list`, `task projects create`, `task projects show`, `task projects update`, and `task projects delete`. Each maps one-to-one to the server project endpoints. Output is raw JSON by default. Accept `--format human` for readable tables/details.

**Definition of Done - Task:**
- `task projects list` calls `GET /projects` and prints a JSON array (`[]` if empty, never `null`)
- `task projects create --name NAME [--description TEXT]` calls `POST /projects`
- `task projects show ID` calls `GET /projects/{id}`
- `task projects update ID [--name NAME] [--description TEXT]` calls `PUT /projects/{id}` with only provided fields
- `task projects delete ID` calls `DELETE /projects/{id}` and prints no JSON on `204`
- 404 exits 3 with `not found: <id>` or the server error message on stderr
- 409 exits 3 with the server conflict message on stderr
- `--format human` renders readable project tables/details without changing JSON defaults
- Command tests verify URL, method, payload, JSON output, human output, and error exits

**Definition of Done - Learning Objective:**
- Practice Click command groups and shared options
- Understand machine-first output: default JSON means agents and scripts do not parse human tables
- Practice command tests with Click's `CliRunner`

---

### C-011: Task subcommands (CRUD)

**Description:**
Implement `task add`, `task list`, `task show`, `task update`, and `task delete`. The `add` and `update` commands accept structured flags (`--title`, `--description`, `--project`, `--tags`). The `list` command accepts `--project`, `--status`, `--limit`, and `--offset` and maps them to query parameters.

**Definition of Done - Task:**
- `task add --title T --project P [--description D] [--tags a,b,c]` calls `POST /tasks`
- `task list [--project P] [--status S] [--limit N] [--offset N]` calls `GET /tasks` with query params
- `task show ID` calls `GET /tasks/{id}`
- `task update ID [--title T] [--description D] [--tags a,b]` calls `PUT /tasks/{id}` with only provided fields
- `task delete ID` calls `DELETE /tasks/{id}` and prints no JSON on `204`
- `--project` accepts either a project ID or a project name; names resolve via `/projects` lookup when the value does not look like a UUID
- `--format human` renders a one-line-per-task table with short id, status, title, and project
- Tests verify every flag combination, query encoding, tag parsing, project-name resolution, JSON output, and human output

**Definition of Done - Learning Objective:**
- Practice Click options, required flags, typed integer options, and validation callbacks
- Understand conditional JSON construction for partial updates
- Practice small UX improvements that do not require server changes

---

### C-012: Task move (status transition)

**Description:**
Implement `task move ID STATUS`, which calls `PATCH /tasks/{id}/move` with `{"status": "..."}`. Validate the status locally against the known list before sending the request. Surface the server's 422 transition error clearly.

**Definition of Done - Task:**
- `task move ID STATUS` validates `STATUS` before hitting the network
- Invalid status exits 2 with `invalid status: X (must be one of: backlog, todo, in-progress, review, done)`
- Valid status calls `PATCH /tasks/{id}/move`
- 422 response prints the server's transition error to stderr and exits 3
- `task move` completion suggests status values
- Tests cover valid transition, status typo, invalid server-side transition, and missing task

**Definition of Done - Learning Objective:**
- Understand client-side validation as UX, not security
- Practice respecting server-owned business rules: the CLI validates known status values but does not duplicate the transition graph

---

### C-013: Human-format rendering helpers

**Description:**
Centralize `--format human` rendering in `taskagent_cli/format.py`. JSON output should always be produced by `json.dumps`; human output should be layered on top and used only when requested.

**Definition of Done - Task:**
- `format.py` exports `print_json`, `render_task_table`, `render_project_table`, `render_task_detail`, and `render_project_detail`
- JSON output is deterministic enough for tests: stable key order is preferred
- Human table rendering uses the standard library or a small dependency already justified in `pyproject.toml`
- Color is enabled only when stdout is a TTY and `NO_COLOR` is unset
- Status color mapping: backlog=gray, todo=yellow, in-progress=blue, review=magenta, done=green
- Timestamps are shortened in table view but full RFC3339 in detail view
- ID is truncated to first 8 chars in table view and full UUID in detail view
- Tests cover TTY/non-TTY behavior, `NO_COLOR`, empty lists, long titles, and detail rendering

**Definition of Done - Learning Objective:**
- Understand separation of data and presentation
- Practice predictable JSON output for automation
- Understand terminal color etiquette and the `NO_COLOR` convention

---

### C-014: Command completion via Click

**Description:**
Provide shell completion for bash, zsh, and fish using Click's built-in completion support. Add a `task completion SHELL` helper command that prints the generated script or clear installation instructions.

**Definition of Done - Task:**
- Completion works for bash, zsh, and fish through Click's documented environment-variable mechanism
- `task completion bash`, `task completion zsh`, and `task completion fish` print sourceable completion scripts or exact setup commands
- Static completion covers command names, long flags, and task status values
- `task move <id> <TAB>` completes the five status values for the `STATUS` argument
- README documents persistent completion installation for common shells
- Manual testing notes document any dynamic completion that is skipped

**Definition of Done - Learning Objective:**
- Understand how Click exposes shell completion
- Practice making the CLI self-documenting for humans without changing the machine-facing command contract

---

### C-015: Release packaging and smoke test

**Description:**
Package the CLI as a Python distribution and provide a simple local install path. Add an end-to-end smoke test that starts the Go server, seeds an API key, configures the CLI, and exercises the full CRUD-and-move workflow.

**Definition of Done - Task:**
- `make -C cli build` produces wheel and source distributions under `cli/dist/`
- `pipx install ./cli/dist/*.whl` installs a working `task` command
- `make -C cli test` runs unit and command tests
- `make -C cli smoke` starts the API server with a temporary database, seeds a key, runs the CLI workflow, and tears everything down
- Smoke workflow: seed key -> login -> create project -> create 3 tasks -> move one through every status -> list with filters -> delete tasks -> delete project
- A GitHub Actions workflow runs lint, typecheck, tests, build, and smoke tests on pushes that modify `cli/`, `cmd/`, `internal/`, `migrations/`, or relevant docs

**Definition of Done - Learning Objective:**
- Understand Python wheel/source distribution packaging
- Practice `pipx` as the right install path for end-user Python CLIs
- Understand smoke tests vs unit tests: smoke tests catch endpoint, auth, and JSON-contract integration failures that command tests miss

---

## Summary

| Part | Tasks | Focus | Est. Hours |
|------|-------|-------|------------|
| A | C-000 - C-015 | Python Click CLI client | 10 |

At ~5 hours/week, this is roughly 2 weeks of work. Ship the CLI first, use it against the deployed API, and let real command-line usage inform any later dashboard plan.

**Future work explicitly out of scope for 0.2.0:** React dashboard, MCP server wrapper, GitHub Issues sync, multi-user permissions refinement, real-time updates, drag-and-drop board, and notifications.
