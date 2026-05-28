from __future__ import annotations

import sys

import click

from taskagent_cli import __version__
from taskagent_cli.config import write_config
from taskagent_cli.errors import (
    ApiClientError,
    ApiServerError,
    NetworkError,
    TaskAgentError,
)


class ErrorHandler(click.Group):
    """Custom Click group to handle TaskAgentError exceptions."""

    def invoke(self, ctx: click.Context) -> None:
        """Override invoke to catch TaskAgentError."""
        try:
            super().invoke(ctx)
        except (ApiClientError, ApiServerError, NetworkError) as e:
            click.echo(e.message, err=True)
            ctx.exit(e.exit_code)
        except TaskAgentError as e:
            click.echo(e.message, err=True)
            ctx.exit(e.exit_code)


@click.group(cls=ErrorHandler, invoke_without_command=True)
@click.option("--version", is_flag=True, help="Show version and exit.")
@click.pass_context
def main(ctx: click.Context, version: bool) -> None:
    """TaskAgent CLI - AI-agent-native task tracker."""
    if version:
        click.echo(__version__)
        ctx.exit(0)

    if ctx.invoked_subcommand is None:
        click.echo(ctx.get_help())
        ctx.exit(2)


@main.command()
@click.option("--server", required=True, help="TaskAgent server URL.")
@click.option("--key", help="API key (or - to read from stdin).")
def login(server: str, key: str | None) -> None:
    """Log in to TaskAgent server."""
    if key is None:
        raise click.UsageError("--key is required")

    if key == "-":
        key = click.prompt("Enter API key", hide_input=True)

    write_config(server, key)
    click.echo("Logged in successfully.", err=True)


@main.command()
def whoami() -> None:
    """Display current user/key information."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("GET", "/auth/keys")
    print_json(result)


@main.group()
def projects() -> None:
    """Manage projects."""
    pass


@projects.command(name="list")
def projects_list() -> None:
    """List projects."""
    raise NotImplementedError("TASK-011")


@projects.command(name="create")
def projects_create() -> None:
    """Create a project."""
    raise NotImplementedError("TASK-012")


@projects.command(name="show")
def projects_show() -> None:
    """Show project details."""
    raise NotImplementedError("TASK-011")


@projects.command(name="update")
def projects_update() -> None:
    """Update a project."""
    raise NotImplementedError("TASK-013")


@projects.command(name="delete")
def projects_delete() -> None:
    """Delete a project."""
    raise NotImplementedError("TASK-014")


@main.command()
def add() -> None:
    """Add a new task."""
    raise NotImplementedError("TASK-016")


@main.command()
def list() -> None:
    """List tasks."""
    raise NotImplementedError("TASK-016")


@main.command()
def show() -> None:
    """Show task details."""
    raise NotImplementedError("TASK-016")


@main.command()
def update() -> None:
    """Update a task."""
    raise NotImplementedError("TASK-018")


@main.command()
def delete() -> None:
    """Delete a task."""
    raise NotImplementedError("TASK-016")


@main.command()
def move() -> None:
    """Move a task to a different status."""
    raise NotImplementedError("TASK-021")


@main.command()
def completion() -> None:
    """Generate shell completion script."""
    raise NotImplementedError("TASK-028")


def run() -> None:
    """Run the CLI with proper error handling."""
    try:
        main()
    except TaskAgentError as e:
        click.echo(e.message, err=True)
        sys.exit(e.exit_code)


if __name__ == "__main__":
    run()
