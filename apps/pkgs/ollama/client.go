// Package ollama provides an HTTP client for the Ollama LLM inference API.
//
// It supports both streaming and synchronous chat completions.
// When the Ollama server is not configured (empty baseURL), the client
// reports itself as unavailable via the Available method.
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with an Ollama server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request body for the Ollama /api/chat endpoint.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatResponse is a single response chunk from the Ollama /api/chat endpoint.
type ChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

// NewClient creates an Ollama client for the given base URL.
// If baseURL is empty, the client will report as unavailable.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Available returns true if the client has a configured base URL.
func (c *Client) Available() bool {
	return c != nil && c.baseURL != ""
}

// Chat sends a streaming chat request and returns a channel of ChatResponse chunks.
// The channel is closed when the response is complete or an error occurs.
// Callers should read from the channel until it is closed.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (<-chan ChatResponse, error) {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan ChatResponse, 16)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var chunk ChatResponse
			if err := json.Unmarshal(line, &chunk); err != nil {
				return
			}

			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}

			if chunk.Done {
				return
			}
		}
	}()

	return ch, nil
}

// ChatSync sends a non-streaming chat request and returns the complete response message.
func (c *Client) ChatSync(ctx context.Context, req ChatRequest) (ChatMessage, error) {
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("ollama: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return ChatMessage{}, fmt.Errorf("ollama: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatMessage{}, fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return ChatMessage{}, fmt.Errorf("ollama: server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return ChatMessage{}, fmt.Errorf("ollama: decode response: %w", err)
	}

	return chatResp.Message, nil
}

// Health checks if the Ollama server is reachable.
func (c *Client) Health(ctx context.Context) error {
	if !c.Available() {
		return fmt.Errorf("ollama: client not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("ollama: create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama: health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama: unhealthy (status %d)", resp.StatusCode)
	}

	return nil
}
