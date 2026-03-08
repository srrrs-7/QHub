package evaluations

import (
	"api/src/domain/executionlog"
	"encoding/json"
)

type evaluationResponse struct {
	ID             string          `json:"id"`
	ExecutionLogID string          `json:"execution_log_id"`
	OverallScore   *string         `json:"overall_score"`
	AccuracyScore  *string         `json:"accuracy_score"`
	RelevanceScore *string         `json:"relevance_score"`
	FluencyScore   *string         `json:"fluency_score"`
	SafetyScore    *string         `json:"safety_score"`
	Feedback       string          `json:"feedback"`
	EvaluatorType  string          `json:"evaluator_type"`
	EvaluatorID    string          `json:"evaluator_id"`
	Metadata       json.RawMessage `json:"metadata"`
	CreatedAt      string          `json:"created_at"`
}

func toEvaluationResponse(e executionlog.Evaluation) evaluationResponse {
	return evaluationResponse{
		ID:             e.ID.String(),
		ExecutionLogID: e.ExecutionLogID.String(),
		OverallScore:   e.OverallScore,
		AccuracyScore:  e.AccuracyScore,
		RelevanceScore: e.RelevanceScore,
		FluencyScore:   e.FluencyScore,
		SafetyScore:    e.SafetyScore,
		Feedback:       e.Feedback,
		EvaluatorType:  e.EvaluatorType,
		EvaluatorID:    e.EvaluatorID,
		Metadata:       e.Metadata,
		CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
