package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCORS(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	type args struct {
		method  string
		origin  string
		envVar  string
		headers map[string]string
	}
	type expected struct {
		statusCode             int
		allowOrigin            string
		allowMethods           string
		allowHeaders           string
		exposeHeaders          string
		maxAge                 string
		allowCredentials       string
		allowOriginAbsent      bool
		allowCredentialsAbsent bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "preflight OPTIONS request returns correct headers with wildcard origin",
			args: args{
				method: http.MethodOptions,
				origin: "http://localhost:3000",
				envVar: "",
				headers: map[string]string{
					"Access-Control-Request-Method": "POST",
				},
			},
			expected: expected{
				statusCode:       http.StatusNoContent,
				allowOrigin:      "*",
				allowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
				allowHeaders:     "Authorization, Content-Type, X-API-Key",
				exposeHeaders:    "X-Request-ID",
				maxAge:           "300",
				allowCredentials: "",
				allowCredentialsAbsent: true,
			},
		},
		{
			testName: "simple GET request includes CORS headers",
			args: args{
				method: http.MethodGet,
				origin: "http://localhost:3000",
				envVar: "",
			},
			expected: expected{
				statusCode:    http.StatusOK,
				allowOrigin:   "*",
				exposeHeaders: "X-Request-ID",
				allowCredentialsAbsent: true,
			},
		},
		{
			testName: "allowed origin with specific origins configured",
			args: args{
				method: http.MethodGet,
				origin: "https://app.example.com",
				envVar: "https://app.example.com,https://admin.example.com",
			},
			expected: expected{
				statusCode:       http.StatusOK,
				allowOrigin:      "https://app.example.com",
				exposeHeaders:    "X-Request-ID",
				allowCredentials: "true",
			},
		},
		{
			testName: "preflight with specific origins returns correct headers and credentials",
			args: args{
				method: http.MethodOptions,
				origin: "https://admin.example.com",
				envVar: "https://app.example.com,https://admin.example.com",
				headers: map[string]string{
					"Access-Control-Request-Method": "PUT",
				},
			},
			expected: expected{
				statusCode:       http.StatusNoContent,
				allowOrigin:      "https://admin.example.com",
				allowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
				allowHeaders:     "Authorization, Content-Type, X-API-Key",
				exposeHeaders:    "X-Request-ID",
				maxAge:           "300",
				allowCredentials: "true",
			},
		},

		// 異常系 (Error Cases)
		{
			testName: "disallowed origin is rejected when specific origins configured",
			args: args{
				method: http.MethodGet,
				origin: "https://evil.com",
				envVar: "https://app.example.com",
			},
			expected: expected{
				statusCode:             http.StatusOK,
				allowOriginAbsent:      true,
				allowCredentialsAbsent: true,
			},
		},
		{
			testName: "disallowed origin preflight is rejected",
			args: args{
				method: http.MethodOptions,
				origin: "https://evil.com",
				envVar: "https://app.example.com",
				headers: map[string]string{
					"Access-Control-Request-Method": "POST",
				},
			},
			expected: expected{
				statusCode:             http.StatusNoContent,
				allowOriginAbsent:      true,
				allowCredentialsAbsent: true,
			},
		},

		// 境界値 (Boundary Values)
		{
			testName: "empty CORS_ORIGINS env var defaults to wildcard",
			args: args{
				method: http.MethodGet,
				origin: "http://anything.com",
				envVar: "",
			},
			expected: expected{
				statusCode:             http.StatusOK,
				allowOrigin:            "*",
				exposeHeaders:          "X-Request-ID",
				allowCredentialsAbsent: true,
			},
		},
		{
			testName: "request without Origin header gets no CORS headers",
			args: args{
				method: http.MethodGet,
				origin: "",
				envVar: "",
			},
			expected: expected{
				statusCode:             http.StatusOK,
				allowOriginAbsent:      true,
				allowCredentialsAbsent: true,
			},
		},
		{
			testName: "single origin in env var",
			args: args{
				method: http.MethodGet,
				origin: "https://only.example.com",
				envVar: "https://only.example.com",
			},
			expected: expected{
				statusCode:       http.StatusOK,
				allowOrigin:      "https://only.example.com",
				exposeHeaders:    "X-Request-ID",
				allowCredentials: "true",
			},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "origin with port number is matched correctly",
			args: args{
				method: http.MethodGet,
				origin: "http://localhost:8080",
				envVar: "http://localhost:8080,http://localhost:3000",
			},
			expected: expected{
				statusCode:       http.StatusOK,
				allowOrigin:      "http://localhost:8080",
				exposeHeaders:    "X-Request-ID",
				allowCredentials: "true",
			},
		},
		{
			testName: "origin with trailing whitespace in env var is trimmed",
			args: args{
				method: http.MethodGet,
				origin: "https://app.example.com",
				envVar: " https://app.example.com , https://other.example.com ",
			},
			expected: expected{
				statusCode:       http.StatusOK,
				allowOrigin:      "https://app.example.com",
				exposeHeaders:    "X-Request-ID",
				allowCredentials: "true",
			},
		},

		// 空文字 (Empty/Whitespace)
		{
			testName: "whitespace-only CORS_ORIGINS defaults to wildcard",
			args: args{
				method: http.MethodGet,
				origin: "http://example.com",
				envVar: "   ",
			},
			expected: expected{
				statusCode:             http.StatusOK,
				allowOrigin:            "*",
				exposeHeaders:          "X-Request-ID",
				allowCredentialsAbsent: true,
			},
		},

		// Null/Nil equivalents
		{
			testName: "OPTIONS without Access-Control-Request-Method still returns CORS headers",
			args: args{
				method: http.MethodOptions,
				origin: "http://localhost:3000",
				envVar: "",
			},
			expected: expected{
				statusCode:             http.StatusNoContent,
				allowOrigin:            "*",
				allowMethods:           "GET, POST, PUT, DELETE, OPTIONS",
				allowHeaders:           "Authorization, Content-Type, X-API-Key",
				exposeHeaders:          "X-Request-ID",
				maxAge:                 "300",
				allowCredentialsAbsent: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := CORS(tt.args.envVar)(dummyHandler)

			req := httptest.NewRequest(tt.args.method, "/api/v1/tasks", nil)
			if tt.args.origin != "" {
				req.Header.Set("Origin", tt.args.origin)
			}
			for k, v := range tt.args.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()

			// Check status code
			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			// Check Access-Control-Allow-Origin
			if tt.expected.allowOriginAbsent {
				if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
					t.Errorf("expected Access-Control-Allow-Origin to be absent, got %q", got)
				}
			} else if tt.expected.allowOrigin != "" {
				if diff := cmp.Diff(tt.expected.allowOrigin, resp.Header.Get("Access-Control-Allow-Origin")); diff != "" {
					t.Errorf("Access-Control-Allow-Origin mismatch (-want +got):\n%s", diff)
				}
			}

			// Check Access-Control-Allow-Methods (only on preflight)
			if tt.expected.allowMethods != "" {
				if diff := cmp.Diff(tt.expected.allowMethods, resp.Header.Get("Access-Control-Allow-Methods")); diff != "" {
					t.Errorf("Access-Control-Allow-Methods mismatch (-want +got):\n%s", diff)
				}
			}

			// Check Access-Control-Allow-Headers (only on preflight)
			if tt.expected.allowHeaders != "" {
				if diff := cmp.Diff(tt.expected.allowHeaders, resp.Header.Get("Access-Control-Allow-Headers")); diff != "" {
					t.Errorf("Access-Control-Allow-Headers mismatch (-want +got):\n%s", diff)
				}
			}

			// Check Access-Control-Expose-Headers
			if tt.expected.exposeHeaders != "" {
				if diff := cmp.Diff(tt.expected.exposeHeaders, resp.Header.Get("Access-Control-Expose-Headers")); diff != "" {
					t.Errorf("Access-Control-Expose-Headers mismatch (-want +got):\n%s", diff)
				}
			}

			// Check Access-Control-Max-Age (only on preflight)
			if tt.expected.maxAge != "" {
				if diff := cmp.Diff(tt.expected.maxAge, resp.Header.Get("Access-Control-Max-Age")); diff != "" {
					t.Errorf("Access-Control-Max-Age mismatch (-want +got):\n%s", diff)
				}
			}

			// Check Access-Control-Allow-Credentials
			if tt.expected.allowCredentialsAbsent {
				if got := resp.Header.Get("Access-Control-Allow-Credentials"); got != "" {
					t.Errorf("expected Access-Control-Allow-Credentials to be absent, got %q", got)
				}
			} else if tt.expected.allowCredentials != "" {
				if diff := cmp.Diff(tt.expected.allowCredentials, resp.Header.Get("Access-Control-Allow-Credentials")); diff != "" {
					t.Errorf("Access-Control-Allow-Credentials mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
