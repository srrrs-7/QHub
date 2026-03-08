package tasks

import (
	"api/src/routes/response"
	"net/http"
)

type listResponse struct {
	Tasks []taskResponse `json:"tasks"`
}

func (h *TaskHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := newListRequest(r)
		if _, err := req.validate(); err != nil {
			response.HandleError(w, err)
			return
		}

		tasks, err := h.repo.FindAll(r.Context())
		if err != nil {
			response.HandleError(w, err)
			return
		}

		items := make([]taskResponse, len(tasks))
		for i, t := range tasks {
			items[i] = toTaskResponse(t)
		}

		response.OK(w, listResponse{Tasks: items})
	}
}
