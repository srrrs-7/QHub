package apikeys

import (
	"api/src/domain/apperror"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

// List returns all API keys for the given organization.
func (h *ApiKeyHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		keys, err := h.q.ListApiKeysByOrganization(r.Context(), orgID)
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "api_key"))
			return
		}

		response.OK(w, response.MapSlice(keys, toApiKeyResponse))
	}
}
