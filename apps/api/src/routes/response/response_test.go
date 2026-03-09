package response

import (
	"api/src/domain/apperror"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testResponse struct {
	Message string `json:"message,omitempty"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Count   int    `json:"count,omitempty"`
}

// customAppError is a test error type that doesn't match any known error name
type customAppError struct {
	errName    string
	domainName string
	err        error
}

func (e customAppError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return e.errName + " [" + e.domainName + "]"
}

func (e customAppError) ErrorName() string {
	return e.errName
}

func (e customAppError) DomainName() string {
	return e.domainName
}

func (e customAppError) Unwrap() error {
	return e.err
}

// unencodableType is a type that cannot be JSON encoded (contains channels)
type unencodableType struct {
	Ch chan int
}

func TestOK(t *testing.T) {
	type args struct {
		body any
	}
	type expected struct {
		statusCode  int
		contentType string
		bodyCheck   func(*testing.T, *http.Response)
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "valid JSON response with message",
			args: args{
				body: testResponse{Message: "success"},
			},
			expected: expected{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result map[string]string
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if diff := cmp.Diff(map[string]string{"message": "success"}, result); diff != "" {
						t.Errorf("body mismatch (-want +got):\n%s", diff)
					}
				},
			},
		},
		{
			testName: "valid JSON response with ID",
			args: args{
				body: testResponse{ID: "123", Name: "Test"},
			},
			expected: expected{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if result.ID != "123" || result.Name != "Test" {
						t.Errorf("unexpected response: %+v", result)
					}
				},
			},
		},
		// 異常系: Empty body
		{
			testName: "empty struct body",
			args: args{
				body: testResponse{},
			},
			expected: expected{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
				},
			},
		},
		// 境界値: Large response
		{
			testName: "large response body",
			args: args{
				body: testResponse{Message: strings.Repeat("a", 1000)},
			},
			expected: expected{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if len(result.Message) != 1000 {
						t.Errorf("expected message length 1000, got %d", len(result.Message))
					}
				},
			},
		},
		// 特殊文字: Unicode and special characters
		{
			testName: "response with Unicode characters",
			args: args{
				body: testResponse{Message: "成功 ✓ 日本語"},
			},
			expected: expected{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if result.Message != "成功 ✓ 日本語" {
						t.Errorf("Unicode not preserved: got %s", result.Message)
					}
				},
			},
		},
		// Nil: nil body (will be encoded as null)
		{
			testName: "nil body",
			args: args{
				body: nil,
			},
			expected: expected{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result any
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
				},
			},
		},
		// 異常系: JSON encoding error (status code is written before encoding)
		{
			testName: "unencodable body",
			args: args{
				body: unencodableType{Ch: make(chan int)},
			},
			expected: expected{
				statusCode:  http.StatusOK, // Status is written before encoding attempt
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					// When encoding fails, http.Error writes error message
					// Status code is already set to 200, so encoding error appears in body
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			OK(w, tt.args.body)

			resp := w.Result()
			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expected.contentType, resp.Header.Get("Content-Type")); diff != "" {
				t.Errorf("Content-Type mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.bodyCheck != nil {
				tt.expected.bodyCheck(t, resp)
			}
		})
	}
}

func TestCreated(t *testing.T) {
	type args struct {
		body any
	}
	type expected struct {
		statusCode  int
		contentType string
		bodyCheck   func(*testing.T, *http.Response)
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "resource created with ID",
			args: args{
				body: testResponse{ID: "123"},
			},
			expected: expected{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result map[string]string
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if diff := cmp.Diff(map[string]string{"id": "123"}, result); diff != "" {
						t.Errorf("body mismatch (-want +got):\n%s", diff)
					}
				},
			},
		},
		{
			testName: "resource created with full object",
			args: args{
				body: testResponse{ID: "456", Name: "NewResource", Message: "Created successfully"},
			},
			expected: expected{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if result.ID != "456" || result.Name != "NewResource" {
						t.Errorf("unexpected response: %+v", result)
					}
				},
			},
		},
		// 異常系: Empty body
		{
			testName: "empty response body",
			args: args{
				body: testResponse{},
			},
			expected: expected{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
				},
			},
		},
		// 境界値: Large ID
		{
			testName: "resource with very long ID",
			args: args{
				body: testResponse{ID: strings.Repeat("a", 500)},
			},
			expected: expected{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if len(result.ID) != 500 {
						t.Errorf("expected ID length 500, got %d", len(result.ID))
					}
				},
			},
		},
		// 特殊文字: Unicode ID
		{
			testName: "resource with Unicode ID",
			args: args{
				body: testResponse{ID: "リソース-123", Name: "日本語"},
			},
			expected: expected{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result testResponse
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if result.ID != "リソース-123" || result.Name != "日本語" {
						t.Errorf("Unicode not preserved: %+v", result)
					}
				},
			},
		},
		// Nil
		{
			testName: "nil body",
			args: args{
				body: nil,
			},
			expected: expected{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result any
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
				},
			},
		},
		// 異常系: JSON encoding error (status code is written before encoding)
		{
			testName: "unencodable body",
			args: args{
				body: unencodableType{Ch: make(chan int)},
			},
			expected: expected{
				statusCode:  http.StatusCreated, // Status is written before encoding attempt
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					// When encoding fails, http.Error writes error message
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			Created(w, tt.args.body)

			resp := w.Result()
			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expected.contentType, resp.Header.Get("Content-Type")); diff != "" {
				t.Errorf("Content-Type mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.bodyCheck != nil {
				tt.expected.bodyCheck(t, resp)
			}
		})
	}
}

func TestAccepted(t *testing.T) {
	tests := []struct {
		testName string
		expected struct {
			statusCode int
		}
	}{
		// 正常系
		{
			testName: "returns 202 Accepted",
			expected: struct{ statusCode int }{
				statusCode: http.StatusAccepted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			Accepted(w)

			if diff := cmp.Diff(tt.expected.statusCode, w.Code); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			// Verify no body is written
			if w.Body.Len() != 0 {
				t.Errorf("expected no content, got body length: %d", w.Body.Len())
			}
		})
	}
}

func TestNoContent(t *testing.T) {
	tests := []struct {
		testName string
		expected struct {
			statusCode int
		}
	}{
		// 正常系
		{
			testName: "returns 204 No Content",
			expected: struct{ statusCode int }{
				statusCode: http.StatusNoContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			NoContent(w)

			if diff := cmp.Diff(tt.expected.statusCode, w.Code); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			// Verify no body is written
			if w.Body.Len() != 0 {
				t.Errorf("expected no content, got body length: %d", w.Body.Len())
			}
		})
	}
}

func TestHandleAppError(t *testing.T) {
	type args struct {
		err apperror.AppError
	}
	type expected struct {
		statusCode  int
		contentType string
		body        map[string]string
		bodyCheck   func(*testing.T, *http.Response)
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - ValidationError maps to BadRequest
		{
			testName: "ValidationError returns 400 Bad Request",
			args: args{
				err: apperror.NewValidationError(nil, "TaskDomain"),
			},
			expected: expected{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.ValidationErrorName,
					"domain":  "TaskDomain",
					"message": "ValidationError [TaskDomain]",
				},
			},
		},
		{
			testName: "NotFoundError returns 404 Not Found",
			args: args{
				err: apperror.NewNotFoundError(nil, "TaskDomain"),
			},
			expected: expected{
				statusCode:  http.StatusNotFound,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.NotFoundErrorName,
					"domain":  "TaskDomain",
					"message": "NotFoundError [TaskDomain]",
				},
			},
		},
		{
			testName: "UnauthorizedError returns 401 Unauthorized",
			args: args{
				err: apperror.NewUnauthorizedError(nil, "AuthDomain"),
			},
			expected: expected{
				statusCode:  http.StatusUnauthorized,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.UnauthorizedErrorName,
					"domain":  "AuthDomain",
					"message": "UnauthorizedError [AuthDomain]",
				},
			},
		},
		{
			testName: "ForbiddenError returns 403 Forbidden",
			args: args{
				err: apperror.NewForbiddenError(nil, "AuthDomain"),
			},
			expected: expected{
				statusCode:  http.StatusForbidden,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.ForbiddenErrorName,
					"domain":  "AuthDomain",
					"message": "ForbiddenError [AuthDomain]",
				},
			},
		},
		{
			testName: "BadRequestError returns 400 Bad Request",
			args: args{
				err: apperror.NewBadRequestError(nil, "RequestDomain"),
			},
			expected: expected{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.BadRequestErrorName,
					"domain":  "RequestDomain",
					"message": "BadRequestError [RequestDomain]",
				},
			},
		},
		{
			testName: "ConflictError returns 409 Conflict",
			args: args{
				err: apperror.NewConflictError(nil, "TaskDomain"),
			},
			expected: expected{
				statusCode:  http.StatusConflict,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.ConflictErrorName,
					"domain":  "TaskDomain",
					"message": "ConflictError [TaskDomain]",
				},
			},
		},
		{
			testName: "DatabaseError returns 500 Internal Server Error",
			args: args{
				err: apperror.NewDatabaseError(nil, "DBDomain"),
			},
			expected: expected{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.DatabaseErrorName,
					"domain":  "DBDomain",
					"message": "DatabaseError [DBDomain]",
				},
			},
		},
		{
			testName: "InternalServerError returns 500 Internal Server Error",
			args: args{
				err: apperror.NewInternalServerError(nil, "SystemDomain"),
			},
			expected: expected{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				body: map[string]string{
					"type":    apperror.InternalServerErrorName,
					"domain":  "SystemDomain",
					"message": "InternalServerError [SystemDomain]",
				},
			},
		},
		// 異常系 - Unknown error type triggers unexpectedError
		{
			testName: "Unknown error type returns 500 Internal Server Error",
			args: args{
				err: customAppError{
					errName:    "UnknownError",
					domainName: "TestDomain",
					err:        nil,
				},
			},
			expected: expected{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				body: map[string]string{
					"type":    "UnknownError",
					"domain":  "TestDomain",
					"message": "UnknownError [TestDomain]",
				},
			},
		},
		// 境界値 - Error with underlying error
		{
			testName: "ValidationError with underlying error",
			args: args{
				err: apperror.NewValidationError(
					fmt.Errorf("json unmarshal error"),
					"TaskDomain",
				),
			},
			expected: expected{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result map[string]string
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if result["type"] != apperror.ValidationErrorName {
						t.Errorf("expected type %s, got %s", apperror.ValidationErrorName, result["type"])
					}
					if result["domain"] != "TaskDomain" {
						t.Errorf("expected domain TaskDomain, got %s", result["domain"])
					}
					// Message should include underlying error
					if !strings.Contains(result["message"], "ValidationError") {
						t.Errorf("expected message to contain ValidationError, got %s", result["message"])
					}
				},
			},
		},
		// 特殊文字 - Error with Unicode in domain name
		{
			testName: "Error with Unicode domain name",
			args: args{
				err: apperror.NewNotFoundError(nil, "タスクドメイン"),
			},
			expected: expected{
				statusCode:  http.StatusNotFound,
				contentType: "application/json",
				bodyCheck: func(t *testing.T, resp *http.Response) {
					var result map[string]string
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						t.Fatalf("failed to decode response body: %v", err)
					}
					if result["domain"] != "タスクドメイン" {
						t.Errorf("Unicode domain not preserved: got %s", result["domain"])
					}
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.args.err)

			resp := w.Result()
			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expected.contentType, resp.Header.Get("Content-Type")); diff != "" {
				t.Errorf("Content-Type mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.bodyCheck != nil {
				tt.expected.bodyCheck(t, resp)
			} else if tt.expected.body != nil {
				var result map[string]string
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if diff := cmp.Diff(tt.expected.body, result); diff != "" {
					t.Errorf("body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestHandleErrorWithNonAppError(t *testing.T) {
	type args struct {
		err error
	}
	type expected struct {
		statusCode int
		bodyContains string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - plain error falls through to http.Error
		{
			testName: "plain error returns 500",
			args: args{
				err: fmt.Errorf("something went wrong"),
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "something went wrong",
			},
		},
		// 特殊文字 - Unicode error message
		{
			testName: "error with Unicode message",
			args: args{
				err: fmt.Errorf("エラーが発生しました"),
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "エラーが発生しました",
			},
		},
		// 境界値 - wrapped plain error (not AppError)
		{
			testName: "wrapped plain error returns 500",
			args: args{
				err: fmt.Errorf("outer: %w", fmt.Errorf("inner error")),
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "outer: inner error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.args.err)

			resp := w.Result()
			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.expected.bodyContains) {
				t.Errorf("expected body to contain %q, got %q", tt.expected.bodyContains, body)
			}
		})
	}
}

func TestMapSlice(t *testing.T) {
	type testCase[S any, D any] struct {
		testName string
		src      []S
		fn       func(S) D
		expected []D
	}

	t.Run("string to int conversion", func(t *testing.T) {
		tests := []testCase[string, int]{
			// 正常系
			{
				testName: "convert multiple items",
				src:      []string{"a", "bb", "ccc"},
				fn:       func(s string) int { return len(s) },
				expected: []int{1, 2, 3},
			},
			// 空文字 - empty slice
			{
				testName: "empty slice returns empty slice",
				src:      []string{},
				fn:       func(s string) int { return len(s) },
				expected: []int{},
			},
			// 境界値 - single element
			{
				testName: "single element slice",
				src:      []string{"hello"},
				fn:       func(s string) int { return len(s) },
				expected: []int{5},
			},
			// 特殊文字
			{
				testName: "Unicode strings",
				src:      []string{"日本", "🎉"},
				fn:       func(s string) int { return len(s) },
				expected: []int{6, 4}, // byte length
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				result := MapSlice(tt.src, tt.fn)
				if diff := cmp.Diff(tt.expected, result); diff != "" {
					t.Errorf("result mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("struct transformation", func(t *testing.T) {
		type input struct {
			Name string
			Age  int
		}
		type output struct {
			Label string
		}

		tests := []testCase[input, output]{
			// 正常系
			{
				testName: "transform structs",
				src:      []input{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}},
				fn:       func(i input) output { return output{Label: i.Name} },
				expected: []output{{Label: "Alice"}, {Label: "Bob"}},
			},
			// Nil - nil slice
			{
				testName: "nil slice returns empty slice",
				src:      nil,
				fn:       func(i input) output { return output{Label: i.Name} },
				expected: []output{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				result := MapSlice(tt.src, tt.fn)
				if diff := cmp.Diff(tt.expected, result); diff != "" {
					t.Errorf("result mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("identity transformation", func(t *testing.T) {
		// 境界値 - large slice
		src := make([]int, 1000)
		for i := range src {
			src[i] = i
		}
		result := MapSlice(src, func(i int) int { return i * 2 })
		if len(result) != 1000 {
			t.Errorf("expected length 1000, got %d", len(result))
		}
		if result[0] != 0 {
			t.Errorf("expected first element 0, got %d", result[0])
		}
		if result[999] != 1998 {
			t.Errorf("expected last element 1998, got %d", result[999])
		}
	})
}

func TestHandleAppErrorUnknownErrorName(t *testing.T) {
	tests := []struct {
		testName string
		err      apperror.AppError
		expected int
	}{
		// 異常系 - custom error name not in errorStatusMap
		{
			testName: "ServiceUnavailableError maps to 500",
			err: customAppError{
				errName:    "ServiceUnavailableError",
				domainName: "Search",
				err:        nil,
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "RateLimitError maps to 500",
			err: customAppError{
				errName:    "RateLimitError",
				domainName: "API",
				err:        nil,
			},
			expected: http.StatusInternalServerError,
		},
		// 空文字 - empty error name
		{
			testName: "empty error name maps to 500",
			err: customAppError{
				errName:    "",
				domainName: "Test",
				err:        nil,
			},
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.err)

			if diff := cmp.Diff(tt.expected, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
