#!/usr/bin/env bash
# container-smoke.sh — TASK-005: Verify Container API Round Trip
#
# Usage:
#   ./scripts/container-smoke.sh [compose-dir]
#
# Prerequisites:
#   - Docker / Docker Compose available
#   - taskagent:local image built (make docker-build)
#
# This script performs a full round-trip verification:
#   1. Starts a clean Compose environment with an empty data directory.
#   2. Seeds an API key inside the container.
#   3. Captures the generated raw API key from the seed command output.
#   4. Starts (or restarts) the main Compose service.
#   5. Polls http://localhost:8080/health until it returns 200.
#   6. Sends a POST /projects request with the seeded X-API-Key header.
#   7. Asserts the response status is 201 and body contains a valid id and name.

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
echo "  Compose service is still running at localhost:8080"
echo "  To inspect logs: docker compose -f $COMPOSE_DIR/docker-compose.yml logs -f"
echo "  To stop: docker compose -f $COMPOSE_DIR/docker-compose.yml down"
echo ""
