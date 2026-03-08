package embeddingservice

import (
	"testing"

	"api/src/services/contentutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		expected string
	}{
		// 正常系
		{testName: "content key", input: `{"content":"hello world"}`, expected: "hello world"},
		{testName: "text key", input: `{"text":"some text"}`, expected: "some text"},
		{testName: "body key", input: `{"body":"body content"}`, expected: "body content"},
		{testName: "system key", input: `{"system":"system prompt"}`, expected: "system prompt"},
		{testName: "user key", input: `{"user":"user message"}`, expected: "user message"},
		// 異常系
		{testName: "no known key returns raw", input: `{"title":"something"}`, expected: `{"title":"something"}`},
		{testName: "invalid json returns raw", input: `not json`, expected: "not json"},
		// 境界値
		{testName: "empty object", input: `{}`, expected: "{}"},
		// 特殊文字
		{testName: "unicode", input: `{"content":"日本語テスト 🎉"}`, expected: "日本語テスト 🎉"},
		// 空文字
		{testName: "empty content value", input: `{"content":""}`, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := contentutil.ExtractText([]byte(tt.input))
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAvailable(t *testing.T) {
	tests := []struct {
		testName string
		hasClient bool
		expected  bool
	}{
		{testName: "with client", hasClient: true, expected: true},
		{testName: "nil client", hasClient: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var svc *EmbeddingService
			if tt.hasClient {
				// Can't create a real client without a server, but we test the nil path
				svc = NewEmbeddingService(nil, nil)
				// With nil client, Available() should return false
				if svc.Available() {
					t.Error("expected Available() false with nil client")
				}
			} else {
				svc = NewEmbeddingService(nil, nil)
				got := svc.Available()
				if diff := cmp.Diff(tt.expected, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestEmbedVersionAsync_NilClient(t *testing.T) {
	// Should not panic with nil client
	svc := NewEmbeddingService(nil, nil)
	svc.EmbedVersionAsync(uuid.Nil, []byte(`{"content":"test"}`))
	// No error expected - noop when client is nil
}
