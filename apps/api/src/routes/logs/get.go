package logs

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *LogHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		log, err := h.logRepo.FindByID(r.Context(), id)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toLogResponse(log))
	}
}
