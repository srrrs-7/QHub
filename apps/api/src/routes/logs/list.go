package logs

import (
	"api/src/routes/response"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func (h *LogHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 20
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if offset < 0 {
			offset = 0
		}

		promptIDStr := r.URL.Query().Get("prompt_id")
		orgIDStr := r.URL.Query().Get("org_id")

		if promptIDStr != "" {
			promptID, err := uuid.Parse(promptIDStr)
			if err != nil {
				response.HandleError(w, err)
				return
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
			return
		}

		if orgIDStr != "" {
			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				response.HandleError(w, err)
				return
			}

			logs, err := h.logRepo.FindAllByOrg(r.Context(), orgID, limit, offset)
			if err != nil {
				response.HandleError(w, err)
				return
			}

			response.OK(w, listLogsResponse{
				Data:  response.MapSlice(logs, toLogResponse),
				Total: int64(len(logs)),
			})
			return
		}

		response.OK(w, listLogsResponse{
			Data:  []logResponse{},
			Total: 0,
		})
	}
}
