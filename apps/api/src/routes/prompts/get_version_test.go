package prompts

import (
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

func TestGetVersionHandler(t *testing.T) {
	t.Run("200 OK by version number", func(t *testing.T) {
		type expected struct {
			statusCode    int
			versionNumber int
			status        string
		}

		tests := []struct {
			testName     string
			versionParam string
			expected     expected
		}{
			// 正常系
			{
				testName:     "get version 1",
				versionParam: "1",
				expected:     expected{statusCode: http.StatusOK, versionNumber: 1, status: "draft"},
			},
			{
				testName:     "get version 2",
				versionParam: "2",
				expected:     expected{statusCode: http.StatusOK, versionNumber: 2, status: "draft"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createVersions(t, q, promptID, authorID, 2)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/"+tt.versionParam, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", tt.versionParam)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).GetVersion()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body versionResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.versionNumber, body.VersionNumber); diff != "" {
					t.Errorf("version_number mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.status, body.Status); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("200 OK latest keyword", func(t *testing.T) {
		q := setupTestHandler(t)
		projectID := createTestProject(t, q)
		promptID := createTestPromptRecord(t, q, projectID)
		authorID := createTestUser(t, q)
		createVersions(t, q, promptID, authorID, 3)

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/latest", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID)
		rctx.URLParams.Add("version", "latest")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		pRepo := prompt_repository.NewPromptRepository(q)
		vRepo := prompt_repository.NewVersionRepository(q)
		handler := NewPromptHandler(pRepo, vRepo, nil, nil).GetVersion()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body versionResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(3, body.VersionNumber); diff != "" {
			t.Errorf("version_number mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("200 OK production keyword", func(t *testing.T) {
		q := setupTestHandler(t)
		projectID := createTestProject(t, q)
		promptID := createTestPromptRecord(t, q, projectID)
		authorID := createTestUser(t, q)
		createVersions(t, q, promptID, authorID, 2)

		// Promote version 1 to production: draft -> review -> production
		parsedPromptID, _ := uuid.Parse(promptID)
		pid := prompt.PromptIDFromUUID(parsedPromptID)
		vRepo := prompt_repository.NewVersionRepository(q)
		pRepo := prompt_repository.NewPromptRepository(q)

		v1, err := vRepo.FindByPromptAndNumber(t.Context(), pid, 1)
		if err != nil {
			t.Fatalf("failed to find version 1: %v", err)
		}
		_, err = vRepo.UpdateStatus(t.Context(), v1.ID, prompt.StatusReview)
		if err != nil {
			t.Fatalf("failed to update to review: %v", err)
		}
		_, err = vRepo.UpdateStatus(t.Context(), v1.ID, prompt.StatusProduction)
		if err != nil {
			t.Fatalf("failed to update to production: %v", err)
		}
		v := 1
		_, err = pRepo.UpdateProductionVersion(t.Context(), pid, &v)
		if err != nil {
			t.Fatalf("failed to update production version: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/production", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID)
		rctx.URLParams.Add("version", "production")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler := NewPromptHandler(pRepo, vRepo, nil, nil).GetVersion()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body versionResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(1, body.VersionNumber); diff != "" {
			t.Errorf("version_number mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("production", body.Status); diff != "" {
			t.Errorf("status mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName     string
			versionParam string
		}{
			// 異常系
			{testName: "non-existent version number", versionParam: "999"},
			// 境界値
			{testName: "version zero", versionParam: "0"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createVersions(t, q, promptID, authorID, 1)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/"+tt.versionParam, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", tt.versionParam)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).GetVersion()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusNotFound {
					t.Errorf("expected status 404, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName     string
			promptID     string
			versionParam string
		}{
			// 異常系
			{testName: "invalid prompt_id", promptID: "not-a-uuid", versionParam: "1"},
			// 特殊文字
			{testName: "non-numeric version", promptID: uuid.New().String(), versionParam: "abc"},
			// 空文字
			{testName: "empty prompt_id", promptID: "", versionParam: "1"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptID+"/versions/"+tt.versionParam, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				rctx.URLParams.Add("version", tt.versionParam)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).GetVersion()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("404 production when no production version exists", func(t *testing.T) {
		q := setupTestHandler(t)
		projectID := createTestProject(t, q)
		promptID := createTestPromptRecord(t, q, projectID)
		authorID := createTestUser(t, q)
		createVersions(t, q, promptID, authorID, 1) // only draft

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/production", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID)
		rctx.URLParams.Add("version", "production")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		pRepo := prompt_repository.NewPromptRepository(q)
		vRepo := prompt_repository.NewVersionRepository(q)
		handler := NewPromptHandler(pRepo, vRepo, nil, nil).GetVersion()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
}
