package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"web/src/client"
	"web/src/routes"
)

// formRequest creates an HTTP request with form-encoded body.
func formRequest(method, path string, values url.Values) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// jsonRequest creates an HTTP request with JSON body.
func jsonRequest(method, path string, body any) *http.Request {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// bodyString reads and returns the response body as string.
func bodyString(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(b)
}

// --- HealthHandler ---

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		testName       string
		expectedStatus int
		expectedBody   string
	}{
		{
			testName:       "returns OK",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(&client.MockClient{})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if body != tt.expectedBody {
				t.Errorf("body: want %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

// --- CreateOrganization ---

func TestCreateOrganization(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
		isError        bool
	}{
		{
			testName: "happy path - creates organization and returns list",
			client:   &client.MockClient{},
			form: url.Values{
				"name": {"Acme Corp"},
				"slug": {"acme-corp"},
				"plan": {"pro"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Org",
			isError:        false,
		},
		{
			testName: "api error - returns snackbar error",
			client:   client.NewMockClientWithError(fmt.Errorf("api down")),
			form: url.Values{
				"name": {"Acme Corp"},
				"slug": {"acme-corp"},
				"plan": {"pro"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
			isError:        true,
		},
		{
			testName:       "empty form values - still processes",
			client:         &client.MockClient{},
			form:           url.Values{},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Org",
			isError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/organizations", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- CreateProject ---

func TestCreateProject(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - creates project and returns success snackbar",
			client:   &client.MockClient{},
			form: url.Values{
				"name":        {"My Project"},
				"slug":        {"my-project"},
				"description": {"A test project"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--success",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("forbidden")),
			form:           url.Values{"name": {"P"}, "slug": {"p"}, "description": {"d"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName:       "empty form - still processes",
			client:         &client.MockClient{},
			form:           url.Values{},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/orgs/org-1/projects", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- UpdateProject ---

func TestUpdateProject(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - updates project and returns project list",
			client:   &client.MockClient{},
			form: url.Values{
				"name":        {"Updated Project"},
				"description": {"Updated desc"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Project",
		},
		{
			testName:       "api error on update - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("update failed")),
			form:           url.Values{"name": {"X"}, "description": {"Y"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName: "api error on GetOrganization after update - returns snackbar error",
			client: &client.MockClient{
				GetOrganizationFn: func(_ context.Context, _ string) (*client.Organization, error) {
					return nil, fmt.Errorf("org not found")
				},
			},
			form:           url.Values{"name": {"X"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPut, "/partials/orgs/org-1/projects/my-project", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- DeleteProject ---

func TestDeleteProject(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - deletes project and returns project list",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Project",
		},
		{
			testName:       "api error on delete - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("delete failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/partials/orgs/org-1/projects/my-project", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- CreatePrompt ---

func TestCreatePrompt(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - creates prompt and returns prompt list",
			client:   &client.MockClient{},
			form: url.Values{
				"name":        {"My Prompt"},
				"slug":        {"my-prompt"},
				"prompt_type": {"chat"},
				"description": {"A test prompt"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Prompt",
		},
		{
			testName:       "api error on create - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("create failed")),
			form:           url.Values{"name": {"P"}, "slug": {"p"}, "prompt_type": {"chat"}, "description": {"d"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName:       "empty form values - still processes",
			client:         &client.MockClient{},
			form:           url.Values{},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/projects/proj-1/prompts", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- UpdatePrompt ---

func TestUpdatePrompt(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - updates prompt and returns header",
			client:   &client.MockClient{},
			form: url.Values{
				"name":        {"Updated Prompt"},
				"description": {"Updated desc"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "Updated Prompt",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("update failed")),
			form:           url.Values{"name": {"X"}, "description": {"Y"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPut, "/partials/projects/proj-1/prompts/my-prompt", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- CreateVersion ---

func TestCreateVersion(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		checkError     bool
	}{
		{
			testName: "happy path - creates version and returns version items",
			client:   &client.MockClient{},
			form: url.Values{
				"content":            {"Write a poem about {{topic}}"},
				"change_description": {"Initial version"},
			},
			expectedStatus: http.StatusOK,
			checkError:     false,
		},
		{
			testName:       "api error on create - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("version create failed")),
			form:           url.Values{"content": {"Hello"}, "change_description": {"test"}},
			expectedStatus: http.StatusOK,
			checkError:     true,
		},
		{
			testName:       "empty content - still processes",
			client:         &client.MockClient{},
			form:           url.Values{},
			expectedStatus: http.StatusOK,
			checkError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/prompts/prompt-1/versions", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if tt.checkError {
				if !strings.Contains(body, "snackbar--error") {
					t.Errorf("body should contain snackbar--error, got %q", body)
				}
			} else {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			}
		})
	}
}

// --- GetVersionDetail ---

func TestGetVersionDetail(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns version detail HTML",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns 404",
			client:         client.NewMockClientWithError(fmt.Errorf("not found")),
			expectedStatus: http.StatusNotFound,
			bodyContains:   "Version not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/prompts/prompt-1/versions/1", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- UpdateVersionStatus ---

func TestUpdateVersionStatus(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		body           any
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - updates status and returns version detail",
			client:         &client.MockClient{},
			body:           map[string]string{"status": "production"},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("status update failed")),
			body:           map[string]string{"status": "production"},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName:       "invalid JSON - returns 400",
			client:         &client.MockClient{},
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			bodyContains:   "Bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()

			var req *http.Request
			if tt.body == nil {
				req = httptest.NewRequest(http.MethodPut, "/partials/prompts/prompt-1/versions/1/status", strings.NewReader("not json"))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = jsonRequest(http.MethodPut, "/partials/prompts/prompt-1/versions/1/status", tt.body)
			}

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- GetLint ---

func TestGetLint(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns lint result HTML",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("lint failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/prompts/prompt-1/versions/1/lint", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- GetTextDiff ---

func TestGetTextDiff(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns text diff HTML",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("diff failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/prompts/prompt-1/versions/2/text-diff", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- CompareVersions ---

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		queryParams    string
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns comparison HTML",
			client:         &client.MockClient{},
			queryParams:    "?v1=1&v2=2",
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "missing v1 - returns 400",
			client:         &client.MockClient{},
			queryParams:    "?v2=2",
			expectedStatus: http.StatusBadRequest,
			bodyContains:   "v1 and v2 query params are required",
		},
		{
			testName:       "missing v2 - returns 400",
			client:         &client.MockClient{},
			queryParams:    "?v1=1",
			expectedStatus: http.StatusBadRequest,
			bodyContains:   "v1 and v2 query params are required",
		},
		{
			testName:       "missing both - returns 400",
			client:         &client.MockClient{},
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			bodyContains:   "v1 and v2 query params are required",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("compare failed")),
			queryParams:    "?v1=1&v2=2",
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/prompts/prompt-1/compare"+tt.queryParams, nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- GetSemanticDiff ---

func TestGetSemanticDiff(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns semantic diff HTML",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("diff failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/prompts/prompt-1/semantic-diff/1/2", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- CreateTag ---

func TestCreateTag(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - creates tag and returns tag list",
			client:         &client.MockClient{},
			form:           url.Values{"name": {"important"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "test-tag",
		},
		{
			testName:       "api error on create - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("tag create failed")),
			form:           url.Values{"name": {"important"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName:       "empty name - still processes",
			client:         &client.MockClient{},
			form:           url.Values{},
			expectedStatus: http.StatusOK,
			bodyContains:   "test-tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/tags", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- DeleteTag ---

func TestDeleteTag(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		queryParams    string
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - deletes tag and returns updated list",
			client:         &client.MockClient{},
			queryParams:    "?name=old-tag",
			expectedStatus: http.StatusOK,
			bodyContains:   "test-tag",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("delete failed")),
			queryParams:    "?name=old-tag",
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName:       "missing name param - still processes (empty name)",
			client:         &client.MockClient{},
			queryParams:    "",
			expectedStatus: http.StatusOK,
			bodyContains:   "test-tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/partials/tags"+tt.queryParams, nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- AddMember ---

func TestAddMember(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - adds member and returns member list",
			client:   &client.MockClient{},
			form: url.Values{
				"user_id": {"user-2"},
				"role":    {"admin"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error on add - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("add member failed")),
			form:           url.Values{"user_id": {"user-2"}, "role": {"admin"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/orgs/org-1/members", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- RemoveMember ---

func TestRemoveMember(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - removes member and returns member list",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("remove failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/partials/orgs/org-1/members/user-2", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- Search ---

func TestSearch(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns search results HTML",
			client:         &client.MockClient{},
			form:           url.Values{"query": {"test query"}, "org_id": {"org-1"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("search failed")),
			form:           url.Values{"query": {"test"}, "org_id": {"org-1"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
		{
			testName:       "empty query - still processes",
			client:         &client.MockClient{},
			form:           url.Values{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/search", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- CloseSession ---

func TestCloseSession(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - closes session and returns header",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("close failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/partials/consulting/sessions/sess-1/close", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- GetEmbeddingStatus ---

func TestGetEmbeddingStatus(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns embedding status badge",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns error status badge (graceful)",
			client:         client.NewMockClientWithError(fmt.Errorf("embedding service down")),
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/search/embedding-status", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			ct := resp.Header.Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("Content-Type: want text/html, got %q", ct)
			}
		})
	}
}

// --- CreateConsultingSession ---

func TestCreateConsultingSession(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - creates session and returns session list",
			client:         &client.MockClient{},
			form:           url.Values{"title": {"My Session"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "Test Session",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("session create failed")),
			form:           url.Values{"title": {"My Session"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/consulting/sessions", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- SendConsultingMessage ---

func TestSendConsultingMessage(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - sends message and returns chat message HTML",
			client:         &client.MockClient{},
			form:           url.Values{"content": {"Hello, I need help"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("send failed")),
			form:           url.Values{"content": {"Hello"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/consulting/sessions/sess-1/messages", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- CreateEvaluation ---

func TestCreateEvaluation(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - creates evaluation and returns evaluations list",
			client:   &client.MockClient{},
			form: url.Values{
				"evaluator_type": {"human"},
				"overall_score":  {"85"},
				"feedback":       {"Good quality"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("eval create failed")),
			form:           url.Values{"evaluator_type": {"human"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/logs/log-1/evaluations", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- CreateIndustry ---

func TestCreateIndustry(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName: "happy path - creates industry and returns industry list",
			client:   &client.MockClient{},
			form: url.Values{
				"name":        {"Finance"},
				"slug":        {"finance"},
				"description": {"Financial services industry"},
			},
			expectedStatus: http.StatusOK,
			bodyContains:   "Healthcare",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("industry create failed")),
			form:           url.Values{"name": {"Finance"}, "slug": {"finance"}, "description": {"Fin"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/industries", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- CheckCompliance ---

func TestCheckCompliance(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - returns compliance result HTML",
			client:         &client.MockClient{},
			form:           url.Values{"content": {"Check this prompt content"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("compliance check failed")),
			form:           url.Values{"content": {"Check this"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/industries/healthcare/compliance", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- UpdateOrganization ---

func TestUpdateOrganization(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - updates org and returns org info section",
			client:         &client.MockClient{},
			form:           url.Values{"name": {"Updated Org"}, "plan": {"enterprise"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "Updated Org",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("update failed")),
			form:           url.Values{"name": {"X"}, "plan": {"free"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPut, "/partials/orgs/test-org", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			body := bodyString(t, resp)
			if !strings.Contains(body, tt.bodyContains) {
				t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
			}
		})
	}
}

// --- UpdateMemberRole ---

func TestUpdateMemberRole(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - updates role and returns member list",
			client:         &client.MockClient{},
			form:           url.Values{"role": {"admin"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("role update failed")),
			form:           url.Values{"role": {"admin"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPut, "/partials/orgs/org-1/members/user-2", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- CreateAPIKey ---

func TestCreateAPIKey(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		form           url.Values
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - creates key and returns notice + list",
			client:         &client.MockClient{},
			form:           url.Values{"name": {"dev-key"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("key create failed")),
			form:           url.Values{"name": {"dev-key"}},
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := formRequest(http.MethodPost, "/partials/orgs/org-1/api-keys", tt.form)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- DeleteAPIKey ---

func TestDeleteAPIKey(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
		bodyContains   string
	}{
		{
			testName:       "happy path - deletes key and returns updated list",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
			bodyContains:   "text/html",
		},
		{
			testName:       "api error - returns snackbar error",
			client:         client.NewMockClientWithError(fmt.Errorf("delete key failed")),
			expectedStatus: http.StatusOK,
			bodyContains:   "snackbar--error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/partials/orgs/org-1/api-keys/key-1", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.bodyContains == "text/html" {
				ct := resp.Header.Get("Content-Type")
				if !strings.Contains(ct, "text/html") {
					t.Errorf("Content-Type: want text/html, got %q", ct)
				}
			} else {
				body := bodyString(t, resp)
				if !strings.Contains(body, tt.bodyContains) {
					t.Errorf("body should contain %q, got %q", tt.bodyContains, body)
				}
			}
		})
	}
}

// --- GetPromptAnalyticsPartial ---

func TestGetPromptAnalyticsPartial(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
	}{
		{
			testName:       "happy path - returns analytics HTML",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "api error - gracefully returns empty analytics",
			client:         client.NewMockClientWithError(fmt.Errorf("analytics failed")),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/analytics/prompts/prompt-1", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			ct := resp.Header.Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("Content-Type: want text/html, got %q", ct)
			}
		})
	}
}

// --- GetDailyTrendPartial ---

func TestGetDailyTrendPartial(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		queryParams    string
		expectedStatus int
	}{
		{
			testName:       "happy path with default days",
			client:         &client.MockClient{},
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "happy path with custom days",
			client:         &client.MockClient{},
			queryParams:    "?days=7",
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "api error - gracefully returns empty trend",
			client:         client.NewMockClientWithError(fmt.Errorf("trend failed")),
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/analytics/prompts/prompt-1/trend"+tt.queryParams, nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			ct := resp.Header.Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("Content-Type: want text/html, got %q", ct)
			}
		})
	}
}

// --- GetProjectAnalyticsPartial ---

func TestGetProjectAnalyticsPartial(t *testing.T) {
	tests := []struct {
		testName       string
		client         *client.MockClient
		expectedStatus int
	}{
		{
			testName:       "happy path - returns project analytics HTML",
			client:         &client.MockClient{},
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "api error - gracefully returns empty analytics",
			client:         client.NewMockClientWithError(fmt.Errorf("analytics failed")),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			router := routes.NewRouter(tt.client)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/partials/analytics/projects/proj-1", nil)

			router.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("status: want %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			ct := resp.Header.Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("Content-Type: want text/html, got %q", ct)
			}
		})
	}
}
