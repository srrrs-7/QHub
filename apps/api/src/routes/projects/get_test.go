package projects

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/domain/project"
	"api/src/infra/rds/project_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestGetHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode  int
			name        string
			slug        string
			description string
		}

		tests := []struct {
			testName    string
			name        string
			slug        string
			description string
			expected    expected
		}{
			// 正常系
			{
				testName:    "get project by org and slug",
				name:        "My Project",
				slug:        "my-project",
				description: "A test project",
				expected:    expected{statusCode: http.StatusOK, name: "My Project", slug: "my-project", description: "A test project"},
			},
			// 特殊文字
			{
				testName:    "get project with Japanese name",
				name:        "テストプロジェクト",
				slug:        "test-project",
				description: "テスト説明",
				expected:    expected{statusCode: http.StatusOK, name: "テストプロジェクト", slug: "test-project", description: "テスト説明"},
			},
			// 境界値
			{
				testName:    "get project with min length slug",
				name:        "AB",
				slug:        "ab",
				description: "",
				expected:    expected{statusCode: http.StatusOK, name: "AB", slug: "ab", description: ""},
			},
			// 特殊文字 (emoji)
			{
				testName:    "get project with emoji in name",
				name:        "Project 🚀",
				slug:        "project-rocket",
				description: "Launch 🎉",
				expected:    expected{statusCode: http.StatusOK, name: "Project 🚀", slug: "project-rocket", description: "Launch 🎉"},
			},
			// 空文字 (empty description)
			{
				testName:    "get project with empty description",
				name:        "No Description",
				slug:        "no-desc",
				description: "",
				expected:    expected{statusCode: http.StatusOK, name: "No Description", slug: "no-desc", description: ""},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				orgUUID, _ := uuid.Parse(orgID)

				repo := project_repository.NewProjectRepository(q)
				cmd := project.NewProjectCmd(
					orgUUID,
					project.ProjectName(tt.name),
					project.ProjectSlug(tt.slug),
					project.ProjectDescription(tt.description),
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test project: %v", err)
				}

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID+"/projects/"+tt.slug, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body projectResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.name, body.Name); diff != "" {
					t.Errorf("name mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.slug, body.Slug); diff != "" {
					t.Errorf("slug mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.description, body.Description); diff != "" {
					t.Errorf("description mismatch (-want +got):\n%s", diff)
				}
				if body.ID == "" {
					t.Error("expected non-empty ID")
				}
				if body.OrganizationID != orgID {
					t.Errorf("expected organization_id %s, got %s", orgID, body.OrganizationID)
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName    string
			projectSlug string
		}{
			// 異常系
			{testName: "non-existent project slug", projectSlug: "non-existent"},
			// 境界値
			{testName: "min length non-existent slug", projectSlug: "zz"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				repo := project_repository.NewProjectRepository(q)

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID+"/projects/"+tt.projectSlug, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNotFound, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName    string
			orgID       string
			projectSlug string
		}{
			// 異常系 - invalid org_id
			{testName: "invalid org_id format", orgID: "not-a-uuid", projectSlug: "my-project"},
			// 空文字
			{testName: "empty org_id", orgID: "", projectSlug: "my-project"},
			// 異常系 - invalid slug
			{testName: "slug too short", orgID: uuid.New().String(), projectSlug: "a"},
			// 空文字
			{testName: "empty project slug", orgID: uuid.New().String(), projectSlug: ""},
			// 特殊文字
			{testName: "slug with special chars", orgID: uuid.New().String(), projectSlug: "bad@slug"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := project_repository.NewProjectRepository(q)

				req := httptest.NewRequest(http.MethodGet, "/organizations/test/projects/test", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", tt.orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
