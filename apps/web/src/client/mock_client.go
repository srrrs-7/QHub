package client

import "context"

// MockClient is a configurable mock implementation of Client for testing.
// Set fields to control return values. Nil functions use defaults.
type MockClient struct {
	// Organizations
	GetOrganizationFn    func(ctx context.Context, slug string) (*Organization, error)
	ListOrganizationsFn  func(ctx context.Context) ([]Organization, error)
	CreateOrganizationFn func(ctx context.Context, body map[string]string) (*Organization, error)
	UpdateOrganizationFn func(ctx context.Context, slug string, body map[string]string) (*Organization, error)

	// Projects
	ListProjectsFn  func(ctx context.Context, orgID string) ([]Project, error)
	GetProjectFn    func(ctx context.Context, orgID, slug string) (*Project, error)
	CreateProjectFn func(ctx context.Context, orgID string, body map[string]string) (*Project, error)
	UpdateProjectFn func(ctx context.Context, orgID, slug string, body map[string]string) (*Project, error)
	DeleteProjectFn func(ctx context.Context, orgID, slug string) error

	// Prompts
	ListPromptsFn  func(ctx context.Context, projectID string) ([]Prompt, error)
	GetPromptFn    func(ctx context.Context, projectID, slug string) (*Prompt, error)
	CreatePromptFn func(ctx context.Context, projectID string, body map[string]string) (*Prompt, error)
	UpdatePromptFn func(ctx context.Context, projectID, slug string, body map[string]string) (*Prompt, error)

	// Versions
	ListVersionsFn        func(ctx context.Context, promptID string) ([]PromptVersion, error)
	GetVersionFn          func(ctx context.Context, promptID, version string) (*PromptVersion, error)
	CreateVersionFn       func(ctx context.Context, promptID string, body map[string]any) (*PromptVersion, error)
	UpdateVersionStatusFn func(ctx context.Context, promptID, version, status string) (*PromptVersion, error)

	// Execution Logs
	ListLogsFn          func(ctx context.Context) ([]ExecutionLog, error)
	GetLogFn            func(ctx context.Context, id string) (*ExecutionLog, error)
	ListLogEvaluationFn func(ctx context.Context, logID string) ([]Evaluation, error)

	// Evaluations
	ListEvaluationsFn  func(ctx context.Context) ([]Evaluation, error)
	CreateEvaluationFn func(ctx context.Context, logID string, body map[string]any) (*Evaluation, error)

	// Consulting
	ListConsultingSessionsFn  func(ctx context.Context) ([]ConsultingSession, error)
	CreateConsultingSessionFn func(ctx context.Context, body map[string]string) (*ConsultingSession, error)
	GetConsultingSessionFn    func(ctx context.Context, id string) (*ConsultingSession, error)
	ListConsultingMessagesFn  func(ctx context.Context, sessionID string) ([]ConsultingMessage, error)
	SendConsultingMessageFn   func(ctx context.Context, sessionID string, body map[string]string) (*ConsultingMessage, error)
	CloseSessionFn            func(ctx context.Context, sessionID string) (*ConsultingSession, error)

	// Tags
	ListTagsFn  func(ctx context.Context) ([]Tag, error)
	CreateTagFn func(ctx context.Context, body map[string]string) (*Tag, error)
	DeleteTagFn func(ctx context.Context, name string) error

	// Industries
	ListIndustriesFn  func(ctx context.Context) ([]Industry, error)
	GetIndustryFn     func(ctx context.Context, slug string) (*Industry, error)
	CreateIndustryFn  func(ctx context.Context, body map[string]string) (*Industry, error)
	UpdateIndustryFn  func(ctx context.Context, slug string, body map[string]string) (*Industry, error)
	CheckComplianceFn func(ctx context.Context, slug string, body map[string]string) (*ComplianceResult, error)
	ListBenchmarksFn  func(ctx context.Context, slug string) ([]Benchmark, error)

	// Members
	ListMembersFn     func(ctx context.Context, orgID string) ([]Member, error)
	AddMemberFn       func(ctx context.Context, orgID string, body map[string]string) (*Member, error)
	UpdateMemberRolFn func(ctx context.Context, orgID, userID string, body map[string]string) (*Member, error)
	RemoveMemberFn    func(ctx context.Context, orgID, userID string) error

	// API Keys
	ListAPIKeysFn  func(ctx context.Context, orgID string) ([]APIKey, error)
	CreateAPIKeyFn func(ctx context.Context, orgID string, body map[string]string) (*APIKeyCreated, error)
	DeleteAPIKeyFn func(ctx context.Context, orgID, id string) error

	// Analytics
	GetPromptAnalyticsFn  func(ctx context.Context, promptID string) ([]PromptAnalytics, error)
	GetDailyTrendFn       func(ctx context.Context, promptID string, days string) ([]DailyTrend, error)
	GetProjectAnalyticsFn func(ctx context.Context, projectID string) ([]ProjectAnalytics, error)
	GetVersionAnalyticsFn func(ctx context.Context, promptID string, version string) (*PromptAnalytics, error)

	// Diff & Lint
	GetSemanticDiffFn func(ctx context.Context, promptID, v1, v2 string) (*SemanticDiff, error)
	GetTextDiffFn     func(ctx context.Context, promptID, version string) (*TextDiffResult, error)
	GetLintResultFn   func(ctx context.Context, promptID, version string) (*LintResult, error)

	// Version Comparison
	CompareVersionsFn func(ctx context.Context, promptID, v1, v2 string) (*VersionComparison, error)

	// Search
	SemanticSearchFn func(ctx context.Context, body map[string]any) (*SearchResponse, error)

	// Embedding Status
	GetEmbeddingStatusFn func(ctx context.Context) (map[string]string, error)
}

