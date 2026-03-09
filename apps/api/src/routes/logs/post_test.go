package logs

import (
	"api/src/infra/rds/executionlog_repository"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"utils/db/db"
	"utils/testutil"

	"github.com/google/go-cmp/cmp"
)

// seedOrgAndPrompt creates the prerequisite chain: organization -> project -> prompt
// and returns the org ID and prompt ID for use in execution log tests.
// Uses a unique numeric suffix to ensure valid slugs matching ^[a-z0-9][a-z0-9-]*[a-z0-9]$.
var seedCounter int

func seedOrgAndPrompt(t *testing.T, q db.Querier, label string) (orgID, promptID string) {
	t.Helper()
	ctx := context.Background()
	seedCounter++
	slug := fmt.Sprintf("s%d", seedCounter)

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Org " + label,
		Slug: "org-" + slug,
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	proj, err := q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: org.ID,
		Name:           "Project " + label,
		Slug:           "proj-" + slug,
		Description:    sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	prompt, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		ProjectID:  proj.ID,
		Name:       "Prompt " + label,
		Slug:       "prompt-" + slug,
		PromptType: "system",
	})
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}

	return org.ID.String(), prompt.ID.String()
}

func validPostLogBody(orgID, promptID string) string {
	return fmt.Sprintf(`{
		"org_id": %q,
		"prompt_id": %q,
		"version_number": 1,
		"request_body": {"prompt": "hello"},
		"response_body": {"text": "world"},
		"model": "claude-sonnet-4-20250514",
		"provider": "anthropic",
		"input_tokens": 100,
		"output_tokens": 50,
		"total_tokens": 150,
		"latency_ms": 1200,
		"estimated_cost": "0.003000",
		"status": "success",
		"environment": "production",
		"executed_at": %q
	}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
}

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			model      string
			provider   string
			status     string
		}

		tests := []struct {
			testName string
			body     func(orgID, promptID string) string
			expected expected
		}{
			// 正常系 (Happy Path)
			{
				testName: "valid execution log with all fields",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "hello"},
						"response_body": {"text": "world"},
						"model": "claude-sonnet-4-20250514",
						"provider": "anthropic",
						"input_tokens": 100,
						"output_tokens": 50,
						"total_tokens": 150,
						"latency_ms": 1200,
						"estimated_cost": "0.003000",
						"status": "success",
						"environment": "production",
						"metadata": {"key": "value"},
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "claude-sonnet-4-20250514",
					provider:   "anthropic",
					status:     "success",
				},
			},
			{
				testName: "valid execution log with error status",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "hello"},
						"model": "gpt-4",
						"provider": "openai",
						"input_tokens": 50,
						"output_tokens": 0,
						"total_tokens": 50,
						"latency_ms": 500,
						"estimated_cost": "0.001000",
						"status": "error",
						"error_message": "rate limit exceeded",
						"environment": "staging",
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "gpt-4",
					provider:   "openai",
					status:     "error",
				},
			},
			// 境界値 (Boundary Values)
			{
				testName: "zero token counts and latency",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "test"},
						"model": "model-x",
						"provider": "provider-y",
						"input_tokens": 0,
						"output_tokens": 0,
						"total_tokens": 0,
						"latency_ms": 0,
						"estimated_cost": "0.000000",
						"status": "success",
						"environment": "development",
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "model-x",
					provider:   "provider-y",
					status:     "success",
				},
			},
			{
				testName: "high token counts and latency",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "big request"},
						"model": "gpt-4-turbo",
						"provider": "openai",
						"input_tokens": 128000,
						"output_tokens": 32000,
						"total_tokens": 160000,
						"latency_ms": 300000,
						"estimated_cost": "999.999999",
						"status": "success",
						"environment": "production",
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "gpt-4-turbo",
					provider:   "openai",
					status:     "success",
				},
			},
			// 特殊文字 (Special Characters)
			{
				testName: "unicode in model name and request body",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "日本語のプロンプト 🤖"},
						"response_body": {"text": "応答テスト ✅"},
						"model": "claude-日本語",
						"provider": "anthropic",
						"input_tokens": 10,
						"output_tokens": 5,
						"total_tokens": 15,
						"latency_ms": 100,
						"estimated_cost": "0.000100",
						"status": "success",
						"environment": "development",
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "claude-日本語",
					provider:   "anthropic",
					status:     "success",
				},
			},
			{
				testName: "special characters in error message",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "test"},
						"model": "model",
						"provider": "provider",
						"input_tokens": 10,
						"output_tokens": 0,
						"total_tokens": 10,
						"latency_ms": 50,
						"estimated_cost": "0.000010",
						"status": "error",
						"error_message": "Error: <script>alert('xss')</script> & \"quotes\"",
						"environment": "development",
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "model",
					provider:   "provider",
					status:     "error",
				},
			},
			// Null/Nil optional fields
			{
				testName: "nil optional fields (response_body, metadata, error_message)",
				body: func(orgID, promptID string) string {
					return fmt.Sprintf(`{
						"org_id": %q,
						"prompt_id": %q,
						"version_number": 1,
						"request_body": {"prompt": "test"},
						"model": "model",
						"provider": "provider",
						"estimated_cost": "0.000000",
						"status": "success",
						"environment": "production",
						"executed_at": %q
					}`, orgID, promptID, time.Now().UTC().Format(time.RFC3339))
				},
				expected: expected{
					statusCode: http.StatusCreated,
					model:      "model",
					provider:   "provider",
					status:     "success",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID, promptID := seedOrgAndPrompt(t, q, tt.testName)

				body := tt.body(orgID, promptID)
				req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.model, result["model"]); diff != "" {
					t.Errorf("model mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.provider, result["provider"]); diff != "" {
					t.Errorf("provider mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.status, result["status"]); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}

				// Verify ID is returned
				if result["id"] == nil || result["id"] == "" {
					t.Error("expected non-empty id in response")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			body           string
			expectedStatus int
		}{
			// 異常系 (Error Cases)
			{
				testName:       "invalid JSON",
				body:           `{invalid json`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "empty body",
				body:           ``,
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字 (Empty String)
			{
				testName: "empty model",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "",
					"provider": "anthropic",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "empty provider",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "missing required org_id",
				body: fmt.Sprintf(`{
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "missing required prompt_id",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "invalid org_id format",
				body: fmt.Sprintf(`{
					"org_id": "not-a-uuid",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "invalid status value",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "invalid_status",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "invalid environment value",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "invalid_env",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			// 境界値 (Boundary Values)
			{
				testName: "version_number zero (below min=1)",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 0,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "negative input_tokens",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"input_tokens": -1,
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "model exceeds max length",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": %q,
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "success",
					"environment": "production",
					"executed_at": %q
				}`, strings.Repeat("a", 101), time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "error_message exceeds max length",
				body: fmt.Sprintf(`{
					"org_id": "00000000-0000-0000-0000-000000000001",
					"prompt_id": "00000000-0000-0000-0000-000000000002",
					"version_number": 1,
					"request_body": {"prompt": "test"},
					"model": "gpt-4",
					"provider": "openai",
					"estimated_cost": "0.001",
					"status": "error",
					"error_message": %q,
					"environment": "production",
					"executed_at": %q
				}`, strings.Repeat("e", 2001), time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}
