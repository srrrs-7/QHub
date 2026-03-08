package task_repository

import (
	"api/src/domain/task"
	"api/src/infra/rds/repoerr"
	"context"

	"github.com/google/uuid"
)

// FindByID retrieves a single task by its ID.
// Returns NotFoundError if the task does not exist.
func (r *TaskRepository) FindByID(ctx context.Context, id task.TaskID) (task.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	t, err := r.q.GetTask(ctx, uuid.UUID(id))
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

// FindAll retrieves all tasks ordered by creation time.
func (r *TaskRepository) FindAll(ctx context.Context) ([]task.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tasks, err := r.q.ListTasks(ctx)
	if err != nil {
		return nil, repoerr.Handle(err, "TaskRepository", "")
	}

	result := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, task.NewTask(
			task.TaskIDFromUUID(t.ID),
			task.TaskTitle(t.Title),
			task.TaskDescription(t.Description.String),
			task.TaskStatus(t.Status),
		))
	}
	return result, nil
}
