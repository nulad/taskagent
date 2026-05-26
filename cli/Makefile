# Makefile for CLI shell scripts
# Usage: make -C cli <target>

.PHONY: check

SHELL := /bin/bash

# Check target: run ShellCheck on all shell scripts
# Checks: task and all *.sh files under lib/ and test/
check:
	@command -v shellcheck >/dev/null 2>&1 || (echo "Error: shellcheck is not installed. Please install it from https://github.com/koalaman/shellcheck or with: sudo apt-get install shellcheck" && exit 1)
	@bash check.sh
