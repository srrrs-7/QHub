package projects

import (
	"api/src/domain/project"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *ProjectHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := project.NewProjectSlug(chi.URLParam(r, "project_slug"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		p, err := h.repo.FindByOrgAndSlug(r.Context(), orgID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toProjectResponse(p))
	}
}
