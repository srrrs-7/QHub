"""Shared fixtures for QHub SDK tests."""

from __future__ import annotations

import pytest
import httpx
import respx

from qhub.client import QHubClient


BASE_URL = "http://localhost:8080"
TOKEN = "test-token-123"


@pytest.fixture()
def mock_api() -> respx.MockRouter:
    """Create a respx mock router scoped to the base URL."""
    with respx.mock(base_url=BASE_URL) as router:
        yield router


@pytest.fixture()
def client(mock_api: respx.MockRouter) -> QHubClient:
    """Create a QHubClient wired to the mock router."""
    return QHubClient(bearer_token=TOKEN, base_url=BASE_URL)
