package task_repository

import (
	"api/src/domain/apperror"
	"api/src/domain/task"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

func (r *TaskRepository) FindByID(ctx context.Context, id task.TaskID) (task.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	t, err := r.q.GetTask(ctx, uuid.UUID(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.Task{}, apperror.NewNotFoundError(err, "Task")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return task.Task{}, apperror.NewInternalServerError(err, "TaskRepository")
		}
		return task.Task{}, apperror.NewDatabaseError(err, "TaskRepository")
	}

	return task.NewTask(
		task.TaskIDFromUUID(t.ID),
		task.TaskTitle(t.Title),
		task.TaskDescription(t.Description.String),
		task.TaskStatus(t.Status),
	), nil
}

func (r *TaskRepository) FindAll(ctx context.Context) ([]task.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tasks, err := r.q.ListTasks(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, apperror.NewInternalServerError(err, "TaskRepository")
		}
		return nil, apperror.NewDatabaseError(err, "TaskRepository")
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
