package consulting

import (
	domain "api/src/domain/consulting"
	"api/src/routes/requtil"
	"api/src/routes/response"
	"api/src/services/ragservice"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Stream returns an SSE handler that streams session messages.
// GET /consulting/sessions/{session_id}/stream
//
// The handler streams existing messages for the session as SSE events,
// then sends a "done" event. It supports heartbeat pings to keep the
// connection alive and respects client disconnection via context cancellation.
//
// When RAG is available and the query parameter `rag=true` is set along with
// `org_id`, the handler will generate an AI response using the RAG pipeline
// and stream it as SSE "chunk" events before the final "done" event.
func (h *ConsultingHandler) Stream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := requtil.ParseUUID(r, "session_id")
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Verify the session exists
		session, err := h.sessionRepo.FindByID(r.Context(), sessionID)
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

		// Check if RAG generation is requested
		if h.ragSvc != nil && h.ragSvc.Available() && r.URL.Query().Get("rag") == "true" {
			orgID := resolveOrgID(r, session)
			query := r.URL.Query().Get("query")
			if orgID != uuid.Nil && query != "" {
				h.streamRAGResponse(r, sse, sessionID, query, orgID)
			}
		}

		// Signal completion
		_ = sse.WriteDone()
	}
}

// streamRAGResponse runs the RAG pipeline and streams chunks as SSE events.
// It also persists the generated assistant message in the database with
// extracted citations indicating which prompt versions were referenced.
func (h *ConsultingHandler) streamRAGResponse(r *http.Request, sse *SSEWriter, sessionID uuid.UUID, query string, orgID uuid.UUID) {
	result, err := h.ragSvc.GenerateResponse(r.Context(), sessionID, query, orgID)
	if err != nil {
		_ = sse.WriteError(err)
		return
	}

	var fullContent string
	for chunk := range result.Chunks {
		select {
		case <-r.Context().Done():
			return
		default:
			fullContent += chunk
			data, _ := json.Marshal(map[string]string{"chunk": chunk})
			if err := sse.WriteEvent("chunk", string(data)); err != nil {
				return
			}
		}
	}

	// Persist the generated assistant message with citations
	if fullContent != "" {
		// Extract citations by matching context items against the generated text
		citations := result.ExtractCitationsFromResponse(fullContent)

		msg := domain.Message{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullContent,
			Citations: ragservice.MarshalCitations(citations),
		}
		created, err := h.messageRepo.Create(r.Context(), msg)
		if err != nil {
			_ = sse.WriteError(err)
			return
		}
		// Send the final persisted message
		_ = sse.WriteMessage(toMessageResponse(created))
	}
}

// resolveOrgID extracts the org_id from the query parameter or session.
func resolveOrgID(r *http.Request, session domain.Session) uuid.UUID {
	if orgIDStr := r.URL.Query().Get("org_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			return id
		}
	}

	// Try org_id from chi URL params
	if orgIDStr := chi.URLParam(r, "org_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			return id
		}
	}

	// Fall back to session's org ID
	return session.OrgID
}
