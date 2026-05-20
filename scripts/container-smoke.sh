#!/usr/bin/env bash
# container-smoke.sh — TASK-005 & TASK-006: Container API Round Trip + SQLite Persistence
#
# Usage:
#   ./scripts/container-smoke.sh [compose-dir]
#
# Prerequisites:
#   - Docker / Docker Compose available
#   - taskagent:local image built (make docker-build)
#
# This script performs a full round-trip verification (TASK-005) followed by
# a persistence verification (TASK-006):
#
#   TASK-005 — Container API Round Trip:
#     1. Starts a clean Compose environment with an empty data directory.
#     2. Seeds an API key inside the container.
#     3. Captures the generated raw API key from the seed command output.
#     4. Starts (or restarts) the main Compose service.
#     5. Polls http://localhost:8080/health until it returns 200.
#     6. Sends a POST /projects request with the seeded X-API-Key header.
#     7. Asserts the response status is 201 and body contains a valid id and name.
#
#   TASK-006 — Compose SQLite Persistence:
#     8. Confirms ./data/taskagent.db exists on the host after the API write.
#     9. Restarts the service with docker compose restart taskagent.
#    10. Polls GET /health until the restarted service is ready.
#    11. Sends GET /projects/{id} with the same X-API-Key.
#    12. Asserts the response status is 200 and body contains the same project id and name.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

COMPOSE_DIR="${1:-$PROJECT_DIR}"

# ── Colour helpers ──────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Colour

ok()  { printf "${GREEN}✓${NC} %s\n" "$*"; }
warn() { printf "${YELLOW}⚠ %s\n" "$*"; }
fail() { printf "${RED}✗ %s\n" "$*"; exit 1; }

# ── Helpers ─────────────────────────────────────────────────────────────────
# Wait up to $1 seconds for the health endpoint to return 200.
wait_for_health() {
  local timeout="${1:-60}"
  local elapsed=0
  local interval=2

  printf "Waiting for service to become healthy (up to ${timeout}s)..."
  while [ "$elapsed" -lt "$timeout" ]; do
    if http_code=$(curl -s -o /dev/null -w '%{http_code}' \
        "http://localhost:8080/health" 2>/dev/null); then
      if [ "$http_code" = "200" ]; then
        echo ""
        ok "Health check passed (HTTP $http_code)"
        return 0
      fi
    fi
    sleep "$interval"
    elapsed=$((elapsed + interval))
  done
  fail "Service did not become healthy within ${timeout}s"
}

# Parse a simple JSON string by key using grep/sed (no jq required).
# Usage: json_val '{"id":"abc"}' id   → prints abc
json_val() {
  local json="$1" key="$2"
  printf '%s' "$json" | grep -oP "\"${key}\":\"\\K[^\"]*" || true
}

# ── Main ────────────────────────────────────────────────────────────────────

cd "$COMPOSE_DIR"

# 1. Clean data directory for a fresh start
warn "Cleaning data directory..."
rm -rf "$COMPOSE_DIR/data"
mkdir -p "$COMPOSE_DIR/data"
chmod 777 "$COMPOSE_DIR/data"

# 2. Tear down any stale Compose environment
warn "Tearing down any stale containers..."
docker compose -f "$COMPOSE_DIR/docker-compose.yml" down --remove-orphans 2>/dev/null || true

# 3. Seed an API key inside a temporary container (one-off, uses same DB volume)
warn "Seeding API key in the container..."
seed_output=$(docker compose -f "$COMPOSE_DIR/docker-compose.yml" run --rm taskagent seed \
  --user admin --label bootstrap 2>&1)

raw_api_key=$(echo "$seed_output" | grep -oP 'key":\s*"\K[^"]+' || true)

if [ -z "$raw_api_key" ]; then
  fail "Seed command did not output an API key. Output was:\n$seed_output"
fi
ok "API key seeded successfully"

# 4. Start (or restart) the main Compose service
warn "Starting Compose service..."
docker compose -f "$COMPOSE_DIR/docker-compose.yml" up -d
ok "Compose service started"

# 5. Poll /health until 200
wait_for_health 60

