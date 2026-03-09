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

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// seedMultipleLogs creates n execution logs for a given org and prompt.
func seedMultipleLogs(t *testing.T, q db.Querier, orgID, promptID uuid.UUID, n int) {
	t.Helper()
	ctx := context.Background()

	for i := 0; i < n; i++ {
		_, err := q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
			OrganizationID: orgID,
			PromptID:       promptID,
			VersionNumber:  1,
			RequestBody:    json.RawMessage(fmt.Sprintf(`{"prompt":"request %d"}`, i)),
			ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(fmt.Sprintf(`{"text":"response %d"}`, i)), Valid: true},
			Model:          "claude-sonnet-4-20250514",
			Provider:       "anthropic",
			InputTokens:    int32(100 + i),
			OutputTokens:   int32(50 + i),
			TotalTokens:    int32(150 + i),
			LatencyMs:      int32(1000 + i*100),
			EstimatedCost:  "0.003000",
			Status:         "success",
			Environment:    "production",
			ExecutedAt:     time.Now().UTC().Add(-time.Duration(i) * time.Minute),
		})
		if err != nil {
			t.Fatalf("failed to seed log %d: %v", i, err)
		}
	}
}

// listTestSeedCounter provides unique slugs for each test seed call.
var listTestSeedCounter int

// seedOrgProjectPrompt creates the FK chain and returns the IDs.
func seedOrgProjectPrompt(t *testing.T, q db.Querier, label string) (orgID, promptID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	listTestSeedCounter++
	slug := fmt.Sprintf("l%d", listTestSeedCounter)

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Org " + label, Slug: "org-" + slug, Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	proj, err := q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: org.ID, Name: "Project " + label, Slug: "proj-" + slug,
		Description: sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	prompt, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		ProjectID: proj.ID, Name: "Prompt " + label, Slug: "prompt-" + slug, PromptType: "system",
	})
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}

	return org.ID, prompt.ID
}

func TestListHandler(t *testing.T) {
	t.Run("200 OK - by prompt_id", func(t *testing.T) {
		type expected struct {
			statusCode int
			total      float64
			dataLen    int
		}

		tests := []struct {
			testName  string
			seedCount int
			query     func(promptID uuid.UUID) string
			expected  expected
		}{
			// 正常系 (Happy Path)
			{
				testName:  "list logs by prompt_id with default pagination",
				seedCount: 3,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      3,
					dataLen:    3,
				},
			},
			{
				testName:  "list logs by prompt_id with limit",
				seedCount: 5,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s&limit=2", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      5,
					dataLen:    2,
				},
			},
			{
				testName:  "list logs by prompt_id with offset",
				seedCount: 5,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s&limit=3&offset=3", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      5,
					dataLen:    2,
				},
			},
			// 境界値 (Boundary Values)
			{
				testName:  "zero results for prompt with no logs",
				seedCount: 0,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      0,
					dataLen:    0,
				},
			},
			{
				testName:  "limit exceeds total count returns all",
				seedCount: 3,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s&limit=100", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      3,
					dataLen:    3,
				},
			},
			{
				testName:  "negative limit defaults to 20",
				seedCount: 3,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s&limit=-1", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      3,
					dataLen:    3,
				},
			},
			{
				testName:  "limit zero defaults to 20",
				seedCount: 3,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s&limit=0", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      3,
					dataLen:    3,
				},
			},
			{
				testName:  "negative offset defaults to 0",
				seedCount: 3,
				query: func(promptID uuid.UUID) string {
					return fmt.Sprintf("?prompt_id=%s&offset=-5", promptID.String())
				},
				expected: expected{
					statusCode: http.StatusOK,
					total:      3,
					dataLen:    3,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID, promptID := seedOrgProjectPrompt(t, q, tt.testName)

				if tt.seedCount > 0 {
					seedMultipleLogs(t, q, orgID, promptID, tt.seedCount)
				}

				queryStr := tt.query(promptID)
				req := httptest.NewRequest(http.MethodGet, "/logs"+queryStr, nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.total, result["total"]); diff != "" {
					t.Errorf("total mismatch (-want +got):\n%s", diff)
				}

				data, ok := result["data"].([]any)
				if !ok {
					t.Fatal("expected data to be an array")
				}
				if diff := cmp.Diff(tt.expected.dataLen, len(data)); diff != "" {
					t.Errorf("data length mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("200 OK - by org_id", func(t *testing.T) {
		tests := []struct {
			testName  string
			seedCount int
			expected  struct {
				statusCode int
				dataLen    int
			}
		}{
			// 正常系 (Happy Path)
			{
				testName:  "list logs by org_id",
				seedCount: 2,
				expected: struct {
					statusCode int
					dataLen    int
				}{
					statusCode: http.StatusOK,
					dataLen:    2,
				},
			},
			// Null/Nil - no logs for org
			{
				testName:  "no logs for org",
				seedCount: 0,
				expected: struct {
					statusCode int
					dataLen    int
				}{
					statusCode: http.StatusOK,
					dataLen:    0,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID, promptID := seedOrgProjectPrompt(t, q, "org-"+tt.testName)

				if tt.seedCount > 0 {
					seedMultipleLogs(t, q, orgID, promptID, tt.seedCount)
				}

				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/logs?org_id=%s", orgID.String()), nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).List()
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
					t.Fatal("expected data to be an array")
				}
				if diff := cmp.Diff(tt.expected.dataLen, len(data)); diff != "" {
					t.Errorf("data length mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("200 OK - no filter returns empty", func(t *testing.T) {
		// 空文字 / Null/Nil - no prompt_id or org_id returns empty
		q := testutil.SetupTestTx(t)

		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		logRepo := executionlog_repository.NewLogRepository(q)
		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewLogHandler(logRepo, evalRepo).List()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(float64(0), result["total"]); diff != "" {
			t.Errorf("total mismatch (-want +got):\n%s", diff)
		}

		data, ok := result["data"].([]any)
		if !ok {
			t.Fatal("expected data to be an array")
		}
		if len(data) != 0 {
			t.Errorf("expected empty data array, got %d items", len(data))
		}
	})

	t.Run("Error Cases", func(t *testing.T) {
		tests := []struct {
			testName string
			query    string
		}{
			// 異常系 (Error Cases)
			{
				testName: "invalid prompt_id UUID format",
				query:    "?prompt_id=not-a-uuid",
			},
			{
				testName: "invalid org_id UUID format",
				query:    "?org_id=invalid-uuid",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/logs"+tt.query, nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				logRepo := executionlog_repository.NewLogRepository(q)
				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewLogHandler(logRepo, evalRepo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				// Invalid UUID parsing returns a plain error, mapped to 500 by HandleError
				if resp.StatusCode == http.StatusOK {
					t.Error("expected non-200 status for invalid UUID, got 200")
				}
				if resp.StatusCode != http.StatusInternalServerError {
					t.Errorf("expected 500 for invalid UUID, got %d", resp.StatusCode)
				}
			})
		}
	})
}
