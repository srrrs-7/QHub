package prompts

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/domain/prompt"
	"api/src/infra/rds/prompt_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestPutVersionStatusHandler(t *testing.T) {
	t.Run("200 OK valid transitions", func(t *testing.T) {
		type expected struct {
			statusCode int
			status     string
		}

		tests := []struct {
			testName    string
			transitions []string // sequence of status transitions to apply
			expected    expected
		}{
			// 正常系 - draft -> review
			{
				testName:    "draft to review",
				transitions: []string{"review"},
				expected:    expected{statusCode: http.StatusOK, status: "review"},
			},
			// 正常系 - draft -> review -> production
			{
				testName:    "draft to review to production",
				transitions: []string{"review", "production"},
				expected:    expected{statusCode: http.StatusOK, status: "production"},
			},
			// 正常系 - draft -> archived (discard)
			{
				testName:    "draft to archived",
				transitions: []string{"archived"},
				expected:    expected{statusCode: http.StatusOK, status: "archived"},
			},
			// 正常系 - full lifecycle: draft -> review -> production -> archived
			{
				testName:    "full lifecycle to archived",
				transitions: []string{"review", "production", "archived"},
				expected:    expected{statusCode: http.StatusOK, status: "archived"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createVersions(t, q, promptID, authorID, 1)

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).PutVersionStatus()

				var resp *http.Response
				for _, status := range tt.transitions {
					jsonBody, _ := json.Marshal(map[string]any{"status": status})

					req := httptest.NewRequest(http.MethodPut, "/prompts/"+promptID+"/versions/1/status", bytes.NewReader(jsonBody))
					req.Header.Set("Content-Type", "application/json")
					testutil.SetAuthHeader(req)

					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("prompt_id", promptID)
					rctx.URLParams.Add("version", "1")
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)
					resp = w.Result()

					if resp.StatusCode != http.StatusOK {
						t.Fatalf("transition to %s failed with status %d", status, resp.StatusCode)
					}
				}

				var body versionResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.status, body.Status); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 invalid transitions", func(t *testing.T) {
		tests := []struct {
			testName       string
			preTransitions []string // transitions to apply before the invalid one
			invalidStatus  string
		}{
			// 異常系 - invalid transitions
			{testName: "draft to production directly", preTransitions: nil, invalidStatus: "production"},
			{testName: "review to draft", preTransitions: []string{"review"}, invalidStatus: "draft"},
			{testName: "review to archived", preTransitions: []string{"review"}, invalidStatus: "archived"},
			{testName: "production to draft", preTransitions: []string{"review", "production"}, invalidStatus: "draft"},
			{testName: "production to review", preTransitions: []string{"review", "production"}, invalidStatus: "review"},
			{testName: "archived to any", preTransitions: []string{"archived"}, invalidStatus: "draft"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createVersions(t, q, promptID, authorID, 1)

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).PutVersionStatus()

				// Apply pre-transitions
				for _, status := range tt.preTransitions {
					jsonBody, _ := json.Marshal(map[string]any{"status": status})

					req := httptest.NewRequest(http.MethodPut, "/prompts/"+promptID+"/versions/1/status", bytes.NewReader(jsonBody))
					req.Header.Set("Content-Type", "application/json")
					testutil.SetAuthHeader(req)

					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("prompt_id", promptID)
					rctx.URLParams.Add("version", "1")
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

					w := httptest.NewRecorder()
					handler.ServeHTTP(w, req)

					if w.Result().StatusCode != http.StatusOK {
						t.Fatalf("pre-transition to %s failed with status %d", status, w.Result().StatusCode)
					}
				}

				// Now try the invalid transition
				jsonBody, _ := json.Marshal(map[string]any{"status": tt.invalidStatus})

				req := httptest.NewRequest(http.MethodPut, "/prompts/"+promptID+"/versions/1/status", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", "1")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("400 Bad Request invalid input", func(t *testing.T) {
		tests := []struct {
			testName string
			promptID string
			version  string
			reqBody  map[string]any
		}{
			// 異常系
			{testName: "invalid prompt_id", promptID: "not-a-uuid", version: "1", reqBody: map[string]any{"status": "review"}},
			{testName: "invalid status value", promptID: uuid.New().String(), version: "1", reqBody: map[string]any{"status": "invalid"}},
			// 特殊文字
			{testName: "non-numeric version", promptID: uuid.New().String(), version: "abc", reqBody: map[string]any{"status": "review"}},
			// 空文字
			{testName: "empty status", promptID: uuid.New().String(), version: "1", reqBody: map[string]any{"status": ""}},
			// Null/Nil
			{testName: "missing status field", promptID: uuid.New().String(), version: "1", reqBody: map[string]any{}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				jsonBody, _ := json.Marshal(tt.reqBody)

				req := httptest.NewRequest(http.MethodPut, "/prompts/"+tt.promptID+"/versions/"+tt.version+"/status", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				rctx.URLParams.Add("version", tt.version)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).PutVersionStatus()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName string
			version  string
		}{
			// 異常系
			{testName: "non-existent version", version: "999"},
			// 境界値
			{testName: "version zero", version: "0"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createVersions(t, q, promptID, authorID, 1)

				jsonBody, _ := json.Marshal(map[string]any{"status": "review"})

				req := httptest.NewRequest(http.MethodPut, "/prompts/"+promptID+"/versions/"+tt.version+"/status", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", tt.version)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).PutVersionStatus()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusNotFound {
					t.Errorf("expected status 404, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("production replaces previous production version", func(t *testing.T) {
		q := setupTestHandler(t)
		projectID := createTestProject(t, q)
		promptID := createTestPromptRecord(t, q, projectID)
		authorID := createTestUser(t, q)
		createVersions(t, q, promptID, authorID, 2)

		pRepo := prompt_repository.NewPromptRepository(q)
		vRepo := prompt_repository.NewVersionRepository(q)
		handler := NewPromptHandler(pRepo, vRepo, nil, nil).PutVersionStatus()

		// Promote version 1: draft -> review -> production
		for _, status := range []string{"review", "production"} {
			jsonBody, _ := json.Marshal(map[string]any{"status": status})
			req := httptest.NewRequest(http.MethodPut, "/prompts/"+promptID+"/versions/1/status", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			testutil.SetAuthHeader(req)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID)
			rctx.URLParams.Add("version", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Result().StatusCode != http.StatusOK {
				t.Fatalf("failed to transition v1 to %s: %d", status, w.Result().StatusCode)
			}
		}

		// Promote version 2: draft -> review -> production
		for _, status := range []string{"review", "production"} {
			jsonBody, _ := json.Marshal(map[string]any{"status": status})
			req := httptest.NewRequest(http.MethodPut, "/prompts/"+promptID+"/versions/2/status", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			testutil.SetAuthHeader(req)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID)
			rctx.URLParams.Add("version", "2")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Result().StatusCode != http.StatusOK {
				t.Fatalf("failed to transition v2 to %s: %d", status, w.Result().StatusCode)
			}
		}

		// Verify version 1 is now archived
		parsedPromptID, _ := uuid.Parse(promptID)
		pid := prompt.PromptIDFromUUID(parsedPromptID)
		v1, err := vRepo.FindByPromptAndNumber(t.Context(), pid, 1)
		if err != nil {
			t.Fatalf("failed to find version 1: %v", err)
		}
		if diff := cmp.Diff("archived", v1.Status.String()); diff != "" {
			t.Errorf("v1 status mismatch (-want +got):\n%s", diff)
		}

		// Verify version 2 is production
		v2, err := vRepo.FindByPromptAndNumber(t.Context(), pid, 2)
		if err != nil {
			t.Fatalf("failed to find version 2: %v", err)
		}
		if diff := cmp.Diff("production", v2.Status.String()); diff != "" {
			t.Errorf("v2 status mismatch (-want +got):\n%s", diff)
		}
	})
}
