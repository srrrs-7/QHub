package prompts

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetLint returns an HTTP handler for GET /prompts/{prompt_id}/versions/{version}/lint.
// It lints a specific prompt version and returns the result.
func (h *PromptHandler) GetLint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		versionNumber, err := strconv.Atoi(chi.URLParam(r, "version"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		result, err := h.lintService.LintVersion(r.Context(), promptID, versionNumber)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, result)
	}
}
