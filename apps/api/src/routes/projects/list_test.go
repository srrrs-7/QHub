package projects

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/domain/organization"
	"api/src/domain/project"
	"api/src/infra/rds/organization_repository"
	"api/src/infra/rds/project_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestListHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName      string
			projectCount  int
			projectNames  []string
			projectSlugs  []string
			expectedCount int
		}{
			// 正常系
			{
				testName:      "list multiple projects",
				projectCount:  3,
				projectNames:  []string{"Project One", "Project Two", "Project Three"},
				projectSlugs:  []string{"project-one", "project-two", "project-three"},
				expectedCount: 3,
			},
			// 正常系 - single project
			{
				testName:      "list single project",
				projectCount:  1,
				projectNames:  []string{"Only Project"},
				projectSlugs:  []string{"only-project"},
				expectedCount: 1,
			},
			// 境界値 - no projects
			{
				testName:      "list empty projects",
				projectCount:  0,
				projectNames:  []string{},
				projectSlugs:  []string{},
				expectedCount: 0,
			},
			// 特殊文字
			{
				testName:      "list projects with Japanese names",
				projectCount:  2,
				projectNames:  []string{"テストプロジェクト", "日本語プロジェクト"},
				projectSlugs:  []string{"test-jp", "nihongo-project"},
				expectedCount: 2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)
				orgUUID, _ := uuid.Parse(orgID)

				repo := project_repository.NewProjectRepository(q)
				for i := 0; i < tt.projectCount; i++ {
					cmd := project.NewProjectCmd(
						orgUUID,
						project.ProjectName(tt.projectNames[i]),
						project.ProjectSlug(tt.projectSlugs[i]),
						project.ProjectDescription(""),
					)
					_, err := repo.Create(t.Context(), cmd)
					if err != nil {
						t.Fatalf("failed to create test project %d: %v", i, err)
					}
				}

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID+"/projects", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body []projectResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expectedCount, len(body)); diff != "" {
					t.Errorf("project count mismatch (-want +got):\n%s", diff)
				}

				for _, p := range body {
					if p.ID == "" {
						t.Error("expected non-empty ID")
					}
					if p.OrganizationID != orgID {
						t.Errorf("expected organization_id %s, got %s", orgID, p.OrganizationID)
					}
				}
			})
		}
	})

	t.Run("200 OK isolation between orgs", func(t *testing.T) {
		// 正常系 - projects are scoped to org
		q := setupTestHandler(t)
		orgID1 := createTestOrg(t, q)
		orgUUID1, _ := uuid.Parse(orgID1)

		// Create a second org with a different slug
		orgRepo := organization_repository.NewOrganizationRepository(q)
		org2Cmd := organization.NewOrganizationCmd(
			organization.OrganizationName("Other Org"),
			organization.OrganizationSlug("other-org"),
			organization.PlanFree,
		)
		org2, err := orgRepo.Create(t.Context(), org2Cmd)
		if err != nil {
			t.Fatalf("failed to create second org: %v", err)
		}
		orgID2 := org2.ID.String()
		orgUUID2 := org2.ID.UUID()

		repo := project_repository.NewProjectRepository(q)

		// Create project in org1
		cmd1 := project.NewProjectCmd(orgUUID1, project.ProjectName("Org1 Project"), project.ProjectSlug("org1-project"), project.ProjectDescription(""))
		_, err = repo.Create(t.Context(), cmd1)
		if err != nil {
			t.Fatalf("failed to create project in org1: %v", err)
		}

		// Create project in org2
		cmd2 := project.NewProjectCmd(orgUUID2, project.ProjectName("Org2 Project"), project.ProjectSlug("org2-project"), project.ProjectDescription(""))
		_, err = repo.Create(t.Context(), cmd2)
		if err != nil {
			t.Fatalf("failed to create project in org2: %v", err)
		}

		// List org1 projects - should only get 1
		req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID1+"/projects", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", orgID1)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()
		handler := NewProjectHandler(repo).List()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []projectResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(1, len(body)); diff != "" {
			t.Errorf("project count mismatch (-want +got):\n%s", diff)
		}

		_ = orgID2 // used for org2 creation
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
		}{
			// 異常系
			{testName: "invalid org_id format", orgID: "not-a-uuid"},
			// 空文字
			{testName: "empty org_id", orgID: ""},
			// 特殊文字
			{testName: "org_id with special chars", orgID: "abc@def"},
			// Null/Nil
			{testName: "org_id zero uuid", orgID: "00000000-0000-0000-0000-00000000000g"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := project_repository.NewProjectRepository(q)

				req := httptest.NewRequest(http.MethodGet, "/organizations/test/projects", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", tt.orgID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewProjectHandler(repo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
