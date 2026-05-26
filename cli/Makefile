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
	@if [ -n "$$(find lib -name '*.sh' -type f 2>/dev/null)" ]; then \
		echo "$$(find lib -name '*.sh' -type f 2>/dev/null)" | while IFS= read -r script; do \
			echo "Checking: $$script"; \
			shellcheck -s sh "$$script"; \
		done; \
	fi
	@if [ -n "$$(find test -name '*.sh' -type f 2>/dev/null)" ]; then \
		echo "$$(find test -name '*.sh' -type f 2>/dev/null)" | while IFS= read -r script; do \
			echo "Checking: $$script"; \
			shellcheck -s sh "$$script"; \
		done; \
	fi
	@echo "ShellCheck completed successfully."
