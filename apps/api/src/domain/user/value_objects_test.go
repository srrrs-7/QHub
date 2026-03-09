package user_test

import (
	"errors"
	"strings"
	"testing"

	"api/src/domain/apperror"
	"api/src/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// --- UserID ---

func TestNewUserID(t *testing.T) {
	type args struct {
		id string
	}
	type expected struct {
		wantErr bool
		errName string
	}

	validUUID := uuid.New().String()

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "valid UUID", args: args{id: validUUID}, expected: expected{wantErr: false}},
		// 異常系
		{testName: "invalid UUID", args: args{id: "not-a-uuid"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "nil-like UUID", args: args{id: "00000000-0000-0000-0000-000000000000"}, expected: expected{wantErr: false}},
		// 特殊文字
		{testName: "special chars", args: args{id: "<script>alert('xss')</script>"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "SQL injection", args: args{id: "'; DROP TABLE users;--"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "Japanese chars", args: args{id: "テスト"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{id: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "whitespace only", args: args{id: "   "}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := user.NewUserID(tt.args.id)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.args.id, result.String())
			}
		})
	}
}

func TestUserIDFromUUID(t *testing.T) {
	u := uuid.New()
	result := user.UserIDFromUUID(u)
	assert.Equal(t, u.String(), result.String())
	assert.Equal(t, u, result.UUID())
}

// --- UserEmail ---

func TestNewUserEmail(t *testing.T) {
	type args struct {
		email string
	}
	type expected struct {
		wantErr bool
		errName string
		value   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "valid email", args: args{email: "user@example.com"}, expected: expected{wantErr: false, value: "user@example.com"}},
		{testName: "email with subdomain", args: args{email: "user@mail.example.com"}, expected: expected{wantErr: false, value: "user@mail.example.com"}},
		{testName: "email with plus tag", args: args{email: "user+tag@example.com"}, expected: expected{wantErr: false, value: "user+tag@example.com"}},
		{testName: "email with dots in local", args: args{email: "first.last@example.com"}, expected: expected{wantErr: false, value: "first.last@example.com"}},
		{testName: "trimmed whitespace", args: args{email: "  user@example.com  "}, expected: expected{wantErr: false, value: "user@example.com"}},
		// 異常系
		{testName: "missing @", args: args{email: "userexample.com"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "missing domain", args: args{email: "user@"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "missing local part", args: args{email: "@example.com"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "double @", args: args{email: "user@@example.com"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "max length 255", args: args{email: strings.Repeat("a", 243) + "@example.com"}, expected: expected{wantErr: false, value: strings.Repeat("a", 243) + "@example.com"}},
		{testName: "over max length 256", args: args{email: strings.Repeat("a", 244) + "@example.com"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "single char local", args: args{email: "a@example.com"}, expected: expected{wantErr: false, value: "a@example.com"}},
		// 特殊文字
		{testName: "no TLD", args: args{email: "user@localhost"}, expected: expected{wantErr: false, value: "user@localhost"}},
		{testName: "SQL injection no at", args: args{email: "'; DROP TABLE users;--"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "just text", args: args{email: "notanemail"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "spaces in local part", args: args{email: "user name@example.com"}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 空文字
		{testName: "empty string", args: args{email: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "whitespace only", args: args{email: "   "}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// Null/Nil (zero value)
		{testName: "zero value string", args: args{email: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := user.NewUserEmail(tt.args.email)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.value, result.String())
			}
		})
	}
}

// --- UserName ---

func TestNewUserName(t *testing.T) {
	type args struct {
		name string
	}
	type expected struct {
		wantErr bool
		errName string
		value   string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{testName: "valid name", args: args{name: "John Doe"}, expected: expected{wantErr: false, value: "John Doe"}},
		{testName: "single character", args: args{name: "A"}, expected: expected{wantErr: false, value: "A"}},
		// 異常系
		{testName: "too long (101 chars)", args: args{name: strings.Repeat("a", 101)}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// 境界値
		{testName: "exactly 100 chars", args: args{name: strings.Repeat("a", 100)}, expected: expected{wantErr: false, value: strings.Repeat("a", 100)}},
		{testName: "min length (1 char)", args: args{name: "X"}, expected: expected{wantErr: false, value: "X"}},
		// 特殊文字
		{testName: "Japanese name", args: args{name: "田中太郎"}, expected: expected{wantErr: false, value: "田中太郎"}},
		{testName: "emoji in name", args: args{name: "User 🚀"}, expected: expected{wantErr: false, value: "User 🚀"}},
		{testName: "SQL injection", args: args{name: "'; DROP TABLE users;--"}, expected: expected{wantErr: false, value: "'; DROP TABLE users;--"}},
		{testName: "XSS attempt", args: args{name: "<script>alert('xss')</script>"}, expected: expected{wantErr: false, value: "<script>alert('xss')</script>"}},
		// 空文字
		{testName: "empty string", args: args{name: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		{testName: "whitespace only", args: args{name: "   "}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
		// Null/Nil (zero value)
		{testName: "zero value string", args: args{name: ""}, expected: expected{wantErr: true, errName: apperror.ValidationErrorName}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result, err := user.NewUserName(tt.args.name)

			if tt.expected.wantErr {
				assert.Error(t, err)
				var appErr apperror.AppError
				assert.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.expected.errName, appErr.ErrorName())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.value, result.String())
			}
		})
	}
}
