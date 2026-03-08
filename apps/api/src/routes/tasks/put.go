package tasks

import (
	"api/src/domain/task"
	"api/src/routes/response"
	"net/http"
)

func (h *TaskHandler) Put() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := newPutRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err = req.validate()
		if err != nil {
			response.HandleError(w, err)
			return
		}

		id, err := task.NewTaskID(req.ID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		status := task.TaskStatusPending
		if req.Status != "" {
			status = task.TaskStatus(req.Status)
		}

		cmd := task.NewTaskCmd(
			task.TaskTitle(req.Title),
			task.TaskDescription(req.Description),
			status,
		)

		t, err := h.repo.Update(r.Context(), id, cmd)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toTaskResponse(t))
	}
}
