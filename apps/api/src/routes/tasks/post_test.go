package tasks

import (
	"api/src/infra/rds/task_repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"utils/testutil"

	"github.com/google/go-cmp/cmp"
)

func TestPostHandler(t *testing.T) {

	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode  int
			title       string
			description string
			status      string
		}

		tests := []struct {
			name     string
			reqBody  map[string]string
			expected expected
		}{
			{
				name: "create task with title and description",
				reqBody: map[string]string{
					"title":       "New Integration Task",
					"description": "Task created in integration test",
				},
				expected: expected{
					statusCode:  http.StatusCreated,
					title:       "New Integration Task",
					description: "Task created in integration test",
					status:      "pending",
				},
			},
			{
				name: "create task with title only",
				reqBody: map[string]string{
					"title": "Task Without Description",
				},
				expected: expected{
					statusCode:  http.StatusCreated,
					title:       "Task Without Description",
					description: "",
					status:      "pending",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				repo := task_repository.NewTaskRepository(q)
				handler := NewTaskHandler(repo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if _, ok := result["id"]; !ok {
					t.Error("response should contain 'id' field")
				}

				if diff := cmp.Diff(tt.expected.title, result["title"]); diff != "" {
					t.Errorf("title mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.description, result["description"]); diff != "" {
					t.Errorf("description mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.status, result["status"]); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}
