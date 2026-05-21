# VPS Deployment Guide

## Overview

TaskAgent runs as a single Docker container backed by SQLite. This guide documents
a provider-neutral deployment flow for any VPS or cloud host. It covers building,
pushing, provisioning a host, starting the service, bootstrapping access, and
maintaining the running instance.

> **Scope:** This guide covers self-hosting with Docker and Docker Compose only.
> Platform-specific deployments (Fly.io, Railway, Render, Kubernetes, Nomad, etc.)
> are intentionally out of scope for this version.

---

## Quick Start

For a minimal deployment, follow these steps:

```bash
# 1. Build and push the Docker image
make docker-build
docker tag taskagent:local <registry.example.com/your-org/taskagent:latest>
docker push <registry.example.com/your-org/taskagent:latest>

# 2. Provision the VPS and prepare data directory
ssh user@your-vps
make data-dir  # Ensures proper ownership for non-root user

# 3. Start the service
make compose-up

# 4. Bootstrap the first API key
docker compose run --rm taskagent seed --user admin --label bootstrap

# 5. Verify health
curl -s http://localhost:8080/health
```

---

## Prerequisites

Before you begin, ensure your local machine and target VPS have the following:

| Requirement | Minimum Version | Notes |
|---|---|---|
| **Docker Engine** | 24+ | Required on both build host and VPS |
| **Docker Compose plugin** | v2+ | Bundled with Docker Desktop; install via `docker compose` |
| **OCI registry account** | — | Push access needed if deploying a remote image (e.g. Docker Hub, GitHub Container Registry, private registry) |
| **Firewall access** | — | Inbound TCP to the chosen listen port (default `8080`) |
| **SSH access** | — | To the VPS for initial provisioning |

---

## 1. Build the Docker Image

On your local machine (or CI build runner):

```bash
make docker-build
```

This produces a multi-stage build output: a static Go binary in a minimal
Alpine runtime image running as non-root user `appuser` (uid 100 / gid 101).

---

## 2. Push to an OCI Registry

Tag the image with your registry URL and push it. Replace placeholders with your
own registry credentials and image name.

```bash
# Tag for a generic registry
docker tag taskagent:local registry.example.com/your-org/taskagent:latest

# Push
docker push registry.example.com/your-org/taskagent:latest
```

Common registry options:

- **Docker Hub:** `docker.io/<username>/taskagent:latest`
- **GitHub Container Registry:** `ghcr.io/<owner>/taskagent:latest`
- **GitLab Container Registry:** `registry.gitlab.com/<project>/taskagent:latest`
- **Private registry:** your own `harbor` or `registry` instance

> **Tip:** Pin to a specific tag or SHA digest in production rather than `latest`.

---

## 3. Provision the VPS

1. **SSH into your VPS** and ensure Docker is installed:

   ```bash
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER
   ```

2. **Clone or copy the repository** (or just the necessary deployment files):

   ```bash
   # Option A — clone the full repo
   git clone https://github.com/your-org/taskagent.git
   cd taskagent

   # Option B — copy only what's needed (Dockerfile, docker-compose.yml, Makefile)
   # Use rsync, scp, or a CI artifact
   ```

3. **Prepare the data directory** with proper ownership for the container's non-root user:

   ```bash
   make data-dir
   ```

   This ensures the `./data` directory has the correct ownership (uid 100 / gid 101) for the
   container's `appuser`, preventing "unable to open database file" errors.

4. **Verify the image pulls correctly** (if using a remote image):

   ```bash
   docker pull registry.example.com/your-org/taskagent:latest
   ```

---

## 4. Start the Service

Start the container using Docker Compose. The compose file expects the image to be
either built locally or pulled from a registry.

```bash
make compose-up
```

This command:

- Runs `make data-dir` to ensure proper data directory ownership
- Pulls the image (if remote) or uses the local image
- Starts the service with all configuration

The service will:

- Bind host port `8080` to container port `8080`
- Mount `./data` to `/data` inside the container for SQLite persistence
- Run as non-root user `appuser`

---

## 5. Configure Environment Variables

TaskAgent reads its configuration from environment variables. The defaults are:

| Variable | Description | Default |
|---|---|---|
| `TASKAGENT_LISTEN_ADDR` | TCP address to listen on | `:8080` |
| `TASKAGENT_DB_PATH` | SQLite database path (inside container) | `/data/taskagent.db` |
| `TASKAGENT_LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | `info` |
| `TASKAGENT_CORS_ORIGINS` | Comma-separated CORS origins | *(empty — all origins allowed)* |

### Setting overrides

**Via `.env` file:** Create a `.env` file in the repository root and add overrides.
Docker Compose automatically loads `.env`:

```bash
# .env
TASKAGENT_LISTEN_ADDR=:8080
TASKAGENT_DB_PATH=/data/taskagent.db
TASKAGENT_LOG_LEVEL=info
TASKAGENT_CORS_ORIGINS=https://app.example.com
```

**Via docker-compose.yml:** Add environment overrides directly in the service definition:

```yaml
services:
  taskagent:
    environment:
      - TASKAGENT_LOG_LEVEL=debug
      - TASKAGENT_CORS_ORIGINS=https://app.example.com
