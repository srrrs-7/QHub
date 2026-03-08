package prompts

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetDiff returns an HTTP handler for GET /prompts/{prompt_id}/semantic-diff/{v1}/{v2}.
// It generates a semantic diff between two prompt versions.
func (h *PromptHandler) GetDiff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		v1, err := strconv.Atoi(chi.URLParam(r, "v1"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		v2, err := strconv.Atoi(chi.URLParam(r, "v2"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		diff, err := h.diffService.GenerateDiff(r.Context(), promptID, v1, v2)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, diff)
	}
}
