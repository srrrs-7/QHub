package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) Get() http.HandlerFunc {
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

		p, err := h.promptRepo.FindByProjectAndSlug(r.Context(), parsedProjectID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toPromptResponse(p))
	}
}
