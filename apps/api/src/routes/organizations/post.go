package organizations

import (
	"api/src/domain/organization"
	"api/src/routes/response"
	"net/http"
)

func (h *OrganizationHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := organization.NewOrganizationName(req.Name)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := organization.NewOrganizationSlug(req.Slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		cmd := organization.NewOrganizationCmd(name, slug, organization.PlanFree)

		org, err := h.repo.Create(r.Context(), cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toOrganizationResponse(org))
	}
}
