package logs

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
	"strconv"
)

// ListByPrompt returns logs filtered by prompt_id from the URL path parameter.
func (h *LogHandler) ListByPrompt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promptID, err := requtil.ParseUUID(r, "prompt_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 20
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if offset < 0 {
			offset = 0
		}

		logs, err := h.logRepo.FindAllByPrompt(r.Context(), promptID, limit, offset)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		total, err := h.logRepo.CountByPrompt(r.Context(), promptID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, listLogsResponse{
			Data:  response.MapSlice(logs, toLogResponse),
			Total: total,
		})
	}
}
