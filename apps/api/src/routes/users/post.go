package users

import (
	"net/http"

	"api/src/domain/apperror"
	"api/src/routes/response"

	"utils/db/db"
)

// Post returns a handler that creates a new user.
func (h *UserHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := newPostRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err = req.sanitizeAndValidate()
		if err != nil {
			response.HandleError(w, err)
			return
		}

		u, err := h.q.CreateUser(r.Context(), db.CreateUserParams{
			Email: req.Email,
			Name:  req.Name,
		})
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "User"))
			return
		}

		response.Created(w, toUserResponse(u))
	}
}
