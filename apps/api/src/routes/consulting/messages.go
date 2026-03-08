package consulting

import (
	domain "api/src/domain/consulting"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

func (h *ConsultingHandler) ListMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := requtil.ParseUUID(r, "session_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		messages, err := h.messageRepo.FindAllBySession(r.Context(), sessionID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.OK(w, response.MapSlice(messages, toMessageResponse))
	}
}

func (h *ConsultingHandler) PostMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := requtil.ParseUUID(r, "session_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		req, err := decodePostMessageRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		msg := domain.Message{
			SessionID:    sessionID,
			Role:         req.Role,
			Content:      req.Content,
			Citations:    req.Citations,
			ActionsTaken: req.ActionsTaken,
		}

		created, err := h.messageRepo.Create(r.Context(), msg)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toMessageResponse(created))
	}
}
