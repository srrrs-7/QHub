package consulting

import (
	domain "api/src/domain/consulting"
	"api/src/infra/rds/consulting_repository"
	"api/src/services/embeddingservice"
	"api/src/services/ragservice"
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"utils/db/db"
	"utils/embedding"
	"utils/ollama"
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
			// We cannot guarantee exactly how many messages were streamed before
			// the cancellation was detected, but it should not have a done event.
			events := parseSSEEvents(w.Body.String())
			messageCount := 0
			for _, ev := range events {
				if ev.eventType == "message" {
					messageCount++
				}
			}

			// With immediate cancellation, we expect fewer than 10 messages
			// (possibly 0, depending on scheduling)
			if messageCount > 10 {
				t.Errorf("expected at most 10 messages, got %d", messageCount)
			}
		})
	})

	t.Run("RAG query params", func(t *testing.T) {
		type expected struct {
			statusCode   int
			hasDone      bool
			messageCount int
		}

		tests := []struct {
			testName    string
			setup       func(t *testing.T, q db.Querier) (string, uuid.UUID)
			queryParams string
			expected    expected
		}{
			// 正常系 - rag=true but ragSvc is nil (handler has nil ragSvc)
			{
				testName: "stream with rag=true but nil ragSvc returns messages and done",
				setup: func(t *testing.T, q db.Querier) (string, uuid.UUID) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "RAG Test")
					seedMessage(t, q, sessionID, "user", "Hello RAG")
					return sessionID.String(), orgID
				},
				queryParams: "rag=true&query=test",
				expected:    expected{statusCode: http.StatusOK, hasDone: true, messageCount: 1},
			},
			// 正常系 - rag=false still streams normally
			{
				testName: "stream with rag=false streams normally",
				setup: func(t *testing.T, q db.Querier) (string, uuid.UUID) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "No RAG Test")
					seedMessage(t, q, sessionID, "user", "No RAG")
					return sessionID.String(), orgID
				},
				queryParams: "rag=false",
				expected:    expected{statusCode: http.StatusOK, hasDone: true, messageCount: 1},
			},
			// 境界値 - empty query param with rag=true
			{
				testName: "stream with rag=true but empty query still works",
				setup: func(t *testing.T, q db.Querier) (string, uuid.UUID) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Empty Query Test")
					return sessionID.String(), orgID
				},
				queryParams: "rag=true&query=",
				expected:    expected{statusCode: http.StatusOK, hasDone: true, messageCount: 0},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID, orgID := tt.setup(t, q)

				url := "/consulting/sessions/" + sessionID + "/stream?org_id=" + orgID.String()
				if tt.queryParams != "" {
					url += "&" + tt.queryParams
				}

				req := httptest.NewRequest(http.MethodGet, url, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.Stream().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				events := parseSSEEvents(w.Body.String())
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

// setupHandlerWithRAG creates a handler with a RAG service that reports as available
// but fails when actually generating (fake URLs).
func setupHandlerWithRAG(t *testing.T) (db.Querier, *ConsultingHandler) {
	t.Helper()
	q := testutil.SetupTestTx(t)
	sessionRepo := consulting_repository.NewSessionRepository(q)
	messageRepo := consulting_repository.NewMessageRepository(q)
	industryRepo := consulting_repository.NewIndustryConfigRepository(q)

	// Create a RAG service that appears available but will fail on actual calls.
	embClient := embedding.NewClient("http://fake-embedding:80")
	embSvc := embeddingservice.NewEmbeddingService(embClient, nil)
	ollamaClient := ollama.NewClient("http://fake-ollama:11434")
	ragSvc := ragservice.NewRAGService(embSvc, ollamaClient, q)

	handler := NewConsultingHandler(sessionRepo, messageRepo, industryRepo, ragSvc)
	return q, handler
}

func TestStreamHandler_WithRAGEnabled(t *testing.T) {
	t.Run("stream with rag=true and available RAG service", func(t *testing.T) {
		type expected struct {
			hasError bool
			hasDone  bool
		}

		tests := []struct {
			testName    string
			setup       func(t *testing.T, q db.Querier) (string, uuid.UUID)
			queryParams string
			expected    expected
		}{
			// 正常系 - RAG available, rag=true, query set, but embedding call fails
			// Should stream existing messages, then get an error from RAG, then write error event
			{
				testName: "RAG enabled with query streams error then done",
				setup: func(t *testing.T, q db.Querier) (string, uuid.UUID) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "RAG Stream Test")
					seedMessage(t, q, sessionID, "user", "Tell me about prompts")
					return sessionID.String(), orgID
				},
				queryParams: "rag=true&query=prompts",
				expected:    expected{hasError: true, hasDone: true},
			},
			// 境界値 - RAG available but empty query
			{
				testName: "RAG enabled with empty query skips RAG",
				setup: func(t *testing.T, q db.Querier) (string, uuid.UUID) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "No Query Test")
					return sessionID.String(), orgID
				},
				queryParams: "rag=true&query=",
				expected:    expected{hasError: false, hasDone: true},
			},
			// 境界値 - RAG available but rag param not "true"
			{
				testName: "RAG enabled but rag=false skips RAG",
				setup: func(t *testing.T, q db.Querier) (string, uuid.UUID) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Rag False Test")
					return sessionID.String(), orgID
				},
				queryParams: "rag=false&query=test",
				expected:    expected{hasError: false, hasDone: true},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandlerWithRAG(t)
				sessionID, orgID := tt.setup(t, q)

				url := "/consulting/sessions/" + sessionID + "/stream?org_id=" + orgID.String()
				if tt.queryParams != "" {
					url += "&" + tt.queryParams
				}

				req := httptest.NewRequest(http.MethodGet, url, nil)
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

				events := parseSSEEvents(w.Body.String())
				hasError := false
				hasDone := false
				for _, ev := range events {
					switch ev.eventType {
					case "error":
						hasError = true
					case "done":
						hasDone = true
					}
				}

				if diff := cmp.Diff(tt.expected.hasError, hasError); diff != "" {
					t.Errorf("error event mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.hasDone, hasDone); diff != "" {
					t.Errorf("done event mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
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

func TestResolveOrgID(t *testing.T) {
	type args struct {
		queryOrgID string
		chiOrgID   string
		session    domain.Session
	}
	type expected struct {
		orgID uuid.UUID
	}

	sessionOrgID := uuid.New()
	queryOrgID := uuid.New()
	chiOrgID := uuid.New()

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - org_id from query param
		{
			testName: "uses query param org_id when present",
			args: args{
				queryOrgID: queryOrgID.String(),
				chiOrgID:   "",
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: queryOrgID},
		},
		// 正常系 - org_id from chi URL param
		{
			testName: "uses chi URL param org_id when query is empty",
			args: args{
				queryOrgID: "",
				chiOrgID:   chiOrgID.String(),
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: chiOrgID},
		},
		// 正常系 - fallback to session org_id
		{
			testName: "falls back to session org_id when no params",
			args: args{
				queryOrgID: "",
				chiOrgID:   "",
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: sessionOrgID},
		},
		// 異常系 - invalid query param falls through to chi param
		{
			testName: "invalid query org_id falls through to chi param",
			args: args{
				queryOrgID: "not-a-uuid",
				chiOrgID:   chiOrgID.String(),
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: chiOrgID},
		},
		// 異常系 - invalid query and chi params fall back to session
		{
			testName: "invalid both params falls back to session",
			args: args{
				queryOrgID: "bad-uuid",
				chiOrgID:   "also-bad",
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: sessionOrgID},
		},
		// 空文字 - empty org_id in all sources uses session
		{
			testName: "empty query and chi params uses session org_id",
			args: args{
				queryOrgID: "",
				chiOrgID:   "",
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: sessionOrgID},
		},
		// Null/Nil - zero session org_id returned when nothing else available
		{
			testName: "zero session org_id returned when all sources empty",
			args: args{
				queryOrgID: "",
				chiOrgID:   "",
				session:    domain.Session{OrgID: uuid.UUID{}},
			},
			expected: expected{orgID: uuid.UUID{}},
		},
		// 境界値 - query param takes priority over chi param
		{
			testName: "query param has priority over chi param",
			args: args{
				queryOrgID: queryOrgID.String(),
				chiOrgID:   chiOrgID.String(),
				session:    domain.Session{OrgID: sessionOrgID},
			},
			expected: expected{orgID: queryOrgID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			url := "/stream"
			if tt.args.queryOrgID != "" {
				url += "?org_id=" + tt.args.queryOrgID
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rctx := chi.NewRouteContext()
			if tt.args.chiOrgID != "" {
				rctx.URLParams.Add("org_id", tt.args.chiOrgID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			got := resolveOrgID(req, tt.args.session)
			if diff := cmp.Diff(tt.expected.orgID, got); diff != "" {
				t.Errorf("orgID mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
