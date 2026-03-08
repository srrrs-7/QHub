package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) Put() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		slugParam := chi.URLParam(r, "prompt_slug")

		parsedProjectID, err := uuid.Parse(projectID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := prompt.NewPromptSlug(slugParam)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		existing, err := h.promptRepo.FindByProjectAndSlug(r.Context(), parsedProjectID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutPromptRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name := existing.Name
		if req.Name != "" {
			name, err = prompt.NewPromptName(req.Name)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		newSlug := existing.Slug
		if req.Slug != "" {
			newSlug, err = prompt.NewPromptSlug(req.Slug)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		desc := existing.Description
		if req.Description != "" {
			desc, err = prompt.NewPromptDescription(req.Description)
			if err != nil {
				response.HandleError(w, err)
				return
			}
		}

		cmd := prompt.PromptCmd{
			ProjectID:   parsedProjectID,
			Name:        name,
			Slug:        newSlug,
			PromptType:  existing.PromptType,
			Description: desc,
		}

		updated, err := h.promptRepo.Update(r.Context(), existing.ID, cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toPromptResponse(updated))
	}
}
