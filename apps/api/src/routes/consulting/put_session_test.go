package consulting

import (
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

func TestPutSessionHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			status     string
		}

		tests := []struct {
			testName string
			reqBody  map[string]any
			expected expected
		}{
			// 正常系 (Happy Path)
			{
				testName: "close an active session",
				reqBody: map[string]any{
					"status": "closed",
				},
				expected: expected{
					statusCode: http.StatusOK,
					status:     "closed",
				},
			},
			{
				testName: "reactivate a session",
				reqBody: map[string]any{
					"status": "active",
				},
				expected: expected{
					statusCode: http.StatusOK,
					status:     "active",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				orgID := seedOrg(t, q)
				sessionID := seedSession(t, q, orgID, "Test Session")

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/consulting/sessions/"+sessionID.String(), bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", sessionID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				handler.PutSession().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.status, result["status"]); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(sessionID.String(), result["id"]); diff != "" {
					t.Errorf("id mismatch (-want +got):\n%s", diff)
				}

				if result["updated_at"] == nil || result["updated_at"] == "" {
					t.Error("expected updated_at to be non-empty")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			sessionID      string
			reqBody        string
			expectedStatus int
		}{
			// 異常系 (Error Cases)
			{
				testName:       "invalid JSON body",
				sessionID:      uuid.New().String(),
				reqBody:        `{invalid json`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "invalid UUID in path",
				sessionID:      "not-a-uuid",
				reqBody:        `{"status": "closed"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "invalid status value",
				sessionID:      uuid.New().String(),
				reqBody:        `{"status": "unknown"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "missing status field",
				sessionID:      uuid.New().String(),
				reqBody:        `{}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字 (Empty String)
			{
				testName:       "empty status",
				sessionID:      uuid.New().String(),
				reqBody:        `{"status": ""}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "empty body",
				sessionID:      uuid.New().String(),
				reqBody:        ``,
				expectedStatus: http.StatusBadRequest,
			},
			// 特殊文字 (Special Characters)
			{
				testName:       "status with special characters",
				sessionID:      uuid.New().String(),
				reqBody:        `{"status": "closed<script>"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "status with SQL injection",
				sessionID:      uuid.New().String(),
				reqBody:        `{"status": "'; DROP TABLE consulting_sessions; --"}`,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				req := httptest.NewRequest(http.MethodPut, "/consulting/sessions/"+tt.sessionID, strings.NewReader(tt.reqBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				handler.PutSession().ServeHTTP(w, req)

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
				testName:       "non-existent session ID",
				expectedStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)
				nonExistentID := uuid.New().String()

				reqBody := `{"status": "closed"}`
				req := httptest.NewRequest(http.MethodPut, "/consulting/sessions/"+nonExistentID, strings.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", nonExistentID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				handler.PutSession().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("Null/Nil: null status in JSON", func(t *testing.T) {
		_, handler := setupHandler(t)

		reqBody := `{"status": null}`
		req := httptest.NewRequest(http.MethodPut, "/consulting/sessions/"+uuid.New().String(), strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.PutSession().ServeHTTP(w, req)

		resp := w.Result()
		// Null status should fail validation
		if resp.StatusCode == http.StatusOK {
			t.Error("expected error for null status, but got 200")
		}
	})

	t.Run("境界値: status exactly 'active' and 'closed'", func(t *testing.T) {
		validStatuses := []string{"active", "closed"}
		for _, status := range validStatuses {
			t.Run(status, func(t *testing.T) {
				q, handler := setupHandler(t)
				orgID := seedOrg(t, q)
				sessionID := seedSession(t, q, orgID, "Boundary Test "+status)

				reqBody := `{"status": "` + status + `"}`
				req := httptest.NewRequest(http.MethodPut, "/consulting/sessions/"+sessionID.String(), strings.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", sessionID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				handler.PutSession().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch for %q (-want +got):\n%s", status, diff)
				}
			})
		}
	})
}
