# Makefile for CLI shell scripts
# Usage: make -C cli <target>

.PHONY: check

SHELL := /bin/bash

# Check target: run ShellCheck on all shell scripts
# Checks: task and all *.sh files under lib/ and test/
check:
	@command -v shellcheck >/dev/null 2>&1 || (echo "Error: shellcheck is not installed. Please install it from https://github.com/koalaman/shellcheck or with: sudo apt-get install shellcheck" && exit 1)
	@echo "Running ShellCheck on CLI shell scripts..."
	@if [ -f "task" ]; then \
		echo "Checking: task"; \
		shellcheck -s sh "task"; \
	fi
	@if [ -f "check.sh" ]; then \
		echo "Checking: check.sh"; \
		shellcheck -s sh "check.sh"; \
	fi
	@for f in $$(find lib -name "*.sh" -type f 2>/dev/null); do \
		[ -n "$$f" ] && { echo "Checking: $$f"; shellcheck -s sh "$$f"; } || exit 1; \
	done
	@for f in $$(find test -name "*.sh" -type f 2>/dev/null); do \
		[ -n "$$f" ] && { echo "Checking: $$f"; shellcheck -s sh "$$f"; } || exit 1; \
	done
	@echo "ShellCheck completed successfully."
