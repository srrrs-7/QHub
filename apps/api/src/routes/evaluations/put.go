package evaluations

import (
	"api/src/domain/executionlog"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *EvaluationHandler) Put() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requtil.ParseUUID(r, "id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePutEvaluationRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		eval := executionlog.Evaluation{
			ID:             id,
			OverallScore:   req.OverallScore,
			AccuracyScore:  req.AccuracyScore,
			RelevanceScore: req.RelevanceScore,
			FluencyScore:   req.FluencyScore,
			SafetyScore:    req.SafetyScore,
			Feedback:       req.Feedback,
			Metadata:       req.Metadata,
		}

		updated, err := h.evalRepo.Update(r.Context(), eval)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toEvaluationResponse(updated))
	}
}
