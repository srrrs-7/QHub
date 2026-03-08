package handlers

import (
	"net/http"

	"web/src/client"
	"web/src/templates"
)

// IndexHandler renders the main page with the task list.
func IndexHandler(apiClient *client.APIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := apiClient.ListTasks(r.Context())
		if err != nil {
			// Render page with empty tasks on error
			tasks = &client.TasksResponse{Tasks: []client.Task{}}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Index(tasks.Tasks).Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

// HealthHandler returns a simple health check response.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
