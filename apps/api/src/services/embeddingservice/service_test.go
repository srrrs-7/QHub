package embeddingservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"api/src/domain/prompt"
	"api/src/services/contentutil"
	"utils/embedding"
	"utils/logger"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}

// --- Mock VersionRepository ---

type mockVersionRepo struct {
	updateEmbeddingFn func(ctx context.Context, id prompt.PromptVersionID, emb []float32) error
	mu                sync.Mutex
	calls             []updateEmbeddingCall
}

type updateEmbeddingCall struct {
	ID        prompt.PromptVersionID
	Embedding []float32
}

func (m *mockVersionRepo) FindByPromptAndNumber(_ context.Context, _ prompt.PromptID, _ int) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) FindAllByPrompt(_ context.Context, _ prompt.PromptID) ([]prompt.PromptVersion, error) {
	return nil, nil
}
func (m *mockVersionRepo) FindLatest(_ context.Context, _ prompt.PromptID) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) FindProduction(_ context.Context, _ prompt.PromptID) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) Create(_ context.Context, _ prompt.VersionCmd, _ int) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) UpdateStatus(_ context.Context, _ prompt.PromptVersionID, _ prompt.VersionStatus) (prompt.PromptVersion, error) {
	return prompt.PromptVersion{}, nil
}
func (m *mockVersionRepo) ArchiveProduction(_ context.Context, _ prompt.PromptID) error {
	return nil
}
func (m *mockVersionRepo) UpdateLintResult(_ context.Context, _ prompt.PromptVersionID, _ json.RawMessage) error {
	return nil
}
func (m *mockVersionRepo) UpdateSemanticDiff(_ context.Context, _ prompt.PromptVersionID, _ json.RawMessage) error {
	return nil
}
func (m *mockVersionRepo) UpdateEmbedding(ctx context.Context, id prompt.PromptVersionID, emb []float32) error {
	m.mu.Lock()
	m.calls = append(m.calls, updateEmbeddingCall{ID: id, Embedding: emb})
	m.mu.Unlock()
	if m.updateEmbeddingFn != nil {
		return m.updateEmbeddingFn(ctx, id, emb)
	}
	return nil
}

var _ prompt.VersionRepository = (*mockVersionRepo)(nil)

// --- Helper: fake TEI server ---

func newFakeTEIServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *embedding.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := embedding.NewClient(srv.URL)
	return srv, client
}

func fakeTEIHandler(embeddings [][]float32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(embeddings)
	}
}

func fakeTEIErrorHandler(statusCode int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}
}

// =============================================================================
// TestNewEmbeddingService
// =============================================================================

