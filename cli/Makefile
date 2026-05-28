.PHONY: lint format-check typecheck test check build smoke

lint:
	ruff check src tests

format-check:
	ruff format --check src tests

typecheck:
	mypy src/taskagent_cli

test:
	pytest tests -v

check: lint format-check typecheck test

build:
	python -m build

smoke:
	task --help
	task --version
