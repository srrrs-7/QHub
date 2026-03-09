"""QHub API client."""

from __future__ import annotations

from typing import Any

import httpx

from qhub.errors import raise_for_status
from qhub.resources.analytics import AnalyticsResource
from qhub.resources.consulting import ConsultingResource
from qhub.resources.evaluations import EvaluationsResource
from qhub.resources.industries import IndustriesResource
from qhub.resources.logs import LogsResource
from qhub.resources.organizations import OrganizationsResource
from qhub.resources.projects import ProjectsResource
from qhub.resources.prompts import PromptsResource
from qhub.resources.search import SearchResource
from qhub.resources.tags import TagsResource
from qhub.resources.versions import VersionsResource

_DEFAULT_BASE_URL = "http://localhost:8080"


class QHubClient:
    """Synchronous client for the QHub API.

    Args:
        bearer_token: Bearer token for authentication.
        base_url: Base URL of the QHub API (default: http://localhost:8080).
        timeout: Request timeout in seconds (default: 30).
        httpx_client: Optional pre-configured httpx.Client to use.

    Example::

        client = QHubClient(bearer_token="my-token")
        orgs = client.organizations.get("my-org")
    """

    def __init__(
        self,
        bearer_token: str,
        *,
        base_url: str = _DEFAULT_BASE_URL,
        timeout: float = 30.0,
        httpx_client: httpx.Client | None = None,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._bearer_token = bearer_token
        self._owns_client = httpx_client is None
        self._http = httpx_client or httpx.Client(timeout=timeout)

        # Resource accessors
        self._organizations: OrganizationsResource | None = None
        self._projects: ProjectsResource | None = None
        self._prompts: PromptsResource | None = None
        self._versions: VersionsResource | None = None
        self._logs: LogsResource | None = None
        self._evaluations: EvaluationsResource | None = None
        self._consulting: ConsultingResource | None = None
        self._tags: TagsResource | None = None
        self._industries: IndustriesResource | None = None
        self._search: SearchResource | None = None
        self._analytics: AnalyticsResource | None = None

    # -- Resource properties (lazy-initialized) ----------------------------

    @property
    def organizations(self) -> OrganizationsResource:
        """Access organization operations."""
        if self._organizations is None:
            self._organizations = OrganizationsResource(self)
        return self._organizations

    @property
    def projects(self) -> ProjectsResource:
        """Access project operations."""
        if self._projects is None:
            self._projects = ProjectsResource(self)
        return self._projects

    @property
    def prompts(self) -> PromptsResource:
        """Access prompt operations."""
        if self._prompts is None:
            self._prompts = PromptsResource(self)
        return self._prompts

    @property
    def versions(self) -> VersionsResource:
        """Access prompt version operations."""
        if self._versions is None:
            self._versions = VersionsResource(self)
        return self._versions

    @property
    def logs(self) -> LogsResource:
        """Access execution log operations."""
        if self._logs is None:
            self._logs = LogsResource(self)
        return self._logs

    @property
    def evaluations(self) -> EvaluationsResource:
        """Access evaluation operations."""
        if self._evaluations is None:
            self._evaluations = EvaluationsResource(self)
        return self._evaluations

    @property
    def consulting(self) -> ConsultingResource:
        """Access consulting session and message operations."""
        if self._consulting is None:
            self._consulting = ConsultingResource(self)
        return self._consulting

    @property
    def tags(self) -> TagsResource:
        """Access tag operations."""
        if self._tags is None:
            self._tags = TagsResource(self)
        return self._tags

    @property
    def industries(self) -> IndustriesResource:
        """Access industry configuration operations."""
        if self._industries is None:
            self._industries = IndustriesResource(self)
        return self._industries

    @property
    def search(self) -> SearchResource:
        """Access semantic search operations."""
        if self._search is None:
            self._search = SearchResource(self)
        return self._search

    @property
    def analytics(self) -> AnalyticsResource:
        """Access analytics operations."""
        if self._analytics is None:
            self._analytics = AnalyticsResource(self)
        return self._analytics

    # -- HTTP helpers ------------------------------------------------------

    def _request(
        self,
        method: str,
        path: str,
        *,
        json: Any | None = None,
        params: dict[str, str] | None = None,
    ) -> Any:
        """Send an HTTP request to the QHub API.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE).
            path: URL path (e.g. /api/v1/organizations).
            json: Optional JSON body.
            params: Optional query parameters.

        Returns:
            Parsed JSON response, or None for 204 No Content.

        Raises:
            QHubError: On non-2xx response.
        """
        url = self._base_url + path
        headers = {
            "Authorization": f"Bearer {self._bearer_token}",
            "Content-Type": "application/json",
        }

        response = self._http.request(
            method,
            url,
            headers=headers,
            json=json,
            params=params,
        )

        raise_for_status(response.status_code, response.text)

        if response.status_code == 204 or not response.content:
            return None

        return response.json()

    # -- Context manager ---------------------------------------------------

    def close(self) -> None:
        """Close the underlying HTTP client (if owned by this instance)."""
        if self._owns_client:
            self._http.close()

    def __enter__(self) -> QHubClient:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()


class AsyncQHubClient:
    """Asynchronous client for the QHub API.

    Args:
        bearer_token: Bearer token for authentication.
        base_url: Base URL of the QHub API (default: http://localhost:8080).
        timeout: Request timeout in seconds (default: 30).
        httpx_client: Optional pre-configured httpx.AsyncClient to use.

    Example::

        async with AsyncQHubClient(bearer_token="my-token") as client:
            org = await client.organizations.get("my-org")
    """

    def __init__(
        self,
        bearer_token: str,
        *,
        base_url: str = _DEFAULT_BASE_URL,
        timeout: float = 30.0,
        httpx_client: httpx.AsyncClient | None = None,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._bearer_token = bearer_token
        self._owns_client = httpx_client is None
        self._http = httpx_client or httpx.AsyncClient(timeout=timeout)

        # Resource accessors (lazy)
        self._organizations: OrganizationsResource | None = None
        self._projects: ProjectsResource | None = None
        self._prompts: PromptsResource | None = None
        self._versions: VersionsResource | None = None
        self._logs: LogsResource | None = None
        self._evaluations: EvaluationsResource | None = None
        self._consulting: ConsultingResource | None = None
        self._tags: TagsResource | None = None
        self._industries: IndustriesResource | None = None
        self._search: SearchResource | None = None
        self._analytics: AnalyticsResource | None = None

    @property
    def organizations(self) -> OrganizationsResource:
        """Access organization operations."""
        if self._organizations is None:
            self._organizations = OrganizationsResource(self)  # type: ignore[arg-type]
        return self._organizations

    @property
    def projects(self) -> ProjectsResource:
        """Access project operations."""
        if self._projects is None:
            self._projects = ProjectsResource(self)  # type: ignore[arg-type]
        return self._projects

    @property
    def prompts(self) -> PromptsResource:
        """Access prompt operations."""
        if self._prompts is None:
            self._prompts = PromptsResource(self)  # type: ignore[arg-type]
        return self._prompts

    @property
    def versions(self) -> VersionsResource:
        """Access prompt version operations."""
        if self._versions is None:
            self._versions = VersionsResource(self)  # type: ignore[arg-type]
        return self._versions

    @property
    def logs(self) -> LogsResource:
        """Access execution log operations."""
        if self._logs is None:
            self._logs = LogsResource(self)  # type: ignore[arg-type]
        return self._logs

    @property
    def evaluations(self) -> EvaluationsResource:
        """Access evaluation operations."""
        if self._evaluations is None:
            self._evaluations = EvaluationsResource(self)  # type: ignore[arg-type]
        return self._evaluations

    @property
    def consulting(self) -> ConsultingResource:
        """Access consulting session and message operations."""
        if self._consulting is None:
            self._consulting = ConsultingResource(self)  # type: ignore[arg-type]
        return self._consulting

    @property
    def tags(self) -> TagsResource:
        """Access tag operations."""
        if self._tags is None:
            self._tags = TagsResource(self)  # type: ignore[arg-type]
        return self._tags

    @property
    def industries(self) -> IndustriesResource:
        """Access industry configuration operations."""
        if self._industries is None:
            self._industries = IndustriesResource(self)  # type: ignore[arg-type]
        return self._industries

    @property
    def search(self) -> SearchResource:
        """Access semantic search operations."""
        if self._search is None:
            self._search = SearchResource(self)  # type: ignore[arg-type]
        return self._search

    @property
    def analytics(self) -> AnalyticsResource:
        """Access analytics operations."""
        if self._analytics is None:
            self._analytics = AnalyticsResource(self)  # type: ignore[arg-type]
        return self._analytics

    # -- HTTP helpers ------------------------------------------------------

    async def _request(
        self,
        method: str,
        path: str,
        *,
        json: Any | None = None,
        params: dict[str, str] | None = None,
    ) -> Any:
        """Send an async HTTP request to the QHub API.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE).
            path: URL path (e.g. /api/v1/organizations).
            json: Optional JSON body.
            params: Optional query parameters.

        Returns:
            Parsed JSON response, or None for 204 No Content.

        Raises:
            QHubError: On non-2xx response.
        """
        url = self._base_url + path
        headers = {
            "Authorization": f"Bearer {self._bearer_token}",
            "Content-Type": "application/json",
        }

        response = await self._http.request(
            method,
            url,
            headers=headers,
            json=json,
            params=params,
        )

        raise_for_status(response.status_code, response.text)

        if response.status_code == 204 or not response.content:
            return None

        return response.json()

    # -- Context manager ---------------------------------------------------

    async def close(self) -> None:
        """Close the underlying HTTP client (if owned by this instance)."""
        if self._owns_client:
            await self._http.aclose()

    async def __aenter__(self) -> AsyncQHubClient:
        return self

    async def __aexit__(self, *args: object) -> None:
        await self.close()
