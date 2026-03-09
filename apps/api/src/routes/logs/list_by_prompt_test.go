package logs

import (
	"api/src/infra/rds/executionlog_repository"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// seedLogForPrompt creates a single execution log for given org and prompt.
func seedLogForPrompt(t *testing.T, q db.Querier, orgID, promptID string) string {
	t.Helper()
	ctx := context.Background()

	log, err := q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
		OrganizationID: uuid.MustParse(orgID),
		PromptID:       uuid.MustParse(promptID),
		VersionNumber:  1,
		RequestBody:    json.RawMessage(`{"prompt":"hello"}`),
		ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"text":"world"}`), Valid: true},
		Model:          "claude-sonnet-4-20250514",
		Provider:       "anthropic",
		InputTokens:    10,
		OutputTokens:   20,
		TotalTokens:    30,
		LatencyMs:      150,
		EstimatedCost:  "0.001",
		Status:         "success",
		Environment:    "development",
		ExecutedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create execution log: %v", err)
	}

	return log.ID.String()
}

func TestListByPromptHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			count      int
		}

		tests := []struct {
			testName  string
			seedCount int
			query     string
			expected  expected
		}{
			// 正常系 (Happy Path)
			{
				testName:  "list logs for prompt with multiple logs",
				seedCount: 3,
				query:     "",
				expected: expected{
					statusCode: http.StatusOK,
					count:      3,
				},
			},
			{
				testName:  "list logs for prompt with single log",
				seedCount: 1,
				query:     "",
				expected: expected{
					statusCode: http.StatusOK,
					count:      1,
				},
			},
			// 境界値 (Boundary Values)
			{
				testName:  "list logs with limit=1",
				seedCount: 3,
				query:     "?limit=1",
				expected: expected{
					statusCode: http.StatusOK,
					count:      1,
				},
			},
			{
				testName:  "list logs with offset=1",
				seedCount: 3,
				query:     "?offset=1",
				expected: expected{
					statusCode: http.StatusOK,
					count:      2,
				},
			},
			{
				testName:  "list logs with limit=100 (max)",
				seedCount: 2,
				query:     "?limit=100",
				expected: expected{
					statusCode: http.StatusOK,
					count:      2,
				},
			},
			// Null/Nil - no logs exist for prompt
			{
				testName:  "empty result for prompt with no logs",
				seedCount: 0,
				query:     "",
				expected: expected{
					statusCode: http.StatusOK,
					count:      0,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID, promptID := seedOrgAndPrompt(t, q, tt.testName)

				for i := 0; i < tt.seedCount; i++ {
					seedLogForPrompt(t, q, orgID, promptID)
				}

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/logs"+tt.query, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).ListByPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := result["data"].([]any)
				if !ok {
					t.Fatal("expected 'data' to be an array")
				}

				if diff := cmp.Diff(tt.expected.count, len(data)); diff != "" {
					t.Errorf("data count mismatch (-want +got):\n%s", diff)
				}

				total, ok := result["total"].(float64)
				if !ok {
					t.Fatal("expected 'total' to be a number")
				}

				if int(total) < tt.expected.count {
					t.Errorf("total should be >= data count, got total=%d count=%d", int(total), tt.expected.count)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			promptID       string
			expectedStatus int
		}{
			// 異常系 (Error Cases)
			{
				testName:       "invalid UUID in path",
				promptID:       "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字 (Empty String)
			{
				testName:       "empty prompt_id",
				promptID:       "",
				expectedStatus: http.StatusBadRequest,
			},
			// 特殊文字 (Special Characters)
			{
				testName:       "SQL injection in prompt_id",
				promptID:       "1-OR-1=1",
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "special characters in prompt_id",
				promptID:       "abc<script>alert</script>",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptID+"/logs", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).ListByPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("境界値: limit and offset edge cases", func(t *testing.T) {
		tests := []struct {
			testName string
			query    string
			wantOK   bool
		}{
			{
				testName: "negative limit defaults to 20",
				query:    "?limit=-1",
				wantOK:   true,
			},
			{
				testName: "zero limit defaults to 20",
				query:    "?limit=0",
				wantOK:   true,
			},
			{
				testName: "limit over 100 defaults to 20",
				query:    "?limit=101",
				wantOK:   true,
			},
			{
				testName: "negative offset defaults to 0",
				query:    "?offset=-1",
				wantOK:   true,
			},
			{
				testName: "non-numeric limit",
				query:    "?limit=abc",
				wantOK:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				_, promptID := seedOrgAndPrompt(t, q, tt.testName)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/logs"+tt.query, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).ListByPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if tt.wantOK {
					if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
						t.Errorf("status code mismatch (-want +got):\n%s", diff)
					}
				}
			})
		}
	})

	t.Run("Null/Nil: non-existent prompt returns empty list", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		nonExistentPromptID := uuid.New().String()

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+nonExistentPromptID+"/logs", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", nonExistentPromptID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		logRepo := executionlog_repository.NewLogRepository(q)
		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewLogHandler(logRepo, evalRepo).ListByPrompt()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		data := result["data"].([]any)
		if len(data) != 0 {
			t.Errorf("expected empty data array, got %d items", len(data))
		}
	})
}
