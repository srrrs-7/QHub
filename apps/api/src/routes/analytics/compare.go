package analytics

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// CompareVersions returns a statistical comparison (Welch's t-test) between
// two prompt versions' execution metrics (latency, tokens, overall score).
func (h *AnalyticsHandler) CompareVersions() http.HandlerFunc {
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

		result, err := h.stats.CompareVersions(r.Context(), promptID, v1, v2)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, result)
	}
}
