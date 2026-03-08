package prompts

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetTextDiff returns an HTTP handler for GET /prompts/{prompt_id}/versions/{version}/text-diff?from={from_version}.
// It produces a line-by-line text diff between two prompt versions.
func (h *PromptHandler) GetTextDiff() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		toVersion, err := strconv.Atoi(chi.URLParam(r, "version"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		fromVersionStr := r.URL.Query().Get("from")
		if fromVersionStr == "" {
			fromVersionStr = strconv.Itoa(toVersion - 1)
		}
		fromVersion, err := strconv.Atoi(fromVersionStr)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		diff, err := h.diffService.GenerateTextDiff(r.Context(), promptID, fromVersion, toVersion)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, diff)
	}
}
