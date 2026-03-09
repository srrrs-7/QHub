package users

import (
	"utils/db/db"
)

// UserHandler handles HTTP requests for user endpoints.
type UserHandler struct {
	q db.Querier
}

// NewUserHandler creates a new UserHandler with the given querier.
func NewUserHandler(q db.Querier) *UserHandler {
	return &UserHandler{q: q}
}
