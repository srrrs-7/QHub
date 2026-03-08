package sdk

import "encoding/json"

// PromptVersion represents a versioned prompt returned by the API.
type PromptVersion struct {
	ID            string          `json:"id"`
	PromptID      string          `json:"prompt_id"`
	VersionNumber int             `json:"version_number"`
	Status        string          `json:"status"`
	Content       json.RawMessage `json:"content"`
	Variables     json.RawMessage `json:"variables"`
	CreatedAt     string          `json:"created_at"`
}

// ExecutionLog represents a prompt execution log entry.
type ExecutionLog struct {
	ID            string          `json:"id,omitempty"`
	OrgID         string          `json:"org_id"`
	PromptID      string          `json:"prompt_id"`
	VersionNumber int             `json:"version_number"`
	RequestBody   json.RawMessage `json:"request_body"`
	ResponseBody  json.RawMessage `json:"response_body,omitempty"`
	Model         string          `json:"model"`
	Provider      string          `json:"provider"`
	InputTokens   int             `json:"input_tokens"`
	OutputTokens  int             `json:"output_tokens"`
	TotalTokens   int             `json:"total_tokens"`
	LatencyMs     int             `json:"latency_ms"`
	EstimatedCost string          `json:"estimated_cost"`
	Status        string          `json:"status"`
	ErrorMessage  string          `json:"error_message,omitempty"`
	Environment   string          `json:"environment"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	ExecutedAt    string          `json:"executed_at"`
	CreatedAt     string          `json:"created_at,omitempty"`
}

// Evaluation represents an evaluation of a prompt execution.
type Evaluation struct {
	ID             string          `json:"id,omitempty"`
	ExecutionLogID string          `json:"execution_log_id"`
	OverallScore   *string         `json:"overall_score,omitempty"`
	AccuracyScore  *string         `json:"accuracy_score,omitempty"`
	RelevanceScore *string         `json:"relevance_score,omitempty"`
	FluencyScore   *string         `json:"fluency_score,omitempty"`
	SafetyScore    *string         `json:"safety_score,omitempty"`
	Feedback       string          `json:"feedback,omitempty"`
	EvaluatorType  string          `json:"evaluator_type"`
	EvaluatorID    string          `json:"evaluator_id,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      string          `json:"created_at,omitempty"`
}
