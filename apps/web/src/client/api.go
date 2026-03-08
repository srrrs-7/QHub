package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		authToken: "dev-token",
	}
}

// --- Generic helpers ---

func (c *APIClient) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	return c.doJSON(req, result)
}

func (c *APIClient) post(ctx context.Context, path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	return c.doJSON(req, result)
}

func (c *APIClient) put(ctx context.Context, path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	return c.doJSON(req, result)
}

func (c *APIClient) doJSON(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
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
	err := c.get(ctx, "/api/v1/organizations/"+slug, &org)
	return &org, err
}

// --- Projects ---

func (c *APIClient) ListProjects(ctx context.Context, orgID string) ([]Project, error) {
	var projects []Project
	err := c.get(ctx, "/api/v1/organizations/"+orgID+"/projects", &projects)
	return projects, err
}

func (c *APIClient) GetProject(ctx context.Context, orgID, slug string) (*Project, error) {
	var project Project
	err := c.get(ctx, "/api/v1/organizations/"+orgID+"/projects/"+slug, &project)
	return &project, err
}

func (c *APIClient) CreateProject(ctx context.Context, orgID string, body map[string]string) (*Project, error) {
	var project Project
	err := c.post(ctx, "/api/v1/organizations/"+orgID+"/projects", body, &project)
	return &project, err
}

// --- Prompts ---

func (c *APIClient) ListPrompts(ctx context.Context, projectID string) ([]Prompt, error) {
	var prompts []Prompt
	err := c.get(ctx, "/api/v1/projects/"+projectID+"/prompts", &prompts)
	return prompts, err
}

func (c *APIClient) GetPrompt(ctx context.Context, projectID, slug string) (*Prompt, error) {
	var prompt Prompt
	err := c.get(ctx, "/api/v1/projects/"+projectID+"/prompts/"+slug, &prompt)
	return &prompt, err
}

func (c *APIClient) CreatePrompt(ctx context.Context, projectID string, body map[string]string) (*Prompt, error) {
	var prompt Prompt
	err := c.post(ctx, "/api/v1/projects/"+projectID+"/prompts", body, &prompt)
	return &prompt, err
}

// --- Versions ---

func (c *APIClient) ListVersions(ctx context.Context, promptID string) ([]PromptVersion, error) {
	var versions []PromptVersion
	err := c.get(ctx, "/api/v1/prompts/"+promptID+"/versions", &versions)
	return versions, err
}

func (c *APIClient) GetVersion(ctx context.Context, promptID, version string) (*PromptVersion, error) {
	var v PromptVersion
	err := c.get(ctx, "/api/v1/prompts/"+promptID+"/versions/"+version, &v)
	return &v, err
}

func (c *APIClient) CreateVersion(ctx context.Context, promptID string, body map[string]any) (*PromptVersion, error) {
	var v PromptVersion
	err := c.post(ctx, "/api/v1/prompts/"+promptID+"/versions", body, &v)
	return &v, err
}

func (c *APIClient) UpdateVersionStatus(ctx context.Context, promptID, version string, status string) (*PromptVersion, error) {
	var v PromptVersion
	err := c.put(ctx, "/api/v1/prompts/"+promptID+"/versions/"+version+"/status", map[string]string{"status": status}, &v)
	return &v, err
}

// --- Legacy Task methods ---

func (c *APIClient) ListTasks(ctx context.Context) (*TasksResponse, error) {
	var result TasksResponse
	err := c.get(ctx, "/api/v1/tasks", &result)
	return &result, err
}

func (c *APIClient) CreateTask(ctx context.Context, task CreateTaskRequest) (*Task, error) {
	var result Task
	err := c.post(ctx, "/api/v1/tasks", task, &result)
	return &result, err
}
