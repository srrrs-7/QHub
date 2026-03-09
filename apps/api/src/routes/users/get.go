package users

import (
	"database/sql"
	"errors"
	"net/http"

	"api/src/domain/apperror"
	"api/src/routes/response"

	"github.com/google/uuid"
)

// Get returns a handler that retrieves a user by ID.
func (h *UserHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := newGetRequest(r)
		if _, err := req.validate(); err != nil {
			response.HandleError(w, err)
			return
		}

		id, err := uuid.Parse(req.ID)
		if err != nil {
			response.HandleError(w, apperror.NewValidationError(err, "UserID"))
			return
		}

		u, err := h.q.GetUser(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.HandleError(w, apperror.NewNotFoundError(err, "User"))
				return
			}
			response.HandleError(w, apperror.NewDatabaseError(err, "User"))
			return
		}

		response.OK(w, toUserResponse(u))
	}
}
