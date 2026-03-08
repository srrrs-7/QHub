package executionlog

import (
	"context"

	"github.com/google/uuid"
)

type LogRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (ExecutionLog, error)
	FindAllByPrompt(ctx context.Context, promptID uuid.UUID, limit, offset int) ([]ExecutionLog, error)
	FindAllByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]ExecutionLog, error)
	Create(ctx context.Context, log ExecutionLog) (ExecutionLog, error)
	CountByPrompt(ctx context.Context, promptID uuid.UUID) (int64, error)
}

type EvaluationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Evaluation, error)
	FindByLogID(ctx context.Context, logID uuid.UUID) ([]Evaluation, error)
	Create(ctx context.Context, eval Evaluation) (Evaluation, error)
}
