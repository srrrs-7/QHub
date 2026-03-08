package evaluations

import (
	"api/src/routes/requtil"
	"encoding/json"
	"net/http"
)

type postEvaluationRequest struct {
	ExecutionLogID string          `json:"execution_log_id" validate:"required,uuid"`
	OverallScore   *string         `json:"overall_score"`
	AccuracyScore  *string         `json:"accuracy_score"`
	RelevanceScore *string         `json:"relevance_score"`
	FluencyScore   *string         `json:"fluency_score"`
	SafetyScore    *string         `json:"safety_score"`
	Feedback       string          `json:"feedback" validate:"omitempty,max=2000"`
	EvaluatorType  string          `json:"evaluator_type" validate:"required,oneof=human auto"`
	EvaluatorID    string          `json:"evaluator_id" validate:"omitempty,max=200"`
	Metadata       json.RawMessage `json:"metadata"`
}

func decodePostEvaluationRequest(r *http.Request) (postEvaluationRequest, error) {
	return requtil.Decode(r, func(req *postEvaluationRequest) {
		req.Feedback = requtil.Sanitize.Sanitize(req.Feedback)
		req.EvaluatorID = requtil.Sanitize.Sanitize(req.EvaluatorID)
	})
}
