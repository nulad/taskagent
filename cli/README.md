# TaskAgent CLI

Python CLI for TaskAgent - an AI-agent-native task tracker.

## Installation

### From Local Wheel (Development)

1. Build the package:
   ```bash
   cd cli
   make build
   ```

2. Install using pipx (recommended for isolated environments):
   ```bash
   pipx install ./dist/taskagent_cli-0.2.0-py3-none-any.whl
   ```

   Or with pip:
   ```bash
   pip install ./dist/taskagent_cli-0.2.0-py3-none-any.whl
   ```

### Development Install

For development with live code changes:

```bash
cd cli
pip install -e .
```

To also install development dependencies:

```bash
pip install -e ".[dev]"
```

## Usage

```bash
task --help                              # Show help
task --version                           # Show version
task login --server <url> --key <key>    # Configure authentication
task whoami                              # Verify credentials
task projects list                       # List projects
task add --title "Task" --project <id>   # Add a task
task list                                # List tasks
task move <id> <status>                  # Change task status
```

## Shell Completion

The CLI supports static shell completion for Bash, Zsh, and Fish. Completion is static in v0.2.0 (commands and flags only; dynamic task/project names are not completed).

### Bash

Add to your `~/.bashrc`:

```bash
eval "$(task completion bash)"
```

### Zsh

Add to your `~/.zshrc`:

```zsh
eval "$(task completion zsh)"
```

### Fish

Add to `~/.config/fish/completions/task.fish`:

```fish
task completion fish | source
```

## Development

Run tests, linting, and type checks:

```bash
cd cli
make check
```

Individual commands:

```bash
make lint          # Run ruff linter
make format-check  # Check code formatting
make typecheck     # Run mypy type checker
make test          # Run pytest
make build         # Build wheel and source distribution
```
