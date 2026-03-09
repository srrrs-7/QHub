"""Industries resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from qhub.types import (
    Benchmark,
    ComplianceCheckResponse,
    CreateIndustryConfigRequest,
    IndustryConfig,
    UpdateIndustryConfigRequest,
)

if TYPE_CHECKING:
    from qhub.client import QHubClient


class IndustriesResource:
    """Manage industry configurations, benchmarks, and compliance checks."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list(self) -> list[IndustryConfig]:
        """List all industry configurations.

        Returns:
            A list of industry configs.
        """
        data = self._client._request("GET", "/api/v1/industries")
        return [IndustryConfig.model_validate(item) for item in data]

    def get(self, slug: str) -> IndustryConfig:
        """Get an industry configuration by slug.

        Args:
            slug: The industry config slug.

        Returns:
            The industry configuration.
        """
        data = self._client._request("GET", f"/api/v1/industries/{slug}")
        return IndustryConfig.model_validate(data)

    def create(
        self,
        *,
        slug: str,
        name: str,
        description: str = "",
        knowledge_base: Any = None,
        compliance_rules: Any = None,
    ) -> IndustryConfig:
        """Create a new industry configuration.

        Args:
            slug: The industry config slug (2-80 chars).
            name: The industry name (1-200 chars).
            description: Optional description (max 1000 chars).
            knowledge_base: Optional knowledge base data (JSON-serializable).
            compliance_rules: Optional compliance rules (JSON-serializable).

        Returns:
            The created industry configuration.
        """
        body = CreateIndustryConfigRequest(
            slug=slug,
            name=name,
            description=description,
            knowledge_base=knowledge_base,
            compliance_rules=compliance_rules,
        )
        data = self._client._request(
            "POST", "/api/v1/industries", json=body.model_dump()
        )
        return IndustryConfig.model_validate(data)

    def update(
        self,
        slug: str,
        *,
        name: str | None = None,
        description: str | None = None,
        knowledge_base: Any = None,
        compliance_rules: Any = None,
    ) -> IndustryConfig:
        """Update an industry configuration.

        Args:
            slug: The industry config slug.
            name: New name (optional).
            description: New description (optional).
            knowledge_base: New knowledge base data (optional).
            compliance_rules: New compliance rules (optional).

        Returns:
            The updated industry configuration.
        """
        body = UpdateIndustryConfigRequest(
            name=name,
            description=description,
            knowledge_base=knowledge_base,
            compliance_rules=compliance_rules,
        )
        payload = body.model_dump(exclude_none=True)
        data = self._client._request(
            "PUT", f"/api/v1/industries/{slug}", json=payload
        )
        return IndustryConfig.model_validate(data)

    def list_benchmarks(self, slug: str) -> list[Benchmark]:
        """List benchmarks for an industry.

        Args:
            slug: The industry config slug.

        Returns:
            A list of benchmarks.
        """
        data = self._client._request(
            "GET", f"/api/v1/industries/{slug}/benchmarks"
        )
        return [Benchmark.model_validate(item) for item in data]

    def compliance_check(self, slug: str, *, content: str) -> ComplianceCheckResponse:
        """Run a compliance check against an industry's rules.

        Args:
            slug: The industry config slug.
            content: The prompt content to check.

        Returns:
            The compliance check result.
        """
        data = self._client._request(
            "POST",
            f"/api/v1/industries/{slug}/compliance-check",
            json={"content": content},
        )
        return ComplianceCheckResponse.model_validate(data)
