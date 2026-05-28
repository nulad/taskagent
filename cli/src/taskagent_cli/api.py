from __future__ import annotations

from typing import Any

import httpx

from taskagent_cli.errors import ApiClientError, ApiServerError, NetworkError


class TaskAgentClient:
    """HTTP client for TaskAgent API."""

    def __init__(
        self,
        base_url: str,
        api_key: str,
        timeout: int = 10,
        client: httpx.Client | None = None,
    ) -> None:
        """Initialize the client.

        Args:
            base_url: Base URL for the API (e.g., http://localhost:8080)
            api_key: API key for authentication
            timeout: Request timeout in seconds
            client: Optional pre-configured httpx.Client for testing
        """
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout
        self.client = client or httpx.Client()

    def _build_url(self, path: str) -> str:
        """Build a full URL from base URL and path."""
        path = path.lstrip("/")
        return f"{self.base_url}/{path}"

    def request(
        self,
        method: str,
        path: str,
        json: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> Any:
        """Make an HTTP request to the API.

        Args:
            method: HTTP method (GET, POST, etc.)
            path: API path (e.g., "/projects")
            json: JSON request body
            params: Query parameters

        Returns:
            Parsed JSON response, or None for 204 responses
        """
        url = self._build_url(path)
        headers = {
            "X-API-Key": self.api_key,
            "Accept": "application/json",
        }

        if json is not None:
            headers["Content-Type"] = "application/json"

        try:
            response = self.client.request(
                method,
                url,
                json=json,
                params=params,
                headers=headers,
                timeout=self.timeout,
            )
        except (httpx.TimeoutException, httpx.TransportError) as e:
            raise NetworkError(f"Network error: {e}") from e

        # Handle 2xx responses
        if 200 <= response.status_code < 300:
            if response.status_code == 204 or not response.content:
                return None
            return response.json()

        # Handle error responses
        error_message = self._extract_error_message(response)

        if 400 <= response.status_code < 500:
            raise ApiClientError(error_message)

        if 500 <= response.status_code < 600:
            raise ApiServerError(error_message)

        # Shouldn't reach here, but handle as client error
        raise ApiClientError(error_message)

    def _extract_error_message(self, response: httpx.Response) -> str:
        """Extract a human-readable error message from a response."""
        try:
            data = response.json()
            if isinstance(data, dict):
                # Try common error message fields
                if "message" in data:
                    return str(data["message"])
                if "error" in data:
                    return str(data["error"])
                if "code" in data:
                    return f"Error: {data['code']}"
        except (ValueError, TypeError):
            pass

        # Fall back to response text
        return response.text or f"HTTP {response.status_code}"
