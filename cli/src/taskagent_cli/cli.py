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
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("GET", "/projects")
    print_json(result or [])


@projects.command(name="create")
@click.option("--name", required=True, help="Project name")
@click.option("--description", help="Project description")
def projects_create(name: str, description: str | None) -> None:
    """Create a project."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    payload: dict[str, str | None] = {"name": name}
    if description is not None:
        payload["description"] = description

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("POST", "/projects", json=payload)
    print_json(result)


@projects.command(name="show")
@click.argument("project_id")
def projects_show(project_id: str) -> None:
    """Show project details."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("GET", f"/projects/{project_id}")
    print_json(result)


@projects.command(name="update")
@click.argument("project_id")
@click.option("--name", help="New project name")
@click.option("--description", help="New project description")
def projects_update(project_id: str, name: str | None, description: str | None) -> None:
    """Update a project."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    if not name and not description:
        raise click.UsageError("At least one of --name or --description is required")

    config = load_config()
    require_auth_config(config)

    payload: dict[str, str] = {}
    if name is not None:
        payload["name"] = name
    if description is not None:
        payload["description"] = description

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("PUT", f"/projects/{project_id}", json=payload)
    print_json(result)


@projects.command(name="delete")
@click.argument("project_id")
def projects_delete(project_id: str) -> None:
    """Delete a project."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("DELETE", f"/projects/{project_id}")
    if result is not None:
        print_json(result)


@main.command()
@click.option("--title", required=True, help="Task title")
@click.option("--project", required=True, help="Project ID or name")
@click.option("--description", help="Task description")
@click.option("--tags", help="Comma-separated tags")
def add(title: str, project: str, description: str | None, tags: str | None) -> None:
    """Add a new task."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json
    from taskagent_cli.models import parse_tags, resolve_project_id

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    project_id = resolve_project_id(client, project)

    payload: dict[str, str | list[str]] = {
        "title": title,
        "project_id": project_id,
    }
    if description is not None:
        payload["description"] = description
    if tags is not None:
        payload["tags"] = parse_tags(tags)

    result = client.request("POST", "/tasks", json=payload)
    print_json(result)


@main.command(name="list")
@click.option("--project", help="Filter by project ID or name")
@click.option("--status", help="Filter by task status")
@click.option("--limit", type=int, help="Limit number of results")
@click.option("--offset", type=int, help="Offset for pagination")
def task_list(
    project: str | None, status: str | None, limit: int | None, offset: int | None
) -> None:
    """List tasks."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json
    from taskagent_cli.models import VALID_STATUSES, is_valid_status, resolve_project_id

    if status is not None and not is_valid_status(status):
        raise click.UsageError(
            f"Invalid status: {status} (must be one of: {', '.join(VALID_STATUSES)})"
        )

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)

    params: dict[str, str | int] = {}
    if project is not None:
        project_id = resolve_project_id(client, project)
        params["project_id"] = project_id
    if status is not None:
        params["status"] = status
    if limit is not None:
        params["limit"] = limit
    if offset is not None:
        params["offset"] = offset

    result = client.request("GET", "/tasks", params=params)
    print_json(result or [])


@main.command()
@click.argument("task_id")
def show(task_id: str) -> None:
    """Show task details."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("GET", f"/tasks/{task_id}")
    print_json(result)


@main.command()
@click.argument("task_id")
@click.option("--title", help="New task title")
@click.option("--description", help="New task description")
@click.option("--tags", help="New comma-separated tags")
def update(
    task_id: str, title: str | None, description: str | None, tags: str | None
) -> None:
    """Update a task."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json
    from taskagent_cli.models import parse_tags

    if not title and not description and not tags:
        raise click.UsageError(
            "At least one of --title, --description, or --tags is required"
        )

    config = load_config()
    require_auth_config(config)

    payload: dict[str, str | list[str]] = {}
    if title is not None:
        payload["title"] = title
    if description is not None:
        payload["description"] = description
    if tags is not None:
        payload["tags"] = parse_tags(tags)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("PUT", f"/tasks/{task_id}", json=payload)
    print_json(result)


@main.command()
@click.argument("task_id")
def delete(task_id: str) -> None:
    """Delete a task."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json

    config = load_config()
    require_auth_config(config)

    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("DELETE", f"/tasks/{task_id}")
    if result is not None:
        print_json(result)


@main.command()
@click.argument("task_id")
@click.argument("status")
def move(task_id: str, status: str) -> None:
    """Move a task to a different status."""
    from taskagent_cli.api import TaskAgentClient
    from taskagent_cli.config import load_config, require_auth_config
    from taskagent_cli.format import print_json
    from taskagent_cli.models import VALID_STATUSES, is_valid_status

    if not is_valid_status(status):
        raise click.UsageError(
            f"invalid status: {status} (must be one of: {', '.join(VALID_STATUSES)})"
        )

    config = load_config()
    require_auth_config(config)

    payload = {"status": status}
    client = TaskAgentClient(config.server, config.api_key, config.timeout)
    result = client.request("PATCH", f"/tasks/{task_id}/move", json=payload)
    print_json(result)


@main.command()
@click.argument("shell", type=click.Choice(["bash", "zsh", "fish"]))
def completion(shell: str) -> None:
    """Generate shell completion script."""
    if shell == "bash":
        click.echo(
            "# Bash completion for task CLI\n"
            'eval "$(_TASK_COMPLETE=bash_source task)"\n'
        )
    elif shell == "zsh":
        click.echo(
            '# Zsh completion for task CLI\neval "$(_TASK_COMPLETE=zsh_source task)"\n'
        )
    elif shell == "fish":
        click.echo(
            "# Fish completion for task CLI\n"
            "eval (env _TASK_COMPLETE=fish_source task)\n"
        )


def run() -> None:
    """Run the CLI with proper error handling."""
    try:
        main()
    except TaskAgentError as e:
        click.echo(e.message, err=True)
        sys.exit(e.exit_code)


if __name__ == "__main__":
    run()
