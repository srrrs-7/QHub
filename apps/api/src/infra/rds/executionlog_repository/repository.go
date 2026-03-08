package executionlog_repository

import (
	"api/src/domain/executionlog"
	"api/src/infra/rds/repoerr"
	"context"
	"database/sql"
	"time"
	"utils/db/db"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

const dbTimeout = 5 * time.Second

// --- LogRepository ---

type LogRepository struct {
	q db.Querier
}

func NewLogRepository(q db.Querier) *LogRepository {
	return &LogRepository{q: q}
}

var _ executionlog.LogRepository = (*LogRepository)(nil)

func toLog(l db.ExecutionLog) executionlog.ExecutionLog {
	return executionlog.ExecutionLog{
		ID:            l.ID,
		OrgID:         l.OrganizationID,
		PromptID:      l.PromptID,
		VersionNumber: int(l.VersionNumber),
		RequestBody:   l.RequestBody,
		ResponseBody:  l.ResponseBody.RawMessage,
		Model:         l.Model,
		Provider:      l.Provider,
		InputTokens:   int(l.InputTokens),
		OutputTokens:  int(l.OutputTokens),
		TotalTokens:   int(l.TotalTokens),
		LatencyMs:     int(l.LatencyMs),
		EstimatedCost: l.EstimatedCost,
		Status:        l.Status,
		ErrorMessage:  l.ErrorMessage.String,
		Environment:   l.Environment,
		Metadata:      l.Metadata.RawMessage,
		ExecutedAt:    l.ExecutedAt,
		CreatedAt:     l.CreatedAt,
	}
}

func (r *LogRepository) FindByID(ctx context.Context, id uuid.UUID) (executionlog.ExecutionLog, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	l, err := r.q.GetExecutionLog(ctx, id)
	if err != nil {
		return executionlog.ExecutionLog{}, repoerr.Handle(err, "LogRepository", "ExecutionLog")
	}
	return toLog(l), nil
}

func (r *LogRepository) FindAllByPrompt(ctx context.Context, promptID uuid.UUID, limit, offset int) ([]executionlog.ExecutionLog, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	logs, err := r.q.ListExecutionLogsByPrompt(ctx, db.ListExecutionLogsByPromptParams{
		PromptID: promptID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, repoerr.Handle(err, "LogRepository", "")
	}

	result := make([]executionlog.ExecutionLog, 0, len(logs))
	for _, l := range logs {
		result = append(result, toLog(l))
	}
	return result, nil
}

func (r *LogRepository) FindAllByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]executionlog.ExecutionLog, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	logs, err := r.q.ListExecutionLogsByOrg(ctx, db.ListExecutionLogsByOrgParams{
		OrganizationID: orgID,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, repoerr.Handle(err, "LogRepository", "")
	}

	result := make([]executionlog.ExecutionLog, 0, len(logs))
	for _, l := range logs {
		result = append(result, toLog(l))
	}
	return result, nil
}

func (r *LogRepository) Create(ctx context.Context, log executionlog.ExecutionLog) (executionlog.ExecutionLog, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	l, err := r.q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
		OrganizationID: log.OrgID,
		PromptID:       log.PromptID,
		VersionNumber:  int32(log.VersionNumber),
		RequestBody:    log.RequestBody,
		ResponseBody:   pqtype.NullRawMessage{RawMessage: log.ResponseBody, Valid: log.ResponseBody != nil},
		Model:          log.Model,
		Provider:       log.Provider,
		InputTokens:    int32(log.InputTokens),
		OutputTokens:   int32(log.OutputTokens),
		TotalTokens:    int32(log.TotalTokens),
		LatencyMs:      int32(log.LatencyMs),
		EstimatedCost:  log.EstimatedCost,
		Status:         log.Status,
		ErrorMessage:   sql.NullString{String: log.ErrorMessage, Valid: log.ErrorMessage != ""},
		Environment:    log.Environment,
		Metadata:       pqtype.NullRawMessage{RawMessage: log.Metadata, Valid: log.Metadata != nil},
		ExecutedAt:     log.ExecutedAt,
	})
	if err != nil {
		return executionlog.ExecutionLog{}, repoerr.Handle(err, "LogRepository", "")
	}
	return toLog(l), nil
}

