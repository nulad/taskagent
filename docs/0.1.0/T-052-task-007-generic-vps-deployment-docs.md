# TASK-007: Write Generic VPS Deployment Docs

## Token Safety Verification

- **Task ID & Title:** TASK-007: Write Generic VPS Deployment Docs
- **Objective:** Document a concise provider-neutral VPS deployment flow for the Docker image and Compose file.
- **Target Files/Scope:** `docs/deployment.md`
- **Token Budget Heuristic:**
  * Estimated Existing Code Context: Low - this task only needs the T-052 deployment contract, Dockerfile image name, and Compose service name.
  * Dependencies: TASK-002 for Compose conventions; TASK-001 for image build/tag conventions.

## Step-by-step Instructions

1. Create `docs/deployment.md`.
2. Include prerequisites: Docker with Compose plugin, a host directory for SQLite data, firewall access to the chosen port, and a registry account if deploying from a pushed image.
3. Document local image build: `docker build -t taskagent:local .`.
4. Document pushing to any OCI registry using placeholder names, not a provider-specific registry.
5. Document copying or checking out the repository on a VPS.
6. Document creating `./data` and starting the service with `docker compose pull && docker compose up -d` when using a remote image.
7. Document the required environment variables and their defaults.
8. Document bootstrapping the first API key with `docker compose run --rm taskagent seed --user admin --label bootstrap`.
9. Document health verification with `curl http://localhost:8080/health`.
10. Add a short note that Fly, Railway, Render, Kubernetes, Nomad, and other provider-specific manifests are intentionally out of scope for T-052.
11. Keep the document operational and concise; avoid broad Docker tutorials.

## Definition of Done

- `docs/deployment.md` explains build, push, host setup, start, seed, health check, logs, restart, and backup considerations.
- The docs are provider-neutral.
- Commands match `docker-compose.yml` and the Makefile targets from T-052.
