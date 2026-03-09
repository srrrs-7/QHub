"""Prompts resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from qhub.types import CreatePromptRequest, Prompt, UpdatePromptRequest

if TYPE_CHECKING:
    from qhub.client import QHubClient


class PromptsResource:
    """Manage prompts within a project."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list(self, project_id: str) -> list[Prompt]:
        """List all prompts in a project.

        Args:
            project_id: The project UUID.

        Returns:
            A list of prompts.
        """
        data = self._client._request(
            "GET", f"/api/v1/projects/{project_id}/prompts"
        )
        return [Prompt.model_validate(item) for item in data]

    def get(self, project_id: str, prompt_slug: str) -> Prompt:
        """Get a prompt by slug.

        Args:
            project_id: The project UUID.
            prompt_slug: The prompt slug.

        Returns:
            The prompt.
        """
        data = self._client._request(
            "GET", f"/api/v1/projects/{project_id}/prompts/{prompt_slug}"
        )
        return Prompt.model_validate(data)

    def create(
        self,
        project_id: str,
        *,
        name: str,
        slug: str,
        prompt_type: str,
        description: str = "",
    ) -> Prompt:
        """Create a new prompt.

        Args:
            project_id: The project UUID.
            name: The prompt name (2-200 chars).
            slug: The prompt slug (2-80 chars).
            prompt_type: One of: system, user, combined.
            description: Optional description (max 1000 chars).

        Returns:
            The created prompt.
        """
        body = CreatePromptRequest(
            name=name, slug=slug, prompt_type=prompt_type, description=description
        )
        data = self._client._request(
            "POST",
            f"/api/v1/projects/{project_id}/prompts",
            json=body.model_dump(),
        )
        return Prompt.model_validate(data)

    def update(
        self,
        project_id: str,
        prompt_slug: str,
        *,
        name: str | None = None,
        slug: str | None = None,
        description: str | None = None,
    ) -> Prompt:
        """Update a prompt.

        Args:
            project_id: The project UUID.
            prompt_slug: The current prompt slug.
            name: New name (optional).
            slug: New slug (optional).
            description: New description (optional).

        Returns:
            The updated prompt.
        """
        body = UpdatePromptRequest(name=name, slug=slug, description=description)
        payload = body.model_dump(exclude_none=True)
        data = self._client._request(
            "PUT",
            f"/api/v1/projects/{project_id}/prompts/{prompt_slug}",
            json=payload,
        )
        return Prompt.model_validate(data)
