package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")

		parsedProjectID, err := uuid.Parse(projectID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePostPromptRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		name, err := prompt.NewPromptName(req.Name)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := prompt.NewPromptSlug(req.Slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		promptType, err := prompt.NewPromptType(req.PromptType)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		desc, err := prompt.NewPromptDescription(req.Description)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		cmd := prompt.PromptCmd{
			ProjectID:   parsedProjectID,
			Name:        name,
			Slug:        slug,
			PromptType:  promptType,
			Description: desc,
		}

		p, err := h.promptRepo.Create(r.Context(), cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toPromptResponse(p))
	}
}
