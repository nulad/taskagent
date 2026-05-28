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
# Fails fast on the first missing command with an error message.
#
require_cmd() {
    for _cmd in "$@"; do
        if ! command -v "$_cmd" >/dev/null 2>&1; then
            die "missing required command: $_cmd"
        fi
    done
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
# Probes support in a subshell, then enables in the current shell only if
# supported. This ensures pipefail is actually enabled when the shell
# supports it.
#
# shellcheck disable=SC3040
enable_pipefail() {
    # Probe support in a subshell first
    if (set -o pipefail) 2>/dev/null; then
        # Shell supports pipefail, enable it in the current shell
        set -o pipefail
    fi
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
# Note: enable_pipefail() probes pipefail support in a subshell to avoid
# issues with POSIX sh (e.g., dash) that don't support it. If supported,
# pipefail is enabled in the current shell. If not supported, the function
# silently does nothing (no error, no warning).
# ============================================================================
