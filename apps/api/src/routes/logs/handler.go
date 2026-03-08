package logs

import (
	"api/src/domain/executionlog"
)

type LogHandler struct {
	logRepo  executionlog.LogRepository
	evalRepo executionlog.EvaluationRepository
}

func NewLogHandler(logRepo executionlog.LogRepository, evalRepo executionlog.EvaluationRepository) *LogHandler {
	return &LogHandler{logRepo: logRepo, evalRepo: evalRepo}
}