```

Then restart:

```bash
make compose-up
```

---

## 6. Bootstrap the First API Key

Before the service can accept authenticated requests, you must create the first
API key. Use the `seed` subcommand:

```bash
docker compose run --rm taskagent seed --user admin --label bootstrap
```

This command:

- Runs once inside a short-lived container (`--rm`)
- Creates the first API key for the specified user
- Prints the generated key to stdout — **copy it immediately**, it will not be shown again

After bootstrapping, include the key in the `Authorization` header for all API calls:

```bash
curl -H "Authorization: Bearer <your-api-key>" http://localhost:8080/health
```

---

## 7. Verify Health

Confirm the service is running and healthy:

```bash
curl -s http://localhost:8080/health
```

Expected response:

```json
{"status":"healthy"}
```

If the service is not ready yet, wait a few seconds and retry. The container's
built-in healthcheck probes this endpoint every 30 seconds.

---

## 8. Inspect Logs

Stream service logs in real time:

```bash
make compose-logs
```

Or view the last N lines:

```bash
docker compose logs --tail=100 taskagent
```

---

## 9. Restart the Service

After configuration changes or image updates:

```bash
# Pull a new image and restart
make compose-up

# Or restart in place (preserves data)
docker compose restart
```

To stop completely:

```bash
make compose-down
```

To stop and remove everything including the container (data in `./data` is preserved
because it is a bind mount):

```bash
docker compose down --remove-orphans
```

---

## 10. Backup Considerations

The SQLite database file lives at `./data/taskagent.db` on the host. Back this up
regularly:

### Simple file-level backup

```bash
# Stop the service, copy the database, restart
make compose-down
cp ./data/taskagent.db ./data/taskagent.db.backup.$(date +%Y%m%d%H%M%S)
make compose-up
```

### Online backup (no downtime)

Use SQLite's online backup API or the `sqlite3` `.backup` command on the host
while the container runs. Docker's volume mount ensures the file is accessible:

```bash
# Copy while running (may capture an in-consistent state for very large DBs)
cp ./data/taskagent.db ./data/taskagent.db.backup.$(date +%Y%m%d%H%M%S)
```

For zero-downtime online backups, consider:

- Setting up a WAL (Write-Ahead Logging) mode in SQLite
- Using a replication strategy if the database grows beyond single-host limits

### Disaster recovery

Restore from backup:

```bash
make compose-down
cp ./data/taskagent.db.backup.20260101120000 ./data/taskagent.db
make compose-up
```

---

## 11. Makefile Commands Reference

| Command | Purpose |
|---|---|
| `make docker-build` | Build the Docker image locally |
| `make data-dir` | Prepare data directory with proper ownership (uid 100:gid 101) |
| `make compose-up` | Pull image and start service (includes `data-dir` preparation) |
| `make compose-down` | Stop service (preserves data in `./data`) |
| `make compose-logs` | Stream container logs |
| `docker compose run --rm taskagent seed --user admin --label bootstrap` | Create first API key |
| `curl -s http://localhost:8080/health` | Health check |
| `docker compose restart` | Restart service in place |
| `docker compose down --remove-orphans` | Stop and remove containers (data preserved via bind mount) |

---

## Troubleshooting

### `unable to open database file` on startup

The host-side `./data` directory must be writable by the container's `appuser`.

```bash
make data-dir
make compose-down
make compose-up
```

If the issue persists, fix ownership manually:

```bash
docker run --rm -v $(pwd)/data:/data alpine:3.20 chown 100:101 /data
make compose-up
```

### Container exits immediately

Check logs for configuration errors:

```bash
make compose-logs
```

Ensure the image tag in `docker-compose.yml` matches the image you pushed or built.

### Port already in use

Change the listen address in environment variables:

```bash
# In .env or docker-compose.yml
TASKAGENT_LISTEN_ADDR=:9090
```

Then update the `ports` mapping in `docker-compose.yml`:

```yaml
ports:
  - "9090:8080"
```

Restart:

```bash
make compose-up
```

### High disk usage from logs

Rotate container logs by configuring the Docker daemon or using a log driver:

```yaml
services:
  taskagent:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

---

## Quick Reference

| Command | Purpose |
|---|---|
| `make docker-build` | Build the Docker image locally |
| `docker push registry.example.com/org/taskagent:latest` | Push to OCI registry |
| `make compose-up` | Pull image and start service |
| `docker compose run --rm taskagent seed --user admin --label bootstrap` | Create first API key |
| `curl -s http://localhost:8080/health` | Health check |
| `make compose-logs` | Stream logs |
| `make compose-down` | Stop service (data preserved) |
| `docker compose restart` | Restart service in place |
| `docker compose down --remove-orphans` | Stop and remove containers |
| `cp ./data/taskagent.db ./data/taskagent.db.backup` | Backup database |
| `make data-dir` | Prepare data directory with proper ownership |

