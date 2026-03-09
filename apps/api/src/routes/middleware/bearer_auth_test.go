package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"utils/logger"

	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}

func TestBearerAuth(t *testing.T) {
	// alwaysValid is a validator that always returns true.
	alwaysValid := func(token string) (bool, error) { return true, nil }
	// alwaysInvalid is a validator that always returns false.
	alwaysInvalid := func(token string) (bool, error) { return false, nil }
	// alwaysError is a validator that always returns an error.
	alwaysError := func(token string) (bool, error) { return false, errors.New("validation error") }

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	type args struct {
		authHeader string
		validator  TokenValidator
	}
	type expected struct {
		statusCode int
		nextCalled bool
		errMsg     string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "valid bearer token passes through",
			args:     args{authHeader: "Bearer valid-token-123", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "bearer prefix is case-insensitive",
			args:     args{authHeader: "bearer my-token", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "BEARER prefix is case-insensitive",
			args:     args{authHeader: "BEARER my-token", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},

		// 異常系 (Error Cases)
		{
			testName: "missing authorization header returns 401",
			args:     args{authHeader: "", validator: alwaysValid},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "missing authorization header"},
		},
		{
			testName: "malformed header without Bearer prefix returns 401",
			args:     args{authHeader: "Basic abc123", validator: alwaysValid},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid authorization header format"},
		},
		{
			testName: "token only without scheme returns 401",
			args:     args{authHeader: "sometoken", validator: alwaysValid},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid authorization header format"},
		},
		{
			testName: "validator returns false returns 401",
			args:     args{authHeader: "Bearer rejected-token", validator: alwaysInvalid},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid token"},
		},
		{
			testName: "validator returns error returns 401",
			args:     args{authHeader: "Bearer error-token", validator: alwaysError},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "token validation failed"},
		},

		// 境界値 (Boundary Values)
		{
			testName: "Bearer with empty token returns 401",
			args:     args{authHeader: "Bearer ", validator: alwaysValid},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "empty bearer token"},
		},
		{
			testName: "single character token is accepted",
			args:     args{authHeader: "Bearer x", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "very long token is accepted",
			args: args{
				authHeader: "Bearer " + string(make([]byte, 4096)),
				validator:  alwaysValid,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "token with special characters",
			args:     args{authHeader: "Bearer abc!@#$%^&*()_+-=", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "token with unicode characters",
			args:     args{authHeader: "Bearer トークン🔑", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "token with spaces after Bearer",
			args:     args{authHeader: "Bearer token with spaces", validator: alwaysValid},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},

		// 空文字 (Empty/whitespace)
		{
			testName: "whitespace-only authorization header returns 401",
			args:     args{authHeader: "   ", validator: alwaysValid},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid authorization header format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			nextCalled = false

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.args.authHeader != "" {
				req.Header.Set("Authorization", tt.args.authHeader)
			}
			w := httptest.NewRecorder()

			middleware := BearerAuth(tt.args.validator)
			middleware(nextHandler).ServeHTTP(w, req)

			resp := w.Result()

			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expected.nextCalled, nextCalled); diff != "" {
				t.Errorf("nextCalled mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.errMsg != "" {
				var body errorResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if diff := cmp.Diff(tt.expected.errMsg, body.Error); diff != "" {
					t.Errorf("error message mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestBearerAuth_TokenInContext(t *testing.T) {
	expectedToken := "my-secret-token"
	var capturedToken string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedToken = GetBearerToken(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	validator := func(token string) (bool, error) { return true, nil }

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expectedToken)
	w := httptest.NewRecorder()

	BearerAuth(validator)(handler).ServeHTTP(w, req)

	if diff := cmp.Diff(expectedToken, capturedToken); diff != "" {
		t.Errorf("token in context mismatch (-want +got):\n%s", diff)
	}
}

func TestGetBearerToken(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type expected struct {
		token string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "returns token from context",
			args:     args{ctx: context.WithValue(context.Background(), bearerTokenKey, "test-token")},
			expected: expected{token: "test-token"},
		},
		// 異常系 (Error Cases)
		{
			testName: "returns empty string when context has wrong type",
			args:     args{ctx: context.WithValue(context.Background(), bearerTokenKey, 12345)},
			expected: expected{token: ""},
		},
		// 空文字 (Empty/whitespace)
		{
			testName: "returns empty string when no token in context",
			args:     args{ctx: context.Background()},
			expected: expected{token: ""},
		},
		// Null/Nil
		{
			testName: "returns empty string for nil value in context",
			args:     args{ctx: context.WithValue(context.Background(), bearerTokenKey, nil)},
			expected: expected{token: ""},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "returns token with special characters",
			args:     args{ctx: context.WithValue(context.Background(), bearerTokenKey, "トークン🔑!@#")},
			expected: expected{token: "トークン🔑!@#"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetBearerToken(tt.args.ctx)
			if diff := cmp.Diff(tt.expected.token, got); diff != "" {
				t.Errorf("token mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