var _ Client = (*MockClient)(nil)

func (m *MockClient) GetOrganization(ctx context.Context, slug string) (*Organization, error) {
	if m.GetOrganizationFn != nil {
		return m.GetOrganizationFn(ctx, slug)
	}
	return &Organization{ID: "org-1", Name: "Test Org", Slug: slug, Plan: "free"}, nil
}

func (m *MockClient) ListOrganizations(ctx context.Context) ([]Organization, error) {
	if m.ListOrganizationsFn != nil {
		return m.ListOrganizationsFn(ctx)
	}
	return []Organization{{ID: "org-1", Name: "Test Org", Slug: "test-org", Plan: "free"}}, nil
}

func (m *MockClient) CreateOrganization(ctx context.Context, body map[string]string) (*Organization, error) {
	if m.CreateOrganizationFn != nil {
		return m.CreateOrganizationFn(ctx, body)
	}
	return &Organization{ID: "org-new", Name: body["name"], Slug: body["slug"], Plan: body["plan"]}, nil
}

func (m *MockClient) UpdateOrganization(ctx context.Context, slug string, body map[string]string) (*Organization, error) {
	if m.UpdateOrganizationFn != nil {
		return m.UpdateOrganizationFn(ctx, slug, body)
	}
	return &Organization{ID: "org-1", Name: body["name"], Slug: slug, Plan: body["plan"]}, nil
}

func (m *MockClient) ListProjects(ctx context.Context, orgID string) ([]Project, error) {
	if m.ListProjectsFn != nil {
		return m.ListProjectsFn(ctx, orgID)
	}
	return []Project{{ID: "proj-1", OrganizationID: orgID, Name: "Test Project", Slug: "test-project"}}, nil
}

func (m *MockClient) GetProject(ctx context.Context, orgID, slug string) (*Project, error) {
	if m.GetProjectFn != nil {
		return m.GetProjectFn(ctx, orgID, slug)
	}
	return &Project{ID: "proj-1", OrganizationID: orgID, Name: "Test Project", Slug: slug}, nil
}

func (m *MockClient) CreateProject(ctx context.Context, orgID string, body map[string]string) (*Project, error) {
	if m.CreateProjectFn != nil {
		return m.CreateProjectFn(ctx, orgID, body)
	}
	return &Project{ID: "proj-new", OrganizationID: orgID, Name: body["name"], Slug: body["slug"]}, nil
}

func (m *MockClient) UpdateProject(ctx context.Context, orgID, slug string, body map[string]string) (*Project, error) {
	if m.UpdateProjectFn != nil {
		return m.UpdateProjectFn(ctx, orgID, slug, body)
	}
	return &Project{ID: "proj-1", OrganizationID: orgID, Name: body["name"], Slug: slug}, nil
}

