package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"web/src/client"
	"web/src/routes"
)

// setupRouter creates a chi router wired with the given mock client.
func setupRouter(t *testing.T, mock *client.MockClient) http.Handler {
	t.Helper()
	return routes.NewRouter(mock)
}

// assertStatus checks that the response status code matches the expected value.
func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("status code mismatch (-want +got):\n%s", diff)
	}
}

// assertContentType checks that the Content-Type header contains the expected value.
func assertContentType(t *testing.T, header http.Header, want string) {
	t.Helper()
	ct := header.Get("Content-Type")
	if !strings.Contains(ct, want) {
		t.Errorf("Content-Type = %q, want to contain %q", ct, want)
	}
}

// assertBodyContains checks that the response body contains all expected substrings.
func assertBodyContains(t *testing.T, body string, substrings ...string) {
	t.Helper()
	for _, s := range substrings {
		if !strings.Contains(body, s) {
			t.Errorf("body does not contain %q\nbody (first 500 chars): %s", s, truncate(body, 500))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// --- Test: Index ---

func TestIndexPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - returns 200 with welcome page",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab", "Welcome to PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Organizations ---

func TestOrganizationsPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - lists organizations",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Organizations", "Test Org"},
		},
		{
			testName: "API error - renders empty list",
			mock: &client.MockClient{
				ListOrganizationsFn: func(_ context.Context) ([]client.Organization, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Organizations", "No organizations yet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: OrganizationDetail ---

func TestOrganizationDetailPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows org with projects",
			path:           "/organizations/test-org",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Org", "Test Project"},
		},
		{
			testName: "org not found - returns 404",
			path:     "/organizations/nonexistent",
			mock: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, _ string) (*client.Organization, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Organization not found"},
		},
		{
			testName: "projects API error - renders org with empty projects",
			path:     "/organizations/test-org",
			mock: &client.MockClient{
				ListProjectsFn: func(_ context.Context, _ string) ([]client.Project, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Org"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Projects ---

func TestProjectsPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows projects for org",
			path:           "/orgs/test-org/projects",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Project"},
		},
		{
			testName: "org not found - returns 404",
			path:     "/orgs/nonexistent/projects",
			mock: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, _ string) (*client.Organization, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Organization not found"},
		},
		{
			testName: "projects API error - renders empty list",
			path:     "/orgs/test-org/projects",
			mock: &client.MockClient{
				ListProjectsFn: func(_ context.Context, _ string) ([]client.Project, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"test-org"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Prompts ---

func TestPromptsPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows prompts for project",
			path:           "/orgs/test-org/projects/test-project/prompts",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Prompt"},
		},
		{
			testName: "org not found - returns 404",
			path:     "/orgs/nonexistent/projects/test-project/prompts",
			mock: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, _ string) (*client.Organization, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Organization not found"},
		},
		{
			testName: "project not found - returns 404",
			path:     "/orgs/test-org/projects/nonexistent/prompts",
			mock: &client.MockClient{
				GetProjectFn: func(_ context.Context, _, _ string) (*client.Project, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Project not found"},
		},
		{
			testName: "prompts API error - renders empty list",
			path:     "/orgs/test-org/projects/test-project/prompts",
			mock: &client.MockClient{
				ListPromptsFn: func(_ context.Context, _ string) ([]client.Prompt, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"test-project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: PromptDetail ---

func TestPromptDetailPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows prompt with versions",
			path:           "/orgs/test-org/projects/test-project/prompts/test-prompt",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Prompt", "draft"},
		},
		{
			testName:       "happy path with specific version",
			path:           "/orgs/test-org/projects/test-project/prompts/test-prompt/v/1",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Prompt"},
		},
		{
			testName: "org not found - returns 404",
			path:     "/orgs/nonexistent/projects/test-project/prompts/test-prompt",
			mock: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, _ string) (*client.Organization, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Organization not found"},
		},
		{
			testName: "project not found - returns 404",
			path:     "/orgs/test-org/projects/nonexistent/prompts/test-prompt",
			mock: &client.MockClient{
				GetProjectFn: func(_ context.Context, _, _ string) (*client.Project, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Project not found"},
		},
		{
			testName: "prompt not found - returns 404",
			path:     "/orgs/test-org/projects/test-project/prompts/nonexistent",
			mock: &client.MockClient{
				GetPromptFn: func(_ context.Context, _, _ string) (*client.Prompt, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Prompt not found"},
		},
		{
			testName: "versions API error - renders prompt with empty versions",
			path:     "/orgs/test-org/projects/test-project/prompts/test-prompt",
			mock: &client.MockClient{
				ListVersionsFn: func(_ context.Context, _ string) ([]client.PromptVersion, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Logs ---

func TestLogsPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - lists logs",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"gpt-4", "success"},
		},
		{
			testName: "API error - renders empty list",
			mock: &client.MockClient{
				ListLogsFn: func(_ context.Context) ([]client.ExecutionLog, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/logs", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: LogDetail ---

func TestLogDetailPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows log detail",
			path:           "/logs/log-123",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"gpt-4", "success"},
		},
		{
			testName: "log not found - returns 404",
			path:     "/logs/nonexistent",
			mock: &client.MockClient{
				GetLogFn: func(_ context.Context, _ string) (*client.ExecutionLog, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Log not found"},
		},
		{
			testName: "evaluations API error - renders log with empty evals",
			path:     "/logs/log-123",
			mock: &client.MockClient{
				ListLogEvaluationFn: func(_ context.Context, _ string) ([]client.Evaluation, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"gpt-4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Consulting ---

func TestConsultingPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - lists sessions",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Session"},
		},
		{
			testName: "API error - renders empty list",
			mock: &client.MockClient{
				ListConsultingSessionsFn: func(_ context.Context) ([]client.ConsultingSession, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/consulting", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Chat ---

func TestChatPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows chat session",
			path:           "/consulting/sess-123",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Session"},
		},
		{
			testName: "session not found - returns 404",
			path:     "/consulting/nonexistent",
			mock: &client.MockClient{
				GetConsultingSessionFn: func(_ context.Context, _ string) (*client.ConsultingSession, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Session not found"},
		},
		{
			testName: "messages API error - renders chat with empty messages",
			path:     "/consulting/sess-123",
			mock: &client.MockClient{
				ListConsultingMessagesFn: func(_ context.Context, _ string) ([]client.ConsultingMessage, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Session"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Search ---

func TestSearchPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - returns 200 with search page",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/search", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Settings ---

func TestSettingsPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows org settings",
			path:           "/orgs/test-org/settings",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Org"},
		},
		{
			testName: "org not found - returns 404",
			path:     "/orgs/nonexistent/settings",
			mock: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, _ string) (*client.Organization, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Organization not found"},
		},
		{
			testName: "members API error - renders settings with empty members",
			path:     "/orgs/test-org/settings",
			mock: &client.MockClient{
				ListMembersFn: func(_ context.Context, _ string) ([]client.Member, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Org"},
		},
		{
			testName: "API keys error - renders settings with empty keys",
			path:     "/orgs/test-org/settings",
			mock: &client.MockClient{
				ListAPIKeysFn: func(_ context.Context, _ string) ([]client.APIKey, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Test Org"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Analytics ---

func TestAnalyticsPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows analytics dashboard",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
		{
			testName:       "all API errors - renders with empty data",
			mock:           client.NewMockClientWithError(fmt.Errorf("api error")),
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: PromptAnalytics ---

func TestPromptAnalyticsPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows prompt analytics",
			path:           "/analytics/prompts/prompt-123",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab", "prompt-123"},
		},
		{
			testName: "API error - renders with empty data",
			path:     "/analytics/prompts/prompt-123",
			mock: &client.MockClient{
				GetPromptAnalyticsFn: func(_ context.Context, _ string) ([]client.PromptAnalytics, error) {
					return nil, fmt.Errorf("api error")
				},
				GetDailyTrendFn: func(_ context.Context, _, _ string) ([]client.DailyTrend, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: ProjectAnalytics ---

func TestProjectAnalyticsPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows project analytics",
			path:           "/analytics/projects/proj-123",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
		{
			testName: "API error - renders with empty data",
			path:     "/analytics/projects/proj-123",
			mock: &client.MockClient{
				GetProjectAnalyticsFn: func(_ context.Context, _ string) ([]client.ProjectAnalytics, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Tags ---

func TestTagsPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - lists tags",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"test-tag"},
		},
		{
			testName: "API error - renders empty list",
			mock: &client.MockClient{
				ListTagsFn: func(_ context.Context) ([]client.Tag, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/tags", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Industries ---

func TestIndustriesPage(t *testing.T) {
	tests := []struct {
		testName       string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - lists industries",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Healthcare"},
		},
		{
			testName: "API error - renders empty list",
			mock: &client.MockClient{
				ListIndustriesFn: func(_ context.Context) ([]client.Industry, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"PromptLab"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, "/industries", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			assertContentType(t, w.Header(), "text/html")
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: IndustryDetail ---

func TestIndustryDetailPage(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
		expectedBody   []string
	}{
		{
			testName:       "happy path - shows industry detail",
			path:           "/industries/healthcare",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Healthcare"},
		},
		{
			testName: "industry not found - returns 404",
			path:     "/industries/nonexistent",
			mock: &client.MockClient{
				GetIndustryFn: func(_ context.Context, _ string) (*client.Industry, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Industry not found"},
		},
		{
			testName: "benchmarks API error - renders industry with empty benchmarks",
			path:     "/industries/healthcare",
			mock: &client.MockClient{
				ListBenchmarksFn: func(_ context.Context, _ string) ([]client.Benchmark, error) {
					return nil, fmt.Errorf("api error")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"Healthcare"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
			if tt.expectedStatus == http.StatusOK {
				assertContentType(t, w.Header(), "text/html")
			}
			assertBodyContains(t, w.Body.String(), tt.expectedBody...)
		})
	}
}

// --- Test: Special characters and edge cases ---

func TestPageHandlersSpecialCharacters(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		mock           *client.MockClient
		expectedStatus int
	}{
		{
			testName: "org slug with unicode - org not found is 404",
			path:     "/organizations/%E3%83%86%E3%82%B9%E3%83%88",
			mock: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, slug string) (*client.Organization, error) {
					return nil, fmt.Errorf("not found")
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			testName: "log ID with special chars",
			path:     "/logs/log-with-dashes-123",
			mock: &client.MockClient{
				GetLogFn: func(_ context.Context, id string) (*client.ExecutionLog, error) {
					if id != "log-with-dashes-123" {
						t.Errorf("unexpected log ID: %s", id)
					}
					return &client.ExecutionLog{ID: id, Model: "gpt-4", Status: "success"}, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "empty path segment handled by router as 404",
			path:           "/organizations/",
			mock:           &client.MockClient{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, tt.mock)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
		})
	}
}

// --- Test: Method not allowed ---

func TestPageHandlersMethodNotAllowed(t *testing.T) {
	tests := []struct {
		testName       string
		method         string
		path           string
		expectedStatus int
	}{
		{
			testName:       "POST to index returns 405",
			method:         http.MethodPost,
			path:           "/",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			testName:       "PUT to organizations returns 405",
			method:         http.MethodPut,
			path:           "/organizations",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			testName:       "DELETE to logs returns 405",
			method:         http.MethodDelete,
			path:           "/logs",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, &client.MockClient{})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
		})
	}
}

// --- Test: Nonexistent routes ---

func TestPageHandlersNotFoundRoutes(t *testing.T) {
	tests := []struct {
		testName       string
		path           string
		expectedStatus int
	}{
		{
			testName:       "completely unknown path returns 404",
			path:           "/this/does/not/exist",
			expectedStatus: http.StatusNotFound,
		},
		{
			testName:       "API prefix without match returns 404",
			path:           "/api/v1/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := setupRouter(t, &client.MockClient{})
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assertStatus(t, w.Code, tt.expectedStatus)
		})
	}
}

// --- Test: Full flow with custom mock data ---

func TestOrganizationDetailPageWithMultipleProjects(t *testing.T) {
	mock := &client.MockClient{
		GetOrganizationFn: func(_ context.Context, slug string) (*client.Organization, error) {
			return &client.Organization{ID: "org-42", Name: "Acme Corp", Slug: slug, Plan: "enterprise"}, nil
		},
		ListProjectsFn: func(_ context.Context, orgID string) ([]client.Project, error) {
			return []client.Project{
				{ID: "p1", OrganizationID: orgID, Name: "Alpha Project", Slug: "alpha"},
				{ID: "p2", OrganizationID: orgID, Name: "Beta Project", Slug: "beta"},
				{ID: "p3", OrganizationID: orgID, Name: "Gamma Project", Slug: "gamma"},
			}, nil
		},
	}

	router := setupRouter(t, mock)
	req := httptest.NewRequest(http.MethodGet, "/organizations/acme", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w.Header(), "text/html")
	body := w.Body.String()
	assertBodyContains(t, body, "Acme Corp", "Alpha Project", "Beta Project", "Gamma Project")
}

func TestChatPageWithMessages(t *testing.T) {
	mock := &client.MockClient{
		GetConsultingSessionFn: func(_ context.Context, id string) (*client.ConsultingSession, error) {
			return &client.ConsultingSession{ID: id, Title: "Debug Session", Status: "active"}, nil
		},
		ListConsultingMessagesFn: func(_ context.Context, _ string) ([]client.ConsultingMessage, error) {
			return []client.ConsultingMessage{
				{ID: "m1", Role: "user", Content: "Hello there"},
				{ID: "m2", Role: "assistant", Content: "Hi, how can I help?"},
			}, nil
		},
	}

	router := setupRouter(t, mock)
	req := httptest.NewRequest(http.MethodGet, "/consulting/sess-42", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w.Header(), "text/html")
	body := w.Body.String()
	assertBodyContains(t, body, "Debug Session", "Hello there", "Hi, how can I help?")
}

func TestPromptDetailWithSpecificVersionParam(t *testing.T) {
	versionRequested := false
	mock := &client.MockClient{
		GetVersionFn: func(_ context.Context, promptID, version string) (*client.PromptVersion, error) {
			versionRequested = true
			if version != "2" {
				t.Errorf("expected version 2, got %s", version)
			}
			return &client.PromptVersion{
				ID:            "ver-2",
				PromptID:      promptID,
				VersionNumber: 2,
				Status:        "review",
			}, nil
		},
	}

	router := setupRouter(t, mock)
	req := httptest.NewRequest(http.MethodGet, "/orgs/test-org/projects/test-project/prompts/test-prompt/v/2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w.Header(), "text/html")
	if !versionRequested {
		t.Error("expected GetVersion to be called with version param")
	}
}

func TestSettingsPageWithAllData(t *testing.T) {
	mock := &client.MockClient{
		GetOrganizationFn: func(_ context.Context, slug string) (*client.Organization, error) {
			return &client.Organization{ID: "org-1", Name: "Settings Org", Slug: slug, Plan: "pro"}, nil
		},
		ListMembersFn: func(_ context.Context, orgID string) ([]client.Member, error) {
			return []client.Member{
				{OrganizationID: orgID, UserID: "user-1", Role: "owner"},
				{OrganizationID: orgID, UserID: "user-2", Role: "member"},
			}, nil
		},
		ListAPIKeysFn: func(_ context.Context, orgID string) ([]client.APIKey, error) {
			return []client.APIKey{
				{ID: "key-1", OrganizationID: orgID, Name: "prod-key", KeyPrefix: "qh_prod"},
			}, nil
		},
	}

	router := setupRouter(t, mock)
	req := httptest.NewRequest(http.MethodGet, "/orgs/test-org/settings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w.Header(), "text/html")
	body := w.Body.String()
	assertBodyContains(t, body, "Settings Org")
}

func TestAnalyticsPageAggregatesAcrossOrgs(t *testing.T) {
	mock := &client.MockClient{
		ListLogsFn: func(_ context.Context) ([]client.ExecutionLog, error) {
			return []client.ExecutionLog{
				{ID: "log-1", Model: "gpt-4", Status: "success", TotalTokens: 100},
				{ID: "log-2", Model: "gpt-4", Status: "error", TotalTokens: 50},
			}, nil
		},
		ListEvaluationsFn: func(_ context.Context) ([]client.Evaluation, error) {
			return []client.Evaluation{}, nil
		},
		ListConsultingSessionsFn: func(_ context.Context) ([]client.ConsultingSession, error) {
			return []client.ConsultingSession{}, nil
		},
		ListOrganizationsFn: func(_ context.Context) ([]client.Organization, error) {
			return []client.Organization{
				{ID: "org-a", Name: "Org A", Slug: "org-a"},
			}, nil
		},
		ListProjectsFn: func(_ context.Context, orgID string) ([]client.Project, error) {
			return []client.Project{
				{ID: "proj-1", OrganizationID: orgID, Name: "Project One", Slug: "project-one"},
			}, nil
		},
	}

	router := setupRouter(t, mock)
	req := httptest.NewRequest(http.MethodGet, "/analytics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w.Header(), "text/html")
}
