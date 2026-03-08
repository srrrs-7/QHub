package repoerr

import (
	"api/src/domain/apperror"
	"context"
	"database/sql"
	"errors"
)

// Handle maps database errors to domain errors.
// For queries that return a single row, pass entity name to get NotFoundError on ErrNoRows.
// For queries that return multiple rows or no result, pass empty string.
func Handle(err error, repoName, entity string) error {
	if errors.Is(err, sql.ErrNoRows) && entity != "" {
		return apperror.NewNotFoundError(err, entity)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperror.NewInternalServerError(err, repoName)
	}
	return apperror.NewDatabaseError(err, repoName)
}