# 6-8. Create a project via the API using the seeded key
warn "Sending POST /projects with seeded API key..."
project_response=$(curl -s -w '\n%{http_code}' \
  -X POST "http://localhost:8080/projects" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $raw_api_key" \
  -d '{"name":"container-smoke"}')

# Extract the status code (last line) and body (everything before the last line)
project_status=$(echo "$project_response" | tail -1)
project_body=$(echo "$project_response" | sed '$d')

# 7. Assert status is 201
if [ "$project_status" != "201" ]; then
  fail "Expected HTTP 201, got $project_status. Body: $project_body"
fi
ok "Project creation returned HTTP $project_status"

# 8. Assert response contains id and name
project_id=$(json_val "$project_body" "id")
project_name=$(json_val "$project_body" "name")

if [ -z "$project_id" ]; then
  fail "Response body missing non-empty 'id' field. Body: $project_body"
fi
if [ "$project_name" != "container-smoke" ]; then
  fail "Response body 'name' mismatch: got '$project_name', expected 'container-smoke'. Body: $project_body"
fi

ok "Response contains id='$project_id' and name='$project_name'"

# ═════════════════════════════════════════════════════════════════════════════
# TASK-006: Verify Compose SQLite Persistence
# ═════════════════════════════════════════════════════════════════════════════

# 9. Confirm ./data/taskagent.db exists on the host after the API write
warn "Verifying host-side SQLite database file..."
db_path="$COMPOSE_DIR/data/taskagent.db"
if [ ! -f "$db_path" ]; then
  fail "Host database file does not exist at $db_path. Persistence bind mount may be misconfigured."
fi
db_size=$(stat -c%s "$db_path" 2>/dev/null || stat -f%z "$db_path" 2>/dev/null || echo "unknown")
ok "Host database file exists at $db_path ($db_size bytes)"

# 10. Restart the service with docker compose restart taskagent
warn "Restarting Compose service to test data persistence..."
docker compose -f "$COMPOSE_DIR/docker-compose.yml" restart taskagent
ok "Compose service restarted"

# 11. Poll GET /health until the restarted service is ready
wait_for_health 60

# 12. Send GET /projects/{id} with the same X-API-Key
warn "Fetching project $project_id after restart..."
get_response=$(curl -s -w '\n%{http_code}' \
  -X GET "http://localhost:8080/projects/$project_id" \
  -H "X-API-Key: $raw_api_key")

get_status=$(echo "$get_response" | tail -1)
get_body=$(echo "$get_response" | sed '$d')

# 13. Assert the response status is 200
if [ "$get_status" != "200" ]; then
  fail "Expected HTTP 200 after restart, got $get_status. Body: $get_body"
fi
ok "Project retrieval returned HTTP $get_status"

# 14. Assert the response body still contains the same project ID and name
get_id=$(json_val "$get_body" "id")
get_name=$(json_val "$get_body" "name")

if [ "$get_id" != "$project_id" ]; then
  fail "Project ID mismatch after restart: got '$get_id', expected '$project_id'. Body: $get_body"
fi
if [ "$get_name" != "$project_name" ]; then
  fail "Project name mismatch after restart: got '$get_name', expected '$project_name'. Body: $get_body"
fi

ok "Persistence verified — project id='$get_id' and name='$get_name' survived restart"

# ── Summary ─────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════"
printf "${GREEN}TASK-005 SMOKE TEST PASSED${NC}\n"
echo "═══════════════════════════════════════════"
echo ""
echo "  Seeded API key label: bootstrap"
echo "  Project ID: $project_id"
echo "  Project name: $project_name"
echo ""
echo "═══════════════════════════════════════════"
printf "${GREEN}TASK-006 PERSISTENCE TEST PASSED${NC}\n"
echo "═══════════════════════════════════════════"
echo ""
echo "  Host database file: $db_path ($db_size bytes)"
echo "  Project data survived Compose restart"
echo "  Compose bind mount confirmed as the persistence boundary"
echo ""
echo "  Compose service is still running at localhost:8080"
echo "  To inspect logs: docker compose -f $COMPOSE_DIR/docker-compose.yml logs -f"
echo "  To stop: docker compose -f $COMPOSE_DIR/docker-compose.yml down"
echo ""
