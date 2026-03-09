package client

import "context"

// Client defines the interface for backend API communication.
// Both APIClient (real) and MockClient (test) implement this.
type Client interface {
	// Organizations
	GetOrganization(ctx context.Context, slug string) (*Organization, error)
	ListOrganizations(ctx context.Context) ([]Organization, error)
	CreateOrganization(ctx context.Context, body map[string]string) (*Organization, error)
	UpdateOrganization(ctx context.Context, slug string, body map[string]string) (*Organization, error)

	// Projects
	ListProjects(ctx context.Context, orgID string) ([]Project, error)
	GetProject(ctx context.Context, orgID, slug string) (*Project, error)
	CreateProject(ctx context.Context, orgID string, body map[string]string) (*Project, error)
	UpdateProject(ctx context.Context, orgID, slug string, body map[string]string) (*Project, error)
	DeleteProject(ctx context.Context, orgID, slug string) error

	// Prompts
	ListPrompts(ctx context.Context, projectID string) ([]Prompt, error)
	GetPrompt(ctx context.Context, projectID, slug string) (*Prompt, error)
	CreatePrompt(ctx context.Context, projectID string, body map[string]string) (*Prompt, error)
	UpdatePrompt(ctx context.Context, projectID, slug string, body map[string]string) (*Prompt, error)

	// Versions
	ListVersions(ctx context.Context, promptID string) ([]PromptVersion, error)
	GetVersion(ctx context.Context, promptID, version string) (*PromptVersion, error)
	CreateVersion(ctx context.Context, promptID string, body map[string]any) (*PromptVersion, error)
	UpdateVersionStatus(ctx context.Context, promptID, version, status string) (*PromptVersion, error)

	// Execution Logs
	ListLogs(ctx context.Context) ([]ExecutionLog, error)
	GetLog(ctx context.Context, id string) (*ExecutionLog, error)
	ListLogEvaluations(ctx context.Context, logID string) ([]Evaluation, error)

	// Evaluations
	ListEvaluations(ctx context.Context) ([]Evaluation, error)
	CreateEvaluation(ctx context.Context, logID string, body map[string]any) (*Evaluation, error)
	GetEvaluation(ctx context.Context, id string) (*Evaluation, error)
	UpdateEvaluation(ctx context.Context, id string, body map[string]any) (*Evaluation, error)

	// Consulting
	ListConsultingSessions(ctx context.Context) ([]ConsultingSession, error)
	CreateConsultingSession(ctx context.Context, body map[string]string) (*ConsultingSession, error)
	GetConsultingSession(ctx context.Context, id string) (*ConsultingSession, error)
	ListConsultingMessages(ctx context.Context, sessionID string) ([]ConsultingMessage, error)
	SendConsultingMessage(ctx context.Context, sessionID string, body map[string]string) (*ConsultingMessage, error)
	CloseSession(ctx context.Context, sessionID string) (*ConsultingSession, error)

	// Tags
	ListTags(ctx context.Context) ([]Tag, error)
	CreateTag(ctx context.Context, body map[string]string) (*Tag, error)
	DeleteTag(ctx context.Context, name string) error
	ListPromptTags(ctx context.Context, promptID string) ([]Tag, error)
	AddPromptTag(ctx context.Context, promptID, tagID string) error
	RemovePromptTag(ctx context.Context, promptID, tagID string) error

	// Industries
	ListIndustries(ctx context.Context) ([]Industry, error)
	GetIndustry(ctx context.Context, slug string) (*Industry, error)
	CreateIndustry(ctx context.Context, body map[string]string) (*Industry, error)
	UpdateIndustry(ctx context.Context, slug string, body map[string]string) (*Industry, error)
	CheckCompliance(ctx context.Context, slug string, body map[string]string) (*ComplianceResult, error)
	ListBenchmarks(ctx context.Context, slug string) ([]Benchmark, error)

	// Members
	ListMembers(ctx context.Context, orgID string) ([]Member, error)
	AddMember(ctx context.Context, orgID string, body map[string]string) (*Member, error)
	UpdateMemberRole(ctx context.Context, orgID, userID string, body map[string]string) (*Member, error)
	RemoveMember(ctx context.Context, orgID, userID string) error

	// API Keys
	ListAPIKeys(ctx context.Context, orgID string) ([]APIKey, error)
	CreateAPIKey(ctx context.Context, orgID string, body map[string]string) (*APIKeyCreated, error)
	DeleteAPIKey(ctx context.Context, orgID, id string) error

	// Analytics
	GetPromptAnalytics(ctx context.Context, promptID string) ([]PromptAnalytics, error)
	GetDailyTrend(ctx context.Context, promptID string, days string) ([]DailyTrend, error)
	GetProjectAnalytics(ctx context.Context, projectID string) ([]ProjectAnalytics, error)
	GetVersionAnalytics(ctx context.Context, promptID string, version string) (*PromptAnalytics, error)

	// Diff & Lint
	GetSemanticDiff(ctx context.Context, promptID, v1, v2 string) (*SemanticDiff, error)
	GetTextDiff(ctx context.Context, promptID, version string) (*TextDiffResult, error)
	GetLintResult(ctx context.Context, promptID, version string) (*LintResult, error)

	// Version Comparison
	CompareVersions(ctx context.Context, promptID, v1, v2 string) (*VersionComparison, error)

	// Logs (create)
	CreateLog(ctx context.Context, body map[string]any) (*ExecutionLog, error)

	// Search
	SemanticSearch(ctx context.Context, body map[string]any) (*SearchResponse, error)

	// Embedding Status
	GetEmbeddingStatus(ctx context.Context) (map[string]string, error)
}

// Compile-time check that APIClient implements Client.
var _ Client = (*APIClient)(nil)
