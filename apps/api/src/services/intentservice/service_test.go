package intentservice

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClassify(t *testing.T) {
	type args struct {
		message string
	}
	type expected struct {
		intentType string
		confidence float64
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - improve intent
		{
			testName: "improve keyword",
			args:     args{message: "Can you improve this prompt?"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "better keyword",
			args:     args{message: "How can I make this better?"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "optimize keyword",
			args:     args{message: "Please optimize my prompt"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "enhance keyword",
			args:     args{message: "Enhance the quality of this prompt"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "refine keyword",
			args:     args{message: "I want to refine my prompt"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},

		// 正常系 - compare intent
		{
			testName: "compare keyword",
			args:     args{message: "Compare version 1 and version 2"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},
		{
			testName: "difference keyword",
			args:     args{message: "What is the difference between these?"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},
		{
			testName: "diff keyword",
			args:     args{message: "Show me the diff"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},
		{
			testName: "versus keyword",
			args:     args{message: "Version 1 versus version 2"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},

		// 正常系 - create intent
		{
			testName: "create keyword",
			args:     args{message: "Create a new prompt for customer support"},
			expected: expected{intentType: IntentCreate, confidence: 0.8},
		},
		{
			testName: "write keyword",
			args:     args{message: "Write a prompt for me"},
			expected: expected{intentType: IntentCreate, confidence: 0.8},
		},
		{
			testName: "generate keyword",
			args:     args{message: "Generate a draft prompt"},
			expected: expected{intentType: IntentCreate, confidence: 0.8},
		},
		{
			testName: "new prompt keyword",
			args:     args{message: "I need a new prompt"},
			expected: expected{intentType: IntentCreate, confidence: 0.8},
		},

		// 正常系 - compliance intent
		{
			testName: "compliance keyword",
			args:     args{message: "Is this prompt compliance ready?"},
			expected: expected{intentType: IntentCompliance, confidence: 0.8},
		},
		{
			testName: "hipaa keyword",
			args:     args{message: "Does this meet HIPAA requirements?"},
			expected: expected{intentType: IntentCompliance, confidence: 0.8},
		},
		{
			testName: "gdpr keyword",
			args:     args{message: "Check GDPR compliance"},
			expected: expected{intentType: IntentCompliance, confidence: 0.8},
		},
		{
			testName: "regulation keyword",
			args:     args{message: "What regulation applies here?"},
			expected: expected{intentType: IntentCompliance, confidence: 0.8},
		},

		// 正常系 - best practice intent
		{
			testName: "best practice keyword",
			args:     args{message: "What are the best practice for prompts?"},
			expected: expected{intentType: IntentBestPractice, confidence: 0.7},
		},
		{
			testName: "recommendation keyword",
			args:     args{message: "Any recommendation for this?"},
			expected: expected{intentType: IntentBestPractice, confidence: 0.7},
		},
		{
			testName: "guideline keyword",
			args:     args{message: "Show me the guideline"},
			expected: expected{intentType: IntentBestPractice, confidence: 0.7},
		},
		{
			testName: "tip keyword",
			args:     args{message: "Any tip for writing prompts?"},
			expected: expected{intentType: IntentBestPractice, confidence: 0.7},
		},

		// 正常系 - explain intent
		{
			testName: "explain keyword",
			args:     args{message: "Can you explain this prompt?"},
			expected: expected{intentType: IntentExplain, confidence: 0.7},
		},
		{
			testName: "what does keyword",
			args:     args{message: "What does this prompt do?"},
			expected: expected{intentType: IntentExplain, confidence: 0.7},
		},
		{
			testName: "how does keyword",
			args:     args{message: "How does this work?"},
			expected: expected{intentType: IntentExplain, confidence: 0.7},
		},

		// 正常系 - general intent
		{
			testName: "general conversation",
			args:     args{message: "Hello, how are you?"},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},
		{
			testName: "ambiguous message",
			args:     args{message: "I have a question about prompts"},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},

		// 特殊文字 - Japanese patterns
		{
			testName: "Japanese improve (改善)",
			args:     args{message: "このプロンプトを改善してください"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "Japanese compare (比較)",
			args:     args{message: "バージョンを比較してください"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},
		{
			testName: "Japanese create (作成)",
			args:     args{message: "新しいプロンプトを作成して"},
			expected: expected{intentType: IntentCreate, confidence: 0.8},
		},
		{
			testName: "Japanese compliance (コンプライアンス)",
			args:     args{message: "コンプライアンスのチェックをお願いします"},
			expected: expected{intentType: IntentCompliance, confidence: 0.8},
		},
		{
			testName: "Japanese best practice (ベストプラクティス)",
			args:     args{message: "ベストプラクティスを教えて"},
			expected: expected{intentType: IntentBestPractice, confidence: 0.7},
		},
		{
			testName: "Japanese explain (説明)",
			args:     args{message: "このプロンプトの説明をお願いします"},
			expected: expected{intentType: IntentExplain, confidence: 0.7},
		},

		// 空文字
		{
			testName: "empty string returns general",
			args:     args{message: ""},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},
		{
			testName: "whitespace only returns general",
			args:     args{message: "   "},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},

		// 特殊文字 - special characters
		{
			testName: "SQL injection returns general",
			args:     args{message: "'; DROP TABLE prompts; --"},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},
		{
			testName: "emoji only returns general",
			args:     args{message: "🚀🎉💡"},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},

		// 境界値 - mixed case
		{
			testName: "uppercase IMPROVE",
			args:     args{message: "IMPROVE this prompt"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "mixed case Compare",
			args:     args{message: "Compare these prompts"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},

		// 境界値 - multiple intent signals (first match wins)
		{
			testName: "improve takes priority over compare",
			args:     args{message: "Improve by comparing with the original"},
			expected: expected{intentType: IntentImprove, confidence: 0.8},
		},
		{
			testName: "compare takes priority over create",
			args:     args{message: "Compare and then create a new version"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},

		// Null/Nil - single character
		{
			testName: "single char returns general",
			args:     args{message: "x"},
			expected: expected{intentType: IntentGeneral, confidence: 0.5},
		},

		// 境界値 - pattern as substring
		{
			testName: "vs as separate word",
			args:     args{message: "v1 vs v2"},
			expected: expected{intentType: IntentCompare, confidence: 0.8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := Classify(tt.args.message)

			if diff := cmp.Diff(tt.expected.intentType, got.Type); diff != "" {
				t.Errorf("intent type mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.confidence, got.Confidence); diff != "" {
				t.Errorf("confidence mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClassify_EntitiesField(t *testing.T) {
	// Null/Nil - entities should be nil by default
	t.Run("entities are nil by default", func(t *testing.T) {
		got := Classify("improve this prompt")
		if got.Entities != nil {
			t.Errorf("expected nil entities, got %v", got.Entities)
		}
	})
}

func TestMatchesAny(t *testing.T) {
	type args struct {
		text     string
		patterns []string
	}
	type expected struct {
		matches bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "matches first pattern", args: args{text: "hello world", patterns: []string{"hello", "goodbye"}}, expected: expected{matches: true}},
		{testName: "matches last pattern", args: args{text: "hello world", patterns: []string{"goodbye", "world"}}, expected: expected{matches: true}},
		// 異常系
		{testName: "no match", args: args{text: "hello world", patterns: []string{"foo", "bar"}}, expected: expected{matches: false}},
		// 空文字
		{testName: "empty text", args: args{text: "", patterns: []string{"hello"}}, expected: expected{matches: false}},
		// Null/Nil
		{testName: "empty patterns", args: args{text: "hello", patterns: []string{}}, expected: expected{matches: false}},
		{testName: "nil patterns", args: args{text: "hello", patterns: nil}, expected: expected{matches: false}},
		// 境界値
		{testName: "empty pattern matches everything", args: args{text: "hello", patterns: []string{""}}, expected: expected{matches: true}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := matchesAny(tt.args.text, tt.args.patterns)
			if diff := cmp.Diff(tt.expected.matches, got); diff != "" {
				t.Errorf("matchesAny mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
