# TaskAgent CLI — Build Tasks

## Overview

**Project:** TaskAgent — an AI-agent-native task tracker
**Version:** 0.2.0
**Component covered here:** Shell CLI client
**Prerequisite:** HTTP API server from `docs/0.1.0/build-tasks.md` (must be deployable and usable end-to-end)
**Estimated effort:** ~10 hours across 2 weeks at ~5 hrs/week

### MVP Scope Guardrails

The API server is the source of truth. The CLI is a thin consumer of the same REST surface. v0.2.0 scope:
- Shell CLI implements the full API surface an AI agent needs (projects + tasks + move)
- Auth is API-key only (`X-API-Key` header)
- No drag-and-drop, no real-time updates, no MCP wrapper, no GitHub Issues sync in 0.2.0

The CLI is AI-first: JSON output is the default, human output is opt-in.

---

## Shell CLI Client

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

## Summary

| Part | Tasks | Focus | Est. Hours |
|------|-------|-------|------------|
| A | C-000 - C-015 | Shell CLI client | 10 |

At ~5 hours/week, this is roughly 2 weeks of work. Ship the CLI first, use it against the deployed API, and let real command-line usage inform any later dashboard plan.

**Future work explicitly out of scope for 0.2.0:** React dashboard, MCP server wrapper, GitHub Issues sync, multi-user permissions refinement, real-time updates, drag-and-drop board, and notifications.
