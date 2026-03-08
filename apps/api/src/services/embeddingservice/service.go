package embeddingservice

import (
	"context"
	"encoding/json"
	"fmt"

	"utils/embedding"
	db "utils/db/db"
	"utils/logger"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// EmbeddingService generates and stores embeddings for prompt versions.
type EmbeddingService struct {
	client *embedding.Client
	q      db.Querier
}

// NewEmbeddingService creates a new EmbeddingService.
// If client is nil, embedding generation is disabled (noop).
func NewEmbeddingService(client *embedding.Client, q db.Querier) *EmbeddingService {
	return &EmbeddingService{client: client, q: q}
}

// Available returns true if the embedding service is enabled.
func (s *EmbeddingService) Available() bool {
	return s.client != nil
}

// EmbedVersion generates an embedding for the given prompt version
// and stores it in the database. This runs asynchronously (fire-and-forget).
func (s *EmbeddingService) EmbedVersionAsync(versionID uuid.UUID, content json.RawMessage) {
	if s.client == nil {
		return
	}

	go func() {
		ctx := context.Background()
		if err := s.embedAndStore(ctx, versionID, content); err != nil {
			logger.Error("embedding generation failed", "version_id", versionID, "error", err)
		}
	}()
}

// EmbedVersion generates an embedding synchronously.
func (s *EmbeddingService) EmbedVersion(ctx context.Context, versionID uuid.UUID, content json.RawMessage) error {
	if s.client == nil {
		return nil
	}
	return s.embedAndStore(ctx, versionID, content)
}

// GenerateEmbedding generates an embedding for arbitrary text (used for search queries).
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if s.client == nil {
		return nil, fmt.Errorf("embedding service not available")
	}
	return s.client.EmbedOne(ctx, text)
}

func (s *EmbeddingService) embedAndStore(ctx context.Context, versionID uuid.UUID, content json.RawMessage) error {
	text := extractText(content)
	if text == "" {
		return nil
	}

	emb, err := s.client.EmbedOne(ctx, text)
	if err != nil {
		return fmt.Errorf("generate embedding: %w", err)
	}

	if err := s.q.UpdatePromptVersionEmbedding(ctx, db.UpdatePromptVersionEmbeddingParams{
		ID:        versionID,
		Embedding: emb,
	}); err != nil {
		return fmt.Errorf("store embedding: %w", err)
	}

	logger.Info("embedding generated", "version_id", versionID, "dimensions", len(emb))
	return nil
}

// extractText extracts the text content from the JSONB content field.
func extractText(raw json.RawMessage) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return string(raw)
	}
	for _, key := range []string{"content", "text", "body", "system", "user"} {
		if v, ok := obj[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return string(raw)
}

// ensure pq is imported for array serialization
var _ = pq.Array
