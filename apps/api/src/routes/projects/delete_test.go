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

func TestDeleteHandler(t *testing.T) {
	t.Run("204 No Content", func(t *testing.T) {
		tests := []struct {
			testName    string
			projectName string
			projectSlug string
		}{
			// 正常系
			{testName: "delete existing project", projectName: "To Delete", projectSlug: "to-delete"},
			// 特殊文字
			{testName: "delete project with Japanese name", projectName: "削除プロジェクト", projectSlug: "delete-jp"},
			// 境界値
			{testName: "delete project with min length slug", projectName: "AB", projectSlug: "ab"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				orgUUID, _ := uuid.Parse(orgID)

				repo := project_repository.NewProjectRepository(q)
				cmd := project.NewProjectCmd(
					orgUUID,
					project.ProjectName(tt.projectName),
					project.ProjectSlug(tt.projectSlug),
					project.ProjectDescription(""),
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test project: %v", err)
				}

				req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID+"/projects/"+tt.projectSlug, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				// Verify project is actually deleted by trying to get it
				getReq := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID+"/projects/"+tt.projectSlug, nil)
				getRctx := chi.NewRouteContext()
				getRctx.URLParams.Add("org_id", orgID)
				getRctx.URLParams.Add("project_slug", tt.projectSlug)
				getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, getRctx))
				testutil.SetAuthHeader(getReq)

				getW := httptest.NewRecorder()
				getHandler := NewProjectHandler(repo).Get()
				getHandler.ServeHTTP(getW, getReq)

				getResp := getW.Result()
				if getResp.StatusCode != http.StatusNotFound {
					t.Errorf("expected deleted project to return 404, got %d", getResp.StatusCode)
				}
			})
		}
	})

	t.Run("204 then list excludes deleted", func(t *testing.T) {
		// 正常系 - verify deletion removes from list
		q := setupTestHandler(t)
		orgID := createTestOrg(t, q)
		orgUUID, _ := uuid.Parse(orgID)

		repo := project_repository.NewProjectRepository(q)

		// Create two projects
		cmd1 := project.NewProjectCmd(orgUUID, project.ProjectName("Keep Me"), project.ProjectSlug("keep-me"), project.ProjectDescription(""))
		_, err := repo.Create(t.Context(), cmd1)
		if err != nil {
			t.Fatalf("failed to create project 1: %v", err)
		}

		cmd2 := project.NewProjectCmd(orgUUID, project.ProjectName("Delete Me"), project.ProjectSlug("delete-me"), project.ProjectDescription(""))
		_, err = repo.Create(t.Context(), cmd2)
		if err != nil {
			t.Fatalf("failed to create project 2: %v", err)
		}

		// Delete second project
		delReq := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID+"/projects/delete-me", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", orgID)
		rctx.URLParams.Add("project_slug", "delete-me")
		delReq = delReq.WithContext(context.WithValue(delReq.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(delReq)

		delW := httptest.NewRecorder()
		NewProjectHandler(repo).Delete().ServeHTTP(delW, delReq)

		if diff := cmp.Diff(http.StatusNoContent, delW.Result().StatusCode); diff != "" {
			t.Fatalf("delete status mismatch (-want +got):\n%s", diff)
		}

		// List should return only 1 project
		listReq := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID+"/projects", nil)
		listRctx := chi.NewRouteContext()
		listRctx.URLParams.Add("org_id", orgID)
		listReq = listReq.WithContext(context.WithValue(listReq.Context(), chi.RouteCtxKey, listRctx))
		testutil.SetAuthHeader(listReq)

		listW := httptest.NewRecorder()
		NewProjectHandler(repo).List().ServeHTTP(listW, listReq)

		var body []projectResponse
		if err := json.NewDecoder(listW.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(1, len(body)); diff != "" {
			t.Errorf("project count after delete mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName    string
			projectSlug string
		}{
			// 異常系
			{testName: "non-existent project", projectSlug: "non-existent"},
			// 境界値
			{testName: "min length non-existent slug", projectSlug: "zz"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				repo := project_repository.NewProjectRepository(q)

				req := httptest.NewRequest(http.MethodDelete, "/organizations/"+orgID+"/projects/"+tt.projectSlug, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Delete()
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

				req := httptest.NewRequest(http.MethodDelete, "/organizations/test/projects/test", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", tt.orgID)
				rctx.URLParams.Add("project_slug", tt.projectSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