func TestNewEmbeddingService(t *testing.T) {
	tests := []struct {
		testName      string
		clientNil     bool
		repoNil       bool
		wantAvailable bool
	}{
		// 正常系
		{testName: "with client and repo", clientNil: false, repoNil: false, wantAvailable: true},
		{testName: "with client nil repo", clientNil: false, repoNil: true, wantAvailable: true},
		// 異常系
		{testName: "nil client with repo", clientNil: true, repoNil: false, wantAvailable: false},
		// Null/Nil
		{testName: "both nil", clientNil: true, repoNil: true, wantAvailable: false},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var client *embedding.Client
			if !tt.clientNil {
				_, client = newFakeTEIServer(t, fakeTEIHandler(nil))
			}
			var repo prompt.VersionRepository
			if !tt.repoNil {
				repo = &mockVersionRepo{}
			}

			svc := NewEmbeddingService(client, repo)

			if svc == nil {
				t.Fatal("expected non-nil service")
			}
			if diff := cmp.Diff(tt.wantAvailable, svc.Available()); diff != "" {
				t.Errorf("Available() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// =============================================================================
// TestAvailable
// =============================================================================

func TestAvailable(t *testing.T) {
	tests := []struct {
		testName  string
		hasClient bool
		expected  bool
	}{
		// 正常系
		{testName: "with non-nil client returns true", hasClient: true, expected: true},
		// Null/Nil
		{testName: "with nil client returns false", hasClient: false, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var client *embedding.Client
			if tt.hasClient {
				_, client = newFakeTEIServer(t, fakeTEIHandler(nil))
			}
			svc := NewEmbeddingService(client, nil)
			got := svc.Available()
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("Available() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// =============================================================================
// TestGenerateEmbedding
// =============================================================================

func TestGenerateEmbedding(t *testing.T) {
	type args struct {
		text      string
		clientNil bool
	}
	type expected struct {
		wantErr    bool
		errContain string
		embedding  []float32
	}

	tests := []struct {
		testName    string
		teiHandler  http.HandlerFunc
		args        args
		expected    expected
	}{
		// 正常系
		{
			testName:   "successful embedding generation",
			teiHandler: fakeTEIHandler([][]float32{{0.1, 0.2, 0.3}}),
			args:       args{text: "hello world"},
			expected:   expected{wantErr: false, embedding: []float32{0.1, 0.2, 0.3}},
		},
		// 異常系
		{
			testName: "nil client returns error",
			args:     args{text: "hello", clientNil: true},
			expected: expected{wantErr: true, errContain: "not available"},
		},
		{
			testName:   "TEI server error returns error",
			teiHandler: fakeTEIErrorHandler(http.StatusInternalServerError, "internal error"),
			args:       args{text: "hello"},
			expected:   expected{wantErr: true, errContain: "500"},
		},
		// 特殊文字
		{
			testName:   "unicode text",
			teiHandler: fakeTEIHandler([][]float32{{0.4, 0.5}}),
			args:       args{text: "日本語テスト 🎉"},
			expected:   expected{wantErr: false, embedding: []float32{0.4, 0.5}},
		},
		{
			testName:   "emoji text",
			teiHandler: fakeTEIHandler([][]float32{{0.6, 0.7}}),
			args:       args{text: "Hello 🌍🚀"},
			expected:   expected{wantErr: false, embedding: []float32{0.6, 0.7}},
		},
		{
			testName:   "SQL injection text",
			teiHandler: fakeTEIHandler([][]float32{{0.8, 0.9}}),
			args:       args{text: "'; DROP TABLE prompts; --"},
			expected:   expected{wantErr: false, embedding: []float32{0.8, 0.9}},
		},
		// 空文字
		{
			testName:   "empty text still calls TEI",
			teiHandler: fakeTEIHandler([][]float32{{0.0}}),
			args:       args{text: ""},
			expected:   expected{wantErr: false, embedding: []float32{0.0}},
		},
		// 境界値
		{
			testName:   "very long text",
			teiHandler: fakeTEIHandler([][]float32{{1.0, 2.0}}),
			args:       args{text: strings.Repeat("a", 10000)},
			expected:   expected{wantErr: false, embedding: []float32{1.0, 2.0}},
		},
		{
			testName:   "single character",
			teiHandler: fakeTEIHandler([][]float32{{0.1}}),
			args:       args{text: "a"},
			expected:   expected{wantErr: false, embedding: []float32{0.1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var client *embedding.Client
			if !tt.args.clientNil {
				_, client = newFakeTEIServer(t, tt.teiHandler)
			}
			svc := NewEmbeddingService(client, nil)

			got, err := svc.GenerateEmbedding(context.Background(), tt.args.text)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.expected.errContain != "" && !strings.Contains(err.Error(), tt.expected.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.expected.errContain)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.expected.embedding, got); diff != "" {
					t.Errorf("embedding mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// =============================================================================
// TestEmbedVersion
// =============================================================================

func TestEmbedVersion(t *testing.T) {
	type args struct {
		clientNil bool
		content   string
		versionID uuid.UUID
	}
	type expected struct {
		wantErr        bool
		errContain     string
		embeddingCalls int
	}

	fixedID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		testName   string
		teiHandler http.HandlerFunc
		repoErr    error
		args       args
		expected   expected
	}{
		// 正常系
		{
			testName:   "successful embed and store",
			teiHandler: fakeTEIHandler([][]float32{{0.1, 0.2, 0.3}}),
			args:       args{content: `{"content":"hello world"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		// Null/Nil
		{
			testName: "nil client returns nil (noop)",
			args:     args{clientNil: true, content: `{"content":"hello"}`, versionID: fixedID},
			expected: expected{wantErr: false, embeddingCalls: 0},
		},
		// 空文字
		{
			testName:   "empty content value skips embedding",
			teiHandler: fakeTEIHandler([][]float32{{0.1}}),
			args:       args{content: `{"content":""}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 0},
		},
		// 異常系
		{
			testName:   "TEI server error",
			teiHandler: fakeTEIErrorHandler(http.StatusInternalServerError, "server error"),
			args:       args{content: `{"content":"hello"}`, versionID: fixedID},
			expected:   expected{wantErr: true, errContain: "generate embedding"},
		},
		{
			testName:   "repo UpdateEmbedding error",
			teiHandler: fakeTEIHandler([][]float32{{0.1, 0.2}}),
			repoErr:    fmt.Errorf("database connection lost"),
			args:       args{content: `{"content":"hello"}`, versionID: fixedID},
			expected:   expected{wantErr: true, errContain: "store embedding", embeddingCalls: 1},
		},
		// 特殊文字
		{
			testName:   "unicode content",
			teiHandler: fakeTEIHandler([][]float32{{0.5, 0.6}}),
			args:       args{content: `{"content":"日本語テスト 🎉"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "SQL injection in content",
			teiHandler: fakeTEIHandler([][]float32{{0.7}}),
			args:       args{content: `{"content":"'; DROP TABLE versions; --"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		// 境界値
		{
			testName:   "nil UUID version ID",
			teiHandler: fakeTEIHandler([][]float32{{0.1}}),
			args:       args{content: `{"content":"test"}`, versionID: uuid.Nil},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "plain text content (not JSON)",
			teiHandler: fakeTEIHandler([][]float32{{0.2, 0.3}}),
			args:       args{content: `just plain text`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "content with text key",
			teiHandler: fakeTEIHandler([][]float32{{0.4}}),
			args:       args{content: `{"text":"some text here"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "content with body key",
			teiHandler: fakeTEIHandler([][]float32{{0.5}}),
			args:       args{content: `{"body":"body content"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "content with system key",
			teiHandler: fakeTEIHandler([][]float32{{0.6}}),
			args:       args{content: `{"system":"system prompt"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "content with user key",
			teiHandler: fakeTEIHandler([][]float32{{0.7}}),
			args:       args{content: `{"user":"user message"}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "empty JSON object falls through to raw",
			teiHandler: fakeTEIHandler([][]float32{{0.8}}),
			args:       args{content: `{}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
		{
			testName:   "nested JSON no known key uses raw",
			teiHandler: fakeTEIHandler([][]float32{{0.9}}),
			args:       args{content: `{"metadata":{"key":"value"}}`, versionID: fixedID},
			expected:   expected{wantErr: false, embeddingCalls: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var client *embedding.Client
			if !tt.args.clientNil {
				_, client = newFakeTEIServer(t, tt.teiHandler)
			}

			repo := &mockVersionRepo{}
			if tt.repoErr != nil {
				repo.updateEmbeddingFn = func(_ context.Context, _ prompt.PromptVersionID, _ []float32) error {
					return tt.repoErr
				}
			}

			svc := NewEmbeddingService(client, repo)
			err := svc.EmbedVersion(context.Background(), tt.args.versionID, json.RawMessage(tt.args.content))

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tt.expected.errContain != "" && !strings.Contains(err.Error(), tt.expected.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.expected.errContain)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			repo.mu.Lock()
			gotCalls := len(repo.calls)
			repo.mu.Unlock()
			if diff := cmp.Diff(tt.expected.embeddingCalls, gotCalls); diff != "" {
				t.Errorf("UpdateEmbedding call count mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// =============================================================================
// TestEmbedVersionAsync
// =============================================================================

func TestEmbedVersionAsync(t *testing.T) {
	fixedID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	t.Run("nil client returns immediately (noop)", func(t *testing.T) {
		svc := NewEmbeddingService(nil, nil)
		// Should not panic
		svc.EmbedVersionAsync(fixedID, json.RawMessage(`{"content":"test"}`))
		// No goroutine spawned, returns immediately
	})

	t.Run("nil client with nil UUID", func(t *testing.T) {
		svc := NewEmbeddingService(nil, nil)
		svc.EmbedVersionAsync(uuid.Nil, nil)
	})

	t.Run("nil client with empty content", func(t *testing.T) {
		svc := NewEmbeddingService(nil, nil)
		svc.EmbedVersionAsync(fixedID, json.RawMessage(``))
	})

	t.Run("with client embeds asynchronously", func(t *testing.T) {
		_, client := newFakeTEIServer(t, fakeTEIHandler([][]float32{{0.1, 0.2, 0.3}}))
		repo := &mockVersionRepo{}
		svc := NewEmbeddingService(client, repo)

		svc.EmbedVersionAsync(fixedID, json.RawMessage(`{"content":"hello async"}`))

		// Wait for goroutine to complete
		deadline := time.After(3 * time.Second)
		for {
			repo.mu.Lock()
			n := len(repo.calls)
			repo.mu.Unlock()
			if n > 0 {
				break
			}
			select {
			case <-deadline:
				t.Fatal("timed out waiting for async embedding")
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}

		repo.mu.Lock()
		if diff := cmp.Diff(1, len(repo.calls)); diff != "" {
			t.Errorf("call count mismatch (-want +got):\n%s", diff)
		}
		repo.mu.Unlock()
	})

	t.Run("with client and empty content does not call repo", func(t *testing.T) {
		_, client := newFakeTEIServer(t, fakeTEIHandler([][]float32{{0.1}}))
		repo := &mockVersionRepo{}
		svc := NewEmbeddingService(client, repo)

		svc.EmbedVersionAsync(fixedID, json.RawMessage(`{"content":""}`))

		// Give the goroutine time to run
		time.Sleep(100 * time.Millisecond)

		repo.mu.Lock()
		if diff := cmp.Diff(0, len(repo.calls)); diff != "" {
			t.Errorf("call count mismatch (-want +got):\n%s", diff)
		}
		repo.mu.Unlock()
	})

	t.Run("with client TEI error logs but does not panic", func(t *testing.T) {
		_, client := newFakeTEIServer(t, fakeTEIErrorHandler(http.StatusInternalServerError, "fail"))
		repo := &mockVersionRepo{}
		svc := NewEmbeddingService(client, repo)

		svc.EmbedVersionAsync(fixedID, json.RawMessage(`{"content":"hello"}`))

		// Give the goroutine time to run and log error
		time.Sleep(100 * time.Millisecond)

		repo.mu.Lock()
		if diff := cmp.Diff(0, len(repo.calls)); diff != "" {
			t.Errorf("call count mismatch (-want +got):\n%s", diff)
		}
		repo.mu.Unlock()
	})
}

// =============================================================================
// TestEmbedVersion_ContextCancelled
// =============================================================================

func TestEmbedVersion_ContextCancelled(t *testing.T) {
	// The TEI server delays, but the context is already cancelled
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([][]float32{{0.1}})
	}

	_, client := newFakeTEIServer(t, handler)
	repo := &mockVersionRepo{}
	svc := NewEmbeddingService(client, repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := svc.EmbedVersion(ctx, uuid.New(), json.RawMessage(`{"content":"hello"}`))
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !strings.Contains(err.Error(), "generate embedding") {
		t.Errorf("error %q should contain 'generate embedding'", err.Error())
	}
}

// =============================================================================
// TestExtractText (comprehensive, extending existing coverage)
// =============================================================================

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
		{testName: "content takes priority over text", input: `{"content":"first","text":"second"}`, expected: "first"},
		{testName: "text takes priority over body", input: `{"text":"first","body":"second"}`, expected: "first"},
		// 異常系
		{testName: "no known key returns raw", input: `{"title":"something"}`, expected: `{"title":"something"}`},
		{testName: "invalid json returns raw", input: `not json`, expected: "not json"},
		{testName: "non-string content value returns raw", input: `{"content":123}`, expected: `{"content":123}`},
		{testName: "null content value returns raw", input: `{"content":null}`, expected: `{"content":null}`},
		{testName: "array content value returns raw", input: `{"content":["a","b"]}`, expected: `{"content":["a","b"]}`},
		// 境界値
		{testName: "empty object", input: `{}`, expected: "{}"},
		{testName: "very long content", input: fmt.Sprintf(`{"content":"%s"}`, strings.Repeat("x", 5000)), expected: strings.Repeat("x", 5000)},
		{testName: "single char content", input: `{"content":"a"}`, expected: "a"},
		// 特殊文字
		{testName: "unicode content", input: `{"content":"日本語テスト 🎉"}`, expected: "日本語テスト 🎉"},
		{testName: "emoji only", input: `{"content":"🚀🌍🎉"}`, expected: "🚀🌍🎉"},
		{testName: "sql injection", input: `{"content":"'; DROP TABLE x; --"}`, expected: "'; DROP TABLE x; --"},
		{testName: "html tags", input: `{"content":"<script>alert(1)</script>"}`, expected: "<script>alert(1)</script>"},
		{testName: "newlines and tabs", input: `{"content":"line1\nline2\ttab"}`, expected: "line1\nline2\ttab"},
		// 空文字
		{testName: "empty content value", input: `{"content":""}`, expected: ""},
		{testName: "whitespace content", input: `{"content":"   "}`, expected: "   "},
		// Null/Nil
		{testName: "json array", input: `["a","b"]`, expected: `["a","b"]`},
		{testName: "json number", input: `42`, expected: `42`},
		{testName: "json null", input: `null`, expected: `null`},
		{testName: "nested object no known key", input: `{"meta":{"nested":"val"}}`, expected: `{"meta":{"nested":"val"}}`},
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

// =============================================================================
// TestEmbedVersion_UpdateEmbeddingVerifiesData
// =============================================================================

func TestEmbedVersion_UpdateEmbeddingVerifiesData(t *testing.T) {
	expectedEmb := []float32{0.11, 0.22, 0.33, 0.44}
	versionID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	_, client := newFakeTEIServer(t, fakeTEIHandler([][]float32{expectedEmb}))
	repo := &mockVersionRepo{}
	svc := NewEmbeddingService(client, repo)

	err := svc.EmbedVersion(context.Background(), versionID, json.RawMessage(`{"content":"verify data"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(repo.calls))
	}

	call := repo.calls[0]
	expectedID := prompt.PromptVersionIDFromUUID(versionID)
	if diff := cmp.Diff(expectedID.UUID(), call.ID.UUID()); diff != "" {
		t.Errorf("version ID mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedEmb, call.Embedding); diff != "" {
		t.Errorf("embedding mismatch (-want +got):\n%s", diff)
	}
}

// =============================================================================
// TestGenerateEmbedding_NilClient
// =============================================================================

func TestGenerateEmbedding_NilClient(t *testing.T) {
	svc := NewEmbeddingService(nil, nil)

	got, err := svc.GenerateEmbedding(context.Background(), "some text")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("error %q should contain 'not available'", err.Error())
	}
	if got != nil {
		t.Errorf("expected nil embedding, got %v", got)
	}
}
