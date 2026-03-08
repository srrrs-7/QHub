package env

import (
	"os"
	"strings"
	"testing"
)

func TestGetString(t *testing.T) {
	type args struct {
		key        string
		setupEnv   map[string]string
		cleanupEnv []string
	}

	type expected struct {
		value   string
		wantErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系: Valid environment variables
		{
			testName: "valid env var with normal string",
			args: args{
				key:      "TEST_STRING_VAR",
				setupEnv: map[string]string{"TEST_STRING_VAR": "hello world"},
			},
			expected: expected{value: "hello world", wantErr: false},
		},
		{
			testName: "valid env var with single character",
			args: args{
				key:      "TEST_SINGLE_CHAR",
				setupEnv: map[string]string{"TEST_SINGLE_CHAR": "a"},
			},
			expected: expected{value: "a", wantErr: false},
		},
		{
			testName: "valid env var with numbers in string",
			args: args{
				key:      "TEST_STRING_WITH_NUMBERS",
				setupEnv: map[string]string{"TEST_STRING_WITH_NUMBERS": "version123"},
			},
			expected: expected{value: "version123", wantErr: false},
		},

		// 異常系: Invalid cases
		{
			testName: "env var not found",
			args: args{
				key:        "TEST_NONEXISTENT_VAR",
				setupEnv:   map[string]string{},
				cleanupEnv: []string{"TEST_NONEXISTENT_VAR"},
			},
			expected: expected{wantErr: true},
		},

		// 境界値: Boundary values
		{
			testName: "very long string (5000 chars)",
			args: args{
				key:      "TEST_LONG_STRING",
				setupEnv: map[string]string{"TEST_LONG_STRING": generateLongString(5000)},
			},
			expected: expected{value: generateLongString(5000), wantErr: false},
		},
		{
			testName: "string with numbers at boundaries",
			args: args{
				key:      "TEST_BOUNDARY_NUMBERS",
				setupEnv: map[string]string{"TEST_BOUNDARY_NUMBERS": "0123456789"},
			},
			expected: expected{value: "0123456789", wantErr: false},
		},

		// 特殊文字: Special characters
		{
			testName: "string with emoji",
			args: args{
				key:      "TEST_EMOJI_STRING",
				setupEnv: map[string]string{"TEST_EMOJI_STRING": "Task 📋 Done ✓"},
			},
			expected: expected{value: "Task 📋 Done ✓", wantErr: false},
		},
		{
			testName: "string with Japanese characters",
			args: args{
				key:      "TEST_JAPANESE_STRING",
				setupEnv: map[string]string{"TEST_JAPANESE_STRING": "タスク 完了"},
			},
			expected: expected{value: "タスク 完了", wantErr: false},
		},
		{
			testName: "string with special symbols",
			args: args{
				key:      "TEST_SPECIAL_SYMBOLS",
				setupEnv: map[string]string{"TEST_SPECIAL_SYMBOLS": "!@#$%^&*()_+-=[]{}|;:,.<>?"},
			},
			expected: expected{value: "!@#$%^&*()_+-=[]{}|;:,.<>?", wantErr: false},
		},
		{
			testName: "string with newlines and tabs",
			args: args{
				key:      "TEST_WHITESPACE_SPECIAL",
				setupEnv: map[string]string{"TEST_WHITESPACE_SPECIAL": "line1\nline2\ttab"},
			},
			expected: expected{value: "line1\nline2\ttab", wantErr: false},
		},

		// 空文字: Empty string
		{
			testName: "empty string env var",
			args: args{
				key:      "TEST_EMPTY_STRING",
				setupEnv: map[string]string{"TEST_EMPTY_STRING": ""},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "whitespace only string",
			args: args{
				key:      "TEST_WHITESPACE_ONLY",
				setupEnv: map[string]string{"TEST_WHITESPACE_ONLY": "   "},
			},
			expected: expected{value: "   ", wantErr: false},
		},

		// Nil/Unset: Variable not set at all
		{
			testName: "completely unset env var (Nil case)",
			args: args{
				key:        "TEST_COMPLETELY_UNSET",
				setupEnv:   map[string]string{},
				cleanupEnv: []string{"TEST_COMPLETELY_UNSET"},
			},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			for k, v := range tt.args.setupEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			for _, k := range tt.args.cleanupEnv {
				os.Unsetenv(k)
				defer os.Unsetenv(k)
			}

			value, err := GetString(tt.args.key)

			if tt.expected.wantErr {
				if err == nil {
					t.Errorf("GetString(%q) expected error, got value: %q", tt.args.key, value)
				}
			} else {
				if err != nil {
					t.Errorf("GetString(%q) unexpected error: %v", tt.args.key, err)
				}
				if value != tt.expected.value {
					t.Errorf("GetString(%q) = %q, want %q", tt.args.key, value, tt.expected.value)
				}
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	type args struct {
		key        string
		setupEnv   map[string]string
		cleanupEnv []string
	}

	type expected struct {
		value   int
		wantErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系: Valid integer values
		{
			testName: "valid positive integer",
			args: args{
				key:      "TEST_INT_POSITIVE",
				setupEnv: map[string]string{"TEST_INT_POSITIVE": "42"},
			},
			expected: expected{value: 42, wantErr: false},
		},
		{
			testName: "valid negative integer",
			args: args{
				key:      "TEST_INT_NEGATIVE",
				setupEnv: map[string]string{"TEST_INT_NEGATIVE": "-15"},
			},
			expected: expected{value: -15, wantErr: false},
		},
		{
			testName: "zero value",
			args: args{
				key:      "TEST_INT_ZERO",
				setupEnv: map[string]string{"TEST_INT_ZERO": "0"},
			},
			expected: expected{value: 0, wantErr: false},
		},
		{
			testName: "single digit",
			args: args{
				key:      "TEST_INT_SINGLE",
				setupEnv: map[string]string{"TEST_INT_SINGLE": "5"},
			},
			expected: expected{value: 5, wantErr: false},
		},

		// 異常系: Invalid integer formats
		{
			testName: "invalid format: string with letters",
			args: args{
				key:      "TEST_INT_LETTERS",
				setupEnv: map[string]string{"TEST_INT_LETTERS": "abc123"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "digits followed by text",
			args: args{
				key:      "TEST_INT_DIGITS_TEXT",
				setupEnv: map[string]string{"TEST_INT_DIGITS_TEXT": "123abc"},
			},
			expected: expected{value: 123, wantErr: false},
		},
		{
			testName: "leading zero followed by x",
			args: args{
				key:      "TEST_INT_ZERO_X",
				setupEnv: map[string]string{"TEST_INT_ZERO_X": "0x10"},
			},
			expected: expected{value: 0, wantErr: false},
		},
		{
			testName: "invalid format: non-numeric start",
			args: args{
				key:      "TEST_INT_NON_NUMERIC",
				setupEnv: map[string]string{"TEST_INT_NON_NUMERIC": "a123"},
			},
			expected: expected{wantErr: true},
		},

		// 境界値: Boundary values
		{
			testName: "very large positive number",
			args: args{
				key:      "TEST_INT_LARGE_POS",
				setupEnv: map[string]string{"TEST_INT_LARGE_POS": "9223372036854775800"},
			},
			expected: expected{value: 9223372036854775800, wantErr: false},
		},
		{
			testName: "very large negative number",
			args: args{
				key:      "TEST_INT_LARGE_NEG",
				setupEnv: map[string]string{"TEST_INT_LARGE_NEG": "-9223372036854775800"},
			},
			expected: expected{value: -9223372036854775800, wantErr: false},
		},
		{
			testName: "integer with leading zeros",
			args: args{
				key:      "TEST_INT_LEADING_ZEROS",
				setupEnv: map[string]string{"TEST_INT_LEADING_ZEROS": "00042"},
			},
			expected: expected{value: 42, wantErr: false},
		},
		{
			testName: "negative number with plus sign prefix (invalid)",
			args: args{
				key:      "TEST_INT_PLUS_NEG",
				setupEnv: map[string]string{"TEST_INT_PLUS_NEG": "-+42"},
			},
			expected: expected{wantErr: true},
		},

		// 特殊文字: Special characters
		{
			testName: "emoji before digits (invalid)",
			args: args{
				key:      "TEST_INT_EMOJI_FIRST",
				setupEnv: map[string]string{"TEST_INT_EMOJI_FIRST": "📋42"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "Japanese characters before digits (invalid)",
			args: args{
				key:      "TEST_INT_JAPANESE_FIRST",
				setupEnv: map[string]string{"TEST_INT_JAPANESE_FIRST": "タスク42"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "special symbols before digits (invalid)",
			args: args{
				key:      "TEST_INT_SYMBOLS_FIRST",
				setupEnv: map[string]string{"TEST_INT_SYMBOLS_FIRST": "!@#42"},
			},
			expected: expected{wantErr: true},
		},

		// 空文字: Empty string
		{
			testName: "empty string env var",
			args: args{
				key:      "TEST_INT_EMPTY",
				setupEnv: map[string]string{"TEST_INT_EMPTY": ""},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "whitespace only",
			args: args{
				key:      "TEST_INT_WHITESPACE",
				setupEnv: map[string]string{"TEST_INT_WHITESPACE": "   "},
			},
			expected: expected{wantErr: true},
		},

		// Nil/Unset: Variable not set
		{
			testName: "completely unset env var",
			args: args{
				key:        "TEST_INT_UNSET",
				setupEnv:   map[string]string{},
				cleanupEnv: []string{"TEST_INT_UNSET"},
			},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			for k, v := range tt.args.setupEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			for _, k := range tt.args.cleanupEnv {
				os.Unsetenv(k)
				defer os.Unsetenv(k)
			}

			value, err := GetInt(tt.args.key)

			if tt.expected.wantErr {
				if err == nil {
					t.Errorf("GetInt(%q) expected error, got value: %d", tt.args.key, value)
				}
			} else {
				if err != nil {
					t.Errorf("GetInt(%q) unexpected error: %v", tt.args.key, err)
				}
				if value != tt.expected.value {
					t.Errorf("GetInt(%q) = %d, want %d", tt.args.key, value, tt.expected.value)
				}
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	type args struct {
		key        string
		setupEnv   map[string]string
		cleanupEnv []string
	}

	type expected struct {
		value   bool
		wantErr bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系: Valid boolean values
		{
			testName: "true value (lowercase)",
			args: args{
				key:      "TEST_BOOL_TRUE_LOWER",
				setupEnv: map[string]string{"TEST_BOOL_TRUE_LOWER": "true"},
			},
			expected: expected{value: true, wantErr: false},
		},
		{
			testName: "true value (uppercase)",
			args: args{
				key:      "TEST_BOOL_TRUE_UPPER",
				setupEnv: map[string]string{"TEST_BOOL_TRUE_UPPER": "TRUE"},
			},
			expected: expected{value: true, wantErr: false},
		},
		{
			testName: "true value (mixed case)",
			args: args{
				key:      "TEST_BOOL_TRUE_MIXED",
				setupEnv: map[string]string{"TEST_BOOL_TRUE_MIXED": "True"},
			},
			expected: expected{value: true, wantErr: false},
		},
		{
			testName: "1 value for true",
			args: args{
				key:      "TEST_BOOL_ONE",
				setupEnv: map[string]string{"TEST_BOOL_ONE": "1"},
			},
			expected: expected{value: true, wantErr: false},
		},
		{
			testName: "yes value (lowercase)",
			args: args{
				key:      "TEST_BOOL_YES_LOWER",
				setupEnv: map[string]string{"TEST_BOOL_YES_LOWER": "yes"},
			},
			expected: expected{value: true, wantErr: false},
		},
		{
			testName: "yes value (uppercase)",
			args: args{
				key:      "TEST_BOOL_YES_UPPER",
				setupEnv: map[string]string{"TEST_BOOL_YES_UPPER": "YES"},
			},
			expected: expected{value: true, wantErr: false},
		},
		{
			testName: "false value (lowercase)",
			args: args{
				key:      "TEST_BOOL_FALSE_LOWER",
				setupEnv: map[string]string{"TEST_BOOL_FALSE_LOWER": "false"},
			},
			expected: expected{value: false, wantErr: false},
		},
		{
			testName: "false value (uppercase)",
			args: args{
				key:      "TEST_BOOL_FALSE_UPPER",
				setupEnv: map[string]string{"TEST_BOOL_FALSE_UPPER": "FALSE"},
			},
			expected: expected{value: false, wantErr: false},
		},
		{
			testName: "0 value for false",
			args: args{
				key:      "TEST_BOOL_ZERO",
				setupEnv: map[string]string{"TEST_BOOL_ZERO": "0"},
			},
			expected: expected{value: false, wantErr: false},
		},
		{
			testName: "no value (lowercase)",
			args: args{
				key:      "TEST_BOOL_NO_LOWER",
				setupEnv: map[string]string{"TEST_BOOL_NO_LOWER": "no"},
			},
			expected: expected{value: false, wantErr: false},
		},
		{
			testName: "no value (uppercase)",
			args: args{
				key:      "TEST_BOOL_NO_UPPER",
				setupEnv: map[string]string{"TEST_BOOL_NO_UPPER": "NO"},
			},
			expected: expected{value: false, wantErr: false},
		},

		// 異常系: Invalid boolean formats
		{
			testName: "invalid value: random string",
			args: args{
				key:      "TEST_BOOL_INVALID_STRING",
				setupEnv: map[string]string{"TEST_BOOL_INVALID_STRING": "maybe"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "invalid value: on/off",
			args: args{
				key:      "TEST_BOOL_ON_OFF",
				setupEnv: map[string]string{"TEST_BOOL_ON_OFF": "on"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "invalid value: numeric other than 0/1",
			args: args{
				key:      "TEST_BOOL_NUMERIC_2",
				setupEnv: map[string]string{"TEST_BOOL_NUMERIC_2": "2"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "invalid value: negative number",
			args: args{
				key:      "TEST_BOOL_NEGATIVE",
				setupEnv: map[string]string{"TEST_BOOL_NEGATIVE": "-1"},
			},
			expected: expected{wantErr: true},
		},

		// 境界値: Boundary values
		{
			testName: "true with extra whitespace around",
			args: args{
				key:      "TEST_BOOL_SPACES_AROUND",
				setupEnv: map[string]string{"TEST_BOOL_SPACES_AROUND": "  true  "},
			},
			expected: expected{wantErr: true},
		},

		// 特殊文字: Special characters
		{
			testName: "boolean with emoji",
			args: args{
				key:      "TEST_BOOL_EMOJI",
				setupEnv: map[string]string{"TEST_BOOL_EMOJI": "true📋"},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "boolean with Japanese characters",
			args: args{
				key:      "TEST_BOOL_JAPANESE",
				setupEnv: map[string]string{"TEST_BOOL_JAPANESE": "trueタスク"},
			},
			expected: expected{wantErr: true},
		},

		// 空文字: Empty string
		{
			testName: "empty string env var",
			args: args{
				key:      "TEST_BOOL_EMPTY",
				setupEnv: map[string]string{"TEST_BOOL_EMPTY": ""},
			},
			expected: expected{wantErr: true},
		},
		{
			testName: "whitespace only",
			args: args{
				key:      "TEST_BOOL_WHITESPACE",
				setupEnv: map[string]string{"TEST_BOOL_WHITESPACE": "   "},
			},
			expected: expected{wantErr: true},
		},

		// Nil/Unset: Variable not set
		{
			testName: "completely unset env var",
			args: args{
				key:        "TEST_BOOL_UNSET",
				setupEnv:   map[string]string{},
				cleanupEnv: []string{"TEST_BOOL_UNSET"},
			},
			expected: expected{wantErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			for k, v := range tt.args.setupEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			for _, k := range tt.args.cleanupEnv {
				os.Unsetenv(k)
				defer os.Unsetenv(k)
			}

			value, err := GetBool(tt.args.key)

			if tt.expected.wantErr {
				if err == nil {
					t.Errorf("GetBool(%q) expected error, got value: %v", tt.args.key, value)
				}
			} else {
				if err != nil {
					t.Errorf("GetBool(%q) unexpected error: %v", tt.args.key, err)
				}
				if value != tt.expected.value {
					t.Errorf("GetBool(%q) = %v, want %v", tt.args.key, value, tt.expected.value)
				}
			}
		})
	}
}

// Helper function to generate long strings for boundary testing
func generateLongString(length int) string {
	var result strings.Builder
	for range length {
		result.WriteString("a")
	}
	return result.String()
}
