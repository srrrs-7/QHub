package organizations

import (
	"api/src/domain/organization"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *OrganizationHandler) Put() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slugParam := chi.URLParam(r, "org_slug")

		slug, err := organization.NewOrganizationSlug(slugParam)
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

		// Use existing values if not provided in request
		name := org.Name
		if req.Name != "" {
			name, err = organization.NewOrganizationName(req.Name)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		newSlug := org.Slug
		if req.Slug != "" {
			newSlug, err = organization.NewOrganizationSlug(req.Slug)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		plan := org.Plan
		if req.Plan != "" {
			plan, err = organization.NewPlan(req.Plan)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		cmd := organization.NewOrganizationCmd(name, newSlug, plan)

		updated, err := h.repo.Update(r.Context(), org.ID, cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toOrganizationResponse(updated))
	}
}
