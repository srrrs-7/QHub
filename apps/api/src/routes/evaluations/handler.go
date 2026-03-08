package evaluations

import (
	"api/src/domain/executionlog"
)

type EvaluationHandler struct {
	evalRepo executionlog.EvaluationRepository
}

func NewEvaluationHandler(evalRepo executionlog.EvaluationRepository) *EvaluationHandler {
	return &EvaluationHandler{evalRepo: evalRepo}
}
