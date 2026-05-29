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
taskagent --help                              # Show help
taskagent --version                           # Show version
taskagent login --server <url> --key <key>    # Configure authentication
taskagent whoami                              # Verify credentials
taskagent projects list                       # List projects
taskagent add --title "Task" --project <id>   # Add a task
taskagent list                                # List tasks
taskagent move <id> <status>                  # Change task status
```

## Shell Completion

The CLI supports static shell completion for Bash, Zsh, and Fish. Completion is static in v0.2.0 (commands and flags only; dynamic task/project names are not completed).

### Bash

Add to your `~/.bashrc`:

```bash
eval "$(taskagent completion bash)"
```

### Zsh

Add to your `~/.zshrc`:

```zsh
eval "$(taskagent completion zsh)"
```

### Fish

Add to `~/.config/fish/completions/taskagent.fish`:

```fish
taskagent completion fish | source
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
