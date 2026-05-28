#!/usr/bin/env sh
# test_config.sh — Config loading precedence tests
# shellcheck shell=sh disable=SC2129
#
# Tests the precedence: CLI flag > env > file > default
#
# Usage: ./cli/test/test_config.sh

set -e

PASS=0
FAIL=0

# ------------------------------------------------------------------
# Helpers
# ------------------------------------------------------------------
pass() { PASS=$((PASS + 1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1"; }
assert_eq() {
    _label="$1"; _expected="$2"; _actual="$3"
    if [ "$_expected" = "$_actual" ]; then
        pass "$_label"
    else
        fail "$_label (expected '$_expected', got '$_actual')"
    fi
}

run_test() {
    # Runs the provided test script in a subprocess with the given env setup.
    # Usage: run_test "TASKAGENT_HOME=/some/path" 'cat > /some/path/config <<...; . script; load_config; echo ...'
    _test_env="$1"
    _test_body="$2"

    sh -c "$_test_env; $_test_body"
}

# ------------------------------------------------------------------
# Setup: create a temporary TASKAGENT_HOME
# ------------------------------------------------------------------
_TMPHOME=$(mktemp -d)
export TASKAGENT_HOME="$_TMPHOME"
echo "Temp dir: $_TMPHOME"

# ------------------------------------------------------------------
# Test 1: default_taskagent_home() respects env and falls back
# ------------------------------------------------------------------
echo ""
echo "Test 1: default_taskagent_home()"

# With env set
assert_eq "env override" "$_TMPHOME" "$(sh -c '. cli/lib/common.sh; default_taskagent_home')"

# Without env set (default)
assert_eq "default to ~/.taskagent" "$HOME/.taskagent" "$(sh -c 'unset TASKAGENT_HOME 2>/dev/null; . cli/lib/common.sh; default_taskagent_home')"

# ------------------------------------------------------------------
# Test 2: config file is loaded when env vars are absent
# ------------------------------------------------------------------
echo ""
echo "Test 2: Config file values are loaded when env vars are absent"

# Create config file
printf '%s\n' 'TASKAGENT_SERVER=http://from-file:8080' > "$_TMPHOME/config"
printf '%s\n' 'TASKAGENT_API_KEY=sk-from-file' >> "$_TMPHOME/config"
printf '%s\n' 'TASKAGENT_TIMEOUT=30' >> "$_TMPHOME/config"

result=$(sh -c '. cli/lib/common.sh; load_config; echo "S=$TASKAGENT_SERVER"; echo "K=$TASKAGENT_API_KEY"; echo "T=$TASKAGENT_TIMEOUT"' 2>&1)

assert_eq "file SERVER" "S=http://from-file:8080" "$(echo "$result" | head -1)"
assert_eq "file API_KEY" "K=sk-from-file" "$(echo "$result" | sed -n '2p')"
assert_eq "file TIMEOUT" "T=30" "$(echo "$result" | sed -n '3p')"

# ------------------------------------------------------------------
# Test 3: env vars override config file values
# ------------------------------------------------------------------
echo ""
echo "Test 3: Environment variables override config file values"

result=$(sh -c '
    TASKAGENT_SERVER=http://from-env:9090
    TASKAGENT_API_KEY=sk-from-env
    TASKAGENT_TIMEOUT=99
    . cli/lib/common.sh
    load_config
    echo "S=$TASKAGENT_SERVER"
    echo "K=$TASKAGENT_API_KEY"
    echo "T=$TASKAGENT_TIMEOUT"
' 2>&1)

assert_eq "env SERVER overrides file" "S=http://from-env:9090" "$(echo "$result" | head -1)"
assert_eq "env API_KEY overrides file" "K=sk-from-env" "$(echo "$result" | sed -n '2p')"
assert_eq "env TIMEOUT overrides file" "T=99" "$(echo "$result" | sed -n '3p')"

# ------------------------------------------------------------------
# Test 4: defaults are applied after env and file values
# ------------------------------------------------------------------
echo ""
echo "Test 4: Defaults are applied after file and env values"

# Default not applied when file has a value
result=$(sh -c '. cli/lib/common.sh; load_config; set_config_defaults; echo "T=$TASKAGENT_TIMEOUT"' 2>&1)
assert_eq "default not applied (file set value)" "T=30" "$(echo "$result" | head -1)"

# Default applied when file is absent
rm -f "$_TMPHOME/config"
result=$(sh -c '. cli/lib/common.sh; load_config; set_config_defaults; echo "T=$TASKAGENT_TIMEOUT"' 2>&1)
assert_eq "default TIMEOUT=10 applied" "T=10" "$(echo "$result" | head -1)"

# Default not applied when env has a value
result=$(sh -c 'TASKAGENT_TIMEOUT=42; . cli/lib/common.sh; load_config; set_config_defaults; echo "T=$TASKAGENT_TIMEOUT"' 2>&1)
assert_eq "default TIMEOUT not applied (env set value)" "T=42" "$(echo "$result" | head -1)"

# ------------------------------------------------------------------
# Test 5: require_auth_config() errors when config is missing
# ------------------------------------------------------------------
echo ""
echo "Test 5: require_auth_config() errors when config is missing"

# Helper: run a test in a temp script to isolate die() exits
_test_auth() {
    _tmpscript=$(mktemp)
    sh -c "
        set +e
        $_1
    " 2>&1
    rm -f "$_tmpscript"
}

# With nothing configured — expect error mentioning both
result=$(sh -c '
    set +e
    unset TASKAGENT_SERVER TASKAGENT_API_KEY TASKAGENT_TIMEOUT 2>/dev/null || true
    . cli/lib/common.sh
    load_config
    set_config_defaults
    require_auth_config 2>&1
') || true

case "$result" in
    *"TASKAGENT_SERVER"*) pass "require_auth_config catches missing SERVER" ;;
    *) fail "require_auth_config should mention TASKAGENT_SERVER" ;;
esac

# With TASKAGENT_API_KEY present — expect error for missing SERVER
result=$(sh -c '
    set +e
    TASKAGENT_API_KEY=sk-some-key
    unset TASKAGENT_SERVER 2>/dev/null || true
    . cli/lib/common.sh
    load_config
    set_config_defaults
    require_auth_config 2>&1
') || true

case "$result" in
    *"TASKAGENT_SERVER"*) pass "require_auth_config catches missing SERVER when API_KEY exists" ;;
    *) fail "require_auth_config should mention TASKAGENT_SERVER" ;;
esac

# With TASKAGENT_SERVER present — expect error for missing API_KEY
result=$(sh -c '
    set +e
    TASKAGENT_SERVER=http://example.com
    unset TASKAGENT_API_KEY 2>/dev/null || true
    . cli/lib/common.sh
    load_config
    set_config_defaults
    require_auth_config 2>&1
') || true

case "$result" in
    *"TASKAGENT_API_KEY"*) pass "require_auth_config catches missing API_KEY when SERVER exists" ;;
    *) fail "require_auth_config should mention TASKAGENT_API_KEY" ;;
esac

# With both set — should NOT error
result=$(sh -c '
    TASKAGENT_SERVER=http://example.com
    TASKAGENT_API_KEY=sk-valid
    . cli/lib/common.sh
    load_config
    set_config_defaults
    require_auth_config
    echo "OK"
' 2>&1)

if [ "$result" = "OK" ]; then
    pass "require_auth_config succeeds when both config values present"
else
    fail "require_auth_config should not error when both config values present (got: $result)"
fi

# ------------------------------------------------------------------
# Test 6: CLI flag override (simulated)
# ------------------------------------------------------------------
echo ""
echo "Test 6: Command-specific flag overrides work after load_config"

printf '%s\n' 'TASKAGENT_SERVER=http://from-file:8080' > "$_TMPHOME/config"
printf '%s\n' 'TASKAGENT_API_KEY=sk-from-file' >> "$_TMPHOME/config"

result=$(sh -c '. cli/lib/common.sh; load_config; set_config_defaults; TASKAGENT_SERVER=http://from-flag:7070; echo "S=$TASKAGENT_SERVER"' 2>&1)
assert_eq "CLI flag override wins" "S=http://from-flag:7070" "$(echo "$result" | head -1)"

# ------------------------------------------------------------------
# Test 7: Config with comments and empty lines
# ------------------------------------------------------------------
echo ""
echo "Test 7: Config file handles comments and empty lines"

printf '%s\n' '# This is a comment' > "$_TMPHOME/config"
printf '%s\n' '' > "$_TMPHOME/config"
printf '%s\n' 'TASKAGENT_SERVER=http://comment-test:8080' >> "$_TMPHOME/config"
printf '%s\n' '# Another comment' >> "$_TMPHOME/config"
printf '%s\n' 'TASKAGENT_API_KEY=sk-comment-test' >> "$_TMPHOME/config"

result=$(sh -c '. cli/lib/common.sh; load_config; echo "S=$TASKAGENT_SERVER"; echo "K=$TASKAGENT_API_KEY"' 2>&1)

assert_eq "comment file SERVER" "S=http://comment-test:8080" "$(echo "$result" | head -1)"
assert_eq "comment file API_KEY" "K=sk-comment-test" "$(echo "$result" | sed -n '2p')"

# ------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------
rm -rf "$_TMPHOME"

# ------------------------------------------------------------------
# Summary
# ------------------------------------------------------------------
echo ""
echo "=============================="
echo "Results: $PASS passed, $FAIL failed"
echo "=============================="

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
exit 0
