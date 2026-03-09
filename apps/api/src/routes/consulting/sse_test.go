package consulting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// flushRecorder wraps httptest.ResponseRecorder and implements http.Flusher.
type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed int
}

func (f *flushRecorder) Flush() {
	f.flushed++
}

// nonFlusher is an http.ResponseWriter that does NOT implement http.Flusher.
type nonFlusher struct {
	http.ResponseWriter
}

func TestNewSSEWriter(t *testing.T) {
	type args struct {
		w http.ResponseWriter
	}
	type expected struct {
		wantErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "create with flusher-capable writer",
			args:     args{w: &flushRecorder{ResponseRecorder: httptest.NewRecorder()}},
			expected: expected{wantErr: false},
		},
		// 異常系
		{
			testName: "create with non-flusher writer returns error",
			args:     args{w: &nonFlusher{}},
			expected: expected{wantErr: true},
		},
		// Null/Nil - nil writer wrapped in nonFlusher
		{
			testName: "create with nil underlying writer returns error",
			args:     args{w: &nonFlusher{ResponseWriter: nil}},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := NewSSEWriter(tt.args.w)
			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if got != nil {
					t.Error("expected nil SSEWriter on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Fatal("expected non-nil SSEWriter")
				}
			}
		})
	}
}

func TestSSEWriter_WriteEvent(t *testing.T) {
	type args struct {
		event string
		data  string
	}
	type expected struct {
		output  string
		flushed bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "write message event",
			args:     args{event: "message", data: `{"role":"assistant","content":"hello"}`},
			expected: expected{output: "event: message\ndata: {\"role\":\"assistant\",\"content\":\"hello\"}\n\n", flushed: true},
		},
		{
			testName: "write done event",
			args:     args{event: "done", data: `{"status":"complete"}`},
			expected: expected{output: "event: done\ndata: {\"status\":\"complete\"}\n\n", flushed: true},
		},
		{
			testName: "write error event",
			args:     args{event: "error", data: `{"error":"something went wrong"}`},
			expected: expected{output: "event: error\ndata: {\"error\":\"something went wrong\"}\n\n", flushed: true},
		},
		// 特殊文字
		{
			testName: "write event with Japanese data",
			args:     args{event: "message", data: `{"content":"こんにちは"}`},
			expected: expected{output: "event: message\ndata: {\"content\":\"こんにちは\"}\n\n", flushed: true},
		},
		{
			testName: "write event with emoji data",
			args:     args{event: "message", data: `{"content":"Hello 🌍"}`},
			expected: expected{output: "event: message\ndata: {\"content\":\"Hello 🌍\"}\n\n", flushed: true},
		},
		// 空文字
		{
			testName: "write event with empty data",
			args:     args{event: "message", data: ""},
			expected: expected{output: "event: message\ndata: \n\n", flushed: true},
		},
		// 境界値
		{
			testName: "write event with single character data",
			args:     args{event: "message", data: "x"},
			expected: expected{output: "event: message\ndata: x\n\n", flushed: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
			sw, err := NewSSEWriter(rec)
			if err != nil {
				t.Fatalf("unexpected error creating SSEWriter: %v", err)
			}

			err = sw.WriteEvent(tt.args.event, tt.args.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected.output, rec.Body.String()); diff != "" {
				t.Errorf("output mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.flushed && rec.flushed == 0 {
				t.Error("expected flush to be called")
			}
		})
	}
}