func (r *LogRepository) CountByPrompt(ctx context.Context, promptID uuid.UUID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	count, err := r.q.CountExecutionLogsByPrompt(ctx, promptID)
	if err != nil {
		return 0, repoerr.Handle(err, "LogRepository", "")
	}
	return count, nil
}

// --- EvaluationRepository ---

type EvaluationRepository struct {
	q db.Querier
}

func NewEvaluationRepository(q db.Querier) *EvaluationRepository {
	return &EvaluationRepository{q: q}
}

var _ executionlog.EvaluationRepository = (*EvaluationRepository)(nil)

func toEvaluation(e db.Evaluation) executionlog.Evaluation {
	var overall, accuracy, relevance, fluency, safety *string
	if e.OverallScore.Valid {
		overall = &e.OverallScore.String
	}
	if e.AccuracyScore.Valid {
		accuracy = &e.AccuracyScore.String
	}
	if e.RelevanceScore.Valid {
		relevance = &e.RelevanceScore.String
	}
	if e.FluencyScore.Valid {
		fluency = &e.FluencyScore.String
	}
	if e.SafetyScore.Valid {
		safety = &e.SafetyScore.String
	}

	return executionlog.Evaluation{
		ID:             e.ID,
		ExecutionLogID: e.ExecutionLogID,
		OverallScore:   overall,
		AccuracyScore:  accuracy,
		RelevanceScore: relevance,
		FluencyScore:   fluency,
		SafetyScore:    safety,
		Feedback:       e.Feedback.String,
		EvaluatorType:  e.EvaluatorType,
		EvaluatorID:    e.EvaluatorID.String,
		Metadata:       e.Metadata.RawMessage,
		CreatedAt:      e.CreatedAt,
	}
}

func (r *EvaluationRepository) FindByID(ctx context.Context, id uuid.UUID) (executionlog.Evaluation, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	e, err := r.q.GetEvaluation(ctx, id)
	if err != nil {
		return executionlog.Evaluation{}, repoerr.Handle(err, "EvaluationRepository", "Evaluation")
	}
	return toEvaluation(e), nil
}

func (r *EvaluationRepository) FindByLogID(ctx context.Context, logID uuid.UUID) ([]executionlog.Evaluation, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	evals, err := r.q.ListEvaluationsByLog(ctx, logID)
	if err != nil {
		return nil, repoerr.Handle(err, "EvaluationRepository", "")
	}

	result := make([]executionlog.Evaluation, 0, len(evals))
	for _, e := range evals {
		result = append(result, toEvaluation(e))
	}
	return result, nil
}

func (r *EvaluationRepository) Create(ctx context.Context, eval executionlog.Evaluation) (executionlog.Evaluation, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	e, err := r.q.CreateEvaluation(ctx, db.CreateEvaluationParams{
		ExecutionLogID: eval.ExecutionLogID,
		OverallScore:   toNullString(eval.OverallScore),
		AccuracyScore:  toNullString(eval.AccuracyScore),
		RelevanceScore: toNullString(eval.RelevanceScore),
		FluencyScore:   toNullString(eval.FluencyScore),
		SafetyScore:    toNullString(eval.SafetyScore),
		Feedback:       sql.NullString{String: eval.Feedback, Valid: eval.Feedback != ""},
		EvaluatorType:  eval.EvaluatorType,
		EvaluatorID:    sql.NullString{String: eval.EvaluatorID, Valid: eval.EvaluatorID != ""},
		Metadata:       pqtype.NullRawMessage{RawMessage: eval.Metadata, Valid: eval.Metadata != nil},
	})
	if err != nil {
		return executionlog.Evaluation{}, repoerr.Handle(err, "EvaluationRepository", "")
	}
	return toEvaluation(e), nil
}

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
