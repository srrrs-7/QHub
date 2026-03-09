/** Base error class for all QHub API errors. */
export class QHubError extends Error {
  /** HTTP status code from the API response. */
  readonly statusCode: number;
  /** Raw response body text. */
  readonly body: string;

  constructor(statusCode: number, body: string) {
    super(`QHub API error ${statusCode}: ${body}`);
    this.name = "QHubError";
    this.statusCode = statusCode;
    this.body = body;
  }
}

/** Thrown when the request is invalid (400). */
export class ValidationError extends QHubError {
  constructor(body: string) {
    super(400, body);
    this.name = "ValidationError";
  }
}

/** Thrown when authentication fails (401). */
export class AuthenticationError extends QHubError {
  constructor(body: string) {
    super(401, body);
    this.name = "AuthenticationError";
  }
}

/** Thrown when the user is not authorized (403). */
export class ForbiddenError extends QHubError {
  constructor(body: string) {
    super(403, body);
    this.name = "ForbiddenError";
  }
}

/** Thrown when the requested resource is not found (404). */
export class NotFoundError extends QHubError {
  constructor(body: string) {
    super(404, body);
    this.name = "NotFoundError";
  }
}

/** Thrown when a conflict occurs (409). */
export class ConflictError extends QHubError {
  constructor(body: string) {
    super(409, body);
    this.name = "ConflictError";
  }
}

/** Thrown when rate limited (429). */
export class RateLimitError extends QHubError {
  constructor(body: string) {
    super(429, body);
    this.name = "RateLimitError";
  }
}

/** Thrown when the server encounters an internal error (500+). */
export class InternalServerError extends QHubError {
  constructor(statusCode: number, body: string) {
    super(statusCode, body);
    this.name = "InternalServerError";
  }
}

/**
 * Creates the appropriate typed error from an HTTP status code and response body.
 */
export function createErrorFromStatus(
  statusCode: number,
  body: string,
): QHubError {
  switch (statusCode) {
    case 400:
      return new ValidationError(body);
    case 401:
      return new AuthenticationError(body);
    case 403:
      return new ForbiddenError(body);
    case 404:
      return new NotFoundError(body);
    case 409:
      return new ConflictError(body);
    case 429:
      return new RateLimitError(body);
    default:
      if (statusCode >= 500) {
        return new InternalServerError(statusCode, body);
      }
      return new QHubError(statusCode, body);
  }
}
