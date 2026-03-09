package evaluations

import (
	"api/src/infra/rds/executionlog_repository"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestGetHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode    int
			evaluatorType string
			hasScores     bool
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q interface{ CreateEvaluation(ctx context.Context, arg interface{}) }) (evalID string)
			expected expected
		}{
			{
				testName: "正常系: get existing evaluation with scores",
				expected: expected{
					statusCode:    http.StatusOK,
					evaluatorType: "human",
					hasScores:     true,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				logID := seedExecutionLog(t, q)
				evalID := seedEvaluation(t, q, logID)

				req := httptest.NewRequest(http.MethodGet, "/evaluations/"+evalID, nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", evalID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(evalID, result["id"]); diff != "" {
					t.Errorf("id mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(logID, result["execution_log_id"]); diff != "" {
					t.Errorf("execution_log_id mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.evaluatorType, result["evaluator_type"]); diff != "" {
					t.Errorf("evaluator_type mismatch (-want +got):\n%s", diff)
				}

				if _, ok := result["created_at"]; !ok {
					t.Error("response should contain 'created_at' field")
				}

				// Verify score fields exist
				if result["overall_score"] == nil {
					t.Error("expected overall_score to be present")
				}
				if result["accuracy_score"] == nil {
					t.Error("expected accuracy_score to be present")
				}
			})
		}
	})

	t.Run("正常系: get evaluation with null scores", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		// Create evaluation without scores via the handler
		reqBody := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   "auto",
			"feedback":         "No scores provided",
		}
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		// Create via Post handler
		postReq := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		postReq.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(postReq)
		postW := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		NewEvaluationHandler(evalRepo).Post().ServeHTTP(postW, postReq)

		var postResult map[string]any
		if err := json.NewDecoder(postW.Result().Body).Decode(&postResult); err != nil {
			t.Fatalf("failed to decode post response: %v", err)
		}
		evalID := postResult["id"].(string)

		// Now Get it
		getReq := httptest.NewRequest(http.MethodGet, "/evaluations/"+evalID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", evalID)
		getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(getReq)
		getW := httptest.NewRecorder()

		NewEvaluationHandler(evalRepo).Get().ServeHTTP(getW, getReq)

		resp := getW.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["overall_score"] != nil {
			t.Errorf("expected overall_score to be nil, got %v", result["overall_score"])
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName       string
			evalID         string
			expectedStatus int
		}{
			{
				testName:       "異常系: non-existent evaluation",
				evalID:         "00000000-0000-0000-0000-000000000000",
				expectedStatus: http.StatusNotFound,
			},
			{
				testName:       "異常系: random UUID that does not exist",
				evalID:         uuid.New().String(),
				expectedStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/evaluations/"+tt.evalID, nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.evalID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request - Invalid ID", func(t *testing.T) {
		tests := []struct {
			testName       string
			evalID         string
			expectedStatus int
		}{
			{
				testName:       "異常系: invalid UUID format",
				evalID:         "not-a-valid-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "空文字: empty string as ID",
				evalID:         "",
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "特殊文字: unicode in ID",
				evalID:         "日本語テスト",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/evaluations/test", nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.evalID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}
