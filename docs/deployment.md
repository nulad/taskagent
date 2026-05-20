# Deployment Guide

## Overview

The taskagent service is containerised with Docker / Docker Compose.
A single `taskagent` service runs a Go binary backed by SQLite.
Data files are persisted via a host bind mount (`./data:/data`).

## Prerequisites

- Docker Engine 24+
- Docker Compose v2+
- Go 1.25+ (only needed for local builds without Docker)

## Quick Start

### 1. Build the image

```bash
make docker-build          # builds taskagent:local
```

### 2. Prepare the data directory

The bind mount at `./data:/data` must be owned by the container's non-root
`appuser` (uid 100 / gid 101 in the image) for SQLite to function.
Docker Compose creates bind-mount directories as `root` on the host, so the
application would fail with `unable to open database file` without permission setup.

The `make compose-up` target runs the preflight automatically:

```bash
make data-dir              # manual preflight (also run by compose-up)
make compose-up            # builds + starts with data-dir preflight
```

The preflight does two things:

1. Creates `./data` if it does not already exist.
2. Sets ownership to match the `appuser` uid/gid extracted from the image,
   using a lightweight Docker container to perform `chown` (no `sudo` needed).

### 3. Verify

```bash
curl -s http://localhost:8080/health
```

Expected response: `{"status":"healthy"}` (or similar, depending on TASK-005
health-check contract).

## Makefile Targets

| Target          | Description                                        |
|-----------------|----------------------------------------------------|
| `build`         | Compile the Go binary (`go build ./...`)           |
| `run`           | Run locally (`go run ./cmd/server`)                |
| `test`          | Run unit tests (`go test ./...`)                   |
| `lint`          | Static analysis (`go vet ./...`)                   |
| `docker-build`  | Build the Docker image (`taskagent:local`)         |
| `data-dir`      | **Preflight** — prepare `./data` with correct perm.|
| `compose-up`    | Build + start service (runs `data-dir` first)      |
| `compose-down`  | Stop and remove containers / networks              |
| `compose-logs`  | Stream logs (`docker compose logs -f taskagent`)   |
| `compose-smoke` | Smoke-verification helper (delegates to TASK-005)  |

## Environment Variables

| Variable              | Description                            | Default       |
|-----------------------|----------------------------------------|---------------|
| `TASKAGENT_LISTEN_ADDR`  | TCP address to listen on            | `:8080`       |
| `TASKAGENT_DB_PATH`      | SQLite database path (inside container) | `/data/taskagent.db` |
| `TASKAGENT_LOG_LEVEL`    | Logging level (`debug`, `info`, `warn`, `error`) | `info` |
| `TASKAGENT_CORS_ORIGINS` | Comma-separated CORS origins       | *(empty — all allowed)* |

Pass overrides via a `.env` file or inline:

```bash
TASKAGENT_LOG_LEVEL=debug make compose-up
```

## Docker Compose Contract

```yaml
services:
  taskagent:
    build:
      context: .
      dockerfile: Dockerfile
    image: taskagent:local
    ports:
      - "8080:8080"
    environment:
      - TASKAGENT_LISTEN_ADDR=:8080
      - TASKAGENT_DB_PATH=/data/taskagent.db
      - TASKAGENT_LOG_LEVEL=info
      - TASKAGENT_CORS_ORIGINS=${TASKAGENT_CORS_ORIGINS:-}
    volumes:
      - ./data:/data
    healthcheck:              # defined in Dockerfile
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      start_period: 5s
      retries: 3
```

## Troubleshooting

### `unable to open database file` on startup

**Cause:** The host-side `./data` directory is owned by a different user than
the container's `appuser` (uid 100 / gid 101). SQLite cannot create/open the database.

**Fix:**

```bash
make data-dir              # re-apply correct ownership
make compose-down          # stop stale container
make compose-up            # restart
```

Or manually via Docker:

```bash
# Find the uid/gid used by the container:
docker run --rm --entrypoint id taskagent:local appuser
# e.g. uid=100(appuser) gid=101(appgroup)

# Apply to host via Docker (no sudo needed):
docker run --rm -v $(pwd)/data:/data alpine:3.20 chown 100:101 /data
```

### `./data` contains stale data from a previous run

Delete and re-initialise:

```bash
rm -rf ./data/*
make compose-down
make compose-up
```

### Docker Compose fails with permission errors

Ensure the current user has Docker access:

```bash
# Add user to docker group (Linux, one-time):
sudo usermod -aG docker $USER
newgrp docker
```

## Non-Root Runtime

The container runs as a dedicated non-root user (`appuser:appgroup`, uid 100 / gid 101).
The `./data` bind mount must be writable by this user for SQLite to function.

The `data-dir` preflight uses a lightweight Alpine container to perform `chown` on
`./data`, so no `sudo` is required on the host — Docker handles it via its socket
permission model.
