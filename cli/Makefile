# cli/Makefile — ShellCheck linting for CLI shell scripts

SHELL := /bin/sh

SHELLCHECK := shellcheck

.PHONY: check

# Collect files to check: cli/task (if present) and all *.sh under cli/lib/ and cli/test/
FILES :=
FILES += $(wildcard task)
FILES += $(wildcard lib/*.sh)
FILES += $(wildcard test/*.sh)

check:
	@if ! command -v $(SHELLCHECK) >/dev/null 2>&1; then \
		echo "ERROR: shellcheck is not installed."; \
		echo "Install it before running 'make check'."; \
		exit 1; \
	fi
	@if [ -z "$(FILES)" ]; then \
		echo "No shell scripts found to check."; \
		exit 1; \
	fi
	@$(SHELLCHECK) -s sh $(FILES)
