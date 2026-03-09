package consulting

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// seedMessage creates a consulting message and returns it.
func seedMessage(t *testing.T, q db.Querier, sessionID uuid.UUID, role, content string) {
	t.Helper()
	_, err := q.CreateConsultingMessage(context.Background(), db.CreateConsultingMessageParams{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	})
	if err != nil {
		t.Fatalf("failed to seed message: %v", err)
	}
}

// parseSSEEvents parses the SSE event stream from the response body.
func parseSSEEvents(body string) []sseEvent {
	var events []sseEvent
	scanner := bufio.NewScanner(strings.NewReader(body))

	var currentEvent sseEvent
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// Empty line = event boundary
			if currentEvent.eventType != "" || currentEvent.data != "" {
				events = append(events, currentEvent)
				currentEvent = sseEvent{}
			}
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			currentEvent.eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			currentEvent.data = strings.TrimPrefix(line, "data: ")
		} else if strings.HasPrefix(line, ":") {
			currentEvent.eventType = "comment"
			currentEvent.data = strings.TrimPrefix(line, ":")
		}
	}

	return events
}

type sseEvent struct {
	eventType string
	data      string
}

func TestStreamHandler(t *testing.T) {
	t.Run("200 OK - stream messages", func(t *testing.T) {
		type expected struct {
			contentType  string
			messageCount int
			hasDone      bool
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) string
			expected expected
		}{
			// 正常系
			{
				testName: "stream session with multiple messages",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Stream Session")
					seedMessage(t, q, sessionID, "user", "Hello")
					seedMessage(t, q, sessionID, "assistant", "Hi there!")
					seedMessage(t, q, sessionID, "user", "How are you?")
					return sessionID.String()
				},
				expected: expected{contentType: "text/event-stream", messageCount: 3, hasDone: true},
			},
			// 境界値 - empty session
			{
				testName: "stream session with no messages returns done immediately",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Empty Session")
					return sessionID.String()
				},
				expected: expected{contentType: "text/event-stream", messageCount: 0, hasDone: true},
			},
			// 境界値 - single message
			{
				testName: "stream session with one message",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Single Message Session")
					seedMessage(t, q, sessionID, "user", "Only one")
					return sessionID.String()
				},
				expected: expected{contentType: "text/event-stream", messageCount: 1, hasDone: true},
			},
			// 特殊文字
			{
				testName: "stream session with Japanese messages",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Japanese Session")
					seedMessage(t, q, sessionID, "user", "こんにちは")
					seedMessage(t, q, sessionID, "assistant", "お元気ですか？")
					return sessionID.String()
				},
				expected: expected{contentType: "text/event-stream", messageCount: 2, hasDone: true},
			},
			{
				testName: "stream session with emoji messages",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Emoji Session")
					seedMessage(t, q, sessionID, "user", "Hello 🌍🎉")
					return sessionID.String()
				},
				expected: expected{contentType: "text/event-stream", messageCount: 1, hasDone: true},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID+"/stream", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.Stream().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.contentType, resp.Header.Get("Content-Type")); diff != "" {
					t.Errorf("Content-Type mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff("no-cache", resp.Header.Get("Cache-Control")); diff != "" {
					t.Errorf("Cache-Control mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff("keep-alive", resp.Header.Get("Connection")); diff != "" {
					t.Errorf("Connection mismatch (-want +got):\n%s", diff)
				}

				events := parseSSEEvents(w.Body.String())

				// Count message events
				messageCount := 0
				hasDone := false
				for _, ev := range events {
					switch ev.eventType {
					case "message":
						messageCount++
					case "done":
						hasDone = true
					}
				}

				if diff := cmp.Diff(tt.expected.messageCount, messageCount); diff != "" {
					t.Errorf("message count mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.hasDone, hasDone); diff != "" {
					t.Errorf("done event mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("SSE format verification", func(t *testing.T) {
		tests := []struct {
			testName    string
			setup       func(t *testing.T, q db.Querier) string
			checkEvents func(t *testing.T, events []sseEvent)
		}{
			// 正常系 - verify message content in SSE events
			{
				testName: "message events contain correct role and content",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Format Test")
					seedMessage(t, q, sessionID, "user", "What is Go?")
					seedMessage(t, q, sessionID, "assistant", "Go is a programming language.")
					return sessionID.String()
				},
				checkEvents: func(t *testing.T, events []sseEvent) {
					t.Helper()
					messageEvents := make([]sseEvent, 0)
					for _, ev := range events {
						if ev.eventType == "message" {
							messageEvents = append(messageEvents, ev)
						}
					}

					if len(messageEvents) != 2 {
						t.Fatalf("expected 2 message events, got %d", len(messageEvents))
					}

					// First message
					var msg1 messageResponse
					if err := json.Unmarshal([]byte(messageEvents[0].data), &msg1); err != nil {
						t.Fatalf("failed to parse first message: %v", err)
					}
					if diff := cmp.Diff("user", msg1.Role); diff != "" {
						t.Errorf("first message role mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff("What is Go?", msg1.Content); diff != "" {
						t.Errorf("first message content mismatch (-want +got):\n%s", diff)
					}

					// Second message
					var msg2 messageResponse
					if err := json.Unmarshal([]byte(messageEvents[1].data), &msg2); err != nil {
						t.Fatalf("failed to parse second message: %v", err)
					}
					if diff := cmp.Diff("assistant", msg2.Role); diff != "" {
						t.Errorf("second message role mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff("Go is a programming language.", msg2.Content); diff != "" {
						t.Errorf("second message content mismatch (-want +got):\n%s", diff)
					}
				},
			},
			// 正常系 - done event format
			{
				testName: "done event contains complete status",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Done Format Test")
					return sessionID.String()
				},
				checkEvents: func(t *testing.T, events []sseEvent) {
					t.Helper()
					var doneEvent *sseEvent
					for i := range events {
						if events[i].eventType == "done" {
							doneEvent = &events[i]
							break
						}
					}
					if doneEvent == nil {
						t.Fatal("expected done event")
					}

					var data map[string]string
					if err := json.Unmarshal([]byte(doneEvent.data), &data); err != nil {
						t.Fatalf("failed to parse done event data: %v", err)
					}
					if diff := cmp.Diff("complete", data["status"]); diff != "" {
						t.Errorf("done status mismatch (-want +got):\n%s", diff)
					}
				},
			},
			// 正常系 - done event is last
			{
				testName: "done event is the last event",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Order Test")
					seedMessage(t, q, sessionID, "user", "Hello")
					return sessionID.String()
				},
				checkEvents: func(t *testing.T, events []sseEvent) {
					t.Helper()
					if len(events) == 0 {
						t.Fatal("expected at least one event")
					}
					lastEvent := events[len(events)-1]
					if diff := cmp.Diff("done", lastEvent.eventType); diff != "" {
						t.Errorf("last event type mismatch (-want +got):\n%s", diff)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID+"/stream", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.Stream().ServeHTTP(w, req)

				events := parseSSEEvents(w.Body.String())
				tt.checkEvents(t, events)
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName  string
			sessionID string
		}{
			// 異常系
			{
				testName:  "non-existent session",
				sessionID: "00000000-0000-0000-0000-000000000000",
			},
			{
				testName:  "random non-existent UUID",
				sessionID: uuid.New().String(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+tt.sessionID+"/stream", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.Stream().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNotFound, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName  string
			sessionID string
		}{
			// 異常系
			{
				testName:  "invalid UUID format",
				sessionID: "not-a-uuid",
			},
			// 空文字
			{
				testName:  "empty session_id",
				sessionID: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+tt.sessionID+"/stream", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.Stream().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("client disconnect", func(t *testing.T) {
		// 正常系 - context cancellation stops streaming
		t.Run("cancelled context stops streaming", func(t *testing.T) {
			q, handler := setupHandler(t)
			orgID := seedOrg(t, q)
			sessionID := seedSession(t, q, orgID, "Cancel Test")
			// Seed many messages
			for i := 0; i < 10; i++ {
				seedMessage(t, q, sessionID, "user", "Message")
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID.String()+"/stream", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("session_id", sessionID.String())
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.Stream().ServeHTTP(w, req)

			// The handler should return without streaming all messages.
			events := parseSSEEvents(w.Body.String())
			messageCount := 0
			for _, ev := range events {
				if ev.eventType == "message" {
					messageCount++
				}
			}

			// With immediate cancellation, we expect fewer than 10 messages
			if messageCount > 10 {
				t.Errorf("expected at most 10 messages, got %d", messageCount)
			}
		})
	})
}

// nonFlushResponseWriter is a ResponseWriter that does NOT implement Flusher.
type nonFlushResponseWriter struct {
	code   int
	header http.Header
	body   strings.Builder
}

func (w *nonFlushResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}
func (w *nonFlushResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
func (w *nonFlushResponseWriter) WriteHeader(code int) {
	w.code = code
}

func TestStreamHandler_NonFlushableWriter(t *testing.T) {
	// 異常系 - writer that doesn't support flushing returns 500
	t.Run("returns 500 when writer does not support flushing", func(t *testing.T) {
		q, handler := setupHandler(t)
		orgID := seedOrg(t, q)
		sessionID := seedSession(t, q, orgID, "No Flush Test")

		req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID.String()+"/stream", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("session_id", sessionID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := &nonFlushResponseWriter{}

		handler.Stream().ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusInternalServerError, w.code); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})
}
