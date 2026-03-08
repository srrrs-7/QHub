package client

// Task represents a task from the API.
// Note: API returns "status" as string ("pending" or "completed"), not boolean.
type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// IsCompleted returns true if the task status is "completed".
func (t Task) IsCompleted() bool {
	return t.Status == "completed"
}

// TasksResponse represents the API response for listing tasks.
type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

// CreateTaskRequest represents the request body for creating a task.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
