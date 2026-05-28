from __future__ import annotations

from typing import Any

VALID_STATUSES = ("backlog", "todo", "in-progress", "review", "done")


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
