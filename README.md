# TaskAgent

TaskAgent is an AI-agent-native task manager for tracking projects and tasks through a small HTTP API and a scriptable CLI.

It is designed for humans and coding agents that need a predictable task backend: create projects, add tasks, list work by status, move tasks through a workflow, and integrate with automation using JSON over HTTP.

## Features

- Project and task management over HTTP
- Python CLI for day-to-day use and automation
- API-key authentication with a one-time bootstrap command
- SQLite persistence with automatic migrations
- Docker and Docker Compose deployment
- JSON responses by default, with human-readable CLI output where supported

## Status

TaskAgent is early software. The current server and CLI are usable for local and self-hosted workflows, but APIs and packaging may still change.

## Requirements

- Go 1.25+
- Docker 24+ and Docker Compose v2, if running with containers
- Python 3.9+, if using or developing the CLI

## Quick Start

Build and start the server with Docker Compose:

```bash
make docker-build
make compose-up
```

In another shell, create the first API key:

```bash
SEED_USER=admin LABEL=bootstrap make compose-run-seed
```

The seed command prints a generated API key once. Use that key for the CLI and HTTP API.

Install the CLI from the local package:

```bash
cd cli
pip install -e .
```

Log in and create your first task:

```bash
taskagent login --server http://localhost:8080 --key <your-api-key>
taskagent projects list --format human
taskagent add --project default --title "Write README" --description "Draft public project docs"
taskagent list --format human
```

Check the server health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## Running Locally

Run the Go server directly:

```bash
make run
```

By default this starts the API on `:8080` and stores SQLite data in `taskagent.db`.

Common development commands:

```bash
make build
make test
make lint
```

Container commands:

```bash
make docker-build
make compose-up
make compose-logs
make compose-down
make compose-smoke
```

## Configuration

The server reads configuration from environment variables.

| Variable | Description | Default |
| --- | --- | --- |
| `TASKAGENT_LISTEN_ADDR` | HTTP listen address | `:8080` |
| `TASKAGENT_DB_PATH` | SQLite database path | `taskagent.db` locally, `/data/taskagent.db` in Docker Compose |
| `TASKAGENT_LOG_LEVEL` | Log level: `debug`, `info`, `warn`, or `error` | `info` |
| `TASKAGENT_CORS_ORIGINS` | Comma-separated list of allowed browser origins | empty |

Example:

```bash
TASKAGENT_LISTEN_ADDR=:9090 \
TASKAGENT_DB_PATH=/tmp/taskagent.db \
TASKAGENT_LOG_LEVEL=debug \
make run
```

The CLI reads configuration from explicit command options, environment variables, and `~/.taskagent/config.json`.

| Variable | Description | Default |
| --- | --- | --- |
| `TASKAGENT_HOME` | Directory for CLI config | `~/.taskagent` |
| `TASKAGENT_SERVER` | Server URL | `http://localhost:8080` |
| `TASKAGENT_API_KEY` | API key used by the CLI | empty |
| `TASKAGENT_TIMEOUT` | Request timeout in seconds | `10` |

Run `taskagent login --server <url> --key <key>` to write CLI config with owner-only file permissions.

## CLI Usage

Common commands:

```bash
taskagent --help
taskagent --version
taskagent login --server http://localhost:8080 --key <key>
taskagent whoami

taskagent projects list
taskagent projects create --name "Website" --description "Website tasks"
taskagent projects show <project-id>
taskagent projects update <project-id> --name "New name"
taskagent projects delete <project-id>

taskagent add --project <project-id-or-name> --title "Task title"
taskagent list --project <project-id-or-name> --status todo
taskagent show <task-id>
taskagent update <task-id> --title "Updated title" --tags "docs,api"
taskagent move <task-id> in-progress
taskagent delete <task-id>
```

Valid task statuses are:

```text
backlog, todo, in-progress, review, done, closed
```

See [cli/README.md](cli/README.md) for CLI packaging, development installation, and shell completion details.

## HTTP API

All routes except `GET /health` require an API key in the `X-API-Key` header.

```bash
curl -H "X-API-Key: <key>" http://localhost:8080/projects
```

Core endpoints:

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check |
| `GET` | `/projects` | List projects |
| `POST` | `/projects` | Create a project |
| `GET` | `/projects/{id}` | Get a project |
| `PUT` | `/projects/{id}` | Update a project |
| `DELETE` | `/projects/{id}` | Delete a project |
| `GET` | `/tasks` | List tasks; supports `project_id`, `status`, `limit`, and `offset` |
| `POST` | `/tasks` | Create a task |
| `GET` | `/tasks/{id}` | Get a task |
| `PUT` | `/tasks/{id}` | Update a task |
| `PATCH` | `/tasks/{id}/move` | Move a task to another status |
| `DELETE` | `/tasks/{id}` | Delete a task |
| `GET` | `/auth/keys` | List API keys |
| `POST` | `/auth/keys` | Create an API key for an existing user |
| `DELETE` | `/auth/keys/{id}` | Delete an API key |

Create a project:

```bash
curl -X POST http://localhost:8080/projects \
  -H "X-API-Key: <key>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Website","description":"Website work"}'
```

Create a task:

```bash
curl -X POST http://localhost:8080/tasks \
  -H "X-API-Key: <key>" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"<project-id>","title":"Ship homepage","tags":["frontend"]}'
```

Move a task:

```bash
curl -X PATCH http://localhost:8080/tasks/<task-id>/move \
  -H "X-API-Key: <key>" \
  -H "Content-Type: application/json" \
  -d '{"status":"in-progress"}'
```

## Architecture

TaskAgent is a small monolithic service:

- `cmd/server`: server entrypoint and `seed` subcommand
- `internal/config`: environment-based configuration
- `internal/handler`: HTTP handlers for auth, projects, and tasks
- `internal/middleware`: API key auth, CORS, request IDs, and request logging
- `internal/service`: project and task business logic
- `internal/store`: SQLite persistence
- `migrations`: embedded database schema migrations
- `cli`: Python CLI package

The server opens a SQLite database, runs migrations, wires stores into services and handlers, and serves JSON over `net/http`. Docker Compose mounts `./data` into the container so the SQLite database survives container restarts.

## Deployment

The default deployment target is a single Docker container backed by SQLite.

```bash
make docker-build
make compose-up
SEED_USER=admin LABEL=bootstrap make compose-run-seed
```

For production-style VPS deployments:

1. Build and tag the Docker image.
2. Push it to an OCI registry.
3. Run it on a host with persistent storage mounted at `/data`.
4. Set `TASKAGENT_DB_PATH=/data/taskagent.db`.
5. Bootstrap the first key with the `seed` subcommand.
6. Put TLS and any public routing in front of the service with your preferred reverse proxy.

See [docs/0.1.0/deployment.md](docs/0.1.0/deployment.md) for the longer VPS deployment guide.

## Contributing

Clone the repository and run the test suite before opening a pull request:

```bash
make test
make lint
make -C cli check
```

For CLI changes, install development dependencies:

```bash
cd cli
pip install -e ".[dev]"
make check
```

Contribution guidelines:

- Keep changes focused and covered by tests.
- Use the existing Go package boundaries for server changes.
- Use the existing Click command style for CLI changes.
- Update documentation when behavior, commands, or configuration changes.

## License

TaskAgent is released under the [MIT License](LICENSE).
