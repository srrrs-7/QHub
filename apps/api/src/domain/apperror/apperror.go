package apperror

import "fmt"

// Error name constants define the canonical names for each error type.
// These are used for error identification and logging.
const (
	NotFoundErrorName       = "NotFoundError"
	ValidationErrorName     = "ValidationError"
	UnauthorizedErrorName   = "UnauthorizedError"
	InternalServerErrorName = "InternalServerError"
	BadRequestErrorName     = "BadRequestError"
	ConflictErrorName       = "ConflictError"
	ForbiddenErrorName      = "ForbiddenError"
	DatabaseErrorName       = "DatabaseError"
)

// AppError is the common error interface for the application.
// It extends the standard error interface with additional methods for
// error categorization and domain identification.
type AppError interface {
	error
	// ErrorName returns the canonical name of the error type.
	ErrorName() string
	// DomainName returns the domain or context where the error occurred.
	DomainName() string
	// Unwrap returns the underlying error, if any.
	Unwrap() error
}

// baseErr is the base error structure that implements AppError.
// All specific error types embed this structure.
type baseErr struct {
	errName    string // The canonical name of the error type
	domainName string // The domain or context where the error occurred
	err        error  // The underlying error, if any
}

// ErrorName returns the canonical name of the error type.
func (e baseErr) ErrorName() string {
	return e.errName
}

// DomainName returns the domain or context where the error occurred.
func (e baseErr) DomainName() string {
	return e.domainName
}

// Error implements the error interface and formats the error message.
// It includes the error name, domain, and underlying error if present.
func (e baseErr) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s [%s]: %v", e.errName, e.domainName, e.err)
	}
	return fmt.Sprintf("%s [%s]", e.errName, e.domainName)
}

// Unwrap returns the underlying error for error chain support.
// This enables the use of errors.Is and errors.As.
func (e baseErr) Unwrap() error {
	return e.err
}

// NotFoundError represents an error when a requested resource is not found.
type NotFoundError struct {
	baseErr
}

// NewNotFoundError creates a new NotFoundError with the given underlying error and domain name.
func NewNotFoundError(err error, dName string) NotFoundError {
	return NotFoundError{
		baseErr: baseErr{
			errName:    NotFoundErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// ValidationError represents an error when input validation fails.
type ValidationError struct {
	baseErr
}

// NewValidationError creates a new ValidationError with the given underlying error and domain name.
func NewValidationError(err error, dName string) ValidationError {
	return ValidationError{
		baseErr: baseErr{
			errName:    ValidationErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// UnauthorizedError represents an error when authentication is required or has failed.
type UnauthorizedError struct {
	baseErr
}

// NewUnauthorizedError creates a new UnauthorizedError with the given underlying error and domain name.
func NewUnauthorizedError(err error, dName string) UnauthorizedError {
	return UnauthorizedError{
		baseErr: baseErr{
			errName:    UnauthorizedErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// InternalServerError represents an unexpected internal server error.
type InternalServerError struct {
	baseErr
}

// NewInternalServerError creates a new InternalServerError with the given underlying error and domain name.
func NewInternalServerError(err error, dName string) InternalServerError {
	return InternalServerError{
		baseErr: baseErr{
			errName:    InternalServerErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// BadRequestError represents an error when the request is malformed or invalid.
type BadRequestError struct {
	baseErr
}

// NewBadRequestError creates a new BadRequestError with the given underlying error and domain name.
func NewBadRequestError(err error, dName string) BadRequestError {
	return BadRequestError{
		baseErr: baseErr{
			errName:    BadRequestErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// ConflictError represents an error when the request conflicts with the current state.
type ConflictError struct {
	baseErr
}

// NewConflictError creates a new ConflictError with the given underlying error and domain name.
func NewConflictError(err error, dName string) ConflictError {
	return ConflictError{
		baseErr: baseErr{
			errName:    ConflictErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// ForbiddenError represents an error when access to a resource is forbidden.
type ForbiddenError struct {
	baseErr
}

// NewForbiddenError creates a new ForbiddenError with the given underlying error and domain name.
func NewForbiddenError(err error, dName string) ForbiddenError {
	return ForbiddenError{
		baseErr: baseErr{
			errName:    ForbiddenErrorName,
			domainName: dName,
			err:        err,
		},
	}
}

// DatabaseError represents an error that occurred during database operations.
type DatabaseError struct {
	baseErr
}

// NewDatabaseError creates a new DatabaseError with the given underlying error and domain name.
func NewDatabaseError(err error, dName string) DatabaseError {
	return DatabaseError{
		baseErr: baseErr{
			errName:    DatabaseErrorName,
			domainName: dName,
			err:        err,
		},
	}
}
