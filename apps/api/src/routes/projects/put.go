package projects

import (
	"api/src/domain/project"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *ProjectHandler) Put() http.HandlerFunc {
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

		req, err := decodePutRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := requtil.MergeField(existing.Name, req.Name, project.NewProjectName)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		newSlug, err := requtil.MergeField(existing.Slug, req.Slug, project.NewProjectSlug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		desc, err := requtil.MergeField(existing.Description, req.Description, project.NewProjectDescription)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		updated, err := h.repo.Update(r.Context(), existing.ID, project.NewProjectCmd(orgID, name, newSlug, desc))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toProjectResponse(updated))
	}
}
