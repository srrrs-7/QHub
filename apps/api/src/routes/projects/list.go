package projects

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *ProjectHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		projects, err := h.repo.FindAllByOrg(r.Context(), orgID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(projects, toProjectResponse))
	}
}
