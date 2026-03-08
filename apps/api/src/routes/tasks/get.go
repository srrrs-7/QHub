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

type getResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func GetHandler(q db.Querier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := types.Pipe2(
			newGetRequest(r).validate(),
			func(req getRequest) types.Result[task.Task, apperror.AppError] {
				return task_repository.FindTaskByID(q, r.Context(), task.NewTaskID(req.ID))
			},
			func(t task.Task) getResponse {
				return getResponse{
					ID:          t.ID.String(),
					Title:       t.Title.String(),
					Description: t.Description.String(),
					Status:      t.Status.String(),
				}
			},
		)

		res.Match(
			func(resp getResponse) {
				response.OK(w, resp)
			},
			func(e apperror.AppError) {
				response.HandleAppError(w, e)
			},
		)
	}
}
