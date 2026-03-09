package logs

import (
	"api/src/infra/rds/executionlog_repository"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"utils/testutil"

	"github.com/google/go-cmp/cmp"
)

func TestPostBatchHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			count      int
		}

		tests := []struct {
			testName string
			body     func(orgID, promptID string) string
			expected expected
		}{
			// 正常系 (Happy Path)
			{
				testName: "batch create two logs",
				body: func(orgID, promptID string) string {
					now := time.Now().UTC().Format(time.RFC3339)
					return fmt.Sprintf(`{
						"logs": [
							{
								"org_id": %q,
								"prompt_id": %q,
								"version_number": 1,
								"request_body": {"prompt": "first"},
								"model": "gpt-4",
								"provider": "openai",
								"estimated_cost": "0.001000",
								"status": "success",
								"environment": "production",
								"executed_at": %q
							},
							{
								"org_id": %q,
								"prompt_id": %q,
								"version_number": 1,
								"request_body": {"prompt": "second"},
								"model": "claude-sonnet-4-20250514",
								"provider": "anthropic",
								"estimated_cost": "0.002000",
								"status": "success",
								"environment": "staging",
								"executed_at": %q
							}
						]
					}`, orgID, promptID, now, orgID, promptID, now)
				},
				expected: expected{
					statusCode: http.StatusCreated,
					count:      2,
				},
			},
			// 境界値 (Boundary Values)
			{
				testName: "batch create single log (min=1)",
				body: func(orgID, promptID string) string {
					now := time.Now().UTC().Format(time.RFC3339)
					return fmt.Sprintf(`{
						"logs": [
							{
								"org_id": %q,
								"prompt_id": %q,
								"version_number": 1,
								"request_body": {"prompt": "single"},
								"model": "gpt-4",
								"provider": "openai",
								"estimated_cost": "0.001000",
								"status": "success",
								"environment": "development",
								"executed_at": %q
							}
						]
					}`, orgID, promptID, now)
				},
				expected: expected{
					statusCode: http.StatusCreated,
					count:      1,
				},
			},
			// 特殊文字 (Special Characters)
			{
				testName: "batch create with unicode content",
				body: func(orgID, promptID string) string {
					now := time.Now().UTC().Format(time.RFC3339)
					return fmt.Sprintf(`{
						"logs": [
							{
								"org_id": %q,
								"prompt_id": %q,
								"version_number": 1,
								"request_body": {"prompt": "日本語テスト 🤖"},
								"response_body": {"text": "応答テスト ✅"},
								"model": "model-日本語",
								"provider": "provider-テスト",
								"estimated_cost": "0.001000",
								"status": "success",
								"environment": "production",
								"executed_at": %q
							}
						]
					}`, orgID, promptID, now)
				},
				expected: expected{
					statusCode: http.StatusCreated,
					count:      1,
				},
			},
			// Null/Nil - optional fields omitted
			{
				testName: "batch create with minimal required fields",
				body: func(orgID, promptID string) string {
					now := time.Now().UTC().Format(time.RFC3339)
					return fmt.Sprintf(`{
						"logs": [
							{
								"org_id": %q,
								"prompt_id": %q,
								"version_number": 1,
								"request_body": {"prompt": "minimal"},
								"model": "model",
								"provider": "provider",
								"estimated_cost": "0.000000",
								"status": "success",
								"environment": "production",
								"executed_at": %q
							}
						]
					}`, orgID, promptID, now)
				},
				expected: expected{
					statusCode: http.StatusCreated,
					count:      1,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID, promptID := seedOrgAndPrompt(t, q, "batch-"+tt.testName)

				body := tt.body(orgID, promptID)
				req := httptest.NewRequest(http.MethodPost, "/logs/batch", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).PostBatch()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result []map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.count, len(result)); diff != "" {
					t.Errorf("result count mismatch (-want +got):\n%s", diff)
				}

				// Verify each result has a non-empty ID
				for i, r := range result {
					if r["id"] == nil || r["id"] == "" {
						t.Errorf("expected non-empty id for result[%d]", i)
					}
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
			// 空文字 / Null/Nil
			{
				testName:       "empty logs array",
				body:           `{"logs": []}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "missing logs field",
				body:           `{}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 異常系 - invalid entry in batch
			{
				testName: "one invalid entry in batch - missing model",
				body: fmt.Sprintf(`{
					"logs": [
						{
							"org_id": "00000000-0000-0000-0000-000000000001",
							"prompt_id": "00000000-0000-0000-0000-000000000002",
							"version_number": 1,
							"request_body": {"prompt": "test"},
							"model": "",
							"provider": "openai",
							"estimated_cost": "0.001000",
							"status": "success",
							"environment": "production",
							"executed_at": %q
						}
					]
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
			// 境界値 - version_number below minimum
			{
				testName: "version_number zero in batch entry",
				body: fmt.Sprintf(`{
					"logs": [
						{
							"org_id": "00000000-0000-0000-0000-000000000001",
							"prompt_id": "00000000-0000-0000-0000-000000000002",
							"version_number": 0,
							"request_body": {"prompt": "test"},
							"model": "gpt-4",
							"provider": "openai",
							"estimated_cost": "0.001000",
							"status": "success",
							"environment": "production",
							"executed_at": %q
						}
					]
				}`, time.Now().UTC().Format(time.RFC3339)),
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodPost, "/logs/batch", bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).PostBatch()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}
