package task_repository

import (
	"api/src/domain/apperror"
	"api/src/domain/task"
	"context"
	"database/sql"
	"errors"
	"time"
	"utils/db/db"
	"utils/types"

	"github.com/google/uuid"
)

const dbTimeout = 5 * time.Second

func FindTaskByID(q db.Querier, ctx context.Context, id task.TaskID) types.Result[task.Task, apperror.AppError] {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	return types.Map(
		types.MapErr(
			types.FromPair(q.GetTask(ctx, uuid.UUID(id))),
			func(e error) apperror.AppError {
				if errors.Is(e, sql.ErrNoRows) {
					return apperror.NewNotFoundError(e, "Task")
				}
				if errors.Is(e, context.DeadlineExceeded) {
					return apperror.NewInternalServerError(e, "TaskRepository")
				}
				return apperror.NewDatabaseError(e, "TaskRepository")
			},
		),
		func(t db.Task) task.Task {
			return task.NewTask(
				task.TaskID(t.ID),
				task.TaskTitle(t.Title),
				task.TaskDescription(t.Description.String),
				task.TaskStatus(t.Status),
			)
		},
	)
}

func FindAllTasks(q db.Querier, ctx context.Context) types.Result[[]task.Task, apperror.AppError] {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	return types.Map(
		types.MapErr(
			types.FromPair(q.ListTasks(ctx)),
			func(e error) apperror.AppError {
				if errors.Is(e, context.DeadlineExceeded) {
					return apperror.NewInternalServerError(e, "TaskRepository")
				}
				return apperror.NewDatabaseError(e, "TaskRepository")
			},
		),
		func(tasks []db.Task) []task.Task {
			result := make([]task.Task, 0, len(tasks))
			for _, t := range tasks {
				result = append(result, task.NewTask(
					task.TaskID(t.ID),
					task.TaskTitle(t.Title),
					task.TaskDescription(t.Description.String),
					task.TaskStatus(t.Status),
				))
			}
			return result
		},
	)
}
