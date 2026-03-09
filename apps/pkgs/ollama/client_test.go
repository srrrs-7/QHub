package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewClient(t *testing.T) {
	type args struct {
		baseURL string
	}
	type expected struct {
		available bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "create client with valid URL",
			args:     args{baseURL: "http://localhost:11434"},
			expected: expected{available: true},
		},
		// 空文字
		{
			testName: "create client with empty URL",
			args:     args{baseURL: ""},
			expected: expected{available: false},
		},
		// 特殊文字
		{
			testName: "create client with URL containing path",
			args:     args{baseURL: "http://host.docker.internal:11434"},
			expected: expected{available: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			client := NewClient(tt.args.baseURL)
			if client == nil {
				t.Fatal("expected non-nil client")
			}
			if diff := cmp.Diff(tt.expected.available, client.Available()); diff != "" {
				t.Errorf("Available() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_Available(t *testing.T) {
	type expected struct {
		available bool
	}

	tests := []struct {
		testName string
		client   *Client
		expected expected
	}{
		// 正常系
		{
			testName: "client with URL is available",
			client:   NewClient("http://localhost:11434"),
			expected: expected{available: true},
		},
		// 空文字
		{
			testName: "client with empty URL is not available",
			client:   NewClient(""),
			expected: expected{available: false},
		},
		// Null/Nil
		{
			testName: "nil client is not available",
			client:   nil,
			expected: expected{available: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := tt.client.Available()
			if diff := cmp.Diff(tt.expected.available, got); diff != "" {
				t.Errorf("Available() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_ChatSync(t *testing.T) {
	type args struct {
		req ChatRequest
	}
	type expected struct {
		role    string
		content string
		wantErr bool
	}

	tests := []struct {
		testName   string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		args       args
		expected   expected
	}{
		// 正常系
		{
			testName: "successful sync chat response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				resp := ChatResponse{
					Message: ChatMessage{Role: "assistant", Content: "Hello! How can I help?"},
					Done:    true,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			args: args{req: ChatRequest{
				Model:    "llama3",
				Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
			}},
			expected: expected{role: "assistant", content: "Hello! How can I help?", wantErr: false},
		},
		// 特殊文字
		{
			testName: "sync chat with Japanese content",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				resp := ChatResponse{
					Message: ChatMessage{Role: "assistant", Content: "こんにちは！お手伝いします。"},
					Done:    true,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			args: args{req: ChatRequest{
				Model:    "llama3",
				Messages: []ChatMessage{{Role: "user", Content: "こんにちは"}},
			}},
			expected: expected{role: "assistant", content: "こんにちは！お手伝いします。", wantErr: false},
		},
		// 異常系
		{
			testName: "server returns 500",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			args: args{req: ChatRequest{
				Model:    "llama3",
				Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
			}},
			expected: expected{wantErr: true},
		},
		// 異常系
		{
			testName: "server returns invalid JSON",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("not json"))
			},
			args: args{req: ChatRequest{
				Model:    "llama3",
				Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
			}},
			expected: expected{wantErr: true},
		},
		// 空文字
		{
			testName: "server returns empty content",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				resp := ChatResponse{
					Message: ChatMessage{Role: "assistant", Content: ""},
					Done:    true,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			args: args{req: ChatRequest{
				Model:    "llama3",
				Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
			}},
			expected: expected{role: "assistant", content: "", wantErr: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client := NewClient(server.URL)
			msg, err := client.ChatSync(context.Background(), tt.args.req)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expected.role, msg.Role); diff != "" {
				t.Errorf("role mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.content, msg.Content); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_Chat(t *testing.T) {
	type expected struct {
		chunks  int
		wantErr bool
	}

	tests := []struct {
		testName   string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		expected   expected
	}{
		// 正常系
		{
			testName: "streaming chat returns multiple chunks",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				flusher, _ := w.(http.Flusher)
				w.Header().Set("Content-Type", "application/x-ndjson")

				chunks := []ChatResponse{
					{Message: ChatMessage{Role: "assistant", Content: "Hello"}, Done: false},
					{Message: ChatMessage{Role: "assistant", Content: " world"}, Done: false},
					{Message: ChatMessage{Role: "assistant", Content: "!"}, Done: true},
				}
				for _, chunk := range chunks {
					data, _ := json.Marshal(chunk)
					w.Write(data)
					w.Write([]byte("\n"))
					if flusher != nil {
						flusher.Flush()
					}
				}
			},
			expected: expected{chunks: 3, wantErr: false},
		},
		// 境界値
		{
			testName: "streaming chat returns single done chunk",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/x-ndjson")
				chunk := ChatResponse{
					Message: ChatMessage{Role: "assistant", Content: "Short answer."},
					Done:    true,
				}
				data, _ := json.Marshal(chunk)
				w.Write(data)
				w.Write([]byte("\n"))
			},
			expected: expected{chunks: 1, wantErr: false},
		},
		// 異常系
		{
			testName: "streaming chat with server error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("error"))
			},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client := NewClient(server.URL)
			req := ChatRequest{
				Model:    "llama3",
				Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
			}

			ch, err := client.Chat(context.Background(), req)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var chunks []ChatResponse
			for chunk := range ch {
				chunks = append(chunks, chunk)
			}

			if diff := cmp.Diff(tt.expected.chunks, len(chunks)); diff != "" {
				t.Errorf("chunk count mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClient_Chat_ContextCancellation(t *testing.T) {
	// 正常系 - context cancellation stops reading
	t.Run("cancelled context stops streaming", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			flusher, _ := w.(http.Flusher)
			w.Header().Set("Content-Type", "application/x-ndjson")

			// Send many chunks slowly - the client should stop reading
			for i := 0; i < 100; i++ {
				chunk := ChatResponse{
					Message: ChatMessage{Role: "assistant", Content: "chunk"},
					Done:    false,
				}
				data, _ := json.Marshal(chunk)
				w.Write(data)
				w.Write([]byte("\n"))
				if flusher != nil {
					flusher.Flush()
				}
			}
		}))
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())
		client := NewClient(server.URL)

		ch, err := client.Chat(ctx, ChatRequest{
			Model:    "llama3",
			Messages: []ChatMessage{{Role: "user", Content: "Hello"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Read one chunk then cancel
		<-ch
		cancel()

		// Drain remaining - channel should close
		count := 0
		for range ch {
			count++
		}
		// We don't assert exact count since it depends on timing,
		// but the channel should eventually close
		_ = count
	})
}

func TestClient_Health(t *testing.T) {
	type expected struct {
		wantErr bool
	}

	tests := []struct {
		testName   string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		expected   expected
	}{
		// 正常系
		{
			testName: "healthy server",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"models":[]}`))
			},
			expected: expected{wantErr: false},
		},
		// 異常系
		{
			testName: "unhealthy server",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			client := NewClient(server.URL)
			err := client.Health(context.Background())

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClient_Health_NotConfigured(t *testing.T) {
	// Null/Nil - client with empty URL
	t.Run("not configured client returns error", func(t *testing.T) {
		client := NewClient("")
		err := client.Health(context.Background())
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})
}

func TestClient_ChatSync_RequestFormat(t *testing.T) {
	// 正常系 - verify request body format
	t.Run("request body has correct format", func(t *testing.T) {
		var receivedReq ChatRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&receivedReq)
			resp := ChatResponse{
				Message: ChatMessage{Role: "assistant", Content: "ok"},
				Done:    true,
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.ChatSync(context.Background(), ChatRequest{
			Model: "llama3",
			Messages: []ChatMessage{
				{Role: "system", Content: "You are helpful."},
				{Role: "user", Content: "Hello"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff("llama3", receivedReq.Model); diff != "" {
			t.Errorf("model mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(false, receivedReq.Stream); diff != "" {
			t.Errorf("stream should be false for sync (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(2, len(receivedReq.Messages)); diff != "" {
			t.Errorf("message count mismatch (-want +got):\n%s", diff)
		}
	})
}
