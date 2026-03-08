package prompts

import (
	"api/src/domain/prompt"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *PromptHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requtil.ParseUUID(r, "project_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		slug, err := prompt.NewPromptSlug(chi.URLParam(r, "prompt_slug"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		p, err := h.promptRepo.FindByProjectAndSlug(r.Context(), projectID, slug)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toPromptResponse(p))
	}
}
