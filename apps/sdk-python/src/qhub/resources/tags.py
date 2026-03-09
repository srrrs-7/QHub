"""Tags resource."""

from __future__ import annotations

from typing import TYPE_CHECKING

from qhub.types import CreateTagRequest, Tag

if TYPE_CHECKING:
    from qhub.client import QHubClient


class TagsResource:
    """Manage tags and prompt-tag associations."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list(self) -> list[Tag]:
        """List all tags.

        Returns:
            A list of tags.
        """
        data = self._client._request("GET", "/api/v1/tags")
        return [Tag.model_validate(item) for item in data]

    def create(self, *, org_id: str, name: str, color: str) -> Tag:
        """Create a new tag.

        Args:
            org_id: The organization UUID.
            name: Tag name (1-100 chars).
            color: Tag color (1-20 chars).

        Returns:
            The created tag.
        """
        body = CreateTagRequest(org_id=org_id, name=name, color=color)
        data = self._client._request(
            "POST", "/api/v1/tags", json=body.model_dump()
        )
        return Tag.model_validate(data)

    def delete(self, tag_id: str) -> None:
        """Delete a tag.

        Args:
            tag_id: The tag UUID.
        """
        self._client._request("DELETE", f"/api/v1/tags/{tag_id}")

    def list_by_prompt(self, prompt_id: str) -> list[Tag]:
        """List tags attached to a prompt.

        Args:
            prompt_id: The prompt UUID.

        Returns:
            A list of tags.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/tags"
        )
        return [Tag.model_validate(item) for item in data]

    def add_to_prompt(self, prompt_id: str, *, tag_id: str) -> None:
        """Add a tag to a prompt.

        Args:
            prompt_id: The prompt UUID.
            tag_id: The tag UUID.
        """
        self._client._request(
            "POST",
            f"/api/v1/prompts/{prompt_id}/tags",
            json={"tag_id": tag_id},
        )

    def remove_from_prompt(self, prompt_id: str, *, tag_id: str) -> None:
        """Remove a tag from a prompt.

        Args:
            prompt_id: The prompt UUID.
            tag_id: The tag UUID.
        """
        self._client._request(
            "DELETE", f"/api/v1/prompts/{prompt_id}/tags/{tag_id}"
        )
