package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client calls the Hugging Face Text Embeddings Inference (TEI) API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a TEI embedding client.
// baseURL is the TEI server URL, e.g. "http://embedding:80".
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// embedRequest is the TEI /embed request body.
type embedRequest struct {
	Inputs    []string `json:"inputs"`
	Normalize bool     `json:"normalize"`
	Truncate  bool     `json:"truncate"`
}

// Embed generates embeddings for the given texts.
// Returns a slice of float32 slices, one per input text.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embedRequest{
		Inputs:    texts,
		Normalize: true,
		Truncate:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding: TEI returned %d: %s", resp.StatusCode, string(respBody))
	}

	var embeddings [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		return nil, fmt.Errorf("embedding: decode response: %w", err)
	}

	return embeddings, nil
}

// EmbedOne generates an embedding for a single text.
func (c *Client) EmbedOne(ctx context.Context, text string) ([]float32, error) {
	results, err := c.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("embedding: empty response")
	}
	return results[0], nil
}

// Health checks if the TEI server is healthy.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("embedding: health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("embedding: unhealthy (status %d)", resp.StatusCode)
	}
	return nil
}
