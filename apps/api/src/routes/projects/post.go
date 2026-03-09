package projects

import (
	"api/src/domain/project"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *ProjectHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		orgID, err := requtil.ParseUUID(r, "org_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := project.NewProjectName(req.Name)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := project.NewProjectSlug(req.Slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		desc, err := project.NewProjectDescription(req.Description)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		p, err := h.repo.Create(r.Context(), project.NewProjectCmd(orgID, name, slug, desc))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toProjectResponse(p))
	}
}
