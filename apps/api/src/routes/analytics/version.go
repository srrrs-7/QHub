package analytics

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	db "utils/db/db"
)

func (h *AnalyticsHandler) GetVersionAnalytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		version, err := strconv.Atoi(chi.URLParam(r, "version"))
		if err != nil {
			response.HandleError(w, err)
			return
		}

		row, err := h.q.GetPromptVersionAnalytics(r.Context(), db.GetPromptVersionAnalyticsParams{
			PromptID:      promptID,
			VersionNumber: int32(version),
		})
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toVersionAnalyticsResponse(row))
	}
}
