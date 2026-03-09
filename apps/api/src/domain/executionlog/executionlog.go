// Package executionlog defines domain entities for tracking prompt
// execution history and quality evaluations.
//
// An ExecutionLog records each invocation of a prompt version against
// an LLM, capturing request/response payloads, token usage, latency,
// and cost. Evaluations score these executions on multiple dimensions.
package executionlog

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ExecutionLog records a single prompt execution against an LLM.
type ExecutionLog struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	PromptID      uuid.UUID
	VersionNumber int
	RequestBody   json.RawMessage
	ResponseBody  json.RawMessage
	Model         string // e.g. "claude-sonnet-4-20250514"
	Provider      string // e.g. "anthropic"
	InputTokens   int
	OutputTokens  int
	TotalTokens   int
	LatencyMs     int
	EstimatedCost string // decimal cost string
	Status        string // "success", "error"
	ErrorMessage  string
	Environment   string // "development", "staging", "production"
	Metadata      json.RawMessage
	ExecutedAt    time.Time
	CreatedAt     time.Time
}

// Evaluation scores an execution log on multiple quality dimensions.
type Evaluation struct {
	ID             uuid.UUID
	ExecutionLogID uuid.UUID
	OverallScore   *string // nullable decimal scores
	AccuracyScore  *string
	RelevanceScore *string
	FluencyScore   *string
	SafetyScore    *string
	Feedback       string
	EvaluatorType  string // "human" or "auto"
	EvaluatorID    string
	Metadata       json.RawMessage
	CreatedAt      time.Time
}
