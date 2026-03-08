package logs

import (
	"api/src/routes/response"
	"net/http"
)

func (h *LogHandler) PostBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostBatchLogRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		results := make([]logResponse, 0, len(req.Logs))
		for _, logReq := range req.Logs {
			log := toLogEntity(logReq)
			created, err := h.logRepo.Create(r.Context(), log)
			if err != nil {
				response.HandleError(w, err)
				return
			}
			results = append(results, toLogResponse(created))
		}

		response.Created(w, results)
	}
}
