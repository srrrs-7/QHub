package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"utils/logger"

	"github.com/google/go-cmp/cmp"
)

func TestLogger(t *testing.T) {
	type expected struct {
		statusCode int
		nextCalled bool
	}

	tests := []struct {
		testName    string
		nextHandler http.HandlerFunc
		expected    expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "calls next handler and logs request",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "captures 404 status code",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			expected: expected{statusCode: http.StatusNotFound, nextCalled: true},
		},
		{
			testName: "captures 500 status code",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			expected: expected{statusCode: http.StatusInternalServerError, nextCalled: true},
		},
		{
			testName: "captures 201 status code",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			}),
			expected: expected{statusCode: http.StatusCreated, nextCalled: true},
		},
		{
			testName: "captures 204 status code",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			expected: expected{statusCode: http.StatusNoContent, nextCalled: true},
		},
		// 境界値 (Boundary Values)
		{
			testName: "default status 200 when WriteHeader not called",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Do not call WriteHeader — implicit 200
				w.Write([]byte("ok"))
			}),
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "captures 301 redirect status",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
			}),
			expected: expected{statusCode: http.StatusMovedPermanently, nextCalled: true},
		},
		{
			testName: "captures 400 bad request status",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			expected: expected{statusCode: http.StatusBadRequest, nextCalled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			nextCalled := false
			wrappedNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				tt.nextHandler.ServeHTTP(w, r)
			})

			req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
			w := httptest.NewRecorder()

			Logger(wrappedNext).ServeHTTP(w, req)

			resp := w.Result()
			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.nextCalled, nextCalled); diff != "" {
				t.Errorf("nextCalled mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLogger_WithDifferentHTTPMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			Logger(nextHandler).ServeHTTP(w, req)

			if !nextCalled {
				t.Error("expected next handler to be called")
			}
			if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLogLevelForStatus(t *testing.T) {
	tests := []struct {
		testName      string
		status        int
		expectFn      logFunc
		expectFnLabel string
	}{
		// 正常系 (Happy Path) — 2xx/3xx → Info
		{testName: "200 → Info", status: 200, expectFn: logger.Info, expectFnLabel: "Info"},
		{testName: "201 → Info", status: 201, expectFn: logger.Info, expectFnLabel: "Info"},
		{testName: "301 → Info", status: 301, expectFn: logger.Info, expectFnLabel: "Info"},
		{testName: "399 → Info", status: 399, expectFn: logger.Info, expectFnLabel: "Info"},
		// 境界値 (Boundary Values)
		{testName: "400 → Warn", status: 400, expectFn: logger.Warn, expectFnLabel: "Warn"},
		{testName: "401 → Warn", status: 401, expectFn: logger.Warn, expectFnLabel: "Warn"},
		{testName: "404 → Warn", status: 404, expectFn: logger.Warn, expectFnLabel: "Warn"},
		{testName: "499 → Warn", status: 499, expectFn: logger.Warn, expectFnLabel: "Warn"},
		{testName: "500 → Error", status: 500, expectFn: logger.Error, expectFnLabel: "Error"},
		{testName: "503 → Error", status: 503, expectFn: logger.Error, expectFnLabel: "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := logLevelForStatus(tt.status)
			// Compare function identity via reflect.ValueOf.Pointer().
			gotPtr := reflect.ValueOf(got).Pointer()
			wantPtr := reflect.ValueOf(tt.expectFn).Pointer()
			if gotPtr != wantPtr {
				t.Errorf("logLevelForStatus(%d) returned wrong function, want %s", tt.status, tt.expectFnLabel)
			}
		})
	}
}

func TestOutcomeFromStatus(t *testing.T) {
	type args struct{ status int }
	type expected struct{ outcome string }

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path) — 2xx
		{testName: "200 is success", args: args{200}, expected: expected{"success"}},
		{testName: "201 is success", args: args{201}, expected: expected{"success"}},
		{testName: "204 is success", args: args{204}, expected: expected{"success"}},

		// 境界値 (Boundary Values)
		{testName: "399 is success", args: args{399}, expected: expected{"success"}},
		{testName: "400 is client_error", args: args{400}, expected: expected{"client_error"}},
		{testName: "404 is client_error", args: args{404}, expected: expected{"client_error"}},
		{testName: "499 is client_error", args: args{499}, expected: expected{"client_error"}},
		{testName: "500 is server_error", args: args{500}, expected: expected{"server_error"}},
		{testName: "503 is server_error", args: args{503}, expected: expected{"server_error"}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := outcomeFromStatus(tt.args.status)
			if diff := cmp.Diff(tt.expected.outcome, got); diff != "" {
				t.Errorf("outcome mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStripAPIPrefix(t *testing.T) {
	type args struct{ pattern string }
	type expected struct{ result string }

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{testName: "strips /api/v1 prefix", args: args{"/api/v1/organizations"}, expected: expected{"/organizations"}},
		{testName: "strips /api/v1 from nested route", args: args{"/api/v1/organizations/{org_id}/projects"}, expected: expected{"/organizations/{org_id}/projects"}},

		// 異常系 (Error Cases)
		{testName: "no prefix passthrough", args: args{"/health"}, expected: expected{"/health"}},
		{testName: "partial prefix not stripped", args: args{"/api/organizations"}, expected: expected{"/api/organizations"}},

		// 空文字 (Empty/whitespace)
		{testName: "empty string", args: args{""}, expected: expected{""}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := stripAPIPrefix(tt.args.pattern)
			if diff := cmp.Diff(tt.expected.result, got); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLogger_RequestLogPropagation(t *testing.T) {
	tests := []struct {
		testName   string
		downstream http.HandlerFunc // simulates auth+RBAC middleware
		wantStatus int
		wantUserID string
		wantOrgID  string
		wantAuth   string
	}{
		// 正常系 — success path: all WHO fields set
		{
			testName: "WHO fields populated on 200 success",
			downstream: func(w http.ResponseWriter, r *http.Request) {
				rl := logger.RequestLogFrom(r.Context())
				rl.UserID = "user-abc"
				rl.OrgID = "org-xyz"
				rl.AuthMethod = "bearer"
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantUserID: "user-abc",
			wantOrgID:  "org-xyz",
			wantAuth:   "bearer",
		},
		// 異常系 — 4xx: partial WHO fields (e.g. org_id known, user_id unknown)
		{
			testName: "WHO fields partially populated on 401 client error",
			downstream: func(w http.ResponseWriter, r *http.Request) {
				// Simulates RequireRole: org_id parsed from URL, user_id missing.
				rl := logger.RequestLogFrom(r.Context())
				rl.OrgID = "org-xyz"
				rl.AuthMethod = "bearer"
				// user_id left empty (X-User-ID header absent)
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantStatus: http.StatusUnauthorized,
			wantUserID: "",
			wantOrgID:  "org-xyz",
			wantAuth:   "bearer",
		},
		// 空文字 / Null — no auth at all
		{
			testName: "empty WHO fields when no auth middleware runs",
			downstream: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantUserID: "",
			wantOrgID:  "",
			wantAuth:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Capture the RequestLog that Logger creates, by reading it from
			// inside the downstream (same pointer is accessible from the request context).
			var capturedRL *logger.RequestLog
			wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRL = logger.RequestLogFrom(r.Context())
				tt.downstream(w, r)
			})

			req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations", nil)
			w := httptest.NewRecorder()
			Logger(wrapped).ServeHTTP(w, req)

			if diff := cmp.Diff(tt.wantStatus, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
			if capturedRL == nil {
				t.Fatal("RequestLog was not initialised by Logger middleware")
			}
			if diff := cmp.Diff(tt.wantUserID, capturedRL.UserID); diff != "" {
				t.Errorf("UserID mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantOrgID, capturedRL.OrgID); diff != "" {
				t.Errorf("OrgID mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantAuth, capturedRL.AuthMethod); diff != "" {
				t.Errorf("AuthMethod mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResponseWriter(t *testing.T) {
	t.Run("newResponseWriter defaults to 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := newResponseWriter(w)

		if diff := cmp.Diff(http.StatusOK, rw.statusCode); diff != "" {
			t.Errorf("default status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("WriteHeader captures status code", func(t *testing.T) {
		tests := []struct {
			testName   string
			statusCode int
		}{
			{testName: "200 OK", statusCode: http.StatusOK},
			{testName: "201 Created", statusCode: http.StatusCreated},
			{testName: "204 No Content", statusCode: http.StatusNoContent},
			{testName: "301 Moved", statusCode: http.StatusMovedPermanently},
			{testName: "400 Bad Request", statusCode: http.StatusBadRequest},
			{testName: "401 Unauthorized", statusCode: http.StatusUnauthorized},
			{testName: "403 Forbidden", statusCode: http.StatusForbidden},
			{testName: "404 Not Found", statusCode: http.StatusNotFound},
			{testName: "500 Internal Server Error", statusCode: http.StatusInternalServerError},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				w := httptest.NewRecorder()
				rw := newResponseWriter(w)

				rw.WriteHeader(tt.statusCode)

				if diff := cmp.Diff(tt.statusCode, rw.statusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("WriteHeader delegates to underlying ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := newResponseWriter(w)

		rw.WriteHeader(http.StatusNotFound)

		// The underlying recorder should also have the status code
		if diff := cmp.Diff(http.StatusNotFound, w.Result().StatusCode); diff != "" {
			t.Errorf("underlying status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Write delegates to underlying ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		rw := newResponseWriter(w)

		body := []byte("hello world")
		n, err := rw.Write(body)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff := cmp.Diff(len(body), n); diff != "" {
			t.Errorf("bytes written mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("hello world", w.Body.String()); diff != "" {
			t.Errorf("body mismatch (-want +got):\n%s", diff)
		}
	})
}
