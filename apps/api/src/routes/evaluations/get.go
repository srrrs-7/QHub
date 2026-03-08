package evaluations

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *EvaluationHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		eval, err := h.evalRepo.FindByID(r.Context(), id)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toEvaluationResponse(eval))
	}
}