func (m *MockClient) DeleteProject(ctx context.Context, orgID, slug string) error {
	if m.DeleteProjectFn != nil {
		return m.DeleteProjectFn(ctx, orgID, slug)
	}
	return nil
}

func (m *MockClient) ListPrompts(ctx context.Context, projectID string) ([]Prompt, error) {
	if m.ListPromptsFn != nil {
		return m.ListPromptsFn(ctx, projectID)
	}
	return []Prompt{{ID: "prompt-1", ProjectID: projectID, Name: "Test Prompt", Slug: "test-prompt", PromptType: "chat", LatestVersion: 1}}, nil
}

func (m *MockClient) GetPrompt(ctx context.Context, projectID, slug string) (*Prompt, error) {
	if m.GetPromptFn != nil {
		return m.GetPromptFn(ctx, projectID, slug)
	}
	return &Prompt{ID: "prompt-1", ProjectID: projectID, Name: "Test Prompt", Slug: slug, PromptType: "chat", LatestVersion: 1}, nil
}

func (m *MockClient) CreatePrompt(ctx context.Context, projectID string, body map[string]string) (*Prompt, error) {
	if m.CreatePromptFn != nil {
		return m.CreatePromptFn(ctx, projectID, body)
	}
	return &Prompt{ID: "prompt-new", ProjectID: projectID, Name: body["name"], Slug: body["slug"]}, nil
}

func (m *MockClient) UpdatePrompt(ctx context.Context, projectID, slug string, body map[string]string) (*Prompt, error) {
	if m.UpdatePromptFn != nil {
		return m.UpdatePromptFn(ctx, projectID, slug, body)
	}
	return &Prompt{ID: "prompt-1", ProjectID: projectID, Name: body["name"], Slug: slug}, nil
}

func (m *MockClient) ListVersions(ctx context.Context, promptID string) ([]PromptVersion, error) {
	if m.ListVersionsFn != nil {
		return m.ListVersionsFn(ctx, promptID)
	}
	return []PromptVersion{{ID: "ver-1", PromptID: promptID, VersionNumber: 1, Status: "draft", ChangeDescription: "Initial"}}, nil
}

func (m *MockClient) GetVersion(ctx context.Context, promptID, version string) (*PromptVersion, error) {
	if m.GetVersionFn != nil {
		return m.GetVersionFn(ctx, promptID, version)
	}
	return &PromptVersion{ID: "ver-1", PromptID: promptID, VersionNumber: 1, Status: "draft"}, nil
}

func (m *MockClient) CreateVersion(ctx context.Context, promptID string, body map[string]any) (*PromptVersion, error) {
	if m.CreateVersionFn != nil {
		return m.CreateVersionFn(ctx, promptID, body)
	}
	return &PromptVersion{ID: "ver-new", PromptID: promptID, VersionNumber: 2, Status: "draft"}, nil
}

func (m *MockClient) UpdateVersionStatus(ctx context.Context, promptID, version, status string) (*PromptVersion, error) {
	if m.UpdateVersionStatusFn != nil {
		return m.UpdateVersionStatusFn(ctx, promptID, version, status)
	}
	return &PromptVersion{ID: "ver-1", PromptID: promptID, VersionNumber: 1, Status: status}, nil
}

func (m *MockClient) ListLogs(ctx context.Context) ([]ExecutionLog, error) {
	if m.ListLogsFn != nil {
		return m.ListLogsFn(ctx)
	}
	return []ExecutionLog{{ID: "log-1", Model: "gpt-4", Status: "success", LatencyMs: 150, TotalTokens: 500}}, nil
}

func (m *MockClient) GetLog(ctx context.Context, id string) (*ExecutionLog, error) {
	if m.GetLogFn != nil {
		return m.GetLogFn(ctx, id)
	}
	return &ExecutionLog{ID: id, Model: "gpt-4", Status: "success", LatencyMs: 150, TotalTokens: 500}, nil
}

func (m *MockClient) ListLogEvaluations(ctx context.Context, logID string) ([]Evaluation, error) {
	if m.ListLogEvaluationFn != nil {
		return m.ListLogEvaluationFn(ctx, logID)
	}
	return []Evaluation{}, nil
}

func (m *MockClient) ListEvaluations(ctx context.Context) ([]Evaluation, error) {
	if m.ListEvaluationsFn != nil {
		return m.ListEvaluationsFn(ctx)
	}
	return []Evaluation{}, nil
}

