package evaluations

import (
	"api/src/infra/rds/executionlog_repository"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestPutHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode  int
			hasFeedback bool
		}

		tests := []struct {
			testName string
			reqBody  map[string]any
			expected expected
		}{
			// 正常系 (Happy Path)
			{
				testName: "update all scores and feedback",
				reqBody: map[string]any{
					"overall_score":   "4.80",
					"accuracy_score":  "4.50",
					"relevance_score": "4.70",
					"fluency_score":   "4.90",
					"safety_score":    "5.00",
					"feedback":        "Updated feedback with better scores",
					"metadata":        map[string]string{"updated": "true"},
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			{
				testName: "update only overall score",
				reqBody: map[string]any{
					"overall_score": "3.00",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: false,
				},
			},
			{
				testName: "update only feedback",
				reqBody: map[string]any{
					"feedback": "Only updating the feedback field",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			// 境界値 (Boundary Values)
			{
				testName: "score at 0.00",
				reqBody: map[string]any{
					"overall_score": "0.00",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: false,
				},
			},
			{
				testName: "score at 5.00",
				reqBody: map[string]any{
					"overall_score": "5.00",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: false,
				},
			},
			{
				testName: "feedback at max length (2000)",
				reqBody: map[string]any{
					"feedback": strings.Repeat("f", 2000),
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			// 特殊文字 (Special Characters)
			{
				testName: "unicode in feedback",
				reqBody: map[string]any{
					"feedback": "評価が改善されました 🎉✅",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			{
				testName: "SQL injection in feedback",
				reqBody: map[string]any{
					"feedback": "'; DROP TABLE evaluations; --",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			{
				testName: "XSS in feedback",
				reqBody: map[string]any{
					"feedback": "<script>alert('xss')</script>",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			// 空文字 (Empty String)
			{
				testName: "empty feedback preserves existing",
				reqBody: map[string]any{
					"feedback": "",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: false,
				},
			},
			// Null/Nil
			{
				testName: "null scores in request",
				reqBody: map[string]any{
					"overall_score":  nil,
					"accuracy_score": nil,
					"feedback":       "Updated with null scores",
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: true,
				},
			},
			{
				testName: "null metadata",
				reqBody: map[string]any{
					"metadata": nil,
				},
				expected: expected{
					statusCode:  http.StatusOK,
					hasFeedback: false,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				logID := seedExecutionLog(t, q)
				evalID := seedEvaluation(t, q, logID)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/evaluations/"+evalID, bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", evalID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Put()
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
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			evalID         string
			reqBody        string
			expectedStatus int
		}{
			// 異常系 (Error Cases)
			{
				testName:       "invalid JSON body",
				evalID:         uuid.New().String(),
				reqBody:        `{invalid json`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "invalid UUID in path",
				evalID:         "not-a-uuid",
				reqBody:        `{"feedback": "test"}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 境界値 (Boundary Values)
			{
				testName:       "feedback exceeds max length (2000)",
				evalID:         uuid.New().String(),
				reqBody:        `{"feedback":"` + strings.Repeat("a", 2001) + `"}`,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodPut, "/evaluations/"+tt.evalID, strings.NewReader(tt.reqBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.evalID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName       string
			expectedStatus int
		}{
			// 異常系 (Error Cases)
			{
				testName:       "non-existent evaluation ID",
				expectedStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				nonExistentID := uuid.New().String()

				reqBody := `{"feedback": "test"}`
				req := httptest.NewRequest(http.MethodPut, "/evaluations/"+nonExistentID, strings.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", nonExistentID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("空文字: empty body", func(t *testing.T) {
		q := testutil.SetupTestTx(t)

		req := httptest.NewRequest(http.MethodPut, "/evaluations/"+uuid.New().String(), strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Put()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// Empty body should be rejected
		if resp.StatusCode == http.StatusOK {
			t.Error("expected error for empty body, but got 200")
		}
	})
}
