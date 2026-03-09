package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api/src/services/statsservice"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	db "utils/db/db"
	"utils/testutil"
)

// setupTwoVersionData creates org, project, prompt, two versions, and execution
// logs with different metrics for each version so the t-test has data to compare.
func setupTwoVersionData(t *testing.T, q db.Querier) (promptID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Stats Org",
		Slug: "stats-org-" + uuid.New().String()[:8],
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	proj, err := q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: org.ID,
		Name:           "Stats Project",
		Slug:           "stats-proj-" + uuid.New().String()[:8],
		Description:    sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email: "stats-" + uuid.New().String()[:8] + "@example.com",
		Name:  "Stats User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	prompt, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		ProjectID:   proj.ID,
		Name:        "Stats Prompt",
		Slug:        "stats-prompt-" + uuid.New().String()[:8],
		PromptType:  "system",
		Description: sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}

	// Version 1
	_, err = q.CreatePromptVersion(ctx, db.CreatePromptVersionParams{
		PromptID:          prompt.ID,
		VersionNumber:     1,
		Status:            "archived",
		Content:           json.RawMessage(`{"text":"v1"}`),
		Variables:         pqtype.NullRawMessage{},
		ChangeDescription: sql.NullString{String: "version 1", Valid: true},
		AuthorID:          user.ID,
	})
	if err != nil {
		t.Fatalf("failed to create version 1: %v", err)
	}

	// Version 2
	_, err = q.CreatePromptVersion(ctx, db.CreatePromptVersionParams{
		PromptID:          prompt.ID,
		VersionNumber:     2,
		Status:            "production",
		Content:           json.RawMessage(`{"text":"v2"}`),
		Variables:         pqtype.NullRawMessage{},
		ChangeDescription: sql.NullString{String: "version 2", Valid: true},
		AuthorID:          user.ID,
	})
	if err != nil {
		t.Fatalf("failed to create version 2: %v", err)
	}

	// Execution logs for version 1: latency ~200ms, tokens ~150
	for i := 0; i < 10; i++ {
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
			LatencyMs:      int32(195 + i),
			EstimatedCost:  "0.002000",
			Status:         "success",
			ErrorMessage:   sql.NullString{},
			Environment:    "production",
			Metadata:       pqtype.NullRawMessage{},
			ExecutedAt:     time.Now().Add(-time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatalf("failed to create execution log v1: %v", err)
		}
	}

	// Execution logs for version 2: latency ~100ms, tokens ~80
	for i := 0; i < 10; i++ {
		_, err = q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
			OrganizationID: org.ID,
			PromptID:       prompt.ID,
			VersionNumber:  2,
			RequestBody:    json.RawMessage(`{"input":"test"}`),
			ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"output":"ok"}`), Valid: true},
			Model:          "gpt-4",
			Provider:       "openai",
			InputTokens:    50,
			OutputTokens:   30,
			TotalTokens:    80,
			LatencyMs:      int32(95 + i),
			EstimatedCost:  "0.001000",
			Status:         "success",
			ErrorMessage:   sql.NullString{},
			Environment:    "production",
			Metadata:       pqtype.NullRawMessage{},
			ExecutedAt:     time.Now().Add(-time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatalf("failed to create execution log v2: %v", err)
		}
	}

	return prompt.ID
}

func TestCompareVersions(t *testing.T) {
	type args struct {
		promptID string
		v1       string
		v2       string
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
		// 正常系 - valid comparison
		{
			testName: "valid comparison between v1 and v2",
			args:     args{v1: "1", v2: "2"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusOK},
		},
		// 異常系 - invalid UUID
		{
			testName: "invalid prompt_id UUID",
			args:     args{promptID: "not-a-uuid", v1: "1", v2: "2"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 異常系 - invalid v1 (not a number)
		{
			testName: "invalid v1 not a number",
			args:     args{v1: "abc", v2: "2"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - invalid v2 (not a number)
		{
			testName: "invalid v2 not a number",
			args:     args{v1: "1", v2: "xyz"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - same version numbers
		{
			testName: "same version numbers returns validation error",
			args:     args{v1: "1", v2: "1"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 異常系 - negative version number
		{
			testName: "negative version number",
			args:     args{v1: "-1", v2: "2"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 異常系 - zero version number
		{
			testName: "zero version number",
			args:     args{v1: "0", v2: "2"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 異常系 - non-existent version (no execution data)
		{
			testName: "non-existent version returns not found",
			args:     args{v1: "1", v2: "999"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusNotFound},
		},
		// 空文字 - empty prompt_id
		{
			testName: "empty prompt_id",
			args:     args{promptID: "", v1: "1", v2: "2"},
			expected: expected{statusCode: http.StatusBadRequest},
		},
		// 空文字 - empty v1
		{
			testName: "empty v1",
			args:     args{v1: "", v2: "2"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 特殊文字 - special chars in v1
		{
			testName: "special characters in v1",
			args:     args{v1: "<script>", v2: "2"},
			setup: func(t *testing.T, q db.Querier) string {
				return setupTwoVersionData(t, q).String()
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 正常系 - non-existent prompt (no data for either version)
		{
			testName: "non-existent prompt returns not found",
			args:     args{promptID: uuid.New().String(), v1: "1", v2: "2"},
			expected: expected{statusCode: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			promptID := tt.args.promptID
			if tt.setup != nil {
				promptID = tt.setup(t, q)
			}

			handler := NewAnalyticsHandler(q).CompareVersions()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID)
			rctx.URLParams.Add("v1", tt.args.v1)
			rctx.URLParams.Add("v2", tt.args.v2)

			target := fmt.Sprintf("/prompts/%s/versions/%s/%s/compare", promptID, tt.args.v1, tt.args.v2)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s\nbody: %s", diff, w.Body.String())
			}
		})
	}

	// Data integrity check
	t.Run("response contains correct comparison data", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		promptID := setupTwoVersionData(t, q)

		handler := NewAnalyticsHandler(q).CompareVersions()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())
		rctx.URLParams.Add("v1", "1")
		rctx.URLParams.Add("v2", "2")

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID.String()+"/versions/1/2/compare", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Fatalf("status code mismatch (-want +got):\n%s\nbody: %s", diff, w.Body.String())
		}

		var result statsservice.ComparisonResult
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify structure
		if result.PromptID != promptID.String() {
			t.Errorf("expected prompt_id %s, got %s", promptID.String(), result.PromptID)
		}
		if result.VersionANumber != 1 {
			t.Errorf("expected version_a 1, got %d", result.VersionANumber)
		}
		if result.VersionBNumber != 2 {
			t.Errorf("expected version_b 2, got %d", result.VersionBNumber)
		}

		// Should have 3 metrics
		if len(result.Metrics) != 3 {
			t.Fatalf("expected 3 metrics, got %d", len(result.Metrics))
		}

		// Check metric names
		metricNames := map[string]bool{}
		for _, m := range result.Metrics {
			metricNames[m.MetricName] = true
		}
		for _, name := range []string{"latency_ms", "total_tokens", "overall_score"} {
			if !metricNames[name] {
				t.Errorf("missing metric: %s", name)
			}
		}

		// Latency: v1 ~200ms vs v2 ~100ms => significant difference, v2 wins (lower is better)
		for _, m := range result.Metrics {
			if m.MetricName == "latency_ms" {
				if !m.IsSignificant {
					t.Error("expected latency comparison to be significant")
				}
				if m.Winner != "v2" {
					t.Errorf("expected v2 to win latency, got %s", m.Winner)
				}
				if m.VersionA.N != 10 || m.VersionB.N != 10 {
					t.Errorf("expected 10 samples each, got A=%d B=%d", m.VersionA.N, m.VersionB.N)
				}
			}
			if m.MetricName == "total_tokens" {
				if !m.IsSignificant {
					t.Error("expected total_tokens comparison to be significant")
				}
				if m.Winner != "v2" {
					t.Errorf("expected v2 to win total_tokens, got %s", m.Winner)
				}
			}
		}

		// Overall winner should be v2 (wins latency and tokens)
		if result.OverallWinner != "v2" {
			t.Errorf("expected overall winner v2, got %s", result.OverallWinner)
		}
	})
}
