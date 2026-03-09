package ragservice

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"api/src/services/embeddingservice"
	"utils/embedding"
	"utils/ollama"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// newAvailableEmbSvc creates an EmbeddingService backed by a real embedding.Client
// pointing at a dummy URL. This makes Available() return true.
func newAvailableEmbSvc() *embeddingservice.EmbeddingService {
	client := embedding.NewClient("http://fake-embedding:80")
	return embeddingservice.NewEmbeddingService(client, nil)
}

// newUnavailableEmbSvc creates an EmbeddingService with nil client (Available() == false).
func newUnavailableEmbSvc() *embeddingservice.EmbeddingService {
	return embeddingservice.NewEmbeddingService(nil, nil)
}

func TestRAGService_Available(t *testing.T) {
	type expected struct {
		available bool
	}

	tests := []struct {
		testName string
		svc      *RAGService
		expected expected
	}{
		// 正常系
		{
			testName: "available when both services configured",
			svc: &RAGService{
				embSvc:       newAvailableEmbSvc(),
				ollamaClient: ollama.NewClient("http://localhost:11434"),
			},
			expected: expected{available: true},
		},
		// 異常系 - nil embedding service
		{
			testName: "unavailable when embedding service is nil",
			svc: &RAGService{
				embSvc:       nil,
				ollamaClient: ollama.NewClient("http://localhost:11434"),
			},
			expected: expected{available: false},
		},
		// 異常系 - embedding service not available (nil client)
		{
			testName: "unavailable when embedding client is nil",
			svc: &RAGService{
				embSvc:       newUnavailableEmbSvc(),
				ollamaClient: ollama.NewClient("http://localhost:11434"),
			},
			expected: expected{available: false},
		},
		// 異常系 - nil ollama client
		{
			testName: "unavailable when ollama client is nil",
			svc: &RAGService{
				embSvc:       newAvailableEmbSvc(),
				ollamaClient: nil,
			},
			expected: expected{available: false},
		},
		// 異常系 - ollama client not configured
		{
			testName: "unavailable when ollama URL is empty",
			svc: &RAGService{
				embSvc:       newAvailableEmbSvc(),
				ollamaClient: ollama.NewClient(""),
			},
			expected: expected{available: false},
		},
		// Null/Nil
		{
			testName: "unavailable when service is nil",
			svc:      nil,
			expected: expected{available: false},
		},
		// 境界値 - both nil
		{
			testName: "unavailable when both services nil",
			svc: &RAGService{
				embSvc:       nil,
				ollamaClient: nil,
			},
			expected: expected{available: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := tt.svc.Available()
			if diff := cmp.Diff(tt.expected.available, got); diff != "" {
				t.Errorf("Available() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRAGService_GenerateResponse_Unavailable(t *testing.T) {
	type args struct {
		sessionID uuid.UUID
		userMsg   string
		orgID     uuid.UUID
	}
	type expected struct {
		wantErr    bool
		errContain string
	}

	tests := []struct {
		testName string
		svc      *RAGService
		args     args
		expected expected
	}{
		// 異常系 - nil embedding service
		{
			testName: "error when embedding service nil",
			svc: &RAGService{
				embSvc:       nil,
				ollamaClient: ollama.NewClient("http://localhost:11434"),
			},
			args:     args{sessionID: uuid.New(), userMsg: "test", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		// 異常系 - nil ollama
		{
			testName: "error when ollama client nil",
			svc: &RAGService{
				embSvc:       newUnavailableEmbSvc(),
				ollamaClient: nil,
			},
			args:     args{sessionID: uuid.New(), userMsg: "test", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		// 空文字
		{
			testName: "error when service nil",
			svc:      nil,
			args:     args{sessionID: uuid.New(), userMsg: "", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ch, err := tt.svc.GenerateResponse(context.Background(), tt.args.sessionID, tt.args.userMsg, tt.args.orgID)
			if !tt.expected.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if ch != nil {
				t.Error("expected nil channel on error")
			}
			if !strings.Contains(err.Error(), tt.expected.errContain) {
				t.Errorf("error %q should contain %q", err.Error(), tt.expected.errContain)
			}
		})
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	type args struct {
		items []contextItem
	}
	type expected struct {
		contains    []string
		notContains []string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "prompt with context items",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Customer Support",
						PromptSlug:    "customer-support",
						VersionNumber: 3,
						Content:       "You are a customer support agent.",
						Similarity:    0.92,
					},
					{
						PromptName:    "Sales Follow-up",
						PromptSlug:    "sales-followup",
						VersionNumber: 1,
						Content:       "Write a follow-up email.",
						Similarity:    0.78,
					},
				},
			},
			expected: expected{
				contains: []string{
					"QHub",
					"Relevant Prompt Context",
					"Customer Support",
					"v3",
					"0.92",
					"customer support agent",
					"Sales Follow-up",
					"v1",
					"0.78",
					"follow-up email",
				},
			},
		},
		// 境界値 - empty items
		{
			testName: "prompt with no context items",
			args:     args{items: []contextItem{}},
			expected: expected{
				contains:    []string{"QHub", "No relevant prompt examples"},
				notContains: []string{"Relevant Prompt Context"},
			},
		},
		// Null/Nil
		{
			testName: "prompt with nil items",
			args:     args{items: nil},
			expected: expected{
				contains:    []string{"No relevant prompt examples"},
				notContains: []string{"Relevant Prompt Context"},
			},
		},
		// 境界値 - single item
		{
			testName: "prompt with single item",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Only Prompt",
						PromptSlug:    "only-prompt",
						VersionNumber: 1,
						Content:       "Single content.",
						Similarity:    0.95,
					},
				},
			},
			expected: expected{
				contains: []string{"1. Only Prompt", "v1", "0.95", "Single content."},
			},
		},
		// 特殊文字
		{
			testName: "prompt with Japanese content",
			args: args{
				items: []contextItem{
					{
						PromptName:    "日本語プロンプト",
						PromptSlug:    "japanese-prompt",
						VersionNumber: 2,
						Content:       "あなたは日本語のアシスタントです。",
						Similarity:    0.85,
					},
				},
			},
			expected: expected{
				contains: []string{"日本語プロンプト", "あなたは日本語のアシスタントです。"},
			},
		},
		// 特殊文字 - emoji
		{
			testName: "prompt with emoji content",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Fun Prompt",
						PromptSlug:    "fun-prompt",
						VersionNumber: 1,
						Content:       "Be fun and engaging!",
						Similarity:    0.75,
					},
				},
			},
			expected: expected{
				contains: []string{"Fun Prompt", "Be fun and engaging!"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := BuildSystemPrompt(tt.args.items)

			for _, want := range tt.expected.contains {
				if !strings.Contains(got, want) {
					t.Errorf("prompt should contain %q, got:\n%s", want, got)
				}
			}
			for _, notwant := range tt.expected.notContains {
				if strings.Contains(got, notwant) {
					t.Errorf("prompt should NOT contain %q, got:\n%s", notwant, got)
				}
			}
		})
	}
}

func TestExtractContentText(t *testing.T) {
	type args struct {
		content json.RawMessage
	}
	type expected struct {
		text string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - JSON string
		{
			testName: "plain JSON string",
			args:     args{content: json.RawMessage(`"Hello world"`)},
			expected: expected{text: "Hello world"},
		},
		// 正常系 - object with text field
		{
			testName: "object with text field",
			args:     args{content: json.RawMessage(`{"text":"Some prompt text"}`)},
			expected: expected{text: "Some prompt text"},
		},
		// 正常系 - object with content field
		{
			testName: "object with content field",
			args:     args{content: json.RawMessage(`{"content":"Some content"}`)},
			expected: expected{text: "Some content"},
		},
		// 異常系 - arbitrary JSON falls back to raw
		{
			testName: "object without text or content returns raw JSON",
			args:     args{content: json.RawMessage(`{"foo":"bar"}`)},
			expected: expected{text: `{"foo":"bar"}`},
		},
		// 空文字
		{
			testName: "empty content returns empty string",
			args:     args{content: json.RawMessage{}},
			expected: expected{text: ""},
		},
		// Null/Nil
		{
			testName: "nil content returns empty string",
			args:     args{content: nil},
			expected: expected{text: ""},
		},
		// 特殊文字
		{
			testName: "Japanese string content",
			args:     args{content: json.RawMessage(`"日本語コンテンツ"`)},
			expected: expected{text: "日本語コンテンツ"},
		},
		// 特殊文字
		{
			testName: "emoji string content",
			args:     args{content: json.RawMessage(`"Hello 🌍"`)},
			expected: expected{text: "Hello 🌍"},
		},
		// 境界値
		{
			testName: "empty JSON string",
			args:     args{content: json.RawMessage(`""`)},
			expected: expected{text: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := extractContentText(tt.args.content)
			if diff := cmp.Diff(tt.expected.text, got); diff != "" {
				t.Errorf("text mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewRAGService(t *testing.T) {
	// 正常系
	t.Run("creates service with all dependencies", func(t *testing.T) {
		ollamaClient := ollama.NewClient("http://localhost:11434")
		embSvc := newUnavailableEmbSvc()

		svc := NewRAGService(embSvc, ollamaClient, nil)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if diff := cmp.Diff(DefaultModel, svc.model); diff != "" {
			t.Errorf("model mismatch (-want +got):\n%s", diff)
		}
	})

	// Null/Nil - nil dependencies
	t.Run("creates service with nil dependencies", func(t *testing.T) {
		svc := NewRAGService(nil, nil, nil)
		if svc == nil {
			t.Fatal("expected non-nil service even with nil deps")
		}
		if svc.Available() {
			t.Error("service with nil deps should not be available")
		}
	})
}

func TestBuildSystemPrompt_NumberingFormat(t *testing.T) {
	// 境界値 - verify numbering with multiple items
	t.Run("items are numbered sequentially", func(t *testing.T) {
		items := make([]contextItem, 5)
		for i := range items {
			items[i] = contextItem{
				PromptName:    "Prompt",
				PromptSlug:    "prompt",
				VersionNumber: int32(i + 1),
				Content:       "content",
				Similarity:    0.9,
			}
		}

		got := BuildSystemPrompt(items)

		// Check that "### " headers appear 5 times
		headerCount := strings.Count(got, "### ")
		if diff := cmp.Diff(5, headerCount); diff != "" {
			t.Errorf("header count mismatch (-want +got):\n%s", diff)
		}
	})
}
