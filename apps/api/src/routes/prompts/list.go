package prompts

import (
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")

		parsedProjectID, err := uuid.Parse(projectID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		prompts, err := h.promptRepo.FindAllByProject(r.Context(), parsedProjectID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(prompts, toPromptResponse))
	}
}
