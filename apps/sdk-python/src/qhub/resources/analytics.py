"""Analytics resource."""

from __future__ import annotations

from typing import TYPE_CHECKING

from qhub.types import (
    DailyTrend,
    ProjectAnalytics,
    PromptAnalytics,
    VersionAnalytics,
)

if TYPE_CHECKING:
    from qhub.client import QHubClient


class AnalyticsResource:
    """Access analytics for prompts, versions, and projects."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def prompt(self, prompt_id: str) -> list[PromptAnalytics]:
        """Get analytics for a prompt across all versions.

        Args:
            prompt_id: The prompt UUID.

        Returns:
            A list of per-version analytics.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/analytics"
        )
        return [PromptAnalytics.model_validate(item) for item in data]

    def version(self, prompt_id: str, version: int) -> VersionAnalytics:
        """Get analytics for a specific prompt version.

        Args:
            prompt_id: The prompt UUID.
            version: The version number.

        Returns:
            The version analytics.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/versions/{version}/analytics"
        )
        return VersionAnalytics.model_validate(data)

    def project(self, project_id: str) -> list[ProjectAnalytics]:
        """Get analytics for all prompts in a project.

        Args:
            project_id: The project UUID.

        Returns:
            A list of per-prompt analytics.
        """
        data = self._client._request(
            "GET", f"/api/v1/projects/{project_id}/analytics"
        )
        return [ProjectAnalytics.model_validate(item) for item in data]

    def daily_trend(self, prompt_id: str) -> list[DailyTrend]:
        """Get daily execution trends for a prompt.

        Args:
            prompt_id: The prompt UUID.

        Returns:
            A list of daily trend data points.
        """
        data = self._client._request(
            "GET", f"/api/v1/prompts/{prompt_id}/trend"
        )
        return [DailyTrend.model_validate(item) for item in data]
