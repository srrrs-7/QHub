package tasks

import (
	"api/src/domain/apperror"
	"api/src/domain/task"
	"api/src/infra/rds/task_repository"
	"api/src/routes/response"
	"net/http"
	"utils/db/db"
	"utils/types"
)

type putResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func PutHandler(q db.Querier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := types.Pipe2(
			types.FlatMap(
				newPutRequest(r),
				func(req putRequest) types.Result[putRequest, apperror.AppError] {
					return req.validate()
				},
			),
			func(req putRequest) types.Result[task.Task, apperror.AppError] {
				status := task.TaskStatusPending
				if req.Status != "" {
					status = task.TaskStatus(req.Status)
				}
				cmd := task.NewTaskCmd(
					task.TaskTitle(req.Title),
					task.TaskDescription(req.Description),
					status,
				)
				return task_repository.UpdateTask(q, r.Context(), task.NewTaskID(req.ID), cmd)
			},
			func(t task.Task) putResponse {
				return putResponse{
					ID:          t.ID.String(),
					Title:       t.Title.String(),
					Description: t.Description.String(),
					Status:      t.Status.String(),
				}
			},
		)

		res.Match(
			func(resp putResponse) {
				response.OK(w, resp)
			},
			func(e apperror.AppError) {
				response.HandleAppError(w, e)
			},
		)
	}
}
