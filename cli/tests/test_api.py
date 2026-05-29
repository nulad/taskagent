"""Tests for API client error handling."""

from unittest.mock import Mock

import httpx
import pytest

from taskagent_cli.api import TaskAgentClient
from taskagent_cli.errors import ApiClientError, ApiServerError, NetworkError


def test_api_client_success() -> None:
    """Test successful API request."""
    mock_response = Mock(spec=httpx.Response)
    mock_response.status_code = 200
    mock_response.json.return_value = {"id": "123", "name": "Test"}

    mock_client = Mock(spec=httpx.Client)
    mock_client.request.return_value = mock_response

    client = TaskAgentClient("http://localhost:8080", "test-key", client=mock_client)
    result = client.request("GET", "/projects/123")

    assert result == {"id": "123", "name": "Test"}
    mock_client.request.assert_called_once()


def test_api_client_204_response() -> None:
    """Test empty 204 response."""
    mock_response = Mock(spec=httpx.Response)
    mock_response.status_code = 204
    mock_response.content = b""

    mock_client = Mock(spec=httpx.Client)
    mock_client.request.return_value = mock_response

    client = TaskAgentClient("http://localhost:8080", "test-key", client=mock_client)
    result = client.request("DELETE", "/projects/123")

    assert result is None


def test_api_client_4xx_error() -> None:
    """Test 4xx error handling."""
    mock_response = Mock(spec=httpx.Response)
    mock_response.status_code = 404
    mock_response.json.return_value = {"error": "Not found"}

    mock_client = Mock(spec=httpx.Client)
    mock_client.request.return_value = mock_response

    client = TaskAgentClient("http://localhost:8080", "test-key", client=mock_client)

    with pytest.raises(ApiClientError) as exc_info:
        client.request("GET", "/projects/123")

    assert exc_info.value.exit_code == 3


def test_api_client_5xx_error() -> None:
    """Test 5xx error handling."""
    mock_response = Mock(spec=httpx.Response)
    mock_response.status_code = 500
    mock_response.json.return_value = {"error": "Internal server error"}

    mock_client = Mock(spec=httpx.Client)
    mock_client.request.return_value = mock_response

    client = TaskAgentClient("http://localhost:8080", "test-key", client=mock_client)

    with pytest.raises(ApiServerError) as exc_info:
        client.request("GET", "/projects/123")

    assert exc_info.value.exit_code == 4


def test_api_client_network_error() -> None:
    """Test network error handling."""
    mock_client = Mock(spec=httpx.Client)
    mock_client.request.side_effect = httpx.TimeoutException("timeout")

    client = TaskAgentClient("http://localhost:8080", "test-key", client=mock_client)

    with pytest.raises(NetworkError) as exc_info:
        client.request("GET", "/projects/123")

    assert exc_info.value.exit_code == 5


def test_api_client_headers() -> None:
    """Test that API client sets correct headers."""
    mock_response = Mock(spec=httpx.Response)
    mock_response.status_code = 200
    mock_response.json.return_value = {}

    mock_client = Mock(spec=httpx.Client)
    mock_client.request.return_value = mock_response

    client = TaskAgentClient("http://localhost:8080", "test-key", client=mock_client)
    client.request("POST", "/tasks", json={"title": "Test"})

    # Check that headers were set correctly
    call_kwargs = mock_client.request.call_args[1]
    assert call_kwargs["headers"]["X-API-Key"] == "test-key"
    assert call_kwargs["headers"]["Accept"] == "application/json"
    assert call_kwargs["headers"]["Content-Type"] == "application/json"
