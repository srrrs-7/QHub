"""Tests for QHubClient initialization and HTTP behavior."""

from __future__ import annotations

import httpx
import pytest
import respx

from qhub.client import QHubClient
from qhub.errors import (
    APIError,
    AuthenticationError,
    ForbiddenError,
    NotFoundError,
    RateLimitError,
    ValidationError,
)


BASE_URL = "http://localhost:8080"
TOKEN = "test-token-123"


class TestClientInit:
    """Client initialization tests."""

    def test_default_base_url(self) -> None:
        c = QHubClient(bearer_token=TOKEN)
        assert c._base_url == "http://localhost:8080"

    def test_custom_base_url(self) -> None:
        c = QHubClient(bearer_token=TOKEN, base_url="https://api.example.com")
        assert c._base_url == "https://api.example.com"

    def test_trailing_slash_stripped(self) -> None:
        c = QHubClient(bearer_token=TOKEN, base_url="https://api.example.com/")
        assert c._base_url == "https://api.example.com"

    def test_custom_httpx_client(self) -> None:
        custom = httpx.Client()
        c = QHubClient(bearer_token=TOKEN, httpx_client=custom)
        assert c._http is custom
        assert c._owns_client is False
        custom.close()

    def test_context_manager(self) -> None:
        with QHubClient(bearer_token=TOKEN) as c:
            assert c._http is not None

    def test_lazy_resource_properties(self) -> None:
        c = QHubClient(bearer_token=TOKEN)
        assert c._organizations is None
        _ = c.organizations
        assert c._organizations is not None

    def test_all_resource_properties_exist(self) -> None:
        c = QHubClient(bearer_token=TOKEN)
        props = [
            "organizations",
            "projects",
            "prompts",
            "versions",
            "logs",
            "evaluations",
            "consulting",
            "tags",
            "industries",
            "search",
            "analytics",
        ]
        for prop in props:
            assert hasattr(c, prop), f"Missing property: {prop}"
            resource = getattr(c, prop)
            assert resource is not None


class TestAuthHeader:
    """Verify the Authorization header is set correctly."""

    @respx.mock(base_url=BASE_URL)
    def test_bearer_token_sent(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(200, json=[])
        )
        c = QHubClient(bearer_token="my-secret-token", base_url=BASE_URL)
        c.tags.list()
        req = respx_mock.calls[0].request
        assert req.headers["authorization"] == "Bearer my-secret-token"


class TestErrorHandling:
    """Test that HTTP errors are mapped to the correct exception types."""

    @respx.mock(base_url=BASE_URL)
    def test_400_raises_validation_error(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(400, text="bad request")
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        with pytest.raises(ValidationError, match="bad request"):
            c.tags.list()

    @respx.mock(base_url=BASE_URL)
    def test_401_raises_authentication_error(
        self, respx_mock: respx.MockRouter
    ) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(401, text="unauthorized")
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        with pytest.raises(AuthenticationError):
            c.tags.list()

    @respx.mock(base_url=BASE_URL)
    def test_403_raises_forbidden_error(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(403, text="forbidden")
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        with pytest.raises(ForbiddenError):
            c.tags.list()

    @respx.mock(base_url=BASE_URL)
    def test_404_raises_not_found_error(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/organizations/missing").mock(
            return_value=httpx.Response(404, text="not found")
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        with pytest.raises(NotFoundError):
            c.organizations.get("missing")

    @respx.mock(base_url=BASE_URL)
    def test_429_raises_rate_limit_error(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(429, text="rate limited")
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        with pytest.raises(RateLimitError):
            c.tags.list()

    @respx.mock(base_url=BASE_URL)
    def test_500_raises_api_error(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.get("/api/v1/tags").mock(
            return_value=httpx.Response(500, text="internal error")
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        with pytest.raises(APIError) as exc_info:
            c.tags.list()
        assert exc_info.value.status_code == 500

    @respx.mock(base_url=BASE_URL)
    def test_204_returns_none(self, respx_mock: respx.MockRouter) -> None:
        respx_mock.delete("/api/v1/tags/abc-123").mock(
            return_value=httpx.Response(204)
        )
        c = QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
        result = c.tags.delete("abc-123")
        assert result is None
