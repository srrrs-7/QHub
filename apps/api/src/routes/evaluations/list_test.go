package evaluations

import (
	"api/src/infra/rds/executionlog_repository"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestListHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			count      int
		}

		tests := []struct {
			testName  string
			seedCount int
			expected  expected
		}{
			{
				testName:  "正常系: list evaluations for log with one evaluation",
				seedCount: 1,
				expected: expected{
					statusCode: http.StatusOK,
					count:      1,
				},
			},
			{
				testName:  "正常系: list evaluations for log with multiple evaluations",
				seedCount: 3,
				expected: expected{
					statusCode: http.StatusOK,
					count:      3,
				},
			},
			{
				testName:  "境界値: list evaluations for log with zero evaluations",
				seedCount: 0,
				expected: expected{
					statusCode: http.StatusOK,
					count:      0,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				logID := seedExecutionLog(t, q)

				// Seed evaluations
				for i := 0; i < tt.seedCount; i++ {
					_, err := q.CreateEvaluation(context.Background(), db.CreateEvaluationParams{
						ExecutionLogID: uuid.MustParse(logID),
						OverallScore:   sql.NullString{String: "4.00", Valid: true},
						EvaluatorType:  "human",
						EvaluatorID:    sql.NullString{String: "user-" + uuid.New().String()[:8], Valid: true},
					})
					if err != nil {
						t.Fatalf("failed to seed evaluation %d: %v", i, err)
					}
				}

				req := httptest.NewRequest(http.MethodGet, "/logs/"+logID+"/evaluations", nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("log_id", logID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).List()
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
					t.Errorf("count mismatch (-want +got):\n%s", diff)
				}

				// Verify each item has required fields
				for _, item := range result {
					if _, ok := item["id"]; !ok {
						t.Error("each evaluation should have 'id' field")
					}
					if _, ok := item["execution_log_id"]; !ok {
						t.Error("each evaluation should have 'execution_log_id' field")
					}
					if diff := cmp.Diff(logID, item["execution_log_id"]); diff != "" {
						t.Errorf("execution_log_id mismatch (-want +got):\n%s", diff)
					}
				}
			})
		}
	})

	t.Run("正常系: list returns evaluations only for the specified log", func(t *testing.T) {
		q := testutil.SetupTestTx(t)

		// Create two execution logs
		logID1 := seedExecutionLog(t, q)
		logID2 := seedExecutionLog(t, q)

		// Seed 2 evaluations for log1
		for i := 0; i < 2; i++ {
			_, err := q.CreateEvaluation(context.Background(), db.CreateEvaluationParams{
				ExecutionLogID: uuid.MustParse(logID1),
				EvaluatorType:  "human",
			})
			if err != nil {
				t.Fatalf("failed to seed evaluation for log1: %v", err)
			}
		}

		// Seed 1 evaluation for log2
		_, err := q.CreateEvaluation(context.Background(), db.CreateEvaluationParams{
			ExecutionLogID: uuid.MustParse(logID2),
			EvaluatorType:  "auto",
		})
		if err != nil {
			t.Fatalf("failed to seed evaluation for log2: %v", err)
		}

		// List evaluations for log1 - should return 2
		req := httptest.NewRequest(http.MethodGet, "/logs/"+logID1+"/evaluations", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("log_id", logID1)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		NewEvaluationHandler(evalRepo).List().ServeHTTP(w, req)

		var result []map[string]any
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(2, len(result)); diff != "" {
			t.Errorf("count mismatch for log1 (-want +got):\n%s", diff)
		}

		// List evaluations for log2 - should return 1
		req2 := httptest.NewRequest(http.MethodGet, "/logs/"+logID2+"/evaluations", nil)
		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("log_id", logID2)
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))
		testutil.SetAuthHeader(req2)
		w2 := httptest.NewRecorder()

		NewEvaluationHandler(evalRepo).List().ServeHTTP(w2, req2)

		var result2 []map[string]any
		if err := json.NewDecoder(w2.Result().Body).Decode(&result2); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(1, len(result2)); diff != "" {
			t.Errorf("count mismatch for log2 (-want +got):\n%s", diff)
		}
	})

	t.Run("正常系: list for non-existent log returns empty array", func(t *testing.T) {
		q := testutil.SetupTestTx(t)

		nonExistentLogID := uuid.New().String()

		req := httptest.NewRequest(http.MethodGet, "/logs/"+nonExistentLogID+"/evaluations", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("log_id", nonExistentLogID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		NewEvaluationHandler(evalRepo).List().ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var result []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result == nil {
			// Accept nil or empty slice, but ensure it decodes as empty
			result = []map[string]any{}
		}
		if len(result) != 0 {
			t.Errorf("expected empty array, got %d items", len(result))
		}
	})

	t.Run("400 Bad Request - Invalid log_id", func(t *testing.T) {
		tests := []struct {
			testName       string
			logID          string
			expectedStatus int
		}{
			{
				testName:       "異常系: invalid UUID format for log_id",
				logID:          "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "空文字: empty log_id",
				logID:          "",
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "特殊文字: unicode characters in log_id",
				logID:          "テスト-ログ-ID",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/logs/test/evaluations", nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("log_id", tt.logID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("特殊文字: evaluations with unicode data are correctly returned", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		// Seed evaluation with unicode feedback
		_, err := q.CreateEvaluation(context.Background(), db.CreateEvaluationParams{
			ExecutionLogID: uuid.MustParse(logID),
			EvaluatorType:  "human",
			Feedback:       sql.NullString{String: "とても良い結果です 🌟", Valid: true},
			EvaluatorID:    sql.NullString{String: "評価者-太郎", Valid: true},
		})
		if err != nil {
			t.Fatalf("failed to seed evaluation: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/logs/"+logID+"/evaluations", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("log_id", logID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		NewEvaluationHandler(evalRepo).List().ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var result []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 evaluation, got %d", len(result))
		}

		if diff := cmp.Diff("とても良い結果です 🌟", result[0]["feedback"]); diff != "" {
			t.Errorf("feedback mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff("評価者-太郎", result[0]["evaluator_id"]); diff != "" {
			t.Errorf("evaluator_id mismatch (-want +got):\n%s", diff)
		}
	})
}
