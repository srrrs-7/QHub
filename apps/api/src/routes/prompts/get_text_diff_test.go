package prompts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/infra/rds/prompt_repository"
	"api/src/services/diffservice"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestGetTextDiffHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName     string
			version      string
			fromQuery    string // query param ?from=N
			expectedFrom int
			expectedTo   int
		}{
			// 正常系 - explicit from version
			{
				testName:     "text diff with explicit from version",
				version:      "2",
				fromQuery:    "1",
				expectedFrom: 1,
				expectedTo:   2,
			},
			// 正常系 - implicit from version (defaults to version-1)
			{
				testName:     "text diff with implicit from version",
				version:      "2",
				fromQuery:    "",
				expectedFrom: 1,
				expectedTo:   2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createDiffVersions(t, q, promptID, authorID)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				diffSvc := diffservice.NewDiffService(vRepo, nil)
				handler := NewPromptHandler(pRepo, vRepo, diffSvc, nil, nil).GetTextDiff()

				url := "/prompts/" + promptID + "/versions/" + tt.version + "/text-diff"
				if tt.fromQuery != "" {
					url += "?from=" + tt.fromQuery
				}

				req := httptest.NewRequest(http.MethodGet, url, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("version", tt.version)
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

				if fromVersion, ok := body["from_version"].(float64); ok {
					if diff := cmp.Diff(tt.expectedFrom, int(fromVersion)); diff != "" {
						t.Errorf("from_version mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Error("expected from_version field")
				}

				if toVersion, ok := body["to_version"].(float64); ok {
					if diff := cmp.Diff(tt.expectedTo, int(toVersion)); diff != "" {
						t.Errorf("to_version mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Error("expected to_version field")
				}

				if _, ok := body["hunks"]; !ok {
					t.Error("expected hunks field in text diff response")
				}
				if _, ok := body["stats"]; !ok {
					t.Error("expected stats field in text diff response")
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName  string
			version   string
			fromQuery string
		}{
			// 異常系
			{testName: "non-existent to version", version: "999", fromQuery: "1"},
			{testName: "non-existent from version", version: "2", fromQuery: "999"},
			// 境界値
			{testName: "version zero", version: "0", fromQuery: ""},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createDiffVersions(t, q, promptID, authorID)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				diffSvc := diffservice.NewDiffService(vRepo, nil)
				handler := NewPromptHandler(pRepo, vRepo, diffSvc, nil, nil).GetTextDiff()

				url := "/prompts/" + promptID + "/versions/" + tt.version + "/text-diff"
				if tt.fromQuery != "" {
					url += "?from=" + tt.fromQuery
				}

				req := httptest.NewRequest(http.MethodGet, url, nil)
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
			testName  string
			promptID  string
			version   string
			fromQuery string
		}{
			// 異常系
			{testName: "invalid prompt_id", promptID: "not-a-uuid", version: "2", fromQuery: ""},
			// 特殊文字
			{testName: "non-numeric version", promptID: uuid.New().String(), version: "abc", fromQuery: ""},
			{testName: "non-numeric from", promptID: uuid.New().String(), version: "2", fromQuery: "abc"},
			// 空文字
			{testName: "empty prompt_id", promptID: "", version: "2", fromQuery: ""},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				diffSvc := diffservice.NewDiffService(vRepo, nil)
				handler := NewPromptHandler(pRepo, vRepo, diffSvc, nil, nil).GetTextDiff()

				url := "/prompts/" + tt.promptID + "/versions/" + tt.version + "/text-diff"
				if tt.fromQuery != "" {
					url += "?from=" + tt.fromQuery
				}

				req := httptest.NewRequest(http.MethodGet, url, nil)
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
