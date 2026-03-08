package task_repository

import (
	"api/src/domain/apperror"
	"api/src/domain/task"
	"context"
	"database/sql"
	"errors"
	"utils/db/db"
	"utils/types"

	"github.com/google/uuid"
)

func CreateTask(q db.Querier, ctx context.Context, cmd task.TaskCmd) types.Result[task.Task, apperror.AppError] {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	desc := cmd.Description.String()
	return types.Map(
		types.MapErr(
			types.FromPair(q.CreateTask(ctx, db.CreateTaskParams{
				Title:       cmd.Title.String(),
				Description: sql.NullString{String: desc, Valid: desc != ""},
				Status:      task.TaskStatusPending.String(), // Default status
				Priority:    "medium",                        // Default priority
			})),
			func(e error) apperror.AppError {
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

func UpdateTask(q db.Querier, ctx context.Context, id task.TaskID, cmd task.TaskCmd) types.Result[task.Task, apperror.AppError] {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	status := cmd.Status.String()
	if status == "" {
		status = task.TaskStatusPending.String()
	}

	title := cmd.Title.String()
	desc := cmd.Description.String()

	return types.Map(
		types.MapErr(
			types.FromPair(q.UpdateTask(ctx, db.UpdateTaskParams{
				ID:          uuid.UUID(id),
				Title:       sql.NullString{String: title, Valid: title != ""},
				Description: sql.NullString{String: desc, Valid: desc != ""},
				Status:      sql.NullString{String: status, Valid: status != ""},
				Priority:    sql.NullString{String: "medium", Valid: true}, // Default
			})),
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
