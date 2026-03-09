package organizations

import (
	"api/src/routes/response"
	"net/http"
)

func (h *OrganizationHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgs, err := h.repo.FindAll(r.Context())
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(orgs, toOrganizationResponse))
	}
}
