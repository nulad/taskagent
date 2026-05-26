# common.sh - Common shell helper functions
# POSIX sh compatible - no Bash extensions, no [[ ... ]], no arrays, no source
#
# This file is designed to be sourced by other scripts, not executed directly.
# It provides utility functions for logging, error handling, and validation.
#
# POSIX sh compatible - no Bash extensions, no [[ ... ]], no arrays, no source
#

# ============================================================================
# die() - Print error message to stderr and exit
# ============================================================================
# Usage: die [exit_code] "message"
# Example: die "Failed to connect"
#          die 2 "Connection failed with timeout"
#
die() {
    _die_exit_code="1"
    _die_message=""

    # Check if first argument is numeric (exit code) or not (message)
    if [ "$1" -eq "$1" ] 2>/dev/null; then
        # First argument is numeric, treat as exit code
        _die_exit_code="${1:-1}"
        _die_message="${2:-}"
    else
        # First argument is not numeric, treat as message
        _die_message="${1:-}"
    fi

    if [ -n "$_die_message" ]; then
        printf '%s\n' "$_die_message" >&2
    fi
    exit "$_die_exit_code"
}

# ============================================================================
# log() - Print human-readable message to stderr
# ============================================================================
# Usage: log "message"
# Example: log "Starting installation..."
#          log "Config file: $CONFIG_FILE"
#
log() {
    _log_message="${1:-}"
    if [ -n "$_log_message" ]; then
        printf '%s\n' "$_log_message" >&2
    fi
}

# ============================================================================
# require_cmd() - Verify required commands are available
# ============================================================================
# Usage: require_cmd cmd1 [cmd2] [cmd3] ...
# Example: require_cmd curl jq
#          require_cmd grep sed awk
#
# Fails with an error message if any specified command is not found.
#
require_cmd() {
    _req_cmd_count="${#@}"
    _missing_cmds=""

    for _cmd in "$@"; do
        if ! command -v "$_cmd" >/dev/null 2>&1; then
            if [ -n "$_missing_cmds" ]; then
                _missing_cmds="$_missing_cmds, $_cmd"
            else
                _missing_cmds="$_cmd"
            fi
        fi
    done

    if [ -n "$_missing_cmds" ]; then
        die "missing required command: $_missing_cmds"
    fi
}

# ============================================================================
# enable_pipefail() - Enable pipefail if supported by the shell
# ============================================================================
# pipefail causes a pipeline to return the exit status of the last command
# to exit with a non-zero status (rather than zero) when any command in
# the pipeline fails.
#
# This helper is best-effort and only enables pipefail if the shell supports
# it (e.g., most modern sh implementations and bash). It does not fail if
# pipefail is not available.
#
enable_pipefail() {
    # Enable pipefail if supported by the shell (bash 4+, dash 0.5.5+, etc.)
    # pipefail causes a pipeline to return the exit status of the last command
    # that exited with a non-zero status, rather than zero.
    # This is best-effort; if the shell doesn't support pipefail, this does nothing.
    # Note: In POSIX sh (e.g., dash), this may produce a warning but does not fail.
    (set -o pipefail 2>/dev/null || true) || true
}

# ============================================================================
# Usage example (for reference, not part of the library):
#
#   . cli/lib/common.sh
#
#   require_cmd curl jq
#   enable_pipefail
#
#   log "Starting download..."
#   if command -v curl >/dev/null; then
#       curl -s https://example.com | jq .
#       log "Download completed"
#   fi
#
#   die "Download failed"
#

# ============================================================================
# Note: enable_pipefail() uses a subshell to avoid issues with POSIX sh
# (e.g., dash) that don't support pipefail. The function may produce a
# warning but will not fail the script. To ensure pipefail is enabled,
# use bash explicitly: `#!/bin/bash` instead of `#!/bin/sh`.
# The double `|| true` ensures the function never fails, even when
# pipefail is not supported in POSIX sh.
# ============================================================================
