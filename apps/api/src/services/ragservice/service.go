// Package ragservice implements a Retrieval-Augmented Generation (RAG) pipeline
// for the consulting chat feature. It combines semantic search over prompt
// versions with LLM generation via Ollama to produce context-aware responses.
//
// When either the embedding service or Ollama client is not configured,
// the service reports itself as unavailable via the Available method.
package ragservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"api/src/services/embeddingservice"
	"api/src/services/intentservice"
	"utils/db/db"
	"utils/logger"
	"utils/ollama"

	"github.com/google/uuid"
)

const (
	// DefaultTopK is the default number of search results to retrieve.
	DefaultTopK = 5
	// DefaultMinScore is the minimum similarity score to include a result.
	DefaultMinScore = 0.5
	// DefaultModel is the default Ollama model to use for generation.
	DefaultModel = "llama3"
)

// RAGService orchestrates the RAG pipeline: embed query, search, build context, generate.
type RAGService struct {
	embSvc       *embeddingservice.EmbeddingService
	ollamaClient *ollama.Client
	q            db.Querier
	model        string
}

// NewRAGService creates a new RAGService.
// If embSvc is nil or ollamaClient is nil/unconfigured, the service will be unavailable.
func NewRAGService(embSvc *embeddingservice.EmbeddingService, ollamaClient *ollama.Client, q db.Querier) *RAGService {
	return &RAGService{
		embSvc:       embSvc,
		ollamaClient: ollamaClient,
		q:            q,
		model:        DefaultModel,
	}
}

// Available returns true if both the embedding service and Ollama client are configured.
func (s *RAGService) Available() bool {
	if s == nil {
		return false
	}
	if s.embSvc == nil || !s.embSvc.Available() {
		return false
	}
	if s.ollamaClient == nil || !s.ollamaClient.Available() {
		return false
	}
	return true
}

// contextItem represents a single search result used as context for generation.
type contextItem struct {
	PromptID      uuid.UUID
	PromptName    string
	PromptSlug    string
	VersionNumber int32
	Content       string
	Similarity    float64
}

// RAGResult holds the streaming response channel and the context items
// used for generation. After the stream is fully consumed, callers should
// call ExtractCitationsFromResponse to determine which context items were referenced.
type RAGResult struct {
	// Chunks is the channel of streamed response text chunks.
	Chunks <-chan string
	// contextItems are the search results provided as context to the LLM.
	contextItems []contextItem
	// promptIDs maps prompt slugs to their UUIDs for citation output.
	promptIDs map[string]uuid.UUID
}

// ExtractCitationsFromResponse extracts citations by matching the generated
// response text against the context items that were used for generation.
func (r *RAGResult) ExtractCitationsFromResponse(responseText string) []Citation {
	if r == nil {
		return nil
	}
	return ExtractCitations(responseText, r.contextItems, r.promptIDs)
}

