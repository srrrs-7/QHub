package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	apiURL    string
	authToken string
	outputFmt string
)

var rootCmd = &cobra.Command{
	Use:     "qhub",
	Short:   "QHub CLI — Prompt version management and intelligence",
	Version: "0.1.0",
	Long: `QHub is a prompt/answer version management system with consulting,
execution logging, and prompt intelligence features.

Key features:
  • Prompt version management with lifecycle (draft -> review -> production)
  • Semantic diff and quality linting for prompt versions
  • Execution log tracking and evaluation analytics
  • AI consulting chat sessions with industry-specific configurations
  • Semantic search across prompts via embeddings
  • Organization, project, and member management

Environment variables:
  QHUB_API_URL   API server URL (default: http://localhost:8080)
  QHUB_TOKEN     Authentication token (default: dev-token)

Get started:
  qhub org create --name "My Org" --slug my-org
  qhub project --org <org-id> create --name "My Project" --slug my-proj
  qhub prompt --project <proj-id> create --name "My Prompt" --slug my-prompt
  qhub version --prompt <prompt-id> create --content "You are a helpful assistant"`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", envOrDefault("QHUB_API_URL", "http://localhost:8080"), "API server URL")
	rootCmd.PersistentFlags().StringVar(&authToken, "token", envOrDefault("QHUB_TOKEN", "dev-token"), "Authentication token")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format: table, json")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// --- API client helpers ---

func apiGet(path string, result any) error {
	return apiDo(http.MethodGet, path, nil, result)
}

func apiPost(path string, body any, result any) error {
	return apiDo(http.MethodPost, path, body, result)
}

func apiPut(path string, body any, result any) error {
	return apiDo(http.MethodPut, path, body, result)
}

func apiDelete(path string) error {
	return apiDo(http.MethodDelete, path, nil, nil)
}

func apiDo(method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, apiURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
