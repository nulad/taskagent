from __future__ import annotations

import re
from typing import TYPE_CHECKING, Any

from taskagent_cli.errors import ApiClientError

if TYPE_CHECKING:
    from taskagent_cli.api import TaskAgentClient

VALID_STATUSES = ("backlog", "todo", "in-progress", "review", "done")

# UUID pattern (simplified)
UUID_PATTERN = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", re.IGNORECASE
)


def is_valid_status(status: str) -> bool:
    """Check if a status value is valid."""
    return status in VALID_STATUSES


def parse_tags(tags_str: str) -> list[str]:
    """Parse comma-separated tags into a trimmed list.

    Drops empty entries.
    """
    if not tags_str:
        return []
    return [tag.strip() for tag in tags_str.split(",") if tag.strip()]


def build_partial_update_payload(
    title: str | None = None,
    description: str | None = None,
    tags: str | None = None,
    status: str | None = None,
) -> dict[str, Any]:
    """Build update payload with only explicitly provided fields."""
    payload: dict[str, Any] = {}
    if title is not None:
        payload["title"] = title
    if description is not None:
        payload["description"] = description
    if tags is not None:
        payload["tags"] = parse_tags(tags)
    if status is not None:
        payload["status"] = status
    return payload


def resolve_project_id(client: TaskAgentClient, project_ref: str) -> str:
    """Resolve a project name or ID to a project ID.

    If project_ref looks like a UUID, return it unchanged.
    Otherwise, fetch projects and find one matching the name.

    Args:
        client: Authenticated API client
        project_ref: Project ID (UUID) or project name

    Returns:
        Project ID

    Raises:
        ApiClientError: If project not found or ambiguous
    """
    # If it looks like a UUID, return as-is
    if UUID_PATTERN.match(project_ref):
        return project_ref

    # Otherwise, look up by name
    try:
        projects = client.request("GET", "/projects")
    except ApiClientError as e:
        raise ApiClientError(f"Failed to resolve project: {e.message}") from e

    if not isinstance(projects, list):
        raise ApiClientError("Invalid projects response from server")

    # Find projects matching the name
    matches = [
        p for p in projects if isinstance(p, dict) and p.get("name") == project_ref
    ]

    if not matches:
        raise ApiClientError(f"project not found: {project_ref}")

    if len(matches) > 1:
        raise ApiClientError(
            f"ambiguous project name: {project_ref} (use project ID to disambiguate)"
        )

    project_id = matches[0].get("id")
    if not isinstance(project_id, str) or not project_id:
        raise ApiClientError("Invalid project response: missing id")

    return project_id