func (m *MockClient) CreateEvaluation(ctx context.Context, logID string, body map[string]any) (*Evaluation, error) {
	if m.CreateEvaluationFn != nil {
		return m.CreateEvaluationFn(ctx, logID, body)
	}
	return &Evaluation{ID: "eval-new", ExecutionLogID: logID, EvaluatorType: "human"}, nil
}

func (m *MockClient) ListConsultingSessions(ctx context.Context) ([]ConsultingSession, error) {
	if m.ListConsultingSessionsFn != nil {
		return m.ListConsultingSessionsFn(ctx)
	}
	return []ConsultingSession{{ID: "sess-1", Title: "Test Session", Status: "active"}}, nil
}

func (m *MockClient) CreateConsultingSession(ctx context.Context, body map[string]string) (*ConsultingSession, error) {
	if m.CreateConsultingSessionFn != nil {
		return m.CreateConsultingSessionFn(ctx, body)
	}
	return &ConsultingSession{ID: "sess-new", Title: body["title"], Status: "active"}, nil
}

func (m *MockClient) GetConsultingSession(ctx context.Context, id string) (*ConsultingSession, error) {
	if m.GetConsultingSessionFn != nil {
		return m.GetConsultingSessionFn(ctx, id)
	}
	return &ConsultingSession{ID: id, Title: "Test Session", Status: "active"}, nil
}

func (m *MockClient) ListConsultingMessages(ctx context.Context, sessionID string) ([]ConsultingMessage, error) {
	if m.ListConsultingMessagesFn != nil {
		return m.ListConsultingMessagesFn(ctx, sessionID)
	}
	return []ConsultingMessage{}, nil
}

func (m *MockClient) SendConsultingMessage(ctx context.Context, sessionID string, body map[string]string) (*ConsultingMessage, error) {
	if m.SendConsultingMessageFn != nil {
		return m.SendConsultingMessageFn(ctx, sessionID, body)
	}
	return &ConsultingMessage{ID: "msg-new", SessionID: sessionID, Role: "user", Content: body["content"]}, nil
}

func (m *MockClient) CloseSession(ctx context.Context, sessionID string) (*ConsultingSession, error) {
	if m.CloseSessionFn != nil {
		return m.CloseSessionFn(ctx, sessionID)
	}
	return &ConsultingSession{ID: sessionID, Title: "Test Session", Status: "closed"}, nil
}

func (m *MockClient) ListTags(ctx context.Context) ([]Tag, error) {
	if m.ListTagsFn != nil {
		return m.ListTagsFn(ctx)
	}
	return []Tag{{ID: "tag-1", Name: "test-tag", Color: "#ff0000"}}, nil
}

func (m *MockClient) CreateTag(ctx context.Context, body map[string]string) (*Tag, error) {
	if m.CreateTagFn != nil {
		return m.CreateTagFn(ctx, body)
	}
	return &Tag{ID: "tag-new", Name: body["name"], Color: body["color"]}, nil
}

func (m *MockClient) DeleteTag(ctx context.Context, name string) error {
	if m.DeleteTagFn != nil {
		return m.DeleteTagFn(ctx, name)
	}
	return nil
}

func (m *MockClient) ListIndustries(ctx context.Context) ([]Industry, error) {
	if m.ListIndustriesFn != nil {
		return m.ListIndustriesFn(ctx)
	}
	return []Industry{{ID: "ind-1", Name: "Healthcare", Slug: "healthcare"}}, nil
}

func (m *MockClient) GetIndustry(ctx context.Context, slug string) (*Industry, error) {
	if m.GetIndustryFn != nil {
		return m.GetIndustryFn(ctx, slug)
	}
	return &Industry{ID: "ind-1", Name: "Healthcare", Slug: slug}, nil
}

func (m *MockClient) CreateIndustry(ctx context.Context, body map[string]string) (*Industry, error) {
	if m.CreateIndustryFn != nil {
		return m.CreateIndustryFn(ctx, body)
	}
	return &Industry{ID: "ind-new", Name: body["name"], Slug: body["slug"]}, nil
}

func (m *MockClient) UpdateIndustry(ctx context.Context, slug string, body map[string]string) (*Industry, error) {
	if m.UpdateIndustryFn != nil {
		return m.UpdateIndustryFn(ctx, slug, body)
	}
	return &Industry{ID: "ind-1", Name: body["name"], Slug: slug}, nil
}

