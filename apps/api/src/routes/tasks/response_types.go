package tasks

import "api/src/domain/task"

// taskResponse is the unified JSON response type for task endpoints.
type taskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func toTaskResponse(t task.Task) taskResponse {
	return taskResponse{
		ID:          t.ID.String(),
		Title:       t.Title.String(),
		Description: t.Description.String(),
		Status:      t.Status.String(),
	}
}
