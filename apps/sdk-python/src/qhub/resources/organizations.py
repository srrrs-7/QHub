"""Organizations resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from qhub.types import (
    CreateOrganizationRequest,
    Organization,
    UpdateOrganizationRequest,
)

if TYPE_CHECKING:
    from qhub.client import QHubClient


class OrganizationsResource:
    """Manage organizations via the QHub API.

    Organizations are the top-level grouping entity.
    """

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def get(self, org_slug: str) -> Organization:
        """Get an organization by slug.

        Args:
            org_slug: The organization slug.

        Returns:
            The organization.
        """
        data = self._client._request("GET", f"/api/v1/organizations/{org_slug}")
        return Organization.model_validate(data)

    def create(
        self,
        *,
        name: str,
        slug: str,
    ) -> Organization:
        """Create a new organization.

        Args:
            name: The organization name (2-100 chars).
            slug: The organization slug (2-50 chars).

        Returns:
            The created organization.
        """
        body = CreateOrganizationRequest(name=name, slug=slug)
        data = self._client._request(
            "POST", "/api/v1/organizations", json=body.model_dump()
        )
        return Organization.model_validate(data)

    def update(
        self,
        org_slug: str,
        *,
        name: str | None = None,
        slug: str | None = None,
        plan: str | None = None,
    ) -> Organization:
        """Update an organization.

        Args:
            org_slug: The current organization slug.
            name: New name (optional).
            slug: New slug (optional).
            plan: New plan (optional, one of: free, pro, team, enterprise).

        Returns:
            The updated organization.
        """
        body = UpdateOrganizationRequest(name=name, slug=slug, plan=plan)
        payload = body.model_dump(exclude_none=True)
        data = self._client._request(
            "PUT", f"/api/v1/organizations/{org_slug}", json=payload
        )
        return Organization.model_validate(data)
