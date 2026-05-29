"""Smoke tests for TaskAgent CLI.

These tests verify the CLI is functional end-to-end.
"""

from click.testing import CliRunner

from taskagent_cli.cli import main


def test_cli_help() -> None:
    """Test CLI help is accessible."""
    runner = CliRunner()
    result = runner.invoke(main, ["--help"])
    assert result.exit_code == 0
    assert "TaskAgent CLI" in result.output
    assert "Commands:" in result.output


def test_cli_version() -> None:
    """Test CLI version is accessible."""
    runner = CliRunner()
    result = runner.invoke(main, ["--version"])
    assert result.exit_code == 0
    assert "0.2.0" in result.output


def test_projects_help() -> None:
    """Test projects subcommand help."""
    runner = CliRunner()
    result = runner.invoke(main, ["projects", "--help"])
    assert result.exit_code == 0
    assert "list" in result.output.lower()
    assert "create" in result.output.lower()


def test_completion_bash() -> None:
    """Test bash completion generation."""
    runner = CliRunner()
    result = runner.invoke(main, ["completion", "bash"])
    assert result.exit_code == 0
    assert "bash_source" in result.output


def test_completion_zsh() -> None:
    """Test zsh completion generation."""
    runner = CliRunner()
    result = runner.invoke(main, ["completion", "zsh"])
    assert result.exit_code == 0
    assert "zsh_source" in result.output


def test_completion_fish() -> None:
    """Test fish completion generation."""
    runner = CliRunner()
    result = runner.invoke(main, ["completion", "fish"])
    assert result.exit_code == 0
    assert "fish_source" in result.output


def test_invalid_command() -> None:
    """Test invalid command handling."""
    runner = CliRunner()
    result = runner.invoke(main, ["nonexistent"])
    assert result.exit_code == 2
    assert "Error" in result.output or "no such command" in result.output.lower()


def test_no_subcommand() -> None:
    """Test running with no subcommand exits with status 2."""
    runner = CliRunner()
    result = runner.invoke(main, [])
    assert result.exit_code == 2
    assert "Usage:" in result.output
