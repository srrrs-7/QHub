package executionlog

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ExecutionLog struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	PromptID       uuid.UUID
	VersionNumber  int
	RequestBody    json.RawMessage
	ResponseBody   json.RawMessage
	Model          string
	Provider       string
	InputTokens    int
	OutputTokens   int
	TotalTokens    int
	LatencyMs      int
	EstimatedCost  string
	Status         string
	ErrorMessage   string
	Environment    string
	Metadata       json.RawMessage
	ExecutedAt     time.Time
	CreatedAt      time.Time
}

type Evaluation struct {
	ID             uuid.UUID
	ExecutionLogID uuid.UUID
	OverallScore   *string
	AccuracyScore  *string
	RelevanceScore *string
	FluencyScore   *string
	SafetyScore    *string
	Feedback       string
	EvaluatorType  string
	EvaluatorID    string
	Metadata       json.RawMessage
	CreatedAt      time.Time
}
