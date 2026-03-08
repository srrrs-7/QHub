package tasks

import (
	"api/src/domain/task"
	"api/src/routes/response"
	"net/http"
)

func (h *TaskHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := newGetRequest(r)
		if _, err := req.validate(); err != nil {
			response.HandleError(w, err)
			return
		}

		id, err := task.NewTaskID(req.ID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		t, err := h.repo.FindByID(r.Context(), id)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, toTaskResponse(t))
	}
}
