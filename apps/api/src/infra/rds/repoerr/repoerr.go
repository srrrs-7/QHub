// Package repoerr provides a centralised mapping from database errors
// to domain-layer AppError types.
//
// All repository implementations use Handle() to convert sql.ErrNoRows,
// context.DeadlineExceeded, and other database errors into the appropriate
// NotFoundError, InternalServerError, or DatabaseError.
package repoerr

import (
	"api/src/domain/apperror"
	"context"
	"database/sql"
	"errors"
)

// Handle maps a database error to the appropriate domain error.
//
//   - sql.ErrNoRows → NotFoundError  (when entity != "")
//   - context.DeadlineExceeded → InternalServerError
//   - anything else → DatabaseError
//
// For list queries that may legitimately return zero rows, pass entity = ""
// so that ErrNoRows is treated as a database error rather than not-found.
func Handle(err error, repoName, entity string) error {
	if errors.Is(err, sql.ErrNoRows) && entity != "" {
		return apperror.NewNotFoundError(err, entity)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperror.NewInternalServerError(err, repoName)
	}
	return apperror.NewDatabaseError(err, repoName)
}
