package diffservice

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildDiff(t *testing.T) {
	tests := []struct {
		testName      string
		fromContent   string
		toContent     string
		expectChanges int
		expectTone    string
	}{
		// 正常系
		{
			testName:      "identical content produces no changes",
			fromContent:   "Hello world",
			toContent:     "Hello world",
			expectChanges: 0,
			expectTone:    "",
		},
		// 異常系: content changed
		{
			testName:      "content length change detected",
			fromContent:   "Short prompt.",
			toContent:     "A much longer prompt with additional instructions and details.",
			expectChanges: 1,
			expectTone:    "",
		},
		// 変数追加
		{
			testName:      "variable added",
			fromContent:   "Hello world",
			toContent:     "Hello {{name}}",
			expectChanges: 2, // length change + variable added
			expectTone:    "",
		},
		// 変数削除
		{
			testName:      "variable removed",
			fromContent:   "Hello {{name}}",
			toContent:     "Hello world",
			expectChanges: 2, // length change + variable removed
			expectTone:    "",
		},
		// トーンシフト
		{
			testName:      "tone shift detected",
			fromContent:   "just do it simply and cool",
			toContent:     "Please ensure you must comply with the required standards",
			expectChanges: 2, // length change + tone shift
			expectTone:    "casual → formal",
		},
		// 空文字
		{
			testName:      "empty to content",
			fromContent:   "some content",
			toContent:     "",
			expectChanges: 1,
			expectTone:    "",
		},
		// 特殊文字
		{
			testName:      "unicode content diff",
			fromContent:   "こんにちは",
			toContent:     "こんにちは世界 🌍",
			expectChanges: 1,
			expectTone:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := buildDiff(tt.fromContent, tt.toContent)

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result.Changes) < tt.expectChanges {
				t.Errorf("expected at least %d changes, got %d: %+v", tt.expectChanges, len(result.Changes), result.Changes)
			}
			if tt.expectTone != "" {
				if diff := cmp.Diff(tt.expectTone, result.ToneShift); diff != "" {
					t.Errorf("tone shift mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestDetectTone(t *testing.T) {
	tests := []struct {
		testName string
		content  string
		expected string
	}{
		{testName: "neutral content", content: "Write a response about dogs.", expected: "neutral"},
		{testName: "formal content", content: "Please ensure you kindly shall must comply.", expected: "formal"},
		{testName: "casual content", content: "Just simply do it like okay hey cool.", expected: "casual"},
		{testName: "strict content", content: "You must always never required mandatory shall not.", expected: "strict"},
		{testName: "friendly content", content: "Feel free, no worries, happy to help, glad, welcome.", expected: "friendly"},
		{testName: "empty content", content: "", expected: "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := detectTone(tt.content)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("tone mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFindVariables(t *testing.T) {
	tests := []struct {
		testName string
		content  string
		expected map[string]bool
	}{
		{testName: "no variables", content: "Hello world", expected: map[string]bool{}},
		{testName: "single variable", content: "Hello {{name}}", expected: map[string]bool{"name": true}},
		{testName: "multiple variables", content: "{{name}} is {{age}} years old", expected: map[string]bool{"name": true, "age": true}},
		{testName: "duplicate variable", content: "{{x}} and {{x}}", expected: map[string]bool{"x": true}},
		{testName: "empty content", content: "", expected: map[string]bool{}},
		{testName: "nested braces ignored", content: "{{{wrong}}}", expected: map[string]bool{"wrong": true}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := findVariables(tt.content)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLcsStrings(t *testing.T) {
	tests := []struct {
		testName string
		a        []string
		b        []string
		expected []string
	}{
		{testName: "identical", a: []string{"a", "b", "c"}, b: []string{"a", "b", "c"}, expected: []string{"a", "b", "c"}},
		{testName: "one added", a: []string{"a", "c"}, b: []string{"a", "b", "c"}, expected: []string{"a", "c"}},
		{testName: "one removed", a: []string{"a", "b", "c"}, b: []string{"a", "c"}, expected: []string{"a", "c"}},
		{testName: "completely different", a: []string{"a"}, b: []string{"b"}, expected: []string{}},
		{testName: "empty first", a: []string{}, b: []string{"a"}, expected: []string{}},
		{testName: "empty second", a: []string{"a"}, b: []string{}, expected: []string{}},
		{testName: "both empty", a: []string{}, b: []string{}, expected: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := lcsStrings(tt.a, tt.b)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtractContent(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		expected string
	}{
		{testName: "content key", input: `{"content":"hello world"}`, expected: "hello world"},
		{testName: "text key", input: `{"text":"some text"}`, expected: "some text"},
		{testName: "body key", input: `{"body":"body text"}`, expected: "body text"},
		{testName: "no match", input: `{"title":"ignored"}`, expected: `{"title":"ignored"}`},
		{testName: "invalid json", input: `not json at all`, expected: "not json at all"},
		{testName: "empty object", input: `{}`, expected: "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := extractContent([]byte(tt.input))
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildSummary(t *testing.T) {
	tests := []struct {
		testName string
		changes  int
		expected string
	}{
		{testName: "no changes", changes: 0, expected: "No significant changes detected"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := buildSummary(nil)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
