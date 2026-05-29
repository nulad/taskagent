"""Tests for Click commands."""

from pathlib import Path
from unittest.mock import patch

from click.testing import CliRunner

from taskagent_cli.cli import main
from taskagent_cli.config import Config


def test_login_command() -> None:
    """Test login command."""
    runner = CliRunner()
    with patch("taskagent_cli.cli.write_config") as mock_write:
        result = runner.invoke(
            main, ["login", "--server", "http://localhost:8080", "--key", "test-key"]
        )
        assert result.exit_code == 0
        mock_write.assert_called_once_with("http://localhost:8080", "test-key")


def test_login_missing_key() -> None:
    """Test login without key."""
    runner = CliRunner()
    result = runner.invoke(main, ["login", "--server", "http://localhost:8080"])
    assert result.exit_code == 2


def test_projects_list() -> None:
    """Test projects list command."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = [{"id": "1", "name": "Project 1"}]

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["projects", "list"])
            assert result.exit_code == 0
            assert "Project 1" in result.output


def test_task_add() -> None:
    """Test task add command."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = {"id": "1", "title": "Test Task"}

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            # Use a UUID-like project ID to avoid name resolution
            result = runner.invoke(
                main,
                [
                    "add",
                    "--title",
                    "Test Task",
                    "--project",
                    "12345678-1234-1234-1234-123456789012",
                ],
            )
            assert result.exit_code == 0
            # Should call request once for creating the task
            assert mock_client.request.called


def test_task_move() -> None:
    """Test task move command."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = {"id": "1", "status": "todo"}

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["move", "task-1", "todo"])
            assert result.exit_code == 0


def test_task_move_invalid_status() -> None:
    """Test task move with invalid status."""
    runner = CliRunner()
    result = runner.invoke(main, ["move", "task-1", "invalid-status"])
    assert result.exit_code == 2
    assert "invalid status" in result.output.lower()


def test_completion_bash() -> None:
    """Test completion command for bash."""
    runner = CliRunner()
    result = runner.invoke(main, ["completion", "bash"])
    assert result.exit_code == 0
    assert "bash_source" in result.output


def test_completion_zsh() -> None:
    """Test completion command for zsh."""
    runner = CliRunner()
    result = runner.invoke(main, ["completion", "zsh"])
    assert result.exit_code == 0
    assert "zsh_source" in result.output


def test_completion_fish() -> None:
    """Test completion command for fish."""
    runner = CliRunner()
    result = runner.invoke(main, ["completion", "fish"])
    assert result.exit_code == 0
    assert "fish_source" in result.output


def test_projects_list_human_format() -> None:
    """Test projects list with human format."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = [
            {"id": "550e8400-e29b-41d4-a716-446655440000", "name": "Project 1", "created_at": "", "updated_at": ""}
        ]

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["projects", "list", "--format", "human"])
            assert result.exit_code == 0
            assert "Project 1" in result.output
            assert "550e8400" in result.output


def test_projects_show_human_format() -> None:
    """Test projects show with human format."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "name": "Project 1",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-02T00:00:00Z",
        }

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["projects", "show", "proj-1", "--format", "human"])
            assert result.exit_code == 0
            assert "Project 1" in result.output
            assert "550e8400-e29b-41d4-a716-446655440000" in result.output


def test_task_list_human_format() -> None:
    """Test task list with human format."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = [
            {"id": "550e8400-e29b-41d4-a716-446655440000", "title": "Task 1", "status": "todo", "project_id": "p1", "updated_at": ""}
        ]

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["list", "--format", "human"])
            assert result.exit_code == 0
            assert "Task 1" in result.output
            assert "todo" in result.output


def test_task_show_human_format() -> None:
    """Test task show with human format."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "title": "Task 1",
            "status": "todo",
            "project_id": "p1",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-02T00:00:00Z",
        }

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["show", "task-1", "--format", "human"])
            assert result.exit_code == 0
            assert "Task 1" in result.output
            assert "550e8400-e29b-41d4-a716-446655440000" in result.output


def test_whoami_human_format() -> None:
    """Test whoami with human format."""
    runner = CliRunner()
    with patch("taskagent_cli.api.TaskAgentClient") as mock_client_cls:
        mock_client = mock_client_cls.return_value
        mock_client.request.return_value = [
            {"id": 1, "label": "my-key", "user_name": "alice", "created_at": "2024-01-01T00:00:00Z"}
        ]

        with patch("taskagent_cli.config.load_config") as mock_load:
            mock_config = Config(
                server="http://localhost:8080",
                api_key="test-key",
                timeout=10,
                home=Path.home() / ".taskagent",
            )
            mock_load.return_value = mock_config

            result = runner.invoke(main, ["whoami", "--format", "human"])
            assert result.exit_code == 0
            assert "my-key" in result.output
            assert "alice" in result.output


def test_invalid_format_flag() -> None:
    """Test that invalid format value exits with code 2."""
    runner = CliRunner()
    result = runner.invoke(main, ["projects", "list", "--format", "xml"])
    assert result.exit_code == 2
