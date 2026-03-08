package tasks

import (
	"api/src/domain/task"
	"api/src/routes/response"
	"net/http"
)

func (h *TaskHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := newPostRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err = req.validate()
		if err != nil {
			response.HandleError(w, err)
			return
		}

		cmd := task.NewTaskCmd(
			task.TaskTitle(req.Title),
			task.TaskDescription(req.Description),
			task.TaskStatusPending,
		)

		t, err := h.repo.Create(r.Context(), cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toTaskResponse(t))
	}
}
