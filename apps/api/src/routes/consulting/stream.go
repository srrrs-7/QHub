package consulting

import (
	"api/src/routes/requtil"
	"api/src/routes/response"
	"net/http"
)

// Stream returns an SSE handler that streams session messages.
// GET /consulting/sessions/{session_id}/stream
//
// The handler streams existing messages for the session as SSE events,
// then sends a "done" event. It supports heartbeat pings to keep the
// connection alive and respects client disconnection via context cancellation.
//
// Designed for future LLM integration: the streaming loop can be extended
// to read from a channel of messages instead of (or in addition to) the database.
func (h *ConsultingHandler) Stream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := requtil.ParseUUID(r, "session_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Verify the session exists
		_, err = h.sessionRepo.FindByID(r.Context(), sessionID)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Set SSE headers before creating the writer
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		sse, err := NewSSEWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Fetch existing messages for the session
		messages, err := h.messageRepo.FindAllBySession(r.Context(), sessionID)
		if err != nil {
			sse.WriteError(err)
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
		sse.WriteDone()
	}
}
