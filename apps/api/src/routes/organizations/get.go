package organizations

import (
	"api/src/domain/organization"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *OrganizationHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug, err := organization.NewOrganizationSlug(chi.URLParam(r, "org_slug"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		org, err := h.repo.FindBySlug(r.Context(), slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toOrganizationResponse(org))
	}
}
