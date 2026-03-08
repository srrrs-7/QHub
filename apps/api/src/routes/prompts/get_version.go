package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *PromptHandler) GetVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID := chi.URLParam(r, "prompt_id")
		versionParam := chi.URLParam(r, "version")

		parsedPromptID, err := uuid.Parse(promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Handle special version aliases
		switch versionParam {
		case "latest":
			v, err := h.versionRepo.FindLatest(r.Context(), prompt.PromptIDFromUUID(parsedPromptID))
			if err != nil {
				response.HandleError(w, err)
				return
			}
			response.OK(w, toVersionResponse(v))
			return
		case "production":
			v, err := h.versionRepo.FindProduction(r.Context(), prompt.PromptIDFromUUID(parsedPromptID))
			if err != nil {
				response.HandleError(w, err)
				return
			}
			response.OK(w, toVersionResponse(v))
			return
		}

		// Parse as version number
		number, err := strconv.Atoi(versionParam)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		v, err := h.versionRepo.FindByPromptAndNumber(r.Context(), prompt.PromptIDFromUUID(parsedPromptID), number)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toVersionResponse(v))
	}
}
