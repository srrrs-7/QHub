"""Evaluations resource."""

from __future__ import annotations

from typing import TYPE_CHECKING

from qhub.types import CreateEvaluationRequest, Evaluation

if TYPE_CHECKING:
    from qhub.client import QHubClient


class EvaluationsResource:
    """Manage prompt execution evaluations."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def get(self, evaluation_id: str) -> Evaluation:
        """Get an evaluation by ID.

        Args:
            evaluation_id: The evaluation UUID.

        Returns:
            The evaluation.
        """
        data = self._client._request(
            "GET", f"/api/v1/evaluations/{evaluation_id}"
        )
        return Evaluation.model_validate(data)

    def create(self, request: CreateEvaluationRequest) -> Evaluation:
        """Create an evaluation.

        Args:
            request: The evaluation creation request.

        Returns:
            The created evaluation.
        """
        data = self._client._request(
            "POST", "/api/v1/evaluations", json=request.model_dump(exclude_none=True)
        )
        return Evaluation.model_validate(data)

    def list_by_log(self, log_id: str) -> list[Evaluation]:
        """List evaluations for a specific execution log.

        Args:
            log_id: The execution log UUID.

        Returns:
            A list of evaluations.
        """
        data = self._client._request(
            "GET", f"/api/v1/logs/{log_id}/evaluations"
        )
        return [Evaluation.model_validate(item) for item in data]
