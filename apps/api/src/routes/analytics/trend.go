package analytics

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"time"

	db "utils/db/db"
)

func (h *AnalyticsHandler) GetDailyTrend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		now := time.Now()
		start := now.AddDate(0, 0, -30)
		end := now

		if s := r.URL.Query().Get("start"); s != "" {
			parsed, err := time.Parse("2006-01-02", s)
			if err != nil {
				response.HandleError(w, err)
				return
			}
			start = parsed
		}

		if e := r.URL.Query().Get("end"); e != "" {
			parsed, err := time.Parse("2006-01-02", e)
			if err != nil {
				response.HandleError(w, err)
				return
			}
			end = parsed
		}

		rows, err := h.q.GetDailyTrend(r.Context(), db.GetDailyTrendParams{
			PromptID:     promptID,
			ExecutedAt:   start,
			ExecutedAt_2: end,
		})
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(rows, toDailyTrendResponse))
	}
}
