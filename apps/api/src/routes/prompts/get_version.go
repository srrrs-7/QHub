package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *PromptHandler) GetVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		pid := prompt.PromptIDFromUUID(promptID)
		versionParam := chi.URLParam(r, "version")

		switch versionParam {
		case "latest":
			v, err := h.versionRepo.FindLatest(r.Context(), pid)
			if err != nil {
				response.HandleError(w, err)
				return
			}
			response.OK(w, toVersionResponse(v))
			return
		case "production":
			v, err := h.versionRepo.FindProduction(r.Context(), pid)
			if err != nil {
				response.HandleError(w, err)
				return
			}
			response.OK(w, toVersionResponse(v))
			return
		}

		number, err := strconv.Atoi(versionParam)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		v, err := h.versionRepo.FindByPromptAndNumber(r.Context(), pid, number)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toVersionResponse(v))
	}
}
