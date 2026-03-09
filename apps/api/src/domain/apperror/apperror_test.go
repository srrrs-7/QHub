package apperror_test

import (
	"api/src/domain/apperror"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewNotFoundError tests the NewNotFoundError constructor
func TestNewNotFoundError(t *testing.T) {
	type args struct {
		err        error
		domainName string
	}
	type expected struct {
		errorName  string
		domainName string
		hasErr     bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "create NotFoundError with error",
			args:     args{err: errors.New("not found"), domainName: "Task"},
			expected: expected{errorName: "NotFoundError", domainName: "Task", hasErr: true},
		},
		{
			testName: "create NotFoundError with different domain",
			args:     args{err: errors.New("resource not found"), domainName: "User"},
			expected: expected{errorName: "NotFoundError", domainName: "User", hasErr: true},
		},

		// 📏 境界値
		{
			testName: "NotFoundError with nil underlying error",
			args:     args{err: nil, domainName: "Task"},
			expected: expected{errorName: "NotFoundError", domainName: "Task", hasErr: false},
		},

		// 🔤 特殊文字
		{
			testName: "NotFoundError with emoji in domain",
			args:     args{err: errors.New("not found"), domainName: "Task📋"},
			expected: expected{errorName: "NotFoundError", domainName: "Task📋", hasErr: true},
		},
		{
			testName: "NotFoundError with Japanese domain",
			args:     args{err: errors.New("見つかりません"), domainName: "タスク"},
			expected: expected{errorName: "NotFoundError", domainName: "タスク", hasErr: true},
		},

		// 📭 空文字
		{
			testName: "NotFoundError with empty domain name",
			args:     args{err: errors.New("not found"), domainName: ""},
			expected: expected{errorName: "NotFoundError", domainName: "", hasErr: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := apperror.NewNotFoundError(tt.args.err, tt.args.domainName)

			assert.Equal(t, tt.expected.errorName, result.ErrorName())
			assert.Equal(t, tt.expected.domainName, result.DomainName())

			if tt.expected.hasErr {
				assert.NotNil(t, result.Unwrap())
			} else {
				assert.Nil(t, result.Unwrap())
			}
		})
	}
}

// TestNewValidationError tests the NewValidationError constructor
func TestNewValidationError(t *testing.T) {
	type args struct {
		err        error
		domainName string
	}
	type expected struct {
		errorName  string
		domainName string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "create ValidationError",
			args:     args{err: errors.New("invalid input"), domainName: "Task"},
			expected: expected{errorName: "ValidationError", domainName: "Task"},
		},

		// 🔤 特殊文字
		{
			testName: "ValidationError with emoji in error",
			args:     args{err: errors.New("invalid 🔥"), domainName: "Task"},
			expected: expected{errorName: "ValidationError", domainName: "Task"},
		},

		// 📭 空文字
		{
			testName: "ValidationError with empty error message",
			args:     args{err: errors.New(""), domainName: "Task"},
			expected: expected{errorName: "ValidationError", domainName: "Task"},
		},

		// ⚠️ Nil
		{
			testName: "ValidationError with nil error",
			args:     args{err: nil, domainName: "Task"},
			expected: expected{errorName: "ValidationError", domainName: "Task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := apperror.NewValidationError(tt.args.err, tt.args.domainName)

			assert.Equal(t, tt.expected.errorName, result.ErrorName())
			assert.Equal(t, tt.expected.domainName, result.DomainName())
		})
	}
}

// TestNewDatabaseError tests the NewDatabaseError constructor
func TestNewDatabaseError(t *testing.T) {
	type args struct {
		err        error
		domainName string
	}
	type expected struct {
		errorName  string
		domainName string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "create DatabaseError",
			args:     args{err: errors.New("connection failed"), domainName: "TaskRepository"},
			expected: expected{errorName: "DatabaseError", domainName: "TaskRepository"},
		},

		// 📭 空文字
		{
			testName: "DatabaseError with empty domain",
			args:     args{err: errors.New("sql error"), domainName: ""},
			expected: expected{errorName: "DatabaseError", domainName: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := apperror.NewDatabaseError(tt.args.err, tt.args.domainName)

			assert.Equal(t, tt.expected.errorName, result.ErrorName())
			assert.Equal(t, tt.expected.domainName, result.DomainName())
		})
	}
}

// TestAllErrorConstructors tests all error type constructors
func TestAllErrorConstructors(t *testing.T) {
	testErr := errors.New("test error")
	domain := "TestDomain"

	constructors := map[string]struct {
		constructor func(error, string) apperror.AppError
		errorName   string
	}{
		"NewNotFoundError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewNotFoundError(e, d) },
			errorName:   "NotFoundError",
		},
		"NewValidationError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewValidationError(e, d) },
			errorName:   "ValidationError",
		},
		"NewUnauthorizedError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewUnauthorizedError(e, d) },
			errorName:   "UnauthorizedError",
		},
		"NewInternalServerError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewInternalServerError(e, d) },
			errorName:   "InternalServerError",
		},
		"NewBadRequestError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewBadRequestError(e, d) },
			errorName:   "BadRequestError",
		},
		"NewConflictError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewConflictError(e, d) },
			errorName:   "ConflictError",
		},
		"NewForbiddenError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewForbiddenError(e, d) },
			errorName:   "ForbiddenError",
		},
		"NewDatabaseError": {
			constructor: func(e error, d string) apperror.AppError { return apperror.NewDatabaseError(e, d) },
			errorName:   "DatabaseError",
		},
	}

	for name, tc := range constructors {
		t.Run(name, func(t *testing.T) {
			result := tc.constructor(testErr, domain)

			assert.Equal(t, tc.errorName, result.ErrorName())
			assert.Equal(t, domain, result.DomainName())
			assert.Equal(t, testErr, result.Unwrap())
		})
	}
}

