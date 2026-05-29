"""Tests for config loading and storage."""

from pathlib import Path

import pytest

from taskagent_cli.config import Config, load_config, require_auth_config, write_config
from taskagent_cli.errors import UsageError


def test_load_config_defaults() -> None:
    """Test loading config with defaults."""
    config = load_config()
    assert config.server == "http://localhost:8080"
    assert config.timeout == 10


def test_write_and_load_config(tmp_path: Path) -> None:
    """Test writing and loading config from file."""
    import os

    # Override the home directory for testing
    os.environ["TASKAGENT_HOME"] = str(tmp_path)

    write_config("http://api.example.com", "test-key-123")

    config = load_config()
    assert config.server == "http://api.example.com"
    assert config.api_key == "test-key-123"

    # Check file permissions
    config_file = tmp_path / "config.json"
    assert config_file.exists()

    # Cleanup
    del os.environ["TASKAGENT_HOME"]


def test_env_vars_override_file(tmp_path: Path) -> None:
    """Test environment variables override config file."""
    import os

    os.environ["TASKAGENT_HOME"] = str(tmp_path)
    os.environ["TASKAGENT_SERVER"] = "http://env-server.example.com"
    os.environ["TASKAGENT_API_KEY"] = "env-key"

    write_config("http://file-server.example.com", "file-key")

    config = load_config()
    assert config.server == "http://env-server.example.com"
    assert config.api_key == "env-key"

    # Cleanup
    del os.environ["TASKAGENT_HOME"]
    del os.environ["TASKAGENT_SERVER"]
    del os.environ["TASKAGENT_API_KEY"]


def test_require_auth_config_with_key() -> None:
    """Test require_auth_config passes with valid config."""
    config = Config(
        server="http://localhost:8080",
        api_key="test-key",
        timeout=10,
        home=Path.home() / ".taskagent",
    )
    # Should not raise
    require_auth_config(config)


def test_require_auth_config_missing_key() -> None:
    """Test require_auth_config fails without API key."""
    config = Config(
        server="http://localhost:8080",
        api_key="",
        timeout=10,
        home=Path.home() / ".taskagent",
    )
    with pytest.raises(UsageError):
        require_auth_config(config)
