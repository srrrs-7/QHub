package evaluations

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *EvaluationHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logID, err := requtil.ParseUUID(r, "log_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		evals, err := h.evalRepo.FindByLogID(r.Context(), logID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(evals, toEvaluationResponse))
	}
}
