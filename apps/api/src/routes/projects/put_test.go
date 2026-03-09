package projects

import (
	"bytes"
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

func TestPutHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode  int
			name        string
			slug        string
			description string
		}

		tests := []struct {
			testName string
			reqBody  map[string]string
			expected expected
		}{
			// 正常系 - update all fields
			{
				testName: "update all fields",
				reqBody:  map[string]string{"name": "Updated Project", "slug": "updated-project", "description": "Updated description"},
				expected: expected{statusCode: http.StatusOK, name: "Updated Project", slug: "updated-project", description: "Updated description"},
			},
			// 正常系 - update name only (partial update)
			{
				testName: "update name only",
				reqBody:  map[string]string{"name": "New Name Only"},
				expected: expected{statusCode: http.StatusOK, name: "New Name Only", slug: "original-project", description: "Original description"},
			},
			// 正常系 - update slug only
			{
				testName: "update slug only",
				reqBody:  map[string]string{"slug": "new-slug"},
				expected: expected{statusCode: http.StatusOK, name: "Original Project", slug: "new-slug", description: "Original description"},
			},
			// 正常系 - update description only
			{
				testName: "update description only",
				reqBody:  map[string]string{"description": "New description"},
				expected: expected{statusCode: http.StatusOK, name: "Original Project", slug: "original-project", description: "New description"},
			},
			// 特殊文字
			{
				testName: "update name with Japanese",
				reqBody:  map[string]string{"name": "更新されたプロジェクト"},
				expected: expected{statusCode: http.StatusOK, name: "更新されたプロジェクト", slug: "original-project", description: "Original description"},
			},
			// 特殊文字 (emoji)
			{
				testName: "update name with emoji",
				reqBody:  map[string]string{"name": "Updated 🎉"},
				expected: expected{statusCode: http.StatusOK, name: "Updated 🎉", slug: "original-project", description: "Original description"},
			},
			// 境界値
			{
				testName: "update name to min length",
				reqBody:  map[string]string{"name": "AB"},
				expected: expected{statusCode: http.StatusOK, name: "AB", slug: "original-project", description: "Original description"},
			},
			// Null/Nil - empty body keeps existing values
			{
				testName: "empty body keeps existing values",
				reqBody:  map[string]string{},
				expected: expected{statusCode: http.StatusOK, name: "Original Project", slug: "original-project", description: "Original description"},
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
					project.ProjectName("Original Project"),
					project.ProjectSlug("original-project"),
					project.ProjectDescription("Original description"),
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test project: %v", err)
				}

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID+"/projects/original-project", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", "original-project")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Put()
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

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			reqBody  map[string]string
		}{
			// 異常系 - invalid field values
			{testName: "name too short", reqBody: map[string]string{"name": "A"}},
			{testName: "slug too short", reqBody: map[string]string{"slug": "a"}},
			// 特殊文字 - invalid slug characters
			{testName: "slug with uppercase", reqBody: map[string]string{"slug": "Invalid-Slug"}},
			{testName: "slug with spaces", reqBody: map[string]string{"slug": "bad slug"}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				orgUUID, _ := uuid.Parse(orgID)

				repo := project_repository.NewProjectRepository(q)
				cmd := project.NewProjectCmd(
					orgUUID,
					project.ProjectName("Test Project"),
					project.ProjectSlug("test-project"),
					project.ProjectDescription("Test"),
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test project: %v", err)
				}

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID+"/projects/test-project", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", "test-project")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
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
			{testName: "non-existent project", projectSlug: "non-existent"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				repo := project_repository.NewProjectRepository(q)

				body := map[string]string{"name": "Updated"}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID+"/projects/"+tt.projectSlug, bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNotFound, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		q := setupTestHandler(t)
		orgID := createTestOrg(t, q)
		orgUUID, _ := uuid.Parse(orgID)

		repo := project_repository.NewProjectRepository(q)
		cmd := project.NewProjectCmd(
			orgUUID,
			project.ProjectName("Test Project"),
			project.ProjectSlug("test-project"),
			project.ProjectDescription("Test"),
		)
		_, err := repo.Create(t.Context(), cmd)
		if err != nil {
			t.Fatalf("failed to create test project: %v", err)
		}

		req := httptest.NewRequest(http.MethodPut, "/organizations/"+orgID+"/projects/test-project", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", orgID)
		rctx.URLParams.Add("project_slug", "test-project")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()
		handler := NewProjectHandler(repo).Put()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("400 Invalid Slug Param", func(t *testing.T) {
		tests := []struct {
			testName    string
			orgID       string
			projectSlug string
		}{
			// 空文字
			{testName: "empty org_id", orgID: "", projectSlug: "my-project"},
			// 異常系
			{testName: "invalid org_id", orgID: "not-uuid", projectSlug: "my-project"},
			// 空文字
			{testName: "empty project slug", orgID: uuid.New().String(), projectSlug: ""},
			// 境界値
			{testName: "slug param too short", orgID: uuid.New().String(), projectSlug: "a"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := project_repository.NewProjectRepository(q)

				body := map[string]string{"name": "Updated"}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/test/projects/test", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", tt.orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
