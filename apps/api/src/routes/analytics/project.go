package analytics

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *AnalyticsHandler) GetProjectAnalytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := requtil.ParseUUID(r, "project_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		rows, err := h.q.GetProjectAnalytics(r.Context(), projectID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(rows, toProjectAnalyticsResponse))
	}
}