func (m *MockClient) CheckCompliance(ctx context.Context, slug string, body map[string]string) (*ComplianceResult, error) {
	if m.CheckComplianceFn != nil {
		return m.CheckComplianceFn(ctx, slug, body)
	}
	return &ComplianceResult{Compliant: true, Violations: []ComplianceIssue{}}, nil
}

func (m *MockClient) ListBenchmarks(ctx context.Context, slug string) ([]Benchmark, error) {
	if m.ListBenchmarksFn != nil {
		return m.ListBenchmarksFn(ctx, slug)
	}
	return []Benchmark{}, nil
}

func (m *MockClient) ListMembers(ctx context.Context, orgID string) ([]Member, error) {
	if m.ListMembersFn != nil {
		return m.ListMembersFn(ctx, orgID)
	}
	return []Member{{OrganizationID: orgID, UserID: "user-1", Role: "owner"}}, nil
}

func (m *MockClient) AddMember(ctx context.Context, orgID string, body map[string]string) (*Member, error) {
	if m.AddMemberFn != nil {
		return m.AddMemberFn(ctx, orgID, body)
	}
	return &Member{OrganizationID: orgID, UserID: body["user_id"], Role: body["role"]}, nil
}

func (m *MockClient) UpdateMemberRole(ctx context.Context, orgID, userID string, body map[string]string) (*Member, error) {
	if m.UpdateMemberRolFn != nil {
		return m.UpdateMemberRolFn(ctx, orgID, userID, body)
	}
	return &Member{OrganizationID: orgID, UserID: userID, Role: body["role"]}, nil
}

func (m *MockClient) RemoveMember(ctx context.Context, orgID, userID string) error {
	if m.RemoveMemberFn != nil {
		return m.RemoveMemberFn(ctx, orgID, userID)
	}
	return nil
}

func (m *MockClient) ListAPIKeys(ctx context.Context, orgID string) ([]APIKey, error) {
	if m.ListAPIKeysFn != nil {
		return m.ListAPIKeysFn(ctx, orgID)
	}
	return []APIKey{{ID: "key-1", OrganizationID: orgID, Name: "dev-key", KeyPrefix: "qh_abc"}}, nil
}

func (m *MockClient) CreateAPIKey(ctx context.Context, orgID string, body map[string]string) (*APIKeyCreated, error) {
	if m.CreateAPIKeyFn != nil {
		return m.CreateAPIKeyFn(ctx, orgID, body)
	}
	return &APIKeyCreated{ID: "key-new", OrganizationID: orgID, Name: body["name"], Key: "qh_full_key_123", KeyPrefix: "qh_ful"}, nil
}

func (m *MockClient) DeleteAPIKey(ctx context.Context, orgID, id string) error {
	if m.DeleteAPIKeyFn != nil {
		return m.DeleteAPIKeyFn(ctx, orgID, id)
	}
	return nil
}

func (m *MockClient) GetPromptAnalytics(ctx context.Context, promptID string) ([]PromptAnalytics, error) {
	if m.GetPromptAnalyticsFn != nil {
		return m.GetPromptAnalyticsFn(ctx, promptID)
	}
	return []PromptAnalytics{}, nil
}

func (m *MockClient) GetDailyTrend(ctx context.Context, promptID string, days string) ([]DailyTrend, error) {
	if m.GetDailyTrendFn != nil {
		return m.GetDailyTrendFn(ctx, promptID, days)
	}
	return []DailyTrend{}, nil
}

func (m *MockClient) GetProjectAnalytics(ctx context.Context, projectID string) ([]ProjectAnalytics, error) {
	if m.GetProjectAnalyticsFn != nil {
		return m.GetProjectAnalyticsFn(ctx, projectID)
	}
	return []ProjectAnalytics{}, nil
}

func (m *MockClient) GetVersionAnalytics(ctx context.Context, promptID string, version string) (*PromptAnalytics, error) {
	if m.GetVersionAnalyticsFn != nil {
		return m.GetVersionAnalyticsFn(ctx, promptID, version)
	}
	return &PromptAnalytics{}, nil
}

