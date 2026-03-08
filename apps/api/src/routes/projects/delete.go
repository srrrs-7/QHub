package projects

import (
	"api/src/domain/project"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *ProjectHandler) Delete() http.HandlerFunc {
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

		existing, err := h.repo.FindByOrgAndSlug(r.Context(), orgID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		if err := h.repo.Delete(r.Context(), existing.ID); err != nil {
			response.HandleError(w, err)
			return
		}

		response.NoContent(w)
	}
}
