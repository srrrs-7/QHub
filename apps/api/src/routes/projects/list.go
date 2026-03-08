package projects

import (
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ProjectHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "org_id")

		req, err := newListRequest(orgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		parsedOrgID, err := uuid.Parse(req.OrgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		projects, err := h.repo.FindAllByOrg(r.Context(), parsedOrgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(projects, toProjectResponse))
	}
}
