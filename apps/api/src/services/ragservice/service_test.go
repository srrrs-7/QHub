package ragservice

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"api/src/services/embeddingservice"
	"api/src/services/intentservice"
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
			result, err := tt.svc.GenerateResponse(context.Background(), tt.args.sessionID, tt.args.userMsg, tt.args.orgID)
			if !tt.expected.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if result != nil {
				t.Error("expected nil result on error")
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
			got := BuildSystemPrompt(tt.args.items, nil)

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

		got := BuildSystemPrompt(items, nil)

		// Check that "### " headers appear 5 times
		headerCount := strings.Count(got, "### ")
		if diff := cmp.Diff(5, headerCount); diff != "" {
			t.Errorf("header count mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestExtractContentText_EdgeCases(t *testing.T) {
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
		// 正常系 - nested object with text field containing non-string value
		{
			testName: "object with text field as number falls back to raw",
			args:     args{content: json.RawMessage(`{"text":123}`)},
			expected: expected{text: `{"text":123}`},
		},
		// 正常系 - object with content field as number falls back to raw
		{
			testName: "object with content field as number falls back to raw",
			args:     args{content: json.RawMessage(`{"content":456}`)},
			expected: expected{text: `{"content":456}`},
		},
		// 正常系 - object with both text and content prefers text
		{
			testName: "object with both text and content returns text",
			args:     args{content: json.RawMessage(`{"text":"from text","content":"from content"}`)},
			expected: expected{text: "from text"},
		},
		// 異常系 - invalid JSON (not even a valid string or object)
		{
			testName: "invalid JSON returns raw string",
			args:     args{content: json.RawMessage(`{invalid}`)},
			expected: expected{text: `{invalid}`},
		},
		// 境界値 - JSON array falls back to raw
		{
			testName: "JSON array returns raw string",
			args:     args{content: json.RawMessage(`["a","b","c"]`)},
			expected: expected{text: `["a","b","c"]`},
		},
		// 境界値 - JSON number falls back to raw
		{
			testName: "JSON number returns raw string",
			args:     args{content: json.RawMessage(`42`)},
			expected: expected{text: `42`},
		},
		// 境界値 - JSON boolean falls back to raw
		{
			testName: "JSON boolean returns raw string",
			args:     args{content: json.RawMessage(`true`)},
			expected: expected{text: `true`},
		},
		// 境界値 - JSON null unmarshals to empty string
		{
			testName: "JSON null returns empty string",
			args:     args{content: json.RawMessage(`null`)},
			expected: expected{text: ""},
		},
		// 特殊文字 - deeply nested object
		{
			testName: "nested object without text or content returns raw",
			args:     args{content: json.RawMessage(`{"data":{"nested":"value"}}`)},
			expected: expected{text: `{"data":{"nested":"value"}}`},
		},
		// 特殊文字 - string with special JSON characters
		{
			testName: "string with escaped quotes",
			args:     args{content: json.RawMessage(`"line1\nline2\ttab"`)},
			expected: expected{text: "line1\nline2\ttab"},
		},
		// 特殊文字 - object with text containing unicode
		{
			testName: "object with text field containing unicode",
			args:     args{content: json.RawMessage(`{"text":"日本語テキスト 🎉"}`)},
			expected: expected{text: "日本語テキスト 🎉"},
		},
		// 境界値 - whitespace-only JSON string
		{
			testName: "whitespace-only JSON string",
			args:     args{content: json.RawMessage(`"   "`)},
			expected: expected{text: "   "},
		},
		// 境界値 - very long string
		{
			testName: "very long string content",
			args:     args{content: json.RawMessage(`"` + strings.Repeat("a", 10000) + `"`)},
			expected: expected{text: strings.Repeat("a", 10000)},
		},
		// 空文字 - object with empty text field
		{
			testName: "object with empty text field",
			args:     args{content: json.RawMessage(`{"text":""}`)},
			expected: expected{text: ""},
		},
		// 空文字 - object with empty content field
		{
			testName: "object with empty content field",
			args:     args{content: json.RawMessage(`{"content":""}`)},
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

func TestBuildSystemPrompt_EdgeCases(t *testing.T) {
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
		// 境界値 - zero similarity
		{
			testName: "item with zero similarity",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Zero Score",
						PromptSlug:    "zero-score",
						VersionNumber: 1,
						Content:       "Zero similarity content.",
						Similarity:    0.0,
					},
				},
			},
			expected: expected{
				contains: []string{"Zero Score", "0.00", "Zero similarity content."},
			},
		},
		// 境界値 - perfect similarity
		{
			testName: "item with perfect similarity",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Perfect Match",
						PromptSlug:    "perfect-match",
						VersionNumber: 5,
						Content:       "Exact match content.",
						Similarity:    1.0,
					},
				},
			},
			expected: expected{
				contains: []string{"Perfect Match", "1.00", "v5"},
			},
		},
		// 特殊文字 - content with SQL injection
		{
			testName: "item with SQL injection in content",
			args: args{
				items: []contextItem{
					{
						PromptName:    "SQL Test",
						PromptSlug:    "sql-test",
						VersionNumber: 1,
						Content:       "'; DROP TABLE prompts; --",
						Similarity:    0.5,
					},
				},
			},
			expected: expected{
				contains: []string{"SQL Test", "'; DROP TABLE prompts; --"},
			},
		},
		// 特殊文字 - content with newlines and special formatting
		{
			testName: "item with multiline content",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Multiline",
						PromptSlug:    "multiline",
						VersionNumber: 1,
						Content:       "Line 1\nLine 2\nLine 3",
						Similarity:    0.8,
					},
				},
			},
			expected: expected{
				contains: []string{"Multiline", "Line 1\nLine 2\nLine 3"},
			},
		},
		// 空文字 - empty content in item
		{
			testName: "item with empty content",
			args: args{
				items: []contextItem{
					{
						PromptName:    "Empty Content",
						PromptSlug:    "empty-content",
						VersionNumber: 1,
						Content:       "",
						Similarity:    0.6,
					},
				},
			},
			expected: expected{
				contains: []string{"Empty Content", "Relevant Prompt Context"},
			},
		},
		// 境界値 - large number of items
		{
			testName: "ten items all included",
			args: args{
				items: func() []contextItem {
					items := make([]contextItem, 10)
					for i := range items {
						items[i] = contextItem{
							PromptName:    "Prompt " + strings.Repeat("x", i),
							PromptSlug:    "prompt",
							VersionNumber: int32(i + 1),
							Content:       "content " + strings.Repeat("y", i),
							Similarity:    0.9,
						}
					}
					return items
				}(),
			},
			expected: expected{
				contains: []string{"Relevant Prompt Context", "### 1.", "### 10."},
			},
		},
		// 境界値 - version number 0
		{
			testName: "item with version number zero",
			args: args{
				items: []contextItem{
					{
						PromptName:    "V0 Prompt",
						PromptSlug:    "v0-prompt",
						VersionNumber: 0,
						Content:       "Version zero.",
						Similarity:    0.7,
					},
				},
			},
			expected: expected{
				contains: []string{"V0 Prompt", "v0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := BuildSystemPrompt(tt.args.items, nil)

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

func TestBuildSystemPrompt_WithIntent(t *testing.T) {
	items := []contextItem{
		{PromptName: "Test Prompt", PromptSlug: "test", VersionNumber: 1, Content: "Hello"},
	}

	tests := []struct {
		testName    string
		intent      *intentservice.Intent
		contains    []string
		notContains []string
	}{
		{
			testName:    "nil intent omits intent section",
			intent:      nil,
			notContains: []string{"User Intent"},
		},
		{
			testName:    "general intent omits intent section",
			intent:      &intentservice.Intent{Type: intentservice.IntentGeneral, Confidence: 0.5},
			notContains: []string{"User Intent"},
		},
		{
			testName: "improve intent adds intent section",
			intent:   &intentservice.Intent{Type: intentservice.IntentImprove, Confidence: 0.8},
			contains: []string{"User Intent: improve", "80%"},
		},
		{
			testName: "compliance intent adds intent section",
			intent:   &intentservice.Intent{Type: intentservice.IntentCompliance, Confidence: 0.8},
			contains: []string{"User Intent: compliance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := BuildSystemPrompt(items, tt.intent)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("prompt should contain %q, got:\n%s", want, got)
				}
			}
			for _, notwant := range tt.notContains {
				if strings.Contains(got, notwant) {
					t.Errorf("prompt should NOT contain %q, got:\n%s", notwant, got)
				}
			}
		})
	}
}

func TestRAGService_Available_EdgeCases(t *testing.T) {
	type expected struct {
		available bool
	}

	tests := []struct {
		testName string
		svc      *RAGService
		expected expected
	}{
		// 正常系 - verify available service with querier
		{
			testName: "available with all deps including querier",
			svc: &RAGService{
				embSvc:       newAvailableEmbSvc(),
				ollamaClient: ollama.NewClient("http://localhost:11434"),
				q:            nil, // querier doesn't affect availability
				model:        DefaultModel,
			},
			expected: expected{available: true},
		},
		// 境界値 - ollama with trailing slash URL
		{
			testName: "available with trailing slash in ollama URL",
			svc: &RAGService{
				embSvc:       newAvailableEmbSvc(),
				ollamaClient: ollama.NewClient("http://localhost:11434/"),
			},
			expected: expected{available: true},
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

func TestNewRAGService_EdgeCases(t *testing.T) {
	// 正常系 - with all deps
	t.Run("model is set to default", func(t *testing.T) {
		svc := NewRAGService(newAvailableEmbSvc(), ollama.NewClient("http://localhost:11434"), nil)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if diff := cmp.Diff(DefaultModel, svc.model); diff != "" {
			t.Errorf("model mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(true, svc.Available()); diff != "" {
			t.Errorf("Available mismatch (-want +got):\n%s", diff)
		}
	})

	// 境界値 - service stores all dependencies
	t.Run("stores embedding service reference", func(t *testing.T) {
		embSvc := newAvailableEmbSvc()
		svc := NewRAGService(embSvc, nil, nil)
		if svc.embSvc != embSvc {
			t.Error("expected embedding service to be stored")
		}
	})

	// 境界値 - service stores ollama client
	t.Run("stores ollama client reference", func(t *testing.T) {
		client := ollama.NewClient("http://localhost:11434")
		svc := NewRAGService(nil, client, nil)
		if svc.ollamaClient != client {
			t.Error("expected ollama client to be stored")
		}
	})
}

func TestGenerateResponse_UnavailableEdgeCases(t *testing.T) {
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
		// 異常系 - both services nil
		{
			testName: "error when both embedding and ollama nil",
			svc: &RAGService{
				embSvc:       nil,
				ollamaClient: nil,
			},
			args:     args{sessionID: uuid.New(), userMsg: "test query", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		// 異常系 - embedding available but ollama empty URL
		{
			testName: "error when ollama has empty URL",
			svc: &RAGService{
				embSvc:       newAvailableEmbSvc(),
				ollamaClient: ollama.NewClient(""),
			},
			args:     args{sessionID: uuid.New(), userMsg: "test query", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		// 空文字 - empty user message still fails if not available
		{
			testName: "error with empty message when not available",
			svc: &RAGService{
				embSvc:       nil,
				ollamaClient: ollama.NewClient("http://localhost:11434"),
			},
			args:     args{sessionID: uuid.New(), userMsg: "", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		// Null/Nil - zero-value UUID still fails if not available
		{
			testName: "error with zero UUID when not available",
			svc: &RAGService{
				embSvc:       newUnavailableEmbSvc(),
				ollamaClient: nil,
			},
			args:     args{sessionID: uuid.UUID{}, userMsg: "test", orgID: uuid.UUID{}},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		// 特殊文字 - Japanese message still fails if not available
		{
			testName: "error with Japanese message when not available",
			svc:      nil,
			args:     args{sessionID: uuid.New(), userMsg: "日本語のクエリ", orgID: uuid.New()},
			expected: expected{wantErr: true, errContain: "not available"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := tt.svc.GenerateResponse(context.Background(), tt.args.sessionID, tt.args.userMsg, tt.args.orgID)
			if !tt.expected.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if result != nil {
				t.Error("expected nil result on error")
			}
			if !strings.Contains(err.Error(), tt.expected.errContain) {
				t.Errorf("error %q should contain %q", err.Error(), tt.expected.errContain)
			}
		})
	}
}
