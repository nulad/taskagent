from __future__ import annotations

import json
import os
from dataclasses import dataclass
from pathlib import Path

from taskagent_cli.errors import UsageError


@dataclass
class Config:
    """TaskAgent CLI configuration."""

    server: str
    api_key: str
    timeout: int
    home: Path


def default_taskagent_home() -> Path:
    """Get the default TaskAgent home directory."""
    env_home = os.environ.get("TASKAGENT_HOME")
    if env_home:
        return Path(env_home)
    return Path.home() / ".taskagent"


def load_config(
    server: str | None = None,
    api_key: str | None = None,
    timeout: int | None = None,
) -> Config:
    """Load config with precedence: explicit > env vars > config file > defaults."""
    home = default_taskagent_home()
    config_file = home / "config.json"

    # Load from config file if it exists
    file_config: dict[str, str | int] = {}
    if config_file.exists():
        try:
            with open(config_file) as f:
                file_config = json.load(f)
        except (json.JSONDecodeError, OSError):
            file_config = {}

    # Apply precedence: explicit > env > file > defaults
    final_server = (
        server
        or os.environ.get("TASKAGENT_SERVER")
        or file_config.get("server")
        or "http://localhost:8080"
    )

    final_api_key = (
        api_key
        or os.environ.get("TASKAGENT_API_KEY")
        or file_config.get("api_key")
        or ""
    )

    timeout_env = os.environ.get("TASKAGENT_TIMEOUT")
    final_timeout = (
        timeout
        or (int(timeout_env) if timeout_env else None)
        or (
            int(file_config["timeout"])
            if isinstance(file_config.get("timeout"), int)
            else None
        )
        or 10
    )

    return Config(
        server=str(final_server),
        api_key=str(final_api_key),
        timeout=int(final_timeout),
        home=home,
    )


def require_auth_config(config: Config) -> None:
    """Raise a UsageError if server or API key is missing."""
    if not config.server or not config.api_key:
        raise UsageError("Missing server or API key. Run `task login` to configure.")


def write_config(server: str, api_key: str) -> None:
    """Write config to file with secure permissions."""
    home = default_taskagent_home()
    home.mkdir(mode=0o700, parents=True, exist_ok=True)

    config_data = {
        "server": server,
        "api_key": api_key,
    }

    config_file = home / "config.json"
    with open(config_file, "w") as f:
        json.dump(config_data, f)

    # Set file permissions to 0600 (readable/writable by owner only)
    config_file.chmod(0o600)
