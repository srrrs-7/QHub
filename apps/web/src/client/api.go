package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		authToken:  "dev-token",
	}
}

func (c *APIClient) do(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// --- Organizations ---

func (c *APIClient) GetOrganization(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	return &org, c.do(ctx, http.MethodGet, "/api/v1/organizations/"+slug, nil, &org)
}

// --- Projects ---

func (c *APIClient) ListProjects(ctx context.Context, orgID string) ([]Project, error) {
	var projects []Project
	return projects, c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/projects", nil, &projects)
}

func (c *APIClient) GetProject(ctx context.Context, orgID, slug string) (*Project, error) {
	var project Project
	return &project, c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/projects/"+slug, nil, &project)
}

// --- Prompts ---

func (c *APIClient) ListPrompts(ctx context.Context, projectID string) ([]Prompt, error) {
	var prompts []Prompt
	return prompts, c.do(ctx, http.MethodGet, "/api/v1/projects/"+projectID+"/prompts", nil, &prompts)
}

func (c *APIClient) GetPrompt(ctx context.Context, projectID, slug string) (*Prompt, error) {
	var prompt Prompt
	return &prompt, c.do(ctx, http.MethodGet, "/api/v1/projects/"+projectID+"/prompts/"+slug, nil, &prompt)
}

func (c *APIClient) CreatePrompt(ctx context.Context, projectID string, body map[string]string) (*Prompt, error) {
	var prompt Prompt
	return &prompt, c.do(ctx, http.MethodPost, "/api/v1/projects/"+projectID+"/prompts", body, &prompt)
}

// --- Versions ---

func (c *APIClient) ListVersions(ctx context.Context, promptID string) ([]PromptVersion, error) {
	var versions []PromptVersion
	return versions, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/versions", nil, &versions)
}

func (c *APIClient) GetVersion(ctx context.Context, promptID, version string) (*PromptVersion, error) {
	var v PromptVersion
	return &v, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/versions/"+version, nil, &v)
}

func (c *APIClient) CreateVersion(ctx context.Context, promptID string, body map[string]any) (*PromptVersion, error) {
	var v PromptVersion
	return &v, c.do(ctx, http.MethodPost, "/api/v1/prompts/"+promptID+"/versions", body, &v)
}

func (c *APIClient) UpdateVersionStatus(ctx context.Context, promptID, version, status string) (*PromptVersion, error) {
	var v PromptVersion
	return &v, c.do(ctx, http.MethodPut, "/api/v1/prompts/"+promptID+"/versions/"+version+"/status", map[string]string{"status": status}, &v)
}

// --- Organizations ---

func (c *APIClient) ListOrganizations(ctx context.Context) ([]Organization, error) {
	var orgs []Organization
	// The API doesn't have a list-all endpoint; we use a search or just return empty.
	// For now, we rely on a convention: the index page will link to known orgs.
	return orgs, c.do(ctx, http.MethodGet, "/api/v1/organizations", nil, &orgs)
}

func (c *APIClient) CreateOrganization(ctx context.Context, body map[string]string) (*Organization, error) {
	var org Organization
	return &org, c.do(ctx, http.MethodPost, "/api/v1/organizations", body, &org)
}

// --- Execution Logs ---

func (c *APIClient) ListLogs(ctx context.Context) ([]ExecutionLog, error) {
	var logs []ExecutionLog
	return logs, c.do(ctx, http.MethodGet, "/api/v1/logs", nil, &logs)
}

func (c *APIClient) GetLog(ctx context.Context, id string) (*ExecutionLog, error) {
	var l ExecutionLog
	return &l, c.do(ctx, http.MethodGet, "/api/v1/logs/"+id, nil, &l)
}

func (c *APIClient) ListLogEvaluations(ctx context.Context, logID string) ([]Evaluation, error) {
	var evals []Evaluation
	return evals, c.do(ctx, http.MethodGet, "/api/v1/logs/"+logID+"/evaluations", nil, &evals)
}

// --- Consulting ---

func (c *APIClient) ListConsultingSessions(ctx context.Context) ([]ConsultingSession, error) {
	var sessions []ConsultingSession
	return sessions, c.do(ctx, http.MethodGet, "/api/v1/consulting/sessions", nil, &sessions)
}

func (c *APIClient) CreateConsultingSession(ctx context.Context, body map[string]string) (*ConsultingSession, error) {
	var session ConsultingSession
	return &session, c.do(ctx, http.MethodPost, "/api/v1/consulting/sessions", body, &session)
}

func (c *APIClient) GetConsultingSession(ctx context.Context, id string) (*ConsultingSession, error) {
	var session ConsultingSession
	return &session, c.do(ctx, http.MethodGet, "/api/v1/consulting/sessions/"+id, nil, &session)
}

func (c *APIClient) ListConsultingMessages(ctx context.Context, sessionID string) ([]ConsultingMessage, error) {
	var msgs []ConsultingMessage
	return msgs, c.do(ctx, http.MethodGet, "/api/v1/consulting/sessions/"+sessionID+"/messages", nil, &msgs)
}

func (c *APIClient) SendConsultingMessage(ctx context.Context, sessionID string, body map[string]string) (*ConsultingMessage, error) {
	var msg ConsultingMessage
	return &msg, c.do(ctx, http.MethodPost, "/api/v1/consulting/sessions/"+sessionID+"/messages", body, &msg)
}

// --- Tags ---

func (c *APIClient) ListTags(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	return tags, c.do(ctx, http.MethodGet, "/api/v1/tags", nil, &tags)
}

func (c *APIClient) CreateTag(ctx context.Context, body map[string]string) (*Tag, error) {
	var tag Tag
	return &tag, c.do(ctx, http.MethodPost, "/api/v1/tags", body, &tag)
}

func (c *APIClient) DeleteTag(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/tags?name="+name, nil, nil)
}

// --- Industries ---

func (c *APIClient) ListIndustries(ctx context.Context) ([]Industry, error) {
	var industries []Industry
	return industries, c.do(ctx, http.MethodGet, "/api/v1/industries", nil, &industries)
}

func (c *APIClient) GetIndustry(ctx context.Context, slug string) (*Industry, error) {
	var ind Industry
	return &ind, c.do(ctx, http.MethodGet, "/api/v1/industries/"+slug, nil, &ind)
}

func (c *APIClient) CreateIndustry(ctx context.Context, body map[string]string) (*Industry, error) {
	var ind Industry
	return &ind, c.do(ctx, http.MethodPost, "/api/v1/industries", body, &ind)
}

func (c *APIClient) UpdateIndustry(ctx context.Context, slug string, body map[string]string) (*Industry, error) {
	var ind Industry
	return &ind, c.do(ctx, http.MethodPut, "/api/v1/industries/"+slug, body, &ind)
}

func (c *APIClient) CheckCompliance(ctx context.Context, slug string, body map[string]string) (*ComplianceResult, error) {
	var result ComplianceResult
	return &result, c.do(ctx, http.MethodPost, "/api/v1/industries/"+slug+"/compliance", body, &result)
}

func (c *APIClient) ListBenchmarks(ctx context.Context, slug string) ([]Benchmark, error) {
	var benchmarks []Benchmark
	return benchmarks, c.do(ctx, http.MethodGet, "/api/v1/industries/"+slug+"/benchmarks", nil, &benchmarks)
}

// --- Members ---

func (c *APIClient) ListMembers(ctx context.Context, orgID string) ([]Member, error) {
	var members []Member
	return members, c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/members", nil, &members)
}

func (c *APIClient) AddMember(ctx context.Context, orgID string, body map[string]string) (*Member, error) {
	var member Member
	return &member, c.do(ctx, http.MethodPost, "/api/v1/organizations/"+orgID+"/members", body, &member)
}

func (c *APIClient) UpdateMemberRole(ctx context.Context, orgID, userID string, body map[string]string) (*Member, error) {
	var member Member
	return &member, c.do(ctx, http.MethodPut, "/api/v1/organizations/"+orgID+"/members/"+userID, body, &member)
}

func (c *APIClient) RemoveMember(ctx context.Context, orgID, userID string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/organizations/"+orgID+"/members/"+userID, nil, nil)
}

// --- API Keys ---

func (c *APIClient) ListAPIKeys(ctx context.Context, orgID string) ([]APIKey, error) {
	var keys []APIKey
	return keys, c.do(ctx, http.MethodGet, "/api/v1/organizations/"+orgID+"/api-keys", nil, &keys)
}

func (c *APIClient) CreateAPIKey(ctx context.Context, orgID string, body map[string]string) (*APIKeyCreated, error) {
	var key APIKeyCreated
	return &key, c.do(ctx, http.MethodPost, "/api/v1/organizations/"+orgID+"/api-keys", body, &key)
}

func (c *APIClient) DeleteAPIKey(ctx context.Context, orgID, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/organizations/"+orgID+"/api-keys/"+id, nil, nil)
}

// --- Update Organization ---

func (c *APIClient) UpdateOrganization(ctx context.Context, slug string, body map[string]string) (*Organization, error) {
	var org Organization
	return &org, c.do(ctx, http.MethodPut, "/api/v1/organizations/"+slug, body, &org)
}

// --- Analytics ---

func (c *APIClient) GetPromptAnalytics(ctx context.Context, promptID string) ([]PromptAnalytics, error) {
	var result []PromptAnalytics
	return result, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/analytics", nil, &result)
}

func (c *APIClient) GetDailyTrend(ctx context.Context, promptID string, days string) ([]DailyTrend, error) {
	var result []DailyTrend
	path := "/api/v1/prompts/" + promptID + "/trend"
	if days != "" {
		path += "?days=" + days
	}
	return result, c.do(ctx, http.MethodGet, path, nil, &result)
}

// --- Project Analytics ---

func (c *APIClient) GetProjectAnalytics(ctx context.Context, projectID string) ([]ProjectAnalytics, error) {
	var result []ProjectAnalytics
	return result, c.do(ctx, http.MethodGet, "/api/v1/projects/"+projectID+"/analytics", nil, &result)
}

// --- Version Analytics ---

func (c *APIClient) GetVersionAnalytics(ctx context.Context, promptID string, version string) (*PromptAnalytics, error) {
	var result PromptAnalytics
	return &result, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/versions/"+version+"/analytics", nil, &result)
}

// --- Close Session ---

func (c *APIClient) CloseSession(ctx context.Context, sessionID string) (*ConsultingSession, error) {
	var session ConsultingSession
	body := map[string]string{"status": "closed"}
	return &session, c.do(ctx, http.MethodPut, "/api/v1/consulting/sessions/"+sessionID, body, &session)
}

// --- Embedding Status ---

func (c *APIClient) GetEmbeddingStatus(ctx context.Context) (map[string]string, error) {
	var result map[string]string
	return result, c.do(ctx, http.MethodGet, "/api/v1/search/embedding-status", nil, &result)
}

// --- Diff ---

func (c *APIClient) GetSemanticDiff(ctx context.Context, promptID, v1, v2 string) (*SemanticDiff, error) {
	var result SemanticDiff
	return &result, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/semantic-diff/"+v1+"/"+v2, nil, &result)
}

func (c *APIClient) GetTextDiff(ctx context.Context, promptID, version string) (*TextDiffResult, error) {
	var result TextDiffResult
	return &result, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/versions/"+version+"/text-diff", nil, &result)
}

// --- Lint ---

func (c *APIClient) GetLintResult(ctx context.Context, promptID, version string) (*LintResult, error) {
	var result LintResult
	return &result, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/versions/"+version+"/lint", nil, &result)
}

// --- Search ---

func (c *APIClient) SemanticSearch(ctx context.Context, body map[string]any) (*SearchResponse, error) {
	var result SearchResponse
	return &result, c.do(ctx, http.MethodPost, "/api/v1/search/semantic", body, &result)
}

// --- Evaluations ---

func (c *APIClient) ListEvaluations(ctx context.Context) ([]Evaluation, error) {
	var evals []Evaluation
	return evals, c.do(ctx, http.MethodGet, "/api/v1/evaluations", nil, &evals)
}

// --- Projects (CRUD) ---

func (c *APIClient) CreateProject(ctx context.Context, orgID string, body map[string]string) (*Project, error) {
	var project Project
	return &project, c.do(ctx, http.MethodPost, "/api/v1/organizations/"+orgID+"/projects", body, &project)
}

func (c *APIClient) UpdateProject(ctx context.Context, orgID, slug string, body map[string]string) (*Project, error) {
	var project Project
	return &project, c.do(ctx, http.MethodPut, "/api/v1/organizations/"+orgID+"/projects/"+slug, body, &project)
}

func (c *APIClient) DeleteProject(ctx context.Context, orgID, slug string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/organizations/"+orgID+"/projects/"+slug, nil, nil)
}

// --- Prompts (update) ---

func (c *APIClient) UpdatePrompt(ctx context.Context, projectID, slug string, body map[string]string) (*Prompt, error) {
	var prompt Prompt
	return &prompt, c.do(ctx, http.MethodPut, "/api/v1/projects/"+projectID+"/prompts/"+slug, body, &prompt)
}

// --- Version Comparison ---

func (c *APIClient) CompareVersions(ctx context.Context, promptID, v1, v2 string) (*VersionComparison, error) {
	var result VersionComparison
	return &result, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/versions/"+v1+"/"+v2+"/compare", nil, &result)
}

// --- Evaluations (create/get/update) ---

func (c *APIClient) CreateEvaluation(ctx context.Context, logID string, body map[string]any) (*Evaluation, error) {
	var eval Evaluation
	return &eval, c.do(ctx, http.MethodPost, "/api/v1/logs/"+logID+"/evaluations", body, &eval)
}

func (c *APIClient) GetEvaluation(ctx context.Context, id string) (*Evaluation, error) {
	var eval Evaluation
	return &eval, c.do(ctx, http.MethodGet, "/api/v1/evaluations/"+id, nil, &eval)
}

func (c *APIClient) UpdateEvaluation(ctx context.Context, id string, body map[string]any) (*Evaluation, error) {
	var eval Evaluation
	return &eval, c.do(ctx, http.MethodPut, "/api/v1/evaluations/"+id, body, &eval)
}

// --- Prompt Tags ---

func (c *APIClient) ListPromptTags(ctx context.Context, promptID string) ([]Tag, error) {
	var tags []Tag
	return tags, c.do(ctx, http.MethodGet, "/api/v1/prompts/"+promptID+"/tags", nil, &tags)
}

func (c *APIClient) AddPromptTag(ctx context.Context, promptID, tagID string) error {
	return c.do(ctx, http.MethodPost, "/api/v1/prompts/"+promptID+"/tags", map[string]string{"tag_id": tagID}, nil)
}

func (c *APIClient) RemovePromptTag(ctx context.Context, promptID, tagID string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/prompts/"+promptID+"/tags/"+tagID, nil, nil)
}

// --- Logs (create) ---

func (c *APIClient) CreateLog(ctx context.Context, body map[string]any) (*ExecutionLog, error) {
	var l ExecutionLog
	return &l, c.do(ctx, http.MethodPost, "/api/v1/logs", body, &l)
}
