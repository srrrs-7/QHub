package organizations

import (
	"api/src/domain/organization"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *OrganizationHandler) Put() http.HandlerFunc {
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

		req, err := decodePutRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := requtil.MergeField(org.Name, req.Name, organization.NewOrganizationName)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		newSlug, err := requtil.MergeField(org.Slug, req.Slug, organization.NewOrganizationSlug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		plan, err := requtil.MergeField(org.Plan, req.Plan, organization.NewPlan)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		updated, err := h.repo.Update(r.Context(), org.ID, organization.NewOrganizationCmd(name, newSlug, plan))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toOrganizationResponse(updated))
	}
}
