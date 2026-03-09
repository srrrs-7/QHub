"""Consulting resource."""

from __future__ import annotations

from typing import TYPE_CHECKING

from qhub.types import (
    ConsultingMessage,
    ConsultingSession,
    CreateMessageRequest,
    CreateSessionRequest,
)

if TYPE_CHECKING:
    from qhub.client import QHubClient


class ConsultingResource:
    """Manage consulting sessions and messages."""

    def __init__(self, client: QHubClient) -> None:
        self._client = client

    def list_sessions(self) -> list[ConsultingSession]:
        """List all consulting sessions.

        Returns:
            A list of consulting sessions.
        """
        data = self._client._request("GET", "/api/v1/consulting/sessions")
        return [ConsultingSession.model_validate(item) for item in data]

    def get_session(self, session_id: str) -> ConsultingSession:
        """Get a consulting session by ID.

        Args:
            session_id: The session UUID.

        Returns:
            The consulting session.
        """
        data = self._client._request(
            "GET", f"/api/v1/consulting/sessions/{session_id}"
        )
        return ConsultingSession.model_validate(data)

    def create_session(
        self,
        *,
        org_id: str,
        title: str,
        industry_config_id: str = "",
    ) -> ConsultingSession:
        """Create a new consulting session.

        Args:
            org_id: The organization UUID.
            title: Session title (1-200 chars).
            industry_config_id: Optional industry config UUID.

        Returns:
            The created consulting session.
        """
        body = CreateSessionRequest(
            org_id=org_id, title=title, industry_config_id=industry_config_id
        )
        data = self._client._request(
            "POST", "/api/v1/consulting/sessions", json=body.model_dump()
        )
        return ConsultingSession.model_validate(data)

    def list_messages(self, session_id: str) -> list[ConsultingMessage]:
        """List messages in a consulting session.

        Args:
            session_id: The session UUID.

        Returns:
            A list of messages.
        """
        data = self._client._request(
            "GET", f"/api/v1/consulting/sessions/{session_id}/messages"
        )
        return [ConsultingMessage.model_validate(item) for item in data]

    def create_message(
        self,
        session_id: str,
        *,
        role: str,
        content: str,
        citations: object = None,
        actions_taken: object = None,
    ) -> ConsultingMessage:
        """Create a message in a consulting session.

        Args:
            session_id: The session UUID.
            role: Message role (one of: user, assistant, system).
            content: The message content.
            citations: Optional citations (JSON-serializable).
            actions_taken: Optional actions taken (JSON-serializable).

        Returns:
            The created message.
        """
        body = CreateMessageRequest(
            role=role,
            content=content,
            citations=citations,
            actions_taken=actions_taken,
        )
        data = self._client._request(
            "POST",
            f"/api/v1/consulting/sessions/{session_id}/messages",
            json=body.model_dump(),
        )
        return ConsultingMessage.model_validate(data)
