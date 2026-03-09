"""Custom exceptions for the QHub SDK."""

from __future__ import annotations


class QHubError(Exception):
    """Base exception for all QHub SDK errors."""

    def __init__(self, message: str) -> None:
        self.message = message
        super().__init__(message)


class APIError(QHubError):
    """Error returned by the QHub API."""

    def __init__(self, status_code: int, message: str) -> None:
        self.status_code = status_code
        super().__init__(f"API error {status_code}: {message}")


class ValidationError(QHubError):
    """Input validation failed (HTTP 400/422)."""


class NotFoundError(QHubError):
    """Requested resource was not found (HTTP 404)."""


class AuthenticationError(QHubError):
    """Authentication failed (HTTP 401)."""


class ForbiddenError(QHubError):
    """Authorization failed (HTTP 403)."""


class ConflictError(QHubError):
    """Resource conflict (HTTP 409)."""


class RateLimitError(QHubError):
    """Rate limit exceeded (HTTP 429)."""


def raise_for_status(status_code: int, body: str) -> None:
    """Raise an appropriate exception based on HTTP status code.

    Args:
        status_code: The HTTP response status code.
        body: The response body text.

    Raises:
        ValidationError: For 400 or 422 status codes.
        AuthenticationError: For 401 status codes.
        ForbiddenError: For 403 status codes.
        NotFoundError: For 404 status codes.
        ConflictError: For 409 status codes.
        RateLimitError: For 429 status codes.
        APIError: For any other non-2xx status codes.
    """
    if 200 <= status_code < 300:
        return

    error_classes: dict[int, type[QHubError]] = {
        400: ValidationError,
        401: AuthenticationError,
        403: ForbiddenError,
        404: NotFoundError,
        409: ConflictError,
        422: ValidationError,
        429: RateLimitError,
    }

    cls = error_classes.get(status_code)
    if cls is not None:
        raise cls(body)
    raise APIError(status_code, body)
