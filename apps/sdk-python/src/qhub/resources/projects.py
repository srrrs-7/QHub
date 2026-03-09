"""Projects resource."""

from __future__ import annotations

from typing import TYPE_CHECKING

from qhub.types import CreateProjectRequest, Project, UpdateProjectRequest

if TYPE_CHECKING:
    from qhub.client import QHubClient


class ProjectsResource:
    """Manage projects within an organization."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list(self, org_id: str) -> list[Project]:
        """List all projects in an organization.

        Args:
            org_id: The organization UUID.

        Returns:
            A list of projects.
        """
        data = self._client._request(
            "GET", f"/api/v1/organizations/{org_id}/projects"
        )
        return [Project.model_validate(item) for item in data]

    def get(self, org_id: str, project_slug: str) -> Project:
        """Get a project by slug.

        Args:
            org_id: The organization UUID.
            project_slug: The project slug.

        Returns:
            The project.
        """
        data = self._client._request(
            "GET", f"/api/v1/organizations/{org_id}/projects/{project_slug}"
        )
        return Project.model_validate(data)

    def create(
        self,
        org_id: str,
        *,
        name: str,
        slug: str,
        description: str = "",
    ) -> Project:
        """Create a new project.

        Args:
            org_id: The organization UUID.
            name: The project name (2-100 chars).
            slug: The project slug (2-50 chars).
            description: Optional description (max 500 chars).

        Returns:
            The created project.
        """
        body = CreateProjectRequest(
            organization_id=org_id, name=name, slug=slug, description=description
        )
        data = self._client._request(
            "POST", f"/api/v1/organizations/{org_id}/projects", json=body.model_dump()
        )
        return Project.model_validate(data)

    def update(
        self,
        org_id: str,
        project_slug: str,
        *,
        name: str | None = None,
        slug: str | None = None,
        description: str | None = None,
    ) -> Project:
        """Update a project.

        Args:
            org_id: The organization UUID.
            project_slug: The current project slug.
            name: New name (optional).
            slug: New slug (optional).
            description: New description (optional).

        Returns:
            The updated project.
        """
        body = UpdateProjectRequest(name=name, slug=slug, description=description)
        payload = body.model_dump(exclude_none=True)
        data = self._client._request(
            "PUT",
            f"/api/v1/organizations/{org_id}/projects/{project_slug}",
            json=payload,
        )
        return Project.model_validate(data)

    def delete(self, org_id: str, project_slug: str) -> None:
        """Delete a project.

        Args:
            org_id: The organization UUID.
            project_slug: The project slug.
        """
        self._client._request(
            "DELETE", f"/api/v1/organizations/{org_id}/projects/{project_slug}"
        )
