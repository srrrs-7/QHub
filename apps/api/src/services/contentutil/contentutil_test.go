package contentutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// --- ExtractText ---

func TestExtractText(t *testing.T) {
	type args struct {
		raw json.RawMessage
	}
	type expected struct {
		result string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "extracts content key",
			args:     args{raw: json.RawMessage(`{"content":"hello world"}`)},
			expected: expected{result: "hello world"},
		},
		{
			testName: "extracts text key",
			args:     args{raw: json.RawMessage(`{"text":"some text"}`)},
			expected: expected{result: "some text"},
		},
		{
			testName: "extracts body key",
			args:     args{raw: json.RawMessage(`{"body":"body content"}`)},
			expected: expected{result: "body content"},
		},
		{
			testName: "extracts system key",
			args:     args{raw: json.RawMessage(`{"system":"system prompt"}`)},
			expected: expected{result: "system prompt"},
		},
		{
			testName: "extracts user key",
			args:     args{raw: json.RawMessage(`{"user":"user message"}`)},
			expected: expected{result: "user message"},
		},
		{
			testName: "content key takes priority over text",
			args:     args{raw: json.RawMessage(`{"content":"first","text":"second"}`)},
			expected: expected{result: "first"},
		},
		{
			testName: "text key takes priority over body",
			args:     args{raw: json.RawMessage(`{"text":"first","body":"second"}`)},
			expected: expected{result: "first"},
		},

		// 異常系 (Error Cases)
		{
			testName: "invalid JSON returns raw string",
			args:     args{raw: json.RawMessage(`not json at all`)},
			expected: expected{result: "not json at all"},
		},
		{
			testName: "JSON array returns raw string",
			args:     args{raw: json.RawMessage(`[1,2,3]`)},
			expected: expected{result: "[1,2,3]"},
		},
		{
			testName: "non-string content value returns raw",
			args:     args{raw: json.RawMessage(`{"content":123}`)},
			expected: expected{result: `{"content":123}`},
		},
		{
			testName: "non-string text value returns raw",
			args:     args{raw: json.RawMessage(`{"text":true}`)},
			expected: expected{result: `{"text":true}`},
		},
		{
			testName: "no matching keys returns raw",
			args:     args{raw: json.RawMessage(`{"foo":"bar","baz":"qux"}`)},
			expected: expected{result: `{"foo":"bar","baz":"qux"}`},
		},
		{
			testName: "content key with null value skips to next key",
			args:     args{raw: json.RawMessage(`{"content":null,"text":"fallback"}`)},
			expected: expected{result: "fallback"},
		},

		// 境界値 (Boundary Values)
		{
			testName: "very long content string",
			args:     args{raw: json.RawMessage(`{"content":"` + strings.Repeat("a", 10000) + `"}`)},
			expected: expected{result: strings.Repeat("a", 10000)},
		},
		{
			testName: "single character content",
			args:     args{raw: json.RawMessage(`{"content":"x"}`)},
			expected: expected{result: "x"},
		},
		{
			testName: "nested object in content returns raw",
			args:     args{raw: json.RawMessage(`{"content":{"nested":"value"}}`)},
			expected: expected{result: `{"content":{"nested":"value"}}`},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "content with emoji",
			args:     args{raw: json.RawMessage(`{"content":"Hello 🌍🎉"}`)},
			expected: expected{result: "Hello 🌍🎉"},
		},
		{
			testName: "content with Japanese characters",
			args:     args{raw: json.RawMessage(`{"content":"こんにちは世界"}`)},
			expected: expected{result: "こんにちは世界"},
		},
		{
			testName: "content with newlines and tabs",
			args:     args{raw: json.RawMessage(`{"content":"line1\nline2\ttab"}`)},
			expected: expected{result: "line1\nline2\ttab"},
		},
		{
			testName: "content with HTML tags",
			args:     args{raw: json.RawMessage(`{"content":"<b>bold</b>"}`)},
			expected: expected{result: "<b>bold</b>"},
		},
		{
			testName: "content with SQL injection",
			args:     args{raw: json.RawMessage(`{"content":"'; DROP TABLE users;--"}`)},
			expected: expected{result: "'; DROP TABLE users;--"},
		},
		{
			testName: "content with unicode escapes",
			args:     args{raw: json.RawMessage(`{"content":"\u0048\u0065\u006c\u006c\u006f"}`)},
			expected: expected{result: "Hello"},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty content string",
			args:     args{raw: json.RawMessage(`{"content":""}`)},
			expected: expected{result: ""},
		},
		{
			testName: "whitespace-only content",
			args:     args{raw: json.RawMessage(`{"content":"   "}`)},
			expected: expected{result: "   "},
		},
		{
			testName: "empty JSON object",
			args:     args{raw: json.RawMessage(`{}`)},
			expected: expected{result: "{}"},
		},

		// Null/Nil
		{
			testName: "JSON null literal",
			args:     args{raw: json.RawMessage(`null`)},
			expected: expected{result: "null"},
		},
		{
			testName: "empty raw message",
			args:     args{raw: json.RawMessage(``)},
			expected: expected{result: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := ExtractText(tt.args.raw)
			if diff := cmp.Diff(tt.expected.result, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// --- FindVariables ---

func TestFindVariables(t *testing.T) {
	type args struct {
		content string
	}
	type expected struct {
		result map[string]bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "single variable",
			args:     args{content: "Hello {{name}}!"},
			expected: expected{result: map[string]bool{"name": true}},
		},
		{
			testName: "multiple variables",
			args:     args{content: "Dear {{first_name}} {{last_name}}, your order {{order_id}} is ready."},
			expected: expected{result: map[string]bool{"first_name": true, "last_name": true, "order_id": true}},
		},
		{
			testName: "duplicate variables are deduplicated",
			args:     args{content: "{{name}} said hello to {{name}}"},
			expected: expected{result: map[string]bool{"name": true}},
		},
		{
			testName: "variable with underscores",
			args:     args{content: "{{my_var_name}}"},
			expected: expected{result: map[string]bool{"my_var_name": true}},
		},
		{
			testName: "variable with digits",
			args:     args{content: "{{var1}} and {{var2}}"},
			expected: expected{result: map[string]bool{"var1": true, "var2": true}},
		},

		// 異常系 (Error Cases)
		{
			testName: "no variables in content",
			args:     args{content: "Just plain text without variables."},
			expected: expected{result: map[string]bool{}},
		},
		{
			testName: "single braces are not variables",
			args:     args{content: "{name} is not a variable"},
			expected: expected{result: map[string]bool{}},
		},
		{
			testName: "triple braces captures inner variable",
			args:     args{content: "{{{name}}}"},
			expected: expected{result: map[string]bool{"name": true}},
		},
		{
			testName: "variable with spaces is not matched",
			args:     args{content: "{{ name }}"},
			expected: expected{result: map[string]bool{}},
		},
		{
			testName: "variable with hyphen is not fully matched",
			args:     args{content: "{{my-var}}"},
			expected: expected{result: map[string]bool{}},
		},

		// 境界値 (Boundary Values)
		{
			testName: "single character variable name",
			args:     args{content: "{{x}}"},
			expected: expected{result: map[string]bool{"x": true}},
		},
		{
			testName: "very long variable name",
			args:     args{content: "{{" + strings.Repeat("a", 1000) + "}}"},
			expected: expected{result: map[string]bool{strings.Repeat("a", 1000): true}},
		},
		{
			testName: "variable at start of string",
			args:     args{content: "{{start}} of text"},
			expected: expected{result: map[string]bool{"start": true}},
		},
		{
			testName: "variable at end of string",
			args:     args{content: "end of {{text}}"},
			expected: expected{result: map[string]bool{"text": true}},
		},
		{
			testName: "adjacent variables",
			args:     args{content: "{{a}}{{b}}"},
			expected: expected{result: map[string]bool{"a": true, "b": true}},
		},
		{
			testName: "many variables",
			args: args{content: func() string {
				var sb strings.Builder
				for i := 0; i < 100; i++ {
					sb.WriteString("{{var")
					sb.WriteString(strings.Repeat("x", i))
					sb.WriteString("}} ")
				}
				return sb.String()
			}()},
			expected: expected{result: func() map[string]bool {
				m := make(map[string]bool)
				for i := 0; i < 100; i++ {
					m["var"+strings.Repeat("x", i)] = true
				}
				return m
			}()},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "surrounding emoji does not affect variable extraction",
			args:     args{content: "🎉 {{greeting}} 🎉"},
			expected: expected{result: map[string]bool{"greeting": true}},
		},
		{
			testName: "Japanese text around variable",
			args:     args{content: "こんにちは{{name}}さん"},
			expected: expected{result: map[string]bool{"name": true}},
		},
		{
			testName: "SQL injection around variable",
			args:     args{content: "'; DROP TABLE {{table_name}};--"},
			expected: expected{result: map[string]bool{"table_name": true}},
		},
		{
			testName: "HTML around variable",
			args:     args{content: "<div>{{content}}</div>"},
			expected: expected{result: map[string]bool{"content": true}},
		},
		{
			testName: "newlines between variables",
			args:     args{content: "{{first}}\n{{second}}\n{{third}}"},
			expected: expected{result: map[string]bool{"first": true, "second": true, "third": true}},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty string",
			args:     args{content: ""},
			expected: expected{result: map[string]bool{}},
		},
		{
			testName: "whitespace only",
			args:     args{content: "   \t\n  "},
			expected: expected{result: map[string]bool{}},
		},
		{
			testName: "empty braces",
			args:     args{content: "{{}}"},
			expected: expected{result: map[string]bool{}},
		},

		// Null/Nil - Go strings cannot be nil, test zero value
		{
			testName: "zero value empty string",
			args:     args{content: string("")},
			expected: expected{result: map[string]bool{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := FindVariables(tt.args.content)
			if diff := cmp.Diff(tt.expected.result, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
