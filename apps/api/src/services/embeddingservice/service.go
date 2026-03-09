// Package embeddingservice generates and stores vector embeddings for
// prompt versions, enabling semantic search across the prompt library.
//
// When EMBEDDING_URL is not configured, all methods gracefully no-op.
package embeddingservice

import (
	"context"
	"encoding/json"
	"fmt"

	"api/src/domain/prompt"
	"api/src/services/contentutil"
	"utils/embedding"
	"utils/logger"

	"github.com/google/uuid"
)

// EmbeddingService generates and stores embeddings for prompt versions.
type EmbeddingService struct {
	client      *embedding.Client
	versionRepo prompt.VersionRepository
}

// NewEmbeddingService creates a new EmbeddingService.
// If client is nil, embedding generation is disabled (noop).
func NewEmbeddingService(client *embedding.Client, versionRepo prompt.VersionRepository) *EmbeddingService {
	return &EmbeddingService{client: client, versionRepo: versionRepo}
}

// Available returns true if the embedding service is enabled.
func (s *EmbeddingService) Available() bool {
	return s.client != nil
}

// EmbedVersionAsync generates an embedding for the given prompt version
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

// embedAndStore generates an embedding from content and persists it.
func (s *EmbeddingService) embedAndStore(ctx context.Context, versionID uuid.UUID, content json.RawMessage) error {
	text := contentutil.ExtractText(content)
	if text == "" {
		return nil
	}

	emb, err := s.client.EmbedOne(ctx, text)
	if err != nil {
		return fmt.Errorf("generate embedding: %w", err)
	}

	if err := s.versionRepo.UpdateEmbedding(ctx, prompt.PromptVersionIDFromUUID(versionID), emb); err != nil {
		return fmt.Errorf("store embedding: %w", err)
	}

	logger.Info("embedding generated", "version_id", versionID, "dimensions", len(emb))
	return nil
}