// TestError tests the Error() method formatting
func TestError(t *testing.T) {
	type args struct {
		underlying error
	}
	type expected struct {
		contains []string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "error with underlying",
			args: args{
				underlying: errors.New("title is required"),
			},
			expected: expected{
				contains: []string{"ValidationError", "[Task]", "title is required"},
			},
		},
		{
			testName: "error without underlying",
			args: args{
				underlying: nil,
			},
			expected: expected{
				contains: []string{"NotFoundError", "[Task]"},
			},
		},

		// 🔤 特殊文字
		{
			testName: "error with emoji in domain",
			args: args{
				underlying: errors.New("test"),
			},
			expected: expected{
				contains: []string{"ValidationError", "[Task📋]", "test"},
			},
		},
		{
			testName: "error with Japanese in message",
			args: args{
				underlying: errors.New("タイトルが必要です"),
			},
			expected: expected{
				contains: []string{"ValidationError", "[タスク]", "タイトルが必要です"},
			},
		},

		// 📭 空文字
		{
			testName: "empty domain name",
			args: args{
				underlying: errors.New("test"),
			},
			expected: expected{
				contains: []string{"ValidationError", "[]", "test"},
			},
		},

		// 📏 境界値
		{
			testName: "very long error message",
			args: args{
				underlying: errors.New(strings.Repeat("a", 1000)),
			},
			expected: expected{
				contains: []string{"DatabaseError", "[Repository]", strings.Repeat("a", 100)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			var err apperror.AppError

			// Create appropriate error based on test case
			switch {
			case strings.Contains(tt.testName, "without underlying"):
				err = apperror.NewNotFoundError(tt.args.underlying, "Task")
			case strings.Contains(tt.testName, "emoji in domain"):
				err = apperror.NewValidationError(tt.args.underlying, "Task📋")
			case strings.Contains(tt.testName, "Japanese"):
				err = apperror.NewValidationError(tt.args.underlying, "タスク")
			case strings.Contains(tt.testName, "empty domain"):
				err = apperror.NewValidationError(tt.args.underlying, "")
			case strings.Contains(tt.testName, "very long"):
				err = apperror.NewDatabaseError(tt.args.underlying, "Repository")
			default:
				err = apperror.NewValidationError(tt.args.underlying, "Task")
			}

			errorMsg := err.Error()

			for _, substr := range tt.expected.contains {
				assert.Contains(t, errorMsg, substr)
			}
		})
	}
}

// TestUnwrap tests the Unwrap method
func TestUnwrap(t *testing.T) {
	type args struct {
		underlying error
	}
	type expected struct {
		isNil bool
		msg   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "unwrap returns underlying error",
			args:     args{underlying: errors.New("original error")},
			expected: expected{isNil: false, msg: "original error"},
		},

		// ⚠️ Nil
		{
			testName: "unwrap with nil underlying",
			args:     args{underlying: nil},
			expected: expected{isNil: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := apperror.NewValidationError(tt.args.underlying, "Task")
			unwrapped := err.Unwrap()

			if tt.expected.isNil {
				assert.Nil(t, unwrapped)
			} else {
				assert.NotNil(t, unwrapped)
				assert.Equal(t, tt.expected.msg, unwrapped.Error())
			}
		})
	}
}

// TestErrorChain tests error chain with errors.Is
func TestErrorChain(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := apperror.NewValidationError(originalErr, "Task")

	// ✅ 正常系: errors.Is should work with wrapped errors
	t.Run("errors.Is works with wrapped error", func(t *testing.T) {
		assert.True(t, errors.Is(appErr, originalErr))
	})

	// ❌ 異常系: errors.Is should return false for different error
	t.Run("errors.Is returns false for different error", func(t *testing.T) {
		differentErr := errors.New("different error")
		assert.False(t, errors.Is(appErr, differentErr))
	})
}

// TestErrorName tests the ErrorName method
func TestErrorName(t *testing.T) {
	tests := []struct {
		testName  string
		err       apperror.AppError
		errorName string
	}{
		{testName: "NotFoundError", err: apperror.NewNotFoundError(nil, "Task"), errorName: "NotFoundError"},
		{testName: "ValidationError", err: apperror.NewValidationError(nil, "Task"), errorName: "ValidationError"},
		{testName: "DatabaseError", err: apperror.NewDatabaseError(nil, "Repo"), errorName: "DatabaseError"},
		{testName: "UnauthorizedError", err: apperror.NewUnauthorizedError(nil, "Auth"), errorName: "UnauthorizedError"},
		{testName: "InternalServerError", err: apperror.NewInternalServerError(nil, "Server"), errorName: "InternalServerError"},
		{testName: "BadRequestError", err: apperror.NewBadRequestError(nil, "Request"), errorName: "BadRequestError"},
		{testName: "ConflictError", err: apperror.NewConflictError(nil, "Resource"), errorName: "ConflictError"},
		{testName: "ForbiddenError", err: apperror.NewForbiddenError(nil, "Access"), errorName: "ForbiddenError"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.errorName, tt.err.ErrorName())
		})
	}
}

// TestDomainName tests the DomainName method
func TestDomainName(t *testing.T) {
	tests := []struct {
		testName   string
		domainName string
	}{
		{testName: "Task domain", domainName: "Task"},
		{testName: "User domain", domainName: "User"},
		{testName: "Repository domain", domainName: "TaskRepository"},
		{testName: "Empty domain", domainName: ""},
		{testName: "Japanese domain", domainName: "タスク"},
		{testName: "Emoji domain", domainName: "Task📋"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := apperror.NewValidationError(nil, tt.domainName)
			assert.Equal(t, tt.domainName, err.DomainName())
		})
	}
}
