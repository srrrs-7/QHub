"""Search resource."""

from __future__ import annotations

from typing import TYPE_CHECKING

from qhub.types import SearchResponse, SemanticSearchRequest

if TYPE_CHECKING:
    from qhub.client import QHubClient


class SearchResource:
    """Semantic search across prompt versions."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def semantic(
        self,
        *,
        query: str,
        org_id: str,
        limit: int = 10,
        min_score: float = 0.0,
    ) -> SearchResponse:
        """Search prompt versions by semantic similarity.

        Args:
            query: The search query text.
            org_id: The organization UUID to search within.
            limit: Max results (1-50, default 10).
            min_score: Minimum similarity score threshold.

        Returns:
            The search response with matching results.
        """
        body = SemanticSearchRequest(
            query=query, org_id=org_id, limit=limit, min_score=min_score
        )
        data = self._client._request(
            "POST", "/api/v1/search/semantic", json=body.model_dump()
        )
        return SearchResponse.model_validate(data)

    def embedding_status(self) -> dict[str, str]:
        """Check the embedding service status.

        Returns:
            A dict with the embedding service status.
        """
        data = self._client._request("GET", "/api/v1/search/embedding-status")
        return data
