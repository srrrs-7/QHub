package logs

import (
	"api/src/infra/rds/executionlog_repository"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// getTestSeedCounter provides unique slugs for each test seed call.
var getTestSeedCounter int

// seedExecutionLog creates a full FK chain and an execution log, returning the log ID, org ID, and prompt ID.
func seedExecutionLog(t *testing.T, q db.Querier, label string) (logID, orgID, promptID string) {
	t.Helper()
	ctx := context.Background()
	getTestSeedCounter++
	slug := fmt.Sprintf("g%d", getTestSeedCounter)

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

	log, err := q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
		OrganizationID: org.ID,
		PromptID:       prompt.ID,
		VersionNumber:  1,
		RequestBody:    json.RawMessage(`{"prompt":"hello"}`),
		ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"text":"world"}`), Valid: true},
		Model:          "claude-sonnet-4-20250514",
		Provider:       "anthropic",
		InputTokens:    100,
		OutputTokens:   50,
		TotalTokens:    150,
		LatencyMs:      1200,
		EstimatedCost:  "0.003000",
		Status:         "success",
		ErrorMessage:   sql.NullString{},
		Environment:    "production",
		Metadata:       pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"key":"value"}`), Valid: true},
		ExecutedAt:     time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to create execution log: %v", err)
	}

	return log.ID.String(), org.ID.String(), prompt.ID.String()
}

func TestGetHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			model      string
			provider   string
			status     string
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) string
			expected expected
		}{
			// 正常系 (Happy Path)
			{
				testName: "get existing execution log",
				setup: func(t *testing.T, q db.Querier) string {
					logID, _, _ := seedExecutionLog(t, q, "get-existing")
					return logID
				},
				expected: expected{
					statusCode: http.StatusOK,
					model:      "claude-sonnet-4-20250514",
					provider:   "anthropic",
					status:     "success",
				},
			},
			// 特殊文字 (Special Characters) - seeded with unicode
			{
				testName: "get log with unicode content",
				setup: func(t *testing.T, q db.Querier) string {
					ctx := context.Background()
					getTestSeedCounter++
					slug := fmt.Sprintf("gu%d", getTestSeedCounter)
					org, _ := q.CreateOrganization(ctx, db.CreateOrganizationParams{
						Name: "Unicode Org", Slug: "org-" + slug, Plan: "free",
					})
					proj, _ := q.CreateProject(ctx, db.CreateProjectParams{
						OrganizationID: org.ID, Name: "Unicode Proj", Slug: "proj-" + slug,
						Description: sql.NullString{String: "desc", Valid: true},
					})
					prompt, _ := q.CreatePrompt(ctx, db.CreatePromptParams{
						ProjectID: proj.ID, Name: "Unicode Prompt", Slug: "prompt-" + slug, PromptType: "system",
					})
					log, err := q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
						OrganizationID: org.ID,
						PromptID:       prompt.ID,
						VersionNumber:  1,
						RequestBody:    json.RawMessage(`{"prompt":"日本語テスト 🤖"}`),
						ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"text":"応答 ✅"}`), Valid: true},
						Model:          "claude-日本語",
						Provider:       "anthropic",
						InputTokens:    10,
						OutputTokens:   5,
						TotalTokens:    15,
						LatencyMs:      100,
						EstimatedCost:  "0.000100",
						Status:         "success",
						Environment:    "development",
						ExecutedAt:     time.Now().UTC(),
					})
					if err != nil {
						t.Fatalf("failed to seed unicode log: %v", err)
					}
					return log.ID.String()
				},
				expected: expected{
					statusCode: http.StatusOK,
					model:      "claude-日本語",
					provider:   "anthropic",
					status:     "success",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				logID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/logs/"+logID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", logID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(logID, result["id"]); diff != "" {
					t.Errorf("id mismatch (-want +got):\n%s", diff)
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
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName       string
			logID          string
			expectedStatus int
		}{
			// 異常系 (Error Cases)
			{
				testName:       "non-existent log ID",
				logID:          "00000000-0000-0000-0000-000000000000",
				expectedStatus: http.StatusNotFound,
			},
			{
				testName:       "random UUID that does not exist",
				logID:          uuid.New().String(),
				expectedStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/logs/"+tt.logID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.logID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			logID          string
			expectedStatus int
		}{
			// 異常系 (Error Cases) - invalid UUID format
			{
				testName:       "invalid UUID format",
				logID:          "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字 (Empty String)
			{
				testName:       "empty string ID",
				logID:          "",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/logs/"+tt.logID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.logID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}