func (m *MockClient) GetSemanticDiff(ctx context.Context, promptID, v1, v2 string) (*SemanticDiff, error) {
	if m.GetSemanticDiffFn != nil {
		return m.GetSemanticDiffFn(ctx, promptID, v1, v2)
	}
	return &SemanticDiff{Summary: "Test diff", Changes: []DiffChange{{Category: "tone", Description: "Changed", Impact: "low"}}}, nil
}

func (m *MockClient) GetTextDiff(ctx context.Context, promptID, version string) (*TextDiffResult, error) {
	if m.GetTextDiffFn != nil {
		return m.GetTextDiffFn(ctx, promptID, version)
	}
	return &TextDiffResult{FromVersion: 1, ToVersion: 2, Hunks: []TextDiffHunk{}, Stats: TextDiffStats{}}, nil
}

func (m *MockClient) GetLintResult(ctx context.Context, promptID, version string) (*LintResult, error) {
	if m.GetLintResultFn != nil {
		return m.GetLintResultFn(ctx, promptID, version)
	}
	return &LintResult{Score: 85, Issues: []LintIssue{}, Passed: []string{"length-check"}}, nil
}

func (m *MockClient) CompareVersions(ctx context.Context, promptID, v1, v2 string) (*VersionComparison, error) {
	if m.CompareVersionsFn != nil {
		return m.CompareVersionsFn(ctx, promptID, v1, v2)
	}
	return &VersionComparison{PromptID: promptID, VersionA: 1, VersionB: 2, OverallWinner: "inconclusive"}, nil
}

func (m *MockClient) SemanticSearch(ctx context.Context, body map[string]any) (*SearchResponse, error) {
	if m.SemanticSearchFn != nil {
		return m.SemanticSearchFn(ctx, body)
	}
	q, _ := body["query"].(string)
	return &SearchResponse{Query: q, Results: []SearchResult{}, Total: 0}, nil
}

func (m *MockClient) GetEmbeddingStatus(ctx context.Context) (map[string]string, error) {
	if m.GetEmbeddingStatusFn != nil {
		return m.GetEmbeddingStatusFn(ctx)
	}
	return map[string]string{"status": "healthy"}, nil
}

