import json
import os
import sys
from typing import Any, Optional


# ANSI color codes
_RESET = "\033[0m"
_GRAY = "\033[90m"
_YELLOW = "\033[33m"
_BLUE = "\033[34m"
_MAGENTA = "\033[35m"
_GREEN = "\033[32m"

_STATUS_COLORS = {
    "backlog": _GRAY,
    "todo": _YELLOW,
    "in-progress": _BLUE,
    "review": _MAGENTA,
    "done": _GREEN,
}


def color_enabled() -> bool:
    """Return True if ANSI colors should be used."""
    return sys.stdout.isatty() and os.environ.get("NO_COLOR") is None


def _colorize(text: str, color: str) -> str:
    """Wrap text in color codes if color is enabled."""
    if color_enabled():
        return f"{color}{text}{_RESET}"
    return text


def _status_color(status: str) -> str:
    """Return colorized status string."""
    color = _STATUS_COLORS.get(status, "")
    return _colorize(status, color)


def _truncate_id(id_str: str) -> str:
    """Truncate ID to first 8 characters."""
    return id_str[:8] if id_str else ""


def _fmt_timestamp(ts: Optional[str]) -> str:
    """Format timestamp for table view (just return as-is for detail)."""
    return ts or ""


def _column_widths(rows: list[list[str]], headers: list[str]) -> list[int]:
    """Calculate max width for each column."""
    widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if i < len(widths):
                widths[i] = max(widths[i], len(cell))
    return widths


def _render_table(headers: list[str], rows: list[list[str]]) -> str:
    """Render a simple text table."""
    if not rows:
        return "  ".join(headers)
    widths = _column_widths(rows, headers)
    lines: list[str] = []
    header_line = "  ".join(
        h.ljust(widths[i]) for i, h in enumerate(headers)
    )
    lines.append(header_line)
    lines.append("-" * len(header_line))
    for row in rows:
        lines.append("  ".join(
            cell.ljust(widths[i]) for i, cell in enumerate(row)
        ))
    return "\n".join(lines)


def print_json(data: Any) -> None:
    """Print data as JSON to stdout."""
    json.dump(data, fp=sys.stdout)
    sys.stdout.write("\n")


def render_project_table(projects: list[dict[str, Any]]) -> str:
    """Render a list of projects as a human-readable table."""
    if not projects:
        return "No projects found."
    headers = ["ID", "Name", "Description", "Created", "Updated"]
    rows: list[list[str]] = []
    for project in projects:
        rows.append([
            _truncate_id(project.get("id", "")),
            project.get("name", ""),
            (project.get("description", "") or "")[:40],
            _fmt_timestamp(project.get("created_at", "")),
            _fmt_timestamp(project.get("updated_at", "")),
        ])
    return _render_table(headers, rows)


def render_project_detail(project: dict[str, Any]) -> str:
    """Render a single project in human-readable detail format."""
    lines: list[str] = [
        f"ID:          {project.get('id', '')}",
        f"Name:        {project.get('name', '')}",
        f"Description: {project.get('description', '') or ''}",
        f"Created:     {project.get('created_at', '')}",
        f"Updated:     {project.get('updated_at', '')}",
    ]
    return "\n".join(lines)


def render_task_table(tasks: list[dict[str, Any]]) -> str:
    """Render a list of tasks as a human-readable table."""
    if not tasks:
        return "No tasks found."
    headers = ["ID", "Status", "Title", "Project", "Updated"]
    rows: list[list[str]] = []
    for task in tasks:
        status = task.get("status", "")
        project_ref = task.get("project_name") or _truncate_id(task.get("project_id", ""))
        rows.append([
            _truncate_id(task.get("id", "")),
            _status_color(status),
            task.get("title", "")[:50],
            project_ref,
            _fmt_timestamp(task.get("updated_at", "")),
        ])
    return _render_table(headers, rows)


def render_task_detail(task: dict[str, Any]) -> str:
    """Render a single task in human-readable detail format."""
    status = task.get("status", "")
    tags = task.get("tags")
    tags_str = ", ".join(tags) if tags else ""
    lines: list[str] = [
        f"ID:          {task.get('id', '')}",
        f"Title:       {task.get('title', '')}",
        f"Status:      {_status_color(status)}",
        f"Project:     {task.get('project_id', '')}",
        f"Description: {task.get('description', '') or ''}",
        f"Tags:        {tags_str}",
        f"Created:     {task.get('created_at', '')}",
        f"Updated:     {task.get('updated_at', '')}",
    ]
    return "\n".join(lines)


def render_auth_key_detail(data: list[dict[str, Any]]) -> str:
    """Render auth key information in human-readable format."""
    if not data:
        return "No API keys found."
    if len(data) == 1:
        key = data[0]
        lines: list[str] = [
            f"ID:        {key.get('id', '')}",
            f"Label:     {key.get('label', '')}",
            f"User:      {key.get('user_name', '')}",
            f"Created:   {key.get('created_at', '')}",
        ]
        return "\n".join(lines)
    headers = ["ID", "Label", "User", "Created"]
    rows: list[list[str]] = []
    for key in data:
        rows.append([
            str(key.get("id", "")),
            key.get("label", ""),
            key.get("user_name", ""),
            key.get("created_at", ""),
        ])
    return _render_table(headers, rows)
