package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "http://localhost:8080"

// Client is the PromptLab SDK client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client for the SDK client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) {
		cl.httpClient = c
	}
}

// WithBaseURL sets the base URL for the PromptLab API.
func WithBaseURL(url string) Option {
	return func(cl *Client) {
		cl.baseURL = url
	}
}

// NewClient creates a new PromptLab SDK client.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// GetPromptLatest retrieves the latest version of a prompt by slug.
func (c *Client) GetPromptLatest(ctx context.Context, promptSlug string) (*PromptVersion, error) {
	path := fmt.Sprintf("/api/v1/prompts/%s/versions/latest", promptSlug)
	var pv PromptVersion
	if err := c.do(ctx, http.MethodGet, path, nil, &pv); err != nil {
		return nil, err
	}
	return &pv, nil
}

// GetPromptVersion retrieves a specific version of a prompt by slug and version number.
func (c *Client) GetPromptVersion(ctx context.Context, promptSlug string, version int) (*PromptVersion, error) {
	path := fmt.Sprintf("/api/v1/prompts/%s/versions/%d", promptSlug, version)
	var pv PromptVersion
	if err := c.do(ctx, http.MethodGet, path, nil, &pv); err != nil {
		return nil, err
	}
	return &pv, nil
}

// Log creates an execution log entry.
func (c *Client) Log(ctx context.Context, log *ExecutionLog) (*ExecutionLog, error) {
	var result ExecutionLog
	if err := c.do(ctx, http.MethodPost, "/api/v1/logs", log, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LogBatch creates multiple execution log entries in a single request.
func (c *Client) LogBatch(ctx context.Context, logs []ExecutionLog) ([]ExecutionLog, error) {
	var result []ExecutionLog
	if err := c.do(ctx, http.MethodPost, "/api/v1/logs/batch", logs, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Evaluate creates an evaluation for a prompt execution.
func (c *Client) Evaluate(ctx context.Context, eval *Evaluation) (*Evaluation, error) {
	var result Evaluation
	if err := c.do(ctx, http.MethodPost, "/api/v1/evaluations", eval, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// do executes an HTTP request against the PromptLab API.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := string(respBody)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    msg,
		}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshal response body: %w", err)
		}
	}

	return nil
}
