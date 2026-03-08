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
