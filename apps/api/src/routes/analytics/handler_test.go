package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	db "utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// setupAnalyticsData creates org, project, prompt, version, and execution logs for testing.
func setupAnalyticsData(t *testing.T, q db.Querier) (orgID, projectID, promptID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Test Org",
		Slug: "test-org-" + uuid.New().String()[:8],
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	proj, err := q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: org.ID,
		Name:           "Test Project",
		Slug:           "test-proj-" + uuid.New().String()[:8],
		Description:    sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email: "test-" + uuid.New().String()[:8] + "@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	prompt, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		ProjectID:   proj.ID,
		Name:        "Test Prompt",
		Slug:        "test-prompt-" + uuid.New().String()[:8],
		PromptType:  "system",
		Description: sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}

	_, err = q.CreatePromptVersion(ctx, db.CreatePromptVersionParams{
		PromptID:          prompt.ID,
		VersionNumber:     1,
		Status:            "production",
		Content:           json.RawMessage(`{"text":"hello"}`),
		Variables:         pqtype.NullRawMessage{},
		ChangeDescription: sql.NullString{String: "initial", Valid: true},
		AuthorID:          user.ID,
	})
	if err != nil {
		t.Fatalf("failed to create version: %v", err)
	}

	// Create execution logs
	for i := 0; i < 3; i++ {
		_, err = q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
			OrganizationID: org.ID,
			PromptID:       prompt.ID,
			VersionNumber:  1,
			RequestBody:    json.RawMessage(`{"input":"test"}`),
			ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"output":"ok"}`), Valid: true},
			Model:          "gpt-4",
			Provider:       "openai",
			InputTokens:    100,
			OutputTokens:   50,
			TotalTokens:    150,
			LatencyMs:      200,
			EstimatedCost:  "0.001000",
			Status:         "success",
			ErrorMessage:   sql.NullString{},
			Environment:    "production",
			Metadata:       pqtype.NullRawMessage{},
			ExecutedAt:     time.Now().Add(-time.Duration(i) * 24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("failed to create execution log: %v", err)
		}
	}

	return org.ID, proj.ID, prompt.ID
}

func TestGetProjectAnalytics(t *testing.T) {
	type args struct {
		projectID string
	}
	type expected struct {
		statusCode int
	}

	tests := []struct {
		testName string
		args     args
		setup    func(t *testing.T, q db.Querier) string
		expected expected
	}{
		// 正常系
		{
			testName: "valid project with execution logs",
			setup: func(t *testing.T, q db.Querier) string {
				_, projectID, _ := setupAnalyticsData(t, q)
				return projectID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 正常系 - empty results
		{
			testName: "valid project with no execution logs",
			setup: func(t *testing.T, q db.Querier) string {
				ctx := context.Background()
				org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
					Name: "Empty Org", Slug: "empty-org-" + uuid.New().String()[:8], Plan: "free",
				})
				if err != nil {
					t.Fatalf("failed to create org: %v", err)
				}
				proj, err := q.CreateProject(ctx, db.CreateProjectParams{
					OrganizationID: org.ID, Name: "Empty Project",
					Slug:        "empty-proj-" + uuid.New().String()[:8],
					Description: sql.NullString{},
				})
				if err != nil {
					t.Fatalf("failed to create project: %v", err)
				}
				return proj.ID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 異常系 - invalid UUID
		{
			testName: "invalid project_id UUID",
			args:     args{projectID: "not-a-uuid"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 空文字 - empty project_id
		{
			testName: "empty project_id",
			args:     args{projectID: ""},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 正常系 - non-existent project returns empty list
		{
			testName: "non-existent project UUID returns empty",
			args:     args{projectID: uuid.New().String()},
			expected: expected{statusCode: http.StatusOK},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			projectID := tt.args.projectID
			if tt.setup != nil {
				projectID = tt.setup(t, q)
			}

			handler := NewAnalyticsHandler(q).GetProjectAnalytics()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("project_id", projectID)
			req := httptest.NewRequest(http.MethodGet, "/analytics/projects/"+projectID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// Data integrity check
	t.Run("response contains correct analytics data", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		_, projectID, _ := setupAnalyticsData(t, q)

		handler := NewAnalyticsHandler(q).GetProjectAnalytics()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("project_id", projectID.String())
		req := httptest.NewRequest(http.MethodGet, "/analytics/projects/"+projectID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var result []projectAnalyticsResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 prompt analytics, got %d", len(result))
		}

		if result[0].TotalExecutions != 3 {
			t.Errorf("expected 3 total executions, got %d", result[0].TotalExecutions)
		}
		if result[0].PromptName != "Test Prompt" {
			t.Errorf("expected prompt name 'Test Prompt', got %s", result[0].PromptName)
		}
	})
}

func TestGetPromptAnalytics(t *testing.T) {
	type args struct {
		promptID string
	}
	type expected struct {
		statusCode int
	}

	tests := []struct {
		testName string
		args     args
		setup    func(t *testing.T, q db.Querier) string
		expected expected
	}{
		// 正常系
		{
			testName: "valid prompt with execution logs",
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 異常系 - invalid UUID
		{
			testName: "invalid prompt_id UUID",
			args:     args{promptID: "not-a-uuid"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 空文字
		{
			testName: "empty prompt_id",
			args:     args{promptID: ""},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 正常系 - non-existent prompt
		{
			testName: "non-existent prompt UUID returns empty",
			args:     args{promptID: uuid.New().String()},
			expected: expected{statusCode: http.StatusOK},
		},
		// 特殊文字
		{
			testName: "special characters in prompt_id",
			args:     args{promptID: "<script>alert('xss')</script>"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			promptID := tt.args.promptID
			if tt.setup != nil {
				promptID = tt.setup(t, q)
			}

			handler := NewAnalyticsHandler(q).GetPromptAnalytics()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID)
			req := httptest.NewRequest(http.MethodGet, "/analytics/prompts/"+promptID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// Data integrity check
	t.Run("response contains correct prompt analytics", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		_, _, promptID := setupAnalyticsData(t, q)

		handler := NewAnalyticsHandler(q).GetPromptAnalytics()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())
		req := httptest.NewRequest(http.MethodGet, "/analytics/prompts/"+promptID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var result []promptAnalyticsResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 version analytics, got %d", len(result))
		}

		if result[0].VersionNumber != 1 {
			t.Errorf("expected version 1, got %d", result[0].VersionNumber)
		}
		if result[0].TotalExecutions != 3 {
			t.Errorf("expected 3 executions, got %d", result[0].TotalExecutions)
		}
		if result[0].SuccessCount != 3 {
			t.Errorf("expected 3 successes, got %d", result[0].SuccessCount)
		}
		if result[0].ErrorCount != 0 {
			t.Errorf("expected 0 errors, got %d", result[0].ErrorCount)
		}
	})
}

func TestGetDailyTrend(t *testing.T) {
	type args struct {
		promptID   string
		startParam string
		endParam   string
	}
	type expected struct {
		statusCode int
	}

	tests := []struct {
		testName string
		args     args
		setup    func(t *testing.T, q db.Querier) string
		expected expected
	}{
		// 正常系 - default date range (last 30 days)
		{
			testName: "valid prompt with default date range",
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 正常系 - custom date range
		{
			testName: "valid prompt with custom start and end",
			args: args{
				startParam: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				endParam:   time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 正常系 - only start param
		{
			testName: "valid prompt with only start param",
			args: args{
				startParam: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 正常系 - only end param
		{
			testName: "valid prompt with only end param",
			args: args{
				endParam: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 異常系 - invalid UUID
		{
			testName: "invalid prompt_id UUID",
			args:     args{promptID: "bad-uuid"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 異常系 - invalid start date format
		{
			testName: "invalid start date format",
			args: args{
				startParam: "2024/01/01",
			},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - invalid end date format
		{
			testName: "invalid end date format",
			args: args{
				endParam: "not-a-date",
			},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 空文字 - empty prompt_id
		{
			testName: "empty prompt_id",
			args:     args{promptID: ""},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 境界値 - future date range returns empty
		{
			testName: "future date range returns empty results",
			args: args{
				startParam: "2099-01-01",
				endParam:   "2099-12-31",
			},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 正常系 - non-existent prompt
		{
			testName: "non-existent prompt UUID returns empty",
			args:     args{promptID: uuid.New().String()},
			expected: expected{statusCode: http.StatusOK},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			promptID := tt.args.promptID
			if tt.setup != nil {
				promptID = tt.setup(t, q)
			}

			handler := NewAnalyticsHandler(q).GetDailyTrend()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID)

			target := "/analytics/prompts/" + promptID + "/trend"
			sep := "?"
			if tt.args.startParam != "" {
				target += sep + "start=" + tt.args.startParam
				sep = "&"
			}
			if tt.args.endParam != "" {
				target += sep + "end=" + tt.args.endParam
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// Data integrity check
	t.Run("response contains daily trend data", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		_, _, promptID := setupAnalyticsData(t, q)

		handler := NewAnalyticsHandler(q).GetDailyTrend()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())

		start := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		end := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		req := httptest.NewRequest(http.MethodGet, "/analytics/prompts/"+promptID.String()+"/trend?start="+start+"&end="+end, nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Fatalf("status code mismatch (-want +got):\n%s", diff)
		}

		var result []dailyTrendResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// We created 3 logs on 3 different days
		if len(result) == 0 {
			t.Error("expected non-empty trend data")
		}

		for _, r := range result {
			if r.Day == "" {
				t.Error("expected non-empty day field")
			}
			if r.TotalExecutions <= 0 {
				t.Error("expected positive total_executions")
			}
		}
	})
}

func TestGetVersionAnalytics(t *testing.T) {
	type args struct {
		promptID string
		version  string
	}
	type expected struct {
		statusCode int
	}

	tests := []struct {
		testName string
		args     args
		setup    func(t *testing.T, q db.Querier) string
		expected expected
	}{
		// 正常系
		{
			testName: "valid prompt and version with execution logs",
			args:     args{version: "1"},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 異常系 - invalid UUID
		{
			testName: "invalid prompt_id UUID",
			args:     args{promptID: "not-a-uuid", version: "1"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 異常系 - invalid version (not a number)
		{
			testName: "invalid version (not a number)",
			args:     args{version: "abc"},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - non-existent version (no rows in result set)
		{
			testName: "non-existent version returns error",
			args:     args{version: "999"},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 空文字 - empty prompt_id
		{
			testName: "empty prompt_id",
			args:     args{promptID: "", version: "1"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 空文字 - empty version
		{
			testName: "empty version string",
			args:     args{version: ""},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 境界値 - version 0
		{
			testName: "version 0 returns error (no data)",
			args:     args{version: "0"},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 特殊文字
		{
			testName: "special characters in version",
			args:     args{version: "<script>"},
			setup: func(t *testing.T, q db.Querier) string {
				_, _, promptID := setupAnalyticsData(t, q)
				return promptID.String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			promptID := tt.args.promptID
			if tt.setup != nil {
				promptID = tt.setup(t, q)
			}

			handler := NewAnalyticsHandler(q).GetVersionAnalytics()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID)
			rctx.URLParams.Add("version", tt.args.version)
			req := httptest.NewRequest(http.MethodGet, "/analytics/prompts/"+promptID+"/versions/"+tt.args.version, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}

	// Data integrity check
	t.Run("response contains correct version analytics", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		_, _, promptID := setupAnalyticsData(t, q)

		handler := NewAnalyticsHandler(q).GetVersionAnalytics()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())
		rctx.URLParams.Add("version", "1")
		req := httptest.NewRequest(http.MethodGet, "/analytics/prompts/"+promptID.String()+"/versions/1", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Fatalf("status code mismatch (-want +got):\n%s", diff)
		}

		var result versionAnalyticsResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.PromptID != promptID.String() {
			t.Errorf("expected prompt_id %s, got %s", promptID.String(), result.PromptID)
		}
		if result.VersionNumber != 1 {
			t.Errorf("expected version 1, got %d", result.VersionNumber)
		}
		if result.TotalExecutions != 3 {
			t.Errorf("expected 3 executions, got %d", result.TotalExecutions)
		}
		if result.SuccessCount != 3 {
			t.Errorf("expected 3 successes, got %d", result.SuccessCount)
		}
		if result.ErrorCount != 0 {
			t.Errorf("expected 0 errors, got %d", result.ErrorCount)
		}
	})
}

func TestNewAnalyticsHandler(t *testing.T) {
	// Nil querier
	t.Run("nil querier", func(t *testing.T) {
		h := NewAnalyticsHandler(nil)
		if h == nil {
			t.Fatal("expected non-nil handler")
		}
		if h.q != nil {
			t.Error("expected nil querier")
		}
	})

	// Valid querier
	t.Run("valid querier", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		h := NewAnalyticsHandler(q)
		if h == nil {
			t.Fatal("expected non-nil handler")
		}
		if h.q == nil {
			t.Error("expected non-nil querier")
		}
	})
}
