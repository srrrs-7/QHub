package ragservice

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestExtractCitations(t *testing.T) {
	promptID1 := uuid.New()
	promptID2 := uuid.New()

	type args struct {
		responseText string
		items        []contextItem
		promptIDs    map[string]uuid.UUID
	}
	type expected struct {
		count int
		slugs []string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - matches by prompt name
		{
			testName: "matches prompt by name mention",
			args: args{
				responseText: "Based on the Customer Support prompt, you should handle tickets promptly.",
				items: []contextItem{
					{PromptName: "Customer Support", PromptSlug: "customer-support", VersionNumber: 3, Similarity: 0.92},
				},
				promptIDs: map[string]uuid.UUID{"customer-support": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"customer-support"}},
		},
		// 正常系 - matches by slug
		{
			testName: "matches prompt by slug mention",
			args: args{
				responseText: "The sales-followup template is a good starting point.",
				items: []contextItem{
					{PromptName: "Sales Follow-up", PromptSlug: "sales-followup", VersionNumber: 1, Similarity: 0.78},
				},
				promptIDs: map[string]uuid.UUID{"sales-followup": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"sales-followup"}},
		},
		// 正常系 - matches by version number pattern "v3"
		{
			testName: "matches prompt by version pattern v3",
			args: args{
				responseText: "I recommend using v3 of the prompt for better results.",
				items: []contextItem{
					{PromptName: "UniqueXYZ", PromptSlug: "unique-xyz", VersionNumber: 3, Similarity: 0.85},
				},
				promptIDs: map[string]uuid.UUID{"unique-xyz": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"unique-xyz"}},
		},
		// 正常系 - matches by "version N" pattern
		{
			testName: "matches prompt by version N pattern",
			args: args{
				responseText: "Version 2 includes improvements over the previous iteration.",
				items: []contextItem{
					{PromptName: "UniqueABC", PromptSlug: "unique-abc", VersionNumber: 2, Similarity: 0.80},
				},
				promptIDs: map[string]uuid.UUID{"unique-abc": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"unique-abc"}},
		},
		// 正常系 - multiple matches
		{
			testName: "matches multiple prompts",
			args: args{
				responseText: "Combining the Customer Support approach with the Sales Follow-up template yields great results.",
				items: []contextItem{
					{PromptName: "Customer Support", PromptSlug: "customer-support", VersionNumber: 3, Similarity: 0.92},
					{PromptName: "Sales Follow-up", PromptSlug: "sales-followup", VersionNumber: 1, Similarity: 0.78},
				},
				promptIDs: map[string]uuid.UUID{
					"customer-support": promptID1,
					"sales-followup":   promptID2,
				},
			},
			expected: expected{count: 2, slugs: []string{"customer-support", "sales-followup"}},
		},
		// 正常系 - case insensitive name match
		{
			testName: "matches case insensitively",
			args: args{
				responseText: "the customer support prompt is very effective.",
				items: []contextItem{
					{PromptName: "Customer Support", PromptSlug: "customer-support", VersionNumber: 1, Similarity: 0.90},
				},
				promptIDs: map[string]uuid.UUID{"customer-support": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"customer-support"}},
		},
		// 異常系 - no match in response
		{
			testName: "no citations when response does not reference any item",
			args: args{
				responseText: "Here is a general recommendation for your workflow.",
				items: []contextItem{
					{PromptName: "Customer Support", PromptSlug: "customer-support", VersionNumber: 3, Similarity: 0.92},
				},
				promptIDs: map[string]uuid.UUID{"customer-support": promptID1},
			},
			expected: expected{count: 0, slugs: nil},
		},
		// 異常系 - nil promptIDs map
		{
			testName: "nil promptIDs map produces zero UUID",
			args: args{
				responseText: "The Customer Support prompt is great.",
				items: []contextItem{
					{PromptName: "Customer Support", PromptSlug: "customer-support", VersionNumber: 1, Similarity: 0.85},
				},
				promptIDs: nil,
			},
			expected: expected{count: 1, slugs: []string{"customer-support"}},
		},
		// 境界値 - empty response text
		{
			testName: "empty response text returns nil",
			args: args{
				responseText: "",
				items: []contextItem{
					{PromptName: "Test", PromptSlug: "test", VersionNumber: 1, Similarity: 0.9},
				},
				promptIDs: map[string]uuid.UUID{"test": promptID1},
			},
			expected: expected{count: 0, slugs: nil},
		},
		// 境界値 - empty items
		{
			testName: "empty items returns nil",
			args: args{
				responseText: "Some response text.",
				items:        []contextItem{},
				promptIDs:    map[string]uuid.UUID{},
			},
			expected: expected{count: 0, slugs: nil},
		},
		// 境界値 - version 0
		{
			testName: "matches version 0 pattern",
			args: args{
				responseText: "Using v0 of the draft, you can see the initial structure.",
				items: []contextItem{
					{PromptName: "UniqueQRS", PromptSlug: "unique-qrs", VersionNumber: 0, Similarity: 0.70},
				},
				promptIDs: map[string]uuid.UUID{"unique-qrs": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"unique-qrs"}},
		},
		// 境界値 - similarity 0.0 and 1.0
		{
			testName: "preserves similarity score in citation",
			args: args{
				responseText: "The Perfect Match prompt is ideal.",
				items: []contextItem{
					{PromptName: "Perfect Match", PromptSlug: "perfect-match", VersionNumber: 1, Similarity: 1.0},
				},
				promptIDs: map[string]uuid.UUID{"perfect-match": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"perfect-match"}},
		},
		// 特殊文字 - Japanese prompt name
		{
			testName: "matches Japanese prompt name",
			args: args{
				responseText: "日本語プロンプトのテンプレートを使用してください。",
				items: []contextItem{
					{PromptName: "日本語プロンプト", PromptSlug: "japanese-prompt", VersionNumber: 1, Similarity: 0.88},
				},
				promptIDs: map[string]uuid.UUID{"japanese-prompt": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"japanese-prompt"}},
		},
		// 特殊文字 - emoji in response
		{
			testName: "matches prompt when response contains emoji",
			args: args{
				responseText: "Great job! The Fun Prompt is perfect for this use case.",
				items: []contextItem{
					{PromptName: "Fun Prompt", PromptSlug: "fun-prompt", VersionNumber: 1, Similarity: 0.75},
				},
				promptIDs: map[string]uuid.UUID{"fun-prompt": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"fun-prompt"}},
		},
		// 特殊文字 - SQL injection in prompt name
		{
			testName: "handles SQL injection in prompt name safely",
			args: args{
				responseText: "'; DROP TABLE prompts; -- is not a valid approach.",
				items: []contextItem{
					{PromptName: "'; DROP TABLE prompts; --", PromptSlug: "sql-test", VersionNumber: 1, Similarity: 0.5},
				},
				promptIDs: map[string]uuid.UUID{"sql-test": promptID1},
			},
			expected: expected{count: 1, slugs: []string{"sql-test"}},
		},
		// 空文字 - empty prompt name and slug
		{
			testName: "no match when prompt name and slug are empty",
			args: args{
				responseText: "Some response mentioning v1.",
				items: []contextItem{
					{PromptName: "", PromptSlug: "", VersionNumber: 1, Similarity: 0.8},
				},
				promptIDs: map[string]uuid.UUID{},
			},
			expected: expected{count: 1, slugs: []string{""}},
		},
		// 空文字 - whitespace only response
		{
			testName: "whitespace only response returns nil",
			args: args{
				responseText: "   \t\n   ",
				items: []contextItem{
					{PromptName: "Test", PromptSlug: "test", VersionNumber: 1, Similarity: 0.9},
				},
				promptIDs: map[string]uuid.UUID{"test": promptID1},
			},
			expected: expected{count: 0, slugs: nil},
		},
		// Null/Nil - nil items
		{
			testName: "nil items returns nil",
			args: args{
				responseText: "Some response.",
				items:        nil,
				promptIDs:    nil,
			},
			expected: expected{count: 0, slugs: nil},
		},
		// Null/Nil - zero UUID in promptIDs
		{
			testName: "zero UUID in promptIDs map",
			args: args{
				responseText: "The Test Prompt is relevant.",
				items: []contextItem{
					{PromptName: "Test Prompt", PromptSlug: "test-prompt", VersionNumber: 1, Similarity: 0.9},
				},
				promptIDs: map[string]uuid.UUID{"test-prompt": uuid.Nil},
			},
			expected: expected{count: 1, slugs: []string{"test-prompt"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := ExtractCitations(tt.args.responseText, tt.args.items, tt.args.promptIDs)

			if diff := cmp.Diff(tt.expected.count, len(got)); diff != "" {
				t.Errorf("citation count mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.slugs != nil {
				gotSlugs := make([]string, len(got))
				for i, c := range got {
					gotSlugs[i] = c.PromptSlug
				}
				if diff := cmp.Diff(tt.expected.slugs, gotSlugs); diff != "" {
					t.Errorf("citation slugs mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExtractCitations_RelevanceScore(t *testing.T) {
	promptID := uuid.New()

	// 正常系 - verify relevance score is preserved
	t.Run("relevance score matches similarity", func(t *testing.T) {
		items := []contextItem{
			{PromptName: "Test Prompt", PromptSlug: "test-prompt", VersionNumber: 1, Similarity: 0.87},
		}
		promptIDs := map[string]uuid.UUID{"test-prompt": promptID}

		got := ExtractCitations("The Test Prompt is useful.", items, promptIDs)
		if len(got) != 1 {
			t.Fatalf("expected 1 citation, got %d", len(got))
		}
		if diff := cmp.Diff(0.87, got[0].RelevanceScore); diff != "" {
			t.Errorf("relevance score mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(promptID, got[0].PromptID); diff != "" {
			t.Errorf("prompt ID mismatch (-want +got):\n%s", diff)
		}
	})

	// 境界値 - similarity 0
	t.Run("zero similarity preserved", func(t *testing.T) {
		items := []contextItem{
			{PromptName: "Zero Score", PromptSlug: "zero-score", VersionNumber: 1, Similarity: 0.0},
		}
		got := ExtractCitations("The Zero Score is mentioned.", items, nil)
		if len(got) != 1 {
			t.Fatalf("expected 1 citation, got %d", len(got))
		}
		if diff := cmp.Diff(0.0, got[0].RelevanceScore); diff != "" {
			t.Errorf("relevance score mismatch (-want +got):\n%s", diff)
		}
	})

	// 境界値 - similarity 1.0
	t.Run("perfect similarity preserved", func(t *testing.T) {
		items := []contextItem{
			{PromptName: "Perfect", PromptSlug: "perfect", VersionNumber: 1, Similarity: 1.0},
		}
		got := ExtractCitations("The Perfect prompt matches.", items, nil)
		if len(got) != 1 {
			t.Fatalf("expected 1 citation, got %d", len(got))
		}
		if diff := cmp.Diff(1.0, got[0].RelevanceScore); diff != "" {
			t.Errorf("relevance score mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestMarshalCitations(t *testing.T) {
	type args struct {
		citations []Citation
	}
	type expected struct {
		isNil     bool
		validJSON bool
		count     int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "marshals single citation",
			args: args{
				citations: []Citation{
					{PromptID: uuid.Nil, PromptSlug: "test", VersionNumber: 1, RelevanceScore: 0.85},
				},
			},
			expected: expected{isNil: false, validJSON: true, count: 1},
		},
		// 正常系 - multiple citations
		{
			testName: "marshals multiple citations",
			args: args{
				citations: []Citation{
					{PromptID: uuid.Nil, PromptSlug: "first", VersionNumber: 1, RelevanceScore: 0.9},
					{PromptID: uuid.Nil, PromptSlug: "second", VersionNumber: 2, RelevanceScore: 0.8},
				},
			},
			expected: expected{isNil: false, validJSON: true, count: 2},
		},
		// 境界値 - empty slice
		{
			testName: "returns nil for empty slice",
			args:     args{citations: []Citation{}},
			expected: expected{isNil: true},
		},
		// Null/Nil
		{
			testName: "returns nil for nil slice",
			args:     args{citations: nil},
			expected: expected{isNil: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := MarshalCitations(tt.args.citations)

			if tt.expected.isNil {
				if got != nil {
					t.Errorf("expected nil, got %s", string(got))
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil result")
			}

			if tt.expected.validJSON {
				var parsed []Citation
				if err := json.Unmarshal(got, &parsed); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if diff := cmp.Diff(tt.expected.count, len(parsed)); diff != "" {
					t.Errorf("citation count mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestMarshalCitations_RoundTrip(t *testing.T) {
	// 正常系 - round trip preserves data
	t.Run("round trip preserves all fields", func(t *testing.T) {
		promptID := uuid.New()
		original := []Citation{
			{PromptID: promptID, PromptSlug: "my-prompt", VersionNumber: 5, RelevanceScore: 0.92},
		}

		data := MarshalCitations(original)
		if data == nil {
			t.Fatal("expected non-nil marshaled data")
		}

		var restored []Citation
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if diff := cmp.Diff(original, restored); diff != "" {
			t.Errorf("round trip mismatch (-want +got):\n%s", diff)
		}
	})

	// 特殊文字 - Japanese slug round trip
	t.Run("round trip with special characters in slug", func(t *testing.T) {
		original := []Citation{
			{PromptID: uuid.Nil, PromptSlug: "日本語-slug", VersionNumber: 1, RelevanceScore: 0.75},
		}

		data := MarshalCitations(original)
		var restored []Citation
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if diff := cmp.Diff(original, restored); diff != "" {
			t.Errorf("round trip mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestRAGResult_ExtractCitationsFromResponse(t *testing.T) {
	promptID := uuid.New()

	type args struct {
		result       *RAGResult
		responseText string
	}
	type expected struct {
		count int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - extracts citations from result
		{
			testName: "extracts matching citations",
			args: args{
				result: &RAGResult{
					contextItems: []contextItem{
						{PromptName: "Test Prompt", PromptSlug: "test-prompt", VersionNumber: 1, Similarity: 0.9},
					},
					promptIDs: map[string]uuid.UUID{"test-prompt": promptID},
				},
				responseText: "The Test Prompt is relevant here.",
			},
			expected: expected{count: 1},
		},
		// 異常系 - no match
		{
			testName: "returns nil when no match",
			args: args{
				result: &RAGResult{
					contextItems: []contextItem{
						{PromptName: "Unique Prompt", PromptSlug: "unique-prompt", VersionNumber: 99, Similarity: 0.9},
					},
					promptIDs: map[string]uuid.UUID{"unique-prompt": promptID},
				},
				responseText: "Nothing relevant in this response.",
			},
			expected: expected{count: 0},
		},
		// Null/Nil - nil result
		{
			testName: "nil result returns nil",
			args: args{
				result:       nil,
				responseText: "Some response.",
			},
			expected: expected{count: 0},
		},
		// 空文字 - empty response text
		{
			testName: "empty response returns nil",
			args: args{
				result: &RAGResult{
					contextItems: []contextItem{
						{PromptName: "Test", PromptSlug: "test", VersionNumber: 1, Similarity: 0.9},
					},
				},
				responseText: "",
			},
			expected: expected{count: 0},
		},
		// 境界値 - empty context items
		{
			testName: "empty context items returns nil",
			args: args{
				result:       &RAGResult{contextItems: nil, promptIDs: nil},
				responseText: "Some response.",
			},
			expected: expected{count: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := tt.args.result.ExtractCitationsFromResponse(tt.args.responseText)
			if diff := cmp.Diff(tt.expected.count, len(got)); diff != "" {
				t.Errorf("citation count mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMatchesResponse(t *testing.T) {
	type args struct {
		lowerResponse string
		item          contextItem
	}
	type expected struct {
		matches bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - name match
		{
			testName: "matches by prompt name",
			args: args{
				lowerResponse: "the customer support prompt is great",
				item:          contextItem{PromptName: "Customer Support", PromptSlug: "customer-support", VersionNumber: 1},
			},
			expected: expected{matches: true},
		},
		// 正常系 - slug match
		{
			testName: "matches by slug",
			args: args{
				lowerResponse: "using the sales-followup template",
				item:          contextItem{PromptName: "Sales", PromptSlug: "sales-followup", VersionNumber: 1},
			},
			expected: expected{matches: true},
		},
		// 正常系 - version short pattern
		{
			testName: "matches by v3 pattern",
			args: args{
				lowerResponse: "i recommend v3 for production",
				item:          contextItem{PromptName: "UniqueNOP", PromptSlug: "unique-nop", VersionNumber: 3},
			},
			expected: expected{matches: true},
		},
		// 正常系 - version long pattern
		{
			testName: "matches by version 5 pattern",
			args: args{
				lowerResponse: "version 5 includes the latest changes",
				item:          contextItem{PromptName: "UniqueDEF", PromptSlug: "unique-def", VersionNumber: 5},
			},
			expected: expected{matches: true},
		},
		// 異常系 - no match
		{
			testName: "no match when nothing references the item",
			args: args{
				lowerResponse: "here is a general recommendation",
				item:          contextItem{PromptName: "Specific Prompt", PromptSlug: "specific-prompt", VersionNumber: 99},
			},
			expected: expected{matches: false},
		},
		// 境界値 - empty name and slug, version 0 with v0 in text
		{
			testName: "matches v0 when name and slug are empty",
			args: args{
				lowerResponse: "use v0 as a draft",
				item:          contextItem{PromptName: "", PromptSlug: "", VersionNumber: 0},
			},
			expected: expected{matches: true},
		},
		// 境界値 - empty name and slug, no version match
		{
			testName: "no match when name and slug empty and no version reference",
			args: args{
				lowerResponse: "no references here at all",
				item:          contextItem{PromptName: "", PromptSlug: "", VersionNumber: 99},
			},
			expected: expected{matches: false},
		},
		// 空文字 - empty response
		{
			testName: "no match on empty response",
			args: args{
				lowerResponse: "",
				item:          contextItem{PromptName: "Test", PromptSlug: "test", VersionNumber: 1},
			},
			expected: expected{matches: false},
		},
		// 特殊文字 - Japanese name match
		{
			testName: "matches Japanese name",
			args: args{
				lowerResponse: "日本語プロンプトを参照してください",
				item:          contextItem{PromptName: "日本語プロンプト", PromptSlug: "jp", VersionNumber: 1},
			},
			expected: expected{matches: true},
		},
		// Null/Nil - all empty fields
		{
			testName: "no match when all item fields empty and version 0 not in text",
			args: args{
				lowerResponse: "nothing relevant here",
				item:          contextItem{PromptName: "", PromptSlug: "", VersionNumber: 0},
			},
			expected: expected{matches: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := matchesResponse(tt.args.lowerResponse, tt.args.item)
			if diff := cmp.Diff(tt.expected.matches, got); diff != "" {
				t.Errorf("matchesResponse mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
