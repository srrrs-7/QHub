package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewAPIClient_AuthHeader(t *testing.T) {
	type args struct {
		envToken string // value of API_AUTH_TOKEN env var ("" means unset)
	}
	type expected struct {
		authHeader string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "uses API_AUTH_TOKEN when set",
			args:     args{envToken: "my-secret-token"},
			expected: expected{authHeader: "Bearer my-secret-token"},
		},

		// 異常系 (Error Cases)
		{
			testName: "falls back to dev-token when API_AUTH_TOKEN is empty",
			args:     args{envToken: ""},
			expected: expected{authHeader: "Bearer dev-token"},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "handles token with special characters",
			args:     args{envToken: "tok-abc123_XYZ.v2"},
			expected: expected{authHeader: "Bearer tok-abc123_XYZ.v2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.args.envToken != "" {
				t.Setenv("API_AUTH_TOKEN", tt.args.envToken)
			}

			var gotHeader string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeader = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := NewAPIClient(srv.URL)
			// Trigger an HTTP call; ignore the result — we only care about the header.
			//nolint:errcheck
			c.do(t.Context(), http.MethodGet, "/", nil, nil)

			if diff := cmp.Diff(tt.expected.authHeader, gotHeader); diff != "" {
				t.Errorf("Authorization header mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
