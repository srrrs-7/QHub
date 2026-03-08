package task_repository

import (
	"api/src/domain/task"
	"api/src/infra/rds/repoerr"
	"context"
	"database/sql"
	"utils/db/db"

	"github.com/google/uuid"
)

// Create inserts a new task and returns the created entity.
// The task is created with status "pending" and priority "medium".
func (r *TaskRepository) Create(ctx context.Context, cmd task.TaskCmd) (task.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	desc := cmd.Description.String()
	t, err := r.q.CreateTask(ctx, db.CreateTaskParams{
		Title:       cmd.Title.String(),
		Description: sql.NullString{String: desc, Valid: desc != ""},
		Status:      task.TaskStatusPending.String(),
		Priority:    "medium",
	})
	if err != nil {
		return task.Task{}, repoerr.Handle(err, "TaskRepository", "")
	}

	return task.NewTask(
		task.TaskIDFromUUID(t.ID),
		task.TaskTitle(t.Title),
		task.TaskDescription(t.Description.String),
		task.TaskStatus(t.Status),
	), nil
}

// Update modifies an existing task and returns the updated entity.
// Returns NotFoundError if the task does not exist.
func (r *TaskRepository) Update(ctx context.Context, id task.TaskID, cmd task.TaskCmd) (task.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	status := cmd.Status.String()
	if status == "" {
		status = task.TaskStatusPending.String()
	}

	title := cmd.Title.String()
	desc := cmd.Description.String()

	t, err := r.q.UpdateTask(ctx, db.UpdateTaskParams{
		ID:          uuid.UUID(id),
		Title:       sql.NullString{String: title, Valid: title != ""},
		Description: sql.NullString{String: desc, Valid: desc != ""},
		Status:      sql.NullString{String: status, Valid: status != ""},
		Priority:    sql.NullString{String: "medium", Valid: true},
	})
	if err != nil {
		return task.Task{}, repoerr.Handle(err, "TaskRepository", "Task")
	}

	return task.NewTask(
		task.TaskIDFromUUID(t.ID),
		task.TaskTitle(t.Title),
		task.TaskDescription(t.Description.String),
		task.TaskStatus(t.Status),
	), nil
}
