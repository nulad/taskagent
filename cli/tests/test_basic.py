"""Basic tests for taskagent CLI."""

import json
from io import StringIO
from unittest.mock import patch

from click.testing import CliRunner

from taskagent_cli import __version__, format, models
from taskagent_cli.cli import main


def test_version() -> None:
    """Test --version flag."""
    runner = CliRunner()
    result = runner.invoke(main, ["--version"])
    assert result.exit_code == 0
    assert __version__ in result.output


def test_help() -> None:
    """Test --help flag."""
    runner = CliRunner()
    result = runner.invoke(main, ["--help"])
    assert result.exit_code == 0
    assert "TaskAgent CLI" in result.output


def test_no_subcommand() -> None:
    """Test that no subcommand exits 2."""
    runner = CliRunner()
    result = runner.invoke(main, [])
    assert result.exit_code == 2


def test_parse_tags() -> None:
    """Test parse_tags function."""
    assert models.parse_tags("bug, urgent,,") == ["bug", "urgent"]
    assert models.parse_tags("") == []
    assert models.parse_tags("single") == ["single"]


def test_valid_statuses() -> None:
    """Test status validation."""
    assert models.is_valid_status("todo")
    assert not models.is_valid_status("invalid")


def test_print_json() -> None:
    """Test print_json function."""
    with patch("sys.stdout", new_callable=StringIO) as mock_stdout:
        format.print_json({"test": "value"})
        output = mock_stdout.getvalue()
        assert '"test": "value"' in output
        data = json.loads(output)
        assert data["test"] == "value"