// NewMockClientWithError returns a MockClient where all calls return the given error.
func NewMockClientWithError(err error) *MockClient {
	return &MockClient{
		GetOrganizationFn:   func(_ context.Context, _ string) (*Organization, error) { return nil, err },
		ListOrganizationsFn: func(_ context.Context) ([]Organization, error) { return nil, err },
		ListProjectsFn:      func(_ context.Context, _ string) ([]Project, error) { return nil, err },
		GetProjectFn:        func(_ context.Context, _, _ string) (*Project, error) { return nil, err },
		ListPromptsFn:       func(_ context.Context, _ string) ([]Prompt, error) { return nil, err },
		GetPromptFn:         func(_ context.Context, _, _ string) (*Prompt, error) { return nil, err },
		ListVersionsFn:      func(_ context.Context, _ string) ([]PromptVersion, error) { return nil, err },
		GetVersionFn:        func(_ context.Context, _, _ string) (*PromptVersion, error) { return nil, err },
		ListLogsFn:          func(_ context.Context) ([]ExecutionLog, error) { return nil, err },
		GetLogFn:            func(_ context.Context, _ string) (*ExecutionLog, error) { return nil, err },
		ListLogEvaluationFn: func(_ context.Context, _ string) ([]Evaluation, error) { return nil, err },
		ListEvaluationsFn:   func(_ context.Context) ([]Evaluation, error) { return nil, err },
		ListConsultingSessionsFn: func(_ context.Context) ([]ConsultingSession, error) {
			return nil, err
		},
		GetConsultingSessionFn: func(_ context.Context, _ string) (*ConsultingSession, error) {
			return nil, err
		},
		ListConsultingMessagesFn: func(_ context.Context, _ string) ([]ConsultingMessage, error) {
			return nil, err
		},
		ListTagsFn:       func(_ context.Context) ([]Tag, error) { return nil, err },
		ListIndustriesFn: func(_ context.Context) ([]Industry, error) { return nil, err },
		GetIndustryFn:    func(_ context.Context, _ string) (*Industry, error) { return nil, err },
		ListMembersFn:    func(_ context.Context, _ string) ([]Member, error) { return nil, err },
		ListAPIKeysFn:    func(_ context.Context, _ string) ([]APIKey, error) { return nil, err },
		GetPromptAnalyticsFn: func(_ context.Context, _ string) ([]PromptAnalytics, error) {
			return nil, err
		},
		GetDailyTrendFn: func(_ context.Context, _, _ string) ([]DailyTrend, error) { return nil, err },
		GetProjectAnalyticsFn: func(_ context.Context, _ string) ([]ProjectAnalytics, error) {
			return nil, err
		},
		SemanticSearchFn: func(_ context.Context, _ map[string]any) (*SearchResponse, error) {
			return nil, err
		},
		GetEmbeddingStatusFn: func(_ context.Context) (map[string]string, error) { return nil, err },
		CreateOrganizationFn: func(_ context.Context, _ map[string]string) (*Organization, error) {
			return nil, err
		},
		CreateProjectFn: func(_ context.Context, _ string, _ map[string]string) (*Project, error) {
			return nil, err
		},
		CreatePromptFn: func(_ context.Context, _ string, _ map[string]string) (*Prompt, error) {
			return nil, err
		},
		CreateVersionFn: func(_ context.Context, _ string, _ map[string]any) (*PromptVersion, error) {
			return nil, err
		},
		CreateTagFn: func(_ context.Context, _ map[string]string) (*Tag, error) { return nil, err },
		CreateConsultingSessionFn: func(_ context.Context, _ map[string]string) (*ConsultingSession, error) {
			return nil, err
		},
		SendConsultingMessageFn: func(_ context.Context, _ string, _ map[string]string) (*ConsultingMessage, error) {
			return nil, err
		},
		CheckComplianceFn: func(_ context.Context, _ string, _ map[string]string) (*ComplianceResult, error) {
			return nil, err
		},
		ListBenchmarksFn:   func(_ context.Context, _ string) ([]Benchmark, error) { return nil, err },
		CreateEvaluationFn: func(_ context.Context, _ string, _ map[string]any) (*Evaluation, error) { return nil, err },
		UpdateOrganizationFn: func(_ context.Context, _ string, _ map[string]string) (*Organization, error) {
			return nil, err
		},
		UpdateProjectFn: func(_ context.Context, _, _ string, _ map[string]string) (*Project, error) {
			return nil, err
		},
		DeleteProjectFn: func(_ context.Context, _, _ string) error { return err },
		UpdatePromptFn: func(_ context.Context, _, _ string, _ map[string]string) (*Prompt, error) {
			return nil, err
		},
		UpdateVersionStatusFn: func(_ context.Context, _, _, _ string) (*PromptVersion, error) {
			return nil, err
		},
		GetSemanticDiffFn: func(_ context.Context, _, _, _ string) (*SemanticDiff, error) {
			return nil, err
		},
		GetTextDiffFn: func(_ context.Context, _, _ string) (*TextDiffResult, error) { return nil, err },
		GetLintResultFn: func(_ context.Context, _, _ string) (*LintResult, error) {
			return nil, err
		},
		CompareVersionsFn: func(_ context.Context, _, _, _ string) (*VersionComparison, error) {
			return nil, err
		},
		CloseSessionFn: func(_ context.Context, _ string) (*ConsultingSession, error) {
			return nil, err
		},
		AddMemberFn: func(_ context.Context, _ string, _ map[string]string) (*Member, error) {
			return nil, err
		},
		UpdateMemberRolFn: func(_ context.Context, _, _ string, _ map[string]string) (*Member, error) {
			return nil, err
		},
		RemoveMemberFn: func(_ context.Context, _, _ string) error { return err },
		CreateAPIKeyFn: func(_ context.Context, _ string, _ map[string]string) (*APIKeyCreated, error) {
			return nil, err
		},
		DeleteAPIKeyFn:   func(_ context.Context, _, _ string) error { return err },
		DeleteTagFn:      func(_ context.Context, _ string) error { return err },
		CreateIndustryFn: func(_ context.Context, _ map[string]string) (*Industry, error) { return nil, err },
		UpdateIndustryFn: func(_ context.Context, _ string, _ map[string]string) (*Industry, error) { return nil, err },
		GetVersionAnalyticsFn: func(_ context.Context, _, _ string) (*PromptAnalytics, error) {
			return nil, err
		},
	}
}