// GenerateResponse runs the full RAG pipeline:
//  1. Generate embedding for the user message
//  2. Search for similar prompt versions
//  3. Build context from top-k results above min_score
//  4. Construct system prompt with retrieved context
//  5. Stream LLM response via Ollama
//
// Returns a RAGResult containing the streaming channel and context items
// for citation extraction. The caller should use ExtractCitations after
// consuming the full response to determine which prompts were referenced.
func (s *RAGService) GenerateResponse(ctx context.Context, sessionID uuid.UUID, userMessage string, orgID uuid.UUID) (*RAGResult, error) {
	if !s.Available() {
		return nil, fmt.Errorf("ragservice: service not available")
	}

	// Step 1: Generate embedding for the user message
	queryEmbedding, err := s.embSvc.GenerateEmbedding(ctx, userMessage)
	if err != nil {
		return nil, fmt.Errorf("ragservice: generate embedding: %w", err)
	}

	// Step 2: Search for similar prompt versions
	results, err := s.q.SearchPromptVersionsByEmbedding(ctx, db.SearchPromptVersionsByEmbeddingParams{
		Column1:        queryEmbedding,
		OrganizationID: orgID,
		Limit:          int32(DefaultTopK),
	})
	if err != nil {
		return nil, fmt.Errorf("ragservice: search prompt versions: %w", err)
	}

	// Step 3: Filter by min score and extract context + prompt IDs
	items := make([]contextItem, 0, len(results))
	promptIDs := make(map[string]uuid.UUID, len(results))
	for _, row := range results {
		if row.Similarity < DefaultMinScore {
			continue
		}
		items = append(items, contextItem{
			PromptID:      row.PromptID,
			PromptName:    row.PromptName,
			PromptSlug:    row.PromptSlug,
			VersionNumber: row.VersionNumber,
			Content:       extractContentText(row.Content),
			Similarity:    row.Similarity,
		})
		promptIDs[row.PromptSlug] = row.PromptID
	}

	logger.Info("RAG context retrieved",
		"session_id", sessionID,
		"results_total", len(results),
		"results_filtered", len(items),
	)

	// Step 4: Classify user intent and build system prompt with context
	intent := intentservice.Classify(userMessage)
	systemPrompt := BuildSystemPrompt(items, &intent)

	// Step 5: Send to Ollama for streaming generation
	chatReq := ollama.ChatRequest{
		Model: s.model,
		Messages: []ollama.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}

	chatCh, err := s.ollamaClient.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("ragservice: ollama chat: %w", err)
	}

	// Convert ChatResponse channel to string channel
	out := make(chan string, 16)
	go func() {
		defer close(out)
		for chunk := range chatCh {
			if chunk.Message.Content != "" {
				select {
				case out <- chunk.Message.Content:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return &RAGResult{
		Chunks:       out,
		contextItems: items,
		promptIDs:    promptIDs,
	}, nil
}

// BuildSystemPrompt constructs a system prompt incorporating retrieved context items
// and an optional classified intent. When intent is non-nil and not "general",
// it adds an intent hint to guide the LLM response.
// Exported for testability.
func BuildSystemPrompt(items []contextItem, intent *intentservice.Intent) string {
	var sb strings.Builder
	sb.WriteString("You are a helpful AI consulting assistant for QHub, a prompt management platform. ")
	sb.WriteString("Use the following relevant prompt examples from the user's organization to inform your response. ")
	sb.WriteString("Reference specific prompts when relevant.\n\n")

	if intent != nil && intent.Type != intentservice.IntentGeneral {
		sb.WriteString(fmt.Sprintf("## User Intent: %s (confidence: %.0f%%)\n", intent.Type, intent.Confidence*100))
		sb.WriteString("Tailor your response to address this specific intent.\n\n")
	}

	if len(items) == 0 {
		sb.WriteString("No relevant prompt examples were found. Provide a helpful general response.\n")
		return sb.String()
	}

	sb.WriteString("## Relevant Prompt Context\n\n")
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("### %d. %s (v%d, similarity: %.2f)\n", i+1, item.PromptName, item.VersionNumber, item.Similarity))
		sb.WriteString(item.Content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\nUse the above context to provide an informed, specific response. ")
	sb.WriteString("If the context is not directly relevant, acknowledge that and provide general guidance.\n")

	return sb.String()
}

// extractContentText extracts readable text from a JSON content field.
// The content may be a JSON string, or a JSON object with a "text" or "content" field.
func extractContentText(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}

	// Try as plain string
	var str string
	if err := json.Unmarshal(content, &str); err == nil {
		return str
	}

	// Try as object with "text" field
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(content, &obj); err == nil {
		if text, ok := obj["text"]; ok {
			var s string
			if err := json.Unmarshal(text, &s); err == nil {
				return s
			}
		}
		if text, ok := obj["content"]; ok {
			var s string
			if err := json.Unmarshal(text, &s); err == nil {
				return s
			}
		}
	}

	// Fallback: return raw JSON as string
	return string(content)
}
