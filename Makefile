.PHONY: build run test lint docker-build data-dir compose-up compose-down compose-logs compose-smoke

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

docker-build:
	docker build -t taskagent:local .

# Preflight: prepare ./data bind mount with ownership matching the container's
# non-root appuser.  This prevents "unable to open database file" errors caused
# by Docker Compose creating the host-side bind directory as root while the
# container runtime user (appuser) needs write access to SQLite at /data.
# Uses the taskagent image itself to chown via Docker (no sudo needed).
data-dir:
	@mkdir -p ./data; \
	CONTAINER_UID=$$(docker run --rm --entrypoint id taskagent:local -u appuser 2>/dev/null) || CONTAINER_UID=1000; \
	CONTAINER_GID=$$(docker run --rm --entrypoint id taskagent:local -g appuser 2>/dev/null) || CONTAINER_GID=1000; \
	echo ">>> Setting ./data ownership to uid=$${CONTAINER_UID} gid=$${CONTAINER_GID}"; \
	docker run --rm -v $$(pwd)/data:/data alpine:3.20 sh -c "chown $${CONTAINER_UID}:$${CONTAINER_GID} /data" 2>/dev/null || \
	docker run --rm -v $$(pwd)/data:/data alpine:3.20 chown $${CONTAINER_UID}:$${CONTAINER_GID} /data; \
	echo ">>> ./data bind mount prepared successfully"

compose-up: data-dir
	docker compose up --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f taskagent

compose-smoke: docker-build data-dir
	@echo "Running TASK-005 container API round-trip smoke test..."
	@bash ./scripts/container-smoke.sh .
