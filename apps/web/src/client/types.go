package client

import "encoding/json"

// --- Organization ---

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}

// --- Project ---

type Project struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
}

// --- Prompt ---

type Prompt struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	PromptType        string `json:"prompt_type"`
	Description       string `json:"description"`
	LatestVersion     int    `json:"latest_version"`
	ProductionVersion *int   `json:"production_version"`
}

func (p Prompt) StatusLabel() string {
	if p.ProductionVersion != nil {
		return "production"
	}
	return "draft"
}

// --- PromptVersion ---

type PromptVersion struct {
	ID                string          `json:"id"`
	PromptID          string          `json:"prompt_id"`
	VersionNumber     int             `json:"version_number"`
	Status            string          `json:"status"`
	Content           json.RawMessage `json:"content"`
	Variables         json.RawMessage `json:"variables"`
	ChangeDescription string          `json:"change_description"`
	AuthorID          string          `json:"author_id"`
}

func (v PromptVersion) IsProduction() bool { return v.Status == "production" }
func (v PromptVersion) IsDraft() bool      { return v.Status == "draft" }
func (v PromptVersion) IsReview() bool     { return v.Status == "review" }

func (v PromptVersion) ContentString() string {
	if v.Content == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(v.Content, &s); err != nil {
		return string(v.Content)
	}
	return s
}

func (v PromptVersion) VariablesList() []string {
	if v.Variables == nil {
		return nil
	}
	var vars []string
	_ = json.Unmarshal(v.Variables, &vars)
	return vars
}

// --- ExecutionLog ---

type ExecutionLog struct {
	ID            string          `json:"id"`
	OrgID         string          `json:"org_id"`
	PromptID      string          `json:"prompt_id"`
	VersionNumber int             `json:"version_number"`
	RequestBody   json.RawMessage `json:"request_body"`
	ResponseBody  json.RawMessage `json:"response_body"`
	Model         string          `json:"model"`
	Provider      string          `json:"provider"`
	InputTokens   int             `json:"input_tokens"`
	OutputTokens  int             `json:"output_tokens"`
	TotalTokens   int             `json:"total_tokens"`
	LatencyMs     int             `json:"latency_ms"`
	EstimatedCost string          `json:"estimated_cost"`
	Status        string          `json:"status"`
	ErrorMessage  string          `json:"error_message"`
	Environment   string          `json:"environment"`
	Metadata      json.RawMessage `json:"metadata"`
	ExecutedAt    string          `json:"executed_at"`
	CreatedAt     string          `json:"created_at"`
}

func (l ExecutionLog) InputString() string {
	if l.RequestBody == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(l.RequestBody, &s); err != nil {
		return string(l.RequestBody)
	}
	return s
}

func (l ExecutionLog) OutputString() string {
	if l.ResponseBody == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(l.ResponseBody, &s); err != nil {
		return string(l.ResponseBody)
	}
	return s
}

// --- Evaluation ---

type Evaluation struct {
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

func (e Evaluation) DisplayScore() string {
	if e.OverallScore != nil {
		return *e.OverallScore
	}
	return "-"
}

// --- ConsultingSession ---

type ConsultingSession struct {
	ID               string  `json:"id"`
	OrgID            string  `json:"org_id"`
	Title            string  `json:"title"`
	IndustryConfigID *string `json:"industry_config_id"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// --- ConsultingMessage ---

type ConsultingMessage struct {
	ID           string          `json:"id"`
	SessionID    string          `json:"session_id"`
	Role         string          `json:"role"`
	Content      string          `json:"content"`
	Citations    json.RawMessage `json:"citations"`
	ActionsTaken json.RawMessage `json:"actions_taken"`
	CreatedAt    string          `json:"created_at"`
}

// --- Tag ---

type Tag struct {
	ID        string `json:"id"`
	OrgID     string `json:"org_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

// --- Industry ---

type Industry struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Slug            string          `json:"slug"`
	Description     string          `json:"description"`
	KnowledgeBase   json.RawMessage `json:"knowledge_base"`
	ComplianceRules json.RawMessage `json:"compliance_rules"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// --- ComplianceResult ---

type ComplianceResult struct {
	Compliant  bool              `json:"compliant"`
	Violations []ComplianceIssue `json:"violations"`
}

type ComplianceIssue struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// --- Search ---

type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

type SearchResult struct {
	ID                string          `json:"id"`
	PromptID          string          `json:"prompt_id"`
	PromptName        string          `json:"prompt_name"`
	PromptSlug        string          `json:"prompt_slug"`
	VersionNumber     int             `json:"version_number"`
	Status            string          `json:"status"`
	Content           json.RawMessage `json:"content"`
	ChangeDescription string          `json:"change_description"`
	Similarity        float64         `json:"similarity"`
	CreatedAt         string          `json:"created_at"`
}

// --- Analytics ---

type PromptAnalytics struct {
	VersionNumber   int32  `json:"version_number"`
	TotalExecutions int64  `json:"total_executions"`
	AvgTokens       int32  `json:"avg_tokens"`
	AvgLatencyMs    int32  `json:"avg_latency_ms"`
	TotalCost       string `json:"total_cost"`
	SuccessCount    int64  `json:"success_count"`
	ErrorCount      int64  `json:"error_count"`
}

type DailyTrend struct {
	Day             string `json:"day"`
	TotalExecutions int64  `json:"total_executions"`
	AvgTokens       int32  `json:"avg_tokens"`
	AvgLatencyMs    int32  `json:"avg_latency_ms"`
	TotalCost       string `json:"total_cost"`
}

// --- Diff ---

type SemanticDiff struct {
	Summary     string       `json:"summary"`
	Changes     []DiffChange `json:"changes"`
	ToneShift   string       `json:"tone_shift,omitempty"`
	Specificity float64      `json:"specificity_change"`
}

type DiffChange struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type TextDiffResult struct {
	FromVersion int            `json:"from_version"`
	ToVersion   int            `json:"to_version"`
	Hunks       []TextDiffHunk `json:"hunks"`
	Stats       TextDiffStats  `json:"stats"`
}

type TextDiffHunk struct {
	Lines []TextDiffLine `json:"lines"`
}

type TextDiffLine struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	OldLine int    `json:"old_line,omitempty"`
	NewLine int    `json:"new_line,omitempty"`
}

type TextDiffStats struct {
	Added   int `json:"added"`
	Removed int `json:"removed"`
	Equal   int `json:"equal"`
}

// --- Lint ---

type LintResult struct {
	Score  int         `json:"score"`
	Issues []LintIssue `json:"issues"`
	Passed []string    `json:"passed"`
}

type LintIssue struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// --- Benchmark ---

type Benchmark struct {
	ID                string `json:"id"`
	IndustryConfigID  string `json:"industry_config_id"`
	Period            string `json:"period"`
	AvgQualityScore   string `json:"avg_quality_score"`
	AvgLatencyMs      string `json:"avg_latency_ms"`
	AvgCostPerRequest string `json:"avg_cost_per_request"`
	TotalExecutions   int64  `json:"total_executions"`
	P50Quality        string `json:"p50_quality"`
	P90Quality        string `json:"p90_quality"`
	OptInCount        int64  `json:"opt_in_count"`
	CreatedAt         string `json:"created_at"`
}
