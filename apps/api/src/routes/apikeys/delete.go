package apikeys

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"utils/db/db"
)

// Delete revokes an API key by its ID.
func (h *ApiKeyHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		_, err = h.q.RevokeApiKey(r.Context(), db.RevokeApiKeyParams{
			ID:             id,
			OrganizationID: orgID,
		})
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "api_key"))
			return
		}

		response.NoContent(w)
	}
}
