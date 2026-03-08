package handlers

import (
	"net/http"

	"web/src/client"
	"web/src/templates/components"
)

// TaskListPartial handles HTMX requests to refresh the task list.
func TaskListPartial(apiClient *client.APIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tasks, err := apiClient.ListTasks(r.Context())
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = components.TaskList([]client.Task{}).Render(r.Context(), w)
			_ = components.Status("Error loading tasks: "+err.Error(), true).Render(r.Context(), w)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = components.TaskList(tasks.Tasks).Render(r.Context(), w)
	}
}

// AddTaskPartial handles HTMX requests to add a new task.
func AddTaskPartial(apiClient *client.APIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = components.Status("Error parsing form: "+err.Error(), true).Render(r.Context(), w)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")

		// Validate
		if title == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			tasks, _ := apiClient.ListTasks(r.Context())
			if tasks == nil {
				tasks = &client.TasksResponse{Tasks: []client.Task{}}
			}
			_ = components.TaskList(tasks.Tasks).Render(r.Context(), w)
			_ = components.Status("Title is required", true).Render(r.Context(), w)
			return
		}

		// Create task
		_, err := apiClient.CreateTask(r.Context(), client.CreateTaskRequest{
			Title:       title,
			Description: description,
		})

		// Fetch updated list
		tasks, listErr := apiClient.ListTasks(r.Context())
		if tasks == nil {
			tasks = &client.TasksResponse{Tasks: []client.Task{}}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Render task list
		_ = components.TaskList(tasks.Tasks).Render(r.Context(), w)

		// Render status message (OOB swap)
		if err != nil {
			_ = components.Status("Error adding task: "+err.Error(), true).Render(r.Context(), w)
		} else if listErr != nil {
			_ = components.Status("Task added, but failed to refresh list", true).Render(r.Context(), w)
		} else {
			_ = components.Status("Task added successfully!", false).Render(r.Context(), w)
		}
	}
}
