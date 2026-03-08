package analytics

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *AnalyticsHandler) GetPromptAnalytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		rows, err := h.q.GetPromptAnalytics(r.Context(), promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(rows, toPromptAnalyticsResponse))
	}
}
