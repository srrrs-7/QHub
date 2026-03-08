package projects

import (
	"api/src/domain/project"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ProjectHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")
		slugParam := chi.URLParam(r, "project_slug")

		parsedOrgID, err := uuid.Parse(orgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := project.NewProjectSlug(slugParam)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		existing, err := h.repo.FindByOrgAndSlug(r.Context(), parsedOrgID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		err = h.repo.Delete(r.Context(), existing.ID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.NoContent(w)
	}
}
