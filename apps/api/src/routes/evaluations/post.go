package evaluations

import (
	"api/src/domain/executionlog"
	"api/src/routes/response"
	"net/http"

	"github.com/google/uuid"
)

func (h *EvaluationHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostEvaluationRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		eval := executionlog.Evaluation{
			ExecutionLogID: uuid.MustParse(req.ExecutionLogID),
			OverallScore:   req.OverallScore,
			AccuracyScore:  req.AccuracyScore,
			RelevanceScore: req.RelevanceScore,
			FluencyScore:   req.FluencyScore,
			SafetyScore:    req.SafetyScore,
			Feedback:       req.Feedback,
			EvaluatorType:  req.EvaluatorType,
			EvaluatorID:    req.EvaluatorID,
			Metadata:       req.Metadata,
		}

		created, err := h.evalRepo.Create(r.Context(), eval)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toEvaluationResponse(created))
	}
}
