// Package ragservice implements citation extraction for RAG responses.
//
// After the LLM generates a response using context items, ExtractCitations
// checks which context items are actually referenced in the generated text
// by matching prompt names, slugs, and version numbers.
package ragservice

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Citation represents a reference to a prompt version that was used
// in generating an AI response.
type Citation struct {
	PromptID       uuid.UUID `json:"prompt_id"`
	PromptSlug     string    `json:"prompt_slug"`
	VersionNumber  int32     `json:"version_number"`
	RelevanceScore float64   `json:"relevance_score"`
}

// ExtractCitations analyzes the generated response text against the context
// items that were provided to the LLM. A context item is considered cited if
// the response mentions the prompt name, slug, or version number pattern (e.g. "v3").
//
// Each matched context item produces a Citation with the original similarity
// score as the relevance score. Items not referenced in the response are excluded.
func ExtractCitations(responseText string, items []contextItem, promptIDs map[string]uuid.UUID) []Citation {
	if responseText == "" || len(items) == 0 {
		return nil
	}

	lower := strings.ToLower(responseText)
	var citations []Citation

	for _, item := range items {
		if matchesResponse(lower, item) {
			promptID := uuid.Nil
			if promptIDs != nil {
				if id, ok := promptIDs[item.PromptSlug]; ok {
					promptID = id
				}
			}
			citations = append(citations, Citation{
				PromptID:       promptID,
				PromptSlug:     item.PromptSlug,
				VersionNumber:  item.VersionNumber,
				RelevanceScore: item.Similarity,
			})
		}
	}

	return citations
}

// matchesResponse checks if the generated response text references the given
// context item. It checks for:
//  1. Prompt name (case-insensitive)
//  2. Prompt slug (case-insensitive)
//  3. Version number pattern like "v3" or "version 3"
func matchesResponse(lowerResponse string, item contextItem) bool {
	// Check prompt name (case-insensitive)
	if item.PromptName != "" && strings.Contains(lowerResponse, strings.ToLower(item.PromptName)) {
		return true
	}

	// Check prompt slug (case-insensitive)
	if item.PromptSlug != "" && strings.Contains(lowerResponse, strings.ToLower(item.PromptSlug)) {
		return true
	}

	// Check version number patterns: "v3", "version 3"
	vShort := fmt.Sprintf("v%d", item.VersionNumber)
	if strings.Contains(lowerResponse, vShort) {
		return true
	}
	vLong := fmt.Sprintf("version %d", item.VersionNumber)
	return strings.Contains(lowerResponse, vLong)
}

// MarshalCitations converts a slice of citations to JSON for storage
// in the database citations column. Returns nil if there are no citations.
func MarshalCitations(citations []Citation) json.RawMessage {
	if len(citations) == 0 {
		return nil
	}
	data, err := json.Marshal(citations)
	if err != nil {
		return nil
	}
	return data
}
