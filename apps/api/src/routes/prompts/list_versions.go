package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *PromptHandler) ListVersions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		versions, err := h.versionRepo.FindAllByPrompt(r.Context(), prompt.PromptIDFromUUID(promptID))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(versions, toVersionResponse))
	}
}
