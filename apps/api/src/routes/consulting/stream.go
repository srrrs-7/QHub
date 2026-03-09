package consulting

import (
	"net/http"

	"api/src/routes/requtil"
	"api/src/routes/response"
)

// Stream returns an SSE handler that streams session messages.
// GET /consulting/sessions/{session_id}/stream
//
// The handler streams existing messages for the session as SSE events,
// then sends a "done" event. Client disconnection is detected via context
// cancellation.
func (h *ConsultingHandler) Stream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := requtil.ParseUUID(r, "session_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Verify the session exists before committing to SSE.
		if _, err := h.sessionRepo.FindByID(r.Context(), sessionID); err != nil {
			response.HandleError(w, err)
			return
		}

		// Confirm streaming is supported before setting SSE headers.
		sse, err := NewSSEWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Fetch existing messages for the session
		messages, err := h.messageRepo.FindAllBySession(r.Context(), sessionID)
		if err != nil {
			_ = sse.WriteError(err)
			return
		}

		// Stream each message as an SSE event
		for _, msg := range messages {
			select {
			case <-r.Context().Done():
				return
			default:
				if err := sse.WriteMessage(toMessageResponse(msg)); err != nil {
					return
				}
			}
		}

		// Signal completion
		_ = sse.WriteDone()
	}
}
