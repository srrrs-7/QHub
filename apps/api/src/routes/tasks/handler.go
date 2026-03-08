package tasks

import "api/src/domain/task"

// TaskHandler handles HTTP requests for task endpoints.
type TaskHandler struct {
	repo task.TaskRepository
}

// NewTaskHandler creates a new TaskHandler with the given repository.
func NewTaskHandler(repo task.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}
