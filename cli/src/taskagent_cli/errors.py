from __future__ import annotations

EXIT_OK = 0
EXIT_USAGE_ERROR = 2
EXIT_API_CLIENT_ERROR = 3
EXIT_API_SERVER_ERROR = 4
EXIT_NETWORK_ERROR = 5


class TaskAgentError(Exception):
    """Base exception for TaskAgent CLI errors."""

    def __init__(self, message: str, exit_code: int) -> None:
        self.message = message
        self.exit_code = exit_code
        super().__init__(message)


class UsageError(TaskAgentError):
    """Usage or configuration error."""

    def __init__(self, message: str) -> None:
        super().__init__(message, EXIT_USAGE_ERROR)


class ApiClientError(TaskAgentError):
    """API client error (4xx response)."""

    def __init__(self, message: str) -> None:
        super().__init__(message, EXIT_API_CLIENT_ERROR)


class ApiServerError(TaskAgentError):
    """API server error (5xx response)."""

    def __init__(self, message: str) -> None:
        super().__init__(message, EXIT_API_SERVER_ERROR)


class NetworkError(TaskAgentError):
    """Network error (connection, timeout, etc.)."""

    def __init__(self, message: str) -> None:
        super().__init__(message, EXIT_NETWORK_ERROR)
