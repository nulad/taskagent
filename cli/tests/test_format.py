"""Tests for formatting helpers."""

from unittest.mock import patch

import pytest

from taskagent_cli.format import (
    color_enabled,
    render_auth_key_detail,
    render_project_detail,
    render_project_table,
    render_task_detail,
    render_task_table,
)


class TestColorEnabled:
    def test_no_color_env_disables(self) -> None:
        with patch("taskagent_cli.format.sys.stdout.isatty", return_value=True):
            with patch.dict("os.environ", {"NO_COLOR": "1"}):
                assert color_enabled() is False

    def test_non_tty_disables(self) -> None:
        with patch("taskagent_cli.format.sys.stdout.isatty", return_value=False):
            with patch.dict("os.environ", {}, clear=True):
                assert color_enabled() is False

    def test_tty_and_no_env_enables(self) -> None:
        with patch("taskagent_cli.format.sys.stdout.isatty", return_value=True):
            with patch.dict("os.environ", {}, clear=True):
                assert color_enabled() is True


class TestRenderProjectTable:
    def test_empty(self) -> None:
        result = render_project_table([])
        assert result == "No projects found."

    def test_single_project(self) -> None:
        projects = [
            {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "name": "My Project",
                "description": "A test project",
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-01-02T00:00:00Z",
            }
        ]
        result = render_project_table(projects)
        assert "550e8400" in result
        assert "My Project" in result
        assert "A test project" in result
        assert "2024-01-01T00:00:00Z" in result

    def test_truncates_id(self) -> None:
        projects = [
            {"id": "550e8400-e29b-41d4-a716-446655440000", "name": "P", "created_at": "", "updated_at": ""}
        ]
        result = render_project_table(projects)
        assert "550e8400-e29b" not in result
        assert "550e8400" in result


class TestRenderProjectDetail:
    def test_full_fields(self) -> None:
        project = {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "name": "My Project",
            "description": "Details",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-02T00:00:00Z",
        }
        result = render_project_detail(project)
        assert "550e8400-e29b-41d4-a716-446655440000" in result
        assert "My Project" in result
        assert "Details" in result
        assert "2024-01-01T00:00:00Z" in result

    def test_null_description(self) -> None:
        project = {
            "id": "1",
            "name": "P",
            "description": None,
            "created_at": "",
            "updated_at": "",
        }
        result = render_project_detail(project)
        assert "Description:" in result


class TestRenderTaskTable:
    def test_empty(self) -> None:
        result = render_task_table([])
        assert result == "No tasks found."

    def test_single_task_no_color(self) -> None:
        tasks = [
            {
                "id": "550e8400-e29b-41d4-a716-446655440000",
                "title": "Do thing",
                "status": "todo",
                "project_id": "proj-1",
                "updated_at": "2024-01-02T00:00:00Z",
            }
        ]
        with patch("taskagent_cli.format.color_enabled", return_value=False):
            result = render_task_table(tasks)
        assert "550e8400" in result
        assert "Do thing" in result
        assert "todo" in result
        assert "proj-1" in result

    def test_status_colors_when_enabled(self) -> None:
        tasks = [
            {
                "id": "1",
                "title": "T",
                "status": "done",
                "project_id": "p1",
                "updated_at": "",
            }
        ]
        with patch("taskagent_cli.format.color_enabled", return_value=True):
            result = render_task_table(tasks)
        assert "\033[32m" in result  # green for done

    def test_project_name_fallback(self) -> None:
        tasks = [
            {
                "id": "1",
                "title": "T",
                "status": "backlog",
                "project_id": "p1",
                "project_name": "My Project",
                "updated_at": "",
            }
        ]
        with patch("taskagent_cli.format.color_enabled", return_value=False):
            result = render_task_table(tasks)
        assert "My Project" in result


class TestRenderTaskDetail:
    def test_full_fields(self) -> None:
        task = {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "title": "Do thing",
            "status": "todo",
            "project_id": "proj-1",
            "description": "Desc",
            "tags": ["a", "b"],
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-02T00:00:00Z",
        }
        with patch("taskagent_cli.format.color_enabled", return_value=False):
            result = render_task_detail(task)
        assert "550e8400-e29b-41d4-a716-446655440000" in result
        assert "Do thing" in result
        assert "todo" in result
        assert "proj-1" in result
        assert "Desc" in result
        assert "a, b" in result

    def test_empty_tags(self) -> None:
        task = {
            "id": "1",
            "title": "T",
            "status": "backlog",
            "project_id": "p1",
            "tags": [],
            "created_at": "",
            "updated_at": "",
        }
        with patch("taskagent_cli.format.color_enabled", return_value=False):
            result = render_task_detail(task)
        assert "Tags:" in result

    def test_missing_optional_fields(self) -> None:
        task = {
            "id": "1",
            "title": "T",
            "status": "backlog",
            "project_id": "p1",
            "created_at": "",
            "updated_at": "",
        }
        with patch("taskagent_cli.format.color_enabled", return_value=False):
            result = render_task_detail(task)
        assert "Tags:" in result
        assert "Description:" in result


class TestRenderAuthKeyDetail:
    def test_empty(self) -> None:
        result = render_auth_key_detail([])
        assert result == "No API keys found."

    def test_single_key(self) -> None:
        data = [
            {
                "id": 1,
                "label": "my-key",
                "user_name": "alice",
                "created_at": "2024-01-01T00:00:00Z",
            }
        ]
        result = render_auth_key_detail(data)
        assert "1" in result
        assert "my-key" in result
        assert "alice" in result
        assert "2024-01-01T00:00:00Z" in result

    def test_multiple_keys(self) -> None:
        data = [
            {"id": 1, "label": "k1", "user_name": "a", "created_at": "2024-01-01T00:00:00Z"},
            {"id": 2, "label": "k2", "user_name": "b", "created_at": "2024-01-02T00:00:00Z"},
        ]
        result = render_auth_key_detail(data)
        assert "k1" in result
        assert "k2" in result
        assert "a" in result
        assert "b" in result
