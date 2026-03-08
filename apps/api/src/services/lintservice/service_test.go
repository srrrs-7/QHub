package lintservice

import (
	"strings"
	"testing"

	"api/src/domain/intelligence"

	"github.com/google/go-cmp/cmp"
	"github.com/sqlc-dev/pqtype"
)

func TestRunLintRules(t *testing.T) {
	type args struct {
		content   string
		variables map[string]bool
	}
	type expected struct {
		score      int
		issueCount int
		passedKeys []string
		issueRules []string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "clean prompt with format and declared variables",
			args: args{
				content:   "Please respond in JSON format using {{name}} and {{role}}.",
				variables: map[string]bool{"name": true, "role": true},
			},
			expected: expected{score: 100, issueCount: 0, passedKeys: []string{"excessive-length", "missing-output-format", "variable-check", "no-vague-instruction"}},
		},
		// 異常系: undeclared variable
		{
			testName: "undeclared variable triggers error",
			args: args{
				content:   "Hello {{name}}, your role is {{role}}.",
				variables: map[string]bool{"name": true},
			},
			expected: expected{score: 65, issueCount: 2, issueRules: []string{"missing-output-format", "variable-check"}},
		},
		// 境界値: exactly max length (including " json" = 5 chars, so base is maxContentLength - 5)
		{
			testName: "content at max length passes",
			args: args{
				content:   strings.Repeat("a", maxContentLength-5) + " json",
				variables: map[string]bool{},
			},
			expected: expected{score: 100, issueCount: 0, passedKeys: []string{"excessive-length", "missing-output-format", "variable-check", "no-vague-instruction"}},
		},
		// 境界値: one over max length
		{
			testName: "content over max length triggers warning",
			args: args{
				content:   strings.Repeat("a", maxContentLength+1) + " json",
				variables: map[string]bool{},
			},
			expected: expected{score: 90, issueCount: 1, issueRules: []string{"excessive-length"}},
		},
		// 空文字
		{
			testName: "empty content",
			args: args{
				content:   "",
				variables: map[string]bool{},
			},
			expected: expected{score: 90, issueCount: 1, issueRules: []string{"missing-output-format"}},
		},
		// 特殊文字
		{
			testName: "content with Japanese and emoji",
			args: args{
				content:   "こんにちは 🤖 JSON形式で出力してください",
				variables: map[string]bool{},
			},
			expected: expected{score: 100, issueCount: 0, passedKeys: []string{"excessive-length", "missing-output-format", "variable-check", "no-vague-instruction"}},
		},
		// 異常系: vague words
		{
			testName: "vague words trigger info",
			args: args{
				content:   "Please provide a good and appropriate response in JSON format.",
				variables: map[string]bool{},
			},
			expected: expected{score: 95, issueCount: 1, issueRules: []string{"no-vague-instruction"}},
		},
		// Nil variables
		{
			testName: "nil variables map with no content vars",
			args: args{
				content:   "Simple prompt in markdown format.",
				variables: nil,
			},
			expected: expected{score: 100, issueCount: 0},
		},
		// 異常系: multiple issues
		{
			testName: "multiple issues compound score deductions",
			args: args{
				content:   "Write a good response using {{unknown_var}}.",
				variables: map[string]bool{},
			},
			expected: expected{score: 60, issueCount: 3, issueRules: []string{"missing-output-format", "variable-check", "no-vague-instruction"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			vars := tt.args.variables
			if vars == nil {
				vars = map[string]bool{}
			}
			result := runLintRules(tt.args.content, vars)

			if diff := cmp.Diff(tt.expected.score, result.Score); diff != "" {
				t.Errorf("score mismatch (-want +got):\n%s\nissues: %+v", diff, result.Issues)
			}
			if diff := cmp.Diff(tt.expected.issueCount, len(result.Issues)); diff != "" {
				t.Errorf("issue count mismatch (-want +got):\n%s\nissues: %+v", diff, result.Issues)
			}
			if len(tt.expected.passedKeys) > 0 {
				if diff := cmp.Diff(tt.expected.passedKeys, result.Passed); diff != "" {
					t.Errorf("passed keys mismatch (-want +got):\n%s", diff)
				}
			}
			if len(tt.expected.issueRules) > 0 {
				rules := make([]string, len(result.Issues))
				for i, iss := range result.Issues {
					rules[i] = iss.Rule
				}
				if diff := cmp.Diff(tt.expected.issueRules, rules); diff != "" {
					t.Errorf("issue rules mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		testName string
		issues   []intelligence.LintIssue
		expected int
	}{
		{testName: "no issues", issues: nil, expected: 100},
		{testName: "one error", issues: []intelligence.LintIssue{{Severity: "error"}}, expected: 75},
		{testName: "one warning", issues: []intelligence.LintIssue{{Severity: "warning"}}, expected: 90},
		{testName: "one info", issues: []intelligence.LintIssue{{Severity: "info"}}, expected: 95},
		{testName: "all types", issues: []intelligence.LintIssue{{Severity: "error"}, {Severity: "warning"}, {Severity: "info"}}, expected: 60},
		{testName: "score floors at 0", issues: []intelligence.LintIssue{
			{Severity: "error"}, {Severity: "error"}, {Severity: "error"}, {Severity: "error"}, {Severity: "error"},
		}, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := calculateScore(tt.issues)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("score mismatch (-want +got):\n%s", diff)
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
		{testName: "content key", input: `{"content":"hello"}`, expected: "hello"},
		{testName: "text key", input: `{"text":"world"}`, expected: "world"},
		{testName: "body key", input: `{"body":"test"}`, expected: "test"},
		{testName: "no known key", input: `{"other":"value"}`, expected: `{"other":"value"}`},
		{testName: "plain string", input: `"just a string"`, expected: `"just a string"`},
		{testName: "invalid json", input: `not json`, expected: "not json"},
		{testName: "empty object", input: `{}`, expected: "{}"},
		// 特殊文字
		{testName: "unicode content", input: `{"content":"日本語テスト 🎉"}`, expected: "日本語テスト 🎉"},
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

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		testName string
		input    string
		valid    bool
		expected map[string]bool
	}{
		{testName: "array of objects", input: `[{"name":"foo"},{"name":"bar"}]`, valid: true, expected: map[string]bool{"foo": true, "bar": true}},
		{testName: "array of strings", input: `["x","y"]`, valid: true, expected: map[string]bool{"x": true, "y": true}},
		{testName: "null/invalid", input: ``, valid: false, expected: map[string]bool{}},
		{testName: "empty array", input: `[]`, valid: true, expected: map[string]bool{}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			raw := pqtype.NullRawMessage{Valid: tt.valid}
			if tt.valid {
				raw.RawMessage = []byte(tt.input)
			}
			got := extractVariables(raw)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFindVagueWords(t *testing.T) {
	tests := []struct {
		testName string
		content  string
		expected int
	}{
		{testName: "no vague words", content: "Be specific and concise.", expected: 0},
		{testName: "one vague word", content: "Write a good response.", expected: 1},
		{testName: "multiple vague words", content: "Write a good, appropriate, nice response.", expected: 3},
		{testName: "case insensitive", content: "Write a GOOD response.", expected: 1},
		{testName: "empty", content: "", expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := findVagueWords(tt.content)
			if diff := cmp.Diff(tt.expected, len(got)); diff != "" {
				t.Errorf("count mismatch (-want +got):\n%s, found: %v", diff, got)
			}
		})
	}
}
