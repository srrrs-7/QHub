package projects

import (
	"api/src/domain/project"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *ProjectHandler) Put() http.HandlerFunc {
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

		req, err := decodePutRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name := existing.Name
		if req.Name != "" {
			name, err = project.NewProjectName(req.Name)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		newSlug := existing.Slug
		if req.Slug != "" {
			newSlug, err = project.NewProjectSlug(req.Slug)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		desc := existing.Description
		if req.Description != "" {
			desc, err = project.NewProjectDescription(req.Description)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		cmd := project.NewProjectCmd(parsedOrgID, name, newSlug, desc)

		updated, err := h.repo.Update(r.Context(), existing.ID, cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toProjectResponse(updated))
	}
}
