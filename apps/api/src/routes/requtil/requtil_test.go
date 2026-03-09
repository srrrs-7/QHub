package requtil

import (
	"api/src/domain/apperror"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// testRequest is a test struct with validation tags for Decode tests.
type testRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=150"`
}

// testSanitize strips HTML from the Name field.
func testSanitize(r *testRequest) {
	r.Name = Sanitize.Sanitize(r.Name)
}

// --- Decode ---

func TestDecode(t *testing.T) {
	type args struct {
		body     string
		sanitize func(*testRequest)
	}
	type expected struct {
		wantErr bool
		errName string
		result  testRequest
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "valid JSON with all fields",
			args: args{
				body:     `{"name":"Alice","email":"alice@example.com","age":30}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Alice", Email: "alice@example.com", Age: 30},
			},
		},
		{
			testName: "valid JSON with sanitize function",
			args: args{
				body:     `{"name":"<b>Bob</b>","email":"bob@example.com","age":25}`,
				sanitize: testSanitize,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Bob", Email: "bob@example.com", Age: 25},
			},
		},
		{
			testName: "valid JSON with nil sanitize function",
			args: args{
				body:     `{"name":"Charlie","email":"charlie@example.com","age":0}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Charlie", Email: "charlie@example.com", Age: 0},
			},
		},

		// 異常系 (Error Cases)
		{
			testName: "invalid JSON syntax",
			args: args{
				body:     `{invalid json}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.BadRequestErrorName,
			},
		},
		{
			testName: "missing required field name",
			args: args{
				body:     `{"email":"test@example.com","age":20}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},
		{
			testName: "missing required field email",
			args: args{
				body:     `{"name":"Alice","age":20}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},
		{
			testName: "invalid email format",
			args: args{
				body:     `{"name":"Alice","email":"not-an-email","age":20}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},

		// 境界値 (Boundary Values)
		{
			testName: "age at minimum boundary zero",
			args: args{
				body:     `{"name":"Min","email":"min@example.com","age":0}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Min", Email: "min@example.com", Age: 0},
			},
		},
		{
			testName: "age at maximum boundary 150",
			args: args{
				body:     `{"name":"Max","email":"max@example.com","age":150}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Max", Email: "max@example.com", Age: 150},
			},
		},
		{
			testName: "age exceeds maximum boundary",
			args: args{
				body:     `{"name":"Over","email":"over@example.com","age":151}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},
		{
			testName: "age below minimum boundary negative",
			args: args{
				body:     `{"name":"Under","email":"under@example.com","age":-1}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},
		{
			testName: "name at min length 1 char",
			args: args{
				body:     `{"name":"A","email":"a@example.com","age":1}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "A", Email: "a@example.com", Age: 1},
			},
		},
		{
			testName: "name at max length 100 chars",
			args: args{
				body:     fmt.Sprintf(`{"name":"%s","email":"long@example.com","age":1}`, strings.Repeat("a", 100)),
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: strings.Repeat("a", 100), Email: "long@example.com", Age: 1},
			},
		},
		{
			testName: "name over max length 101 chars",
			args: args{
				body:     fmt.Sprintf(`{"name":"%s","email":"long@example.com","age":1}`, strings.Repeat("a", 101)),
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "name with emoji",
			args: args{
				body:     `{"name":"Alice 🎉","email":"alice@example.com","age":25}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Alice 🎉", Email: "alice@example.com", Age: 25},
			},
		},
		{
			testName: "name with Japanese characters",
			args: args{
				body:     `{"name":"太郎","email":"taro@example.com","age":30}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "太郎", Email: "taro@example.com", Age: 30},
			},
		},
		{
			testName: "name with HTML sanitized by function",
			args: args{
				body:     `{"name":"<script>alert('xss')</script>Safe","email":"xss@example.com","age":20}`,
				sanitize: testSanitize,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Safe", Email: "xss@example.com", Age: 20},
			},
		},
		{
			testName: "name with SQL injection characters",
			args: args{
				body:     `{"name":"Robert'; DROP TABLE--","email":"sql@example.com","age":20}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: false,
				result:  testRequest{Name: "Robert'; DROP TABLE--", Email: "sql@example.com", Age: 20},
			},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty body",
			args: args{
				body:     ``,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.BadRequestErrorName,
			},
		},
		{
			testName: "empty JSON object fails validation",
			args: args{
				body:     `{}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},

		// Null/Nil
		{
			testName: "JSON null values for strings",
			args: args{
				body:     `{"name":null,"email":null,"age":0}`,
				sanitize: nil,
			},
			expected: expected{
				wantErr: true,
				errName: apperror.ValidationErrorName,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.args.body))
			req.Header.Set("Content-Type", "application/json")

			got, err := Decode[testRequest](req, tt.args.sanitize)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Errorf("expected AppError but got: %T", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.expected.result, got); diff != "" {
					t.Errorf("result mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// --- ParseUUID ---

func TestParseUUID(t *testing.T) {
	validUUID := uuid.New()

	type args struct {
		paramName  string
		paramValue string
	}
	type expected struct {
		wantErr bool
		errName string
		id      uuid.UUID
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "valid UUID",
			args:     args{paramName: "id", paramValue: validUUID.String()},
			expected: expected{wantErr: false, id: validUUID},
		},
		{
			testName: "valid UUID with different param name",
			args:     args{paramName: "project_id", paramValue: validUUID.String()},
			expected: expected{wantErr: false, id: validUUID},
		},
		{
			testName: "valid nil UUID",
			args:     args{paramName: "id", paramValue: uuid.Nil.String()},
			expected: expected{wantErr: false, id: uuid.Nil},
		},

		// 異常系 (Error Cases)
		{
			testName: "invalid UUID format",
			args:     args{paramName: "id", paramValue: "not-a-uuid"},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "UUID with wrong length",
			args:     args{paramName: "id", paramValue: "12345678-1234-1234-1234"},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "UUID with invalid hex characters",
			args:     args{paramName: "id", paramValue: "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// 境界値 (Boundary Values)
		{
			testName: "max UUID value",
			args:     args{paramName: "id", paramValue: "ffffffff-ffff-ffff-ffff-ffffffffffff"},
			expected: expected{wantErr: false, id: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")},
		},
		{
			testName: "min UUID value all zeros",
			args:     args{paramName: "id", paramValue: "00000000-0000-0000-0000-000000000000"},
			expected: expected{wantErr: false, id: uuid.Nil},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "UUID-like string with emoji",
			args:     args{paramName: "id", paramValue: "🎉🎉🎉🎉-🎉🎉🎉🎉-🎉🎉🎉🎉-🎉🎉🎉🎉"},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "SQL injection as UUID",
			args:     args{paramName: "id", paramValue: "'; DROP TABLE users;--"},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "Japanese characters as UUID",
			args:     args{paramName: "id", paramValue: "テスト"},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty string",
			args:     args{paramName: "id", paramValue: ""},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "whitespace only",
			args:     args{paramName: "id", paramValue: "   "},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// Null/Nil - param not set in route context
		{
			testName: "param not set in context",
			args:     args{paramName: "missing_param", paramValue: ""},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(tt.args.paramName, tt.args.paramValue)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			got, err := ParseUUID(req, tt.args.paramName)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Errorf("expected AppError but got: %T", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.expected.id, got); diff != "" {
					t.Errorf("UUID mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// --- ValidateParams ---

type testParams struct {
	Page    int    `validate:"gte=1,lte=1000"`
	PerPage int    `validate:"gte=1,lte=100"`
	Sort    string `validate:"required,oneof=asc desc"`
}

func TestValidateParams(t *testing.T) {
	type args struct {
		params testParams
	}
	type expected struct {
		wantErr bool
		errName string
		result  testParams
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "valid params",
			args:     args{params: testParams{Page: 1, PerPage: 10, Sort: "asc"}},
			expected: expected{wantErr: false, result: testParams{Page: 1, PerPage: 10, Sort: "asc"}},
		},
		{
			testName: "valid params with desc sort",
			args:     args{params: testParams{Page: 50, PerPage: 50, Sort: "desc"}},
			expected: expected{wantErr: false, result: testParams{Page: 50, PerPage: 50, Sort: "desc"}},
		},

		// 異常系 (Error Cases)
		{
			testName: "invalid sort value",
			args:     args{params: testParams{Page: 1, PerPage: 10, Sort: "invalid"}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "page below minimum",
			args:     args{params: testParams{Page: 0, PerPage: 10, Sort: "asc"}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// 境界値 (Boundary Values)
		{
			testName: "page at min boundary 1",
			args:     args{params: testParams{Page: 1, PerPage: 1, Sort: "asc"}},
			expected: expected{wantErr: false, result: testParams{Page: 1, PerPage: 1, Sort: "asc"}},
		},
		{
			testName: "page at max boundary 1000",
			args:     args{params: testParams{Page: 1000, PerPage: 100, Sort: "asc"}},
			expected: expected{wantErr: false, result: testParams{Page: 1000, PerPage: 100, Sort: "asc"}},
		},
		{
			testName: "page over max boundary 1001",
			args:     args{params: testParams{Page: 1001, PerPage: 10, Sort: "asc"}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "per_page over max boundary 101",
			args:     args{params: testParams{Page: 1, PerPage: 101, Sort: "asc"}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "sort with emoji",
			args:     args{params: testParams{Page: 1, PerPage: 10, Sort: "🎉"}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "sort with Japanese",
			args:     args{params: testParams{Page: 1, PerPage: 10, Sort: "昇順"}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty sort string",
			args:     args{params: testParams{Page: 1, PerPage: 10, Sort: ""}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
		{
			testName: "whitespace sort string",
			args:     args{params: testParams{Page: 1, PerPage: 10, Sort: " "}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},

		// Null/Nil (zero values)
		{
			testName: "zero value struct",
			args:     args{params: testParams{}},
			expected: expected{wantErr: true, errName: apperror.ValidationErrorName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := ValidateParams(tt.args.params)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Errorf("expected AppError but got: %T", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.expected.result, got); diff != "" {
					t.Errorf("result mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// --- MergeField ---

func TestMergeField(t *testing.T) {
	type args struct {
		existing    string
		raw         string
		constructor func(string) (string, error)
	}
	type expected struct {
		wantErr bool
		result  string
	}

	successConstructor := func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}
	failConstructor := func(s string) (string, error) {
		return "", fmt.Errorf("constructor error for: %s", s)
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "non-empty raw replaces existing",
			args: args{
				existing:    "old",
				raw:         "new",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "NEW"},
		},
		{
			testName: "constructor transforms value",
			args: args{
				existing:    "original",
				raw:         "hello world",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "HELLO WORLD"},
		},

		// 異常系 (Error Cases)
		{
			testName: "constructor returns error",
			args: args{
				existing:    "old",
				raw:         "bad-input",
				constructor: failConstructor,
			},
			expected: expected{wantErr: true},
		},

		// 境界値 (Boundary Values)
		{
			testName: "single character raw",
			args: args{
				existing:    "old",
				raw:         "x",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "X"},
		},
		{
			testName: "very long raw string",
			args: args{
				existing:    "old",
				raw:         strings.Repeat("a", 10000),
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: strings.ToUpper(strings.Repeat("a", 10000))},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "raw with emoji",
			args: args{
				existing:    "old",
				raw:         "hello 🎉",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "HELLO 🎉"},
		},
		{
			testName: "raw with Japanese characters",
			args: args{
				existing:    "old",
				raw:         "テスト",
				constructor: func(s string) (string, error) { return s, nil },
			},
			expected: expected{wantErr: false, result: "テスト"},
		},
		{
			testName: "raw with SQL injection",
			args: args{
				existing:    "old",
				raw:         "'; DROP TABLE--",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "'; DROP TABLE--"},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "empty raw returns existing",
			args: args{
				existing:    "keep-me",
				raw:         "",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "keep-me"},
		},
		{
			testName: "empty raw with empty existing",
			args: args{
				existing:    "",
				raw:         "",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: ""},
		},
		{
			testName: "whitespace-only raw is treated as non-empty",
			args: args{
				existing:    "old",
				raw:         "   ",
				constructor: successConstructor,
			},
			expected: expected{wantErr: false, result: "   "},
		},

		// Null/Nil - constructor is always provided but raw empty triggers existing return
		{
			testName: "empty raw with nil-like zero value existing",
			args: args{
				existing:    "",
				raw:         "",
				constructor: failConstructor,
			},
			expected: expected{wantErr: false, result: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := MergeField(tt.args.existing, tt.args.raw, tt.args.constructor)

			if tt.expected.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tt.expected.result, got); diff != "" {
					t.Errorf("result mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
