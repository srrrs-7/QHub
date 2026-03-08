package projects

import (
	"api/src/domain/project"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ProjectHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		slugParam := chi.URLParam(r, "project_slug")

		req, err := newGetRequest(orgID, slugParam)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		parsedOrgID, err := uuid.Parse(req.OrgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := project.NewProjectSlug(req.Slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		p, err := h.repo.FindByOrgAndSlug(r.Context(), parsedOrgID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toProjectResponse(p))
	}
}
