package consulting

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *ConsultingHandler) PutSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutSessionRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		updated, err := h.sessionRepo.UpdateStatus(r.Context(), id, req.Status)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toSessionResponse(updated))
	}
}
