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

type listResponse struct {
	Tasks []taskItem `json:"tasks"`
}

type taskItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func ListHandler(q db.Querier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := types.Pipe2(
			newListRequest(r).validate(),
			func(req listRequest) types.Result[[]task.Task, apperror.AppError] {
				return task_repository.FindAllTasks(q, r.Context())
			},
			func(tasks []task.Task) listResponse {
				items := make([]taskItem, len(tasks))
				for i, t := range tasks {
					items[i] = taskItem{
						ID:          t.ID.String(),
						Title:       t.Title.String(),
						Description: t.Description.String(),
						Status:      t.Status.String(),
					}
				}
				return listResponse{Tasks: items}
			},
		)

		res.Match(
			func(resp listResponse) {
				response.OK(w, resp)
			},
			func(e apperror.AppError) {
				response.HandleAppError(w, e)
			},
		)
	}
}
