package routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"web/src/client"
	"web/src/routes"
)

func newTestRouter() http.Handler {
	return routes.NewRouter(&client.MockClient{})
}

func TestRoutes_Health(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /health: want status %d, got %d", http.StatusOK, w.Code)
	}
	if body := w.Body.String(); body != "OK" {
		t.Errorf("GET /health: want body %q, got %q", "OK", body)
	}
}

func TestRoutes_PageRoutes(t *testing.T) {
	tests := []struct {
		testName string
		method   string
		path     string
		wantCode int
	}{
		{"index page", "GET", "/", 200},
		{"organizations page", "GET", "/organizations", 200},
		{"organization detail", "GET", "/organizations/test-org", 200},
		{"projects page", "GET", "/orgs/test-org/projects", 200},
		{"prompts page", "GET", "/orgs/test-org/projects/test-project/prompts", 200},
		{"prompt detail", "GET", "/orgs/test-org/projects/test-project/prompts/test-prompt", 200},
		{"prompt detail with version", "GET", "/orgs/test-org/projects/test-project/prompts/test-prompt/v/1", 200},
		{"logs page", "GET", "/logs", 200},
		{"log detail", "GET", "/logs/log-1", 200},
		{"consulting page", "GET", "/consulting", 200},
		{"chat page", "GET", "/consulting/sess-1", 200},
		{"search page", "GET", "/search", 200},
		{"settings page", "GET", "/orgs/test-org/settings", 200},
		{"analytics page", "GET", "/analytics", 200},
		{"prompt analytics", "GET", "/analytics/prompts/prompt-1", 200},
		{"project analytics", "GET", "/analytics/projects/proj-1", 200},
		{"tags page", "GET", "/tags", 200},
		{"industries page", "GET", "/industries", 200},
		{"industry detail", "GET", "/industries/healthcare", 200},
	}

	router := newTestRouter()

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("%s %s: want status %d, got %d", tt.method, tt.path, tt.wantCode, w.Code)
			}
		})
	}
}