func TestSSEWriter_WriteMessage(t *testing.T) {
	type expected struct {
		role    string
		content string
	}

	tests := []struct {
		testName string
		msg      messageResponse
		expected expected
	}{
		// 正常系
		{
			testName: "write assistant message",
			msg:      messageResponse{ID: "id-1", SessionID: "sess-1", Role: "assistant", Content: "Hello there", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: "Hello there"},
		},
		{
			testName: "write user message",
			msg:      messageResponse{ID: "id-2", SessionID: "sess-1", Role: "user", Content: "Help me", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "user", content: "Help me"},
		},
		// 特殊文字
		{
			testName: "write message with Japanese content",
			msg:      messageResponse{ID: "id-3", SessionID: "sess-1", Role: "assistant", Content: "日本語の応答", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: "日本語の応答"},
		},
		// 空文字 - empty citations/actions
		{
			testName: "write message with null citations",
			msg:      messageResponse{ID: "id-4", SessionID: "sess-1", Role: "assistant", Content: "response", Citations: nil, CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: "response"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
			sw, err := NewSSEWriter(rec)
			if err != nil {
				t.Fatalf("unexpected error creating SSEWriter: %v", err)
			}

			err = sw.WriteMessage(tt.msg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			body := rec.Body.String()
			if !strings.HasPrefix(body, "event: message\ndata: ") {
				t.Errorf("expected SSE message event prefix, got: %s", body)
			}

			// Extract JSON data from SSE format
			dataLine := strings.TrimPrefix(body, "event: message\ndata: ")
			dataLine = strings.TrimSuffix(dataLine, "\n\n")

			var parsed messageResponse
			if err := json.Unmarshal([]byte(dataLine), &parsed); err != nil {
				t.Fatalf("failed to parse JSON data: %v", err)
			}

			if diff := cmp.Diff(tt.expected.role, parsed.Role); diff != "" {
				t.Errorf("role mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.content, parsed.Content); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSSEWriter_WriteMessage_EdgeCases(t *testing.T) {
	type expected struct {
		role    string
		content string
	}

	tests := []struct {
		testName string
		msg      messageResponse
		expected expected
	}{
		// 境界値 - empty content
		{
			testName: "write message with empty content",
			msg:      messageResponse{ID: "id-empty", SessionID: "sess-1", Role: "assistant", Content: "", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: ""},
		},
		// 特殊文字 - emoji content
		{
			testName: "write message with emoji content",
			msg:      messageResponse{ID: "id-emoji", SessionID: "sess-1", Role: "assistant", Content: "Great work! 🎉👍", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: "Great work! 🎉👍"},
		},
		// 正常系 - message with citations
		{
			testName: "write message with citations JSON",
			msg:      messageResponse{ID: "id-cite", SessionID: "sess-1", Role: "assistant", Content: "Based on data", Citations: json.RawMessage(`[{"source":"doc.pdf"}]`), CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: "Based on data"},
		},
		// 正常系 - message with actions_taken
		{
			testName: "write message with actions taken",
			msg:      messageResponse{ID: "id-action", SessionID: "sess-1", Role: "assistant", Content: "Analyzed", ActionsTaken: json.RawMessage(`["search","analyze"]`), CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "assistant", content: "Analyzed"},
		},
		// 境界値 - system role
		{
			testName: "write system message",
			msg:      messageResponse{ID: "id-sys", SessionID: "sess-1", Role: "system", Content: "You are a helpful assistant.", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "system", content: "You are a helpful assistant."},
		},
		// 特殊文字 - SQL injection in content
		{
			testName: "write message with SQL injection content",
			msg:      messageResponse{ID: "id-sql", SessionID: "sess-1", Role: "user", Content: "'; DROP TABLE users; --", CreatedAt: "2026-01-01T00:00:00Z"},
			expected: expected{role: "user", content: "'; DROP TABLE users; --"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
			sw, err := NewSSEWriter(rec)
			if err != nil {
				t.Fatalf("unexpected error creating SSEWriter: %v", err)
			}

			err = sw.WriteMessage(tt.msg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			body := rec.Body.String()
			if !strings.HasPrefix(body, "event: message\ndata: ") {
				t.Errorf("expected SSE message event prefix, got: %s", body)
			}

			// Extract JSON data from SSE format
			dataLine := strings.TrimPrefix(body, "event: message\ndata: ")
			dataLine = strings.TrimSuffix(dataLine, "\n\n")

			var parsed messageResponse
			if err := json.Unmarshal([]byte(dataLine), &parsed); err != nil {
				t.Fatalf("failed to parse JSON data: %v", err)
			}

			if diff := cmp.Diff(tt.expected.role, parsed.Role); diff != "" {
				t.Errorf("role mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.content, parsed.Content); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%s", diff)
			}

			if rec.flushed == 0 {
				t.Error("expected flush to be called")
			}
		})
	}
}

func TestSSEWriter_WriteDone(t *testing.T) {
	// 正常系
	t.Run("writes done event with complete status", func(t *testing.T) {
		rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
		sw, err := NewSSEWriter(rec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = sw.WriteDone()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := "event: done\ndata: {\"status\":\"complete\"}\n\n"
		if diff := cmp.Diff(want, rec.Body.String()); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestSSEWriter_WriteError(t *testing.T) {
	type args struct {
		err error
	}

	tests := []struct {
		testName     string
		args         args
		wantContains string
	}{
		// 正常系
		{
			testName:     "write standard error",
			args:         args{err: fmt.Errorf("session not found")},
			wantContains: "session not found",
		},
		// 特殊文字
		{
			testName:     "write error with special characters",
			args:         args{err: fmt.Errorf("error: <script>alert('xss')</script>")},
			wantContains: "alert", // JSON encoding escapes < and > as \u003c and \u003e
		},
		// 空文字
		{
			testName:     "write error with empty message",
			args:         args{err: fmt.Errorf("")},
			wantContains: `"error":""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
			sw, err := NewSSEWriter(rec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			err = sw.WriteError(tt.args.err)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			body := rec.Body.String()
			if !strings.HasPrefix(body, "event: error\ndata: ") {
				t.Errorf("expected SSE error event prefix, got: %s", body)
			}
			if !strings.Contains(body, tt.wantContains) {
				t.Errorf("expected body to contain %q, got: %s", tt.wantContains, body)
			}
		})
	}
}

// errorWriter is an http.ResponseWriter + http.Flusher that always returns an error on Write.
type errorWriter struct {
	header http.Header
}

func (e *errorWriter) Header() http.Header {
	if e.header == nil {
		e.header = make(http.Header)
	}
	return e.header
}
func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}
func (e *errorWriter) WriteHeader(_ int) {}
func (e *errorWriter) Flush()            {}

func TestSSEWriter_WriteEvent_Error(t *testing.T) {
	// 異常系 - write returns error
	t.Run("returns error when write fails", func(t *testing.T) {
		ew := &errorWriter{}
		sw, err := NewSSEWriter(ew)
		if err != nil {
			t.Fatalf("unexpected error creating SSEWriter: %v", err)
		}

		err = sw.WriteEvent("message", "test data")
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !strings.Contains(err.Error(), "simulated write error") {
			t.Errorf("expected simulated write error, got: %v", err)
		}
	})
}

func TestSSEWriter_WriteMessage_Error(t *testing.T) {
	// 異常系 - write returns error
	t.Run("returns error when write fails", func(t *testing.T) {
		ew := &errorWriter{}
		sw, err := NewSSEWriter(ew)
		if err != nil {
			t.Fatalf("unexpected error creating SSEWriter: %v", err)
		}

		err = sw.WriteMessage(messageResponse{ID: "id-1", Role: "user", Content: "test"})
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})
}

func TestSSEWriter_Ping_Error(t *testing.T) {
	// 異常系 - write returns error
	t.Run("returns error when write fails", func(t *testing.T) {
		ew := &errorWriter{}
		sw, err := NewSSEWriter(ew)
		if err != nil {
			t.Fatalf("unexpected error creating SSEWriter: %v", err)
		}

		err = sw.Ping()
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})
}

func TestSSEWriter_Ping(t *testing.T) {
	// 正常系
	t.Run("writes ping comment", func(t *testing.T) {
		rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
		sw, err := NewSSEWriter(rec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = sw.Ping()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := ":ping\n\n"
		if diff := cmp.Diff(want, rec.Body.String()); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}

		if rec.flushed == 0 {
			t.Error("expected flush to be called")
		}
	})

	// 境界値 - multiple pings
	t.Run("multiple pings accumulate", func(t *testing.T) {
		rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
		sw, err := NewSSEWriter(rec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := 0; i < 3; i++ {
			if err := sw.Ping(); err != nil {
				t.Fatalf("unexpected error on ping %d: %v", i, err)
			}
		}

		want := ":ping\n\n:ping\n\n:ping\n\n"
		if diff := cmp.Diff(want, rec.Body.String()); diff != "" {
			t.Errorf("output mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(3, rec.flushed); diff != "" {
			t.Errorf("flush count mismatch (-want +got):\n%s", diff)
		}
	})
}
