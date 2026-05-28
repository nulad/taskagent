# common.sh - Common shell helper functions
# shellcheck shell=sh disable=SC3043
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
# Config Loading Precedence
#
# Precedence order (highest to lowest):
#   1. Explicit CLI flags (set by command-specific argument parsers)
#   2. Environment variables (TASKAGENT_SERVER, TASKAGENT_API_KEY, TASKAGENT_TIMEOUT)
#   3. Config file ($TASKAGENT_HOME/config)
#   4. Hardcoded defaults (TASKAGENT_TIMEOUT=10)
#
# Usage:
#   . cli/lib/common.sh
#   load_config          # reads config file if present
#   require_auth_config  # exits with error if server/key not configured
#
# Command-specific flag overrides (set after load_config):
#   TASKAGENT_SERVER="http://example.com"  # env or explicit flag override
#   TASKAGENT_API_KEY="sk-xxx"             # env or explicit flag override
#


# --------------------------------------------------------------------------
# default_taskagent_home() - Resolve the base configuration directory
# --------------------------------------------------------------------------
# Returns the path to the TaskAgent home directory.
# Respects TASKAGENT_HOME env var, falls back to $HOME/.taskagent.
#
# shellcheck disable=SC2016
# Usage: TASKAGENT_HOME="$(default_taskagent_home)"
# --------------------------------------------------------------------------
default_taskagent_home() {
    echo "${TASKAGENT_HOME:-${HOME:-$HOME}/.taskagent}"
}

# --------------------------------------------------------------------------
# load_config() - Load configuration from file
# --------------------------------------------------------------------------
# Reads $TASKAGENT_HOME/config if it exists and sources it.
# The config file must contain KEY=VALUE lines (one per line).
# Supported keys: TASKAGENT_SERVER, TASKAGENT_API_KEY, TASKAGENT_TIMEOUT.
#
# Config file values are only applied when the corresponding environment
# variable is NOT already set — this preserves env-var precedence.
# --------------------------------------------------------------------------
load_config() {
    local _config_dir
    local _config_file

    _config_dir="$(default_taskagent_home)"
    _config_file="${_config_dir}/config"

    # Only source the config file if it exists and is readable
    if [ -r "$_config_file" ]; then
        # Save current env values so we can restore them after sourcing
        # This allows the config file to set defaults without overriding env
        local _saved_server
        local _saved_key
        local _saved_timeout
        _saved_server="${TASKAGENT_SERVER:-}"
        _saved_key="${TASKAGENT_API_KEY:-}"
        _saved_timeout="${TASKAGENT_TIMEOUT:-}"

        # Source the config file — it will set TASKAGENT_* variables
        # shellcheck disable=SC1090
        . "$_config_file"

        # Restore env vars if they were set before sourcing (env > file)
        # If TASKAGENT_SERVER was set before load_config(), keep the env value
        if [ -n "$_saved_server" ]; then
            TASKAGENT_SERVER="$_saved_server"
        fi
        if [ -n "$_saved_key" ]; then
            TASKAGENT_API_KEY="$_saved_key"
        fi
        if [ -n "$_saved_timeout" ]; then
            TASKAGENT_TIMEOUT="$_saved_timeout"
        fi
    fi
}

# --------------------------------------------------------------------------
# require_auth_config() - Validate that auth configuration is present
# --------------------------------------------------------------------------
# Exits with code 2 if TASKAGENT_SERVER or TASKAGENT_API_KEY is unset or
# empty after config loading. Prints a helpful error directing the user to
# run "task login" to set up credentials.
#
# Call this at the start of any authenticated subcommand.
# --------------------------------------------------------------------------
require_auth_config() {
    local _missing=""

    if [ -z "${TASKAGENT_SERVER:-}" ]; then
        _missing="${_missing:+${_missing}, }TASKAGENT_SERVER"
    fi
    if [ -z "${TASKAGENT_API_KEY:-}" ]; then
        _missing="${_missing:+${_missing}, }TASKAGENT_API_KEY"
    fi

    if [ -n "$_missing" ]; then
        die 2 "
missing required configuration:$_missing

Configure ${_missing} via environment variable, config file, or 'task login'.
"
    fi
}

# --------------------------------------------------------------------------
# Config defaults — applied after env and file values
# --------------------------------------------------------------------------
# These defaults are only applied if the variable is still unset after
# load_config() has run and no environment override exists.
#
# TASKAGENT_TIMEOUT defaults to 10 seconds.
# TASKAGENT_SERVER and TASKAGENT_API_KEY have no defaults — they must be
# provided by the user via config file or environment.
#
# Usage: set_config_defaults
# --------------------------------------------------------------------------
set_config_defaults() {
    : "${TASKAGENT_TIMEOUT:=10}"
}

# ============================================================================
# Usage example (for reference, not part of the library):
#
#   . cli/lib/common.sh
#
#   load_config
#   set_config_defaults
#   require_auth_config
#
#   log "Server: $TASKAGENT_SERVER"
#   log "Timeout: $TASKAGENT_TIMEOUT"
#   log "Key: ${TASKAGENT_API_KEY:+***}${TASKAGENT_API_KEY#????}"
#
#   # Later, a command-specific flag parser can override loaded values:
#   #   TASKAGENT_SERVER="http://custom-host:8080"  # CLI flag override
#
# ============================================================================