func TestRoutes_PartialRoutes(t *testing.T) {
	tests := []struct {
		testName    string
		method      string
		path        string
		contentType string
		body        string
		wantCode    int
	}{
		// POST routes (form-encoded)
		{"create organization", "POST", "/partials/organizations", "application/x-www-form-urlencoded", "name=Test&slug=test&plan=free", 200},
		{"create project", "POST", "/partials/orgs/org-1/projects", "application/x-www-form-urlencoded", "name=Proj&slug=proj", 200},
		{"create prompt", "POST", "/partials/projects/proj-1/prompts", "application/x-www-form-urlencoded", "name=P&slug=p&prompt_type=chat", 200},
		{"create version", "POST", "/partials/prompts/prompt-1/versions", "application/x-www-form-urlencoded", "content=hello&change_description=init", 200},
		{"create tag", "POST", "/partials/tags", "application/x-www-form-urlencoded", "name=test-tag", 200},
		{"create industry", "POST", "/partials/industries", "application/x-www-form-urlencoded", "name=Health&slug=health&description=desc", 200},
		{"create session", "POST", "/partials/consulting/sessions", "application/x-www-form-urlencoded", "title=New+Session", 200},
		{"send message", "POST", "/partials/consulting/sessions/sess-1/messages", "application/x-www-form-urlencoded", "content=hello", 200},
		{"search", "POST", "/partials/search", "application/x-www-form-urlencoded", "query=test&org_id=org-1", 200},
		{"add member", "POST", "/partials/orgs/org-1/members", "application/x-www-form-urlencoded", "user_id=u1&role=member", 200},
		{"create api key", "POST", "/partials/orgs/org-1/api-keys", "application/x-www-form-urlencoded", "name=dev-key", 200},
		{"check compliance", "POST", "/partials/industries/healthcare/compliance", "application/x-www-form-urlencoded", "content=test+prompt", 200},
		{"create evaluation", "POST", "/partials/logs/log-1/evaluations", "application/x-www-form-urlencoded", "evaluator_type=human", 200},

		// GET routes
		{"version detail", "GET", "/partials/prompts/prompt-1/versions/1", "", "", 200},
		{"lint", "GET", "/partials/prompts/prompt-1/versions/1/lint", "", "", 200},
		{"text diff", "GET", "/partials/prompts/prompt-1/versions/1/text-diff", "", "", 200},
		{"semantic diff", "GET", "/partials/prompts/prompt-1/semantic-diff/1/2", "", "", 200},
		{"compare versions", "GET", "/partials/prompts/prompt-1/compare?v1=1&v2=2", "", "", 200},
		{"prompt analytics partial", "GET", "/partials/analytics/prompts/prompt-1", "", "", 200},
		{"daily trend", "GET", "/partials/analytics/prompts/prompt-1/trend", "", "", 200},
		{"project analytics partial", "GET", "/partials/analytics/projects/proj-1", "", "", 200},
		{"embedding status", "GET", "/partials/search/embedding-status", "", "", 200},

		// PUT routes
		{"update version status", "PUT", "/partials/prompts/prompt-1/versions/1/status", "application/json", `{"status":"production"}`, 200},
		{"update org", "PUT", "/partials/orgs/test-org", "application/x-www-form-urlencoded", "name=Updated&plan=pro", 200},
		{"update member", "PUT", "/partials/orgs/org-1/members/user-1", "application/x-www-form-urlencoded", "role=admin", 200},
		{"close session", "PUT", "/partials/consulting/sessions/sess-1/close", "", "", 200},
		{"update project", "PUT", "/partials/orgs/org-1/projects/test-project", "application/x-www-form-urlencoded", "name=Updated&description=new", 200},
		{"update prompt", "PUT", "/partials/projects/proj-1/prompts/test-prompt", "application/x-www-form-urlencoded", "name=Updated&description=new", 200},

		// DELETE routes
		{"delete tag", "DELETE", "/partials/tags?name=test", "", "", 200},
		{"delete project", "DELETE", "/partials/orgs/org-1/projects/test-project", "", "", 200},
		{"delete member", "DELETE", "/partials/orgs/org-1/members/user-1", "", "", 200},
		{"delete api key", "DELETE", "/partials/orgs/org-1/api-keys/key-1", "", "", 200},
	}

	router := newTestRouter()

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("%s %s: want status %d, got %d (body: %s)", tt.method, tt.path, tt.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestRoutes_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		testName string
		method   string
		path     string
	}{
		{"POST to GET-only page route", "POST", "/organizations"},
		{"PUT to GET-only page route", "PUT", "/logs"},
		{"DELETE to GET-only page route", "DELETE", "/tags"},
		{"GET to POST-only partial", "GET", "/partials/organizations"},
		{"GET to POST-only partial projects", "GET", "/partials/orgs/org-1/projects"},
		{"GET to POST-only partial tags", "GET", "/partials/tags"},
		{"POST to GET-only partial", "POST", "/partials/prompts/prompt-1/versions/1/lint"},
		{"POST to PUT-only partial", "POST", "/partials/orgs/test-org"},
	}

	router := newTestRouter()

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s: want status %d, got %d", tt.method, tt.path, http.StatusMethodNotAllowed, w.Code)
			}
		})
	}
}

func TestRoutes_NotFound(t *testing.T) {
	tests := []struct {
		testName string
		method   string
		path     string
	}{
		{"non-existent top-level path", "GET", "/does-not-exist"},
		{"non-existent nested path", "GET", "/api/v1/anything"},
		{"non-existent partial path", "GET", "/partials/does-not-exist"},
		{"non-existent deep path", "GET", "/orgs/org/projects/proj/something-else"},
	}

	router := newTestRouter()

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("%s %s: want status %d, got %d", tt.method, tt.path, http.StatusNotFound, w.Code)
			}
		})
	}
}
