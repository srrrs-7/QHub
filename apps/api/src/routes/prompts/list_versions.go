package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) ListVersions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")

		parsedPromptID, err := uuid.Parse(promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		versions, err := h.versionRepo.FindAllByPrompt(r.Context(), prompt.PromptIDFromUUID(parsedPromptID))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(versions, toVersionResponse))
	}
}
