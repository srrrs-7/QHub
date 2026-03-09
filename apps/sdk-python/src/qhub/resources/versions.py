"""Prompt Versions resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from qhub.types import (
    CreateVersionRequest,
    LintResult,
    PromptVersion,
    SemanticDiff,
    TextDiff,
    UpdateVersionStatusRequest,
)

if TYPE_CHECKING:
    from qhub.client import QHubClient


class VersionsResource:
    """Manage prompt versions."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list(self, prompt_id: str) -> list[PromptVersion]:
        """List all versions of a prompt.

        Args:
            prompt_id: The prompt UUID.

        Returns:
            A list of prompt versions.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/versions"
        )
        return [PromptVersion.model_validate(item) for item in data]

    def get(self, prompt_id: str, version: int) -> PromptVersion:
        """Get a specific prompt version.

        Args:
            prompt_id: The prompt UUID.
            version: The version number.

        Returns:
            The prompt version.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/versions/{version}"
        )
        return PromptVersion.model_validate(data)

    def create(
        self,
        prompt_id: str,
        *,
        content: Any,
        author_id: str,
        variables: Any = None,
        change_description: str = "",
    ) -> PromptVersion:
        """Create a new version of a prompt.

        Args:
            prompt_id: The prompt UUID.
            content: The prompt content (JSON-serializable).
            author_id: UUID of the author.
            variables: Optional variables (JSON-serializable).
            change_description: Optional change description (max 500 chars).

        Returns:
            The created prompt version.
        """
        body = CreateVersionRequest(
            content=content,
            author_id=author_id,
            variables=variables,
            change_description=change_description,
        )
        data = self._client._request(
            "POST",
            f"/api/v1/prompts/{prompt_id}/versions",
            json=body.model_dump(),
        )
        return PromptVersion.model_validate(data)

    def update_status(
        self, prompt_id: str, version: int, *, status: str
    ) -> PromptVersion:
        """Update a version's status.

        Args:
            prompt_id: The prompt UUID.
            version: The version number.
            status: New status (one of: draft, review, production, archived).

        Returns:
            The updated prompt version.
        """
        body = UpdateVersionStatusRequest(status=status)
        data = self._client._request(
            "PUT",
            f"/api/v1/prompts/{prompt_id}/versions/{version}/status",
            json=body.model_dump(),
        )
        return PromptVersion.model_validate(data)

    def lint(self, prompt_id: str, version: int) -> LintResult:
        """Lint a prompt version.

        Args:
            prompt_id: The prompt UUID.
            version: The version number.

        Returns:
            The lint result.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/versions/{version}/lint"
        )
        return LintResult.model_validate(data)

    def text_diff(self, prompt_id: str, version: int) -> TextDiff:
        """Get a text diff for a prompt version against its predecessor.

        Args:
            prompt_id: The prompt UUID.
            version: The version number.

        Returns:
            The text diff result.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/versions/{version}/text-diff"
        )
        return TextDiff.model_validate(data)

    def semantic_diff(self, prompt_id: str, v1: int, v2: int) -> SemanticDiff:
        """Get a semantic diff between two prompt versions.

        Args:
            prompt_id: The prompt UUID.
            v1: The first version number.
            v2: The second version number.

        Returns:
            The semantic diff result.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/semantic-diff/{v1}/{v2}"
        )
        return SemanticDiff.model_validate(data)
