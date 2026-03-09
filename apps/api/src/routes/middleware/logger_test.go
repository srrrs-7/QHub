package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
