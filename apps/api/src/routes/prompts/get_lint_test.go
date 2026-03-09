package prompts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/db/db"
	"utils/testutil"

	"api/src/domain/prompt"
	"api/src/infra/rds/prompt_repository"
	"api/src/services/lintservice"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestGetLintHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			hasScore   bool
			hasIssues  bool
		}

		tests := []struct {
			testName string
			content  string
			expected expected
		}{
			// 正常系 - clean prompt with output format
			{
				testName: "lint clean prompt with json format",
				content:  `{"system":"You are an assistant. Output your response in JSON format."}`,
				expected: expected{statusCode: http.StatusOK, hasScore: true, hasIssues: true},
			},
			// 正常系 - prompt with vague words
			{
				testName: "lint prompt with vague instructions",
				content:  `{"system":"Write a good and appropriate response."}`,
				expected: expected{statusCode: http.StatusOK, hasScore: true, hasIssues: true},
			},
			// 特殊文字 - Japanese content
			{
				testName: "lint Japanese content",
				content:  `{"system":"あなたは親切なアシスタントです。JSON形式で回答してください。"}`,
				expected: expected{statusCode: http.StatusOK, hasScore: true, hasIssues: true},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createSingleVersion(t, q, promptID, authorID, tt.content)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				lintSvc := lintservice.NewLintService(vRepo)
				handler := NewPromptHandler(pRepo, vRepo, nil, lintSvc).GetLint()

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/1/lint", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", "1")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if _, ok := body["score"]; !ok && tt.expected.hasScore {
					t.Error("expected score field in lint response")
				}
				if _, ok := body["issues"]; !ok && tt.expected.hasIssues {
					t.Error("expected issues field in lint response")
				}
			})
		}
	})

	t.Run("200 OK lint score reflects content quality", func(t *testing.T) {
		q := setupTestHandler(t)
		projectID := createTestProject(t, q)
		promptID := createTestPromptRecord(t, q, projectID)
		authorID := createTestUser(t, q)

		// Good prompt: has output format, no vague words, no undeclared vars
		createSingleVersion(t, q, promptID, authorID,
			`{"system":"You are a technical assistant. Respond in JSON format with detailed analysis."}`)

		vRepo := prompt_repository.NewVersionRepository(q)
		pRepo := prompt_repository.NewPromptRepository(q)
		lintSvc := lintservice.NewLintService(vRepo)
		handler := NewPromptHandler(pRepo, vRepo, nil, lintSvc).GetLint()

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/1/lint", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID)
		rctx.URLParams.Add("version", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		score, ok := body["score"].(float64)
		if !ok {
			t.Fatal("expected score to be a number")
		}
		if score < 80 {
			t.Errorf("expected high lint score for clean prompt, got %.0f", score)
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
				createSingleVersion(t, q, promptID, authorID, `{"system":"test"}`)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				lintSvc := lintservice.NewLintService(vRepo)
				handler := NewPromptHandler(pRepo, vRepo, nil, lintSvc).GetLint()

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions/"+tt.version+"/lint", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", tt.version)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
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
			testName string
			promptID string
			version  string
		}{
			// 異常系
			{testName: "invalid prompt_id", promptID: "not-a-uuid", version: "1"},
			// 特殊文字
			{testName: "non-numeric version", promptID: uuid.New().String(), version: "abc"},
			// 空文字
			{testName: "empty prompt_id", promptID: "", version: "1"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				lintSvc := lintservice.NewLintService(vRepo)
				handler := NewPromptHandler(pRepo, vRepo, nil, lintSvc).GetLint()

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptID+"/versions/"+tt.version+"/lint", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				rctx.URLParams.Add("version", tt.version)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// createSingleVersion creates one prompt version with the given content.
func createSingleVersion(t *testing.T, q db.Querier, promptID, authorID, content string) {
	t.Helper()

	parsedPromptID, err := uuid.Parse(promptID)
	if err != nil {
		t.Fatalf("failed to parse prompt ID: %v", err)
	}

	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		t.Fatalf("failed to parse author ID: %v", err)
	}

	pRepo := prompt_repository.NewPromptRepository(q)
	vRepo := prompt_repository.NewVersionRepository(q)
	pid := prompt.PromptIDFromUUID(parsedPromptID)

	_, err = vRepo.Create(t.Context(), prompt.VersionCmd{
		PromptID:          pid,
		Content:           json.RawMessage(content),
		Variables:         nil,
		ChangeDescription: prompt.ChangeDescription("Initial version"),
		AuthorID:          parsedAuthorID,
	}, 1)
	if err != nil {
		t.Fatalf("failed to create version: %v", err)
	}

	_, err = pRepo.UpdateLatestVersion(t.Context(), pid, 1)
	if err != nil {
		t.Fatalf("failed to update latest version: %v", err)
	}
}
