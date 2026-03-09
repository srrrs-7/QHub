"""Execution Logs resource."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from qhub.types import CreateLogRequest, ExecutionLog, LogListResponse

if TYPE_CHECKING:
    from qhub.client import QHubClient


class LogsResource:
    """Manage execution logs."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list(
        self,
        *,
        prompt_id: str | None = None,
        org_id: str | None = None,
        limit: int | None = None,
        offset: int | None = None,
    ) -> LogListResponse:
        """List execution logs with optional filtering.

        Args:
            prompt_id: Filter by prompt UUID (optional).
            org_id: Filter by organization UUID (optional).
            limit: Max results to return (optional).
            offset: Number of results to skip (optional).

        Returns:
            Paginated list of execution logs.
        """
        params: dict[str, str] = {}
        if prompt_id is not None:
            params["prompt_id"] = prompt_id
        if org_id is not None:
            params["org_id"] = org_id
        if limit is not None:
            params["limit"] = str(limit)
        if offset is not None:
            params["offset"] = str(offset)

        data = self._client._request("GET", "/api/v1/logs", params=params)
        return LogListResponse.model_validate(data)

    def get(self, log_id: str) -> ExecutionLog:
        """Get an execution log by ID.

        Args:
            log_id: The log UUID.

        Returns:
            The execution log.
        """
        data = self._client._request("GET", f"/api/v1/logs/{log_id}")
        return ExecutionLog.model_validate(data)

    def create(self, request: CreateLogRequest) -> ExecutionLog:
        """Create an execution log entry.

        Args:
            request: The log creation request.

        Returns:
            The created execution log.
        """
        data = self._client._request(
            "POST", "/api/v1/logs", json=request.model_dump()
        )
        return ExecutionLog.model_validate(data)

    def create_batch(self, requests: list[CreateLogRequest]) -> list[ExecutionLog]:
        """Create multiple execution log entries in a single request.

        Args:
            requests: A list of log creation requests (max 100).

        Returns:
            The created execution logs.
        """
        payload = {"logs": [r.model_dump() for r in requests]}
        data = self._client._request("POST", "/api/v1/logs/batch", json=payload)
        return [ExecutionLog.model_validate(item) for item in data]
